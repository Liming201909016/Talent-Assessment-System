-- Local artifact only. Separate staging migration approval required.
-- No old ALTER/DML, no history backfill, no AutoMigrate, no FK disabling.
-- DDL auto-commits: back up/drain writers before an approved installation.
SET NAMES utf8mb4;
SET @mng_draft_cs=(SELECT character_set_name FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='el_exam' AND column_name='id' AND column_type='varchar(64)' AND is_nullable='NO');
SET @mng_draft_coll=(SELECT collation_name FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='el_exam' AND column_name='id');
SET @mng_draft_sql=IF(@mng_draft_cs='utf8mb4' AND @mng_draft_coll REGEXP '^utf8mb4_[a-z0-9_]+$', 'SELECT ''MNG_DRAFT_PARENT_OK''','SELECT * FROM information_schema.MNG_DRAFT_PARENT_MISMATCH');
PREPARE mng_draft_stmt FROM @mng_draft_sql; EXECUTE mng_draft_stmt; DEALLOCATE PREPARE mng_draft_stmt;
-- Inherit only the actual parent identifier collation. Repo ID has no FK:
-- existing el_repo is utf8mb3; the application validates its canonical source.
SET @mng_draft_sql=CONCAT('CREATE TABLE IF NOT EXISTS el_mng_exam_draft (
`exam_id` varchar(64) CHARACTER SET utf8mb4 COLLATE ',@mng_draft_coll,' NOT NULL,
`repo_id` varchar(64) NOT NULL,
`repo_code` varchar(5) NOT NULL,
`lifecycle` varchar(6) NOT NULL,
`created_at` datetime(6) NOT NULL,
`updated_at` datetime(6) NOT NULL,
`frozen_at` datetime(6) NULL,
PRIMARY KEY(exam_id),
CONSTRAINT fk_mng_draft_exam FOREIGN KEY(exam_id) REFERENCES el_exam(id) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin');
PREPARE mng_draft_stmt FROM @mng_draft_sql; EXECUTE mng_draft_stmt; DEALLOCATE PREPARE mng_draft_stmt;
-- CREATE IF NOT EXISTS never repairs drift. Runtime validates every column,
-- ordered primary index and RESTRICT FK before first new002 DML.
SET @mng_draft_columns=(SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='el_mng_exam_draft');
SET @mng_draft_columns_ok=(SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='el_mng_exam_draft' AND (
 (column_name='exam_id' AND column_type='varchar(64)' AND is_nullable='NO' AND character_set_name=@mng_draft_cs AND collation_name=@mng_draft_coll) OR
 (column_name='repo_id' AND column_type='varchar(64)' AND is_nullable='NO' AND character_set_name='utf8mb4' AND collation_name='utf8mb4_bin') OR
 (column_name='repo_code' AND column_type='varchar(5)' AND is_nullable='NO' AND character_set_name='utf8mb4' AND collation_name='utf8mb4_bin') OR
 (column_name='lifecycle' AND column_type='varchar(6)' AND is_nullable='NO' AND character_set_name='utf8mb4' AND collation_name='utf8mb4_bin') OR
 (column_name IN ('created_at','updated_at') AND column_type='datetime(6)' AND is_nullable='NO') OR
 (column_name='frozen_at' AND column_type='datetime(6)' AND is_nullable='YES')));
SET @mng_draft_indexes=(SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='el_mng_exam_draft');
SET @mng_draft_pk=(SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='el_mng_exam_draft' AND index_name='PRIMARY' AND non_unique=0 AND seq_in_index=1 AND column_name='exam_id' AND sub_part IS NULL);
SET @mng_draft_fks=(SELECT COUNT(*) FROM information_schema.key_column_usage WHERE constraint_schema=DATABASE() AND table_name='el_mng_exam_draft' AND referenced_table_name IS NOT NULL);
SET @mng_draft_fk_ok=(SELECT COUNT(*) FROM information_schema.key_column_usage k JOIN information_schema.referential_constraints r ON r.constraint_schema=k.constraint_schema AND r.table_name=k.table_name AND r.constraint_name=k.constraint_name WHERE k.constraint_schema=DATABASE() AND k.table_name='el_mng_exam_draft' AND k.constraint_name='fk_mng_draft_exam' AND k.column_name='exam_id' AND k.ordinal_position=1 AND k.referenced_table_schema=DATABASE() AND k.referenced_table_name='el_exam' AND k.referenced_column_name='id' AND r.update_rule='RESTRICT' AND r.delete_rule='RESTRICT');
SET @mng_draft_engine=(SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name='el_mng_exam_draft' AND engine='InnoDB');
SET @mng_draft_sql=IF(@mng_draft_columns=7 AND @mng_draft_columns_ok=7 AND @mng_draft_indexes=1 AND @mng_draft_pk=1 AND @mng_draft_fks=1 AND @mng_draft_fk_ok=1 AND @mng_draft_engine=1,'SELECT ''MNG_DRAFT_SIGNATURE_OK''','SELECT * FROM information_schema.MNG_DRAFT_SIGNATURE_MISMATCH');
PREPARE mng_draft_stmt FROM @mng_draft_sql; EXECUTE mng_draft_stmt; DEALLOCATE PREPARE mng_draft_stmt;
-- Rollback: retain sidecar and new draft/frozen rows; never drop old tables or
-- remove markers to force new data into legacy. Restore reviewed binaries/UI
-- only with a compatible fail-closed maintenance plan.