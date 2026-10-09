#!/usr/bin/env bash
# Authorized staging backup / isolated rehearsal. Never migrates element.
set -euo pipefail
umask 077
action=${1:?action required}
root=/opt/talent-assessment
[ "$(id -u)" = 0 ] && [ "$(hostname)" = vm-ubuntu-go-dev ]
pid=$(systemctl show talent-assessment -p MainPID --value)
app=local; report=''; dsn=''
while IFS= read -r -d '' item; do
  case "${item%%=*}" in APP_ENV) app=${item#*=};; REPORT_EFFECTIVE_ENV) report=${item#*=};; MYSQL_DSN) dsn=${item#*=};; esac
done < /proc/"$pid"/environ
[ "$report" = staging ]
case "$app" in local|staging|production) ;; *) exit 2;; esac
if [ -z "$dsn" ]; then
  for f in "$root/configs/application.yml" "$root/configs/application.yaml" "$root/configs/application-$app.yml" "$root/configs/application-$app.yaml"; do
    if [ -f "$f" ]; then
      value=$(awk '/^[^[:space:]#]/{section=($0 ~ /^mysql:/)} section && /^[[:space:]]+dsn:/{sub(/^[[:space:]]+dsn:[[:space:]]*/,"");gsub(/\r/,"");sub(/^["\047]/,"");sub(/["\047][[:space:]]*$/,"");print;exit}' "$f")
      if [ -n "$value" ]; then dsn=$value; fi
    fi
  done
fi
case "$dsn" in *'@tcp(127.0.0.1:3306)/element?'*|*'@tcp(localhost:3306)/element?'*) ;; *) echo EFFECTIVE_DB_UNCONFIRMED; exit 3;; esac
unset dsn value item
mysqlq() { mysql --batch --skip-column-names "$@"; }
legacy_digest() {
  mysqldump --single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary element el_exam el_exam_repo el_repo el_qu el_qu_repo el_qu_answer el_paper el_paper_qu el_paper_qu_answer el_candidate el_tester el_mbti_answer | sha256sum | cut -d' ' -f1
}
pdf_inventory() {
  find "$root/tmp" /data/uploadPath -type f -iname '*.pdf' -print0 | sort -z | xargs -0 -r sha256sum
}
if [ "$action" = backup ]; then
  available=$(df -B1 --output=avail "$root" | tail -1 | tr -d ' ')
  size=$(du -sb "$root/dist" "$root/configs" "$root/tmp" "$root/mbti-templates" "$root/mbti-templates-simple" /data/uploadPath | awk '{s+=$1} END{print s}')
  [ "$available" -gt "$((size*4+2147483648))" ]
  b="$root/backups/mng_phase1_20261003_$(date -u +%H%M%S)_$(cat /proc/sys/kernel/random/uuid | tr -d '-' | cut -c1-12)"
  mkdir -m 0700 "$b"
  printf '%s\n' mng-phase1-staging-owned-v1 > "$b/ownership"
  legacy_digest > "$b/legacy-before.sha256"
  pdf_inventory > "$b/pdf-before.sha256"
  mysqlq element -e 'SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE(); SELECT COUNT(*) FROM el_exam; SELECT COUNT(*) FROM el_paper; SELECT COUNT(*) FROM el_paper_qu; SELECT COUNT(*) FROM el_paper_qu_answer; SELECT COUNT(*) FROM el_candidate; SELECT COUNT(*) FROM el_tester;' > "$b/counts-before.txt"
  mysqldump --single-transaction --quick --skip-lock-tables --routines --triggers --events --hex-blob --set-gtid-purged=OFF --no-tablespaces element | gzip -c > "$b/element.sql.gz"
  gzip -t "$b/element.sql.gz"
  # Without --databases: no CREATE DATABASE / USE can redirect restore to element.
  if gzip -dc "$b/element.sql.gz" | grep -E '^((CREATE|DROP) DATABASE|USE )' > "$b/restore-routing-check.txt"; then echo UNSAFE_DUMP_ROUTING; exit 4; fi
  tar -C "$root" -czf "$b/application.tar.gz" server dist configs mbti-templates mbti-templates-simple 2> "$b/application-tar.log"
  tar -C / -czf "$b/files-and-system.tar.gz" opt/talent-assessment/tmp data/uploadPath etc/nginx etc/systemd/system/talent-assessment.service 2> "$b/files-tar.log"
  gzip -t "$b/application.tar.gz" "$b/files-and-system.tar.gz"
  tar -tzf "$b/application.tar.gz" >/dev/null
  tar -tzf "$b/files-and-system.tar.gz" >/dev/null
  pdf_inventory > "$b/pdf-after-backup.sha256"
  cmp "$b/pdf-before.sha256" "$b/pdf-after-backup.sha256"
  chmod 0600 "$b"/*
  (cd "$b"; sha256sum element.sql.gz application.tar.gz files-and-system.tar.gz > SHA256SUMS; sha256sum -c SHA256SUMS)
  [ "$(stat -c '%U:%G:%a' "$b")" = root:root:700 ]
  [ "$(find "$b" -type f ! -perm 0600 | wc -l)" = 0 ]
  printf 'BACKUP_PATH=%s\n' "$b"
  printf 'DATABASE_SHA256=%s\n' "$(sha256sum "$b/element.sql.gz" | cut -d' ' -f1)"
  printf 'DATABASE_GZIP_BYTES=%s PDF_FILES=%s\n' "$(stat -c %s "$b/element.sql.gz")" "$(wc -l < "$b/pdf-before.sha256")"
  printf 'BACKUP_PASS\n'
  exit 0
fi
[ "$action" = rehearse ]
b=${2:?backup required}; payload=${3:?payload required}
[[ "$b" =~ ^/opt/talent-assessment/backups/mng_phase1_20261003_[0-9]{6}_[a-f0-9]{12}$ ]]
[[ "$payload" =~ ^/tmp/mng_phase1_20261003_[a-f0-9]{12}$ ]]
[ "$(cat "$b/ownership")" = mng-phase1-staging-owned-v1 ]
(cd "$b"; sha256sum -c SHA256SUMS)
stamp=$(cat /proc/sys/kernel/random/uuid | tr -d '-' | cut -c1-16)
backup=$b
b="$backup/rehearsal_$stamp"
mkdir -m 0700 "$b"
cp "$backup/legacy-before.sha256" "$backup/pdf-before.sha256" "$b/"
printf 'REHEARSAL_EVIDENCE=%s\n' "$b"
schema="mng_verify_20261003_$stamp"
lockschema="mng_source_lock_test_20261003_$stamp"
owned=0; lockowned=0
cleanup() {
  code=$?
  trap - EXIT
  cleanupfailed=0
  if [ "$lockowned" = 1 ]; then mysqlq -e "DROP DATABASE \`$lockschema\`" || cleanupfailed=1; fi
  if [ "$owned" = 1 ]; then mysqlq -e "DROP DATABASE \`$schema\`" || cleanupfailed=1; fi
  remaining=$(mysqlq -e "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name IN ('$schema','$lockschema')") || cleanupfailed=1
  printf 'TEMP_SCHEMA_REMAINING=%s CLEANUP_ERRORS=%s\n' "${remaining:-unknown}" "$cleanupfailed"
  printf 'EXIT=%s REMAINING=%s ERRORS=%s\n' "$code" "${remaining:-unknown}" "$cleanupfailed" > "$b/cleanup-receipt.txt"
  if [ "$cleanupfailed" != 0 ] || [ "${remaining:-unknown}" != 0 ]; then exit 90; fi
  exit "$code"
}
trap cleanup EXIT
trap 'exit 130' HUP INT TERM
[ "$(mysqlq -e "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name IN ('$schema','$lockschema')")" = 0 ]
mysqlq -e "CREATE DATABASE \`$schema\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci"
owned=1
printf '%s\n' "$schema" > "$b/temp-schema-created.txt"
gzip -dc "$backup/element.sql.gz" | mysql "$schema" > "$b/restore.log" 2>&1
printf 'RESTORE_PASS\n'
gatefailed=0
for phase in first repeat; do
  mysql "$schema" < "$payload/runtime.sql" > "$b/ddl-$phase.log" 2>&1
  tables=$(mysqlq "$schema" -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_'")
  fks=$(mysqlq "$schema" -e "SELECT COUNT(*) FROM information_schema.referential_constraints WHERE constraint_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' AND update_rule='RESTRICT' AND delete_rule='RESTRICT'")
  [ "$tables" = 11 ] && [ "$fks" = 15 ]
  mysqlq "$schema" -e "SELECT table_name,column_name,column_type,is_nullable,IFNULL(character_set_name,''),IFNULL(collation_name,'') FROM information_schema.columns WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name,ordinal_position; SELECT table_name,index_name,non_unique,seq_in_index,column_name,IFNULL(sub_part,0) FROM information_schema.statistics WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name,index_name,seq_in_index; SELECT table_name,constraint_name,referenced_table_name,update_rule,delete_rule FROM information_schema.referential_constraints WHERE constraint_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name,constraint_name;" > "$b/schema-$phase.signature"
  if ! "$payload/verifier" schema "$schema" > "$b/schema-gate-$phase.log" 2>&1; then gatefailed=1; fi
  cat "$b/schema-gate-$phase.log"
  printf 'DDL_%s_TABLES=%s RESTRICT_FKS=%s SIGNATURE_SHA256=%s\n' "$phase" "$tables" "$fks" "$(sha256sum "$b/schema-$phase.signature" | cut -d' ' -f1)"
done
cmp "$b/schema-first.signature" "$b/schema-repeat.signature"
mysqlq -e "CREATE DATABASE \`$lockschema\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci"
lockowned=1
printf '%s\n' "$lockschema" > "$b/lock-schema-created.txt"
"$payload/verifier" lock "$lockschema" "$payload/service.test" > "$b/source-lock.log" 2>&1
cat "$b/source-lock.log"
[ "$(mysqlq "$lockschema" -e 'SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()')" = 0 ]
legacy_digest > "$b/legacy-after.sha256"
pdf_inventory > "$b/pdf-after.sha256"
cmp "$b/legacy-before.sha256" "$b/legacy-after.sha256"
cmp "$b/pdf-before.sha256" "$b/pdf-after.sha256"
[ "$(mysqlq element -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_'")" = 0 ]
printf 'LEGACY_SHA256=%s LEGACY_UNCHANGED=1 PDF_UNCHANGED=1 MAIN_MNG_TABLES=0\n' "$(cat "$b/legacy-after.sha256")"
for s in talent-assessment nginx mysql; do systemctl is-active --quiet "$s"; done
curl --fail --silent --max-time 10 http://127.0.0.1:8092/health
if [ "$gatefailed" != 0 ]; then printf '\nPHASE1_BLOCKED_ACTUAL_SCHEMA_GATE\n'; exit 1; fi
printf '\nPHASE1_REHEARSAL_PASS\n'