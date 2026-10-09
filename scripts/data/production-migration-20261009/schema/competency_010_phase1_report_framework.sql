SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `el_competency_report_content_package` (
  `id` varchar(64) NOT NULL,
  `product_version` varchar(64) NOT NULL,
  `scoring_version` varchar(64) NOT NULL,
  `content_version` varchar(64) NOT NULL,
  `template_version` varchar(64) NOT NULL,
  `audience` varchar(32) NOT NULL,
  `approval_status` varchar(20) NOT NULL DEFAULT 'draft',
  `content_approved_by` varchar(100) NOT NULL DEFAULT '',
  `content_approved_at` datetime DEFAULT NULL,
  `psychometric_approved_by` varchar(100) NOT NULL DEFAULT '',
  `psychometric_approved_at` datetime DEFAULT NULL,
  `question_source_sha256` char(64) NOT NULL DEFAULT '',
  `content_source_sha256` char(64) NOT NULL DEFAULT '',
  `effective_environment` varchar(32) NOT NULL DEFAULT '',
  `disclaimer` varchar(1000) NOT NULL DEFAULT '',
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_competency_report_content_package` (`product_version`,`scoring_version`,`content_version`,`template_version`,`audience`),
  KEY `idx_competency_report_content_package_status` (`approval_status`,`effective_environment`,`content_version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

SET @exists = (SELECT COUNT(*) FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_report_content_package' AND CONSTRAINT_NAME='fk_report_content_package_version');
SET @sql = IF(@exists=0, 'SELECT 1', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- This migration intentionally inserts no package. Formal reporting remains blocked until
-- independently reviewed source hashes, approvers, timestamps and disclaimer are supplied.