-- Copy-only cleanup for the 2026-10-08 005 customer-template report E2E.
-- Retains the reusable 00501/00502 repository fixtures and draft/reissue schemas.
SET NAMES utf8mb4;
SET @e1='1791466773929935301';
SET @e2='1791466835021877003';
SET @cleanup_schema_ok=(SELECT DATABASE()='talent_mng005_local_7081fbec31e3d105');
SET @cleanup_shape_ok=(
  (SELECT COUNT(*) FROM el_exam WHERE (id=@e1 AND title='MNG00501-REPORT-A66722421') OR (id=@e2 AND title='MNG00502-REPORT-B66833339'))=2 AND
  (SELECT COUNT(*) FROM el_mng_exam_profile WHERE exam_id IN (@e1,@e2))=2 AND
  (SELECT COUNT(*) FROM el_candidate WHERE exam_id IN (@e1,@e2))=2 AND
  (SELECT COUNT(*) FROM el_paper WHERE exam_id IN (@e1,@e2))=2 AND
  (SELECT COUNT(*) FROM el_mng_result_run WHERE exam_id IN (@e1,@e2) AND status='completed' AND answered_question_count=140)=2 AND
  (SELECT COUNT(*) FROM el_mng_report_reissue WHERE exam_id IN (@e1,@e2) AND status='completed')=2 AND
  (SELECT COUNT(*) FROM el_mng_reissue_audit WHERE exam_id IN (@e1,@e2) AND action IN ('generate','reuse','view','download'))>=8 AND
  (SELECT COUNT(*) FROM el_mng_report_revision WHERE exam_id IN (@e1,@e2))=0
);
SET @cleanup_sql=IF(@cleanup_schema_ok AND @cleanup_shape_ok,
  'SELECT ''MNG005_REPORT_CLEANUP_PRECONDITION_OK''',
  'SELECT * FROM information_schema.MNG005_REPORT_CLEANUP_PRECONDITION_FAILED');
PREPARE cleanup_stmt FROM @cleanup_sql; EXECUTE cleanup_stmt; DEALLOCATE PREPARE cleanup_stmt;
SET @bundle_00501=(SELECT bundle_id FROM el_mng_exam_profile WHERE exam_id=@e1);
SET @bundle_00502=(SELECT bundle_id FROM el_mng_exam_profile WHERE exam_id=@e2);
START TRANSACTION;
DELETE FROM el_mng_reissue_audit WHERE id IN (SELECT id FROM (SELECT id FROM el_mng_reissue_audit WHERE exam_id IN (@e1,@e2)) owned_audits);
DELETE FROM el_mng_report_reissue WHERE id IN (SELECT id FROM (SELECT id FROM el_mng_report_reissue WHERE exam_id IN (@e1,@e2)) owned_reports);
DELETE FROM el_mng_runtime_receipt WHERE run_id IN (SELECT id FROM el_mng_result_run WHERE exam_id IN (@e1,@e2));
DELETE FROM el_mng_result_dimension WHERE id IN (SELECT id FROM (SELECT d.id FROM el_mng_result_dimension d JOIN el_mng_result_run r ON r.id=d.run_id WHERE r.exam_id IN (@e1,@e2)) owned_dimensions);
DELETE FROM el_mng_result_module WHERE id IN (SELECT id FROM (SELECT m.id FROM el_mng_result_module m JOIN el_mng_result_run r ON r.id=m.run_id WHERE r.exam_id IN (@e1,@e2)) owned_modules);
DELETE FROM el_mng_result_run WHERE id IN (SELECT id FROM (SELECT id FROM el_mng_result_run WHERE exam_id IN (@e1,@e2)) owned_runs);
DELETE FROM el_mng_paper_question_snapshot WHERE id IN (SELECT id FROM (SELECT q.id FROM el_mng_paper_question_snapshot q JOIN el_mng_paper_snapshot p ON p.paper_id=q.paper_id WHERE p.exam_id IN (@e1,@e2)) owned_snapshots);
DELETE FROM el_paper_qu_answer WHERE id IN (SELECT id FROM (SELECT a.id FROM el_paper_qu_answer a JOIN el_paper p ON p.id=a.paper_id WHERE p.exam_id IN (@e1,@e2)) owned_answers);
DELETE FROM el_paper_qu WHERE id IN (SELECT id FROM (SELECT q.id FROM el_paper_qu q JOIN el_paper p ON p.id=q.paper_id WHERE p.exam_id IN (@e1,@e2)) owned_questions);
DELETE FROM el_mng_paper_snapshot WHERE paper_id IN (SELECT id FROM el_paper WHERE exam_id IN (@e1,@e2));
DELETE FROM el_candidate WHERE id IN (SELECT id FROM (SELECT id FROM el_candidate WHERE exam_id IN (@e1,@e2)) owned_candidates);
DELETE FROM el_paper WHERE id IN (SELECT id FROM (SELECT id FROM el_paper WHERE exam_id IN (@e1,@e2)) owned_papers);
DELETE FROM el_mng_exam_draft WHERE exam_id IN (@e1,@e2);
DELETE FROM el_mng_exam_profile WHERE exam_id IN (@e1,@e2);
DELETE FROM el_exam_repo WHERE id IN (SELECT id FROM (SELECT id FROM el_exam_repo WHERE exam_id IN (@e1,@e2)) owned_exam_repos);
DELETE FROM el_exam WHERE id IN (@e1,@e2);
DELETE FROM el_mng_definition_bundle
WHERE id IN (@bundle_00501,@bundle_00502)
  AND id NOT IN (SELECT bundle_id FROM el_mng_exam_profile)
  AND id NOT IN (SELECT bundle_id FROM el_mng_paper_snapshot);
SET @cleanup_result_ok=(
  (SELECT COUNT(*) FROM el_exam WHERE id IN (@e1,@e2))=0 AND
  (SELECT COUNT(*) FROM el_candidate WHERE exam_id IN (@e1,@e2))=0 AND
  (SELECT COUNT(*) FROM el_paper WHERE exam_id IN (@e1,@e2))=0 AND
  (SELECT COUNT(*) FROM el_mng_exam_profile WHERE exam_id IN (@e1,@e2))=0 AND
  (SELECT COUNT(*) FROM el_mng_result_run WHERE exam_id IN (@e1,@e2))=0 AND
  (SELECT COUNT(*) FROM el_mng_report_reissue WHERE exam_id IN (@e1,@e2))=0 AND
  (SELECT COUNT(*) FROM el_mng_reissue_audit WHERE exam_id IN (@e1,@e2))=0 AND
  (SELECT COUNT(*) FROM el_repo WHERE code IN ('00501','00502') AND remark='MNG005_LOCAL_FIXTURE:7081fbec31e3d105:v1')=2 AND
  (SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name IN ('el_mng_exam_draft','el_mng_report_reissue','el_mng_reissue_audit'))=3
);
SET @cleanup_sql=IF(@cleanup_result_ok,
  'SELECT ''MNG005_REPORT_CLEANUP_SIGNATURE_OK''',
  'SELECT * FROM information_schema.MNG005_REPORT_CLEANUP_SIGNATURE_FAILED');
PREPARE cleanup_stmt FROM @cleanup_sql; EXECUTE cleanup_stmt; DEALLOCATE PREPARE cleanup_stmt;
COMMIT;
