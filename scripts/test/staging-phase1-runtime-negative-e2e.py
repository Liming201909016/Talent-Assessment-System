#!/usr/bin/env python3
import base64
import hashlib
import hmac
import json
import re
import subprocess
import time
import urllib.error
import urllib.request
import uuid
import zipfile
from io import BytesIO
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'
TRIGGER = 'phase1_runtime_negative_question_fail'


def b64url(value):
    return base64.urlsafe_b64encode(value).rstrip(b'=').decode()


def config_text():
    environment = subprocess.run(
        ['systemctl', 'show', 'talent-assessment', '--property=Environment', '--value'],
        check=True, capture_output=True, text=True,
    ).stdout
    match = re.search(r'(?:^|\s)APP_ENV=([^\s]+)', environment)
    name = match.group(1) if match else 'local'
    files = [APP_DIR / 'configs' / 'application.yml', APP_DIR / 'configs' / f'application-{name}.yml']
    return '\n'.join(path.read_text(encoding='utf-8-sig').replace('\r', '') for path in files if path.exists())


def scalar(text, section, key, default=''):
    value = default
    for section_match in re.finditer(rf'(?ms)^{re.escape(section)}:\s*\n(.*?)(?=^[A-Za-z][\w-]*:\s*$|\Z)', text):
        key_match = re.search(rf'(?m)^\s+{re.escape(key)}:\s*(.+?)\s*$', section_match.group(1))
        if key_match:
            value = key_match.group(1).strip().strip('"\'')
    return value


def sign(secret, token_id):
    header = b64url(json.dumps({'alg': 'HS512', 'typ': 'JWT'}, separators=(',', ':')).encode())
    payload = b64url(json.dumps({'login_user_key': token_id}, separators=(',', ':')).encode())
    raw = f'{header}.{payload}'
    return f'{raw}.{b64url(hmac.new(secret.encode(), raw.encode(), hashlib.sha512).digest())}'


def mysql(sql):
    return subprocess.run(
        ['sudo', '-n', 'mysql', 'element', '-Nse', sql],
        check=True, capture_output=True, text=True,
    ).stdout.strip()


def mysql_exec(sql):
    subprocess.run(['sudo', '-n', 'mysql', 'element', '-e', sql], check=True)


def api(path, token=None, body=None, method='POST', expect_success=True):
    headers = {}
    if token:
        headers['Authorization'] = 'Bearer ' + token
    payload = None
    if body is not None:
        headers['Content-Type'] = 'application/json'
        payload = json.dumps(body, ensure_ascii=False, separators=(',', ':')).encode()
    call = urllib.request.Request(BASE + path, data=payload, headers=headers, method=method)
    try:
        with urllib.request.urlopen(call, timeout=120) as response:
            raw = response.read()
            status = response.status
            response_headers = dict(response.headers)
    except urllib.error.HTTPError as error:
        raw = error.read()
        status = error.code
        response_headers = dict(error.headers)
    if not expect_success:
        return status, raw, response_headers
    data = json.loads(raw.decode())
    if status >= 400 or data.get('code') not in (0, 200):
        raise RuntimeError(f'{path}: HTTP {status}: {data.get("msg")}')
    return data.get('data') or {}


def paper_api(path, token, body):
    headers = {'Content-Type': 'application/json', 'X-Competency-Token': token}
    call = urllib.request.Request(
        BASE + path,
        data=json.dumps(body, ensure_ascii=False, separators=(',', ':')).encode(),
        headers=headers,
        method='POST',
    )
    with urllib.request.urlopen(call, timeout=120) as response:
        data = json.loads(response.read().decode())
    if data.get('code') not in (0, 200):
        raise RuntimeError(f'{path}: {data.get("msg")}')
    return data.get('data') or {}


def expect_business_error(path, token=None, body=None, method='POST', contains=None):
    status, raw, _ = api(path, token, body, method, expect_success=False)
    if status >= 400:
        raise RuntimeError(f'{path}: unexpected HTTP {status}: {raw.decode(errors="replace")}')
    data = json.loads(raw.decode())
    if data.get('code') in (0, 200):
        raise RuntimeError(f'{path}: expected business rejection, got {data}')
    if contains and contains not in data.get('msg', ''):
        raise RuntimeError(f'{path}: rejection {data.get("msg")!r} missing {contains!r}')
    return data


def create_exam(admin, title):
    exam = api('/exam/api/exam/exam/save', admin, {
        'title': title, 'content': 'temporary phase-1 negative verification',
        'assessmentType': 'competency', 'scoringMode': 'competency_average',
        'joinType': 1, 'openType': 1, 'isOpen': 1, 'answerType': 1,
        'state': 0, 'totalTime': 20, 'repoList': [], 'departIds': [],
    })
    return exam['id']


def publish(admin, exam_id):
    data = api('/exam/api/competency/exams/publish', admin, {'examId': exam_id})
    if data.get('dimensionCount') != 10 or data.get('questionCount') != 90:
        raise RuntimeError(f'publish mismatch: {data}')
    return data


def create_candidate_and_paper(admin, exam_id, suffix, name):
    candidate = api('/exam/api/candidate/save', admin, {
        'examId': exam_id, 'name': name,
        'telephone': '139' + suffix[:8], 'gender': '0',
    })
    participant_token = candidate.get('participantToken')
    if not participant_token:
        raise RuntimeError('participant token missing')
    access = api('/exam/api/competency/participant/create-paper', body={
        'examId': exam_id, 'participantId': candidate['id'],
        'participantType': 'candidate', 'participantToken': participant_token,
    })
    return candidate['id'], access['paperId'], access['paperToken']


def answer_paper(paper_id, paper_token, raw_for_validity, skipped_codes=None):
    skipped_codes = skipped_codes or set()
    detail = paper_api('/exam/api/competency/participant/paper-detail', paper_token, {'paperId': paper_id})
    questions = detail.get('questions') or []
    if len(questions) != 90:
        raise RuntimeError(f'paper question count={len(questions)}')
    for question in questions:
        if question['code'] in skipped_codes:
            continue
        raw = raw_for_validity if question['code'].startswith('P1-VAL-Q') else 3
        paper_api('/exam/api/competency/participant/fill-answer', paper_token, {
            'paperId': paper_id, 'paperQuestionId': question['id'], 'rawValue': raw,
        })


def cleanup_exam(exam_id, paper_id='', candidate_id=''):
    if paper_id:
        mysql_exec(f"""
DELETE FROM el_competency_report_audit WHERE paper_id='{paper_id}';
DELETE FROM el_competency_report WHERE paper_id='{paper_id}';
DELETE FROM el_competency_group_result WHERE paper_id='{paper_id}';
DELETE FROM el_competency_validity_result WHERE paper_id='{paper_id}';
DELETE FROM el_competency_dimension_result WHERE paper_id='{paper_id}';
DELETE FROM el_competency_result WHERE paper_id='{paper_id}';
DELETE FROM el_paper_qu_answer WHERE paper_id='{paper_id}';
DELETE FROM el_paper_qu WHERE paper_id='{paper_id}';
UPDATE el_candidate SET paper_id=NULL,end_time=NULL WHERE id='{candidate_id}';
DELETE FROM el_paper WHERE id='{paper_id}';
""")
    if candidate_id:
        mysql_exec(f"DELETE FROM el_candidate WHERE id='{candidate_id}';")
    if exam_id:
        mysql_exec(f"""
DELETE FROM el_exam_competency_question WHERE exam_id='{exam_id}';
DELETE FROM el_exam_competency_dimension WHERE exam_id='{exam_id}';
DELETE FROM el_exam_competency_group WHERE exam_id='{exam_id}';
DELETE FROM el_exam WHERE id='{exam_id}';
""")


def workbook_xml(payload):
    if not payload.startswith(b'PK'):
        raise RuntimeError('export is not xlsx')
    with zipfile.ZipFile(BytesIO(payload)) as archive:
        return '\n'.join(
            archive.read(name).decode('utf-8', errors='ignore')
            for name in archive.namelist() if name.endswith('.xml')
        )


def download_export(admin, exam_id):
    call = urllib.request.Request(
        BASE + '/exam/api/exam/exam/export-raw-data?examId=' + exam_id,
        headers={'Authorization': 'Bearer ' + admin}, method='GET',
    )
    with urllib.request.urlopen(call, timeout=120) as response:
        payload = response.read()
        if 'spreadsheetml.sheet' not in response.headers.get('Content-Type', ''):
            raise RuntimeError('export content type mismatch')
    return workbook_xml(payload)


def verify_timeout_incomplete(admin, suffix):
    exam_id = candidate_id = paper_id = ''
    try:
        exam_id = create_exam(admin, 'PHASE1-TIMEOUT-' + suffix)
        publish(admin, exam_id)
        candidate_id, paper_id, paper_token = create_candidate_and_paper(admin, exam_id, suffix, 'Phase1 Timeout Verify')
        answer_paper(paper_id, paper_token, 1, {'A1-01-Q01', 'P1-VAL-Q01'})
        mysql_exec(f"UPDATE el_paper SET limit_time=DATE_SUB(NOW(), INTERVAL 1 SECOND) WHERE id='{paper_id}';")
        submitted = paper_api('/exam/api/competency/participant/submit', paper_token, {
            'paperId': paper_id, 'submitType': 'manual',
        })
        if submitted.get('isComplete'):
            raise RuntimeError('timeout submission unexpectedly complete')
        result = api('/exam/api/competency/results/detail', admin, {'paperId': paper_id})
        overall = result['result']
        dimensions = result.get('dimensions') or []
        groups = result.get('groups') or []
        validity = result.get('validity') or {}
        if [overall.get('totalQuestionCount'), overall.get('answeredQuestionCount'), overall.get('dimensionQuestionCount'), overall.get('answeredDimensionQuestionCount')] != [90, 88, 80, 79]:
            raise RuntimeError(f'timeout counts mismatch: {overall}')
        if overall.get('overallScore') is not None or overall.get('evaluationLevel') is not None or overall.get('isComplete') != 0 or overall.get('submitType') != 'timeout':
            raise RuntimeError(f'timeout overall formal fields mismatch: {overall}')
        if len(dimensions) != 10 or sum(row.get('dimensionScore') is None for row in dimensions) != 1:
            raise RuntimeError(f'timeout dimension results mismatch: {dimensions}')
        if len(groups) != 2 or sum(row.get('groupScore') is None for row in groups) != 1:
            raise RuntimeError(f'timeout group results mismatch: {groups}')
        if validity.get('answeredQuestionCount') != 9 or validity.get('validityScore') is not None or validity.get('validityStatus') != 'incomplete' or validity.get('isComplete') != 0:
            raise RuntimeError(f'timeout validity mismatch: {validity}')
        db = mysql(f"""
SELECT CONCAT(
  (SELECT overall_score IS NULL FROM el_competency_result WHERE paper_id='{paper_id}'),'|',
  (SELECT COUNT(*) FROM el_competency_dimension_result WHERE paper_id='{paper_id}' AND dimension_score IS NULL),'|',
  (SELECT COUNT(*) FROM el_competency_group_result WHERE paper_id='{paper_id}' AND group_score IS NULL),'|',
  (SELECT validity_score IS NULL FROM el_competency_validity_result WHERE paper_id='{paper_id}'))
""")
        if db != '1|1|1|1':
            raise RuntimeError(f'timeout null persistence mismatch: {db}')
        incomplete = api('/exam/api/competency/results/paging', admin, {'examId': exam_id, 'current': 1, 'size': 20, 'completion': 'incomplete', 'validity': 'incomplete', 'sortBy': 'submittedAt'})
        ranking = api('/exam/api/competency/results/paging', admin, {'examId': exam_id, 'current': 1, 'size': 20, 'sortBy': 'overallScore'})
        if incomplete.get('total') != 1 or ranking.get('total') != 0:
            raise RuntimeError('timeout filters/ranking mismatch')
        report_error = expect_business_error('/exam/api/competency/admin/report-data?paperId=' + paper_id, admin, method='GET')
        if '未完整作答' not in report_error.get('msg', ''):
            raise RuntimeError(f'timeout report rejection mismatch: {report_error}')
        print('TIMEOUT_INCOMPLETE_PASS')
        print('timeout=answered:88/90|dimension:79/80|overall:null|dimension_null:1|group_null:1|validity:9/10,incomplete,null|ranking:0|report:rejected')
    finally:
        cleanup_exam(exam_id, paper_id, candidate_id)


def verify_questionable(admin, suffix):
    exam_id = candidate_id = paper_id = ''
    try:
        exam_id = create_exam(admin, 'PHASE1-QUESTIONABLE-' + suffix)
        publish(admin, exam_id)
        candidate_id, paper_id, paper_token = create_candidate_and_paper(admin, exam_id, suffix[::-1], 'Phase1 Questionable Verify')
        answer_paper(paper_id, paper_token, 4)
        submitted = paper_api('/exam/api/competency/participant/submit', paper_token, {'paperId': paper_id, 'submitType': 'manual'})
        if not submitted.get('isComplete'):
            raise RuntimeError('questionable submission incomplete')
        result = api('/exam/api/competency/results/detail', admin, {'paperId': paper_id})
        validity = result.get('validity') or {}
        if float(validity.get('validityScore')) != 40 or validity.get('validityStatus') != 'questionable' or validity.get('isComplete') != 1:
            raise RuntimeError(f'questionable validity mismatch: {validity}')
        if float(result['result']['overallScore']) != 30 or result['result'].get('evaluationLevel') != 'weak' or result['result'].get('isComplete') != 1:
            raise RuntimeError(f'questionable overall mismatch: {result["result"]}')
        all_rows = api('/exam/api/competency/results/paging', admin, {'examId': exam_id, 'current': 1, 'size': 20, 'completion': 'all', 'validity': 'all', 'sortBy': 'overallScore'})
        questionable = api('/exam/api/competency/results/paging', admin, {'examId': exam_id, 'current': 1, 'size': 20, 'validity': 'questionable', 'sortBy': 'submittedAt'})
        ranking = api('/exam/api/competency/results/paging', admin, {'examId': exam_id, 'current': 1, 'size': 20, 'sortBy': 'overallScore'})
        if all_rows.get('total') != 1 or questionable.get('total') != 1 or ranking.get('total') != 0:
            raise RuntimeError('questionable filters/default ranking mismatch')
        xml = download_export(admin, exam_id)
        for required in ['效度原始分', '效度状态', 'questionable']:
            if required not in xml:
                raise RuntimeError(f'questionable export missing {required}')
        print('VALIDITY_QUESTIONABLE_PASS')
        print('questionable=score:40|status:questionable|overall:30/weak|explicit_all:1|filter:1|default_ranking:0|export:true')
    finally:
        cleanup_exam(exam_id, paper_id, candidate_id)


def verify_publish_rollback(admin, suffix):
    exam_id = ''
    try:
        exam_id = create_exam(admin, 'PHASE1-ROLLBACK-' + suffix)
        mysql_exec(f"DROP TRIGGER IF EXISTS {TRIGGER};")
        mysql_exec(f"""
CREATE TRIGGER {TRIGGER}
BEFORE INSERT ON el_exam_competency_question
FOR EACH ROW
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='PHASE1_RUNTIME_FORCED_SNAPSHOT_FAILURE';
""")
        rejection = expect_business_error('/exam/api/competency/exams/publish', admin, {'examId': exam_id}, contains='PHASE1_RUNTIME_FORCED_SNAPSHOT_FAILURE')
        counts = mysql(f"""
SELECT CONCAT(
  (SELECT publish_status FROM el_exam WHERE id='{exam_id}'),'|',
  (SELECT COUNT(*) FROM el_exam_competency_group WHERE exam_id='{exam_id}'),'|',
  (SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id='{exam_id}'),'|',
  (SELECT COUNT(*) FROM el_exam_competency_dimension WHERE exam_id='{exam_id}' AND group_id IS NOT NULL))
""")
        if counts != '0|0|0|0':
            raise RuntimeError(f'publish rollback residue: {counts}')
        mysql_exec(f"DROP TRIGGER IF EXISTS {TRIGGER};")
        successful = publish(admin, exam_id)
        if successful.get('questionCount') != 90:
            raise RuntimeError('draft not reusable after rollback')
        print('PUBLISH_ROLLBACK_PASS')
        print('rollback=forced_snapshot_failure:true|publish_status:0|groups:0|questions:0|group_links:0|draft_reusable:true')
    finally:
        mysql_exec(f"DROP TRIGGER IF EXISTS {TRIGGER};")
        cleanup_exam(exam_id)


def main():
    suffix = uuid.uuid4().hex[:12]
    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    token_id = 'phase1-negative-' + suffix
    redis_key = 'login_tokens:' + token_id
    now_ms = int(time.time() * 1000)
    login_user = {
        'userId': 1, 'token': token_id, 'loginTime': now_ms,
        'expireTime': now_ms + 3600000, 'permissions': ['*:*:*'], 'roles': ['admin'], 'user': None,
    }
    subprocess.run(
        ['redis-cli', '-n', redis_db, 'SET', redis_key, json.dumps(login_user, separators=(',', ':')), 'EX', '3600'],
        check=True, stdout=subprocess.DEVNULL,
    )
    admin = sign(secret, token_id)
    try:
        baseline = mysql("""
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_exam WHERE assessment_type='competency'),'|',
  (SELECT COUNT(*) FROM el_competency_result),'|',
  (SELECT COUNT(*) FROM information_schema.TRIGGERS WHERE TRIGGER_SCHEMA=DATABASE() AND TRIGGER_NAME='phase1_runtime_negative_question_fail'))
""")
        if baseline != '0|0|0':
            raise RuntimeError(f'refusing dirty competency baseline: {baseline}')
        verify_timeout_incomplete(admin, suffix)
        verify_questionable(admin, suffix)
        verify_publish_rollback(admin, suffix)
        print('STAGING_PHASE1_NEGATIVE_RUNTIME_PASS')
    finally:
        mysql_exec(f"DROP TRIGGER IF EXISTS {TRIGGER};")
        subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)
        remaining = mysql("""
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_exam WHERE assessment_type='competency'),'|',
  (SELECT COUNT(*) FROM el_competency_result),'|',
  (SELECT COUNT(*) FROM el_competency_group_result),'|',
  (SELECT COUNT(*) FROM el_competency_validity_result),'|',
  (SELECT COUNT(*) FROM information_schema.TRIGGERS WHERE TRIGGER_SCHEMA=DATABASE() AND TRIGGER_NAME='phase1_runtime_negative_question_fail'))
""")
        print('cleanup_remaining=' + remaining)


if __name__ == '__main__':
    main()
