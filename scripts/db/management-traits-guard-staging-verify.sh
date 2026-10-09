#!/usr/bin/env bash
# Isolated real validation only: never applies DDL to element or deploys files.
set -euo pipefail
umask 077
root=/opt/talent-assessment
backup=$root/backups/mng_phase1_20261003_112853_fd24d4ee35a9
payload=${1:?exact payload required}
[[ "$payload" =~ ^/tmp/mng_guard_staging_[a-f0-9]{16}$ ]]
[ "$(id -u)" = 0 ] && [ "$(hostname)" = vm-ubuntu-go-dev ]
stamp=${payload##*_}
schema=mng_guard_audit_test_$stamp
lockschema=mng_source_lock_test_$stamp
evidence=$backup/guard_actual_$stamp
[ "$(cat "$backup/ownership")" = mng-phase1-staging-owned-v1 ]
[ "$(sha256sum "$backup/element.sql.gz" | cut -d' ' -f1)" = 9ab00b04d5b8acfe564d7c3a5031cc375b8959853b39a0e8087f8a4d23b01aa1 ]
(cd "$backup"; sha256sum -c SHA256SUMS)
gzip -t "$backup/element.sql.gz" "$backup/application.tar.gz" "$backup/files-and-system.tar.gz"
tar -tzf "$backup/application.tar.gz" >/dev/null
tar -tzf "$backup/files-and-system.tar.gz" >/dev/null
if gzip -dc "$backup/element.sql.gz" | grep -E '^((CREATE|DROP) DATABASE|USE )' >/dev/null; then exit 4; fi
[ "$(stat -c '%U:%G:%a' "$backup")" = root:root:700 ]
pid=$(systemctl show talent-assessment -p MainPID --value)
report=''
while IFS= read -r -d '' item; do
  case "${item%%=*}" in REPORT_EFFECTIVE_ENV) report=${item#*=};; esac
done < /proc/"$pid"/environ
[ "$report" = staging ]
unset item
mkdir -m 0700 "$evidence"
printf '%s\n' "$schema" "$lockschema" "$payload" > "$evidence/ownership.txt"
q() { mysql --batch --skip-column-names "$@"; }
legacy() { mysqldump --single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary element el_exam el_exam_repo el_repo el_qu el_qu_repo el_qu_answer el_paper el_paper_qu el_paper_qu_answer el_candidate el_tester el_mbti_answer | sha256sum | cut -d' ' -f1; }
pdfs() { find "$root/tmp" /data/uploadPath -type f -iname '*.pdf' -print0 | sort -z | xargs -0 -r sha256sum; }
legacy > "$evidence/legacy-before.sha256"
pdfs > "$evidence/pdf-before.sha256"
cmp "$backup/legacy-before.sha256" "$evidence/legacy-before.sha256"
cmp "$backup/pdf-before.sha256" "$evidence/pdf-before.sha256"
[ "$(wc -l < "$evidence/pdf-before.sha256")" = 465 ]
tar -C "$root" --compare -zf "$backup/application.tar.gz" > "$evidence/application-before.log" 2>&1
# Configuration/unit/template/archive drift is checked without printing content.
tar -C / --compare -zf "$backup/files-and-system.tar.gz" > "$evidence/files-before.log" 2>&1
(cd "$payload"; sha256sum -c SHA256SUMS)
[ "$(sha256sum "$payload/scripts/sql/management_traits_001_runtime.sql" | cut -d' ' -f1)" = 7ff62155861958eee787f735bc3a65eb7797fe39f3b09333182f093e375983e8 ]
[ "$(q -e "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name IN ('$schema','$lockschema')")" = 0 ]
owned=0; lockowned=0
cleanup() {
  code=$?; trap - EXIT; bad=0
  if [ "$lockowned" = 1 ]; then q -e "DROP DATABASE \`$lockschema\`" || bad=1; fi
  if [ "$owned" = 1 ]; then q -e "DROP DATABASE \`$schema\`" || bad=1; fi
  remaining=$(q -e "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name IN ('$schema','$lockschema')") || bad=1
  legacy > "$evidence/legacy-after.sha256" || bad=1
  pdfs > "$evidence/pdf-after.sha256" || bad=1
  cmp "$evidence/legacy-before.sha256" "$evidence/legacy-after.sha256" || bad=1
  cmp "$evidence/pdf-before.sha256" "$evidence/pdf-after.sha256" || bad=1
  tar -C "$root" --compare -zf "$backup/application.tar.gz" > "$evidence/application-after.log" 2>&1 || bad=1
  tar -C / --compare -zf "$backup/files-and-system.tar.gz" > "$evidence/files-after.log" 2>&1 || bad=1
  main=$(q element -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_'") || bad=1
  [ "${main:-unknown}" = 0 ] || bad=1
  for s in talent-assessment nginx mysql; do systemctl is-active --quiet "$s" || bad=1; done
  curl --fail --silent --max-time 10 http://127.0.0.1:8092/health > "$evidence/health.json" || bad=1
  printf 'EXIT=%s TEMP_SCHEMA_REMAINING=%s CLEANUP_ERRORS=%s MAIN_MNG=%s PDF_COUNT=465\n' "$code" "${remaining:-unknown}" "$bad" "${main:-unknown}" | tee "$evidence/receipt.txt"
  printf 'EVIDENCE=%s\n' "$evidence"
  if [ "$bad" != 0 ] || [ "${remaining:-unknown}" != 0 ]; then exit 90; fi
  exit "$code"
}
trap cleanup EXIT
trap 'exit 130' HUP INT TERM
q -e "CREATE DATABASE \`$schema\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci"
owned=1
gzip -dc "$backup/element.sql.gz" | mysql "$schema" > "$evidence/restore.log" 2>&1
[ "$(q "$schema" -e 'SELECT COUNT(*) FROM information_schema.key_column_usage WHERE table_schema=DATABASE() AND referenced_table_schema IS NOT NULL AND referenced_table_schema<>DATABASE()')" = 0 ]
cwd="$payload/Go-based Refactored System/internal/service"
signature() {
 q "$schema" -e "SELECT table_name,column_name,column_type,is_nullable,IFNULL(character_set_name,''),IFNULL(collation_name,'') FROM information_schema.columns WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name,ordinal_position; SELECT table_name,index_name,non_unique,seq_in_index,column_name,IFNULL(sub_part,0) FROM information_schema.statistics WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name,index_name,seq_in_index; SELECT k.table_name,k.constraint_name,k.column_name,k.ordinal_position,k.referenced_table_name,k.referenced_column_name,r.update_rule,r.delete_rule FROM information_schema.key_column_usage k JOIN information_schema.referential_constraints r ON r.constraint_schema=k.constraint_schema AND r.table_name=k.table_name AND r.constraint_name=k.constraint_name WHERE k.table_schema=DATABASE() AND LEFT(k.table_name,7)='el_mng_' ORDER BY k.table_name,k.constraint_name,k.ordinal_position;"
}
for phase in first repeat; do
 mysql "$schema" < "$payload/scripts/sql/management_traits_001_runtime.sql" > "$evidence/ddl-$phase.log" 2>&1
 [ "$(q "$schema" -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_'")" = 11 ]
 [ "$(q "$schema" -e "SELECT COUNT(*) FROM information_schema.referential_constraints WHERE constraint_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' AND update_rule='RESTRICT' AND delete_rule='RESTRICT'")" = 15 ]
 signature > "$evidence/schema-$phase.signature"
 "$payload/bootstrap" schema "$schema" "$payload/service.test" "$cwd" "$pid" > "$evidence/schema-$phase.log" 2>&1
 cat "$evidence/schema-$phase.log"
done
cmp "$evidence/schema-first.signature" "$evidence/schema-repeat.signature"
"$payload/bootstrap" post "$schema" "$payload/service.test" "$cwd" "$pid" > "$evidence/actual-post.log" 2>&1
cat "$evidence/actual-post.log"
if grep -q -- '--- SKIP:' "$evidence/actual-post.log"; then echo ACTUAL_SKIP_BLOCKED; exit 5; fi
signature > "$evidence/schema-post.signature"
cmp "$evidence/schema-first.signature" "$evidence/schema-post.signature"
q -e "CREATE DATABASE \`$lockschema\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci"
lockowned=1
"$payload/bootstrap" lock "$lockschema" "$payload/service.test" "$cwd" "$pid" > "$evidence/actual-lock.log" 2>&1
cat "$evidence/actual-lock.log"
[ "$(grep -c -- '--- PASS:' "$evidence/actual-lock.log")" = 4 ]
if grep -q -- '--- SKIP:' "$evidence/actual-lock.log"; then exit 5; fi
[ "$(q "$lockschema" -e 'SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()')" = 0 ]
printf 'SCHEMA_SIGNATURE_SHA256=%s\n' "$(sha256sum "$evidence/schema-first.signature" | cut -d' ' -f1)"
printf 'ACTUAL_GUARD_STAGING_PASS\n'