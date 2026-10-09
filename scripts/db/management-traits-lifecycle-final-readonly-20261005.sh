#!/usr/bin/env bash
# Read-only completion evidence for the two exact disposable restore receipts.
set -euo pipefail
[ "$(id -u)" = 0 ] && [ "$(hostname)" = vm-ubuntu-go-dev ]
root=/opt/talent-assessment
parent=$root/backups/mng_phase1_20261003_112853_fd24d4ee35a9
q() { mysql --batch --skip-column-names "$@"; }
for stamp in 1ac0b50628a4d637 0ca26bf185d9b25f; do
  base=$parent/lifecycle_$stamp
  [ "$(sed -n '1p' "$base/ownership.txt")" = mng-restored-lifecycle-staging-owned-v1 ]
  [ "$(sed -n '2p' "$base/ownership.txt")" = "mng_lifecycle_test_$stamp" ]
  [ "$(q -e "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name='mng_lifecycle_test_$stamp'")" = 0 ]
  [ ! -e "/tmp/mng_lifecycle_$stamp" ]
  (cd "$base"; sha256sum -c FINAL_SHA256SUMS >/dev/null)
  for name in legacy pdf immutable source; do cmp "$base/$name-before.sha256" "$base/$name-after.sha256"; done
  cmp "$base/cache-before.stat" "$base/cache-after.stat"
  cmp "$base/asset-metadata-before.txt" "$base/asset-metadata-after.txt"
  [ "$(find "$base" -type d ! -perm 0700 | wc -l)" = 0 ]
  [ "$(find "$base" -type f ! -perm 0600 | wc -l)" = 0 ]
  [ "$(sha256sum "$base/element-current.sql.gz" | cut -d' ' -f1)" = 9ebee6a155ca17c7814085d7f839c8c1d59f2ef9ee3baa8dd5ea5f9f7902241f ]
  gzip -t "$base/element-current.sql.gz"
  printf 'EVIDENCE_STAMP=%s MANIFEST_VALID=1 DIR0700_FILE0600=1 RESTORE_DB=0 PAYLOAD=0 BACKUP_VALID=1\n' "$stamp"
  sha256sum "$base/test.log" "$base/FINAL_SHA256SUMS"
  grep -E 'ACTUAL_DATABASE_ACCOUNT|PHASE=|CASE=|FAILURE_STAGE=|NEGATIVE_STAGE=|REMAINING_SERVICE_INTEGRATION|^--- (PASS|FAIL)|RESTORED_LIFECYCLE_CHILD|OWNED_TRIGGER' "$base/test.log" || true
  cat "$base/receipt.txt"
done
q -e "SELECT 'OWNED_RESTORE_GRANTS',COUNT(*) FROM information_schema.schema_privileges WHERE REPLACE(table_schema,CHAR(92),'') IN ('mng_lifecycle_test_1ac0b50628a4d637','mng_lifecycle_test_0ca26bf185d9b25f');"
q element -e "SELECT 'MAIN_MNG_TABLES',COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_';SELECT 'MAIN_MNG_RESTRICT_FK',COUNT(*) FROM information_schema.referential_constraints WHERE constraint_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' AND delete_rule='RESTRICT' AND update_rule='RESTRICT';"
while read -r table; do
  [[ "$table" =~ ^el_mng_[a-z_]+$ ]]
  [ "$(q element -e "SELECT COUNT(*) FROM $table")" = 0 ]
  printf 'MAIN_TABLE=%s ROWS=0\n' "$table"
done < <(q element -e "SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name")
q element -e "SELECT 'MAIN_COUNTS',(SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()),(SELECT COUNT(*) FROM el_exam),(SELECT COUNT(*) FROM el_paper),(SELECT COUNT(*) FROM el_paper_qu),(SELECT COUNT(*) FROM el_paper_qu_answer),(SELECT COUNT(*) FROM el_candidate),(SELECT COUNT(*) FROM el_tester);SELECT 'OTHER_ACTIVE_ASSESSMENTS',COUNT(*) FROM el_paper WHERE state=1 AND limit_time>NOW();SELECT 'EXPIRED_COMPETENCY',COUNT(*) FROM el_paper p INNER JOIN el_exam e ON e.id=p.exam_id WHERE p.state=1 AND p.limit_time<=NOW() AND e.assessment_type='competency';SELECT 'BINARY_LOGGING',@@log_bin,@@log_bin_trust_function_creators;"
[ "$(find "$root/private/management-traits-test-reports" -type f | wc -l)" = 0 ]
pid=$(systemctl show talent-assessment -p MainPID --value)
[ "$pid" = 2002 ]
printf 'MAIN_PID=%s PRIVATE_FILES=0 MAIN_RESTART=0\n' "$pid"
tr '\0' '\n' < "/proc/$pid/environ" | grep -E '^(APP_ENV|REPORT_EFFECTIVE_ENV|MNG_TEST_REPORT_ENV)='
sha256sum "$root/server" "$root/dist/index.html"
stat -c 'SHARED_CACHE=%u:%g:%a' /var/spool/libreoffice/uno_packages/cache/uno_packages
for s in talent-assessment nginx mysql; do printf 'SERVICE=%s STATE=' "$s";systemctl is-active "$s";done
curl --fail --silent --max-time 10 http://127.0.0.1:8092/health;echo
date -u '+FINAL_UTC=%FT%TZ'
printf 'FINAL_READONLY_PASS=1 HTTP_ZERO_UI_UNVERIFIED=1 INDEPENDENT_DATA_SHA_UNVERIFIED=1\n'