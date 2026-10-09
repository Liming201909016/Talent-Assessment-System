#!/usr/bin/env bash
# Retained, owned debugging copy only; never installs the application or migrations.
set +x
set -euo pipefail
umask 077
stamp=${1:?unique ownership required}
[[ "$stamp" =~ ^[a-f0-9]{16}$ ]]
[ "$(hostname)" = vm-ubuntu-go-dev ] && [ "$(id -u)" = 0 ]
schema=talent_mng005_local_$stamp
account=mngdbg_${stamp:0:12}
root=/opt/talent-assessment
backup=$root/backups/mng005_local_$stamp
q() { mysql --batch --skip-column-names "$@"; }
[ "$(q -e "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name='$schema'")" = 0 ]
[ "$(q -e "SELECT COUNT(*) FROM mysql.user WHERE user='$account'")" = 0 ]
[ ! -e "$backup" ] && [ ! -L "$backup" ]
[ "$(df -B1 --output=avail "$root" | tail -1 | tr -d ' ')" -gt 1073741824 ]
mkdir -m 0700 "$backup"
printf '%s\n' "$stamp" "$schema" "$account" > "$backup/ownership"
exec 3>&1
exec > "$backup/operation.stdout.private" 2> "$backup/operation.stderr.private"
phase=baseline
trap 'rc=$?; if [ "$rc" != 0 ]; then printf "DEBUG_COPY_BLOCKED phase=%s exit=%s retained=1\n" "$phase" "$rc" >&3; fi' EXIT
dump() { mysqldump --single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-triggers --routines=false --events=false "$@"; }
facts() {
  dump --skip-comments --compact --hex-blob --order-by-primary element | sha256sum | cut -d' ' -f1
}
assets() {
  find "$root/dist" "$root/configs" "$root/private" "$root/tmp" /data/uploadPath -type f -print0 | sort -z | xargs -0 -r sha256sum
  sha256sum "$root/server" /etc/systemd/system/talent-assessment.service
  find /etc/systemd/system/talent-assessment.service.d -type f -exec sha256sum {} +
  find /var/spool/libreoffice/uno_packages/cache -printf '%p %u:%g:%m:%s:%T@:%C@\n' | sort
}
pid=$(systemctl show talent-assessment -p MainPID --value)
[ "$pid" -gt 0 ]
facts > "$backup/source-before.sha"
assets > "$backup/assets-before.sha"
q element -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE();SELECT COUNT(*) FROM el_exam;SELECT COUNT(*) FROM el_paper;SELECT COUNT(*) FROM el_repo WHERE code IN ('00501','00502');SELECT COUNT(*),MAX(overall_score) FROM el_mng_result_run WHERE exam_id='1791298091700970647' AND status='completed';" > "$backup/counts-before.private"
phase=backup
dump element | gzip > "$backup/element.sql.gz"
gzip -t "$backup/element.sql.gz"
(cd "$backup"; sha256sum element.sql.gz > SHA256SUMS)
# Reject database switches and qualified FK references before any restore.
dump --no-data --skip-comments element > "$backup/schema.private.sql"
if grep -Ei '^[[:space:]]*((CREATE|DROP|ALTER)[[:space:]]+DATABASE|USE[[:space:]]|.*REFERENCES[[:space:]]+`[^`]+`\.)' "$backup/schema.private.sql"; then exit 41; fi
# The original backup stays immutable. Only the private restoration stream is sanitized.
gzip -dc "$backup/element.sql.gz" | sed -E '/^[[:space:]]*(\/\*![0-9]+[[:space:]]+)?SET[[:space:]]+(@@GLOBAL\.|GLOBAL[[:space:]]|@@SESSION\.SQL_LOG_BIN|@@SQL_LOG_BIN|SQL_LOG_BIN)/Id' > "$backup/restore.private.sql"
if grep -Ei '^[[:space:]]*(\/\*![0-9]+[[:space:]]+)?((CREATE|DROP|ALTER)[[:space:]]+DATABASE|USE[[:space:]]|SET.*(GTID_PURGED|SQL_LOG_BIN|@@GLOBAL)|CREATE[[:space:]]+(DEFINER|TRIGGER|PROCEDURE|FUNCTION|EVENT))' "$backup/restore.private.sql"; then exit 42; fi
phase=restore
q -e "CREATE DATABASE \`$schema\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci"
# Dump FK session directives apply exclusively to this private restore connection.
mysql "$schema" < "$backup/restore.private.sql"
[ "$(q "$schema" -e 'SELECT COUNT(*) FROM information_schema.key_column_usage WHERE table_schema=DATABASE() AND referenced_table_schema IS NOT NULL AND referenced_table_schema<>DATABASE()')" = 0 ]
source_tables=$(q -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='element'")
[ "$(q "$schema" -e 'SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()')" = "$source_tables" ]
phase=account
password=$(openssl rand -hex 24)
grant_schema=${schema//_/\\_}
for account_host in localhost 127.0.0.1; do
  q <<SQL
CREATE USER '$account'@'$account_host' IDENTIFIED BY '$password';GRANT SELECT,INSERT,UPDATE,DELETE,CREATE,ALTER,INDEX,REFERENCES ON \`$grant_schema\`.* TO '$account'@'$account_host';
SQL
done
printf '{"schema":"%s","user":"%s","password":"%s","stamp":"%s"}\n' "$schema" "$account" "$password" "$stamp" > "$backup/credentials.private.json"
phase=source_protection
facts > "$backup/source-after.sha"
assets > "$backup/assets-after.sha"
cmp "$backup/source-before.sha" "$backup/source-after.sha"
cmp "$backup/assets-before.sha" "$backup/assets-after.sha"
[ "$(systemctl show talent-assessment -p MainPID --value)" = "$pid" ]
for svc in talent-assessment nginx mysql; do systemctl is-active --quiet "$svc"; done
curl -fsS --max-time 5 http://127.0.0.1:8092/health > "$backup/main-health.json"
(cd "$backup"; sha256sum --check --status SHA256SUMS)
[ "$(find "$backup" -type f ! -perm 0600 | wc -l)" = 0 ]
sha=$(cut -d' ' -f1 "$backup/SHA256SUMS")
printf '{"status":"COPY_RESTORED","schema":"%s","backupSHA":"%s","tables":%s,"crossSchemaFK":0,"sourceUnchanged":true,"assetsUnchanged":true,"mainPID":%s,"mainApplicationDeployed":false}\n' "$schema" "$sha" "$source_tables" "$pid" > "$backup/receipt.json"
phase=complete
# Sole sensitive transport, captured in memory and DPAPI-encrypted by the local launcher.
cat "$backup/credentials.private.json" >&3
cat "$backup/receipt.json" >&3