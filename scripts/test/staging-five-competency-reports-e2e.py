#!/usr/bin/env python3
import base64
import hashlib
import hmac
import json
import re
import subprocess
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'
OUTPUT_DIR = Path('/tmp/five-competency-reports')
SUMMARY = Path('/tmp/five-competency-reports-summary.json')
DIMENSION_IDS = ['competency-d01', 'competency-d02', 'competency-d05', 'competency-d06', 'competency-d32']
CASES = [
    {'index': 1, 'audience': 'frontline_employee', 'target': 1, 'label': '较低'},
    {'index': 2, 'audience': 'frontline_employee', 'target': 2, 'label': '一般'},
    {'index': 3, 'audience': 'frontline_employee', 'target': 3, 'label': '良好'},
    {'index': 4, 'audience': 'leader', 'target': 4, 'label': '较高'},
    {'index': 5, 'audience': 'leader', 'target': 5, 'label': '较高'},
]


def b64url(value):
    return base64.urlsafe_b64encode(value).rstrip(b'=').decode()


def config_text():
    env = subprocess.run(
        ['systemctl', 'show', 'talent-assessment', '--property=Environment', '--value'],
        check=True, capture_output=True, text=True,
    ).stdout
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


def api(path, body=None, token=None, headers=None, method='POST', timeout=120):
    request_headers = {'Content-Type': 'application/json'}
    if token:
        request_headers['Authorization'] = 'Bearer ' + token
    if headers:
        request_headers.update(headers)
    data = None if method == 'GET' else json.dumps(body or {}, ensure_ascii=False).encode()
    request = urllib.request.Request(BASE + path, data=data, headers=request_headers, method=method)
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            payload = json.loads(response.read().decode())
    except urllib.error.HTTPError as error:
        raise RuntimeError(f'{path} HTTP {error.code}: {error.read().decode(errors="replace")}')
    if payload.get('code') not in (0, 200):
        raise RuntimeError(f'{path}: {payload.get("msg")}')
    return payload.get('data')


def download(path, token):
    request = urllib.request.Request(BASE + path, headers={'Authorization': 'Bearer ' + token}, method='GET')
    with urllib.request.urlopen(request, timeout=120) as response:
        return response.read(), dict(response.headers)


def mysql(sql):
    return subprocess.run(
        ['sudo', '-n', 'mysql', 'element', '-Nse', sql],
        check=True, capture_output=True, text=True,
    ).stdout.strip()


def remove_session(redis_db, redis_key):
    subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)


def inspect_pdf(pdf_path, case, expected):
    info = subprocess.run(['pdfinfo', str(pdf_path)], check=True, capture_output=True, text=True).stdout
    page_match = re.search(r'^Pages:\s+(\d+)', info, re.MULTILINE)
    page_count = int(page_match.group(1)) if page_match else 0
    expected_pages = 4 + len(DIMENSION_IDS)
    if page_count != expected_pages:
        raise RuntimeError(f'{pdf_path.name} pages={page_count}, want {expected_pages}')
    text = subprocess.run(['pdftotext', str(pdf_path), '-'], check=True, capture_output=True, text=True).stdout
    required = [
        case['examTitle'], case['participantName'], case['audienceLabel'], '临时测试报告',
        '报告阅读说明', '总体评价', '测评结果分析', '核心含义', '发展提示',
        '不可作为人才决策依据', '沟通表达', '人际交往', '逻辑思维', '持续学习', '严谨性',
    ]
    missing = [item for item in required if item not in text]
    if missing:
        raise RuntimeError(f'{pdf_path.name} missing text: {missing}')
    if str(expected['score']) not in text or case['label'] not in text:
        raise RuntimeError(f'{pdf_path.name} score/level text mismatch')
    return page_count


def cleanup_created(exam_ids, admin):
    for exam_id in reversed(exam_ids):
        try:
            api('/exam/api/exam/exam/delete', {'ids': [exam_id]}, admin)
        except Exception as error:
            print(f'cleanup_failed exam_id={exam_id} error={error}')


def main():
    suffix = uuid.uuid4().hex[:8]
    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    token_id = 'five-competency-' + suffix
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
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    for old in OUTPUT_DIR.glob('*.pdf'):
        old.unlink()
    exam_ids = []
    results = []
    succeeded = False
    try:
        dimensions = api('/exam/api/competency/dimensions/list', {}, admin)
        available = {item['id'] for item in dimensions}
        if any(dimension_id not in available for dimension_id in DIMENSION_IDS):
            raise RuntimeError('required 00401 dimensions are unavailable')
        for case in CASES:
            index = case['index']
            exam_title = f'00401 五配置验收-{index:02d}-{suffix}'
            participant_name = f'五配置受测者{index:02d}'
            audience_label = '领导人员版' if case['audience'] == 'leader' else '基层员工版'
            exam = api('/exam/api/exam/exam/save', {
                'title': exam_title,
                'content': '00401 five-config full assessment and report verification',
                'assessmentType': 'competency',
                'scoringMode': 'competency_average',
                'competencyReportAudience': case['audience'],
                'dimensionIds': DIMENSION_IDS,
                'requiredFields': 'name,gender,age,telephone,affiliation,post',
                'joinType': 1, 'openType': 1, 'isOpen': 1, 'answerType': 1,
                'state': 0, 'totalTime': 30, 'repoList': [], 'departIds': [],
            }, admin)
            exam_id = exam['id']
            exam_ids.append(exam_id)
            published = api('/exam/api/competency/exams/publish', {'examId': exam_id}, admin)
            if published.get('dimensionCount') != 5 or published.get('questionCount') != 40:
                raise RuntimeError(f'{exam_title} publish mismatch: {published}')
            candidate = api('/exam/api/candidate/save', {
                'examId': exam_id, 'name': participant_name,
                'telephone': f'13900000{index:03d}', 'gender': str(index % 2), 'age': 24 + index,
                'affiliation': '00401测试单位', 'post': f'测试岗位{index:02d}',
            }, admin)
            access = api('/exam/api/competency/participant/create-paper', {
                'examId': exam_id, 'participantId': candidate['id'], 'participantType': 'candidate',
                'participantToken': candidate['participantToken'],
            })
            paper_id = access['paperId']
            paper_token = access['paperToken']
            paper = api('/exam/api/competency/participant/paper-detail', {'paperId': paper_id}, headers={'X-Competency-Token': paper_token})
            if paper.get('totalCount') != 40 or len(paper.get('questions') or []) != 40:
                raise RuntimeError(f'{exam_title} paper question count mismatch')
            scoring_rows = mysql(
                "SELECT CONCAT(pq.id,'|',q.scoring_direction) FROM el_paper_qu pq "
                "INNER JOIN el_exam_competency_question q ON q.id=pq.exam_question_id "
                f"WHERE pq.paper_id='{paper_id}'"
            ).splitlines()
            directions = dict(row.split('|') for row in scoring_rows)
            if len(directions) != 40:
                raise RuntimeError(f'{exam_title} scoring metadata mismatch')
            for question in paper['questions']:
                direction = directions[question['id']]
                raw_value = case['target'] if direction == 'forward' else 6 - case['target']
                api('/exam/api/competency/participant/fill-answer', {
                    'paperId': paper_id, 'paperQuestionId': question['id'], 'rawValue': raw_value,
                }, headers={'X-Competency-Token': paper_token})
            submitted = api('/exam/api/competency/participant/submit', {
                'paperId': paper_id, 'submitType': 'manual',
            }, headers={'X-Competency-Token': paper_token})
            if not submitted.get('isComplete'):
                raise RuntimeError(f'{exam_title} submission incomplete')
            detail = api('/exam/api/competency/results/detail', {'paperId': paper_id}, admin)
            result = detail['result']
            dimensions_result = detail.get('dimensions') or []
            if len(dimensions_result) != 5 or len(detail.get('questions') or []) != 40:
                raise RuntimeError(f'{exam_title} result detail mismatch')
            expected_score = float(case['target'])
            if float(result['overallScore']) != expected_score * 5 or float(result['evaluationAverage']) != expected_score:
                raise RuntimeError(f'{exam_title} score mismatch: {result}')
            if any(float(row['dimensionScore']) != expected_score for row in dimensions_result):
                raise RuntimeError(f'{exam_title} dimension score mismatch')
            report = api('/exam/api/competency/reports/generate', {'paperId': paper_id, 'force': True}, admin)
            if report.get('status') != 'completed' or report.get('contentVersion') != 'temp-v1':
                raise RuntimeError(f'{exam_title} report generation failed: {report}')
            pdf, headers = download('/exam/api/competency/reports/download?paperId=' + urllib.parse.quote(paper_id), admin)
            digest = hashlib.sha256(pdf).hexdigest()
            if not pdf.startswith(b'%PDF') or digest != report.get('pdfSha256'):
                raise RuntimeError(f'{exam_title} PDF/hash mismatch')
            file_name = f'{index:02d}-{audience_label}-{participant_name}.pdf'
            pdf_path = OUTPUT_DIR / file_name
            pdf_path.write_bytes(pdf)
            case_context = dict(case, examTitle=exam_title, participantName=participant_name, audienceLabel=audience_label)
            pages = inspect_pdf(pdf_path, case_context, {'score': expected_score})
            report_counts = mysql(
                f"SELECT CONCAT((SELECT COUNT(*) FROM el_competency_report WHERE paper_id='{paper_id}' AND status='completed'),'|',"
                f"(SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id='{paper_id}' AND action IN ('generate','regenerate') AND status=1),'|',"
                f"(SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id='{paper_id}' AND action='download' AND status=1))"
            )
            if report_counts != '1|1|1':
                raise RuntimeError(f'{exam_title} report audit mismatch: {report_counts}')
            results.append({
                'index': index, 'examId': exam_id, 'examTitle': exam_title,
                'participantId': candidate['id'], 'participantName': participant_name,
                'paperId': paper_id, 'audience': case['audience'], 'audienceLabel': audience_label,
                'targetScore': expected_score, 'overallScore': str(result['overallScore']),
                'evaluationAverage': str(result['evaluationAverage']), 'evaluationLevel': result['evaluationLevel'],
                'questionCount': 40, 'dimensionCount': 5, 'pages': pages,
                'pdfBytes': len(pdf), 'pdfSha256': digest, 'fileName': file_name,
            })
            print(f'CASE_{index:02d}_OK exam={exam_id} paper={paper_id} audience={case["audience"]} score={expected_score:.2f} pages={pages}')
        SUMMARY.write_text(json.dumps({'suffix': suffix, 'retained': True, 'reports': results}, ensure_ascii=False, indent=2), encoding='utf-8')
        SUMMARY.chmod(0o600)
        succeeded = True
        print('FIVE_COMPETENCY_REPORTS_E2E_OK')
        print('exam_count=5')
        print('complete_result_count=5')
        print('completed_report_count=5')
        print('pdf_count=5')
        print('retained=true')
    finally:
        if not succeeded:
            cleanup_created(exam_ids, admin)
        remove_session(redis_db, redis_key)


if __name__ == '__main__':
    main()
