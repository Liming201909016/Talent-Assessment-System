#!/usr/bin/env bash
# Explicitly disposable restore only. Main schema, credentials and UI untouched.
set -euo pipefail
umask 077
payload=${1:?owned payload required}
[[ "$payload" =~ ^/tmp/mng_lifecycle_[a-f0-9]{16}$ ]]
[ "$(id -u)" = 0 ] && [ "$(hostname)" = vm-ubuntu-go-dev ]
root=/opt/talent-assessment
backup=$root/backups/mng_phase1_20261003_112853_fd24d4ee35a9
current=$backup/http_772ccd9f3e56
stamp=${payload##*_}
schema=mng_lifecycle_test_$stamp
grant_schema=${schema//_/\\_}
evidence=$backup/lifecycle_$stamp
q() { mysql --batch --skip-column-names "$@"; }
legacy() { mysqldump --single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary element el_exam el_exam_repo el_repo el_qu el_qu_repo el_qu_answer el_paper el_paper_qu el_paper_qu_answer el_candidate el_tester el_mbti_answer | sha256sum | cut -d' ' -f1; }
pdfs() { find "$root/tmp" /data/uploadPath -type f -iname '*.pdf' -print0 | sort -z | xargs -0 -r sha256sum; }
immutable() { find "$root/dist" "$root/configs" /etc/systemd/system/talent-assessment.service.d -type f -print0 | sort -z | xargs -0 sha256sum; sha256sum /etc/systemd/system/talent-assessment.service "$root/server"; }
[ "$(cat "$backup/ownership")" = mng-phase1-staging-owned-v1 ]
[ "$(sha256sum "$current/element-current.sql.gz" | cut -d' ' -f1)" = 9ebee6a155ca17c7814085d7f839c8c1d59f2ef9ee3baa8dd5ea5f9f7902241f ]
gzip -t "$current/element-current.sql.gz"
if gzip -dc "$current/element-current.sql.gz" | grep -E '^((CREATE|DROP) DATABASE|USE )' >/dev/null; then exit 4; fi
[ "$(q -e "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name='$schema'")" = 0 ]
(cd "$payload"; sha256sum -c SHA256SUMS)
mkdir -m 0700 "$evidence"
printf '%s\n' mng-restored-lifecycle-staging-owned-v1 "$schema" "$payload" > "$evidence/ownership.txt"
cp "$payload/SHA256SUMS" "$evidence/payload-SHA256SUMS"
cp "$payload/lifecycle.sh" "$evidence/executed.sh"
legacy > "$evidence/legacy-before.sha256"; pdfs > "$evidence/pdf-before.sha256"; immutable > "$evidence/immutable-before.sha256"
cmp "$current/legacy-after.sha256" "$evidence/legacy-before.sha256"
cmp "$current/pdf-after.sha256" "$evidence/pdf-before.sha256"
[ "$(wc -l < "$evidence/pdf-before.sha256")" = 465 ]
owned=0
granted=0
pid_before=$(systemctl show talent-assessment -p MainPID --value)
stat -c '%u:%g:%a:%s:%Y:%Z' /var/spool/libreoffice/uno_packages/cache/uno_packages > "$evidence/cache-before.stat"
find "$root/configs" "$root/dist" -printf '%p %u:%g:%m\n' | sort > "$evidence/asset-metadata-before.txt"
mysqldump --single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary element el_repo el_qu el_qu_repo el_qu_answer | sha256sum | cut -d' ' -f1 > "$evidence/source-before.sha256"
cp "$current/element-current.sql.gz" "$evidence/element-current.sql.gz"
printf 'BACKUP_REUSED_CURRENT_NO_DRIFT=1 BACKUP_SHA=9ebee6a155ca17c7814085d7f839c8c1d59f2ef9ee3baa8dd5ea5f9f7902241f\n'
finish() {
 code=$?; trap - EXIT; bad=0
 if [ -d "$payload/output" ]; then cp -a "$payload/output" "$evidence/output" || bad=1; fi
 if [ "$granted" = 1 ]; then q -e "REVOKE SELECT,INSERT,UPDATE,DELETE,TRIGGER ON \`$grant_schema\`.* FROM 'positive_app'@'%'" || bad=1; fi
 if [ "$owned" = 1 ]; then
  [ "$(cat "$evidence/ownership.txt" | sed -n '2p')" = "$schema" ] && q -e "DROP DATABASE \`$schema\`" || bad=1
 fi
 remaining=$(q -e "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name='$schema'") || bad=1
 legacy > "$evidence/legacy-after.sha256" || bad=1
 pdfs > "$evidence/pdf-after.sha256" || bad=1
 immutable > "$evidence/immutable-after.sha256" || bad=1
 for name in legacy pdf immutable; do cmp "$evidence/$name-before.sha256" "$evidence/$name-after.sha256" || bad=1; done
 mysqldump --single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary element el_repo el_qu el_qu_repo el_qu_answer | sha256sum | cut -d' ' -f1 > "$evidence/source-after.sha256" || bad=1
 cmp "$evidence/source-before.sha256" "$evidence/source-after.sha256" || bad=1
 stat -c '%u:%g:%a:%s:%Y:%Z' /var/spool/libreoffice/uno_packages/cache/uno_packages > "$evidence/cache-after.stat" || bad=1
 cmp "$evidence/cache-before.stat" "$evidence/cache-after.stat" || bad=1
 find "$root/configs" "$root/dist" -printf '%p %u:%g:%m\n' | sort > "$evidence/asset-metadata-after.txt" || bad=1
 cmp "$evidence/asset-metadata-before.txt" "$evidence/asset-metadata-after.txt" || bad=1
 [ "$(systemctl show talent-assessment -p MainPID --value)" = "$pid_before" ] || bad=1
 [ "$(find "$root/private/management-traits-test-reports" -type f | wc -l)" = 0 ] || bad=1
 while IFS= read -r table; do
  [[ "$table" =~ ^el_mng_[a-z_]+$ ]] || { bad=1; continue; }
  [ "$(q element -e "SELECT COUNT(*) FROM \`$table\`")" = 0 ] || bad=1
 done < <(q element -e "SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_'")
 for s in talent-assessment nginx mysql; do systemctl is-active --quiet "$s" || bad=1; done
 curl --fail --silent --max-time 10 http://127.0.0.1:8092/health > "$evidence/health.json" || bad=1
 (cd "$payload"; sha256sum -c SHA256SUMS) > "$evidence/payload-final-check.txt" || bad=1
 cp "$payload/handler.test" "$payload/bootstrap" "$evidence/" || bad=1
 while read -r hash file; do [[ "$file" =~ ^(handler.test|bootstrap|lifecycle.sh)$ ]] && [ "$(sha256sum "$payload/$file" | cut -d' ' -f1)" = "$hash" ] && rm -- "$payload/$file" || bad=1; done < "$payload/SHA256SUMS"
 rm -- "$payload/SHA256SUMS" || bad=1
 if [ -d "$payload/output" ]; then find "$payload/output" -type f -print0 | while IFS= read -r -d '' f; do rel=${f#"$payload/output/"}; [ -f "$evidence/output/$rel" ] && cmp "$f" "$evidence/output/$rel" && rm -- "$f" || exit 1; done || bad=1; find "$payload/output" -depth -type d -empty -delete; fi
 rmdir "$payload" || bad=1
 [ "$(q -e "SELECT COUNT(*) FROM information_schema.schema_privileges WHERE table_schema='$schema'")" = 0 ] || bad=1
 gzip -t "$evidence/element-current.sql.gz" || bad=1
 sha256sum "$evidence/element-current.sql.gz" > "$evidence/backup-SHA256SUMS"
 find "$evidence" -type f ! -name FINAL_SHA256SUMS ! -name receipt.txt -print0 | sort -z | xargs -0 sha256sum > "$evidence/FINAL_SHA256SUMS"
 find "$evidence" -type d -exec chmod 0700 {} +
 find "$evidence" -type f -exec chmod 0600 {} +
 printf 'EXIT=%s RESTORE_SCHEMA_REMAINING=%s CLEANUP_ERRORS=%s MAIN_MNG_ROWS=0 PRIVATE_FILES=0 OLD_PDFS=465 CURRENT_BASELINE_UNCHANGED=1 MAIN_PID=%s MAIN_RESTART=0 SHARED_MODE_UNCHANGED=1 SERVICE_ONLY_NOT_HTTP=1\n' "$code" "$remaining" "$bad" "$pid_before" | tee "$evidence/receipt.txt"
 printf 'LIFECYCLE_EVIDENCE=%s\n' "$evidence"
 if [ "$bad" != 0 ] || [ "$remaining" != 0 ]; then exit 90; fi
 exit "$code"
}
trap finish EXIT
trap 'exit 130' HUP INT TERM
q -e "CREATE DATABASE \`$schema\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci"
owned=1
gzip -dc "$current/element-current.sql.gz" | mysql "$schema" > "$evidence/restore.log" 2>&1
[ "$(q "$schema" -e 'SELECT COUNT(*) FROM information_schema.key_column_usage WHERE table_schema=DATABASE() AND referenced_table_schema IS NOT NULL AND referenced_table_schema<>DATABASE()')" = 0 ]
[ "$(q "$schema" -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_'")" = 11 ]
q -e "GRANT SELECT,INSERT,UPDATE,DELETE,TRIGGER ON \`$grant_schema\`.* TO 'positive_app'@'%'"
granted=1
printf 'RESTORE_CROSS_SCHEMA_FK=0 MAIN_DDL=0 APPLICATION_ACCOUNT=positive_app TEMP_SCOPE_ONLY=1\n'
pid=$(systemctl show talent-assessment -p MainPID --value)
"$payload/bootstrap" "$schema" "$payload/handler.test" "$payload/output" "$pid" > "$evidence/test.log" 2>&1
cat "$evidence/test.log"
if grep -q -- '--- SKIP:' "$evidence/test.log"; then exit 5; fi
printf 'RESTORED_LIFECYCLE_PASS DDL_MAIN_EXECUTIONS=0 MAIN_BUSINESS_WRITES=0\n'