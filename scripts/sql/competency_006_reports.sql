SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `el_competency_report_text` (
  `id` varchar(100) NOT NULL,
  `content_version` varchar(32) NOT NULL,
  `audience` varchar(32) NOT NULL,
  `content_type` varchar(20) NOT NULL,
  `dimension_id` varchar(64) NOT NULL DEFAULT '',
  `level_code` varchar(20) NOT NULL,
  `content` text NOT NULL,
  `disclaimer` varchar(500) NOT NULL DEFAULT '',
  `is_temporary` tinyint NOT NULL DEFAULT 1,
  `status` tinyint NOT NULL DEFAULT 0,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_competency_report_text_match` (`content_version`,`audience`,`content_type`,`dimension_id`,`level_code`),
  KEY `idx_competency_report_text_active` (`content_version`,`status`,`audience`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `el_competency_report` (
  `id` varchar(64) NOT NULL,
  `paper_id` varchar(64) NOT NULL,
  `exam_id` varchar(64) NOT NULL,
  `audience` varchar(32) NOT NULL,
  `content_version` varchar(32) NOT NULL,
  `text_snapshot` longtext NOT NULL,
  `score_snapshot` longtext NOT NULL,
  `pdf_path` varchar(500) NOT NULL DEFAULT '',
  `pdf_sha256` char(64) NOT NULL DEFAULT '',
  `pdf_size` bigint NOT NULL DEFAULT 0,
  `status` varchar(20) NOT NULL DEFAULT 'generating',
  `error_message` varchar(500) NOT NULL DEFAULT '',
  `generated_by` bigint DEFAULT NULL,
  `generated_at` datetime DEFAULT NULL,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_competency_report_paper_version` (`paper_id`,`content_version`),
  KEY `idx_competency_report_exam_status` (`exam_id`,`status`,`generated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `el_competency_report_audit` (
  `id` varchar(64) NOT NULL,
  `report_id` varchar(64) NOT NULL,
  `paper_id` varchar(64) NOT NULL,
  `action` varchar(20) NOT NULL,
  `operator_id` bigint DEFAULT NULL,
  `status` tinyint NOT NULL DEFAULT 1,
  `error_message` varchar(500) NOT NULL DEFAULT '',
  `client_ip` varchar(64) NOT NULL DEFAULT '',
  `create_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_competency_report_audit_paper` (`paper_id`,`create_time`),
  KEY `idx_competency_report_audit_report` (`report_id`,`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

SELECT CHARACTER_SET_NAME, COLLATION_NAME INTO @paper_cs, @paper_co
FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_paper' AND COLUMN_NAME='id';
SELECT CHARACTER_SET_NAME, COLLATION_NAME INTO @exam_cs, @exam_co
FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam' AND COLUMN_NAME='id';
SET @sql = CONCAT('ALTER TABLE `el_competency_report` MODIFY `paper_id` varchar(64) CHARACTER SET ', @paper_cs, ' COLLATE ', @paper_co, ' NOT NULL');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = CONCAT('ALTER TABLE `el_competency_report` MODIFY `exam_id` varchar(64) CHARACTER SET ', @exam_cs, ' COLLATE ', @exam_co, ' NOT NULL');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = CONCAT('ALTER TABLE `el_competency_report_audit` MODIFY `paper_id` varchar(64) CHARACTER SET ', @paper_cs, ' COLLATE ', @paper_co, ' NOT NULL');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @exists = (SELECT COUNT(*) FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND CONSTRAINT_NAME='fk_competency_report_paper');
SET @sql = IF(@exists=0, 'ALTER TABLE `el_competency_report` ADD CONSTRAINT `fk_competency_report_paper` FOREIGN KEY (`paper_id`) REFERENCES `el_paper` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @exists = (SELECT COUNT(*) FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report' AND CONSTRAINT_NAME='fk_competency_report_exam');
SET @sql = IF(@exists=0, 'ALTER TABLE `el_competency_report` ADD CONSTRAINT `fk_competency_report_exam` FOREIGN KEY (`exam_id`) REFERENCES `el_exam` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @exists = (SELECT COUNT(*) FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report_audit' AND CONSTRAINT_NAME='fk_competency_report_audit_report');
SET @sql = IF(@exists=0, 'ALTER TABLE `el_competency_report_audit` ADD CONSTRAINT `fk_competency_report_audit_report` FOREIGN KEY (`report_id`) REFERENCES `el_competency_report` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @exists = (SELECT COUNT(*) FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report_audit' AND CONSTRAINT_NAME='fk_competency_report_audit_paper');
SET @sql = IF(@exists=0, 'ALTER TABLE `el_competency_report_audit` ADD CONSTRAINT `fk_competency_report_audit_paper` FOREIGN KEY (`paper_id`) REFERENCES `el_paper` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

INSERT INTO `el_competency_report_text`
(`id`,`content_version`,`audience`,`content_type`,`dimension_id`,`level_code`,`content`,`disclaimer`,`is_temporary`,`status`,`create_time`,`update_time`)
SELECT CONCAT('temp-v1-',a.code,'-overall-',l.code), 'temp-v1', a.audience, 'overall', '', l.code,
       CONCAT('【临时测试文案】本次总体评价等级为“',l.label,'”。该文字仅验证报告版本、匹配和渲染流程。'),
       '临时测试文案，仅用于系统功能验证，不可作为人才决策依据。',1,0,NOW(),NOW()
FROM (SELECT 'frontline_employee' audience,'frontline' code UNION ALL SELECT 'leader','leader') a
CROSS JOIN (SELECT 'low' code,'较低' label UNION ALL SELECT 'average','一般' UNION ALL SELECT 'good','良好' UNION ALL SELECT 'high','较高') l
WHERE 1=1
ON DUPLICATE KEY UPDATE `id`=VALUES(`id`);

INSERT INTO `el_competency_report_text`
(`id`,`content_version`,`audience`,`content_type`,`dimension_id`,`level_code`,`content`,`disclaimer`,`is_temporary`,`status`,`create_time`,`update_time`)
SELECT CONCAT('temp-v1-',a.code,'-',d.code,'-',l.code), 'temp-v1', a.audience, 'dimension', d.id, l.code,
       CONCAT('【临时测试文案】',d.code,' ',d.name,'当前等级为“',l.label,'”。本段仅用于验证维度文案匹配和PDF生成，不构成发展建议。'),
       '临时测试文案，仅用于系统功能验证，不可作为人才决策依据。',1,0,NOW(),NOW()
FROM (SELECT 'frontline_employee' audience,'frontline' code UNION ALL SELECT 'leader','leader') a
CROSS JOIN `el_competency_dimension` d
CROSS JOIN (SELECT 'low' code,'较低' label UNION ALL SELECT 'average','一般' UNION ALL SELECT 'good','良好' UNION ALL SELECT 'high','较高') l
WHERE 1=1
ON DUPLICATE KEY UPDATE `id`=VALUES(`id`);
