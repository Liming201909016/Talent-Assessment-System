-- ============================================================
-- 胜任力测验迁移 007：产品、评分、内容与报告模板版本冻结
-- 兼容 MySQL 5.7+；幂等；不删除既有数据
-- 既有 competency 数据回填当前兼容版本；legacy 测评保持空版本。
-- ============================================================
SET NAMES utf8mb4;

-- el_exam：草稿配置与发布冻结版本
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam' AND COLUMN_NAME='competency_product_version';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_exam` ADD COLUMN `competency_product_version` varchar(32) NOT NULL DEFAULT '''' COMMENT ''胜任力产品版本''',
  'SELECT ''el_exam.competency_product_version exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam' AND COLUMN_NAME='competency_scoring_version';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_exam` ADD COLUMN `competency_scoring_version` varchar(32) NOT NULL DEFAULT '''' COMMENT ''胜任力评分版本''',
  'SELECT ''el_exam.competency_scoring_version exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam' AND COLUMN_NAME='competency_content_version';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_exam` ADD COLUMN `competency_content_version` varchar(32) NOT NULL DEFAULT '''' COMMENT ''胜任力内容版本''',
  'SELECT ''el_exam.competency_content_version exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam' AND COLUMN_NAME='competency_report_template_version';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_exam` ADD COLUMN `competency_report_template_version` varchar(32) NOT NULL DEFAULT '''' COMMENT ''胜任力报告模板版本''',
  'SELECT ''el_exam.competency_report_template_version exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- el_competency_result：提交时冻结完整版本集
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='product_version';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_competency_result` ADD COLUMN `product_version` varchar(32) NOT NULL DEFAULT '''' AFTER `submit_type`',
  'SELECT ''el_competency_result.product_version exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='content_version';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_competency_result` ADD COLUMN `content_version` varchar(32) NOT NULL DEFAULT '''' AFTER `scoring_version`',
  'SELECT ''el_competency_result.content_version exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='report_template_version';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_competency_result` ADD COLUMN `report_template_version` varchar(32) NOT NULL DEFAULT '''' AFTER `content_version`',
  'SELECT ''el_competency_result.report_template_version exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- el_competency_report：报告实例冻结实际模板版本
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND COLUMN_NAME='template_version';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_competency_report` ADD COLUMN `template_version` varchar(32) NOT NULL DEFAULT '''' AFTER `content_version`',
  'SELECT ''el_competency_report.template_version exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 报告实例唯一键必须与运行时精确匹配条件一致：paper + content + template。
-- 先建临时替代索引再删除旧二列索引，避免外键在切换瞬间失去 paper_id 左前缀索引。
SET @report_index_columns=(SELECT GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report'
    AND INDEX_NAME='uk_competency_report_paper_version');
SET @report_index_non_unique=(SELECT MIN(NON_UNIQUE)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report'
    AND INDEX_NAME='uk_competency_report_paper_version');
SELECT COUNT(*) INTO @replacement_index_exists FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report'
  AND INDEX_NAME='uk_competency_report_paper_content_template_tmp';
SET @sql=IF(
  NOT (@report_index_columns='paper_id,content_version,template_version' AND @report_index_non_unique=0)
    AND @replacement_index_exists=0,
  'CREATE UNIQUE INDEX `uk_competency_report_paper_content_template_tmp` ON `el_competency_report` (`paper_id`,`content_version`,`template_version`)',
  'SELECT ''replacement report version index not required''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @report_index_exists=(SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report'
    AND INDEX_NAME='uk_competency_report_paper_version');
SET @sql=IF(
  @report_index_exists>0
    AND NOT (@report_index_columns='paper_id,content_version,template_version' AND @report_index_non_unique=0),
  'ALTER TABLE `el_competency_report` DROP INDEX `uk_competency_report_paper_version`',
  'SELECT ''report version index already current''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @report_index_exists=(SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report'
    AND INDEX_NAME='uk_competency_report_paper_version');
SET @replacement_index_exists=(SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report'
    AND INDEX_NAME='uk_competency_report_paper_content_template_tmp');
SET @sql=IF(@report_index_exists=0 AND @replacement_index_exists>0,
  'ALTER TABLE `el_competency_report` RENAME INDEX `uk_competency_report_paper_content_template_tmp` TO `uk_competency_report_paper_version`',
  'SELECT ''report version index rename not required''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @report_index_exists=(SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report'
    AND INDEX_NAME='uk_competency_report_paper_version');
SET @sql=IF(@report_index_exists=0,
  'CREATE UNIQUE INDEX `uk_competency_report_paper_version` ON `el_competency_report` (`paper_id`,`content_version`,`template_version`)',
  'SELECT ''uk_competency_report_paper_version exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 兼容回填：主键条件显式出现，符合 safe-update 约束；不覆盖未来非空版本。
UPDATE `el_exam`
SET `competency_product_version`=IF(`competency_product_version`='', 'competency-generic-v1', `competency_product_version`),
    `competency_scoring_version`=IF(`competency_scoring_version`='', 'competency-v1', `competency_scoring_version`),
    `competency_content_version`=IF(`competency_content_version`='', 'temp-v1', `competency_content_version`),
    `competency_report_template_version`=IF(`competency_report_template_version`='', 'competency-report-v1', `competency_report_template_version`)
WHERE `id` IS NOT NULL
  AND `assessment_type`='competency'
  AND (`competency_product_version`='' OR `competency_scoring_version`='' OR `competency_content_version`='' OR `competency_report_template_version`='');

UPDATE `el_competency_result`
SET `product_version`=IF(`product_version`='', 'competency-generic-v1', `product_version`),
    `scoring_version`=IF(`scoring_version`='', 'competency-v1', `scoring_version`),
    `content_version`=IF(`content_version`='', 'temp-v1', `content_version`),
    `report_template_version`=IF(`report_template_version`='', 'competency-report-v1', `report_template_version`)
WHERE `paper_id` IS NOT NULL
  AND (`product_version`='' OR `scoring_version`='' OR `content_version`='' OR `report_template_version`='');

UPDATE `el_competency_report`
SET `template_version`='competency-report-v1'
WHERE `id` IS NOT NULL AND `template_version`='';

-- 只读验证
SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam'
  AND COLUMN_NAME IN ('competency_product_version','competency_scoring_version','competency_content_version','competency_report_template_version')
ORDER BY ORDINAL_POSITION;

SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result'
  AND COLUMN_NAME IN ('product_version','scoring_version','content_version','report_template_version')
ORDER BY ORDINAL_POSITION;

SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND COLUMN_NAME='template_version';

SELECT assessment_type, competency_product_version, competency_scoring_version,
       competency_content_version, competency_report_template_version, COUNT(*) AS row_count
FROM el_exam
GROUP BY assessment_type, competency_product_version, competency_scoring_version,
         competency_content_version, competency_report_template_version
ORDER BY assessment_type, competency_product_version;
