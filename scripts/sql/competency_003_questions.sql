-- ============================================================
-- 胜任力测验迁移 003：题目主数据字段与索引
-- 兼容：MySQL 5.7+
-- 特性：可重复执行；仅新增列和索引；既有题目的胜任力字段保持 NULL
-- ============================================================

SET NAMES utf8mb4;

-- 1. question_code：胜任力题目全局稳定编号
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_qu'
  AND COLUMN_NAME = 'question_code';
SET @sql = IF(
  @column_exists = 0,
  'ALTER TABLE `el_qu` ADD COLUMN `question_code` varchar(32) DEFAULT NULL COMMENT ''胜任力题目编号，如D01-Q01''',
  'SELECT ''el_qu.question_code already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2. dimension_id：唯一所属胜任力维度
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_qu'
  AND COLUMN_NAME = 'dimension_id';
SET @sql = IF(
  @column_exists = 0,
  'ALTER TABLE `el_qu` ADD COLUMN `dimension_id` varchar(64) DEFAULT NULL COMMENT ''胜任力维度ID''',
  'SELECT ''el_qu.dimension_id already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 关联列必须继承维度主表 id 的字符集和排序规则，兼容 MySQL 5.7/8.0 环境。
SELECT CHARACTER_SET_NAME, COLLATION_NAME
INTO @dimension_id_charset, @dimension_id_collation
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_competency_dimension'
  AND COLUMN_NAME = 'id';
SET @sql = CONCAT(
  'ALTER TABLE `el_qu` MODIFY COLUMN `dimension_id` varchar(64) CHARACTER SET ',
  @dimension_id_charset,
  ' COLLATE ',
  @dimension_id_collation,
  ' DEFAULT NULL COMMENT ''胜任力维度ID'''
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 3. dimension_item_no：维度内题号
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_qu'
  AND COLUMN_NAME = 'dimension_item_no';
SET @sql = IF(
  @column_exists = 0,
  'ALTER TABLE `el_qu` ADD COLUMN `dimension_item_no` int DEFAULT NULL COMMENT ''胜任力维度内题号''',
  'SELECT ''el_qu.dimension_item_no already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 4. observation_point：考察点，仅用于说明和报告追溯
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_qu'
  AND COLUMN_NAME = 'observation_point';
SET @sql = IF(
  @column_exists = 0,
  'ALTER TABLE `el_qu` ADD COLUMN `observation_point` varchar(255) DEFAULT NULL COMMENT ''考察点，不参与计分''',
  'SELECT ''el_qu.observation_point already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 5. scoring_direction：forward/reverse
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_qu'
  AND COLUMN_NAME = 'scoring_direction';
SET @sql = IF(
  @column_exists = 0,
  'ALTER TABLE `el_qu` ADD COLUMN `scoring_direction` varchar(16) DEFAULT NULL COMMENT ''计分方向：forward/reverse''',
  'SELECT ''el_qu.scoring_direction already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 6. question_status：仅用于胜任力题启停；既有题不按此字段筛选
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_qu'
  AND COLUMN_NAME = 'question_status';
SET @sql = IF(
  @column_exists = 0,
  'ALTER TABLE `el_qu` ADD COLUMN `question_status` tinyint NOT NULL DEFAULT 0 COMMENT ''胜任力题状态：0启用1停用''',
  'SELECT ''el_qu.question_status already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 7. 题目编号唯一索引；MySQL 允许多个 NULL，既有题不受影响
SELECT COUNT(*) INTO @index_exists
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_qu'
  AND INDEX_NAME = 'uk_qu_question_code';
SET @sql = IF(
  @index_exists = 0,
  'CREATE UNIQUE INDEX `uk_qu_question_code` ON `el_qu` (`question_code`)',
  'SELECT ''uk_qu_question_code already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 8. 同一维度内题号唯一
SELECT COUNT(*) INTO @index_exists
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_qu'
  AND INDEX_NAME = 'uk_qu_dimension_item';
SET @sql = IF(
  @index_exists = 0,
  'CREATE UNIQUE INDEX `uk_qu_dimension_item` ON `el_qu` (`dimension_id`, `dimension_item_no`)',
  'SELECT ''uk_qu_dimension_item already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 9. 按维度和启停状态查询题目
SELECT COUNT(*) INTO @index_exists
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_qu'
  AND INDEX_NAME = 'idx_qu_dimension_status';
SET @sql = IF(
  @index_exists = 0,
  'CREATE INDEX `idx_qu_dimension_status` ON `el_qu` (`dimension_id`, `question_status`)',
  'SELECT ''idx_qu_dimension_status already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 10. 只读验证
SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, CHARACTER_SET_NAME, COLLATION_NAME
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_qu'
  AND COLUMN_NAME IN (
    'question_code',
    'dimension_id',
    'dimension_item_no',
    'observation_point',
    'scoring_direction',
    'question_status'
  )
ORDER BY ORDINAL_POSITION;

SELECT INDEX_NAME, SEQ_IN_INDEX, COLUMN_NAME, NON_UNIQUE
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_qu'
  AND INDEX_NAME IN ('uk_qu_question_code', 'uk_qu_dimension_item', 'idx_qu_dimension_status')
ORDER BY INDEX_NAME, SEQ_IN_INDEX;

SELECT COUNT(*) AS legacy_rows_with_competency_metadata
FROM el_qu
WHERE question_code IS NOT NULL
   OR dimension_id IS NOT NULL
   OR dimension_item_no IS NOT NULL
   OR observation_point IS NOT NULL
   OR scoring_direction IS NOT NULL;
