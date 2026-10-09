-- 002 TEST sidecar only. MySQL 5.7 / 8.0; no legacy writes or historical backfill.
-- Apply only after backup and old-writer drain. CREATE IF NOT EXISTS is a no-op
-- on repeat; the application rejects drift instead of silently repairing it.
-- DDL auto-commits. On partial installation do NOT enable runtime; complete
-- installation and restart to invalidate the process's schema cache.
SET NAMES utf8mb4;
SET @mng_charset=(SELECT character_set_name FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='el_exam' AND column_name='id');
SET @mng_collation=(SELECT collation_name FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='el_exam' AND column_name='id');
SET @mng_parents=(SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name IN ('el_exam','el_paper','el_paper_qu') AND column_name='id' AND column_type='varchar(64)' AND is_nullable='NO' AND character_set_name=@mng_charset AND collation_name=@mng_collation);
SET @mng_sql=IF(@mng_charset='utf8mb4' AND @mng_collation REGEXP '^utf8mb4_[a-z0-9_]+$' AND @mng_parents=3,'SELECT ''MNG_PARENT_SIGNATURE_OK''','SELECT * FROM information_schema.MNG_PARENT_SIGNATURE_MISMATCH');
PREPARE mng_stmt FROM @mng_sql; EXECUTE mng_stmt; DEALLOCATE PREPARE mng_stmt;
-- If parent collations differ: stop, do not ALTER legacy columns; obtain a
-- reviewed per-column migration. This script never changes existing parents.

SET @mng_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_definition_bundle (
`id` varchar(64) NOT NULL,
`product_version` varchar(64) NOT NULL,
`question_version` varchar(64) NOT NULL,
`scoring_version` varchar(64) NOT NULL,
`norm_version` varchar(64) NOT NULL,
`questionnaire` varchar(255) NOT NULL,
`scoring_manifest` longtext NOT NULL,
`scoring_manifest_sha` char(64) NOT NULL,
`status` varchar(255) NOT NULL,
`created_at` datetime(6) NOT NULL,
PRIMARY KEY(id), UNIQUE KEY uk_mng_bundle_versions(product_version,question_version,scoring_version,norm_version)
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4 COLLATE=',@mng_collation);
PREPARE mng_stmt FROM @mng_sql; EXECUTE mng_stmt; DEALLOCATE PREPARE mng_stmt;

SET @mng_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_exam_profile (
`exam_id` varchar(64) NOT NULL,
`bundle_id` varchar(64) NOT NULL,
`mapping_snapshot` longtext NOT NULL,
`field_contract` longtext NOT NULL,
`mapping_sha` char(64) NOT NULL,
`total_time_minutes` bigint NOT NULL,
`frozen_at` datetime(6) NULL,
`created_at` datetime(6) NOT NULL,
PRIMARY KEY(exam_id), KEY idx_mng_profile_bundle(bundle_id),
CONSTRAINT fk_mng_profile_exam FOREIGN KEY(exam_id) REFERENCES el_exam(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
CONSTRAINT fk_mng_profile_bundle FOREIGN KEY(bundle_id) REFERENCES el_mng_definition_bundle(id) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4 COLLATE=',@mng_collation);
PREPARE mng_stmt FROM @mng_sql; EXECUTE mng_stmt; DEALLOCATE PREPARE mng_stmt;

SET @mng_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_paper_snapshot (
`paper_id` varchar(64) NOT NULL,
`exam_id` varchar(64) NOT NULL,
`profile_exam_id` varchar(64) NULL,
`bundle_id` varchar(64) NOT NULL,
`source` varchar(255) NOT NULL,
`evidence_snapshot` longtext NOT NULL,
`evidence_sha` char(64) NOT NULL,
`mapping_snapshot` longtext NOT NULL,
`mapping_sha` char(64) NOT NULL,
`scoring_manifest_sha` char(64) NOT NULL,
`participant_type` varchar(255) NOT NULL,
`participant_id` varchar(64) NOT NULL,
`participant_snapshot` longtext NOT NULL,
`field_contract` longtext NOT NULL,
`identity_source` varchar(255) NOT NULL,
`source_captured_at` datetime(6) NOT NULL,
`started_at` datetime(6) NOT NULL,
`limit_time` datetime(6) NULL,
`created_at` datetime(6) NOT NULL,
PRIMARY KEY(paper_id), UNIQUE KEY uk_mng_snapshot_paper_exam(paper_id,exam_id),
KEY idx_mng_snapshot_owner(participant_id,participant_type(16)), KEY idx_mng_snapshot_expiry(limit_time,paper_id),
CONSTRAINT fk_mng_snapshot_paper FOREIGN KEY(paper_id) REFERENCES el_paper(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
CONSTRAINT fk_mng_snapshot_exam FOREIGN KEY(exam_id) REFERENCES el_exam(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
CONSTRAINT fk_mng_snapshot_profile FOREIGN KEY(profile_exam_id) REFERENCES el_mng_exam_profile(exam_id) ON DELETE RESTRICT ON UPDATE RESTRICT,
CONSTRAINT fk_mng_snapshot_bundle FOREIGN KEY(bundle_id) REFERENCES el_mng_definition_bundle(id) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4 COLLATE=',@mng_collation);
PREPARE mng_stmt FROM @mng_sql; EXECUTE mng_stmt; DEALLOCATE PREPARE mng_stmt;

SET @mng_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_paper_question_snapshot (
`id` varchar(64) NOT NULL,
`paper_id` varchar(64) NOT NULL,
`paper_question_id` varchar(64) NOT NULL,
`source_question_id` varchar(64) NOT NULL,
`v_number` bigint NOT NULL,
`display_order` bigint NOT NULL,
`dimension_key` varchar(255) NOT NULL,
`reverse` tinyint(1) NOT NULL,
`content` longtext NOT NULL,
`options_snapshot` longtext NOT NULL,
`scoring_snapshot_sha` char(64) NOT NULL,
`selected_option_id` varchar(64) NULL,
`raw_answer` bigint NULL,
`final_score` bigint NULL,
`submitted_at` datetime(6) NULL,
`created_at` datetime(6) NOT NULL,
PRIMARY KEY(id), UNIQUE KEY uk_mng_question_v(paper_id,v_number), UNIQUE KEY uk_mng_question_pq(paper_id,paper_question_id), UNIQUE KEY uk_mng_question_order(paper_id,display_order),
CONSTRAINT fk_mng_question_paper FOREIGN KEY(paper_id) REFERENCES el_mng_paper_snapshot(paper_id) ON DELETE RESTRICT ON UPDATE RESTRICT,
CONSTRAINT fk_mng_question_pq FOREIGN KEY(paper_question_id) REFERENCES el_paper_qu(id) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4 COLLATE=',@mng_collation);
PREPARE mng_stmt FROM @mng_sql; EXECUTE mng_stmt; DEALLOCATE PREPARE mng_stmt;

SET @mng_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_result_run (
`id` varchar(64) NOT NULL,
`paper_id` varchar(64) NOT NULL,
`exam_id` varchar(64) NOT NULL,
`product_version` varchar(64) NOT NULL,
`question_version` varchar(64) NOT NULL,
`scoring_version` varchar(64) NOT NULL,
`norm_version` varchar(64) NOT NULL,
`questionnaire` varchar(255) NOT NULL,
`participant_type` varchar(255) NOT NULL,
`participant_id` varchar(64) NOT NULL,
`scoring_manifest_sha` char(64) NOT NULL,
`input_sha` char(64) NOT NULL,
`status` varchar(255) NOT NULL,
`source` varchar(255) NOT NULL,
`total_question_count` bigint NOT NULL,
`answered_question_count` bigint NOT NULL,
`overall_score` decimal(18,6) NULL,
`overall_norm` decimal(18,6) NULL,
`overall_level` varchar(255) NULL,
`user_time_seconds` bigint NULL,
`submitted_at` datetime(6) NULL,
`created_at` datetime(6) NOT NULL,
PRIMARY KEY(id), UNIQUE KEY uk_mng_run_versions(paper_id,scoring_version,norm_version), UNIQUE KEY uk_mng_run_identity(id,paper_id,exam_id), KEY idx_mng_run_exam(exam_id,submitted_at,id),
CONSTRAINT fk_mng_run_snapshot FOREIGN KEY(paper_id,exam_id) REFERENCES el_mng_paper_snapshot(paper_id,exam_id) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4 COLLATE=',@mng_collation);
PREPARE mng_stmt FROM @mng_sql; EXECUTE mng_stmt; DEALLOCATE PREPARE mng_stmt;

SET @mng_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_result_dimension (
`id` varchar(64) NOT NULL,
`run_id` varchar(64) NOT NULL,
`dimension_key` varchar(255) NOT NULL,
`display_order` bigint NOT NULL,
`dimension_name` varchar(255) NOT NULL,
`module_key` varchar(255) NOT NULL,
`question_count` bigint NOT NULL,
`answered_count` bigint NOT NULL,
`score_sum` bigint NOT NULL,
`score` decimal(18,6) NULL,
`norm` decimal(18,6) NULL,
`level` varchar(255) NULL,
`created_at` datetime(6) NOT NULL,
PRIMARY KEY(id), UNIQUE KEY uk_mng_dimension_key(run_id,dimension_key), UNIQUE KEY uk_mng_dimension_order(run_id,display_order),
CONSTRAINT fk_mng_dimension_run FOREIGN KEY(run_id) REFERENCES el_mng_result_run(id) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4 COLLATE=',@mng_collation);
PREPARE mng_stmt FROM @mng_sql; EXECUTE mng_stmt; DEALLOCATE PREPARE mng_stmt;

SET @mng_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_result_module (
`id` varchar(64) NOT NULL,
`run_id` varchar(64) NOT NULL,
`module_key` varchar(255) NOT NULL,
`display_order` bigint NOT NULL,
`dimension_count` bigint NOT NULL,
`score` decimal(18,6) NULL,
`created_at` datetime(6) NOT NULL,
PRIMARY KEY(id), UNIQUE KEY uk_mng_module_key(run_id,module_key), UNIQUE KEY uk_mng_module_order(run_id,display_order),
CONSTRAINT fk_mng_module_run FOREIGN KEY(run_id) REFERENCES el_mng_result_run(id) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4 COLLATE=',@mng_collation);
PREPARE mng_stmt FROM @mng_sql; EXECUTE mng_stmt; DEALLOCATE PREPARE mng_stmt;

SET @mng_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_runtime_receipt (
`run_id` varchar(64) NOT NULL,
`paper_id` varchar(64) NOT NULL,
`exam_id` varchar(64) NOT NULL,
`participant_type` varchar(16) NOT NULL,
`participant_id` varchar(64) NOT NULL,
`submit_type` varchar(16) NOT NULL,
`started_at` datetime(6) NOT NULL,
`limit_time` datetime(6) NOT NULL,
`submitted_at` datetime(6) NOT NULL,
`user_time_seconds` bigint NOT NULL,
PRIMARY KEY(run_id), UNIQUE KEY uk_mng_receipt_paper(paper_id),
CONSTRAINT fk_mng_receipt_run FOREIGN KEY(run_id,paper_id,exam_id) REFERENCES el_mng_result_run(id,paper_id,exam_id) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4 COLLATE=',@mng_collation);
PREPARE mng_stmt FROM @mng_sql; EXECUTE mng_stmt; DEALLOCATE PREPARE mng_stmt;

SET @mng_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_report_revision (
`id` varchar(64) NOT NULL,
`run_id` varchar(64) NOT NULL,
`paper_id` varchar(64) NOT NULL,
`exam_id` varchar(64) NOT NULL,
`revision` bigint NOT NULL,
`mode` varchar(16) NOT NULL,
`test_title` varchar(255) NOT NULL,
`test_label` varchar(255) NOT NULL,
`content_version` varchar(64) NOT NULL,
`content_sha` char(64) NOT NULL,
`template_version` varchar(64) NOT NULL,
`template_sha` char(64) NOT NULL,
`data_snapshot` longtext NOT NULL,
`data_sha` char(64) NOT NULL,
`file_key` varchar(255) NOT NULL,
`file_sha` char(64) NOT NULL,
`file_bytes` bigint NOT NULL,
`created_by` bigint NOT NULL,
`created_at` datetime(6) NOT NULL,
PRIMARY KEY(id), UNIQUE KEY uk_mng_report_identity(id,paper_id), UNIQUE KEY uk_mng_report_revision(run_id,revision),
CONSTRAINT fk_mng_report_run FOREIGN KEY(run_id,paper_id,exam_id) REFERENCES el_mng_result_run(id,paper_id,exam_id) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4 COLLATE=',@mng_collation);
PREPARE mng_stmt FROM @mng_sql; EXECUTE mng_stmt; DEALLOCATE PREPARE mng_stmt;

SET @mng_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_report_current (
`paper_id` varchar(64) NOT NULL,
`report_id` varchar(64) NOT NULL,
`updated_at` datetime(6) NOT NULL,
PRIMARY KEY(paper_id), UNIQUE KEY uk_mng_current_report(report_id),
CONSTRAINT fk_mng_current_report FOREIGN KEY(report_id,paper_id) REFERENCES el_mng_report_revision(id,paper_id) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4 COLLATE=',@mng_collation);
PREPARE mng_stmt FROM @mng_sql; EXECUTE mng_stmt; DEALLOCATE PREPARE mng_stmt;

SET @mng_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_report_audit (
`id` varchar(64) NOT NULL,
`report_id` varchar(64) NOT NULL,
`actor_id` bigint NOT NULL,
`action` varchar(16) NOT NULL,
`created_at` datetime(6) NOT NULL,
PRIMARY KEY(id), KEY idx_mng_audit_report(report_id,created_at),
CONSTRAINT fk_mng_audit_report FOREIGN KEY(report_id) REFERENCES el_mng_report_revision(id) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4 COLLATE=',@mng_collation);
PREPARE mng_stmt FROM @mng_sql; EXECUTE mng_stmt; DEALLOCATE PREPARE mng_stmt;
SELECT 'MNG_TEST_SIDECAR_DDL_COMPLETE_RESTART_REQUIRED';