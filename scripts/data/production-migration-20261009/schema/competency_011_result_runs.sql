-- ============================================================
-- 胜任力测验迁移 011：版本化评分运行及并行结果表
-- 兼容 MySQL 5.7+；幂等；不修改、不回填旧结果表
-- ============================================================
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `el_competency_result_run` (
  `id` varchar(64) NOT NULL,
  `paper_id` varchar(64) NOT NULL,
  `exam_id` varchar(64) NOT NULL,
  `product_version` varchar(32) NOT NULL,
  `scoring_version` varchar(32) NOT NULL,
  `content_version` varchar(32) NOT NULL,
  `report_template_version` varchar(32) NOT NULL,
  `report_audience` varchar(32) NOT NULL,
  `participant_type` varchar(16) NOT NULL,
  `participant_id` varchar(64) NOT NULL,
  `participant_name` varchar(100) NOT NULL DEFAULT '',
  `participant_telephone` varchar(32) NOT NULL DEFAULT '',
  `participant_age` int DEFAULT NULL,
  `participant_gender` varchar(16) NOT NULL DEFAULT '',
  `participant_affiliation` varchar(255) NOT NULL DEFAULT '',
  `participant_post` varchar(100) NOT NULL DEFAULT '',
  `participant_degree` varchar(100) NOT NULL DEFAULT '',
  `participant_major` varchar(100) NOT NULL DEFAULT '',
  `source` varchar(32) NOT NULL,
  `status` varchar(20) NOT NULL,
  `error_message` varchar(500) NOT NULL DEFAULT '',
  `created_by` bigint DEFAULT NULL,
  `completed_at` datetime DEFAULT NULL,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_competency_result_run_version` (`paper_id`,`scoring_version`),
  KEY `idx_competency_result_run_exam_status` (`exam_id`,`status`,`create_time`),
  KEY `idx_competency_result_run_participant` (`participant_type`,`participant_id`,`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力版本化评分运行';

CREATE TABLE IF NOT EXISTS `el_competency_result_run_overall` (
  `result_run_id` varchar(64) NOT NULL,
  `total_question_count` int NOT NULL DEFAULT 0,
  `answered_question_count` int NOT NULL DEFAULT 0,
  `dimension_question_count` int NOT NULL DEFAULT 0,
  `answered_dimension_question_count` int NOT NULL DEFAULT 0,
  `effective_dimension_count` int NOT NULL DEFAULT 0,
  `overall_score` decimal(18,6) DEFAULT NULL,
  `level_code` varchar(32) DEFAULT NULL,
  `norm_score` decimal(18,6) DEFAULT NULL,
  `norm_comparison_code` varchar(32) DEFAULT NULL,
  `is_complete` tinyint NOT NULL DEFAULT 0,
  `submit_type` varchar(16) NOT NULL DEFAULT '',
  `submitted_at` datetime DEFAULT NULL,
  `user_time` int NOT NULL DEFAULT 0,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`result_run_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力版本化总体结果';

CREATE TABLE IF NOT EXISTS `el_competency_result_run_module` (
  `id` varchar(64) NOT NULL,
  `result_run_id` varchar(64) NOT NULL,
  `module_id` varchar(64) NOT NULL,
  `module_code` varchar(32) NOT NULL,
  `module_name` varchar(100) NOT NULL,
  `display_order` int NOT NULL,
  `total_dimension_count` int NOT NULL DEFAULT 0,
  `effective_dimension_count` int NOT NULL DEFAULT 0,
  `module_score` decimal(18,6) DEFAULT NULL,
  `level_code` varchar(32) DEFAULT NULL,
  `norm_score` decimal(18,6) DEFAULT NULL,
  `norm_comparison_code` varchar(32) DEFAULT NULL,
  `is_complete` tinyint NOT NULL DEFAULT 0,
  `create_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_result_run_module` (`result_run_id`,`module_code`),
  UNIQUE KEY `uk_result_run_module_order` (`result_run_id`,`display_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力版本化模块结果';

CREATE TABLE IF NOT EXISTS `el_competency_result_run_dimension` (
  `id` varchar(64) NOT NULL,
  `result_run_id` varchar(64) NOT NULL,
  `dimension_id` varchar(64) NOT NULL,
  `dimension_code` varchar(32) NOT NULL,
  `dimension_name` varchar(100) NOT NULL,
  `display_order` int NOT NULL,
  `total_question_count` int NOT NULL DEFAULT 0,
  `answered_question_count` int NOT NULL DEFAULT 0,
  `score_sum` int NOT NULL DEFAULT 0,
  `dimension_score` decimal(18,6) DEFAULT NULL,
  `level_code` varchar(32) DEFAULT NULL,
  `norm_score` decimal(18,6) DEFAULT NULL,
  `is_complete` tinyint NOT NULL DEFAULT 0,
  `create_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_result_run_dimension` (`result_run_id`,`dimension_id`),
  UNIQUE KEY `uk_result_run_dimension_order` (`result_run_id`,`display_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力版本化维度结果';

CREATE TABLE IF NOT EXISTS `el_competency_result_run_validity` (
  `result_run_id` varchar(64) NOT NULL,
  `total_question_count` int NOT NULL DEFAULT 0,
  `answered_question_count` int NOT NULL DEFAULT 0,
  `validity_score` decimal(18,6) DEFAULT NULL,
  `validity_status` varchar(32) DEFAULT NULL,
  `is_complete` tinyint NOT NULL DEFAULT 0,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`result_run_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力版本化效度结果';

-- 动态对齐外键列字符集与排序规则。
SELECT CHARACTER_SET_NAME,COLLATION_NAME INTO @paper_cs,@paper_co
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_paper' AND COLUMN_NAME='id';
SELECT CHARACTER_SET_NAME,COLLATION_NAME INTO @exam_cs,@exam_co
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam' AND COLUMN_NAME='id';
SELECT CHARACTER_SET_NAME,COLLATION_NAME INTO @participant_cs,@participant_co
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_candidate' AND COLUMN_NAME='id';
SELECT CHARACTER_SET_NAME,COLLATION_NAME INTO @run_cs,@run_co
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result_run' AND COLUMN_NAME='id';

-- 重跑时仅暂时移除011自身的外键，以允许重复执行字符集对齐；末尾完整恢复。
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_result_run_paper';
SET @sql=IF(@fk_exists>0,'ALTER TABLE `el_competency_result_run` DROP FOREIGN KEY `fk_result_run_paper`','SELECT ''fk_result_run_paper absent before alignment''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_result_run_exam';
SET @sql=IF(@fk_exists>0,'ALTER TABLE `el_competency_result_run` DROP FOREIGN KEY `fk_result_run_exam`','SELECT ''fk_result_run_exam absent before alignment''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 013 may already reference (id,paper_id). Its FK could only have been created
-- after paper_id was compatible, so do not alter that constrained column again.
SELECT COUNT(*) INTO @report_run_fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_competency_report_result_run';
SET @sql=IF(@report_run_fk_exists>0,
  'SELECT ''result_run paper_id alignment retained by 013 foreign key''',
  CONCAT('ALTER TABLE `el_competency_result_run` MODIFY `paper_id` varchar(64) CHARACTER SET ',@paper_cs,' COLLATE ',@paper_co,' NOT NULL'));
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql=CONCAT('ALTER TABLE `el_competency_result_run` MODIFY `exam_id` varchar(64) CHARACTER SET ',@exam_cs,' COLLATE ',@exam_co,' NOT NULL, MODIFY `participant_id` varchar(64) CHARACTER SET ',@participant_cs,' COLLATE ',@participant_co,' NOT NULL');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_result_run_overall';
SET @sql=IF(@fk_exists>0,'ALTER TABLE `el_competency_result_run_overall` DROP FOREIGN KEY `fk_result_run_overall`','SELECT ''fk_result_run_overall absent before alignment''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_result_run_module';
SET @sql=IF(@fk_exists>0,'ALTER TABLE `el_competency_result_run_module` DROP FOREIGN KEY `fk_result_run_module`','SELECT ''fk_result_run_module absent before alignment''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_result_run_dimension';
SET @sql=IF(@fk_exists>0,'ALTER TABLE `el_competency_result_run_dimension` DROP FOREIGN KEY `fk_result_run_dimension`','SELECT ''fk_result_run_dimension absent before alignment''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_result_run_validity';
SET @sql=IF(@fk_exists>0,'ALTER TABLE `el_competency_result_run_validity` DROP FOREIGN KEY `fk_result_run_validity`','SELECT ''fk_result_run_validity absent before alignment''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql=CONCAT('ALTER TABLE `el_competency_result_run_overall` MODIFY `result_run_id` varchar(64) CHARACTER SET ',@run_cs,' COLLATE ',@run_co,' NOT NULL');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql=CONCAT('ALTER TABLE `el_competency_result_run_module` MODIFY `result_run_id` varchar(64) CHARACTER SET ',@run_cs,' COLLATE ',@run_co,' NOT NULL');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql=CONCAT('ALTER TABLE `el_competency_result_run_dimension` MODIFY `result_run_id` varchar(64) CHARACTER SET ',@run_cs,' COLLATE ',@run_co,' NOT NULL');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql=CONCAT('ALTER TABLE `el_competency_result_run_validity` MODIFY `result_run_id` varchar(64) CHARACTER SET ',@run_cs,' COLLATE ',@run_co,' NOT NULL');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 只增加011表之间及其到paper/exam的RESTRICT关系。
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_result_run_paper';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_competency_result_run` ADD CONSTRAINT `fk_result_run_paper` FOREIGN KEY (`paper_id`) REFERENCES `el_paper` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_result_run_paper exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_result_run_exam';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_competency_result_run` ADD CONSTRAINT `fk_result_run_exam` FOREIGN KEY (`exam_id`) REFERENCES `el_exam` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_result_run_exam exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_result_run_overall';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_competency_result_run_overall` ADD CONSTRAINT `fk_result_run_overall` FOREIGN KEY (`result_run_id`) REFERENCES `el_competency_result_run` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_result_run_overall exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_result_run_module';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_competency_result_run_module` ADD CONSTRAINT `fk_result_run_module` FOREIGN KEY (`result_run_id`) REFERENCES `el_competency_result_run` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_result_run_module exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_result_run_dimension';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_competency_result_run_dimension` ADD CONSTRAINT `fk_result_run_dimension` FOREIGN KEY (`result_run_id`) REFERENCES `el_competency_result_run` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_result_run_dimension exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_result_run_validity';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_competency_result_run_validity` ADD CONSTRAINT `fk_result_run_validity` FOREIGN KEY (`result_run_id`) REFERENCES `el_competency_result_run` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_result_run_validity exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 只读结构验证；迁移本身不创建任何评分运行数据。
SELECT TABLE_NAME
FROM information_schema.TABLES
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN (
  'el_competency_result_run','el_competency_result_run_overall',
  'el_competency_result_run_module','el_competency_result_run_dimension',
  'el_competency_result_run_validity'
)
ORDER BY TABLE_NAME;
SELECT CONSTRAINT_NAME
FROM information_schema.REFERENTIAL_CONSTRAINTS
WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME IN (
  'fk_result_run_paper','fk_result_run_exam','fk_result_run_overall',
  'fk_result_run_module','fk_result_run_dimension','fk_result_run_validity'
)
ORDER BY CONSTRAINT_NAME;
