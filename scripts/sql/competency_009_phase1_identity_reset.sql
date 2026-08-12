-- ============================================================
-- 胜任力测验迁移 009：一期 A/B 维度身份重置
-- 兼容 MySQL 5.7+；一次性、带持久化标记；事务内替换源题/文案/维度主数据。
-- 仅允许用户指定的 staging 重置窗口执行；禁止加入 production 通用迁移序列。
-- 前置条件：必须先通过应用既有整链删除清空全部胜任力测评和报告文件。
-- 本迁移不删除传统测评、传统题目、传统答卷，也不导入一期正式题目或评分文案。
-- ============================================================
SET NAMES utf8mb4;

-- 调用方必须在同一数据库会话中显式设置以下变量，脚本不会自行授权：
--   SET @competency_009_target_environment='staging';
--   SET @competency_009_writes_quiesced=1;
-- 第二项表示应用写流量已停止，且整链删除与本迁移将在同一维护窗口完成。
SET @competency_009_guard_sql=IF(
  IFNULL(@competency_009_target_environment,'')='staging'
    AND IFNULL(@competency_009_writes_quiesced,0)=1,
  'SELECT 1',
  'SELECT JSON_EXTRACT(''COMPETENCY_009_REQUIRES_EXPLICIT_STAGING_AUTHORIZATION'',''$'')'
);
PREPARE competency_009_guard_stmt FROM @competency_009_guard_sql;
EXECUTE competency_009_guard_stmt;
DEALLOCATE PREPARE competency_009_guard_stmt;

SET @competency_009_lock_acquired=GET_LOCK('competency-009-phase1-identity-reset',30);
SET @competency_009_guard_sql=IF(
  @competency_009_lock_acquired=1,
  'SELECT 1',
  'SELECT JSON_EXTRACT(''COMPETENCY_009_MIGRATION_LOCK_NOT_ACQUIRED'',''$'')'
);
PREPARE competency_009_guard_stmt FROM @competency_009_guard_sql;
EXECUTE competency_009_guard_stmt;
DEALLOCATE PREPARE competency_009_guard_stmt;

CREATE TABLE IF NOT EXISTS `el_competency_migration` (
  `migration_key` varchar(100) NOT NULL,
  `applied_at` datetime NOT NULL,
  `remark` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`migration_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力一次性数据迁移标记';

-- 依赖清理必须由应用整链删除完成，以便事务删除业务行并在提交后安全删除报告文件。
-- 任一测评、快照、试卷题、结果或报告实例残留时，本迁移都在写入前硬阻断。
SELECT
  (SELECT COUNT(*) FROM `el_exam` WHERE `assessment_type`='competency') +
  (SELECT COUNT(*) FROM `el_exam_competency_group`) +
  (SELECT COUNT(*) FROM `el_exam_competency_dimension`) +
  (SELECT COUNT(*) FROM `el_exam_competency_question`) +
  (SELECT COUNT(*) FROM `el_paper_qu` WHERE `exam_question_id` IS NOT NULL) +
    (SELECT COUNT(*) FROM `el_paper_qu` pq
      INNER JOIN `el_qu` q ON q.`id`=pq.`qu_id`
      WHERE q.`dimension_id` IS NOT NULL OR q.`competency_question_type` IS NOT NULL) +
    (SELECT COUNT(*) FROM `el_paper_qu_answer` pqa
      INNER JOIN `el_qu` q ON q.`id`=pqa.`qu_id`
      WHERE q.`dimension_id` IS NOT NULL OR q.`competency_question_type` IS NOT NULL) +
    (SELECT COUNT(*) FROM `el_user_book` ub
      INNER JOIN `el_qu` q ON q.`id`=ub.`qu_id`
      WHERE q.`dimension_id` IS NOT NULL OR q.`competency_question_type` IS NOT NULL) +
  (SELECT COUNT(*) FROM `el_competency_group_result`) +
  (SELECT COUNT(*) FROM `el_competency_validity_result`) +
  (SELECT COUNT(*) FROM `el_competency_dimension_result`) +
  (SELECT COUNT(*) FROM `el_competency_result`) +
  (SELECT COUNT(*) FROM `el_competency_report_audit`) +
  (SELECT COUNT(*) FROM `el_competency_report`)
INTO @remaining_competency_dependencies;

SELECT COUNT(*)
INTO @identity_reset_already_applied
FROM `el_competency_migration`
WHERE `migration_key`='competency-009-phase1-identity-reset';

-- 已完成的marker允许安全重跑；首次执行只要还有依赖就用确定性错误对象名硬阻断。
SET @competency_009_guard_sql=IF(
  @identity_reset_already_applied=1 OR @remaining_competency_dependencies=0,
  'SELECT 1',
  'SELECT JSON_EXTRACT(''COMPETENCY_009_REMAINING_DEPENDENCIES'',''$'')'
);
PREPARE competency_009_guard_stmt FROM @competency_009_guard_sql;
EXECUTE competency_009_guard_stmt;
DEALLOCATE PREPARE competency_009_guard_stmt;

START TRANSACTION;

INSERT IGNORE INTO `el_competency_migration` (`migration_key`,`applied_at`,`remark`)
VALUES ('competency-009-phase1-identity-reset',NOW(),'一期 A/B 维度身份重置');
SET @apply_reset=ROW_COUNT();

-- 只清理带胜任力身份的源题关系；传统题目元数据均为空，不会进入以下条件。
DELETE qa FROM `el_qu_answer` qa
INNER JOIN `el_qu` q ON q.`id`=qa.`qu_id`
WHERE @apply_reset=1
  AND qa.`id` IS NOT NULL
  AND (q.`dimension_id` IS NOT NULL OR q.`competency_question_type` IS NOT NULL);

DELETE qr FROM `el_qu_repo` qr
INNER JOIN `el_qu` q ON q.`id`=qr.`qu_id`
WHERE @apply_reset=1
  AND qr.`id` IS NOT NULL
  AND (q.`dimension_id` IS NOT NULL OR q.`competency_question_type` IS NOT NULL);

DELETE q FROM `el_qu` q
WHERE @apply_reset=1
  AND q.`id` IS NOT NULL
  AND (q.`dimension_id` IS NOT NULL OR q.`competency_question_type` IS NOT NULL);

-- 旧临时报告文案绑定退役身份/四档内容，不能静默沿用到一期产品。
DELETE rt FROM `el_competency_report_text` rt
WHERE @apply_reset=1
  AND rt.`id` IS NOT NULL;

DELETE d FROM `el_competency_dimension` d
WHERE @apply_reset=1
  AND d.`id` IS NOT NULL;

-- 仅初始化用户已确认的一期十个稳定身份；第二、三期总体池尚未定稿，不在此推断。
INSERT INTO `el_competency_dimension`
(`id`,`code`,`name`,`vird_level`,`applicable_category`,`core_meaning`,`display_order`,`status`,`create_time`,`update_time`)
SELECT 'competency-a1-01','A1-01','逻辑思维','通用能力','基层员工','逻辑分析严谨，推理判断有据',1,0,NOW(),NOW() FROM DUAL WHERE @apply_reset=1
UNION ALL SELECT 'competency-a1-02','A1-02','数字应用','通用能力','基层员工','善用数字化工具与AI技术，具备数据思维',2,0,NOW(),NOW() FROM DUAL WHERE @apply_reset=1
UNION ALL SELECT 'competency-a1-03','A1-03','计划执行','通用能力','基层员工','高效推进计划并达成预期结果',3,0,NOW(),NOW() FROM DUAL WHERE @apply_reset=1
UNION ALL SELECT 'competency-a1-04','A1-04','持续学习','通用能力','基层员工','主动学习，多渠道获取知识并学以致用',4,0,NOW(),NOW() FROM DUAL WHERE @apply_reset=1
UNION ALL SELECT 'competency-a1-05','A1-05','沟通表达','通用能力','基层员工','清晰传递信息，重视倾听与反馈',5,0,NOW(),NOW() FROM DUAL WHERE @apply_reset=1
UNION ALL SELECT 'competency-b1-01','B1-01','敬业奉献','心理素养','基层员工','视工作为使命，全心投入，甘于奉献',6,0,NOW(),NOW() FROM DUAL WHERE @apply_reset=1
UNION ALL SELECT 'competency-b1-02','B1-02','求真务实','心理素养','基层员工','追求真理，尊重事实，注重实效',7,0,NOW(),NOW() FROM DUAL WHERE @apply_reset=1
UNION ALL SELECT 'competency-b1-03','B1-03','自律性','心理素养','基层员工','自我约束，规划在先，言行一致',8,0,NOW(),NOW() FROM DUAL WHERE @apply_reset=1
UNION ALL SELECT 'competency-b1-04','B1-04','成就导向','心理素养','基层员工','追求工作成功，不断挑战更高目标',9,0,NOW(),NOW() FROM DUAL WHERE @apply_reset=1
UNION ALL SELECT 'competency-b1-05','B1-05','合作意识','心理素养','基层员工','主动协作，乐于分享，促成共赢',10,0,NOW(),NOW() FROM DUAL WHERE @apply_reset=1;

COMMIT;

DO RELEASE_LOCK('competency-009-phase1-identity-reset');

-- 只读验证：应为一个迁移标记、十个启用维度、连续顺序和零胜任力源题/旧文案。
SELECT COUNT(*) AS migration_marker_count
FROM `el_competency_migration`
WHERE `migration_key`='competency-009-phase1-identity-reset';

SELECT COUNT(*) AS dimension_count,
       COUNT(DISTINCT `id`) AS unique_ids,
       COUNT(DISTINCT `code`) AS unique_codes,
       COUNT(DISTINCT `name`) AS unique_names,
       COUNT(DISTINCT `display_order`) AS unique_orders,
       MIN(`display_order`) AS min_order,
       MAX(`display_order`) AS max_order,
       SUM(`status`=0) AS enabled_count,
       SUM(`code` REGEXP '^[AB][0-9]+-[0-9]+$') AS ab_code_count
FROM `el_competency_dimension`;

SELECT `id`,`code`,`name`,`vird_level`,`applicable_category`,`core_meaning`,`display_order`,`status`
FROM `el_competency_dimension`
ORDER BY `display_order`;

SELECT
  (SELECT COUNT(*) FROM `el_qu` WHERE `dimension_id` IS NOT NULL OR `competency_question_type` IS NOT NULL) AS competency_source_questions,
  (SELECT COUNT(*) FROM `el_competency_report_text`) AS competency_report_texts,
  (SELECT COUNT(*) FROM `el_exam` WHERE `assessment_type`='competency') AS competency_exams,
  (SELECT COUNT(*) FROM `el_competency_result`) AS competency_results,
  (SELECT COUNT(*) FROM `el_competency_report`) AS competency_reports;
