#!/usr/bin/env bash
# Reuses the existing deployment's baseline, gate, drain and metadata contracts.
set -euo pipefail
umask 077
payload=${1:?owned payload required}
[[ "$payload" =~ ^/tmp/mng_backend_scope_[a-f0-9]{16}$ ]]
[ "$(id -u)" = 0 ] && [ "$(hostname)" = vm-ubuntu-go-dev ]
root=/opt/talent-assessment
backup=$root/backups/mng_phase1_20261003_112853_fd24d4ee35a9
stamp=${payload##*_}
evidence=$backup/backend_$stamp
q() { mysql --batch --skip-column-names "$@"; }
legacy() { mysqldump --single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary element el_exam el_exam_repo el_repo el_qu el_qu_repo el_qu_answer el_paper el_paper_qu el_paper_qu_answer el_candidate el_tester el_mbti_answer | sha256sum | cut -d' ' -f1; }
pdfs() { find "$root/tmp" /data/uploadPath -type f -iname '*.pdf' -print0 | sort -z | xargs -0 -r sha256sum; }
immutable() { find "$root/dist" "$root/configs" /etc/systemd/system/talent-assessment.service.d -type f -print0 | sort -z | xargs -0 sha256sum; sha256sum /etc/systemd/system/talent-assessment.service; }
signature() { q element -e "SELECT table_name,column_name,column_type,is_nullable,IFNULL(character_set_name,''),IFNULL(collation_name,'') FROM information_schema.columns WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name,ordinal_position; SELECT table_name,index_name,non_unique,seq_in_index,column_name,IFNULL(sub_part,0) FROM information_schema.statistics WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name,index_name,seq_in_index; SELECT k.table_name,k.constraint_name,k.column_name,k.ordinal_position,k.referenced_table_name,k.referenced_column_name,r.update_rule,r.delete_rule FROM information_schema.key_column_usage k JOIN information_schema.referential_constraints r ON r.constraint_schema=k.constraint_schema AND r.table_name=k.table_name AND r.constraint_name=k.constraint_name WHERE k.table_schema=DATABASE() AND LEFT(k.table_name,7)='el_mng_' ORDER BY k.table_name,k.constraint_name,k.ordinal_position;"; }
empty_sidecars() {
 local table
 [ "$(q element -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_'")" = 11 ]
 while IFS= read -r table; do
  [[ "$table" =~ ^el_mng_[a-z_]+$ ]] || return 1
  [ "$(q element -e "SELECT COUNT(*) FROM \`$table\`")" = 0 ] || return 1
 done < <(q element -e "SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_'")
}
pid=$(systemctl show talent-assessment -p MainPID --value)
for s in talent-assessment nginx mysql; do systemctl is-active --quiet "$s"; done
while IFS= read -r -d '' item; do
 case "${item%%=*}" in APP_ENV|REPORT_EFFECTIVE_ENV|MYSQL_DSN|JWT_SECRET|SERVER_PORT|LIBREOFFICE_PATH|MNG_TEST_REPORT_ENV|MNG_TEST_REPORT_DIR|MNG_TEST_TEMPLATE_PATH|MNG_TEST_CONTENT_PATH) export "$item";; esac
done < /proc/"$pid"/environ
unset item
[ "${REPORT_EFFECTIVE_ENV:-}" = staging ] && [ "${MNG_TEST_REPORT_ENV:-}" = staging ]
[ "$(systemctl show talent-assessment -p User --value)" = liming ]
[ "$(systemctl show talent-assessment -p WorkingDirectory --value)" = "$root" ]
[ "$(systemctl show talent-assessment -p KillMode --value)" = control-group ]
empty_sidecars
# A restart must not drive another participant's expiry workflow for this test.
[ "$(q element -e "SELECT COUNT(*) FROM el_paper p JOIN el_exam e ON e.id=p.exam_id WHERE p.state=1 AND e.assessment_type='competency' AND e.scoring_mode='competency_average' AND p.limit_time<=NOW()")" = 0 ]
[ "$(cat "$backup/ownership")" = mng-phase1-staging-owned-v1 ]
[ "$(stat -c '%U:%G:%a' "$backup")" = root:root:700 ]
(cd "$backup"; sha256sum -c SHA256SUMS)
gzip -t "$backup/element.sql.gz" "$backup/application.tar.gz" "$backup/files-and-system.tar.gz"
(cd "$payload"; sha256sum -c SHA256SUMS)
[ "$(df -B1 --output=avail "$root" | tail -1 | tr -d ' ')" -gt 2147483648 ]
mkdir -m 0700 "$evidence"
printf '%s\n' mng-backend-only-staging-owned-v1 "$payload" > "$evidence/ownership.txt"
cp "$payload/SHA256SUMS" "$evidence/payload-SHA256SUMS"
cp "$payload/check" "$evidence/checker"
cp "$payload/check-source.go.txt" "$evidence/checker-source.go.txt"
cp "$payload/deploy.sh" "$evidence/deploy-executed.sh"
legacy > "$evidence/legacy-before.sha256"
pdfs > "$evidence/pdf-before.sha256"
immutable > "$evidence/immutable-before.sha256"
signature > "$evidence/schema-before.signature"
[ "$(wc -l < "$evidence/pdf-before.sha256")" = 465 ]
stat -c '%u %g %a' "$root/server" > "$evidence/server.metadata.before"
sha256sum "$root/server" > "$evidence/prior-server.sha256"
cp -a "$root/server" "$evidence/server.before"
mysqldump --single-transaction --quick --routines --triggers --events --set-gtid-purged=OFF --no-tablespaces element | gzip > "$evidence/element-current.sql.gz"
gzip -t "$evidence/element-current.sql.gz"
(cd "$evidence"; sha256sum element-current.sql.gz > current-db-SHA256SUMS; sha256sum -c current-db-SHA256SUMS)
chmod 0600 "$evidence"/*
sudo -u liming "$payload/check" post "$pid" > "$evidence/precheck.log" 2>&1
cat "$evidence/precheck.log"
appuser=$(sed -n 's/.* APP_DB_USER=\([^ ]*\) MNG_TABLES=.*/\1/p' "$evidence/precheck.log")
[[ "$appuser" =~ ^[a-zA-Z0-9_]+$ ]]
stage=prepared; changed=0; success=0
finish() {
 code=$?; trap - EXIT; rollback=not-needed
 if [ "$success" != 1 ] && [ "$changed" = 1 ]; then
  rollback=blocked
  systemctl stop talent-assessment || true
  if empty_sidecars; then
   read -r uid gid mode < "$evidence/server.metadata.before"
   cp "$evidence/server.before" "$root/server.rollback-$stamp"
   chown "$uid:$gid" "$root/server.rollback-$stamp"; chmod "$mode" "$root/server.rollback-$stamp"
   touch -r "$evidence/server.before" "$root/server.rollback-$stamp"
   mv "$root/server.rollback-$stamp" "$root/server"
   if systemctl start talent-assessment; then rollback=restored-previous-backend; fi
  fi
 fi
 printf 'BACKEND_DEPLOYED=%s EXIT=%s STAGE=%s ROLLBACK=%s DDL_EXECUTIONS=0 FRONTEND_CHANGES=0\n' "$success" "$code" "$stage" "$rollback" | tee "$evidence/deploy-receipt.txt"
 printf 'BACKEND_EVIDENCE=%s\n' "$evidence"
 exit "$code"
}
trap finish EXIT
trap 'exit 130' HUP INT TERM
stage=drain; changed=1
cg=$(systemctl show talent-assessment -p ControlGroup --value)
systemctl stop talent-assessment
[ "$(systemctl show talent-assessment -p MainPID --value)" = 0 ] && [ ! -e "/proc/$pid" ]
if [ -n "$cg" ] && [ -f "/sys/fs/cgroup$cg/cgroup.procs" ]; then [ -z "$(cat "/sys/fs/cgroup$cg/cgroup.procs")" ]; fi
if ss -lntp | grep -q ':8092 '; then exit 1; fi
[ "$(q -e "SELECT COUNT(*) FROM information_schema.processlist WHERE user='$appuser'")" = 0 ]
printf 'DRAIN_PASS_PID=%s CONTROL_GROUP_EMPTY=1 LISTENER_8092=0 APP_DB_CONNECTIONS=0\n' "$pid" | tee "$evidence/drain.log"
legacy > "$evidence/legacy-drained.sha256"; cmp "$evidence/legacy-before.sha256" "$evidence/legacy-drained.sha256"
empty_sidecars
stage=install
install -o liming -g liming -m 0755 "$payload/server" "$root/server.mng-new-$stamp"
mv "$root/server.mng-new-$stamp" "$root/server"
"$payload/check" gate - > "$evidence/gate-stopped.log" 2>&1
cat "$evidence/gate-stopped.log"
systemctl start talent-assessment
newpid=$(systemctl show talent-assessment -p MainPID --value)
curl --fail --silent --retry 5 --retry-connrefused --max-time 10 http://127.0.0.1:8092/health > "$evidence/health.json"
stage=postcheck
sudo -u liming "$payload/check" post "$newpid" > "$evidence/postcheck.log" 2>&1
cat "$evidence/postcheck.log"
legacy > "$evidence/legacy-after.sha256"; pdfs > "$evidence/pdf-after.sha256"
immutable > "$evidence/immutable-after.sha256"; signature > "$evidence/schema-final.signature"
for pair in legacy:sha256 pdf:sha256 immutable:sha256; do name=${pair%:*}; cmp "$evidence/$name-before.sha256" "$evidence/$name-after.sha256"; done
cmp "$evidence/schema-before.signature" "$evidence/schema-final.signature"
empty_sidecars
[ "$(q element -e "SELECT COUNT(*) FROM information_schema.referential_constraints WHERE constraint_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' AND update_rule='RESTRICT' AND delete_rule='RESTRICT'")" = 15 ]
for s in talent-assessment nginx mysql; do systemctl is-active "$s"; done
sha256sum "$root/server" "$root/dist/index.html" > "$evidence/deployed-SHA256SUMS"
cat "$evidence/deployed-SHA256SUMS"; cat "$evidence/current-db-SHA256SUMS"
printf 'CURRENT_LEGACY_SHA256=%s PDF_COUNT=465 SIDECAR_TABLES=11 ROWS_EACH=0 FKS=15 IMMUTABLE_UNCHANGED=1 PID=%s\n' "$(cat "$evidence/legacy-after.sha256")" "$newpid"
stage=verified; success=1