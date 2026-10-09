-- Production policy: visible TEST assessments arrive disabled; the customer uses
-- the existing exam-state operation to enable either assessment. No new switch.
SET NAMES utf8mb4;
SET @mng005_owned=(
  (SELECT COUNT(*) FROM el_exam WHERE id IN ('9051103000000000501','9051103000000000502') AND title LIKE '[TEST-保留] MNG005BASE_20261008_c51130cb775ba019%' AND state=0)=2
);
SET @mng005_sql=IF(@mng005_owned,'SELECT ''MNG005_ACTIVATION_POLICY_OK''','SELECT * FROM information_schema.MNG005_ACTIVATION_POLICY_MISMATCH');
PREPARE mng005_stmt FROM @mng005_sql; EXECUTE mng005_stmt; DEALLOCATE PREPARE mng005_stmt;
START TRANSACTION;
UPDATE el_exam SET state=1, update_time=NOW() WHERE id IN ('9051103000000000501','9051103000000000502') AND state=0;
SET @mng005_disabled=(SELECT COUNT(*) FROM el_exam WHERE id IN ('9051103000000000501','9051103000000000502') AND state=1)=2;
SET @mng005_sql=IF(@mng005_disabled,'SELECT ''MNG005_VISIBLE_DISABLED_OK''','SELECT * FROM information_schema.MNG005_VISIBLE_DISABLED_MISMATCH');
PREPARE mng005_stmt FROM @mng005_sql; EXECUTE mng005_stmt; DEALLOCATE PREPARE mng005_stmt;
COMMIT;
