#!/usr/bin/env bash
set -euo pipefail
[ "$(hostname)" = vm-ubuntu-go-dev ]
[ "$(id -u)" = 0 ]
ROOT=/opt/talent-assessment
BACKUP="$ROOT/backups/mng_current_d0e8202eafb14b08"
pid=$(systemctl show talent-assessment -p MainPID --value)
sha() { sha256sum "$1" | cut -d' ' -f1; }
front_manifest() { (cd "$1" && find . -type f -print0 | sort -z | xargs -0 sha256sum) | sha256sum | cut -d' ' -f1; }
config_manifest() { (find "$ROOT/configs" /etc/systemd/system/talent-assessment.service /etc/systemd/system/talent-assessment.service.d -type f -print0 | sort -z | xargs -0 sha256sum) | sha256sum | cut -d' ' -f1; }
protected_db_sha() { tables=$(mysql -NBe "SELECT table_name FROM information_schema.tables WHERE table_schema='element' AND LEFT(table_name,7)='el_mng_' AND table_name NOT IN ('el_mng_exam_draft','el_mng_report_reissue','el_mng_reissue_audit') ORDER BY table_name"); mysqldump --single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary element $tables | sha256sum | cut -d' ' -f1; }
pdf_manifest() { (find "$ROOT/private/management-traits-test-reports" -type f -print0 | sort -z | xargs -0 -r sha256sum) | sha256sum | cut -d' ' -f1; }
for service in talent-assessment nginx mysql; do [ "$(systemctl is-active "$service")" = active ]; printf 'SERVICE_%s=active\n' "$service"; done
[ "$(sha "$ROOT/server")" = d30c8e40e0c99dc525ad42b1269581665dd72acb7e1bfd47958ac0efd1561fc7 ]
[ "$(sha "/proc/$pid/exe")" = d30c8e40e0c99dc525ad42b1269581665dd72acb7e1bfd47958ac0efd1561fc7 ]
[ "$(sha "$ROOT/dist/index.html")" = c4f3b8f76b740244bd6b1d10ac6e24e4650f15367f1c427e4b46055297e21555 ]
[ "$(find "$ROOT/dist" -type f | wc -l)" = 393 ]
[ "$(front_manifest "$ROOT/dist")" = cfe77d952c8cb07a07518c92f3d233b2cf4d0e5d446d696f21ca36d16e0e6139 ]
[ "$(config_manifest)" = ed9837af63bce10ac4005a6e56fafc4565ff267e7a988078e292b63e7da4d53f ]
[ "$(protected_db_sha)" = 5e80140dbc2308c60f38b2f485941f91a83b113fea62f20a6faaed974b9a8645 ]
[ "$(pdf_manifest)" = 3d209586a67cce13c4a93ed34c43ceb9f8a0b11d64aa8549131d726b5bd75066 ]
[ "$(mysql -NBe 'SELECT COUNT(*) FROM element.el_paper WHERE state=1')" = 0 ]
[ "$(mysql -NBe "SELECT COUNT(*) FROM element.el_paper p JOIN element.el_exam e ON e.id=p.exam_id WHERE p.state=1 AND p.limit_time IS NOT NULL AND p.limit_time<=NOW() AND e.assessment_type='competency'")" = 0 ]
[ "$(mysql -NBe "SELECT COUNT(*) FROM element.el_paper p JOIN element.el_exam e ON e.id=p.exam_id WHERE p.state=1 AND p.limit_time IS NOT NULL AND p.limit_time<=NOW() AND e.id IN (SELECT exam_id FROM element.el_mng_exam_profile)")" = 0 ]
[ "$(mysql -NBe 'SELECT (SELECT COUNT(*) FROM element.el_mng_exam_draft)+(SELECT COUNT(*) FROM element.el_mng_report_reissue)+(SELECT COUNT(*) FROM element.el_mng_reissue_audit)')" = 0 ]
[ "$(mysql -NBe "SELECT CAST(overall_score AS CHAR) FROM element.el_mng_result_run WHERE exam_id='1791298091700970647' AND status='completed' ORDER BY submitted_at DESC,id DESC LIMIT 1")" = 58.642639 ]
tr '\0' '\n' < "/proc/$pid/environ" | grep -Fx REPORT_EFFECTIVE_ENV=staging >/dev/null
tr '\0' '\n' < "/proc/$pid/environ" | grep -Fx MNG_TEST_REPORT_ENV=staging >/dev/null
[ "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8092/health)" = 200 ]
[ "$(curl -sS -o /dev/null -w '%{http_code}' http://20.200.136.133/prod-api/health)" = 200 ]
[ "$(find /tmp -maxdepth 1 -type d -name 'mng_current_*' | wc -l)" = 0 ]
[ "$(find "$ROOT" -path "$ROOT/backups" -prune -o -type f \( -name '*.test' -o -name 'reissue-service.test' \) -print | wc -l)" = 0 ]
backup_test_binaries=$(find "$ROOT/backups" -type f \( -name '*.test' -o -name 'reissue-service.test' \) -print | wc -l)
[ "$(find "$ROOT/backups" -type f \( -name '*.test' -o -name 'reissue-service.test' \) ! -perm 0600 -print | wc -l)" = 0 ]
while IFS= read -r top; do [ "$(stat -c %a "$ROOT/backups/$top")" = 700 ]; done < <(find "$ROOT/backups" -type f \( -name '*.test' -o -name 'reissue-service.test' \) -printf '%P\n' | cut -d/ -f1 | sort -u)
[ "$(find /run -maxdepth 1 -type f -name 'mng-rollback-*' | wc -l)" = 0 ]
[ "$(redis-cli -n 1 --scan --pattern 'login_tokens:rollbackdrill*' | wc -l)" = 0 ]
[ "$(stat -c %a "$BACKUP")" = 700 ]
[ "$(stat -c %a "$BACKUP/full-rollback-drill-20261009")" = 700 ]
[ "$(stat -c %a "$BACKUP/full-rollback-drill-20261009-attempt2")" = 700 ]
app_errors=$(journalctl -u talent-assessment --since '-10 minutes' --no-pager 2>/dev/null | grep -Eci 'panic|fatal|segmentation|permission denied' || true)
nginx_5xx=$(tail -n 1500 /var/log/nginx/access.log 2>/dev/null | awk '$9 ~ /^5/{n++} END{print n+0}')
[ "$app_errors" = 0 ]
[ "$nginx_5xx" = 0 ]
printf 'FINAL_PID=%s\nNRESTARTS=%s\nSERVER_SHA=%s\nINDEX_SHA=%s\nFRONT_FILES=393\nFRONT_MANIFEST=%s\nCONFIG_SHA=%s\nPROTECTED_DB_SHA=%s\nPDF_SHA=%s\nFROZEN_002_SCORE=58.642639\nSTATE1=0\nOVERDUE_COMPETENCY=0\nOVERDUE_MNG=0\nADDITIVE_ROWS=0\nPAYLOAD_DIRS=0\nRUNTIME_TEST_BINARIES=0\nRESTRICTED_BACKUP_TEST_BINARIES=%s\nBACKUP_TEST_BINARY_MODE_ERRORS=0\nEPHEMERAL_AUTH_RESIDUE=0\nAPP_ERRORS=%s\nNGINX_5XX=%s\nPOST_DRILL_STATUS=PASS\n' "$pid" "$(systemctl show talent-assessment -p NRestarts --value)" "$(sha "$ROOT/server")" "$(sha "$ROOT/dist/index.html")" "$(front_manifest "$ROOT/dist")" "$(config_manifest)" "$(protected_db_sha)" "$(pdf_manifest)" "$backup_test_binaries" "$app_errors" "$nginx_5xx"
