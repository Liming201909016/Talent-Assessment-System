#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT=/opt/talent-assessment
EXPECTED_HOST=iZ0yosjdcen2p4Z
CLIENT=/run/talent-assessment-paper-orphan-readonly.cnf
cleanup_status=not_started
failed_stage=bootstrap

cleanup() {
  if rm -f "$CLIENT"; then
    cleanup_status=completed
  else
    cleanup_status=failed
  fi
}
trap cleanup EXIT HUP INT TERM

fail() {
  local message=${1:-unknown_error}
  cleanup
  trap - EXIT HUP INT TERM
  python3 - "$failed_stage" "$message" "$cleanup_status" <<'PY'
import json, sys
print(json.dumps({
    "status": "failed",
    "failedStage": sys.argv[1],
    "error": sys.argv[2],
    "createdResources": ["ephemeral_mysql_client_config"],
    "cleanup": sys.argv[3],
    "databaseWrites": 0,
}, ensure_ascii=True, separators=(",", ":")))
PY
  exit 1
}

[ "$(hostname)" = "$EXPECTED_HOST" ] || fail wrong_host
[ "$(id -u)" = 0 ] || fail root_required
[ -f "$ROOT/configs/application-production.yml" ] || fail production_config_missing

failed_stage=config
python3 - "$CLIENT" <<'PY' || fail db_config_unavailable
from pathlib import Path
import re
import sys

client = Path(sys.argv[1])
paths = [
    Path('/opt/talent-assessment/configs/application.yml'),
    Path('/opt/talent-assessment/configs/application-production.yml'),
]
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
client.write_text(
    '[client]\nuser=' + user + '\npassword=' + password + '\nhost=' + host + '\nport=' + port + '\n',
    encoding='utf-8',
)
client.chmod(0o600)
PY

failed_stage=inventory
result=$(python3 - "$CLIENT" <<'PY'
import hashlib
import json
import re
import subprocess
import sys

client = sys.argv[1]
base = [
    'mysql', '--defaults-extra-file=' + client, '--batch', '--skip-column-names',
    '--raw', 'element',
]

def rows(sql):
    statement = 'SET SESSION TRANSACTION READ ONLY; ' + sql
    completed = subprocess.run(base + ['-e', statement], check=True, capture_output=True, text=True)
    return [line.split('\t') for line in completed.stdout.splitlines() if line]

def scalar(sql):
    data = rows(sql)
    if len(data) != 1 or len(data[0]) != 1:
        raise RuntimeError('unexpected_scalar_shape')
    return data[0][0]

def safe_identifier(value):
    if not re.fullmatch(r'[A-Za-z0-9_]+', value):
        raise RuntimeError('unsafe_identifier')
    return '`' + value + '`'

orphan_where = 'NOT EXISTS (SELECT 1 FROM el_exam e WHERE e.id=p.exam_id)'
summary = {
    'orphanPapers': int(scalar('SELECT COUNT(*) FROM el_paper p WHERE ' + orphan_where)),
    'distinctMissingExamIds': int(scalar('SELECT COUNT(DISTINCT p.exam_id) FROM el_paper p WHERE ' + orphan_where)),
    'activeOrphanPapers': int(scalar('SELECT COUNT(*) FROM el_paper p WHERE ' + orphan_where + ' AND p.state=1')),
    'paperQuestions': int(scalar('SELECT COUNT(*) FROM el_paper_qu pq JOIN el_paper p ON p.id=pq.paper_id WHERE ' + orphan_where)),
    'paperAnswers': int(scalar('SELECT COUNT(*) FROM el_paper_qu_answer pqa JOIN el_paper p ON p.id=pqa.paper_id WHERE ' + orphan_where)),
    'sourceQuestionsStillPresent': int(scalar('SELECT COUNT(*) FROM el_paper_qu pq JOIN el_paper p ON p.id=pq.paper_id JOIN el_qu q ON q.id=pq.qu_id WHERE ' + orphan_where)),
    'candidateRefs': int(scalar('SELECT COUNT(*) FROM el_candidate c JOIN el_paper p ON p.id=c.paper_id WHERE ' + orphan_where)),
    'testerRefs': int(scalar('SELECT COUNT(*) FROM el_tester t JOIN el_paper p ON p.id=t.paper_id WHERE ' + orphan_where)),
    'mbtiAnswerRefs': int(scalar('SELECT COUNT(*) FROM el_mbti_answer a JOIN el_paper p ON p.id=a.paper_id WHERE ' + orphan_where)),
    'linkedSystemUsers': int(scalar("SELECT COUNT(*) FROM el_paper p JOIN sys_user u ON CAST(u.user_id AS CHAR)=p.user_id WHERE " + orphan_where)),
}

state_distribution = [
    {'state': int(state), 'papers': int(count)}
    for state, count in rows('SELECT p.state,COUNT(*) FROM el_paper p WHERE ' + orphan_where + ' GROUP BY p.state ORDER BY p.state')
]

date_bounds_row = rows(
    "SELECT COALESCE(DATE_FORMAT(MIN(p.create_time),'%Y-%m-%d'),'NULL'),"
    "COALESCE(DATE_FORMAT(MAX(p.create_time),'%Y-%m-%d'),'NULL'),"
    "COALESCE(DATE_FORMAT(MIN(p.update_time),'%Y-%m-%d'),'NULL'),"
    "COALESCE(DATE_FORMAT(MAX(p.update_time),'%Y-%m-%d'),'NULL') "
    'FROM el_paper p WHERE ' + orphan_where
)[0]
date_bounds = {
    'createMin': date_bounds_row[0],
    'createMax': date_bounds_row[1],
    'updateMin': date_bounds_row[2],
    'updateMax': date_bounds_row[3],
}

by_year = [
    {'year': year, 'state': int(state), 'papers': int(count)}
    for year, state, count in rows(
        "SELECT COALESCE(DATE_FORMAT(p.create_time,'%Y'),'NULL'),p.state,COUNT(*) "
        'FROM el_paper p WHERE ' + orphan_where + ' GROUP BY 1,2 ORDER BY 1,2'
    )
]

missing_exam_groups = [
    {'examIdSha256': digest, 'state': int(state), 'papers': int(count)}
    for digest, state, count in rows(
        "SELECT SHA2(COALESCE(p.exam_id,''),256),p.state,COUNT(*) FROM el_paper p WHERE "
        + orphan_where + ' GROUP BY p.exam_id,p.state ORDER BY COUNT(*) DESC,SHA2(COALESCE(p.exam_id,\'\'),256)'
    )
]

missing_exam_details = []
for row in rows(
    "SELECT SHA2(p.exam_id,256),COUNT(*),SUM(p.state=0),SUM(p.state=2),"
    "COUNT(DISTINCT p.user_id),COUNT(DISTINCT p.title),"
    "COALESCE(DATE_FORMAT(MIN(p.create_time),'%Y-%m-%d'),'NULL'),"
    "COALESCE(DATE_FORMAT(MAX(p.create_time),'%Y-%m-%d'),'NULL'),"
    "(SELECT COUNT(*) FROM el_paper_qu pq JOIN el_paper p2 ON p2.id=pq.paper_id WHERE p2.exam_id=p.exam_id),"
    "(SELECT COUNT(*) FROM el_paper_qu pq JOIN el_paper p2 ON p2.id=pq.paper_id WHERE p2.exam_id=p.exam_id AND pq.answered=1),"
    "(SELECT COUNT(*) FROM el_paper_qu_answer pqa JOIN el_paper p2 ON p2.id=pqa.paper_id WHERE p2.exam_id=p.exam_id),"
    "(SELECT COUNT(*) FROM el_exam_repo er WHERE er.exam_id=p.exam_id),"
    "(SELECT COALESCE(GROUP_CONCAT(DISTINCT r.code ORDER BY r.code),'NONE') FROM el_exam_repo er LEFT JOIN el_repo r ON r.id=er.repo_id WHERE er.exam_id=p.exam_id),"
    "(SELECT COUNT(*) FROM el_user_exam ue WHERE ue.exam_id=p.exam_id),"
    "(SELECT COUNT(DISTINCT ue.user_id) FROM el_user_exam ue WHERE ue.exam_id=p.exam_id),"
    "(SELECT COUNT(*) FROM el_user_book ub WHERE ub.exam_id=p.exam_id) "
    'FROM el_paper p WHERE ' + orphan_where + ' GROUP BY p.exam_id ORDER BY COUNT(*) DESC'
):
    (digest, papers, state0, state2, users, titles, create_min, create_max,
     paper_questions, answered_questions, paper_answers, exam_repo_rows, repo_codes,
     user_exam_rows, user_exam_users, user_book_rows) = row
    missing_exam_details.append({
        'examIdSha256': digest,
        'papers': int(papers),
        'state0': int(state0),
        'state2': int(state2),
        'distinctUsers': int(users),
        'distinctTitles': int(titles),
        'createMin': create_min,
        'createMax': create_max,
        'paperQuestions': int(paper_questions),
        'answeredQuestions': int(answered_questions),
        'paperAnswers': int(paper_answers),
        'examRepoRows': int(exam_repo_rows),
        'repoCodes': repo_codes.split(',') if repo_codes != 'NONE' else [],
        'userExamRows': int(user_exam_rows),
        'userExamUsers': int(user_exam_users),
        'userBookRows': int(user_book_rows),
    })

title_groups = [
    {'titleSha256': digest, 'papers': int(count)}
    for digest, count in rows(
        "SELECT SHA2(COALESCE(p.title,''),256),COUNT(*) FROM el_paper p WHERE "
        + orphan_where + " GROUP BY SHA2(COALESCE(p.title,''),256) ORDER BY COUNT(*) DESC LIMIT 50"
    )
]

repo_distribution = [
    {'repoCode': code, 'paperQuestions': int(count), 'papers': int(papers)}
    for code, count, papers in rows(
        "SELECT COALESCE(r.code,'UNRESOLVED'),COUNT(*),COUNT(DISTINCT p.id) "
        'FROM el_paper p JOIN el_paper_qu pq ON pq.paper_id=p.id '
        'LEFT JOIN el_qu_repo qr ON qr.qu_id=pq.qu_id LEFT JOIN el_repo r ON r.id=qr.repo_id WHERE '
        + orphan_where + " GROUP BY COALESCE(r.code,'UNRESOLVED') ORDER BY COUNT(*) DESC"
    )
]

paper_id_tables = [r[0] for r in rows(
    "SELECT table_name FROM information_schema.columns WHERE table_schema=DATABASE() "
    "AND column_name='paper_id' AND table_name<>'el_paper' ORDER BY table_name"
)]
paper_references = []
for table in paper_id_tables:
    quoted = safe_identifier(table)
    count, papers = rows(
        'SELECT COUNT(*),COUNT(DISTINCT child.paper_id) FROM ' + quoted
        + ' child JOIN el_paper p ON p.id=child.paper_id WHERE ' + orphan_where
    )[0]
    paper_references.append({'table': table, 'rows': int(count), 'papers': int(papers)})

exam_id_tables = [r[0] for r in rows(
    "SELECT table_name FROM information_schema.columns WHERE table_schema=DATABASE() "
    "AND column_name='exam_id' AND table_name NOT IN ('el_exam','el_paper') ORDER BY table_name"
)]
missing_exam_references = []
missing_exam_reference_groups = []
for table in exam_id_tables:
    quoted = safe_identifier(table)
    count, ids = rows(
        'SELECT COUNT(*),COUNT(DISTINCT child.exam_id) FROM ' + quoted
        + ' child WHERE child.exam_id IS NOT NULL AND NOT EXISTS '
          '(SELECT 1 FROM el_exam e WHERE e.id=child.exam_id)'
    )[0]
    missing_exam_references.append({'table': table, 'rows': int(count), 'missingExamIds': int(ids)})
    if int(count) > 0:
        groups = rows(
            'SELECT SHA2(child.exam_id,256),COUNT(*) FROM ' + quoted
            + ' child WHERE child.exam_id IS NOT NULL AND NOT EXISTS '
              '(SELECT 1 FROM el_exam e WHERE e.id=child.exam_id) '
              'GROUP BY child.exam_id ORDER BY COUNT(*) DESC,SHA2(child.exam_id,256)'
        )
        missing_exam_reference_groups.append({
            'table': table,
            'groups': [{'examIdSha256': digest, 'rows': int(group_count)} for digest, group_count in groups],
        })

fingerprint_lines = sorted(
    f"{item['examIdSha256']}:{item['papers']}:{item['state0']}:{item['state2']}"
    for item in missing_exam_details
)
group_fingerprint = hashlib.sha256('\n'.join(fingerprint_lines).encode()).hexdigest()
expected_summary = {
    'orphanPapers': 409,
    'distinctMissingExamIds': 14,
    'activeOrphanPapers': 0,
    'paperQuestions': 20566,
    'paperAnswers': 42963,
    'candidateRefs': 0,
    'testerRefs': 0,
    'mbtiAnswerRefs': 0,
    'linkedSystemUsers': 409,
}
expected_missing_exam_refs = {'el_exam_repo': (24, 24), 'el_user_exam': (19, 12), 'el_user_book': (43, 3)}
actual_missing_exam_refs = {
    item['table']: (item['rows'], item['missingExamIds'])
    for item in missing_exam_references if item['rows'] > 0
}
retention_drift = []
for key, expected in expected_summary.items():
    if summary[key] != expected:
        retention_drift.append({'field': key, 'expected': expected, 'actual': summary[key]})
if group_fingerprint != '26836df28929fb71c2fa4b901a8cc2dc0a1dfd14eb61d1b4fcfad27af1a7e21b':
    retention_drift.append({'field': 'groupFingerprint', 'expected': '26836df28929fb71c2fa4b901a8cc2dc0a1dfd14eb61d1b4fcfad27af1a7e21b', 'actual': group_fingerprint})
if actual_missing_exam_refs != expected_missing_exam_refs:
    retention_drift.append({'field': 'missingExamReferences', 'expected': expected_missing_exam_refs, 'actual': actual_missing_exam_refs})

foreign_keys = [
    {'constraint': name, 'column': column, 'referencedTable': ref_table, 'referencedColumn': ref_column}
    for name, column, ref_table, ref_column in rows(
        "SELECT constraint_name,column_name,referenced_table_name,referenced_column_name "
        "FROM information_schema.key_column_usage WHERE table_schema=DATABASE() "
        "AND table_name='el_paper' AND referenced_table_name IS NOT NULL ORDER BY constraint_name,column_name"
    )
]

print(json.dumps({
    'status': 'passed' if not retention_drift else 'failed',
    'failedStage': None if not retention_drift else 'retention_baseline',
    'error': None if not retention_drift else 'retention_baseline_drift',
    'host': subprocess.run(['hostname'], check=True, capture_output=True, text=True).stdout.strip(),
    'databaseVersion': scalar('SELECT VERSION()'),
    'readOnly': True,
    'databaseWrites': 0,
    'summary': summary,
    'stateDistribution': state_distribution,
    'dateBounds': date_bounds,
    'byYear': by_year,
    'missingExamGroups': missing_exam_groups,
    'missingExamDetails': missing_exam_details,
    'titleGroups': title_groups,
    'repoDistribution': repo_distribution,
    'paperReferences': paper_references,
    'missingExamReferences': missing_exam_references,
    'missingExamReferenceGroups': missing_exam_reference_groups,
    'paperForeignKeys': foreign_keys,
    'retentionBaseline': {
        'status': 'passed' if not retention_drift else 'failed',
        'groupFingerprint': group_fingerprint,
        'drift': retention_drift,
    },
    'createdResources': ['ephemeral_mysql_client_config'],
    'cleanup': 'pending',
}, ensure_ascii=True, separators=(',', ':')))
PY
) || fail inventory_query_failed

failed_stage=cleanup
cleanup
trap - EXIT HUP INT TERM
[ "$cleanup_status" = completed ] || fail cleanup_failed

python3 - "$result" <<'PY'
import json, sys
payload = json.loads(sys.argv[1])
payload['cleanup'] = 'completed'
print(json.dumps(payload, ensure_ascii=True, separators=(',', ':')))
if payload['status'] != 'passed':
    raise SystemExit(2)
PY
