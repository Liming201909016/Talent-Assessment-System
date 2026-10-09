#!/usr/bin/env bash
# Production release: additive competency schema + visible-disabled 005 TEST package.
set -Eeuo pipefail
umask 077

stamp=${1:?16 lowercase hex stamp required}
payload=${2:?absolute migration payload directory required}
release=${3:?absolute release archive required}
release_sha=${4:?release archive SHA-256 required}
payload_manifest_sha=${5:?payload manifest SHA-256 required}
[[ "$stamp" =~ ^[a-f0-9]{16}$ ]]
[[ "$payload" =~ ^/opt/talent-assessment/backups/production_migration_payload_20261009_[a-f0-9]{16}$ ]]
[[ "$release" =~ ^/opt/talent-assessment/backups/production_release_20261009_[a-f0-9]{16}\.tar\.gz$ ]]
[[ "$release_sha" =~ ^[a-f0-9]{64}$ ]]
[[ "$payload_manifest_sha" =~ ^[a-f0-9]{64}$ ]]
[ "$payload" = "/opt/talent-assessment/backups/production_migration_payload_20261009_$stamp" ]
[ "$(hostname)" = iZ0yosjdcen2p4Z ]
[ "$(id -u)" = 0 ]

app=/opt/talent-assessment
backup="$app/backups/production_release_backup_20261009_$stamp"
client=/run/production-release-$stamp.cnf
stage=/tmp/production-release-$stamp
phase=preflight
service_stopped=0
data_installed=0
switched=0
rollback_ready=0
db_mutated=0
nginx_stopped=0

cleanup_secret() { rm -f "$client"; }
nginx_active() { pgrep -x nginx >/dev/null; }
nginx_start() { /etc/init.d/nginx start; }
nginx_stop() { /etc/init.d/nginx stop; }
verify_recovery() {
  [ "$(systemctl is-active talent-assessment)" = active ] || return 1
  nginx_active || return 1
  curl --retry 30 --retry-delay 1 --retry-connrefused --connect-timeout 2 --max-time 2 -sS -f http://127.0.0.1:8092/health >/dev/null || return 1
  curl --retry 10 --retry-delay 1 --retry-connrefused --connect-timeout 2 --max-time 2 -sS -f -H 'Host: 39.106.61.48' http://127.0.0.1:8090/prod-api/health >/dev/null || return 1
}
release_failed() {
  code=$?
  trap - ERR
  set +e
  rollback_ok=1
  if [ "$rollback_ready" != 1 ]; then
    recovery_ok=1
    [ "$service_stopped" = 0 ] || systemctl start talent-assessment || recovery_ok=0
    [ "$nginx_stopped" = 0 ] || nginx_start || recovery_ok=0
    verify_recovery || recovery_ok=0
    printf 'PRODUCTION_RELEASE_PREFLIGHT_OR_BACKUP_FAILED_PHASE=%s\nORIGINAL_EXIT=%s\nRECOVERY_OK=%s\n' "$phase" "$code" "$recovery_ok"
    cleanup_secret
    exit 1
  fi
  if [ "$service_stopped" = 0 ]; then
    systemctl stop talent-assessment || rollback_ok=0
    service_stopped=1
  fi
  [ "$(systemctl is-active talent-assessment || true)" = inactive ] || rollback_ok=0
  sessions=$(mysqlq -e "SELECT COUNT(*) FROM information_schema.PROCESSLIST WHERE DB='element' AND ID<>CONNECTION_ID()" 2>>"$backup/private-errors.log") || rollback_ok=0
  [ "${sessions:-unknown}" = 0 ] || rollback_ok=0
  if [ "$rollback_ok" != 1 ]; then
    printf 'PRODUCTION_RELEASE_FAILED_PHASE=%s\nORIGINAL_EXIT=%s\nROLLBACK_OK=0\nMANUAL_RECOVERY_REQUIRED=1\nBACKUP_ROOT=%s\n' "$phase" "$code" "$backup"
    cleanup_secret
    exit 1
  fi
  if [ "$db_mutated" = 1 ]; then
    charset=$(mysqlq -e "SELECT DEFAULT_CHARACTER_SET_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='element'")
    collation=$(mysqlq -e "SELECT DEFAULT_COLLATION_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='element'")
    [[ "$charset" =~ ^[a-z0-9_]+$ && "$collation" =~ ^[a-z0-9_]+$ ]] || rollback_ok=0
    if [ "$rollback_ok" = 1 ]; then
      mysqlq -e "DROP DATABASE \`element\`; CREATE DATABASE \`element\` CHARACTER SET $charset COLLATE $collation;" 2>>"$backup/private-errors.log" || rollback_ok=0
      gzip -dc "$backup/element.sql.gz" | mysql --defaults-extra-file="$client" element >>"$backup/restore.private" 2>>"$backup/private-errors.log" || rollback_ok=0
    fi
  fi
  if [ "$rollback_ok" != 1 ]; then
    printf 'PRODUCTION_RELEASE_FAILED_PHASE=%s\nORIGINAL_EXIT=%s\nROLLBACK_OK=0\nDATABASE_RESTORE_FAILED=1\nMANUAL_RECOVERY_REQUIRED=1\nBACKUP_ROOT=%s\n' "$phase" "$code" "$backup"
    cleanup_secret
    exit 1
  fi
  rm -rf "$app/server" "$app/dist" "$app/configs" "$app/private" "$app/dist.new" "$app/dist.old-$stamp" || rollback_ok=0
  rm -f /etc/systemd/system/talent-assessment.service.d/20-management-traits-production.conf || rollback_ok=0
  if [ "$rollback_ok" != 1 ]; then
    printf 'PRODUCTION_RELEASE_FAILED_PHASE=%s\nORIGINAL_EXIT=%s\nROLLBACK_OK=0\nFILESYSTEM_REMOVE_FAILED=1\nMANUAL_RECOVERY_REQUIRED=1\nBACKUP_ROOT=%s\n' "$phase" "$code" "$backup"
    cleanup_secret
    exit 1
  fi
  if [ -f "$backup/application.tar.gz" ]; then tar -C "$app" -xzf "$backup/application.tar.gz" || rollback_ok=0; fi
  if [ -f "$backup/system.tar.gz" ]; then tar -C / -xzf "$backup/system.tar.gz" || rollback_ok=0; fi
  systemctl daemon-reload || rollback_ok=0
  if [ "$rollback_ok" != 1 ]; then
    printf 'PRODUCTION_RELEASE_FAILED_PHASE=%s\nORIGINAL_EXIT=%s\nROLLBACK_OK=0\nFILESYSTEM_RESTORE_FAILED=1\nMANUAL_RECOVERY_REQUIRED=1\nBACKUP_ROOT=%s\n' "$phase" "$code" "$backup"
    cleanup_secret
    exit 1
  fi
  systemctl start talent-assessment || rollback_ok=0
  service_stopped=0
  if [ "$nginx_stopped" = 1 ]; then nginx_start || rollback_ok=0; nginx_stopped=0; fi
  verify_recovery || rollback_ok=0
  printf 'PRODUCTION_RELEASE_FAILED_PHASE=%s\nORIGINAL_EXIT=%s\nROLLBACK_OK=%s\nBACKUP_ROOT=%s\n' "$phase" "$code" "$rollback_ok" "$backup"
  cleanup_secret
  exit 1
}
trap cleanup_secret EXIT
trap release_failed ERR HUP INT TERM
mkdir -m 0700 "$backup" "$stage"

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
mysqlq() { mysql --defaults-extra-file="$client" --batch --skip-column-names --raw "$@"; }

[ "$(mysqlq -e 'SELECT VERSION()')" = 5.7.44-log ]
[ "$(mysqlq element -e 'SELECT COUNT(*) FROM el_paper WHERE state=1')" = 0 ]
[ "$(mysqlq -e "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='element' AND LEFT(TABLE_NAME,7)='el_mng_'")" = 0 ]
[ "$(mysqlq element -e "SELECT COUNT(*) FROM el_repo WHERE code IN ('00501','00502')")" = 0 ]
[ "$(sha256sum "$app/server" | cut -d' ' -f1)" = 03397e0faf24a21fb6da4e76ba0776226ab87bf2a3c52f72b475eeed4791e44a ]
[ "$(sha256sum "/proc/$(systemctl show talent-assessment -p MainPID --value)/exe" | cut -d' ' -f1)" = 03397e0faf24a21fb6da4e76ba0776226ab87bf2a3c52f72b475eeed4791e44a ]
(
  cd "$payload"
  [ "$(sha256sum SHA256SUMS | cut -d' ' -f1)" = "$payload_manifest_sha" ]
  sha256sum -c SHA256SUMS
  find . -type f ! -path './SHA256SUMS' -printf '%P\n' | sort > "$backup/payload.actual-files"
  sed -E 's#^[a-f0-9]{64}  (\./)?##' SHA256SUMS | sort > "$backup/payload.expected-files"
  cmp "$backup/payload.actual-files" "$backup/payload.expected-files"
  rm -f "$backup/payload.actual-files" "$backup/payload.expected-files"
)
[ "$(sha256sum "$release" | cut -d' ' -f1)" = "$release_sha" ]
tar -xzf "$release" -C "$stage"
(
  cd "$stage"
  sha256sum -c SHA256SUMS
  find . -type f ! -name SHA256SUMS -printf '%P\n' | sort > "$backup/release.actual-files"
  sed -E 's#^[a-f0-9]{64}  (\./)?##' SHA256SUMS | sort > "$backup/release.expected-files"
  cmp "$backup/release.actual-files" "$backup/release.expected-files"
  rm -f "$backup/release.actual-files" "$backup/release.expected-files"
)
[ "$(sha256sum "$stage/server" | cut -d' ' -f1)" = 753fad7a6134139b11ed3285c418da092c160b4fe81b9baf53dbf70f6d3cf0fc ]
[ "$(sha256sum "$stage/dist/index.html" | cut -d' ' -f1)" = abf93dd1fcd6ca6d94a1da393cc492594f6c6d00117152cd987b9c490bbfcbc4 ]
[ -f "$stage/talent-assessment-production-mng005.conf" ]
[ "$(mysqlq -e "SELECT COUNT(*) FROM information_schema.EVENTS WHERE EVENT_SCHEMA='element'")" = 0 ]

phase=backup
printf '%s\n' "production-release-backup-v1:$stamp" > "$backup/ownership"
systemctl show talent-assessment -p MainPID -p NRestarts -p ActiveState > "$backup/service.before"
sha256sum "$app/server" "$app/dist/index.html" > "$backup/runtime.before"
systemctl stop talent-assessment
service_stopped=1
[ "$(systemctl is-active talent-assessment || true)" = inactive ]
nginx_stop
nginx_stopped=1
! nginx_active
[ "$(mysqlq -e "SELECT COUNT(*) FROM information_schema.PROCESSLIST WHERE DB='element' AND ID<>CONNECTION_ID()")" = 0 ]
mysqldump --defaults-extra-file="$client" --single-transaction --quick --skip-lock-tables --routines --triggers --events --hex-blob --set-gtid-purged=OFF --no-tablespaces element | gzip -c > "$backup/element.sql.gz"
[ "$(mysqlq -e "SELECT COUNT(*) FROM information_schema.PROCESSLIST WHERE DB='element' AND ID<>CONNECTION_ID()")" = 0 ]
gzip -t "$backup/element.sql.gz"
assets=(server dist configs)
[ ! -d "$app/private" ] || assets+=(private)
tar -C "$app" -czf "$backup/application.tar.gz" "${assets[@]}"
tar -czf "$backup/system.tar.gz" /etc/systemd/system/talent-assessment.service /etc/systemd/system/talent-assessment.service.d 2>"$backup/system-tar.log" || [ ! -d /etc/systemd/system/talent-assessment.service.d ]
gzip -t "$backup/application.tar.gz" "$backup/system.tar.gz"
find "$backup" -type f -exec chmod 0600 {} +
(cd "$backup"; sha256sum element.sql.gz application.tar.gz system.tar.gz > SHA256SUMS; sha256sum -c SHA256SUMS)
rollback_ready=1

phase=stop
[ "$(systemctl is-active talent-assessment || true)" = inactive ]

phase=schema
db_mutated=1
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
for file in "${schema_files[@]}"; do mysqlq element < "$payload/schema/$file" > "$backup/$file.private"; done

phase=data
mysqlq element < "$payload/mng005/010_mng005_test_baseline_data.sql" > "$backup/data.private"
data_installed=1
mysqlq element < "$payload/mng005/015_mng005_00502_profile_v2.sql" > "$backup/profile-v2.private"
mysqlq element < "$payload/mng005/020_mng005_customer_activation_state.sql" > "$backup/activation.private"
[ "$(mysqlq element -e "SELECT CONCAT_WS('|',(SELECT COUNT(*) FROM el_repo WHERE code IN ('00501','00502')),(SELECT COUNT(*) FROM el_exam WHERE id IN ('9051103000000000501','9051103000000000502') AND state=1),(SELECT COUNT(*) FROM el_qu WHERE id LIKE 'b501-%' OR id LIKE 'b502-%'),(SELECT COUNT(*) FROM el_mng_result_run WHERE exam_id IN ('9051103000000000501','9051103000000000502')))")" = '2|2|280|2' ]

phase=assets
report_root="$app/private/management-traits-test-reports"
mkdir -m 0700 -p "$report_root/reissues" "$app/configs/export-templates"
cp "$payload/mng005/assets/reports/baseline-c51130cb775ba019/"*.pdf "$report_root/"
cp "$payload/mng005/assets/reports/reissues/"*.pdf "$report_root/reissues/"
cp "$payload/mng005/assets/templates/management-traits-002-test-content-v1.xlsx" "$app/configs/export-templates/"
cp "$payload/mng005/assets/templates/management-traits-002-test-only-v2.docx" "$app/configs/export-templates/"
find "$report_root" -type d -exec chmod 0700 {} +
find "$report_root" -type f -exec chmod 0600 {} +
chmod 0644 "$app/configs/export-templates/management-traits-002-test-content-v1.xlsx" "$app/configs/export-templates/management-traits-002-test-only-v2.docx"

phase=switch
install -m 0755 "$stage/server" "$app/server.new"
mv -f "$app/server.new" "$app/server"
rm -rf "$app/dist.new"
cp -a "$stage/dist" "$app/dist.new"
chown -R root:root "$app/dist.new"
chmod -R 0755 "$app/dist.new"
mv "$app/dist" "$app/dist.old-$stamp"
mv "$app/dist.new" "$app/dist"
mkdir -m 0755 -p /etc/systemd/system/talent-assessment.service.d
install -m 0644 "$stage/talent-assessment-production-mng005.conf" /etc/systemd/system/talent-assessment.service.d/20-management-traits-production.conf
systemctl daemon-reload
switched=1

phase=start
systemctl start talent-assessment
service_stopped=0
curl --retry 30 --retry-delay 1 --retry-connrefused --connect-timeout 2 --max-time 2 -sS -f http://127.0.0.1:8092/health >/dev/null
[ "$(systemctl is-active talent-assessment)" = active ]
[ "$(sha256sum "$app/server" | cut -d' ' -f1)" = 753fad7a6134139b11ed3285c418da092c160b4fe81b9baf53dbf70f6d3cf0fc ]
[ "$(sha256sum "/proc/$(systemctl show talent-assessment -p MainPID --value)/exe" | cut -d' ' -f1)" = 753fad7a6134139b11ed3285c418da092c160b4fe81b9baf53dbf70f6d3cf0fc ]
[ "$(sha256sum "$app/dist/index.html" | cut -d' ' -f1)" = abf93dd1fcd6ca6d94a1da393cc492594f6c6d00117152cd987b9c490bbfcbc4 ]
pid=$(systemctl show talent-assessment -p MainPID --value)
[ "$(tr '\0' '\n' < "/proc/$pid/environ" | sed -n 's/^REPORT_EFFECTIVE_ENV=//p')" = production ]
[ "$(tr '\0' '\n' < "/proc/$pid/environ" | sed -n 's/^MNG_TEST_REPORT_ENV=//p')" = production ]
phase=acceptance
[ "$(curl -sS -o /dev/null -w '%{http_code}' 'http://127.0.0.1:8092/exam/api/management-traits/profile/detail?examId=9051103000000000501')" = 401 ]
[ "$(mysqlq element -e "SELECT CONCAT_WS('|',(SELECT COUNT(*) FROM el_repo WHERE code IN ('00501','00502')),(SELECT COUNT(*) FROM el_exam WHERE id IN ('9051103000000000501','9051103000000000502') AND state=1),(SELECT COUNT(*) FROM el_mng_result_run WHERE exam_id IN ('9051103000000000501','9051103000000000502')),(SELECT COUNT(*) FROM el_mng_report_revision WHERE exam_id IN ('9051103000000000501','9051103000000000502')),(SELECT COUNT(*) FROM el_mng_report_reissue WHERE exam_id IN ('9051103000000000501','9051103000000000502')))")" = '2|2|2|2|2' ]
mysqlq element -e "SELECT 'revision',file_key,file_sha,file_bytes FROM el_mng_report_revision WHERE exam_id IN ('9051103000000000501','9051103000000000502') UNION ALL SELECT 'reissue',file_key,file_sha,file_bytes FROM el_mng_report_reissue WHERE exam_id IN ('9051103000000000501','9051103000000000502') ORDER BY 1,2" > "$stage/report-assets.tsv"
[ "$(wc -l < "$stage/report-assets.tsv")" = 4 ]
while IFS=$'\t' read -r kind key expected_sha expected_bytes; do
  file="$report_root/$key"
  [ "$kind" = reissue ] && file="$report_root/reissues/$key"
  [ -f "$file" ] && [ ! -L "$file" ]
  [ "$(sha256sum "$file" | cut -d' ' -f1)" = "$expected_sha" ]
  [ "$(stat -c %s "$file")" = "$expected_bytes" ]
done < "$stage/report-assets.tsv"
[ "$(sha256sum "$app/configs/export-templates/management-traits-002-test-only-v2.docx" | cut -d' ' -f1)" = 05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c ]
[ "$(sha256sum "$app/configs/export-templates/management-traits-002-test-content-v1.xlsx" | cut -d' ' -f1)" = b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c ]
/www/server/nginx/sbin/nginx -t
nginx_start
nginx_stopped=0
nginx_active
[ "$(curl -sS -o /dev/null -w '%{http_code}' -H 'Host: 39.106.61.48' http://127.0.0.1:8090/prod-api/health)" = 200 ]
curl -sS -H 'Host: 39.106.61.48' http://127.0.0.1:8090/ > "$stage/public-index.html"
[ "$(sha256sum "$stage/public-index.html" | cut -d' ' -f1)" = abf93dd1fcd6ca6d94a1da393cc492594f6c6d00117152cd987b9c490bbfcbc4 ]
rm -rf "$app/dist.old-$stamp" "$stage"
printf 'PRODUCTION_RELEASE_PASS=1\nBACKUP_ROOT=%s\nSERVER_SHA=%s\nINDEX_SHA=%s\nMNG_TABLES=%s\nMNG005_REPOS=%s\n' "$backup" "$(sha256sum "$app/server" | cut -d' ' -f1)" "$(sha256sum "$app/dist/index.html" | cut -d' ' -f1)" "$(mysqlq -e "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='element' AND LEFT(TABLE_NAME,7)='el_mng_'")" "$(mysqlq element -e "SELECT COUNT(*) FROM el_repo WHERE code IN ('00501','00502')")"
phase=completed
