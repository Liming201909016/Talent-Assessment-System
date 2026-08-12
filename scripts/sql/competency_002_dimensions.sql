-- ============================================================
-- 胜任力测验迁移 002：一期 A/B 维度主数据与测评维度选择表
-- 兼容：MySQL 5.7+
-- 特性：可重复执行；不删除现有数据；不覆盖管理员维护的维度主数据
-- 新库仅初始化已确认的一期 10 个维度；既有 D 体系环境由迁移 009 重置。
-- ============================================================

SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `el_competency_dimension` (
  `id` varchar(64) NOT NULL COMMENT '稳定ID',
  `code` varchar(16) NOT NULL COMMENT '稳定维度编码',
  `name` varchar(100) NOT NULL COMMENT '维度名称',
  `vird_level` varchar(100) NOT NULL COMMENT '能力层级',
  `applicable_category` varchar(50) NOT NULL COMMENT '适用对象',
  `core_meaning` varchar(500) NOT NULL COMMENT '核心含义',
  `display_order` int NOT NULL COMMENT '默认展示顺序',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0启用1停用',
  `create_time` datetime DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_competency_dimension_code` (`code`),
  UNIQUE KEY `uk_competency_dimension_name` (`name`),
  UNIQUE KEY `uk_competency_dimension_order` (`display_order`),
  KEY `idx_competency_dimension_status_order` (`status`,`display_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力维度主数据';

CREATE TABLE IF NOT EXISTS `el_exam_competency_dimension` (
  `id` varchar(64) NOT NULL COMMENT 'ID',
  `exam_id` varchar(64) NOT NULL COMMENT '测评ID',
  `dimension_id` varchar(64) NOT NULL COMMENT '维度ID',
  `dimension_code` varchar(16) NOT NULL DEFAULT '' COMMENT '发布快照：维度编号',
  `dimension_name` varchar(100) NOT NULL DEFAULT '' COMMENT '发布快照：维度名称',
  `vird_level` varchar(100) NOT NULL DEFAULT '' COMMENT '发布快照：能力层级',
  `applicable_category` varchar(50) NOT NULL DEFAULT '' COMMENT '发布快照：适用对象',
  `core_meaning` varchar(500) NOT NULL DEFAULT '' COMMENT '发布快照：核心含义',
  `display_order` int NOT NULL DEFAULT 0 COMMENT '发布快照：默认顺序',
  `question_count` int NOT NULL DEFAULT 0 COMMENT '发布时启用题数',
  `create_time` datetime DEFAULT NULL,
  `snapshot_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_exam_competency_dimension` (`exam_id`,`dimension_id`),
  KEY `idx_exam_competency_dimension_order` (`exam_id`,`display_order`),
  KEY `idx_exam_competency_dimension_dimension` (`dimension_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='测评选择的胜任力维度及发布快照';

-- 关联列必须继承现有 el_exam.id 的字符集和排序规则。
-- staging MySQL 8 使用 utf8mb4_0900_ai_ci，MySQL 5.7 环境使用 utf8mb4_general_ci；
-- 动态读取可同时兼容两类环境，并确保重复执行时修复已存在表。
SELECT CHARACTER_SET_NAME, COLLATION_NAME
INTO @exam_id_charset, @exam_id_collation
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_exam'
  AND COLUMN_NAME = 'id';
SET @sql = CONCAT(
  'ALTER TABLE `el_exam_competency_dimension` MODIFY COLUMN `exam_id` varchar(64) CHARACTER SET ',
  @exam_id_charset,
  ' COLLATE ',
  @exam_id_collation,
  ' NOT NULL COMMENT ''测评ID'''
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO `el_competency_dimension`
(`id`,`code`,`name`,`vird_level`,`applicable_category`,`core_meaning`,`display_order`,`status`,`create_time`,`update_time`)
VALUES
('competency-a1-01','A1-01','逻辑思维','通用能力','基层员工','逻辑分析严谨，推理判断有据',1,0,NOW(),NOW()),
('competency-a1-02','A1-02','数字应用','通用能力','基层员工','善用数字化工具与AI技术，具备数据思维',2,0,NOW(),NOW()),
('competency-a1-03','A1-03','计划执行','通用能力','基层员工','高效推进计划并达成预期结果',3,0,NOW(),NOW()),
('competency-a1-04','A1-04','持续学习','通用能力','基层员工','主动学习，多渠道获取知识并学以致用',4,0,NOW(),NOW()),
('competency-a1-05','A1-05','沟通表达','通用能力','基层员工','清晰传递信息，重视倾听与反馈',5,0,NOW(),NOW()),
('competency-b1-01','B1-01','敬业奉献','心理素养','基层员工','视工作为使命，全心投入，甘于奉献',6,0,NOW(),NOW()),
('competency-b1-02','B1-02','求真务实','心理素养','基层员工','追求真理，尊重事实，注重实效',7,0,NOW(),NOW()),
('competency-b1-03','B1-03','自律性','心理素养','基层员工','自我约束，规划在先，言行一致',8,0,NOW(),NOW()),
('competency-b1-04','B1-04','成就导向','心理素养','基层员工','追求工作成功，不断挑战更高目标',9,0,NOW(),NOW()),
('competency-b1-05','B1-05','合作意识','心理素养','基层员工','主动协作，乐于分享，促成共赢',10,0,NOW(),NOW())
ON DUPLICATE KEY UPDATE
  `id`=`id`;

-- 只读验证：新库应为一期 10 条、序号 1-10；既有环境须先执行迁移 009。
SELECT COUNT(*) AS dimension_count,
       MIN(display_order) AS min_order,
       MAX(display_order) AS max_order,
       COUNT(DISTINCT code) AS unique_codes,
       COUNT(DISTINCT name) AS unique_names
FROM el_competency_dimension;

SELECT id, code, name, display_order, status
FROM el_competency_dimension
WHERE code IN ('A1-01','A1-02','A1-03','A1-04','A1-05','B1-01','B1-02','B1-03','B1-04','B1-05')
ORDER BY display_order;
