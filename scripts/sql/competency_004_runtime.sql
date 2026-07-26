-- ============================================================
-- 胜任力测验迁移 004：发布快照、答题字段和结果表
-- 兼容 MySQL 5.7+；幂等；不删除或覆盖既有业务数据
-- ============================================================
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `el_exam_competency_question` (
  `id` varchar(64) NOT NULL,
  `exam_id` varchar(64) NOT NULL,
  `exam_dimension_id` varchar(64) NOT NULL,
  `source_qu_id` varchar(64) NOT NULL,
  `question_code` varchar(32) NOT NULL,
  `dimension_item_no` int NOT NULL,
  `question_content` text NOT NULL,
  `observation_point` varchar(255) NOT NULL,
  `scoring_direction` varchar(16) NOT NULL,
  `options_snapshot` text NOT NULL,
  `source_update_time` datetime DEFAULT NULL,
  `snapshot_order` int NOT NULL,
  `create_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_exam_competency_question_source` (`exam_id`,`source_qu_id`),
  UNIQUE KEY `uk_exam_competency_question_code` (`exam_id`,`question_code`),
  KEY `idx_exam_competency_question_dimension` (`exam_id`,`exam_dimension_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `el_competency_dimension_result` (
  `id` varchar(64) NOT NULL,
  `paper_id` varchar(64) NOT NULL,
  `exam_dimension_id` varchar(64) NOT NULL,
  `dimension_id` varchar(64) NOT NULL,
  `dimension_code` varchar(16) NOT NULL,
  `dimension_name` varchar(100) NOT NULL,
  `display_order` int NOT NULL,
  `total_question_count` int NOT NULL,
  `answered_question_count` int NOT NULL,
  `score_sum` int NOT NULL,
  `dimension_score` decimal(18,6) DEFAULT NULL,
  `level_code` varchar(16) DEFAULT NULL,
  `is_complete` tinyint NOT NULL,
  `create_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_competency_dimension_result` (`paper_id`,`exam_dimension_id`),
  KEY `idx_competency_dimension_score` (`exam_dimension_id`,`dimension_score`,`paper_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `el_competency_result` (
  `paper_id` varchar(64) NOT NULL,
  `exam_id` varchar(64) NOT NULL,
  `total_question_count` int NOT NULL,
  `answered_question_count` int NOT NULL,
  `effective_dimension_count` int NOT NULL,
  `overall_score` decimal(18,6) NOT NULL,
  `evaluation_average` decimal(18,6) DEFAULT NULL,
  `evaluation_level` varchar(16) DEFAULT NULL,
  `report_audience` varchar(32) NOT NULL,
  `is_complete` tinyint NOT NULL,
  `submit_type` varchar(16) NOT NULL,
  `scoring_version` varchar(32) NOT NULL,
  `submitted_at` datetime NOT NULL,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`paper_id`),
  KEY `idx_competency_result_exam_score` (`exam_id`,`overall_score`,`paper_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- el_paper_qu competency fields
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_paper_qu' AND COLUMN_NAME='exam_question_id';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_paper_qu` ADD COLUMN `exam_question_id` varchar(64) DEFAULT NULL COMMENT ''发布题目快照ID''','SELECT ''exam_question_id exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_paper_qu' AND COLUMN_NAME='raw_answer';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_paper_qu` ADD COLUMN `raw_answer` tinyint DEFAULT NULL COMMENT ''原始选择1-5''','SELECT ''raw_answer exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_paper_qu' AND COLUMN_NAME='final_score';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_paper_qu` ADD COLUMN `final_score` tinyint DEFAULT NULL COMMENT ''最终得分1-5''','SELECT ''final_score exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @index_exists FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_paper_qu' AND INDEX_NAME='uk_paper_exam_question';
SET @sql=IF(@index_exists=0,'CREATE UNIQUE INDEX `uk_paper_exam_question` ON `el_paper_qu` (`paper_id`,`exam_question_id`)','SELECT ''uk_paper_exam_question exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @index_exists FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_paper_qu' AND INDEX_NAME='idx_paper_question_answered';
SET @sql=IF(@index_exists=0,'CREATE INDEX `idx_paper_question_answered` ON `el_paper_qu` (`paper_id`,`answered`)','SELECT ''idx_paper_question_answered exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @index_exists FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_paper' AND INDEX_NAME='idx_paper_state_limit_time';
SET @sql=IF(@index_exists=0,'CREATE INDEX `idx_paper_state_limit_time` ON `el_paper` (`state`,`limit_time`)','SELECT ''idx_paper_state_limit_time exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Dynamically align every join key with its referenced existing column.
SELECT CHARACTER_SET_NAME, COLLATION_NAME INTO @exam_cs,@exam_co FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam' AND COLUMN_NAME='id';
SELECT CHARACTER_SET_NAME, COLLATION_NAME INTO @paper_cs,@paper_co FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_paper' AND COLUMN_NAME='id';
SELECT CHARACTER_SET_NAME, COLLATION_NAME INTO @qu_cs,@qu_co FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_qu' AND COLUMN_NAME='id';
SELECT CHARACTER_SET_NAME, COLLATION_NAME INTO @exam_dimension_cs,@exam_dimension_co FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam_competency_dimension' AND COLUMN_NAME='id';
SELECT CHARACTER_SET_NAME, COLLATION_NAME INTO @dimension_cs,@dimension_co FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_dimension' AND COLUMN_NAME='id';
SELECT CHARACTER_SET_NAME, COLLATION_NAME INTO @exam_question_cs,@exam_question_co FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam_competency_question' AND COLUMN_NAME='id';
SET @sql=CONCAT('ALTER TABLE `el_exam_competency_question` MODIFY `exam_id` varchar(64) CHARACTER SET ',@exam_cs,' COLLATE ',@exam_co,' NOT NULL, MODIFY `exam_dimension_id` varchar(64) CHARACTER SET ',@exam_dimension_cs,' COLLATE ',@exam_dimension_co,' NOT NULL, MODIFY `source_qu_id` varchar(64) CHARACTER SET ',@qu_cs,' COLLATE ',@qu_co,' NOT NULL'); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql=CONCAT('ALTER TABLE `el_paper_qu` MODIFY `exam_question_id` varchar(64) CHARACTER SET ',@exam_question_cs,' COLLATE ',@exam_question_co,' DEFAULT NULL'); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql=CONCAT('ALTER TABLE `el_competency_dimension_result` MODIFY `paper_id` varchar(64) CHARACTER SET ',@paper_cs,' COLLATE ',@paper_co,' NOT NULL, MODIFY `exam_dimension_id` varchar(64) CHARACTER SET ',@exam_dimension_cs,' COLLATE ',@exam_dimension_co,' NOT NULL, MODIFY `dimension_id` varchar(64) CHARACTER SET ',@dimension_cs,' COLLATE ',@dimension_co,' NOT NULL'); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql=CONCAT('ALTER TABLE `el_competency_result` MODIFY `paper_id` varchar(64) CHARACTER SET ',@paper_cs,' COLLATE ',@paper_co,' NOT NULL, MODIFY `exam_id` varchar(64) CHARACTER SET ',@exam_cs,' COLLATE ',@exam_co,' NOT NULL'); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN ('el_exam_competency_question','el_competency_dimension_result','el_competency_result') ORDER BY TABLE_NAME;
