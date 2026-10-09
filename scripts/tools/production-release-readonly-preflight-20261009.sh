#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT=/opt/talent-assessment
CLIENT=/run/talent-assessment-production-preflight.cnf
cleanup() { rm -f "$CLIENT"; }
trap cleanup EXIT HUP INT TERM

[ "$(hostname)" = iZ0yosjdcen2p4Z ]
[ "$(id -u)" = 0 ]
[ -f "$ROOT/configs/application-production.yml" ]

python3 - "$CLIENT" <<'PY'
from pathlib import Path
import re, sys
client = Path(sys.argv[1])
paths = [Path('/opt/talent-assessment/configs/application.yml'), Path('/opt/talent-assessment/configs/application-production.yml')]
text = '\n'.join(p.read_text(encoding='utf-8-sig').replace('\r', '') for p in paths if p.exists())
dsn = None
for line in text.splitlines():
    if line.strip().startswith('dsn:'):
        dsn = line.split(':', 1)[1].strip().strip('"\'')
if not dsn:
    raise SystemExit('DB_DSN_UNAVAILABLE')
match = re.fullmatch(r'([^:]+):([^@]*)@tcp\(([^:)]+):(\d+)\)/([^?]+)(?:\?.*)?', dsn)
if not match:
    raise SystemExit('DB_DSN_FORMAT_UNSUPPORTED')
user, password, host, port, database = match.groups()
if database != 'element':
    raise SystemExit('DB_NAME_UNEXPECTED')
client.write_text('[client]\nuser=' + user + '\npassword=' + password + '\nhost=' + host + '\nport=' + port + '\n', encoding='utf-8')
client.chmod(0o600)
PY

mysqlq() { mysql --defaults-extra-file="$CLIENT" --batch --skip-column-names --raw element -e "$1"; }
pid=$(systemctl show talent-assessment -p MainPID --value)
[ "$pid" -gt 0 ]

printf 'HOST=%s\nUSER=%s\nPID=%s\nNRESTARTS=%s\nSERVICE=%s\n' "$(hostname)" "$(id -un)" "$pid" "$(systemctl show talent-assessment -p NRestarts --value)" "$(systemctl is-active talent-assessment)"
printf 'HEALTH_8092=%s\nHEALTH_8090=%s\nHTTP_80=%s\n' "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8092/health)" "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8090/prod-api/health)" "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1/)"
printf 'SERVER_SHA=%s\nPROCESS_SHA=%s\nINDEX_SHA=%s\nFRONT_FILES=%s\n' "$(sha256sum "$ROOT/server" | cut -d' ' -f1)" "$(sha256sum "/proc/$pid/exe" | cut -d' ' -f1)" "$(sha256sum "$ROOT/dist/index.html" | cut -d' ' -f1)" "$(find "$ROOT/dist" -type f | wc -l)"
printf 'MYSQL=%s\nTABLES=%s\nROOT_AVAIL=%s\nMEM_AVAIL_KB=%s\n' "$(mysqlq 'SELECT VERSION()')" "$(mysqlq 'SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()')" "$(df -B1 --output=avail "$ROOT" | tail -1 | tr -d ' ')" "$(awk '/MemAvailable/{print $2}' /proc/meminfo)"
printf 'APP_ENV='; tr '\0' '\n' < "/proc/$pid/environ" | sed -n 's/^APP_ENV=//p'
printf 'REPORT_ENV='; tr '\0' '\n' < "/proc/$pid/environ" | sed -n 's/^REPORT_EFFECTIVE_ENV=//p'
printf 'MNG_ENV='; tr '\0' '\n' < "/proc/$pid/environ" | sed -n 's/^MNG_TEST_REPORT_ENV=//p'
printf 'ACTIVE_PAPERS=%s\nEXPIRED_STATE0=%s\nEXPIRED_COMPETENCY=%s\n' "$(mysqlq 'SELECT COUNT(*) FROM el_paper WHERE state=1')" "$(mysqlq 'SELECT COUNT(*) FROM el_paper WHERE state=0 AND limit_time IS NOT NULL AND limit_time<=NOW()')" "$(mysqlq "SELECT COUNT(*) FROM el_paper p JOIN el_exam e ON e.id=p.exam_id WHERE p.state=1 AND e.assessment_type='competency' AND p.limit_time IS NOT NULL AND p.limit_time<=NOW()")"
printf 'EXAMS=%s\nPAPERS=%s\nCANDIDATES=%s\nTESTERS=%s\nMBTI_ANSWERS=%s\n' "$(mysqlq 'SELECT COUNT(*) FROM el_exam')" "$(mysqlq 'SELECT COUNT(*) FROM el_paper')" "$(mysqlq 'SELECT COUNT(*) FROM el_candidate')" "$(mysqlq 'SELECT COUNT(*) FROM el_tester')" "$(mysqlq 'SELECT COUNT(*) FROM el_mbti_answer')"
printf 'COMPETENCY_TABLES=%s\nCOMPETENCY_QUESTIONS=%s\nCOMPETENCY_RESULTS=%s\nCOMPETENCY_REPORTS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,14)='el_competency_'")" "$(mysqlq "SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NOT NULL OR competency_question_type IS NOT NULL")" "$(mysqlq "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name='el_competency_result'; SELECT IF((SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name='el_competency_result')=1,(SELECT COUNT(*) FROM el_competency_result),0)")" "$(mysqlq "SELECT IF((SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name='el_competency_report')=1,(SELECT COUNT(*) FROM el_competency_report),0)")"
printf 'MNG_TABLES=%s\nMNG005_REPOS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_'")" "$(mysqlq "SELECT COUNT(*) FROM el_repo WHERE code IN ('00501','00502')")"
printf 'COMPETENCY_VERSION_COLUMNS=%s\nRESULT_RUN_TABLES=%s\n' "$(mysqlq "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='el_exam' AND column_name IN ('competency_product_version','competency_scoring_version','competency_content_version','competency_report_template_version')")" "$(mysqlq "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name IN ('el_competency_result_run','el_competency_result_run_overall','el_competency_result_run_module','el_competency_result_run_dimension','el_competency_result_run_validity','el_competency_report_current')")"
printf 'ORPHAN_PAPER_EXAM=%s\n' "$(mysqlq 'SELECT COUNT(*) FROM el_paper p LEFT JOIN el_exam e ON e.id=p.exam_id WHERE e.id IS NULL')"
printf 'FONT_NOTO=%s\nFONT_WQY=%s\nLO=%s\nCHROME=%s\nBACKUP_DIR=%s\n' "$(fc-list | grep -ci 'Noto Sans CJK')" "$(fc-list | grep -ci 'WenQuanYi')" "$(libreoffice --version 2>/dev/null | head -1)" "$(google-chrome --version 2>/dev/null || chromium --version 2>/dev/null || true)" "$([ -d "$ROOT/backups" ] && echo present || echo absent)"
printf 'UNIT_SHA=%s\nNGINX_SHA=%s\nCONFIG_SHA=%s\n' "$(sha256sum /etc/systemd/system/talent-assessment.service | cut -d' ' -f1)" "$(sha256sum /www/server/nginx/conf/nginx.conf | cut -d' ' -f1)" "$(find "$ROOT/configs" -maxdepth 2 -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum | cut -d' ' -f1)"
printf 'READONLY_PREFLIGHT=PASS\n'
