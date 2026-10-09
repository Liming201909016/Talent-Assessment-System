-- ============================================================
-- 胜任力测验迁移 013：报告绑定评分运行及当前展示指针
-- 兼容 MySQL 5.7+；幂等；不回填、不覆盖旧报告或PDF
-- ============================================================
SET NAMES utf8mb4;

-- 旧报告没有result_run，必须保持NULL；新报告写入准确运行ID。
SELECT COUNT(*) INTO @column_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND COLUMN_NAME='result_run_id';
SET @sql=IF(@column_exists=0,
  'ALTER TABLE `el_competency_report` ADD COLUMN `result_run_id` varchar(64) DEFAULT NULL AFTER `exam_id`',
  'SELECT ''el_competency_report.result_run_id exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 复合候选键供跨表复合外键验证paper/audience一致性。
SELECT COUNT(*) INTO @index_exists
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result_run' AND INDEX_NAME='uk_result_run_id_paper';
SET @sql=IF(@index_exists=0,
  'CREATE UNIQUE INDEX `uk_result_run_id_paper` ON `el_competency_result_run` (`id`,`paper_id`)',
  'SELECT ''uk_result_run_id_paper exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @index_exists
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND INDEX_NAME='uk_competency_report_id_paper_audience';
SET @sql=IF(@index_exists=0,
  'CREATE UNIQUE INDEX `uk_competency_report_id_paper_audience` ON `el_competency_report` (`id`,`paper_id`,`audience`)',
  'SELECT ''uk_competency_report_id_paper_audience exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @index_exists
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND INDEX_NAME='uk_competency_report_result_run_version';
SET @sql=IF(@index_exists=0,
  'CREATE UNIQUE INDEX `uk_competency_report_result_run_version` ON `el_competency_report` (`result_run_id`,`content_version`,`template_version`,`audience`)',
  'SELECT ''uk_competency_report_result_run_version exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS `el_competency_report_current` (
  `paper_id` varchar(64) NOT NULL,
  `audience` varchar(32) NOT NULL,
  `report_id` varchar(64) NOT NULL,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`paper_id`,`audience`),
  UNIQUE KEY `uk_competency_report_current_report` (`report_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力报告当前展示指针';

-- 动态对齐所有复合外键列字符集与排序规则。
SELECT CHARACTER_SET_NAME,COLLATION_NAME INTO @run_id_cs,@run_id_co
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result_run' AND COLUMN_NAME='id';
SELECT CHARACTER_SET_NAME,COLLATION_NAME INTO @report_id_cs,@report_id_co
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND COLUMN_NAME='id';
SELECT CHARACTER_SET_NAME,COLLATION_NAME INTO @report_paper_cs,@report_paper_co
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND COLUMN_NAME='paper_id';
SELECT CHARACTER_SET_NAME,COLLATION_NAME INTO @report_audience_cs,@report_audience_co
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND COLUMN_NAME='audience';

SELECT COUNT(*) INTO @report_run_fk_exists
FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND CONSTRAINT_NAME='fk_competency_report_result_run';
SET @sql=IF(@report_run_fk_exists=0,
  CONCAT('ALTER TABLE `el_competency_report` MODIFY `result_run_id` varchar(64) CHARACTER SET ',@run_id_cs,' COLLATE ',@run_id_co,' DEFAULT NULL'),
  'SELECT ''result_run_id alignment already constrained''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @current_report_fk_exists
FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report_current' AND CONSTRAINT_NAME='fk_competency_report_current_report';
SET @sql=IF(@current_report_fk_exists=0,
  CONCAT('ALTER TABLE `el_competency_report_current` MODIFY `paper_id` varchar(64) CHARACTER SET ',@report_paper_cs,' COLLATE ',@report_paper_co,' NOT NULL, MODIFY `audience` varchar(32) CHARACTER SET ',@report_audience_cs,' COLLATE ',@report_audience_co,' NOT NULL, MODIFY `report_id` varchar(64) CHARACTER SET ',@report_id_cs,' COLLATE ',@report_id_co,' NOT NULL'),
  'SELECT ''current report alignment already constrained''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 新报告的run和paper必须属于同一答卷；旧NULL绑定不受影响。
SELECT COUNT(*) INTO @fk_exists
FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND CONSTRAINT_NAME='fk_competency_report_result_run';
SET @sql=IF(@fk_exists=0,
  'ALTER TABLE `el_competency_report` ADD CONSTRAINT `fk_competency_report_result_run` FOREIGN KEY (`result_run_id`,`paper_id`) REFERENCES `el_competency_result_run` (`id`,`paper_id`) ON DELETE RESTRICT ON UPDATE RESTRICT',
  'SELECT ''fk_competency_report_result_run exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 当前指针必须指向同一paper及同一audience的报告。
SELECT COUNT(*) INTO @fk_exists
FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report_current' AND CONSTRAINT_NAME='fk_competency_report_current_report';
SET @sql=IF(@fk_exists=0,
  'ALTER TABLE `el_competency_report_current` ADD CONSTRAINT `fk_competency_report_current_report` FOREIGN KEY (`report_id`,`paper_id`,`audience`) REFERENCES `el_competency_report` (`id`,`paper_id`,`audience`) ON DELETE RESTRICT ON UPDATE RESTRICT',
  'SELECT ''fk_competency_report_current_report exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 只读验证；迁移不创建指针、不回填旧报告。
SELECT COLUMN_NAME,COLUMN_TYPE,IS_NULLABLE,COLUMN_DEFAULT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND COLUMN_NAME='result_run_id';
SELECT INDEX_NAME,GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX) AS columns_in_order,MIN(NON_UNIQUE) AS non_unique
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN ('el_competency_result_run','el_competency_report','el_competency_report_current')
  AND INDEX_NAME IN ('uk_result_run_id_paper','uk_competency_report_id_paper_audience','uk_competency_report_result_run_version','PRIMARY','uk_competency_report_current_report')
GROUP BY TABLE_NAME,INDEX_NAME
ORDER BY TABLE_NAME,INDEX_NAME;
SELECT CONSTRAINT_NAME
FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME IN ('fk_competency_report_result_run','fk_competency_report_current_report')
ORDER BY CONSTRAINT_NAME;
