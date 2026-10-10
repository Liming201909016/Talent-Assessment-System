#!/usr/bin/env bash
set -euo pipefail
umask 077

APP_DIR=/opt/talent-assessment
SERVICE=talent-assessment
EXPECTED_HOST=${1:-}
MYSQL_CLIENT=

cleanup() {
  if [ -n "$MYSQL_CLIENT" ]; then
    rm -f "$MYSQL_CLIENT"
  fi
}
trap cleanup EXIT HUP INT TERM

if [ -n "$EXPECTED_HOST" ] && [ "$(hostname)" != "$EXPECTED_HOST" ]; then
  echo 'INVENTORY_ERROR=unexpected_host'
  exit 1
fi

if [ "$(id -u)" -eq 0 ]; then
  MYSQL_CLIENT=/run/phase1-live-readonly-inventory.cnf
  python3 - "$MYSQL_CLIENT" <<'PY'
from pathlib import Path
import re
import sys

output = Path(sys.argv[1])
paths = (
    Path('/opt/talent-assessment/configs/application.yml'),
    Path('/opt/talent-assessment/configs/application-production.yml'),
)
text = '\n'.join(path.read_text(encoding='utf-8-sig').replace('\r', '') for path in paths if path.exists())
dsn = ''
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
def escape(value):
    return value.replace('\\', '\\\\').replace('"', '\\"')
output.write_text(
    '[client]\nuser="{}"\npassword="{}"\nhost="{}"\nport={}\n'.format(
        escape(user), escape(password), escape(host), port
    ),
    encoding='utf-8',
)
output.chmod(0o600)
PY
  mysqlq() {
    mysql --defaults-extra-file="$MYSQL_CLIENT" --batch --skip-column-names --raw element -e "$1"
  }
else
  mysqlq() {
    sudo -n mysql --batch --skip-column-names --raw element -e "$1"
  }
fi

sha_stream() {
  sha256sum | cut -d' ' -f1
}

mysqlsha() {
  local sql=$1
  local output
  if ! output=$(mysqlq "$sql"); then
    return 1
  fi
  printf '%s' "$output" | sha_stream
}

mysqlnormalizedsha() {
  local sql=$1
  local output
  if ! output=$(mysqlq "$sql"); then
    return 1
  fi
  printf '%s' "$output" | python3 -c "import re,sys; data=sys.stdin.read(); sys.stdout.write(re.sub(r'(?i)\\b(tinyint|smallint|mediumint|int|bigint)\\([0-9]+\\)', r'\\1', data))" | sha_stream
}

DIMENSION_IDS="'competency-a1-01','competency-a1-02','competency-a1-03','competency-a1-04','competency-a1-05','competency-b1-01','competency-b1-02','competency-b1-03','competency-b1-04','competency-b1-05'"

pid=$(systemctl show "$SERVICE" -p MainPID --value)
[ "$pid" -gt 0 ]

proxy_port=8090
if [ "$(hostname)" = 'vm-ubuntu-go-dev' ]; then
  proxy_port=80
fi

printf 'HOST=%s\n' "$(hostname)"
printf 'SERVICE=%s\n' "$(systemctl is-active "$SERVICE")"
printf 'PID=%s\n' "$pid"
printf 'NRESTARTS=%s\n' "$(systemctl show "$SERVICE" -p NRestarts --value)"
printf 'SERVER_SHA=%s\n' "$(sha256sum "$APP_DIR/server" | cut -d' ' -f1)"
printf 'PROCESS_SHA=%s\n' "$(sha256sum "/proc/$pid/exe" | cut -d' ' -f1)"
printf 'INDEX_SHA=%s\n' "$(sha256sum "$APP_DIR/dist/index.html" | cut -d' ' -f1)"
printf 'FRONT_FILES=%s\n' "$(find "$APP_DIR/dist" -type f | wc -l | tr -d ' ')"
printf 'HEALTH_8092=%s\n' "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8092/health)"
printf 'HEALTH_PROXY_PORT=%s\n' "$proxy_port"
printf 'HEALTH_PROXY=%s\n' "$(curl -sS -o /dev/null -w '%{http_code}' "http://127.0.0.1:$proxy_port/prod-api/health")"
printf 'MYSQL_VERSION=%s\n' "$(mysqlq 'SELECT VERSION()')"
printf 'TABLES=%s\n' "$(mysqlq 'SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()')"
printf 'TABLE_LIST=%s\n' "$(mysqlq 'SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() ORDER BY table_name' | paste -sd, -)"
printf 'ACTIVE_PAPERS=%s\n' "$(mysqlq 'SELECT COUNT(*) FROM el_paper WHERE state=1')"
printf 'PHASE1_DIMENSIONS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_competency_dimension WHERE id IN ($DIMENSION_IDS)")"
printf 'PHASE1_QUESTIONS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_qu WHERE dimension_id IN ($DIMENSION_IDS)")"
printf 'PHASE1_QUESTION_DISTRIBUTION=%s\n' "$(mysqlq "SELECT CONCAT_WS('|',COUNT(*),COUNT(DISTINCT question_code),SUM(competency_question_type='dimension'),SUM(competency_question_type='validity'),SUM(competency_question_type='dimension' AND scoring_direction='forward'),SUM(competency_question_type='dimension' AND scoring_direction='reverse'),SUM(competency_question_type='validity' AND scoring_direction='forward'),SUM(question_status=0)) FROM el_qu WHERE dimension_id IN ($DIMENSION_IDS)")"
printf 'PHASE1_DIMENSION_SHA=%s\n' "$(mysqlq "SELECT CONCAT_WS('|',HEX(id),HEX(code),HEX(name),HEX(vird_level),HEX(applicable_category),HEX(core_meaning),display_order,status) FROM el_competency_dimension WHERE id IN ($DIMENSION_IDS) ORDER BY display_order,id" | sha_stream)"
printf 'PHASE1_QUESTION_SHA=%s\n' "$(mysqlq "SELECT CONCAT_WS('|',HEX(question_code),HEX(dimension_id),HEX(competency_question_type),dimension_item_no,HEX(content),HEX(observation_point),HEX(scoring_direction),question_status,HEX(COALESCE(remark,''))) FROM el_qu WHERE dimension_id IN ($DIMENSION_IDS) ORDER BY question_code" | sha_stream)"
printf 'PHASE1_REPO_RELATIONS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_qu_repo WHERE qu_id IN (SELECT id FROM el_qu WHERE dimension_id IN ($DIMENSION_IDS))")"
printf 'PHASE1_ANSWER_ROWS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_qu_answer WHERE qu_id IN (SELECT id FROM el_qu WHERE dimension_id IN ($DIMENSION_IDS))")"
printf 'PHASE1_RUNTIME_REFS=%s\n' "$(mysqlq "SELECT (SELECT COUNT(*) FROM el_exam_competency_dimension WHERE dimension_id IN ($DIMENSION_IDS))+(SELECT COUNT(*) FROM el_exam_competency_question WHERE source_qu_id IN (SELECT id FROM el_qu WHERE dimension_id IN ($DIMENSION_IDS)))")"
printf 'PHASE1_EXAMS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_exam WHERE competency_product_version IN ('competency-frontline-phase1-v1','competency-frontline-phase1-v2')")"
printf 'PHASE1_PAPERS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_paper WHERE exam_id IN (SELECT id FROM el_exam WHERE competency_product_version IN ('competency-frontline-phase1-v1','competency-frontline-phase1-v2'))")"
printf 'PHASE1_RESULTS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_competency_result WHERE exam_id IN (SELECT id FROM el_exam WHERE competency_product_version IN ('competency-frontline-phase1-v1','competency-frontline-phase1-v2'))")"
printf 'PHASE1_REPORTS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_competency_report WHERE exam_id IN (SELECT id FROM el_exam WHERE competency_product_version IN ('competency-frontline-phase1-v1','competency-frontline-phase1-v2'))")"
printf 'REPO_00401=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_repo WHERE code='00401'")"
printf 'EXAMS_BY_REPO_00401=%s\n' "$(mysqlq "SELECT COUNT(DISTINCT e.id) FROM el_exam e INNER JOIN el_exam_repo er ON er.exam_id=e.id INNER JOIN el_repo r ON r.id=er.repo_id WHERE r.code='00401'")"
printf 'COMPETENCY_EXAMS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_exam WHERE assessment_type='competency'")"
printf 'COMPETENCY_PAPERS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_paper WHERE exam_id IN (SELECT id FROM el_exam WHERE assessment_type='competency')")"
printf 'COMPETENCY_RESULTS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_competency_result")"
printf 'COMPETENCY_REPORTS=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_competency_report")"
printf 'REPORT_TEXT_V1=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_competency_report_text WHERE content_version='competency-phase1-content-v1'")"
printf 'REPORT_TEXT_V2=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_competency_report_text WHERE content_version='competency-phase1-content-v2'")"
printf 'REPORT_PACKAGE_V1=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_competency_report_content_package WHERE content_version='competency-phase1-content-v1' AND effective_environment IN ('staging','production')")"
printf 'REPORT_PACKAGE_V2=%s\n' "$(mysqlq "SELECT COUNT(*) FROM el_competency_report_content_package WHERE content_version='competency-phase1-content-v2' AND effective_environment IN ('staging','production')")"
printf 'REPORT_V1_TEXT_SHA=%s\n' "$(mysqlsha "SELECT CONCAT_WS('|',HEX(t.id),HEX(t.content_type),HEX(COALESCE(t.dimension_id,'')),HEX(COALESCE(t.level_code,'')),HEX(t.content),HEX(COALESCE(t.disclaimer,'')),t.is_temporary,t.status) FROM el_competency_report_text AS t WHERE t.content_version='competency-phase1-content-v1' ORDER BY t.id")"
printf 'REPORT_V2_TEXT_SHA=%s\n' "$(mysqlsha "SELECT CONCAT_WS('|',HEX(t.id),HEX(t.content_type),HEX(COALESCE(t.dimension_id,'')),HEX(COALESCE(t.level_code,'')),HEX(t.content),HEX(COALESCE(t.disclaimer,'')),t.is_temporary,t.status) FROM el_competency_report_text AS t WHERE t.content_version='competency-phase1-content-v2' ORDER BY t.id")"
printf 'REPORT_PACKAGE_V1_SHA=%s\n' "$(mysqlsha "SELECT CONCAT_WS('|',HEX(p.product_version),HEX(p.scoring_version),HEX(p.content_version),HEX(p.template_version),HEX(p.audience),HEX(p.approval_status),HEX(COALESCE(p.question_source_sha256,'')),HEX(COALESCE(p.content_source_sha256,'')),HEX(COALESCE(p.disclaimer,''))) FROM el_competency_report_content_package AS p WHERE p.content_version='competency-phase1-content-v1' ORDER BY p.id")"
printf 'REPORT_PACKAGE_V2_SHA=%s\n' "$(mysqlsha "SELECT CONCAT_WS('|',HEX(p.product_version),HEX(p.scoring_version),HEX(p.content_version),HEX(p.template_version),HEX(p.audience),HEX(p.approval_status),HEX(COALESCE(p.question_source_sha256,'')),HEX(COALESCE(p.content_source_sha256,'')),HEX(COALESCE(p.disclaimer,''))) FROM el_competency_report_content_package AS p WHERE p.content_version='competency-phase1-content-v2' ORDER BY p.id")"
printf 'COMPETENCY_SCHEMA_SHA=%s\n' "$(mysqlsha "SELECT CONCAT_WS('|',table_name,column_name,column_type,is_nullable,COALESCE(column_default,'<NULL>'),extra,ordinal_position,COALESCE(collation_name,'')) FROM information_schema.columns WHERE table_schema=DATABASE() AND (table_name LIKE 'el_competency_%' OR table_name IN ('el_exam','el_qu','el_paper','el_paper_qu','el_paper_qu_answer')) ORDER BY table_name,ordinal_position")"
printf 'COMPETENCY_SCHEMA_SEMANTIC_SHA=%s\n' "$(mysqlsha "SELECT CONCAT_WS('|',table_name,column_name,column_type,is_nullable,COALESCE(column_default,'<NULL>'),extra,ordinal_position) FROM information_schema.columns WHERE table_schema=DATABASE() AND (table_name LIKE 'el_competency_%' OR table_name IN ('el_exam','el_qu','el_paper','el_paper_qu','el_paper_qu_answer')) ORDER BY table_name,ordinal_position")"
printf 'COMPETENCY_SCHEMA_PORTABLE_SHA=%s\n' "$(mysqlnormalizedsha "SELECT CONCAT_WS('|',table_name,column_name,column_type,is_nullable,COALESCE(column_default,'<NULL>'),extra,ordinal_position) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name<>'el_competency_migration' AND (table_name LIKE 'el_competency_%' OR table_name IN ('el_exam','el_qu','el_paper','el_paper_qu','el_paper_qu_answer')) ORDER BY table_name,ordinal_position")"
printf 'COMPETENCY_INDEX_SHA=%s\n' "$(mysqlsha "SELECT CONCAT_WS('|',table_name,index_name,non_unique,seq_in_index,column_name,COALESCE(sub_part,'<NULL>'),index_type) FROM information_schema.statistics WHERE table_schema=DATABASE() AND (table_name LIKE 'el_competency_%' OR table_name IN ('el_exam','el_qu','el_paper','el_paper_qu','el_paper_qu_answer')) ORDER BY table_name,index_name,seq_in_index")"
printf 'COMPETENCY_INDEX_PORTABLE_SHA=%s\n' "$(mysqlsha "SELECT CONCAT_WS('|',table_name,index_name,non_unique,seq_in_index,column_name,COALESCE(sub_part,'<NULL>'),index_type) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name<>'el_competency_migration' AND (table_name LIKE 'el_competency_%' OR table_name IN ('el_exam','el_qu','el_paper','el_paper_qu','el_paper_qu_answer')) ORDER BY table_name,index_name,seq_in_index")"
printf 'COMPETENCY_FK_SHA=%s\n' "$(mysqlsha "SELECT CONCAT_WS('|',table_name,constraint_name,column_name,referenced_table_name,referenced_column_name,ordinal_position) FROM information_schema.key_column_usage WHERE table_schema=DATABASE() AND referenced_table_name IS NOT NULL AND (table_name LIKE 'el_competency_%' OR table_name IN ('el_exam','el_qu','el_paper','el_paper_qu','el_paper_qu_answer')) ORDER BY table_name,constraint_name,ordinal_position")"
printf 'V1_TEMPLATE='; if [ -f "$APP_DIR/configs/export-templates/competency-phase1-report.docx" ]; then sha256sum "$APP_DIR/configs/export-templates/competency-phase1-report.docx" | cut -d' ' -f1; else echo absent; fi
printf 'V2_TEMPLATE='; if [ -f "$APP_DIR/configs/export-templates/competency-phase1-report-v2.docx" ]; then sha256sum "$APP_DIR/configs/export-templates/competency-phase1-report-v2.docx" | cut -d' ' -f1; else echo absent; fi
if [ "${2:-}" = 'schema-details' ]; then
  mysqlq "SELECT CONCAT('SCHEMA_DETAIL=',HEX(CONCAT_WS('|',table_name,column_name,column_type,is_nullable,COALESCE(column_default,'<NULL>'),extra,ordinal_position))) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name<>'el_competency_migration' AND (table_name LIKE 'el_competency_%' OR table_name IN ('el_exam','el_qu','el_paper','el_paper_qu','el_paper_qu_answer')) ORDER BY table_name,ordinal_position"
  mysqlq "SELECT CONCAT('INDEX_DETAIL=',HEX(CONCAT_WS('|',table_name,index_name,non_unique,seq_in_index,column_name,COALESCE(sub_part,'<NULL>'),index_type))) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name<>'el_competency_migration' AND (table_name LIKE 'el_competency_%' OR table_name IN ('el_exam','el_qu','el_paper','el_paper_qu','el_paper_qu_answer')) ORDER BY table_name,index_name,seq_in_index"
fi
printf 'INVENTORY=PASS\n'
