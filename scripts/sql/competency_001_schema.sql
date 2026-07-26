-- ============================================================
-- 胜任力测验迁移 001：测评类型、计分模式、报告对象与发布状态
-- 兼容：MySQL 5.7+
-- 特性：可重复执行；仅新增列和索引；不删除或覆盖现有业务数据
-- 说明：既有测评统一按 legacy + legacy 处理，并视为已发布。
--       新建胜任力测评由应用显式写 competency + competency_average + 0（草稿）。
--       胜任力还必须显式写 frontline_employee（基层员工版）或 leader（领导人员版）。
-- ============================================================

SET NAMES utf8mb4;

-- 1. assessment_type：显式测评类型
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_exam'
  AND COLUMN_NAME = 'assessment_type';
SET @sql = IF(
  @column_exists = 0,
  'ALTER TABLE `el_exam` ADD COLUMN `assessment_type` varchar(32) NOT NULL DEFAULT ''legacy'' COMMENT ''测评类型：legacy/competency''',
  'SELECT ''el_exam.assessment_type already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2. scoring_mode：显式计分模式
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_exam'
  AND COLUMN_NAME = 'scoring_mode';
SET @sql = IF(
  @column_exists = 0,
  'ALTER TABLE `el_exam` ADD COLUMN `scoring_mode` varchar(32) NOT NULL DEFAULT ''legacy'' COMMENT ''计分模式：legacy/competency_average''',
  'SELECT ''el_exam.scoring_mode already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 3. competency_report_audience：报告对象版本
-- 既有 legacy 测评保持 NULL；胜任力保存/发布时由应用强制校验非空。
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_exam'
  AND COLUMN_NAME = 'competency_report_audience';
SET @sql = IF(
  @column_exists = 0,
  'ALTER TABLE `el_exam` ADD COLUMN `competency_report_audience` varchar(32) DEFAULT NULL COMMENT ''胜任力报告对象：frontline_employee/leader''',
  'SELECT ''el_exam.competency_report_audience already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 4. publish_status：0=草稿，1=已发布
-- 默认 1 用于兼容既有测评和旧版 INSERT；应用创建胜任力测评时必须显式写 0。
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_exam'
  AND COLUMN_NAME = 'publish_status';
SET @sql = IF(
  @column_exists = 0,
  'ALTER TABLE `el_exam` ADD COLUMN `publish_status` tinyint NOT NULL DEFAULT 1 COMMENT ''发布状态：0草稿，1已发布''',
  'SELECT ''el_exam.publish_status already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 5. published_at：首次成功冻结时间
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_exam'
  AND COLUMN_NAME = 'published_at';
SET @sql = IF(
  @column_exists = 0,
  'ALTER TABLE `el_exam` ADD COLUMN `published_at` datetime DEFAULT NULL COMMENT ''首次成功冻结时间''',
  'SELECT ''el_exam.published_at already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 6. published_by：首次发布管理员
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_exam'
  AND COLUMN_NAME = 'published_by';
SET @sql = IF(
  @column_exists = 0,
  'ALTER TABLE `el_exam` ADD COLUMN `published_by` bigint DEFAULT NULL COMMENT ''首次发布管理员用户ID''',
  'SELECT ''el_exam.published_by already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 7. 类型 + 发布状态复合索引
SELECT COUNT(*) INTO @index_exists
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_exam'
  AND INDEX_NAME = 'idx_exam_assessment_publish';
SET @sql = IF(
  @index_exists = 0,
  'CREATE INDEX `idx_exam_assessment_publish` ON `el_exam` (`assessment_type`, `publish_status`)',
  'SELECT ''idx_exam_assessment_publish already exists'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 8. 只读验证：应返回 6 个字段和 1 个索引定义
SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_exam'
  AND COLUMN_NAME IN (
    'assessment_type',
    'scoring_mode',
    'competency_report_audience',
    'publish_status',
    'published_at',
    'published_by'
  )
ORDER BY ORDINAL_POSITION;

SELECT INDEX_NAME, SEQ_IN_INDEX, COLUMN_NAME
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_exam'
  AND INDEX_NAME = 'idx_exam_assessment_publish'
ORDER BY SEQ_IN_INDEX;

-- 9. 只读数据核对：既有记录必须保持合法组合
SELECT assessment_type, scoring_mode, competency_report_audience, publish_status, COUNT(*) AS row_count
FROM el_exam
GROUP BY assessment_type, scoring_mode, competency_report_audience, publish_status
ORDER BY assessment_type, scoring_mode, competency_report_audience, publish_status;
