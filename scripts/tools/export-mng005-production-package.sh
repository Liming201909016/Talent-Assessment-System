#!/usr/bin/env bash
# Export the exact retained 005 TEST closure from staging for reviewed production migration.
# Run as root on vm-ubuntu-go-dev. The output contains synthetic TEST identity/report data.
set -euo pipefail
umask 077

output=${1:?absolute output directory required}
root=/opt/talent-assessment
marker='MNG005BASE_20261008_c51130cb775ba019'
e1='9051103000000000501'
e2='9051103000000000502'
r1='m5b501-c51130cb775b'
r2='m5b502-c51130cb775b'

[ "$(id -u)" = 0 ]
[ "$(hostname)" = vm-ubuntu-go-dev ]
[[ "$output" =~ ^/tmp/mng005-production-export-[a-z0-9-]+$ ]]
[ ! -e "$output" ]

pid=$(systemctl show talent-assessment -p MainPID --value)
report=''
while IFS= read -r -d '' item; do
  case "${item%%=*}" in REPORT_EFFECTIVE_ENV) report=${item#*=};; esac
done < /proc/"$pid"/environ
[ "$report" = staging ]
unset item

mysqlq() { mysql --batch --skip-column-names element "$@"; }
dump_rows() {
  local table=$1 where=$2
  mysqldump --single-transaction --quick --skip-lock-tables --skip-add-locks \
    --skip-disable-keys --skip-comments --compact --hex-blob --complete-insert \
    --skip-extended-insert --set-gtid-purged=OFF --no-tablespaces \
    --no-create-info --where="$where" element "$table"
}

actual=$(mysqlq -e "
SELECT CONCAT_WS('|',
 (SELECT COUNT(*) FROM el_repo WHERE id IN ('$r1','$r2')),
 (SELECT COUNT(*) FROM el_qu_repo WHERE repo_id IN ('$r1','$r2')),
 (SELECT COUNT(*) FROM el_qu_answer WHERE qu_id IN (SELECT qu_id FROM el_qu_repo WHERE repo_id IN ('$r1','$r2'))),
 (SELECT COUNT(*) FROM el_exam WHERE id IN ('$e1','$e2') AND title LIKE '[TEST-%'),
 (SELECT COUNT(*) FROM el_paper WHERE exam_id IN ('$e1','$e2')),
 (SELECT COUNT(*) FROM el_paper_qu WHERE paper_id IN (SELECT id FROM el_paper WHERE exam_id IN ('$e1','$e2'))),
 (SELECT COUNT(*) FROM el_paper_qu_answer WHERE paper_id IN (SELECT id FROM el_paper WHERE exam_id IN ('$e1','$e2'))),
 (SELECT COUNT(*) FROM el_mng_result_run WHERE exam_id IN ('$e1','$e2')),
 (SELECT COUNT(*) FROM el_mng_report_revision WHERE exam_id IN ('$e1','$e2')),
 (SELECT COUNT(*) FROM el_mng_report_reissue WHERE exam_id IN ('$e1','$e2'))
);")
[ "$actual" = '2|280|1400|2|2|280|1400|2|2|2' ]

mkdir -m 0700 -p "$output/assets/reports/baseline-c51130cb775ba019" "$output/assets/reports/reissues" "$output/assets/templates"
data="$output/010_mng005_test_baseline_data.sql"
cat > "$data" <<'SQL'
-- Exact retained 00501/00502 TEST baseline. Synthetic identities only.
-- Requires management_traits_001_runtime.sql and management_traits_004_report_reissues.sql.
-- Fails before DML when any package-owned identity is already occupied.
SET NAMES utf8mb4;
SET @mng005_preflight=(
  (SELECT COUNT(*) FROM el_repo WHERE code IN ('00501','00502') OR id IN ('m5b501-c51130cb775b','m5b502-c51130cb775b'))=0 AND
  (SELECT COUNT(*) FROM el_exam WHERE id IN ('9051103000000000501','9051103000000000502'))=0 AND
  (SELECT COUNT(*) FROM el_qu WHERE id LIKE 'b501-%' OR id LIKE 'b502-%')=0
);
SET @mng005_sql=IF(@mng005_preflight,'SELECT ''MNG005_TARGET_EMPTY''','SELECT * FROM information_schema.MNG005_TARGET_OCCUPIED');
PREPARE mng005_stmt FROM @mng005_sql; EXECUTE mng005_stmt; DEALLOCATE PREPARE mng005_stmt;
START TRANSACTION;
SQL
{
  dump_rows el_repo "id IN ('$r1','$r2')"
  dump_rows el_qu "id IN (SELECT qu_id FROM el_qu_repo WHERE repo_id IN ('$r1','$r2'))"
  dump_rows el_qu_answer "qu_id IN (SELECT qu_id FROM el_qu_repo WHERE repo_id IN ('$r1','$r2'))"
  dump_rows el_qu_repo "repo_id IN ('$r1','$r2')"
  dump_rows el_exam "id IN ('$e1','$e2')"
  dump_rows el_exam_repo "exam_id IN ('$e1','$e2')"
  dump_rows el_mng_definition_bundle "id IN (SELECT bundle_id FROM el_mng_exam_profile WHERE exam_id IN ('$e1','$e2') UNION SELECT bundle_id FROM el_mng_paper_snapshot WHERE exam_id IN ('$e1','$e2'))"
  dump_rows el_mng_exam_profile "exam_id IN ('$e1','$e2')"
  dump_rows el_candidate "exam_id IN ('$e1','$e2')"
  dump_rows el_paper "exam_id IN ('$e1','$e2')"
  dump_rows el_paper_qu "paper_id IN (SELECT id FROM el_paper WHERE exam_id IN ('$e1','$e2'))"
  dump_rows el_paper_qu_answer "paper_id IN (SELECT id FROM el_paper WHERE exam_id IN ('$e1','$e2'))"
  dump_rows el_mng_paper_snapshot "exam_id IN ('$e1','$e2')"
  dump_rows el_mng_paper_question_snapshot "paper_id IN (SELECT paper_id FROM el_mng_paper_snapshot WHERE exam_id IN ('$e1','$e2'))"
  dump_rows el_mng_result_run "exam_id IN ('$e1','$e2')"
  dump_rows el_mng_result_dimension "run_id IN (SELECT id FROM el_mng_result_run WHERE exam_id IN ('$e1','$e2'))"
  dump_rows el_mng_result_module "run_id IN (SELECT id FROM el_mng_result_run WHERE exam_id IN ('$e1','$e2'))"
  dump_rows el_mng_runtime_receipt "exam_id IN ('$e1','$e2')"
  dump_rows el_mng_report_revision "exam_id IN ('$e1','$e2')"
  dump_rows el_mng_report_current "report_id IN (SELECT id FROM el_mng_report_revision WHERE exam_id IN ('$e1','$e2'))"
  dump_rows el_mng_report_audit "report_id IN (SELECT id FROM el_mng_report_revision WHERE exam_id IN ('$e1','$e2'))"
  dump_rows el_mng_report_reissue "exam_id IN ('$e1','$e2')"
  dump_rows el_mng_reissue_audit "report_id IN (SELECT id FROM el_mng_report_reissue WHERE exam_id IN ('$e1','$e2'))"
} >> "$data"
cat >> "$data" <<'SQL'
SET @mng005_postflight=(
  (SELECT COUNT(*) FROM el_repo WHERE id IN ('m5b501-c51130cb775b','m5b502-c51130cb775b'))=2 AND
  (SELECT COUNT(*) FROM el_qu_repo WHERE repo_id IN ('m5b501-c51130cb775b','m5b502-c51130cb775b'))=280 AND
  (SELECT COUNT(*) FROM el_qu_answer WHERE qu_id IN (SELECT qu_id FROM el_qu_repo WHERE repo_id IN ('m5b501-c51130cb775b','m5b502-c51130cb775b')))=1400 AND
  (SELECT COUNT(*) FROM el_exam WHERE id IN ('9051103000000000501','9051103000000000502') AND title LIKE '[TEST-%')=2 AND
  (SELECT COUNT(*) FROM el_paper WHERE exam_id IN ('9051103000000000501','9051103000000000502'))=2 AND
  (SELECT COUNT(*) FROM el_mng_result_run WHERE exam_id IN ('9051103000000000501','9051103000000000502'))=2 AND
  (SELECT COUNT(*) FROM el_mng_report_revision WHERE exam_id IN ('9051103000000000501','9051103000000000502'))=2 AND
  (SELECT COUNT(*) FROM el_mng_report_reissue WHERE exam_id IN ('9051103000000000501','9051103000000000502'))=2
);
SET @mng005_sql=IF(@mng005_postflight,'SELECT ''MNG005_DATA_OK''','SELECT * FROM information_schema.MNG005_DATA_MISMATCH');
PREPARE mng005_stmt FROM @mng005_sql; EXECUTE mng005_stmt; DEALLOCATE PREPARE mng005_stmt;
COMMIT;
SQL

report_root="$root/private/management-traits-test-reports"
cp --preserve=mode,timestamps "$report_root/baseline-c51130cb775ba019/fffa814c-49ac-4f5f-8ede-f8493bed1055.pdf" "$output/assets/reports/baseline-c51130cb775ba019/"
cp --preserve=mode,timestamps "$report_root/baseline-c51130cb775ba019/910dc08e-fcc1-464c-b755-9626438fb9ab.pdf" "$output/assets/reports/baseline-c51130cb775ba019/"
cp --preserve=mode,timestamps "$report_root/reissues/8d6e59ae-6a9c-4b21-8044-b234d2c36e5f.pdf" "$output/assets/reports/reissues/"
cp --preserve=mode,timestamps "$report_root/reissues/b57c3452-5463-4562-a9d1-3105c90008c3.pdf" "$output/assets/reports/reissues/"
cp --preserve=mode,timestamps "$root/configs/export-templates/management-traits-002-test-only-v2.docx" "$output/assets/templates/"
cp --preserve=mode,timestamps "$root/configs/export-templates/management-traits-002-test-content-v1.xlsx" "$output/assets/templates/"

mysqlq -e "
SELECT 'el_repo',COUNT(*) FROM el_repo WHERE id IN ('$r1','$r2') UNION ALL
SELECT 'el_qu',COUNT(*) FROM el_qu WHERE id IN (SELECT qu_id FROM el_qu_repo WHERE repo_id IN ('$r1','$r2')) UNION ALL
SELECT 'el_qu_answer',COUNT(*) FROM el_qu_answer WHERE qu_id IN (SELECT qu_id FROM el_qu_repo WHERE repo_id IN ('$r1','$r2')) UNION ALL
SELECT 'el_qu_repo',COUNT(*) FROM el_qu_repo WHERE repo_id IN ('$r1','$r2') UNION ALL
SELECT 'el_exam',COUNT(*) FROM el_exam WHERE id IN ('$e1','$e2') UNION ALL
SELECT 'el_paper',COUNT(*) FROM el_paper WHERE exam_id IN ('$e1','$e2') UNION ALL
SELECT 'el_mng_result_run',COUNT(*) FROM el_mng_result_run WHERE exam_id IN ('$e1','$e2') UNION ALL
SELECT 'el_mng_report_revision',COUNT(*) FROM el_mng_report_revision WHERE exam_id IN ('$e1','$e2') UNION ALL
SELECT 'el_mng_report_reissue',COUNT(*) FROM el_mng_report_reissue WHERE exam_id IN ('$e1','$e2');" > "$output/ROW-COUNTS.tsv"
printf '%s\n' \
  'schema=mng005-production-test-export-v1' \
  "source_host=$(hostname)" \
  'source_database=element' \
  "ownership_marker=$marker" \
  'identity_class=synthetic-test-only' \
  'labels_must_retain=[TEST-保留]' \
  'formal_registry_included=false' \
  'draft_rows_included=false' > "$output/MANIFEST.txt"

find "$output" -type f -exec chmod 0600 {} +
(
  cd "$output"
  find . -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum > SHA256SUMS
  sha256sum -c SHA256SUMS
)
printf 'MNG005_EXPORT_PASS=%s\n' "$output"
