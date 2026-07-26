#!/usr/bin/env python3
import argparse, base64, hashlib, hmac, json, re, subprocess, time, urllib.error, urllib.request, uuid
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'
STATE = Path('/tmp/competency-results-verify-state.json')


def b64url(value):
    return base64.urlsafe_b64encode(value).rstrip(b'=').decode()


def config_text():
    env = subprocess.run(['systemctl', 'show', 'talent-assessment', '--property=Environment', '--value'], check=True, capture_output=True, text=True).stdout
    match = re.search(r'(?:^|\s)APP_ENV=([^\s]+)', env)
    name = match.group(1) if match else 'local'
    files = [APP_DIR / 'configs' / 'application.yml', APP_DIR / 'configs' / f'application-{name}.yml']
    return '\n'.join(path.read_text(encoding='utf-8') for path in files if path.exists())


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
    signature = b64url(hmac.new(secret.encode(), raw.encode(), hashlib.sha512).digest())
    return f'{raw}.{signature}'


def api(path, body, token=None, headers=None, method='POST'):
    request_headers = {'Content-Type': 'application/json'}
    if token:
        request_headers['Authorization'] = 'Bearer ' + token
    if headers:
        request_headers.update(headers)
    data = None if method == 'GET' else json.dumps(body, ensure_ascii=False).encode()
    request = urllib.request.Request(BASE + path, data=data, headers=request_headers, method=method)
    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            payload = json.loads(response.read().decode())
    except urllib.error.HTTPError as error:
        raise RuntimeError(f'{path} HTTP {error.code}: {error.read().decode(errors="replace")}')
    if payload.get('code') not in (0, 200):
        raise RuntimeError(f'{path}: {payload.get("msg")}')
    return payload.get('data')


def download(path, token):
    request = urllib.request.Request(BASE + path, headers={'Authorization': 'Bearer ' + token}, method='GET')
    with urllib.request.urlopen(request, timeout=60) as response:
        return response.read(), dict(response.headers)


def mysql(sql):
    return subprocess.run(['sudo', '-n', 'mysql', 'element', '-Nse', sql], check=True, capture_output=True, text=True).stdout.strip()


def remove_session(redis_db, redis_key):
    subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)


def assert_order(data, expected_names, score_field):
    records = data.get('records') or []
    names = [row['participantName'] for row in records]
    if names != expected_names:
        raise RuntimeError(f'{score_field} order mismatch: {names}, want {expected_names}')
    if any(not row.get('participantTelephone') or row.get('participantType') != 'candidate' for row in records):
        raise RuntimeError(f'{score_field} participant projection missing: {records}')
    return records


def cleanup(state):
    exam_id = state['examId']
    try:
        api('/exam/api/exam/exam/delete', {'ids': [exam_id]}, state['adminToken'])
    finally:
        remove_session(state['redisDb'], state['redisKey'])
    remaining = mysql(
        f"SELECT (SELECT COUNT(*) FROM el_exam WHERE id='{exam_id}')+"
        f"(SELECT COUNT(*) FROM el_paper WHERE exam_id='{exam_id}')+"
        f"(SELECT COUNT(*) FROM el_competency_result WHERE exam_id='{exam_id}')+"
        f"(SELECT COUNT(*) FROM el_candidate WHERE exam_id='{exam_id}')"
    )
    if remaining != '0':
        raise RuntimeError('cleanup remaining=' + remaining)
    STATE.unlink(missing_ok=True)
    print('COMPETENCY_RESULTS_CLEANUP_OK')
    print('cleanup_remaining=0')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--keep', action='store_true')
    parser.add_argument('--cleanup', action='store_true')
    args = parser.parse_args()
    if args.cleanup:
        cleanup(json.loads(STATE.read_text(encoding='utf-8')))
        return

    suffix = uuid.uuid4().hex[:10]
    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    token_id = 'competency-results-' + suffix
    redis_key = 'login_tokens:' + token_id
    now_ms = int(time.time() * 1000)
    login_user = {'userId': 1, 'token': token_id, 'loginTime': now_ms, 'expireTime': now_ms + 1800000, 'permissions': ['*:*:*'], 'roles': ['admin'], 'user': None}
    subprocess.run(['redis-cli', '-n', redis_db, 'SET', redis_key, json.dumps(login_user, separators=(',', ':')), 'EX', '1800'], check=True, stdout=subprocess.DEVNULL)
    admin = sign(secret, token_id)
    exam_id = None
    state = None
    try:
        dimensions = api('/exam/api/competency/dimensions/list', {}, admin)
        selected = [item['id'] for item in dimensions if item['code'] in ('D01', 'D02')]
        if selected != ['competency-d01', 'competency-d02']:
            raise RuntimeError(f'D01/D02 dimensions missing: {selected}')
        exam = api('/exam/api/exam/exam/save', {
            'title': 'RESULT-SORT-' + suffix, 'content': 'temporary result management verification',
            'assessmentType': 'competency', 'scoringMode': 'competency_average',
            'competencyReportAudience': 'leader', 'dimensionIds': selected,
            'joinType': 1, 'openType': 1, 'isOpen': 1, 'answerType': 1,
            'state': 0, 'totalTime': 30, 'repoList': [], 'departIds': []
        }, admin)
        exam_id = exam['id']
        api('/exam/api/competency/exams/publish', {'examId': exam_id}, admin)

        profiles = [
            ('Result Low', '13800000001', {'competency-d01': 1, 'competency-d02': 1}),
            ('Result Mixed', '13800000002', {'competency-d01': 4, 'competency-d02': 2}),
            ('Result High', '13800000003', {'competency-d01': 5, 'competency-d02': 3}),
        ]
        paper_ids = []
        for name, telephone, targets in profiles:
            candidate = api('/exam/api/candidate/save', {'examId': exam_id, 'name': name, 'telephone': telephone, 'gender': '0'}, admin)
            access = api('/exam/api/competency/participant/create-paper', {
                'examId': exam_id, 'participantId': candidate['id'], 'participantType': 'candidate',
                'participantToken': candidate['participantToken']
            })
            paper_id = access['paperId']
            paper_ids.append(paper_id)
            paper = api('/exam/api/competency/participant/paper-detail', {'paperId': paper_id}, headers={'X-Competency-Token': access['paperToken']})
            rows = mysql(
                "SELECT CONCAT(pq.id,'|',d.dimension_id,'|',q.scoring_direction) "
                "FROM el_paper_qu pq INNER JOIN el_exam_competency_question q ON q.id=pq.exam_question_id "
                "INNER JOIN el_exam_competency_dimension d ON d.id=q.exam_dimension_id "
                f"WHERE pq.paper_id='{paper_id}'"
            ).splitlines()
            scoring = {}
            for row in rows:
                question_id, dimension_id, direction = row.split('|')
                scoring[question_id] = (dimension_id, direction)
            for question in paper['questions']:
                dimension_id, direction = scoring[question['id']]
                target = targets[dimension_id]
                raw_value = target if direction == 'forward' else 6 - target
                api('/exam/api/competency/participant/fill-answer', {
                    'paperId': paper_id, 'paperQuestionId': question['id'], 'rawValue': raw_value
                }, headers={'X-Competency-Token': access['paperToken']})
            submitted = api('/exam/api/competency/participant/submit', {'paperId': paper_id, 'submitType': 'manual'}, headers={'X-Competency-Token': access['paperToken']})
            if not submitted['isComplete']:
                raise RuntimeError(name + ' submission incomplete')

        for index, paper_id in enumerate(paper_ids, start=1):
            mysql(f"UPDATE el_competency_result SET submitted_at='2026-07-25 12:0{index}:00' WHERE paper_id='{paper_id}'")

        default_page = api('/exam/api/competency/results/paging', {'examId': exam_id, 'current': 1, 'size': 10}, admin)
        assert_order(default_page, ['Result High', 'Result Mixed', 'Result Low'], 'submittedAt desc')
        overall_asc = api('/exam/api/competency/results/paging', {'examId': exam_id, 'current': 1, 'size': 10, 'sortBy': 'overallScore', 'sortDirection': 'asc'}, admin)
        overall_rows = assert_order(overall_asc, ['Result Low', 'Result Mixed', 'Result High'], 'overallScore asc')
        dimension_desc = api('/exam/api/competency/results/paging', {'examId': exam_id, 'current': 1, 'size': 10, 'sortBy': 'dimensionScore', 'sortDirection': 'desc', 'dimensionId': 'competency-d01'}, admin)
        dimension_rows = assert_order(dimension_desc, ['Result High', 'Result Mixed', 'Result Low'], 'D01 desc')
        detail = api('/exam/api/competency/results/detail', {'paperId': paper_ids[2]}, admin)
        if len(detail.get('dimensions') or []) != 2 or len(detail.get('questions') or []) != 16:
            raise RuntimeError(f'detail mismatch: dimensions={len(detail.get("dimensions") or [])}, questions={len(detail.get("questions") or [])}')
        if [float(row['overallScore']) for row in overall_rows] != [2.0, 6.0, 8.0]:
            raise RuntimeError('overall score values mismatch')
        if [float(row['sortDimensionScore']) for row in dimension_rows] != [5.0, 4.0, 1.0]:
            raise RuntimeError('dimension score values mismatch')

        report = api('/exam/api/competency/reports/generate', {'paperId': paper_ids[2], 'force': False}, admin)
        if report.get('status') != 'completed' or report.get('contentVersion') != 'temp-v1' or not report.get('pdfSha256'):
            raise RuntimeError(f'report metadata mismatch: {report}')
        pdf, headers = download('/exam/api/competency/reports/download?paperId=' + paper_ids[2], admin)
        if not pdf.startswith(b'%PDF') or len(pdf) < 1024:
            raise RuntimeError(f'invalid competency report PDF: {len(pdf)} bytes')
        if hashlib.sha256(pdf).hexdigest() != report['pdfSha256']:
            raise RuntimeError('competency report PDF hash mismatch')
        pdf_path = Path('/tmp/competency-report-verify.pdf')
        pdf_path.write_bytes(pdf)
        try:
            text_result = subprocess.run(['pdftotext', str(pdf_path), '-'], check=True, capture_output=True, text=True)
            if '临时测试报告' not in text_result.stdout or '不可作为人才决策依据' not in text_result.stdout:
                raise RuntimeError('temporary competency disclaimer missing from PDF text: ' + repr(text_result.stdout[:500]))
        finally:
            pdf_path.unlink(missing_ok=True)
        report_counts = mysql(
            f"SELECT (SELECT COUNT(*) FROM el_competency_report WHERE paper_id='{paper_ids[2]}' AND status='completed'),"
            f"(SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id='{paper_ids[2]}' AND action='generate' AND status=1),"
            f"(SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id='{paper_ids[2]}' AND action='download' AND status=1)"
        )
        if report_counts != '1\t1\t1':
            raise RuntimeError('report/audit counts mismatch: ' + report_counts)

        state = {'examId': exam_id, 'examTitle': 'RESULT-SORT-' + suffix, 'adminToken': admin, 'redisDb': redis_db, 'redisKey': redis_key}
        STATE.write_text(json.dumps(state), encoding='utf-8')
        STATE.chmod(0o600)
        print('COMPETENCY_RESULTS_VERIFY_OK')
        print('result_count=3')
        print('submitted_desc=Result High|Result Mixed|Result Low')
        print('overall_asc=2.000000|6.000000|8.000000')
        print('dimension_d01_desc=5.000000|4.000000|1.000000')
        print('participant_projection=candidate|name|telephone')
        print('detail=2_dimensions|16_questions')
        print('temporary_report=pdf|hash|disclaimer|audit')
        print('exam_id=' + exam_id)
        if not args.keep:
            cleanup(state)
    except Exception:
        if exam_id and state is None:
            try:
                api('/exam/api/exam/exam/delete', {'ids': [exam_id]}, admin)
            except Exception:
                pass
        remove_session(redis_db, redis_key)
        raise


if __name__ == '__main__':
    main()
