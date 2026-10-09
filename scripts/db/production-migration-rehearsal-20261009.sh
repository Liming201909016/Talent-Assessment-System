#!/usr/bin/env bash
# Production MySQL 5.7 restored-copy rehearsal only. Never writes the element schema.
set -Eeuo pipefail
umask 077

stamp=${1:?16 lowercase hex stamp required}
payload=${2:?absolute package directory required}
[[ "$stamp" =~ ^[a-f0-9]{16}$ ]]
[[ "$payload" =~ ^/opt/talent-assessment/backups/production_migration_payload_20261009_[a-f0-9]{16}$ ]]
[ "$(hostname)" = iZ0yosjdcen2p4Z ]
[ "$(id -u)" = 0 ]
[ -d "$payload/schema" ]
[ -d "$payload/mng005" ]

app=/opt/talent-assessment
backup="$app/backups/production_migration_rehearsal_20261009_$stamp"
schema="production_migration_verify_$stamp"
client=/run/production-migration-rehearsal-$stamp.cnf
asset_root="/tmp/production_migration_assets_$stamp"
created=0
assets_created=0
phase=preflight

cleanup() {
  code=$?
  trap - EXIT
  cleanup_code=0
  rm -f "$client"
  if [ "$assets_created" = 1 ] && [[ "$asset_root" =~ ^/tmp/production_migration_assets_[a-f0-9]{16}$ ]]; then
    rm -rf "$asset_root" || cleanup_code=1
  fi
  if [ "$created" = 1 ] && [[ "$schema" =~ ^production_migration_verify_[a-f0-9]{16}$ ]]; then
    mysql --defaults-extra-file="$backup/client.cnf" -e "DROP DATABASE \`$schema\`;" 2>>"$backup/private-errors.log" || cleanup_code=1
    remaining=$(mysql --defaults-extra-file="$backup/client.cnf" --batch --skip-column-names -e "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='$schema';" 2>>"$backup/private-errors.log") || cleanup_code=1
    [ "${remaining:-unknown}" = 0 ] || cleanup_code=1
  fi
  rm -f "$backup/client.cnf"
  printf 'FINAL_PHASE=%s\nRUN_EXIT=%s\nCLEANUP_EXIT=%s\nOWNED_SCHEMA_REMAINING=%s\nBACKUP_ROOT=%s\n' "$phase" "$code" "$cleanup_code" "${remaining:-unknown}" "$backup"
  if [ "$code" = 0 ] && [ "$cleanup_code" = 0 ]; then exit 0; fi
  exit 1
}
trap cleanup EXIT HUP INT TERM

mkdir -m 0700 "$backup"
python3 - "$client" <<'PY'
from pathlib import Path
import re, sys
text = Path('/opt/talent-assessment/configs/application-production.yml').read_text(encoding='utf-8-sig').replace('\r', '')
dsn = next((line.split(':', 1)[1].strip().strip('"\'') for line in text.splitlines() if line.strip().startswith('dsn:')), None)
match = re.fullmatch(r'([^:]+):([^@]*)@tcp\(([^:)]+):(\d+)\)/(element)(?:\?.*)?', dsn or '')
if not match:
    raise SystemExit('DB_DSN_REJECTED')
user, password, host, port, _ = match.groups()
Path(sys.argv[1]).write_text(f'[client]\nuser={user}\npassword={password}\nhost={host}\nport={port}\n', encoding='utf-8')
Path(sys.argv[1]).chmod(0o600)
PY
cp -p "$client" "$backup/client.cnf"
mysqlq() { mysql --defaults-extra-file="$client" --batch --skip-column-names --raw "$@"; }

[ "$(mysqlq -e 'SELECT VERSION()')" = 5.7.44-log ]
[ "$(mysqlq -e "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='$schema'")" = 0 ]
[ "$(mysqlq element -e "SELECT COUNT(*) FROM el_paper WHERE state=1")" = 0 ]
(
  cd "$payload"
  sha256sum -c SHA256SUMS
)
systemctl show talent-assessment -p MainPID -p ActiveState -p NRestarts > "$backup/service.before"
sha256sum "$app/server" "$app/dist/index.html" > "$backup/runtime.before"

phase=backup
mysqldump --defaults-extra-file="$client" --single-transaction --quick --skip-lock-tables --routines --triggers --events --hex-blob --set-gtid-purged=OFF --no-tablespaces element | gzip -c > "$backup/element.sql.gz"
gzip -t "$backup/element.sql.gz"
gzip -dc "$backup/element.sql.gz" > "$backup/element.restore.sql"
chmod 0600 "$backup/element.restore.sql"
if grep -Ei '^[[:space:]]*(CREATE[[:space:]]+DATABASE|USE[[:space:]])' "$backup/element.restore.sql" >/dev/null ||
  grep -Ei '(^|[^[:alnum:]_])(`element`|element)[[:space:]]*\.' "$backup/element.restore.sql" >/dev/null; then
  echo UNSAFE_RESTORE_ROUTING
  exit 1
fi
sha256sum "$backup/element.sql.gz" > "$backup/SHA256SUMS"

phase=restore
mysqlq -e "CREATE DATABASE \`$schema\` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"
created=1
mysqlq "$schema" < "$backup/element.restore.sql"
rm -f "$backup/element.restore.sql"

schema_files=(
  competency_007_versions.sql
  competency_008_phase1_structures.sql
  competency_010_phase1_report_framework.sql
  competency_011_result_runs.sql
  competency_012_v2_dimension_catalog.sql
  competency_013_report_result_run_binding.sql
  management_traits_001_runtime.sql
  management_traits_003_new_draft.sql
  management_traits_004_report_reissues.sql
)
schema_signature() {
  mysqlq -e "SELECT TABLE_NAME,COLUMN_NAME,ORDINAL_POSITION,COLUMN_TYPE,IS_NULLABLE,COALESCE(COLLATION_NAME,'-') FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='$schema' ORDER BY TABLE_NAME,ORDINAL_POSITION; SELECT TABLE_NAME,INDEX_NAME,NON_UNIQUE,SEQ_IN_INDEX,COLUMN_NAME FROM information_schema.STATISTICS WHERE TABLE_SCHEMA='$schema' ORDER BY TABLE_NAME,INDEX_NAME,SEQ_IN_INDEX; SELECT k.TABLE_NAME,k.CONSTRAINT_NAME,k.ORDINAL_POSITION,k.COLUMN_NAME,k.REFERENCED_TABLE_NAME,k.REFERENCED_COLUMN_NAME,r.UPDATE_RULE,r.DELETE_RULE FROM information_schema.KEY_COLUMN_USAGE k JOIN information_schema.REFERENTIAL_CONSTRAINTS r ON r.CONSTRAINT_SCHEMA=k.CONSTRAINT_SCHEMA AND r.TABLE_NAME=k.TABLE_NAME AND r.CONSTRAINT_NAME=k.CONSTRAINT_NAME WHERE k.TABLE_SCHEMA='$schema' AND k.REFERENCED_TABLE_NAME IS NOT NULL ORDER BY k.TABLE_NAME,k.CONSTRAINT_NAME,k.ORDINAL_POSITION;" | sha256sum
}
stage_assets() {
  [ ! -e "$asset_root" ]
  mkdir -m 0700 -p "$asset_root/reports/reissues" "$asset_root/templates"
  assets_created=1
  cp "$payload/mng005/assets/reports/baseline-c51130cb775ba019/"*.pdf "$asset_root/reports/"
  cp "$payload/mng005/assets/reports/reissues/"*.pdf "$asset_root/reports/reissues/"
  cp "$payload/mng005/assets/templates/"* "$asset_root/templates/"
  find "$asset_root" -type d -exec chmod 0700 {} +
  find "$asset_root" -type f -exec chmod 0600 {} +
}
verify_assets() {
  local kind key expected_sha expected_bytes file actual_sha actual_bytes
  while IFS=$'\t' read -r kind key expected_sha expected_bytes; do
    file="$asset_root/reports/$key"
    [ "$kind" = reissue ] && file="$asset_root/reports/reissues/$key"
    [ -f "$file" ] && [ ! -L "$file" ]
    actual_sha=$(sha256sum "$file" | cut -d' ' -f1)
    actual_bytes=$(stat -c %s "$file")
    [ "$actual_sha" = "$expected_sha" ] && [ "$actual_bytes" = "$expected_bytes" ]
  done < <(mysqlq "$schema" -e "SELECT 'revision',file_key,file_sha,file_bytes FROM el_mng_report_revision WHERE exam_id IN ('9051103000000000501','9051103000000000502') UNION ALL SELECT 'reissue',file_key,file_sha,file_bytes FROM el_mng_report_reissue WHERE exam_id IN ('9051103000000000501','9051103000000000502') ORDER BY 1,2")
  [ "$(find "$asset_root/reports" -type f -name '*.pdf' | wc -l)" = 4 ]
  [ "$(sha256sum "$asset_root/templates/management-traits-002-test-only-v2.docx" | cut -d' ' -f1)" = "$(mysqlq "$schema" -e "SELECT MIN(template_sha) FROM (SELECT template_sha FROM el_mng_report_revision WHERE exam_id IN ('9051103000000000501','9051103000000000502') UNION ALL SELECT template_sha FROM el_mng_report_reissue WHERE exam_id IN ('9051103000000000501','9051103000000000502')) x")" ]
  [ "$(sha256sum "$asset_root/templates/management-traits-002-test-content-v1.xlsx" | cut -d' ' -f1)" = "$(mysqlq "$schema" -e "SELECT MIN(content_sha) FROM (SELECT content_sha FROM el_mng_report_revision WHERE exam_id IN ('9051103000000000501','9051103000000000502') UNION ALL SELECT content_sha FROM el_mng_report_reissue WHERE exam_id IN ('9051103000000000501','9051103000000000502')) x")" ]
}
historical_00502_digest() {
  mysqldump --defaults-extra-file="$client" --single-transaction --quick --skip-lock-tables --skip-add-locks --skip-disable-keys --skip-comments --compact --hex-blob --complete-insert --skip-extended-insert --set-gtid-purged=OFF --no-tablespaces --no-create-info "$schema" \
    el_paper el_paper_qu el_paper_qu_answer el_candidate el_mng_paper_snapshot el_mng_paper_question_snapshot el_mng_result_run el_mng_result_dimension el_mng_result_module el_mng_runtime_receipt el_mng_report_revision el_mng_report_current el_mng_report_audit el_mng_report_reissue el_mng_reissue_audit | sha256sum
}
verify_00502_profile_v2() {
  local i j mapped_question expected_question qid mapped_option expected_option
  mkdir -p "$backup/facts"
  : > "$backup/facts/00502.mapping-question-ids"
  : > "$backup/facts/00502.mapping-option-ids"
  [ "$(mysqlq "$schema" -e "SELECT COUNT(*) FROM el_mng_exam_profile p JOIN el_mng_definition_bundle b ON b.id=p.bundle_id WHERE p.exam_id='9051103000000000502' AND b.question_version='mng-00502-db-current-v2' AND p.mapping_sha=SHA2(CAST(p.mapping_snapshot AS BINARY),256) AND b.scoring_manifest_sha=SHA2(CAST(b.scoring_manifest AS BINARY),256) AND JSON_UNQUOTE(JSON_EXTRACT(p.mapping_snapshot,'$.manifestSha'))=b.scoring_manifest_sha AND JSON_LENGTH(JSON_EXTRACT(p.mapping_snapshot,'$.questions'))=140")" = 1 ]
  for i in $(seq 0 139); do
    qid=$(mysqlq "$schema" -e "SELECT JSON_UNQUOTE(JSON_EXTRACT(mapping_snapshot,'$.questions[$i].sourceQuestionId')) FROM el_mng_exam_profile WHERE exam_id='9051103000000000502'")
    [[ "$qid" =~ ^b502-[A-Za-z0-9_-]+$ ]]
    printf '%s\n' "$qid" >> "$backup/facts/00502.mapping-question-ids"
    mapped_question=$(mysqlq "$schema" -e "SELECT SHA2(CONCAT(JSON_UNQUOTE(JSON_EXTRACT(mapping_snapshot,'$.questions[$i].sourceQuestionId')),0x00,JSON_UNQUOTE(JSON_EXTRACT(mapping_snapshot,'$.questions[$i].content'))),256) FROM el_mng_exam_profile WHERE exam_id='9051103000000000502'")
    expected_question=$(mysqlq "$schema" -e "SELECT SHA2(CONCAT(id,0x00,title),256) FROM el_qu WHERE id='$qid'")
    [ "$mapped_question" = "$expected_question" ]
    [ "$(mysqlq "$schema" -e "SELECT JSON_LENGTH(JSON_EXTRACT(mapping_snapshot,'$.questions[$i].options')) FROM el_mng_exam_profile WHERE exam_id='9051103000000000502'")" = 5 ]
    for j in 0 1 2 3 4; do
      mapped_option=$(mysqlq "$schema" -e "SELECT SHA2(CONCAT(JSON_UNQUOTE(JSON_EXTRACT(mapping_snapshot,'$.questions[$i].options[$j].sourceOptionId')),0x00,JSON_UNQUOTE(JSON_EXTRACT(mapping_snapshot,'$.questions[$i].options[$j].raw')),0x00,JSON_UNQUOTE(JSON_EXTRACT(mapping_snapshot,'$.questions[$i].options[$j].content'))),256) FROM el_mng_exam_profile WHERE exam_id='9051103000000000502'")
      mysqlq "$schema" -e "SELECT JSON_UNQUOTE(JSON_EXTRACT(mapping_snapshot,'$.questions[$i].options[$j].sourceOptionId')) FROM el_mng_exam_profile WHERE exam_id='9051103000000000502'" >> "$backup/facts/00502.mapping-option-ids"
      expected_option=$(mysqlq "$schema" -e "SELECT SHA2(CONCAT(id,0x00,score,0x00,content),256) FROM el_qu_answer WHERE qu_id='$qid' ORDER BY score,id LIMIT $j,1")
      [ "$mapped_option" = "$expected_option" ]
    done
  done
  mysqlq "$schema" -e "SELECT q.id FROM el_qu_repo qr JOIN el_qu q ON q.id=qr.qu_id WHERE qr.repo_id='m5b502-c51130cb775b' ORDER BY qr.sort,qr.id" > "$backup/facts/00502.source-question-ids"
  mysqlq "$schema" -e "SELECT a.id FROM el_qu_repo qr JOIN el_qu_answer a ON a.qu_id=qr.qu_id WHERE qr.repo_id='m5b502-c51130cb775b' ORDER BY qr.sort,qr.id,a.score,a.id" > "$backup/facts/00502.source-option-ids"
  [ "$(wc -l < "$backup/facts/00502.mapping-question-ids")" = 140 ]
  [ "$(sort -u "$backup/facts/00502.mapping-question-ids" | wc -l)" = 140 ]
  [ "$(wc -l < "$backup/facts/00502.mapping-option-ids")" = 700 ]
  [ "$(sort -u "$backup/facts/00502.mapping-option-ids" | wc -l)" = 700 ]
  cmp "$backup/facts/00502.mapping-question-ids" "$backup/facts/00502.source-question-ids"
  cmp "$backup/facts/00502.mapping-option-ids" "$backup/facts/00502.source-option-ids"
}
for round in 1 2; do
  phase="schema-round-$round"
  for file in "${schema_files[@]}"; do
    mysqlq "$schema" < "$payload/schema/$file" > "$backup/$file.round-$round.private"
  done
  schema_signature > "$backup/schema.round-$round.sha"
done
cmp "$backup/schema.round-1.sha" "$backup/schema.round-2.sha"

phase=data-first-apply
mysqlq "$schema" < "$payload/mng005/010_mng005_test_baseline_data.sql" > "$backup/data-first.private"
historical_00502_digest > "$backup/historical.before-profile-v2.sha"
mysqlq "$schema" < "$payload/mng005/015_mng005_00502_profile_v2.sql" > "$backup/profile-v2-first.private"
verify_00502_profile_v2
historical_00502_digest > "$backup/historical.after-profile-v2.sha"
cmp "$backup/historical.before-profile-v2.sha" "$backup/historical.after-profile-v2.sha"
stage_assets
verify_assets
counts_sql="SELECT CONCAT_WS('|',(SELECT COUNT(*) FROM el_repo WHERE code IN ('00501','00502')),(SELECT COUNT(*) FROM el_exam WHERE id IN ('9051103000000000501','9051103000000000502') AND state IN (0,1)),(SELECT COUNT(*) FROM el_qu WHERE id LIKE 'b501-%' OR id LIKE 'b502-%'),(SELECT COUNT(*) FROM el_mng_result_run WHERE exam_id IN ('9051103000000000501','9051103000000000502')),(SELECT COUNT(*) FROM el_mng_report_revision WHERE exam_id IN ('9051103000000000501','9051103000000000502')),(SELECT COUNT(*) FROM el_mng_report_reissue WHERE exam_id IN ('9051103000000000501','9051103000000000502')))"
[ "$(mysqlq "$schema" -e "$counts_sql")" = '2|2|280|2|2|2' ]

phase=rollback-before-activation
mysqlq "$schema" < "$payload/mng005/090_mng005_test_baseline_rollback.sql" > "$backup/rollback-before-activation.private"
[ "$(mysqlq "$schema" -e "$counts_sql")" = '0|0|0|0|0|0' ]
rm -rf "$asset_root"
assets_created=0
[ ! -e "$asset_root" ]

phase=reapply-before-activation
mysqlq "$schema" < "$payload/mng005/010_mng005_test_baseline_data.sql" > "$backup/data-before-activation.private"
mysqlq "$schema" < "$payload/mng005/015_mng005_00502_profile_v2.sql" > "$backup/profile-v2-before-activation.private"
verify_00502_profile_v2
mysqlq "$schema" < "$payload/mng005/020_mng005_customer_activation_state.sql" > "$backup/activation-first.private"
[ "$(mysqlq "$schema" -e "SELECT COUNT(*) FROM el_exam WHERE id IN ('9051103000000000501','9051103000000000502') AND state=1")" = 2 ]
stage_assets
verify_assets
first=$(mysqlq "$schema" -e "$counts_sql")

phase=data-repeat-rejection
if mysqlq "$schema" < "$payload/mng005/010_mng005_test_baseline_data.sql" > "$backup/data-repeat.private" 2>>"$backup/private-errors.log"; then
  echo DATA_REPEAT_UNEXPECTED_SUCCESS
  exit 1
fi
[ "$(mysqlq "$schema" -e "$counts_sql")" = "$first" ]

phase=rollback-after-activation
mysqlq "$schema" < "$payload/mng005/090_mng005_test_baseline_rollback.sql" > "$backup/rollback.private"
[ "$(mysqlq "$schema" -e "$counts_sql")" = '0|0|0|0|0|0' ]
[ "$(mysqlq -e "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='$schema' AND LEFT(TABLE_NAME,7)='el_mng_'")" = 14 ]
rm -rf "$asset_root"
assets_created=0
[ ! -e "$asset_root" ]

phase=reapply
mysqlq "$schema" < "$payload/mng005/010_mng005_test_baseline_data.sql" > "$backup/data-reapply.private"
mysqlq "$schema" < "$payload/mng005/015_mng005_00502_profile_v2.sql" > "$backup/profile-v2-reapply.private"
verify_00502_profile_v2
mysqlq "$schema" < "$payload/mng005/020_mng005_customer_activation_state.sql" > "$backup/activation-reapply.private"
[ "$(mysqlq "$schema" -e "$counts_sql")" = '2|2|280|2|2|2' ]
stage_assets
verify_assets

phase=runtime-unchanged
systemctl show talent-assessment -p MainPID -p ActiveState -p NRestarts > "$backup/service.after"
sha256sum "$app/server" "$app/dist/index.html" > "$backup/runtime.after"
cmp "$backup/service.before" "$backup/service.after"
cmp "$backup/runtime.before" "$backup/runtime.after"
printf 'MYSQL57_RESTORE_PASS=1\nSCHEMA_REPEAT_PASS=1\nDATA_REPEAT_FAIL_CLOSED=1\nROLLBACK_BEFORE_AFTER_ACTIVATION_PASS=1\nPROFILE_V2_140_QUESTIONS_700_OPTIONS_PASS=1\nPROFILE_V2_HISTORICAL_BYTES_UNCHANGED=1\nROLLBACK_REAPPLY_PASS=1\nASSET_INSTALL_VERIFY_CLEANUP_PASS=1\nMAIN_RUNTIME_UNCHANGED=1\n'
phase=completed
