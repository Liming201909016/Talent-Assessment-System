#!/usr/bin/env bash
# Explicitly authorized 002 TEST staging deployment; no legacy writes/backfill.
set -euo pipefail
umask 077
payload=${1:?owned payload required}
[[ "$payload" =~ ^/tmp/mng_deploy_staging_[a-f0-9]{16}$ ]]
[ "$(id -u)" = 0 ] && [ "$(hostname)" = vm-ubuntu-go-dev ]
root=/opt/talent-assessment
backup=$root/backups/mng_phase1_20261003_112853_fd24d4ee35a9
stamp=${payload##*_}
evidence=$backup/deploy_$stamp
dropin=/etc/systemd/system/talent-assessment.service.d/management-traits-test.conf
envfile=$root/configs/management-traits-staging-test.env
private=$root/private/management-traits-test-reports
newdist=$root/dist.mng-new-$stamp
olddist=$root/dist.mng-rollback-$stamp
q() { mysql --batch --skip-column-names "$@"; }
legacy() { mysqldump --single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary element el_exam el_exam_repo el_repo el_qu el_qu_repo el_qu_answer el_paper el_paper_qu el_paper_qu_answer el_candidate el_tester el_mbti_answer | sha256sum | cut -d' ' -f1; }
pdfs() { find "$root/tmp" /data/uploadPath -type f -iname '*.pdf' -print0 | sort -z | xargs -0 -r sha256sum; }
signature() { q element -e "SELECT table_name,column_name,column_type,is_nullable,IFNULL(character_set_name,''),IFNULL(collation_name,'') FROM information_schema.columns WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name,ordinal_position; SELECT table_name,index_name,non_unique,seq_in_index,column_name,IFNULL(sub_part,0) FROM information_schema.statistics WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name,index_name,seq_in_index; SELECT k.table_name,k.constraint_name,k.column_name,k.ordinal_position,k.referenced_table_name,k.referenced_column_name,r.update_rule,r.delete_rule FROM information_schema.key_column_usage k JOIN information_schema.referential_constraints r ON r.constraint_schema=k.constraint_schema AND r.table_name=k.table_name AND r.constraint_name=k.constraint_name WHERE k.table_schema=DATABASE() AND LEFT(k.table_name,7)='el_mng_' ORDER BY k.table_name,k.constraint_name,k.ordinal_position;"; }
empty_sidecars() {
 local table
 while IFS= read -r table; do
  [[ "$table" =~ ^el_mng_[a-z_]+$ ]] || return 1
  [ "$(q element -e "SELECT COUNT(*) FROM \`$table\`")" = 0 ] || return 1
 done < <(q element -e "SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_'")
}
pid=$(systemctl show talent-assessment -p MainPID --value)
# Keep secrets only in inherited process memory; never write or print them.
while IFS= read -r -d '' item; do
 case "${item%%=*}" in APP_ENV|REPORT_EFFECTIVE_ENV|MYSQL_DSN|JWT_SECRET|SERVER_PORT|LIBREOFFICE_PATH) export "$item";; esac
done < /proc/"$pid"/environ
unset item
[ "${REPORT_EFFECTIVE_ENV:-}" = staging ]
[ "$(systemctl show talent-assessment -p User --value)" = liming ]
[ "$(systemctl show talent-assessment -p WorkingDirectory --value)" = "$root" ]
[ "$(systemctl show talent-assessment -p KillMode --value)" = control-group ]
[ ! -e "$dropin" ] && [ ! -e "$envfile" ] && [ ! -e "$private" ]
[ ! -e "$newdist" ] && [ ! -e "$olddist" ]
for f in management-traits-002-test-only-v2.docx management-traits-002-test-content-v1.xlsx; do [ ! -e "$root/configs/export-templates/$f" ]; done
[ "$(cat "$backup/ownership")" = mng-phase1-staging-owned-v1 ]
[ "$(sha256sum "$backup/element.sql.gz" | cut -d' ' -f1)" = 9ab00b04d5b8acfe564d7c3a5031cc375b8959853b39a0e8087f8a4d23b01aa1 ]
(cd "$backup"; sha256sum -c SHA256SUMS)
gzip -t "$backup/element.sql.gz" "$backup/application.tar.gz" "$backup/files-and-system.tar.gz"
[ "$(stat -c '%U:%G:%a' "$backup")" = root:root:700 ]
[ "$(df -B1 --output=avail "$root" | tail -1 | tr -d ' ')" -gt 2147483648 ]
(cd "$payload"; sha256sum -c SHA256SUMS)
[ "$(sha256sum "$payload/runtime.sql" | cut -d' ' -f1)" = 7ff62155861958eee787f735bc3a65eb7797fe39f3b09333182f093e375983e8 ]
mkdir -m 0700 "$evidence"
printf '%s\n' mng-staging-deployment-owned-v1 "$payload" "$olddist" > "$evidence/ownership.txt"
cp "$payload/SHA256SUMS" "$evidence/payload-SHA256SUMS"
cp "$payload/check-source.go.txt" "$evidence/checker-source.go.txt"
cp "$payload/check" "$evidence/checker"
cp "$payload/deploy.sh" "$evidence/deploy-executed.sh"
cp "$payload/runtime.sql" "$evidence/runtime-not-reapplied.sql"
legacy > "$evidence/legacy-before.sha256"
pdfs > "$evidence/pdf-before.sha256"
cmp "$backup/legacy-before.sha256" "$evidence/legacy-before.sha256"
cmp "$backup/pdf-before.sha256" "$evidence/pdf-before.sha256"
tar -C "$root" --compare -zf "$backup/application.tar.gz" > "$evidence/application-before.log" 2>&1
tar -C / --compare -zf "$backup/files-and-system.tar.gz" > "$evidence/files-before.log" 2>&1
[ "$(wc -l < "$evidence/pdf-before.sha256")" = 465 ]
sudo -u liming "$payload/check" pre "$pid" > "$evidence/precheck.log" 2>&1
cat "$evidence/precheck.log"
appuser=$(sed -n 's/.* APP_DB_USER=\([^ ]*\) MNG_TABLES=.*/\1/p' "$evidence/precheck.log")
[[ "$appuser" =~ ^[a-zA-Z0-9_]+$ ]]
existing_tables=$(q element -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_'")
[ "$existing_tables" = 0 ] || [ "$existing_tables" = 11 ]
empty_sidecars
stat -c '%u %g %a' "$root/server" > "$evidence/server.metadata.before"
cp -a "$root/server" "$evidence/server.before"
cp -a /etc/systemd/system/talent-assessment.service "$evidence/unit.before"
tar -C /etc/systemd/system -czf "$evidence/dropins.before.tar.gz" talent-assessment.service.d
tar -C "$root" -czf "$evidence/configs.before.tar.gz" configs
chmod 0600 "$evidence"/*
stage=prepared; changed=0; distchanged=0; configured=0; success=0
finish() {
 code=$?; trap - EXIT; rollback=not-needed
 if [ "$success" != 1 ] && [ "$changed" = 1 ]; then
  rollback=blocked; systemctl stop talent-assessment || true
  if empty_sidecars; then
   if [ "$configured" = 1 ]; then rm -f -- "$dropin" "$envfile"; systemctl daemon-reload; fi
   if [ "$distchanged" = 1 ]; then mv "$root/dist" "$newdist"; mv "$olddist" "$root/dist"; fi
    read -r server_uid server_gid server_mode < "$evidence/server.metadata.before"
    cp "$evidence/server.before" "$root/server.rollback-$stamp"
    chown "$server_uid:$server_gid" "$root/server.rollback-$stamp"
    chmod "$server_mode" "$root/server.rollback-$stamp"
    touch -r "$evidence/server.before" "$root/server.rollback-$stamp"
   mv "$root/server.rollback-$stamp" "$root/server"
   if systemctl start talent-assessment; then rollback=restored-previous-app-schema-preserved; fi
  fi
 fi
 printf 'DEPLOYED=%s EXIT=%s STAGE=%s ROLLBACK=%s DDL_PRESERVED=1\n' "$success" "$code" "$stage" "$rollback" | tee "$evidence/deploy-receipt.txt"
 printf 'DEPLOY_EVIDENCE=%s\n' "$evidence"
 exit "$code"
}
trap finish EXIT
trap 'exit 130' HUP INT TERM
stop_and_drain() {
 local old=$1 cg
 cg=$(systemctl show talent-assessment -p ControlGroup --value)
 systemctl stop talent-assessment
 [ "$(systemctl show talent-assessment -p MainPID --value)" = 0 ]
 [ ! -e "/proc/$old" ]
 if [ -n "$cg" ] && [ -f "/sys/fs/cgroup$cg/cgroup.procs" ]; then [ -z "$(cat "/sys/fs/cgroup$cg/cgroup.procs")" ]; fi
 if ss -lntp | grep -q ':8092 '; then return 1; fi
 [ "$(q -e "SELECT COUNT(*) FROM information_schema.processlist WHERE user='$appuser'")" = 0 ]
 printf 'DRAIN_PASS_PID=%s CONTROL_GROUP_EMPTY=1 LISTENER_8092=0 APP_DB_CONNECTIONS=0\n' "$old" | tee -a "$evidence/drain.log"
}
# Phase A: all old children/HTTP/DB connections terminate before replacement.
stage=old-drain; changed=1
stop_and_drain "$pid"
install -o liming -g liming -m 0755 "$payload/server" "$root/server.mng-new-$stamp"
mv "$root/server.mng-new-$stamp" "$root/server"
systemctl start talent-assessment
newpid=$(systemctl show talent-assessment -p MainPID --value)
sudo -u liming "$payload/check" pre "$newpid" > "$evidence/guard-running-disabled.log" 2>&1
cat "$evidence/guard-running-disabled.log"
curl --fail --silent --retry 5 --retry-connrefused --max-time 10 http://127.0.0.1:8092/health > "$evidence/guard-health.json"
printf 'NEW_GUARD_RUNNING_TEST_DISABLED=PASS PID=%s\n' "$newpid"
# Phase B: stop new guard as well. No backend listener/inflight writer at DDL.
stage=new-guard-drain
stop_and_drain "$newpid"
for phase in first repeat; do
 stage=main-ddl-$phase
 if [ "$existing_tables" = 0 ]; then
  mysql element < "$payload/runtime.sql" > "$evidence/ddl-$phase.log" 2>&1
 else
  printf 'MAIN_DDL_REUSED_NO_EXECUTION TABLES=11\n' > "$evidence/ddl-$phase.log"
 fi
 [ "$(q element -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_'")" = 11 ]
 [ "$(q element -e "SELECT COUNT(*) FROM information_schema.referential_constraints WHERE constraint_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' AND update_rule='RESTRICT' AND delete_rule='RESTRICT'")" = 15 ]
 signature > "$evidence/schema-$phase.signature"
 "$payload/check" gate - > "$evidence/gate-$phase.log" 2>&1
 cat "$evidence/gate-$phase.log"
done
cmp "$evidence/schema-first.signature" "$evidence/schema-repeat.signature"
empty_sidecars
stage=test-assets
install -d -o liming -g liming -m 0700 "$root/private" "$private"
for f in management-traits-002-test-only-v2.docx management-traits-002-test-content-v1.xlsx; do
 install -o liming -g liming -m 0600 "$payload/$f" "$root/configs/export-templates/$f"
done
configured=1
printf '%s\n' 'MNG_TEST_REPORT_ENV=staging' "MNG_TEST_REPORT_DIR=$private" "MNG_TEST_CONTENT_PATH=$root/configs/export-templates/management-traits-002-test-content-v1.xlsx" "MNG_TEST_TEMPLATE_PATH=$root/configs/export-templates/management-traits-002-test-only-v2.docx" > "$envfile"
chown root:liming "$envfile"; chmod 0640 "$envfile"
printf '[Service]\nEnvironmentFile=%s\n' "$envfile" > "$dropin"
chown root:root "$dropin"; chmod 0644 "$dropin"
systemctl daemon-reload
stage=frontend
mkdir -m 0755 "$newdist"
tar -xzf "$payload/dist.tar.gz" -C "$newdist" --strip-components=1
chmod -R 0755 "$newdist"; chown -R root:root "$newdist"
mv "$root/dist" "$olddist"; mv "$newdist" "$root/dist"; distchanged=1
nginx -t > "$evidence/nginx-test.log" 2>&1
stage=restart-postcheck
systemctl start talent-assessment
newpid=$(systemctl show talent-assessment -p MainPID --value)
curl --fail --silent --retry 5 --retry-connrefused --max-time 10 http://127.0.0.1:8092/health > "$evidence/health.json"
sudo -u liming "$payload/check" post "$newpid" > "$evidence/postcheck.log" 2>&1
cat "$evidence/postcheck.log"
signature > "$evidence/schema-final.signature"
cmp "$evidence/schema-first.signature" "$evidence/schema-final.signature"
legacy > "$evidence/legacy-after.sha256"; pdfs > "$evidence/pdf-after.sha256"
cmp "$evidence/legacy-before.sha256" "$evidence/legacy-after.sha256"
cmp "$evidence/pdf-before.sha256" "$evidence/pdf-after.sha256"
empty_sidecars
for s in talent-assessment nginx mysql; do systemctl is-active "$s"; done
sha256sum "$root/server" "$root/dist/index.html" "$envfile" "$dropin" "$root/configs/export-templates/management-traits-002-test-only-v2.docx" "$root/configs/export-templates/management-traits-002-test-content-v1.xlsx" > "$evidence/deployed-SHA256SUMS"
cat "$evidence/deployed-SHA256SUMS"
curl --fail --silent --max-time 15 http://127.0.0.1/prod-api/health
curl --fail --silent --max-time 15 http://127.0.0.1/index.html | sha256sum
printf 'TABLES=11 RESTRICT_FKS=15 SIDECAR_ROWS=0 PRIVATE_FILES=0 OLD_PDFS=465 LEGACY_SHA256=%s SCHEMA_SHA256=%s\n' "$(cat "$evidence/legacy-after.sha256")" "$(sha256sum "$evidence/schema-final.signature" | cut -d' ' -f1)"
printf '%s\n' 'Rollback: stop talent-assessment and drain; verify all el_mng_* empty; remove only new management-traits-test.conf and management-traits-staging-test.env; daemon-reload; restore server.before atomically using saved server.metadata.before uid/gid/mode and server.before mtime (private copy mode600 is not original mode); restore retained dist.mng-rollback directory; start/health. Preserve 11 main DDL tables and formal backups. Never drop tables or restore the whole element database automatically. If new sidecar records exist, stop and require reviewed forward/rollback handling.' > "$evidence/rollback.txt"
stage=verified; success=1
printf 'STAGING_TEST_DEPLOY_PASS\n'