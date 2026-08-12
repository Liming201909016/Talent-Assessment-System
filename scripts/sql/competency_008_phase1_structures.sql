-- ============================================================
-- 胜任力测验迁移 008：题目类型、效度结果与一级维度结果结构
-- 兼容 MySQL 5.7+；幂等；不实现效度阈值、一级聚合或五档规则
-- ============================================================
SET NAMES utf8mb4;

-- 1. 源题题型：NULL 保留传统题；dimension/validity 由胜任力专用流程写入。
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_qu' AND COLUMN_NAME='competency_question_type';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_qu` ADD COLUMN `competency_question_type` varchar(32) DEFAULT NULL COMMENT ''胜任力题型：dimension/validity'' AFTER `scoring_direction`',
  'SELECT ''el_qu.competency_question_type exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 已有带维度的胜任力题来自 003-007，只能是既有维度题；不改传统题和已明确分类的题。
UPDATE `el_qu`
SET `competency_question_type`='dimension'
WHERE `id` IS NOT NULL
  AND `competency_question_type` IS NULL
  AND `dimension_id` IS NOT NULL;

-- 原唯一索引未区分题型，会阻止同一关联维度下的效度题使用独立题号。
-- 先创建临时替代索引验证数据，再移除旧索引并重命名，避免迁移失败时失去唯一性保护。
SELECT GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX SEPARATOR ','), MIN(NON_UNIQUE)
INTO @qu_dimension_item_columns,@qu_dimension_item_non_unique
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_qu' AND INDEX_NAME='uk_qu_dimension_item';
SELECT COUNT(*) INTO @replacement_index_exists FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_qu' AND INDEX_NAME='uk_qu_dimension_item_v2_tmp';
SET @sql=IF(
  NOT (@qu_dimension_item_columns='dimension_id,competency_question_type,dimension_item_no' AND @qu_dimension_item_non_unique=0)
    AND @replacement_index_exists=0,
  'CREATE UNIQUE INDEX `uk_qu_dimension_item_v2_tmp` ON `el_qu` (`dimension_id`,`competency_question_type`,`dimension_item_no`)',
  'SELECT ''replacement question item index not required''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @index_exists FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_qu' AND INDEX_NAME='uk_qu_dimension_item';
SET @sql=IF(
  @index_exists>0
    AND NOT (@qu_dimension_item_columns='dimension_id,competency_question_type,dimension_item_no' AND @qu_dimension_item_non_unique=0),
  'DROP INDEX `uk_qu_dimension_item` ON `el_qu`',
  'SELECT ''uk_qu_dimension_item signature is current or absent''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @index_exists FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_qu' AND INDEX_NAME='uk_qu_dimension_item';
SELECT COUNT(*) INTO @replacement_index_exists FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_qu' AND INDEX_NAME='uk_qu_dimension_item_v2_tmp';
SET @sql=IF(@index_exists=0 AND @replacement_index_exists>0,
  'ALTER TABLE `el_qu` RENAME INDEX `uk_qu_dimension_item_v2_tmp` TO `uk_qu_dimension_item`',
  'SELECT ''question item index rename not required''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @index_exists FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_qu' AND INDEX_NAME='uk_qu_dimension_item';
SET @sql=IF(@index_exists=0,
  'CREATE UNIQUE INDEX `uk_qu_dimension_item` ON `el_qu` (`dimension_id`,`competency_question_type`,`dimension_item_no`)',
  'SELECT ''uk_qu_dimension_item exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @replacement_index_exists FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_qu' AND INDEX_NAME='uk_qu_dimension_item_v2_tmp';
SET @sql=IF(@replacement_index_exists>0,
  'DROP INDEX `uk_qu_dimension_item_v2_tmp` ON `el_qu`',
  'SELECT ''replacement question item index absent''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @index_exists FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_qu' AND INDEX_NAME='idx_qu_competency_type_status';
SET @sql=IF(@index_exists=0,
  'CREATE INDEX `idx_qu_competency_type_status` ON `el_qu` (`competency_question_type`,`question_status`,`dimension_id`)',
  'SELECT ''idx_qu_competency_type_status exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 2. 发布题型快照。当前切片仅新增承载字段，不改变现有发布/计分流程的必填契约。
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam_competency_question' AND COLUMN_NAME='competency_question_type';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_exam_competency_question` ADD COLUMN `competency_question_type` varchar(32) DEFAULT NULL COMMENT ''发布快照题型：dimension/validity'' AFTER `question_code`',
  'SELECT ''el_exam_competency_question.competency_question_type exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

UPDATE `el_exam_competency_question`
SET `competency_question_type`='dimension'
WHERE `id` IS NOT NULL
  AND `competency_question_type` IS NULL;

SELECT COUNT(*) INTO @index_exists FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam_competency_question' AND INDEX_NAME='idx_exam_competency_question_type';
SET @sql=IF(@index_exists=0,
  'CREATE INDEX `idx_exam_competency_question_type` ON `el_exam_competency_question` (`exam_id`,`competency_question_type`,`snapshot_order`)',
  'SELECT ''idx_exam_competency_question_type exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 3. 一级维度发布快照。这里只提供通用分组容器，不写死一期名称或维度映射。
CREATE TABLE IF NOT EXISTS `el_exam_competency_group` (
  `id` varchar(64) NOT NULL,
  `exam_id` varchar(64) NOT NULL,
  `group_code` varchar(32) NOT NULL,
  `group_name` varchar(100) NOT NULL,
  `display_order` int NOT NULL,
  `dimension_count` int NOT NULL DEFAULT 0,
  `question_count` int NOT NULL DEFAULT 0,
  `create_time` datetime DEFAULT NULL,
  `snapshot_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_exam_competency_group_code` (`exam_id`,`group_code`),
  UNIQUE KEY `uk_exam_competency_group_order` (`exam_id`,`display_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='测评一级维度发布快照';

SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam_competency_dimension' AND COLUMN_NAME='group_id';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_exam_competency_dimension` ADD COLUMN `group_id` varchar(64) DEFAULT NULL COMMENT ''一级维度发布快照ID'' AFTER `exam_id`',
  'SELECT ''el_exam_competency_dimension.group_id exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 重跑时先移除 008 自身约束，避免 MySQL 拒绝对外键列重复执行字符集对齐；末尾会完整恢复。
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_ecdim_group';
SET @sql=IF(@fk_exists>0,
  'ALTER TABLE `el_exam_competency_dimension` DROP FOREIGN KEY `fk_ecdim_group`',
  'SELECT ''fk_ecdim_group absent before alignment''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_ecg_exam';
SET @sql=IF(@fk_exists>0,
  'ALTER TABLE `el_exam_competency_group` DROP FOREIGN KEY `fk_ecg_exam`',
  'SELECT ''fk_ecg_exam absent before alignment''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT CHARACTER_SET_NAME,COLLATION_NAME INTO @exam_cs,@exam_co
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam' AND COLUMN_NAME='id';
SELECT CHARACTER_SET_NAME,COLLATION_NAME INTO @exam_group_cs,@exam_group_co
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam_competency_group' AND COLUMN_NAME='id';
SET @sql=CONCAT('ALTER TABLE `el_exam_competency_group` MODIFY COLUMN `exam_id` varchar(64) CHARACTER SET ',@exam_cs,' COLLATE ',@exam_co,' NOT NULL');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql=CONCAT('ALTER TABLE `el_exam_competency_dimension` MODIFY COLUMN `group_id` varchar(64) CHARACTER SET ',@exam_group_cs,' COLLATE ',@exam_group_co,' DEFAULT NULL');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @index_exists FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam_competency_dimension' AND INDEX_NAME='idx_exam_competency_dimension_group';
SET @sql=IF(@index_exists=0,
  'CREATE INDEX `idx_exam_competency_dimension_group` ON `el_exam_competency_dimension` (`exam_id`,`group_id`,`display_order`)',
  'SELECT ''idx_exam_competency_dimension_group exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 4. 一级维度结果。score/level 均可空，避免在客户确认前套用现有四档或假定五档边界。
CREATE TABLE IF NOT EXISTS `el_competency_group_result` (
  `id` varchar(64) NOT NULL,
  `paper_id` varchar(64) NOT NULL,
  `exam_group_id` varchar(64) NOT NULL,
  `group_code` varchar(32) NOT NULL,
  `group_name` varchar(100) NOT NULL,
  `display_order` int NOT NULL,
  `total_dimension_count` int NOT NULL DEFAULT 0,
  `effective_dimension_count` int NOT NULL DEFAULT 0,
  `total_question_count` int NOT NULL DEFAULT 0,
  `answered_question_count` int NOT NULL DEFAULT 0,
  `group_score` decimal(18,6) DEFAULT NULL,
  `level_code` varchar(32) DEFAULT NULL,
  `is_complete` tinyint NOT NULL DEFAULT 0,
  `scoring_version` varchar(32) NOT NULL DEFAULT '',
  `create_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_competency_group_result` (`paper_id`,`exam_group_id`),
  KEY `idx_competency_group_result_score` (`exam_group_id`,`group_score`,`paper_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力一级维度结果';

-- 5. 效度结果。score/status 均可空，迁移不写入方向、阈值或状态结论。
CREATE TABLE IF NOT EXISTS `el_competency_validity_result` (
  `paper_id` varchar(64) NOT NULL,
  `total_question_count` int NOT NULL DEFAULT 0,
  `answered_question_count` int NOT NULL DEFAULT 0,
  `validity_score` decimal(18,6) DEFAULT NULL,
  `validity_status` varchar(32) DEFAULT NULL,
  `is_complete` tinyint NOT NULL DEFAULT 0,
  `scoring_version` varchar(32) NOT NULL DEFAULT '',
  `create_time` datetime DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`paper_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力效度结果';

SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_cgr_paper';
SET @sql=IF(@fk_exists>0,
  'ALTER TABLE `el_competency_group_result` DROP FOREIGN KEY `fk_cgr_paper`',
  'SELECT ''fk_cgr_paper absent before alignment''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_cgr_exam_group';
SET @sql=IF(@fk_exists>0,
  'ALTER TABLE `el_competency_group_result` DROP FOREIGN KEY `fk_cgr_exam_group`',
  'SELECT ''fk_cgr_exam_group absent before alignment''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_cvr_paper';
SET @sql=IF(@fk_exists>0,
  'ALTER TABLE `el_competency_validity_result` DROP FOREIGN KEY `fk_cvr_paper`',
  'SELECT ''fk_cvr_paper absent before alignment''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT CHARACTER_SET_NAME,COLLATION_NAME INTO @paper_cs,@paper_co
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_paper' AND COLUMN_NAME='id';
SET @sql=CONCAT('ALTER TABLE `el_competency_group_result` MODIFY COLUMN `paper_id` varchar(64) CHARACTER SET ',@paper_cs,' COLLATE ',@paper_co,' NOT NULL, MODIFY COLUMN `exam_group_id` varchar(64) CHARACTER SET ',@exam_group_cs,' COLLATE ',@exam_group_co,' NOT NULL');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql=CONCAT('ALTER TABLE `el_competency_validity_result` MODIFY COLUMN `paper_id` varchar(64) CHARACTER SET ',@paper_cs,' COLLATE ',@paper_co,' NOT NULL');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 6. 整体结果保留维度题分母；总题数继续表示完整试卷题数。
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='dimension_question_count';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_competency_result` ADD COLUMN `dimension_question_count` int NOT NULL DEFAULT 0 AFTER `answered_question_count`',
  'SELECT ''el_competency_result.dimension_question_count exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='answered_dimension_question_count';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_competency_result` ADD COLUMN `answered_dimension_question_count` int NOT NULL DEFAULT 0 AFTER `dimension_question_count`',
  'SELECT ''el_competency_result.answered_dimension_question_count exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 一期超时未完成答卷不产生正式总体分，必须以 NULL 持久化而不是伪造范围外的 0 分。
ALTER TABLE `el_competency_result`
  MODIFY COLUMN `overall_score` decimal(18,6) DEFAULT NULL;

UPDATE `el_competency_result`
SET `dimension_question_count`=`total_question_count`,
    `answered_dimension_question_count`=`answered_question_count`
WHERE `paper_id` IS NOT NULL
  AND `product_version`='competency-generic-v1'
  AND `scoring_version`='competency-v1'
  AND (`dimension_question_count`<>`total_question_count`
       OR `answered_dimension_question_count`<>`answered_question_count`);

-- 7. 关系约束。全部使用 RESTRICT，与应用既有完整链删除顺序一致。
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_ecg_exam';
SET @sql=IF(@fk_exists=0,
  'ALTER TABLE `el_exam_competency_group` ADD CONSTRAINT `fk_ecg_exam` FOREIGN KEY (`exam_id`) REFERENCES `el_exam` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT',
  'SELECT ''fk_ecg_exam exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_ecdim_group';
SET @sql=IF(@fk_exists=0,
  'ALTER TABLE `el_exam_competency_dimension` ADD CONSTRAINT `fk_ecdim_group` FOREIGN KEY (`group_id`) REFERENCES `el_exam_competency_group` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT',
  'SELECT ''fk_ecdim_group exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_cgr_paper';
SET @sql=IF(@fk_exists=0,
  'ALTER TABLE `el_competency_group_result` ADD CONSTRAINT `fk_cgr_paper` FOREIGN KEY (`paper_id`) REFERENCES `el_paper` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT',
  'SELECT ''fk_cgr_paper exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_cgr_exam_group';
SET @sql=IF(@fk_exists=0,
  'ALTER TABLE `el_competency_group_result` ADD CONSTRAINT `fk_cgr_exam_group` FOREIGN KEY (`exam_group_id`) REFERENCES `el_exam_competency_group` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT',
  'SELECT ''fk_cgr_exam_group exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_cvr_paper';
SET @sql=IF(@fk_exists=0,
  'ALTER TABLE `el_competency_validity_result` ADD CONSTRAINT `fk_cvr_paper` FOREIGN KEY (`paper_id`) REFERENCES `el_paper` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT',
  'SELECT ''fk_cvr_paper exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 8. 只读验证。
SELECT TABLE_NAME FROM information_schema.TABLES
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN (
  'el_exam_competency_group','el_competency_group_result','el_competency_validity_result'
) ORDER BY TABLE_NAME;
SELECT TABLE_NAME,COLUMN_NAME,IS_NULLABLE,COLUMN_TYPE FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND (
  (TABLE_NAME='el_qu' AND COLUMN_NAME='competency_question_type') OR
  (TABLE_NAME='el_exam_competency_question' AND COLUMN_NAME='competency_question_type') OR
  (TABLE_NAME='el_exam_competency_dimension' AND COLUMN_NAME='group_id') OR
  (TABLE_NAME='el_competency_result' AND COLUMN_NAME IN ('dimension_question_count','answered_dimension_question_count'))
) ORDER BY TABLE_NAME,ORDINAL_POSITION;
SELECT CONSTRAINT_NAME FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME IN (
  'fk_ecg_exam','fk_ecdim_group','fk_cgr_paper','fk_cgr_exam_group','fk_cvr_paper'
) ORDER BY CONSTRAINT_NAME;
