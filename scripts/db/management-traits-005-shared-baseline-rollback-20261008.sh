#!/usr/bin/env bash
set -euo pipefail

marker='MNG005BASE_20261008_c51130cb775ba019'
exam1='9051103000000000501'
exam2='9051103000000000502'
repo1='m5b501-c51130cb775b'
repo2='m5b502-c51130cb775b'
report_root='/opt/talent-assessment/private/management-traits-test-reports/baseline-c51130cb775ba019'

[ "${1:-}" = '--execute-exact-owned-rollback' ] || {
  printf 'Refusing: pass --execute-exact-owned-rollback after explicit approval.\n' >&2
  exit 64
}
[ "$(hostname)" = 'vm-ubuntu-go-dev' ]
[ -d "$report_root" ] && [ ! -L "$report_root" ]
[ "$(sudo -n stat -c '%U:%G:%a' "$report_root")" = 'root:root:700' ]

files=$(sudo -n mysql --batch --skip-column-names element <<SQL
START TRANSACTION READ ONLY;
SELECT file_key FROM el_mng_report_revision
WHERE exam_id IN ('$exam1','$exam2')
ORDER BY exam_id;
COMMIT;
SQL
)
[ "$(printf '%s\n' "$files" | sed '/^$/d' | wc -l)" -eq 2 ]
while IFS= read -r key; do
  [[ "$key" =~ ^[a-f0-9-]{36}\.pdf$ ]]
  [ -f "$report_root/$key" ] && [ ! -L "$report_root/$key" ]
done <<< "$files"

sudo -n mysql --batch --skip-column-names element <<SQL
SET NAMES utf8mb4;
SET SESSION SQL_SAFE_UPDATES=1;
START TRANSACTION;
SELECT id FROM el_exam
WHERE id IN ('$exam1','$exam2')
  AND title IN ('[TEST-保留] $marker 基层员工','[TEST-保留] $marker 干部')
FOR UPDATE;
SELECT id FROM el_repo
WHERE (id='$repo1' AND code='00501' AND remark='$marker:cloned-from:00201')
   OR (id='$repo2' AND code='00502' AND remark='$marker:cloned-from:00202')
FOR UPDATE;

SET @owned_exams=(SELECT COUNT(*) FROM el_exam WHERE (id='$exam1' AND title='[TEST-保留] $marker 基层员工') OR (id='$exam2' AND title='[TEST-保留] $marker 干部'));
SET @owned_repos=(SELECT COUNT(*) FROM el_repo WHERE (id='$repo1' AND code='00501' AND remark='$marker:cloned-from:00201') OR (id='$repo2' AND code='00502' AND remark='$marker:cloned-from:00202'));
SET @owned_reports=(SELECT COUNT(*) FROM el_mng_report_revision WHERE exam_id IN ('$exam1','$exam2') AND mode='test' AND test_label='仅供系统测试，不可作为人才决策依据');
SET @owned_runs=(SELECT COUNT(*) FROM el_mng_result_run WHERE exam_id IN ('$exam1','$exam2') AND status='completed' AND answered_question_count=140);
SET @guard_ok=(@owned_exams=2 AND @owned_repos=2 AND @owned_reports=2 AND @owned_runs=2);
SET @guard_sql=IF(@guard_ok,'SELECT 1','SELECT * FROM information_schema.MNG005_BASELINE_ROLLBACK_GUARD_FAILED');
PREPARE guard_stmt FROM @guard_sql; EXECUTE guard_stmt; DEALLOCATE PREPARE guard_stmt;

CREATE TEMPORARY TABLE rollback_exam_ids (id varchar(64) PRIMARY KEY) ENGINE=MEMORY;
INSERT INTO rollback_exam_ids VALUES ('$exam1'),('$exam2');
CREATE TEMPORARY TABLE rollback_repo_ids (id varchar(64) PRIMARY KEY) ENGINE=MEMORY;
INSERT INTO rollback_repo_ids VALUES ('$repo1'),('$repo2');
CREATE TEMPORARY TABLE rollback_question_ids (id varchar(64) PRIMARY KEY) ENGINE=MEMORY;
INSERT INTO rollback_question_ids
SELECT DISTINCT qr.qu_id FROM el_qu_repo qr JOIN rollback_repo_ids r ON r.id=qr.repo_id;
CREATE TEMPORARY TABLE rollback_bundle_ids (id varchar(64) PRIMARY KEY) ENGINE=MEMORY;
INSERT INTO rollback_bundle_ids
SELECT DISTINCT p.bundle_id FROM el_mng_exam_profile p JOIN rollback_exam_ids e ON e.id=p.exam_id;

DELETE a FROM el_mng_report_audit a JOIN el_mng_report_revision r ON r.id=a.report_id JOIN rollback_exam_ids e ON e.id=r.exam_id;
DELETE c FROM el_mng_report_current c JOIN el_mng_report_revision r ON r.id=c.report_id JOIN rollback_exam_ids e ON e.id=r.exam_id;
DELETE r FROM el_mng_report_revision r JOIN rollback_exam_ids e ON e.id=r.exam_id;
DELETE x FROM el_mng_runtime_receipt x JOIN rollback_exam_ids e ON e.id=x.exam_id;
DELETE d FROM el_mng_result_dimension d JOIN el_mng_result_run r ON r.id=d.run_id JOIN rollback_exam_ids e ON e.id=r.exam_id;
DELETE m FROM el_mng_result_module m JOIN el_mng_result_run r ON r.id=m.run_id JOIN rollback_exam_ids e ON e.id=r.exam_id;
DELETE r FROM el_mng_result_run r JOIN rollback_exam_ids e ON e.id=r.exam_id;
DELETE q FROM el_mng_paper_question_snapshot q JOIN el_mng_paper_snapshot p ON p.paper_id=q.paper_id JOIN rollback_exam_ids e ON e.id=p.exam_id;
DELETE a FROM el_paper_qu_answer a JOIN el_paper p ON p.id=a.paper_id JOIN rollback_exam_ids e ON e.id=p.exam_id;
DELETE q FROM el_paper_qu q JOIN el_paper p ON p.id=q.paper_id JOIN rollback_exam_ids e ON e.id=p.exam_id;
DELETE p FROM el_mng_paper_snapshot p JOIN rollback_exam_ids e ON e.id=p.exam_id;
DELETE c FROM el_candidate c JOIN rollback_exam_ids e ON e.id=c.exam_id;
DELETE p FROM el_paper p JOIN rollback_exam_ids e ON e.id=p.exam_id;
DELETE p FROM el_mng_exam_profile p JOIN rollback_exam_ids e ON e.id=p.exam_id;
DELETE er FROM el_exam_repo er JOIN rollback_exam_ids e ON e.id=er.exam_id;
DELETE e FROM el_exam e JOIN rollback_exam_ids x ON x.id=e.id;
DELETE b FROM el_mng_definition_bundle b JOIN rollback_bundle_ids x ON x.id=b.id
WHERE b.id NOT IN (SELECT bundle_id FROM el_mng_exam_profile) AND b.id NOT IN (SELECT bundle_id FROM el_mng_paper_snapshot);
DELETE a FROM el_qu_answer a JOIN rollback_question_ids q ON q.id=a.qu_id;
DELETE qr FROM el_qu_repo qr JOIN rollback_repo_ids r ON r.id=qr.repo_id;
DELETE q FROM el_qu q JOIN rollback_question_ids x ON x.id=q.id
WHERE q.id NOT IN (SELECT qu_id FROM el_qu_repo);
DELETE r FROM el_repo r JOIN rollback_repo_ids x ON x.id=r.id;
COMMIT;

SELECT 'RESIDUAL',
 (SELECT COUNT(*) FROM el_exam WHERE id IN ('$exam1','$exam2')),
 (SELECT COUNT(*) FROM el_repo WHERE id IN ('$repo1','$repo2')),
 (SELECT COUNT(*) FROM el_mng_report_revision WHERE exam_id IN ('$exam1','$exam2'));
SQL

while IFS= read -r key; do
  sudo -n rm -- "$report_root/$key"
done <<< "$files"
printf 'MNG005_SHARED_BASELINE_ROLLBACK_COMPLETE\n'
