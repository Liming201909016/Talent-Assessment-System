-- Local artifact ONLY: installation requires separate migration/release approval.
-- Independent archival PDFs; no old ALTER/DML, no backfill or current mutation.
SET NAMES utf8mb4;
SET @mng_ri_run=(SELECT collation_name FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='el_mng_result_run' AND column_name='id' AND column_type='varchar(64)' AND is_nullable='NO' AND character_set_name='utf8mb4');
SET @mng_ri_paper=(SELECT collation_name FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='el_mng_result_run' AND column_name='paper_id' AND column_type='varchar(64)' AND is_nullable='NO' AND character_set_name='utf8mb4');
SET @mng_ri_exam=(SELECT collation_name FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='el_mng_result_run' AND column_name='exam_id' AND column_type='varchar(64)' AND is_nullable='NO' AND character_set_name='utf8mb4');
SET @mng_ri_sql=IF(@mng_ri_run REGEXP '^utf8mb4_[a-z0-9_]+$' AND @mng_ri_paper REGEXP '^utf8mb4_[a-z0-9_]+$' AND @mng_ri_exam REGEXP '^utf8mb4_[a-z0-9_]+$','SELECT ''MNG_REISSUE_PARENT_OK''','SELECT * FROM information_schema.MNG_REISSUE_PARENT_MISMATCH');
PREPARE mng_ri_stmt FROM @mng_ri_sql; EXECUTE mng_ri_stmt; DEALLOCATE PREPARE mng_ri_stmt;
SET @mng_ri_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_report_reissue (
id varchar(64) NOT NULL,
run_id varchar(64) CHARACTER SET utf8mb4 COLLATE ',@mng_ri_run,' NOT NULL,
paper_id varchar(64) CHARACTER SET utf8mb4 COLLATE ',@mng_ri_paper,' NOT NULL,
exam_id varchar(64) CHARACTER SET utf8mb4 COLLATE ',@mng_ri_exam,' NOT NULL,
kind varchar(32) NOT NULL,status varchar(16) NOT NULL,
data_snapshot longtext NOT NULL,data_sha char(64) NOT NULL,
template_sha char(64) NOT NULL,content_sha char(64) NOT NULL,
file_key varchar(255) NOT NULL,file_sha char(64) NOT NULL,file_bytes bigint NOT NULL,
created_by bigint NOT NULL,created_at datetime(6) NOT NULL,
PRIMARY KEY(id),UNIQUE KEY uk_mng_reissue_identity(id,paper_id,exam_id),
UNIQUE KEY uk_mng_reissue_input(run_id,data_sha,template_sha,content_sha),
KEY idx_mng_reissue_paper(paper_id,created_at),KEY idx_mng_reissue_run(run_id,paper_id,exam_id),
CONSTRAINT fk_mng_reissue_run FOREIGN KEY(run_id,paper_id,exam_id) REFERENCES el_mng_result_run(id,paper_id,exam_id) ON UPDATE RESTRICT ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin');
PREPARE mng_ri_stmt FROM @mng_ri_sql; EXECUTE mng_ri_stmt; DEALLOCATE PREPARE mng_ri_stmt;
SET @mng_ri_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_reissue_audit (
id varchar(64) NOT NULL,report_id varchar(64) NOT NULL,
paper_id varchar(64) CHARACTER SET utf8mb4 COLLATE ',@mng_ri_paper,' NOT NULL,
exam_id varchar(64) CHARACTER SET utf8mb4 COLLATE ',@mng_ri_exam,' NOT NULL,
actor_id bigint NOT NULL,action varchar(16) NOT NULL,created_at datetime(6) NOT NULL,
PRIMARY KEY(id),KEY idx_mng_reissue_audit(report_id,created_at),
KEY idx_mng_reissue_audit_identity(report_id,paper_id,exam_id),
CONSTRAINT fk_mng_reissue_audit FOREIGN KEY(report_id,paper_id,exam_id) REFERENCES el_mng_report_reissue(id,paper_id,exam_id) ON UPDATE RESTRICT ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin');
PREPARE mng_ri_stmt FROM @mng_ri_sql; EXECUTE mng_ri_stmt; DEALLOCATE PREPARE mng_ri_stmt;
-- CREATE IF NOT EXISTS never repairs drift. The API checks every column,
-- ordered index and composite FK before access, including on repeat install.
-- Retain records and PDFs on binary rollback; never drop evidence or old files.