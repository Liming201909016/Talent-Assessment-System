#!/usr/bin/env python3
import base64
import hashlib
import hmac
import json
import re
import shutil
import subprocess
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'
OUTPUT_DIR = Path('/tmp/sc012-dual-audience')
DIMENSION_IDS = ['competency-d01', 'competency-d02', 'competency-d05', 'competency-d06', 'competency-d32']
CASES = [
    {'audience': 'frontline_employee', 'label': '基层员工版'},
    {'audience': 'leader', 'label': '领导人员版'},
]
RESULT_FIELDS = [
    'totalQuestionCount', 'answeredQuestionCount', 'effectiveDimensionCount',
    'overallScore', 'evaluationAverage', 'evaluationLevel', 'isComplete',
    'submitType', 'scoringVersion', 'participantName', 'participantTelephone',
    'participantAge', 'participantGender', 'participantAffiliation', 'participantPost',
]
DIMENSION_FIELDS = [
    'dimensionId', 'dimensionCode', 'dimensionName', 'displayOrder',
    'totalQuestionCount', 'answeredQuestionCount', 'scoreSum',
    'dimensionScore', 'levelCode', 'isComplete',
]


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
        return response.read()


def mysql(sql):
    return subprocess.run(
        ['sudo', '-n', 'mysql', 'element', '--batch', '--skip-column-names', '-e', sql],
        check=True, capture_output=True, text=True,
    ).stdout.strip()


def sql_quote(value):
    return "'" + str(value).replace("'", "''") + "'"


def fixed_raw_value(question_code):
    match = re.fullmatch(r'D\d{2}-Q(\d{2})', question_code)
    if not match:
        raise RuntimeError(f'unexpected question code: {question_code}')
    return ((int(match.group(1)) - 1) % 5) + 1


def canonical_result(detail):
    result = detail['result']
    result_values = {key: result.get(key) for key in RESULT_FIELDS}
    dimensions = [
        {key: row.get(key) for key in DIMENSION_FIELDS}
        for row in detail.get('dimensions') or []
    ]
    return {'result': result_values, 'dimensions': dimensions}


def report_text_snapshot(paper_id):
    encoded = mysql(
        "SELECT HEX(text_snapshot) FROM el_competency_report "
        f"WHERE paper_id={sql_quote(paper_id)} AND content_version='temp-v1' AND status='completed'"
    )
    if not encoded:
        raise RuntimeError(f'report text snapshot missing: {paper_id}')
    return json.loads(bytes.fromhex(encoded).decode('utf-8'))


def expected_report_text(audience, content_type, dimension_id, level_code):
    encoded = mysql(
        "SELECT HEX(content) FROM el_competency_report_text WHERE content_version='temp-v1' "
        f"AND audience={sql_quote(audience)} AND content_type={sql_quote(content_type)} "
        f"AND dimension_id={sql_quote(dimension_id)} AND level_code={sql_quote(level_code)} AND status=0"
    )
    if not encoded:
        raise RuntimeError(f'report source text missing: {audience}|{content_type}|{dimension_id}|{level_code}')
    return bytes.fromhex(encoded).decode('utf-8')


def verify_text_snapshot(snapshot, audience, detail):
    if snapshot.get('contentVersion') != 'temp-v1' or snapshot.get('audience') != audience:
        raise RuntimeError(f'text snapshot audience/version mismatch: {snapshot}')
    result = detail['result']
    expected_overall = expected_report_text(audience, 'overall', '', result['evaluationLevel'])
    if snapshot.get('overallText') != expected_overall:
        raise RuntimeError(f'overall text does not match exact audience row: {audience}')
    dimension_texts = snapshot.get('dimensionTexts') or {}
    dimensions = detail.get('dimensions') or []
    if set(dimension_texts) != {row['dimensionId'] for row in dimensions}:
        raise RuntimeError(f'dimension text keys mismatch: {audience}')
    for row in dimensions:
        expected = expected_report_text(audience, 'dimension', row['dimensionId'], row['levelCode'])
        if dimension_texts[row['dimensionId']] != expected:
            raise RuntimeError(f'dimension text does not match exact audience row: {audience}|{row["dimensionId"]}')


def pdf_info(pdf_path):
    output = subprocess.run(['pdfinfo', str(pdf_path)], check=True, capture_output=True, text=True).stdout
    pages_match = re.search(r'^Pages:\s+(\d+)', output, re.MULTILINE)
    size_match = re.search(r'^Page size:\s+([\d.]+) x ([\d.]+) pts', output, re.MULTILINE)
    if not pages_match or not size_match:
        raise RuntimeError(f'cannot parse pdfinfo: {pdf_path}')
    return int(pages_match.group(1)), (float(size_match.group(1)), float(size_match.group(2)))


def pdf_pages_text(pdf_path):
    output = subprocess.run(
        ['pdftotext', '-layout', str(pdf_path), '-'],
        check=True, capture_output=True, text=True,
    ).stdout
    pages = output.split('\f')
    if pages and not pages[-1].strip():
        pages.pop()
    return pages


def normalize_page_text(value):
    value = value.replace('基层员工版', '<AUDIENCE>').replace('领导人员版', '<AUDIENCE>')
    value = re.sub(r'\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}', '<DATETIME>', value)
    value = re.sub(r'\d{4}年\d{1,2}月\d{1,2}日', '<DATE>', value)
    value = re.sub(r'[ \t]+', ' ', value)
    value = re.sub(r' *\n *', '\n', value)
    return value.strip()


def ppm_pixels(path):
    data = path.read_bytes()
    match = re.match(br'P6\s+(?:#[^\n]*\s+)*(\d+)\s+(\d+)\s+(\d+)\s', data)
    if not match or int(match.group(3)) != 255:
        raise RuntimeError(f'unsupported PPM: {path}')
    return int(match.group(1)), int(match.group(2)), data[match.end():]


def rasterize(pdf_path, prefix):
    subprocess.run(['pdftoppm', '-r', '72', str(pdf_path), str(prefix)], check=True, stdout=subprocess.DEVNULL)
    return sorted(prefix.parent.glob(prefix.name + '-*.ppm'))


def compare_rasters(left_pdf, right_pdf):
    left_pages = rasterize(left_pdf, OUTPUT_DIR / 'frontline-page')
    right_pages = rasterize(right_pdf, OUTPUT_DIR / 'leader-page')
    if len(left_pages) != len(right_pages):
        raise RuntimeError('raster page count mismatch')
    ratios = []
    for index, (left, right) in enumerate(zip(left_pages, right_pages), 1):
        left_width, left_height, left_data = ppm_pixels(left)
        right_width, right_height, right_data = ppm_pixels(right)
        if (left_width, left_height, len(left_data)) != (right_width, right_height, len(right_data)):
            raise RuntimeError(f'raster dimensions mismatch on page {index}')
        pixels = left_width * left_height
        different = sum(
            left_data[offset:offset + 3] != right_data[offset:offset + 3]
            for offset in range(0, len(left_data), 3)
        )
        ratio = different / pixels
        ratios.append(ratio)
        if ratio > 0.015:
            raise RuntimeError(f'visual mismatch ratio too high on page {index}: {ratio:.6f}')
    return ratios


def cleanup_created(exam_ids, admin):
    errors = []
    for exam_id in reversed(exam_ids):
        try:
            api('/exam/api/exam/exam/delete', {'ids': [exam_id]}, admin)
        except Exception as error:
            errors.append(f'{exam_id}:{error}')
    return errors


def main():
    suffix = uuid.uuid4().hex[:8]
    title = f'SC-012 同答案双受众-{suffix}'
    participant_name = f'SC012受测者-{suffix}'
    telephone = '139' + str(int(time.time()))[-8:]
    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    token_id = 'sc012-' + suffix
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
    exam_ids = []
    records = []
    cleanup_errors = []
    if OUTPUT_DIR.exists():
        shutil.rmtree(OUTPUT_DIR)
    OUTPUT_DIR.mkdir(parents=True)
    try:
        for case in CASES:
            exam = api('/exam/api/exam/exam/save', {
                'title': title,
                'content': 'SC-012 identical-answer dual-audience report verification',
                'assessmentType': 'competency', 'scoringMode': 'competency_average',
                'competencyReportAudience': case['audience'], 'dimensionIds': DIMENSION_IDS,
                'requiredFields': 'name,gender,age,telephone,affiliation,post',
                'joinType': 1, 'openType': 1, 'isOpen': 1, 'answerType': 1,
                'state': 0, 'totalTime': 30, 'repoList': [], 'departIds': [],
            }, admin)
            exam_id = exam['id']
            exam_ids.append(exam_id)
            published = api('/exam/api/competency/exams/publish', {'examId': exam_id}, admin)
            if published.get('dimensionCount') != 5 or published.get('questionCount') != 40:
                raise RuntimeError(f'publish mismatch: {case["audience"]}|{published}')
            candidate = api('/exam/api/candidate/save', {
                'examId': exam_id, 'name': participant_name, 'telephone': telephone,
                'gender': '0', 'age': 30, 'affiliation': 'SC-012测试单位', 'post': 'SC-012测试岗位',
            }, admin)
            access = api('/exam/api/competency/participant/create-paper', {
                'examId': exam_id, 'participantId': candidate['id'], 'participantType': 'candidate',
                'participantToken': candidate['participantToken'],
            })
            paper_id = access['paperId']
            paper_token = access['paperToken']
            paper = api('/exam/api/competency/participant/paper-detail', {'paperId': paper_id}, headers={'X-Competency-Token': paper_token})
            if paper.get('totalCount') != 40 or len(paper.get('questions') or []) != 40:
                raise RuntimeError(f'paper size mismatch: {case["audience"]}')
            rows = mysql(
                "SELECT CONCAT(pq.id,'|',q.question_code,'|',q.scoring_direction) FROM el_paper_qu pq "
                "INNER JOIN el_exam_competency_question q ON q.id=pq.exam_question_id "
                f"WHERE pq.paper_id={sql_quote(paper_id)}"
            ).splitlines()
            question_map = {}
            for row in rows:
                paper_question_id, question_code, direction = row.split('|')
                question_map[paper_question_id] = (question_code, direction)
            if len(question_map) != 40:
                raise RuntimeError(f'question metadata mismatch: {case["audience"]}')
            submitted_raw = {}
            for question in paper['questions']:
                question_code, _ = question_map[question['id']]
                raw_value = fixed_raw_value(question_code)
                submitted_raw[question_code] = raw_value
                api('/exam/api/competency/participant/fill-answer', {
                    'paperId': paper_id, 'paperQuestionId': question['id'], 'rawValue': raw_value,
                }, headers={'X-Competency-Token': paper_token})
            submitted = api('/exam/api/competency/participant/submit', {
                'paperId': paper_id, 'submitType': 'manual',
            }, headers={'X-Competency-Token': paper_token})
            if not submitted.get('isComplete'):
                raise RuntimeError(f'incomplete submission: {case["audience"]}')
            detail = api('/exam/api/competency/results/detail', {'paperId': paper_id}, admin)
            report = api('/exam/api/competency/reports/generate', {'paperId': paper_id, 'force': True}, admin)
            if report.get('status') != 'completed' or report.get('audience') != case['audience']:
                raise RuntimeError(f'report generation mismatch: {case["audience"]}|{report}')
            pdf = download('/exam/api/competency/reports/download?paperId=' + urllib.parse.quote(paper_id), admin)
            digest = hashlib.sha256(pdf).hexdigest()
            if not pdf.startswith(b'%PDF') or digest != report.get('pdfSha256'):
                raise RuntimeError(f'PDF/hash mismatch: {case["audience"]}')
            pdf_path = OUTPUT_DIR / (case['audience'] + '.pdf')
            pdf_path.write_bytes(pdf)
            snapshot = report_text_snapshot(paper_id)
            verify_text_snapshot(snapshot, case['audience'], detail)
            records.append({
                'case': case, 'examId': exam_id, 'paperId': paper_id,
                'raw': submitted_raw, 'detail': detail, 'canonical': canonical_result(detail),
                'snapshot': snapshot, 'pdf': pdf_path, 'sha256': digest,
            })

        left, right = records
        if left['raw'] != right['raw'] or len(left['raw']) != 40:
            raise RuntimeError('raw answers are not identical by question code')
        if left['canonical'] != right['canonical']:
            raise RuntimeError('persisted score facts or dimension order differ across audiences')
        if left['detail']['result']['reportAudience'] != 'frontline_employee' or right['detail']['result']['reportAudience'] != 'leader':
            raise RuntimeError('result audience snapshots are incorrect')

        left_pages, left_size = pdf_info(left['pdf'])
        right_pages, right_size = pdf_info(right['pdf'])
        if left_pages != 9 or right_pages != 9 or left_size != right_size:
            raise RuntimeError(f'PDF structure mismatch: pages={left_pages}/{right_pages}, size={left_size}/{right_size}')
        if abs(left_size[0] - 595.92) > 0.1 or abs(left_size[1] - 841.92) > 0.1:
            raise RuntimeError(f'PDF is not A4: {left_size}')
        left_text = [normalize_page_text(page) for page in pdf_pages_text(left['pdf'])]
        right_text = [normalize_page_text(page) for page in pdf_pages_text(right['pdf'])]
        if left_text != right_text:
            different_pages = [str(index + 1) for index, pair in enumerate(zip(left_text, right_text)) if pair[0] != pair[1]]
            raise RuntimeError('normalized PDF page text differs: pages=' + ','.join(different_pages))
        raster_ratios = compare_rasters(left['pdf'], right['pdf'])

        score = left['detail']['result']
        dimension_order = [row['dimensionCode'] for row in left['detail']['dimensions']]
        print('SC012_DUAL_AUDIENCE_REPORT_PASS')
        print('raw_answers_identical=40/40')
        print(f'overall_score={score["overallScore"]}|evaluation_average={score["evaluationAverage"]}|level={score["evaluationLevel"]}')
        print('dimension_order=' + ','.join(dimension_order))
        print('dimension_facts_identical=5/5')
        print('text_exact_audience_match=overall:2/2|dimensions:10/10')
        print('pdf_pages=9/9|page_size=A4|normalized_text_pages=9/9')
        print('raster_mismatch_max=' + format(max(raster_ratios), '.6f'))
        print('raster_mismatch_by_page=' + ','.join(format(value, '.6f') for value in raster_ratios))
        print('pdf_sha256_frontline=' + left['sha256'])
        print('pdf_sha256_leader=' + right['sha256'])
    finally:
        cleanup_errors = cleanup_created(exam_ids, admin)
        remaining = mysql(
            "SELECT CONCAT("
            f"(SELECT COUNT(*) FROM el_exam WHERE title={sql_quote(title)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_result r INNER JOIN el_exam e ON e.id=r.exam_id WHERE e.title={sql_quote(title)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_report r INNER JOIN el_exam e ON e.id=r.exam_id WHERE e.title={sql_quote(title)})"
            ")"
        )
        subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)
        shutil.rmtree(OUTPUT_DIR, ignore_errors=True)
        print('cleanup_exam_result_report=' + remaining)
        print('cleanup_session=done')
        if cleanup_errors:
            raise RuntimeError('cleanup failures: ' + ';'.join(cleanup_errors))
        if remaining != '0|0|0':
            raise RuntimeError('SC-012 cleanup incomplete: ' + remaining)


if __name__ == '__main__':
    main()
