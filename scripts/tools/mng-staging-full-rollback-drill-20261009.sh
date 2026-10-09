#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT=/opt/talent-assessment
BACKUP=/opt/talent-assessment/backups/mng_current_d0e8202eafb14b08
DRILL="$BACKUP/full-rollback-drill-20261009-attempt3-schema"
OLD_SERVER_SHA=4179fd3f4e9b60e8479f9727c0d3592a7ada188f5ef6ee855fdbd423d944b36c
OLD_INDEX_SHA=52eecf04a77a809816bf11b18d61f23a415ad030957dd67cd6307f714b560860
NEW_SERVER_SHA=d30c8e40e0c99dc525ad42b1269581665dd72acb7e1bfd47958ac0efd1561fc7
NEW_INDEX_SHA=c4f3b8f76b740244bd6b1d10ac6e24e4650f15367f1c427e4b46055297e21555
EXPECTED_ADDITIVE_SCHEMA_SHA=1a9f16e82a3facd55d418d87bf85771c4d6d23eb5edc88c531b8d96c30ddbe27
HISTORICAL_REHEARSAL_SCHEMA_SHA=1ebadfd6818723bc83465ed4c85d43ff0a9da6e5f397e1e53537ef082ae2ec3c
FROZEN_EXAM=1791298091700970647
manual_restarts=0
changed=0
restored_new=0
old_pid=0
rollback_pid=0
final_pid=0
auth_uuid=""
auth_header=""

fail() { printf 'FAIL=%s\n' "$1" >&2; return 1; }
scalar() { mysql --batch --skip-column-names element -e "$1"; }
sha() { sha256sum "$1" | cut -d' ' -f1; }
front_manifest() {
  local dir=$1
  (cd "$dir" && find . -type f -print0 | sort -z | xargs -0 sha256sum) | sha256sum | cut -d' ' -f1
}
config_manifest() {
  (find "$ROOT/configs" /etc/systemd/system/talent-assessment.service /etc/systemd/system/talent-assessment.service.d -type f -print0 | sort -z | xargs -0 sha256sum) | sha256sum | cut -d' ' -f1
}
protected_db_sha() {
  local tables
  tables=$(mysql --batch --skip-column-names element -e "SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' AND table_name NOT IN ('el_mng_exam_draft','el_mng_report_reissue','el_mng_reissue_audit') ORDER BY table_name")
  mysqldump --single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary element $tables | sha256sum | cut -d' ' -f1
}
additive_schema_receipt() {
  mysql --batch --skip-column-names --raw element <<'SQL'
SELECT CONCAT('T|',HEX(TABLE_NAME),'|',HEX(ENGINE),'|',IFNULL(HEX(TABLE_COLLATION),'N')) FROM information_schema.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN ('el_mng_exam_draft','el_mng_report_reissue','el_mng_reissue_audit') ORDER BY TABLE_NAME;
SELECT CONCAT('C|',HEX(TABLE_NAME),'|',LPAD(ORDINAL_POSITION,6,'0'),'|',HEX(COLUMN_NAME),'|',HEX(COLUMN_TYPE),'|',IS_NULLABLE,'|',IF(COLUMN_DEFAULT IS NULL,'N',CONCAT('V',HEX(COLUMN_DEFAULT))),'|',HEX(EXTRA)) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN ('el_mng_exam_draft','el_mng_report_reissue','el_mng_reissue_audit') ORDER BY TABLE_NAME,ORDINAL_POSITION;
SELECT CONCAT('I|',HEX(TABLE_NAME),'|',HEX(INDEX_NAME),'|',NON_UNIQUE,'|',LPAD(SEQ_IN_INDEX,6,'0'),'|',HEX(COLUMN_NAME)) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN ('el_mng_exam_draft','el_mng_report_reissue','el_mng_reissue_audit') ORDER BY TABLE_NAME,INDEX_NAME,SEQ_IN_INDEX;
SELECT CONCAT('F|',HEX(k.TABLE_NAME),'|',HEX(k.CONSTRAINT_NAME),'|',LPAD(k.ORDINAL_POSITION,6,'0'),'|',HEX(k.COLUMN_NAME),'|',HEX(k.REFERENCED_TABLE_NAME),'|',HEX(k.REFERENCED_COLUMN_NAME),'|',r.UPDATE_RULE,'|',r.DELETE_RULE) FROM information_schema.KEY_COLUMN_USAGE k JOIN information_schema.REFERENTIAL_CONSTRAINTS r ON r.CONSTRAINT_SCHEMA=k.CONSTRAINT_SCHEMA AND r.TABLE_NAME=k.TABLE_NAME AND r.CONSTRAINT_NAME=k.CONSTRAINT_NAME WHERE k.TABLE_SCHEMA=DATABASE() AND k.TABLE_NAME IN ('el_mng_exam_draft','el_mng_report_reissue','el_mng_reissue_audit') AND k.REFERENCED_TABLE_NAME IS NOT NULL ORDER BY k.TABLE_NAME,k.CONSTRAINT_NAME,k.ORDINAL_POSITION;
SQL
}
capture_additive_schema() {
  local phase=$1 receipt="$DRILL/$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]').additive-schema.canonical"
  [ "$(scalar "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name IN ('el_mng_exam_draft','el_mng_report_reissue','el_mng_reissue_audit')")" = 3 ] || fail "$phase-additive-schema-table-count"
  additive_schema_receipt > "$receipt"
  chmod 0600 "$receipt"
  local signature
  signature=$(sha256sum "$receipt" | cut -d' ' -f1)
  [ "$signature" = "$EXPECTED_ADDITIVE_SCHEMA_SHA" ] || fail "$phase-additive-schema-expected-mismatch"
  printf -v "${phase}_ADDITIVE_SCHEMA_SHA" '%s' "$signature"
  printf '%s_ADDITIVE_SCHEMA_SHA=%s\n' "$phase" "$signature"
}
pdf_manifest() {
  local dir="$ROOT/private/management-traits-test-reports"
  if [ ! -d "$dir" ]; then printf empty; return; fi
  (find "$dir" -type f -print0 | sort -z | xargs -0 -r sha256sum) | sha256sum | cut -d' ' -f1
}
route_status() {
  local method=$1 url=$2 body=${3-}
  if [ "$method" = POST ]; then
    curl -sS -o /dev/null -w '%{http_code}' -H 'Content-Type: application/json' -d "$body" "http://127.0.0.1:8092$url"
  else
    curl -sS -o /dev/null -w '%{http_code}' "http://127.0.0.1:8092$url"
  fi
}
auth_route_status() {
  local method=$1 url=$2 body=${3-}
  if [ "$method" = POST ]; then
    curl -sS -o /dev/null -w '%{http_code}' -H "@$auth_header" -H 'Content-Type: application/json' -d "$body" "http://127.0.0.1:8092$url"
  else
    curl -sS -o /dev/null -w '%{http_code}' -H "@$auth_header" "http://127.0.0.1:8092$url"
  fi
}
create_ephemeral_admin() {
  auth_uuid="rollbackdrill$(date +%s%N)"
  auth_header=$(mktemp /run/mng-rollback-auth-header.XXXXXX)
  local login_json
  login_json=$(mktemp /run/mng-rollback-login.XXXXXX)
  chmod 0600 "$auth_header" "$login_json"
  python3 - "$ROOT/configs/application-production.yml" "$auth_uuid" "$auth_header" "$login_json" <<'PY'
import base64, hashlib, hmac, json, os, sys, time
config_path, token_id, header_path, login_path = sys.argv[1:]
lines = open(config_path, encoding='utf-8').read().splitlines()
in_jwt = False
secret = None
for line in lines:
    if line and not line.startswith((' ', '\t', '#')):
        in_jwt = line.strip() == 'jwt:'
        continue
    if in_jwt and line.strip().startswith('secret:'):
        secret = line.split(':', 1)[1].strip().strip('"\'')
        break
if not secret:
    raise SystemExit('jwt secret unavailable')
def b64(value):
    return base64.urlsafe_b64encode(value).rstrip(b'=')
head = b64(json.dumps({'alg':'HS512','typ':'JWT'}, separators=(',',':')).encode())
body = b64(json.dumps({'login_user_key':token_id}, separators=(',',':')).encode())
signing = head + b'.' + body
signature = b64(hmac.new(secret.encode(), signing, hashlib.sha512).digest())
token = (signing + b'.' + signature).decode()
now = int(time.time() * 1000)
login = {'userId':1,'deptId':None,'token':token_id,'loginTime':now,'expireTime':now+120000,'ipaddr':'127.0.0.1','browser':'rollback-drill','os':'linux','permissions':['*:*:*'],'roles':['admin'],'user':None}
with open(header_path, 'w', encoding='utf-8', newline='\n') as f:
    f.write('Authorization: Bearer ' + token + '\n')
with open(login_path, 'w', encoding='utf-8', newline='\n') as f:
    json.dump(login, f, ensure_ascii=True, separators=(',',':'))
os.chmod(header_path, 0o600)
os.chmod(login_path, 0o600)
PY
  redis-cli -n 1 -x SETEX "login_tokens:$auth_uuid" 120 < "$login_json" >/dev/null
  rm -f "$login_json"
}
delete_ephemeral_admin() {
  if [ -n "$auth_uuid" ]; then redis-cli -n 1 DEL "login_tokens:$auth_uuid" >/dev/null 2>&1 || true; fi
  if [ -n "$auth_header" ]; then rm -f "$auth_header"; fi
}
assert_services() {
  for service in talent-assessment nginx mysql; do
    [ "$(systemctl is-active "$service")" = active ] || fail "service-$service-not-active"
  done
}
assert_env() {
  local pid=$1
  tr '\0' '\n' < "/proc/$pid/environ" | grep -Fx REPORT_EFFECTIVE_ENV=staging >/dev/null
  tr '\0' '\n' < "/proc/$pid/environ" | grep -Fx MNG_TEST_REPORT_ENV=staging >/dev/null
}
assert_additive_empty() {
  [ "$(scalar "SELECT (SELECT COUNT(*) FROM el_mng_exam_draft)+(SELECT COUNT(*) FROM el_mng_report_reissue)+(SELECT COUNT(*) FROM el_mng_reissue_audit)")" = 0 ] || fail additive-tables-not-empty
}
assert_data_invariants() {
  [ "$(protected_db_sha)" = "$BASE_PROTECTED_DB" ] || fail protected-db-hash-drift
  [ "$(pdf_manifest)" = "$BASE_PDF" ] || fail pdf-hash-drift
  [ "$(scalar "SELECT CAST(overall_score AS CHAR) FROM el_mng_result_run WHERE exam_id='$FROZEN_EXAM' AND status='completed' ORDER BY submitted_at DESC,id DESC LIMIT 1")" = 58.642639 ] || fail frozen-002-score-drift
  assert_additive_empty
}
assert_no_errors_since() {
  local since=$1 nginx_line=$2 phase=$3
  local app_errors nginx_errors
  app_errors=$(journalctl -u talent-assessment --since "@$since" --no-pager 2>/dev/null | grep -Eci 'panic|fatal|segmentation|permission denied' || true)
  nginx_errors=$(tail -n +$((nginx_line+1)) /var/log/nginx/access.log 2>/dev/null | awk '$9 ~ /^5/{n++} END{print n+0}')
  printf '%s_APP_ERRORS=%s\n%s_NGINX_5XX=%s\n' "$phase" "$app_errors" "$phase" "$nginx_errors"
  [ "$app_errors" = 0 ] || fail "$phase-app-errors"
  [ "$nginx_errors" = 0 ] || fail "$phase-nginx-5xx"
}
stop_service() {
  local pid=$1
  systemctl stop talent-assessment
  [ "$(systemctl show talent-assessment -p MainPID --value)" = 0 ]
  [ ! -e "/proc/$pid" ]
  ! ss -lntp | grep -q ':8092 '
}
start_service() {
  systemctl start talent-assessment
  manual_restarts=$((manual_restarts+1))
  started_pid=$(systemctl show talent-assessment -p MainPID --value)
  [ "$started_pid" -gt 0 ]
  curl -fsS --retry 8 --retry-connrefused --max-time 10 http://127.0.0.1:8092/health >/dev/null
}
restore_new_best_effort() {
  set +e
  systemctl stop talent-assessment
  if [ -f "$DRILL/new.server" ]; then
    read -r uid gid mode mtime < "$DRILL/new.server.metadata"
    install -o "$uid" -g "$gid" -m "$mode" "$DRILL/new.server" "$ROOT/server.recovery"
    touch -d "@$mtime" "$ROOT/server.recovery"
    mv -f "$ROOT/server.recovery" "$ROOT/server"
  fi
  if [ -d "$DRILL/dist.swap" ] && [ "$(sha "$DRILL/dist.swap/index.html" 2>/dev/null)" = "$NEW_INDEX_SHA" ]; then
    perl -e 'syscall(316,-100,$ARGV[0],-100,$ARGV[1],2)==0 or die "exchange failed\n"' "$ROOT/dist" "$DRILL/dist.swap"
  fi
  systemctl start talent-assessment
  local pid
  pid=$(systemctl show talent-assessment -p MainPID --value)
  if [ "$pid" -gt 0 ] && [ "$(sha "$ROOT/server" 2>/dev/null)" = "$NEW_SERVER_SHA" ] && [ "$(sha "$ROOT/dist/index.html" 2>/dev/null)" = "$NEW_INDEX_SHA" ]; then restored_new=1; fi
  set -e
}
finish() {
  local rc=$?
  trap - EXIT HUP INT TERM
  delete_ephemeral_admin
  if [ "$rc" -ne 0 ] && [ "$changed" = 1 ] && [ "$restored_new" != 1 ]; then restore_new_best_effort; fi
  printf 'ROLLBACK_DRILL_EXIT=%s RESTORED_NEW_AFTER_FAILURE=%s MANUAL_RESTARTS=%s\n' "$rc" "$restored_new" "$manual_restarts"
  exit "$rc"
}
trap finish EXIT
trap 'exit 130' HUP INT TERM

[ "$(hostname)" = vm-ubuntu-go-dev ]
[ "$(id -u)" = 0 ]
[ -d "$BACKUP" ]
[ "$(stat -c %a "$BACKUP")" = 700 ]
(cd "$BACKUP" && sha256sum -c SHA256SUMS)
[ "$(sha "$BACKUP/server.before")" = "$OLD_SERVER_SHA" ]
[ ! -e "$DRILL" ]
mkdir -m 0700 "$DRILL"

assert_services
old_pid=$(systemctl show talent-assessment -p MainPID --value)
[ "$old_pid" -gt 0 ]
[ "$(sha "$ROOT/server")" = "$NEW_SERVER_SHA" ]
[ "$(sha "/proc/$old_pid/exe")" = "$NEW_SERVER_SHA" ]
[ "$(sha "$ROOT/dist/index.html")" = "$NEW_INDEX_SHA" ]
[ "$(find "$ROOT/dist" -type f | wc -l)" = 393 ]
[ "$(scalar 'SELECT COUNT(*) FROM el_paper WHERE state=1')" = 0 ]
[ "$(scalar "SELECT COUNT(*) FROM el_paper p JOIN el_exam e ON e.id=p.exam_id WHERE p.state=1 AND p.limit_time IS NOT NULL AND p.limit_time<=NOW() AND e.assessment_type='competency'")" = 0 ]
[ "$(scalar "SELECT COUNT(*) FROM el_paper p JOIN el_exam e ON e.id=p.exam_id WHERE p.state=1 AND p.limit_time IS NOT NULL AND p.limit_time<=NOW() AND e.id IN (SELECT exam_id FROM el_mng_exam_profile)")" = 0 ]
[ "$(ss -Htn state established '( sport = :8092 )' | wc -l)" = 0 ]
assert_env "$old_pid"
assert_additive_empty
capture_additive_schema PRE
create_ephemeral_admin

BASE_CONFIG=$(config_manifest)
BASE_PROTECTED_DB=$(protected_db_sha)
BASE_PDF=$(pdf_manifest)
BASE_NEW_FRONT=$(front_manifest "$ROOT/dist")
BASE_NGINX_LINE=$(wc -l < /var/log/nginx/access.log)
BASE_TIME=$(date +%s)
printf 'BASE_PID=%s\nBASE_CONFIG_SHA=%s\nBASE_PROTECTED_DB_SHA=%s\nBASE_PDF_SHA=%s\nBASE_NEW_FRONT_MANIFEST=%s\n' "$old_pid" "$BASE_CONFIG" "$BASE_PROTECTED_DB" "$BASE_PDF" "$BASE_NEW_FRONT"

stat -c '%u %g %a %Y' "$ROOT/server" > "$DRILL/new.server.metadata"
cp -a "$ROOT/server" "$DRILL/new.server"
chmod 0600 "$DRILL/new.server" "$DRILL/new.server.metadata"
mkdir -m 0755 "$DRILL/dist.swap"
tar -xzf "$BACKUP/dist.before.tar.gz" -C "$DRILL/dist.swap" --strip-components=1
chown -R root:root "$DRILL/dist.swap"
chmod -R 0755 "$DRILL/dist.swap"
OLD_FRONT_MANIFEST=$(front_manifest "$DRILL/dist.swap")
[ "$(sha "$DRILL/dist.swap/index.html")" = "$OLD_INDEX_SHA" ]
printf 'OLD_FRONT_FILES=%s\nOLD_FRONT_INDEX_SHA=%s\nOLD_FRONT_MANIFEST=%s\n' "$(find "$DRILL/dist.swap" -type f | wc -l)" "$OLD_INDEX_SHA" "$OLD_FRONT_MANIFEST"

changed=1
rollback_time=$(date +%s)
rollback_nginx_line=$(wc -l < /var/log/nginx/access.log)
stop_service "$old_pid"
read -r old_uid old_gid old_mode old_mtime < "$BACKUP/server.metadata"
install -o "$old_uid" -g "$old_gid" -m "$old_mode" "$BACKUP/server.before" "$ROOT/server.rollback-drill"
touch -d "@$old_mtime" "$ROOT/server.rollback-drill"
mv -f "$ROOT/server.rollback-drill" "$ROOT/server"
perl -e 'syscall(316,-100,$ARGV[0],-100,$ARGV[1],2)==0 or die "exchange failed\n"' "$ROOT/dist" "$DRILL/dist.swap"
start_service
rollback_pid=$started_pid
[ "$rollback_pid" != "$old_pid" ]
[ "$(sha "$ROOT/server")" = "$OLD_SERVER_SHA" ]
[ "$(sha "/proc/$rollback_pid/exe")" = "$OLD_SERVER_SHA" ]
[ "$(sha "$ROOT/dist/index.html")" = "$OLD_INDEX_SHA" ]
[ "$(front_manifest "$ROOT/dist")" = "$OLD_FRONT_MANIFEST" ]
[ "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8092/health)" = 200 ]
assert_services
assert_env "$rollback_pid"
[ "$(config_manifest)" = "$BASE_CONFIG" ]
assert_data_invariants
capture_additive_schema OLD
legacy_detail_old=$(route_status POST /exam/api/exam/exam/detail '{}')
legacy_results_old=$(route_status POST /exam/api/management-traits/results/list '{}')
reissue_old=$(auth_route_status GET '/exam/api/management-traits/report-reissues/qualification?runId=rollback-drill')
printf 'OLD_LEGACY_DETAIL_HTTP=%s\nOLD_LEGACY_RESULTS_HTTP=%s\nOLD_REISSUE_AUTH_HTTP=%s\n' "$legacy_detail_old" "$legacy_results_old" "$reissue_old"
[ "$legacy_detail_old" != 404 ]
[ "$legacy_results_old" != 404 ]
[ "$reissue_old" = 404 ]
assert_no_errors_since "$rollback_time" "$rollback_nginx_line" OLD_PHASE

reapply_time=$(date +%s)
reapply_nginx_line=$(wc -l < /var/log/nginx/access.log)
stop_service "$rollback_pid"
read -r new_uid new_gid new_mode new_mtime < "$DRILL/new.server.metadata"
install -o "$new_uid" -g "$new_gid" -m "$new_mode" "$DRILL/new.server" "$ROOT/server.reapply-drill"
touch -d "@$new_mtime" "$ROOT/server.reapply-drill"
mv -f "$ROOT/server.reapply-drill" "$ROOT/server"
perl -e 'syscall(316,-100,$ARGV[0],-100,$ARGV[1],2)==0 or die "exchange failed\n"' "$ROOT/dist" "$DRILL/dist.swap"
start_service
final_pid=$started_pid
[ "$final_pid" != "$rollback_pid" ]
[ "$final_pid" != "$old_pid" ]
[ "$(sha "$ROOT/server")" = "$NEW_SERVER_SHA" ]
[ "$(sha "/proc/$final_pid/exe")" = "$NEW_SERVER_SHA" ]
[ "$(sha "$ROOT/dist/index.html")" = "$NEW_INDEX_SHA" ]
[ "$(find "$ROOT/dist" -type f | wc -l)" = 393 ]
[ "$(front_manifest "$ROOT/dist")" = "$BASE_NEW_FRONT" ]
[ "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8092/health)" = 200 ]
assert_services
assert_env "$final_pid"
[ "$(config_manifest)" = "$BASE_CONFIG" ]
assert_data_invariants
capture_additive_schema POST
legacy_detail_new=$(route_status POST /exam/api/exam/exam/detail '{}')
legacy_results_new=$(route_status POST /exam/api/management-traits/results/list '{}')
reissue_new=$(route_status GET '/exam/api/management-traits/report-reissues/qualification?runId=rollback-drill')
reissue_new_auth=$(auth_route_status GET '/exam/api/management-traits/report-reissues/qualification?runId=rollback-drill')
printf 'NEW_LEGACY_DETAIL_HTTP=%s\nNEW_LEGACY_RESULTS_HTTP=%s\nNEW_REISSUE_UNAUTH_HTTP=%s\nNEW_REISSUE_AUTH_HTTP=%s\n' "$legacy_detail_new" "$legacy_results_new" "$reissue_new" "$reissue_new_auth"
[ "$legacy_detail_new" != 404 ]
[ "$legacy_results_new" != 404 ]
[ "$reissue_new" = 401 ] || [ "$reissue_new" = 403 ]
[ "$reissue_new_auth" != 404 ]
assert_no_errors_since "$reapply_time" "$reapply_nginx_line" NEW_PHASE
[ "$manual_restarts" = 2 ]
[ "$PRE_ADDITIVE_SCHEMA_SHA" = "$OLD_ADDITIVE_SCHEMA_SHA" ]
[ "$OLD_ADDITIVE_SCHEMA_SHA" = "$POST_ADDITIVE_SCHEMA_SHA" ]

printf 'PID_SEQUENCE=%s,%s,%s\n' "$old_pid" "$rollback_pid" "$final_pid"
printf 'RESTART_COUNT=2\n'
printf 'FINAL_SERVER_SHA=%s\nFINAL_INDEX_SHA=%s\nFINAL_FRONT_FILES=393\n' "$NEW_SERVER_SHA" "$NEW_INDEX_SHA"
printf 'CONFIG_SHA_UNCHANGED=%s\nPROTECTED_DB_SHA_UNCHANGED=%s\nPDF_SHA_UNCHANGED=%s\nFROZEN_002_SCORE=58.642639\nADDITIVE_TABLE_ROWS=0\n' "$BASE_CONFIG" "$BASE_PROTECTED_DB" "$BASE_PDF"
printf 'EXPECTED_ADDITIVE_SCHEMA_SHA=%s\nPRE_ADDITIVE_SCHEMA_SHA=%s\nOLD_ADDITIVE_SCHEMA_SHA=%s\nPOST_ADDITIVE_SCHEMA_SHA=%s\n' "$EXPECTED_ADDITIVE_SCHEMA_SHA" "$PRE_ADDITIVE_SCHEMA_SHA" "$OLD_ADDITIVE_SCHEMA_SHA" "$POST_ADDITIVE_SCHEMA_SHA"
printf 'ADDITIVE_SCHEMA_SIGNATURES_EQUAL=1\nADDITIVE_SCHEMA_EXPECTED_MATCH=1\n'
printf 'HISTORICAL_REHEARSAL_SCHEMA_SHA=%s\nHISTORICAL_REHEARSAL_EXACT_COMPARABLE=0\n' "$HISTORICAL_REHEARSAL_SCHEMA_SHA"
printf 'FULL_ROLLBACK_STATUS=PASS\n'
restored_new=1
