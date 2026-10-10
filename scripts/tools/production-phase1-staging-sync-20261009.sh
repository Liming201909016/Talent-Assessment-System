#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT=/opt/talent-assessment
SERVICE=talent-assessment
EXPECTED_HOST=iZ0yosjdcen2p4Z
PAYLOAD=${1:?payload directory required}
STAMP=${2:?release stamp required}
BACKUP="$ROOT/backups/phase1_staging_sync_${STAMP}"
CLIENT=/run/phase1-staging-sync.cnf
WORK=/tmp/phase1-staging-sync-work-${STAMP}
OLD_SERVER_SHA=f850575b1dfa6eac3f5b4533148baf715eabc8afa13ccd32c7655d0507a7d600
OLD_INDEX_SHA=abf93dd1fcd6ca6d94a1da393cc492594f6c6d00117152cd987b9c490bbfcbc4
NEW_SERVER_SHA=f850575b1dfa6eac3f5b4533148baf715eabc8afa13ccd32c7655d0507a7d600
NEW_INDEX_SHA=593d4a20d890d73bf47f9a519a22fffbdb4b539afd5e1cc1dcc2fdebff73cfde
V1_TEMPLATE_SHA=54b167fcc02737ba44e0432a930e1f5657db57cec87d9063f1ee7c111c4932bf
V2_TEMPLATE_SHA=a814c36e3759c8ff2cf5c17148e0939b477f17d4d3f430bb6a22f530022f9f5a
DIMENSION_SHA=ac6d4290d450506897cb9935ab98bdd0ad67c457a406ee778f0029e906e5d282
QUESTION_SHA=db1554e787634c0cc18cacfeb41db633f1827b8ade46c58114eb226f0675685b
V1_TEXT_SHA=c52b2b195c373cbe87fe0c6826bc1ebf8858e3a028eaa68caa1c5dc2751ba41f
V1_PACKAGE_SHA=85b4df6a90ff0e679df9f0574d7714a9814ceb7b28c6dad1c4dc565942520e0e
DIMENSION_IDS="'competency-a1-01','competency-a1-02','competency-a1-03','competency-a1-04','competency-a1-05','competency-b1-01','competency-b1-02','competency-b1-03','competency-b1-04','competency-b1-05'"
mutating=0
completed=0

cleanup() {
  rm -f "$CLIENT"
  rm -rf "$WORK"
  if [ -d "$PAYLOAD" ]; then rm -rf "$PAYLOAD"; fi
}

rollback() {
  local rc=$?
  if [ "$mutating" -eq 1 ] && [ "$completed" -eq 0 ]; then
    set +e
    systemctl stop "$SERVICE"
    cp -a "$BACKUP/server" "$ROOT/server"
    rm -rf "$ROOT/dist"
    mkdir -p "$ROOT/dist"
    tar -xzf "$BACKUP/dist.tar.gz" -C "$ROOT/dist"
    rm -rf "$ROOT/configs/export-templates"
    cp -a "$BACKUP/export-templates" "$ROOT/configs/export-templates"
    gunzip -c "$BACKUP/element.sql.gz" | mysql --defaults-extra-file="$CLIENT" element
    chown -R root:root "$ROOT/server" "$ROOT/dist" "$ROOT/configs/export-templates"
    chmod 755 "$ROOT/server"
    find "$ROOT/dist" -type d -exec chmod 755 {} +
    find "$ROOT/dist" -type f -exec chmod 644 {} +
    systemctl start "$SERVICE"
    echo 'ROLLBACK_ATTEMPTED=1'
    set -e
  fi
  cleanup
  exit "$rc"
}
trap rollback EXIT HUP INT TERM

[ "$(hostname)" = "$EXPECTED_HOST" ]
[ "$(id -u)" -eq 0 ]
[ "$(systemctl is-active "$SERVICE")" = active ]
[ ! -e "$BACKUP" ]
[ -d "$PAYLOAD" ]
for f in server staging-dist.tar.gz v1-template.docx v2-template.docx v1-text.sql v1-package.sql; do [ -f "$PAYLOAD/$f" ]; done
[ "$(sha256sum "$PAYLOAD/server" | cut -d' ' -f1)" = "$NEW_SERVER_SHA" ]
[ "$(sha256sum "$PAYLOAD/v1-template.docx" | cut -d' ' -f1)" = "$V1_TEMPLATE_SHA" ]
[ "$(sha256sum "$PAYLOAD/v2-template.docx" | cut -d' ' -f1)" = "$V2_TEMPLATE_SHA" ]
[ "$(sha256sum "$ROOT/server" | cut -d' ' -f1)" = "$OLD_SERVER_SHA" ]
[ "$(sha256sum "$ROOT/dist/index.html" | cut -d' ' -f1)" = "$OLD_INDEX_SHA" ]

python3 - "$CLIENT" <<'PY'
from pathlib import Path
import re, sys
out=Path(sys.argv[1])
text='\n'.join(p.read_text(encoding='utf-8-sig').replace('\r','') for p in (Path('/opt/talent-assessment/configs/application.yml'),Path('/opt/talent-assessment/configs/application-production.yml')) if p.exists())
dsn=''
for line in text.splitlines():
    if line.strip().startswith('dsn:'): dsn=line.split(':',1)[1].strip().strip('"\'')
m=re.fullmatch(r'([^:]+):([^@]*)@tcp\(([^:)]+):(\d+)\)/([^?]+)(?:\?.*)?',dsn)
if not m or m.group(5)!='element': raise SystemExit('DB_DSN_INVALID')
u,p,h,port,_=m.groups()
def esc(v): return v.replace('\\','\\\\').replace('"','\\"')
out.write_text('[client]\nuser="%s"\npassword="%s"\nhost="%s"\nport=%s\n'%(esc(u),esc(p),esc(h),port),encoding='utf-8')
out.chmod(0o600)
PY
mysqlq(){ mysql --defaults-extra-file="$CLIENT" --batch --skip-column-names --raw element -e "$1"; }
sha_stream(){ sha256sum | cut -d' ' -f1; }
mysqlsha(){ local output; output=$(mysqlq "$1") || return 1; printf '%s' "$output" | sha_stream; }

pid=$(systemctl show "$SERVICE" -p MainPID --value); [ "$pid" -gt 0 ]
[ "$(sha256sum "/proc/$pid/exe" | cut -d' ' -f1)" = "$OLD_SERVER_SHA" ]
[ "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8092/health)" = 200 ]
[ "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8090/prod-api/health)" = 200 ]
[ "$(mysqlq 'SELECT COUNT(*) FROM el_paper WHERE state=1')" = 0 ]
[ "$(mysqlq "SELECT COUNT(*) FROM el_exam WHERE competency_product_version IN ('competency-frontline-phase1-v1','competency-frontline-phase1-v2')")" = 0 ]
[ "$(mysqlq "SELECT COUNT(*) FROM el_competency_dimension WHERE id IN ($DIMENSION_IDS)")" = 10 ]
[ "$(mysqlq "SELECT COUNT(*) FROM el_qu WHERE dimension_id IN ($DIMENSION_IDS)")" = 90 ]
[ "$(mysqlq "SELECT CONCAT_WS('|',COUNT(*),COUNT(DISTINCT question_code),SUM(competency_question_type='dimension'),SUM(competency_question_type='validity'),SUM(competency_question_type='dimension' AND scoring_direction='forward'),SUM(competency_question_type='dimension' AND scoring_direction='reverse'),SUM(competency_question_type='validity' AND scoring_direction='forward'),SUM(question_status=0)) FROM el_qu WHERE dimension_id IN ($DIMENSION_IDS)")" = '90|90|80|10|62|18|10|90' ]
[ "$(mysqlq "SELECT CONCAT_WS('|',HEX(id),HEX(code),HEX(name),HEX(vird_level),HEX(applicable_category),HEX(core_meaning),display_order,status) FROM el_competency_dimension WHERE id IN ($DIMENSION_IDS) ORDER BY display_order,id" | sha_stream)" = "$DIMENSION_SHA" ]
[ "$(mysqlq "SELECT CONCAT_WS('|',HEX(question_code),HEX(dimension_id),HEX(competency_question_type),dimension_item_no,HEX(content),HEX(observation_point),HEX(scoring_direction),question_status,HEX(COALESCE(remark,''))) FROM el_qu WHERE dimension_id IN ($DIMENSION_IDS) ORDER BY question_code" | sha_stream)" = "$QUESTION_SHA" ]
[ "$(grep -c '^INSERT INTO `el_competency_report_text`' "$PAYLOAD/v1-text.sql")" = 66 ]
[ "$(grep -c '^INSERT INTO `el_competency_report_content_package`' "$PAYLOAD/v1-package.sql")" = 1 ]

rm -rf "$WORK"; install -d -m 700 "$WORK/new-dist"
tar -xzf "$PAYLOAD/staging-dist.tar.gz" -C "$WORK/new-dist"
new_index=$(find "$WORK/new-dist" -type f -name index.html -print -quit); [ -n "$new_index" ]
[ "$(sha256sum "$new_index" | cut -d' ' -f1)" = "$NEW_INDEX_SHA" ]
[ "$(find "$WORK/new-dist" -type f | wc -l | tr -d ' ')" = 393 ]

install -d -m 700 "$BACKUP"
mysqldump --defaults-extra-file="$CLIENT" --single-transaction --routines --triggers --events --add-drop-table element | gzip -9 > "$BACKUP/element.sql.gz"
gzip -t "$BACKUP/element.sql.gz"
cp -a "$ROOT/server" "$BACKUP/server"
tar -czf "$BACKUP/dist.tar.gz" -C "$ROOT/dist" .
cp -a "$ROOT/configs/export-templates" "$BACKUP/export-templates"
sha256sum "$BACKUP/element.sql.gz" "$BACKUP/server" "$BACKUP/dist.tar.gz" > "$BACKUP/SHA256SUMS"
find "$BACKUP" -type d -exec chmod 700 {} +
find "$BACKUP" -type f -exec chmod 600 {} +

{
  echo 'SET NAMES utf8mb4;'
  echo 'BEGIN;'
  echo "DELETE FROM el_competency_report_text WHERE content_version='competency-phase1-content-v1';"
  echo "DELETE FROM el_competency_report_content_package WHERE content_version='competency-phase1-content-v1';"
  cat "$PAYLOAD/v1-text.sql"
  cat "$PAYLOAD/v1-package.sql"
  echo "UPDATE el_competency_report_content_package SET effective_environment='production', update_time=NOW() WHERE content_version='competency-phase1-content-v1' AND approval_status='approved';"
  echo 'COMMIT;'
} > "$WORK/import.sql"
chmod 600 "$WORK/import.sql"

mutating=1
systemctl stop "$SERVICE"
install -m 755 "$PAYLOAD/server" "$ROOT/server"
rm -rf "$ROOT/dist.next"; mv "$WORK/new-dist" "$ROOT/dist.next"
rm -rf "$ROOT/dist"; mv "$ROOT/dist.next" "$ROOT/dist"
install -m 644 "$PAYLOAD/v1-template.docx" "$ROOT/configs/export-templates/competency-phase1-report.docx"
install -m 644 "$PAYLOAD/v2-template.docx" "$ROOT/configs/export-templates/competency-phase1-report-v2.docx"
mysql --defaults-extra-file="$CLIENT" element < "$WORK/import.sql"
chown -R root:root "$ROOT/server" "$ROOT/dist" "$ROOT/configs/export-templates"
chmod 755 "$ROOT/server"
find "$ROOT/dist" -type d -exec chmod 755 {} +
find "$ROOT/dist" -type f -exec chmod 644 {} +
systemctl start "$SERVICE"

for _ in $(seq 1 30); do [ "$(systemctl is-active "$SERVICE")" = active ] && curl -fsS http://127.0.0.1:8092/health >/dev/null && break; done
[ "$(systemctl is-active "$SERVICE")" = active ]
pid=$(systemctl show "$SERVICE" -p MainPID --value); [ "$pid" -gt 0 ]
[ "$(sha256sum "$ROOT/server" | cut -d' ' -f1)" = "$NEW_SERVER_SHA" ]
[ "$(sha256sum "/proc/$pid/exe" | cut -d' ' -f1)" = "$NEW_SERVER_SHA" ]
[ "$(sha256sum "$ROOT/dist/index.html" | cut -d' ' -f1)" = "$NEW_INDEX_SHA" ]
[ "$(find "$ROOT/dist" -type f | wc -l | tr -d ' ')" = 393 ]
[ "$(sha256sum "$ROOT/configs/export-templates/competency-phase1-report.docx" | cut -d' ' -f1)" = "$V1_TEMPLATE_SHA" ]
[ "$(sha256sum "$ROOT/configs/export-templates/competency-phase1-report-v2.docx" | cut -d' ' -f1)" = "$V2_TEMPLATE_SHA" ]
[ "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8092/health)" = 200 ]
[ "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8090/prod-api/health)" = 200 ]
[ "$(mysqlq "SELECT COUNT(*) FROM el_competency_report_text WHERE content_version='competency-phase1-content-v1' AND is_temporary=0 AND status=0")" = 66 ]
[ "$(mysqlq "SELECT COUNT(*) FROM el_competency_report_content_package WHERE content_version='competency-phase1-content-v1' AND approval_status='approved' AND effective_environment='production'")" = 1 ]
[ "$(mysqlsha "SELECT CONCAT_WS('|',HEX(t.id),HEX(t.content_type),HEX(COALESCE(t.dimension_id,'')),HEX(COALESCE(t.level_code,'')),HEX(t.content),HEX(COALESCE(t.disclaimer,'')),t.is_temporary,t.status) FROM el_competency_report_text AS t WHERE t.content_version='competency-phase1-content-v1' ORDER BY t.id")" = "$V1_TEXT_SHA" ]
[ "$(mysqlsha "SELECT CONCAT_WS('|',HEX(p.product_version),HEX(p.scoring_version),HEX(p.content_version),HEX(p.template_version),HEX(p.audience),HEX(p.approval_status),HEX(COALESCE(p.question_source_sha256,'')),HEX(COALESCE(p.content_source_sha256,'')),HEX(COALESCE(p.disclaimer,''))) FROM el_competency_report_content_package AS p WHERE p.content_version='competency-phase1-content-v1' ORDER BY p.id")" = "$V1_PACKAGE_SHA" ]
[ "$(mysqlq 'SELECT COUNT(*) FROM el_paper WHERE state=1')" = 0 ]

completed=1
trap cleanup EXIT HUP INT TERM
printf 'BACKUP=%s\n' "$BACKUP"
printf 'PID=%s\n' "$pid"
printf 'SERVER_SHA=%s\nINDEX_SHA=%s\nV1_TEMPLATE_SHA=%s\nV2_TEMPLATE_SHA=%s\n' "$NEW_SERVER_SHA" "$NEW_INDEX_SHA" "$V1_TEMPLATE_SHA" "$V2_TEMPLATE_SHA"
printf 'PHASE1_STAGING_SYNC=PASS\n'
