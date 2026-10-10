#!/usr/bin/env python3
import base64
import hashlib
import hmac
import json
import os
import re
import subprocess
import time
import urllib.error
import urllib.request
import uuid
import zipfile
import xml.etree.ElementTree as ET
from io import BytesIO
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'


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


def request(path, token=None, method='POST', body=None, expect_success=True):
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
    if expect_success:
        data = json.loads(raw.decode())
        if status >= 400 or data.get('code') not in (0, 200):
            raise RuntimeError(f'{path}: HTTP {status}: {data.get("msg")}')
        return data.get('data') or {}, response_headers
    return status, raw, response_headers


def paper_request(path, paper_token, body):
    headers = {'Content-Type': 'application/json', 'X-Competency-Token': paper_token}
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


def xlsx_summary(payload):
    if not payload.startswith(b'PK'):
        raise RuntimeError('export is not an xlsx zip')
    with zipfile.ZipFile(BytesIO(payload)) as archive:
        names = set(archive.namelist())
        workbook = archive.read('xl/workbook.xml').decode('utf-8', errors='strict')
        row_counts = []
        for index in range(1, 4):
            worksheet = f'xl/worksheets/sheet{index}.xml'
            if worksheet not in names:
                raise RuntimeError(f'export missing {worksheet}')
            content = archive.read(worksheet).decode('utf-8', errors='strict')
            row_counts.append(len(re.findall(r'<row(?:\s|>)', content)))
        return len(re.findall(r'<sheet(?:\s|>)', workbook)), row_counts


def main():
    suffix = uuid.uuid4().hex[:12]
    exam_id = ''
    candidate_id = ''
    paper_id = ''
    baseline = ''
    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    token_id = 'phase1-runtime-' + suffix
    redis_key = 'login_tokens:' + token_id
    now_ms = int(time.time() * 1000)
    login_user = {
        'userId': 1, 'token': token_id, 'loginTime': now_ms,
        'expireTime': now_ms + 1800000, 'permissions': ['*:*:*'], 'roles': ['admin'], 'user': None,
    }
    subprocess.run(
        ['redis-cli', '-n', redis_db, 'SET', redis_key, json.dumps(login_user, separators=(',', ':')), 'EX', '1800'],
        check=True, stdout=subprocess.DEVNULL,
    )
    admin_token = sign(secret, token_id)
    try:
        baseline = mysql("""
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_exam WHERE assessment_type='competency'),'|',
  (SELECT COUNT(*) FROM el_competency_result),'|',
  (SELECT COUNT(*) FROM el_competency_group_result),'|',
  (SELECT COUNT(*) FROM el_competency_validity_result))
""")
        if not re.fullmatch(r'\d+\|\d+\|\d+\|\d+', baseline):
            raise RuntimeError(f'invalid competency runtime baseline: {baseline}')
        package_count = mysql("SELECT COUNT(*) FROM el_competency_report_content_package")

        exam, _ = request('/exam/api/exam/exam/save', admin_token, body={
            'title': 'PHASE1-RUNTIME-' + suffix,
            'content': 'temporary staging phase-1 runtime verification',
            'assessmentType': 'competency', 'scoringMode': 'competency_average',
            'joinType': 1, 'openType': 1, 'isOpen': 1, 'answerType': 1,
            'state': 0, 'totalTime': 20, 'repoList': [], 'departIds': [],
        })
        exam_id = exam['id']
        published, _ = request('/exam/api/competency/exams/publish', admin_token, body={'examId': exam_id})
        if published.get('questionCount') != 90 or published.get('dimensionCount') != 10:
            raise RuntimeError(f'publish mismatch: {published}')
        repeated_publish, _ = request('/exam/api/competency/exams/publish', admin_token, body={'examId': exam_id})
        if not repeated_publish.get('alreadyPublished') or repeated_publish.get('questionCount') != 90:
            raise RuntimeError(f'publish idempotency mismatch: {repeated_publish}')

        snapshot_counts = mysql(f"""
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_exam_competency_group WHERE exam_id='{exam_id}'),'|',
  (SELECT COUNT(*) FROM el_exam_competency_dimension WHERE exam_id='{exam_id}'),'|',
  (SELECT COUNT(*) FROM el_exam_competency_dimension WHERE exam_id='{exam_id}' AND group_id IS NOT NULL),'|',
  (SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id='{exam_id}'),'|',
  (SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id='{exam_id}' AND competency_question_type='dimension'),'|',
  (SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id='{exam_id}' AND competency_question_type='validity'))
""")
        if snapshot_counts != '2|10|10|90|80|10':
            raise RuntimeError(f'snapshot counts mismatch: {snapshot_counts}')

        candidate, _ = request('/exam/api/candidate/save', admin_token, body={
            'examId': exam_id, 'name': 'Phase1 Runtime Verify',
            'telephone': '139' + suffix[:8], 'gender': '0',
        })
        candidate_id = candidate['id']
        participant_token = candidate.get('participantToken')
        if not participant_token:
            raise RuntimeError('participant token missing')

        access, _ = request('/exam/api/competency/participant/create-paper', body={
            'examId': exam_id, 'participantId': candidate_id,
            'participantType': 'candidate', 'participantToken': participant_token,
        })
        paper_id = access['paperId']
        paper_token = access['paperToken']
        restored, _ = request('/exam/api/competency/participant/create-paper', body={
            'examId': exam_id, 'participantId': candidate_id,
            'participantType': 'candidate', 'participantToken': participant_token,
        })
        if restored.get('paperId') != paper_id:
            raise RuntimeError('paper restore created another paper')

        detail = paper_request('/exam/api/competency/participant/paper-detail', paper_token, {'paperId': paper_id})
        questions = detail.get('questions') or []
        if detail.get('totalCount') != 90 or len(questions) != 90 or len({row['id'] for row in questions}) != 90:
            raise RuntimeError('paper does not contain 90 unique questions')
        expected_labels = ['完全不符合', '比较不符合', '不确定', '比较符合', '完全符合']
        for question in questions:
            if [item['label'] for item in question.get('options') or []] != expected_labels:
                raise RuntimeError(f'option snapshot mismatch: {question.get("code")}')
            raw_value = 1 if question.get('code', '').startswith('P1-VAL-Q') else 3
            paper_request('/exam/api/competency/participant/fill-answer', paper_token, {
                'paperId': paper_id, 'paperQuestionId': question['id'], 'rawValue': raw_value,
            })

        submitted = paper_request('/exam/api/competency/participant/submit', paper_token, {
            'paperId': paper_id, 'submitType': 'manual',
        })
        if not submitted.get('isComplete'):
            raise RuntimeError(f'submit incomplete: {submitted}')
        repeated_submit = paper_request('/exam/api/competency/participant/submit', paper_token, {
            'paperId': paper_id, 'submitType': 'manual',
        })
        if not repeated_submit.get('alreadySubmitted'):
            raise RuntimeError('submit idempotency failed')

        result_data, _ = request('/exam/api/competency/results/detail', admin_token, body={'paperId': paper_id})
        result = result_data['result']
        dimensions = result_data.get('dimensions') or []
        groups = result_data.get('groups') or []
        validity = result_data.get('validity') or {}
        if len(dimensions) != 10 or len(groups) != 2:
            raise RuntimeError(f'result cardinality mismatch: dimensions={len(dimensions)} groups={len(groups)}')
        if float(result['overallScore']) != 30 or result.get('evaluationLevel') != 'weak' or result.get('isComplete') != 1:
            raise RuntimeError(f'overall result mismatch: {result}')
        if result.get('totalQuestionCount') != 90 or result.get('dimensionQuestionCount') != 80:
            raise RuntimeError(f'question counts mismatch: {result}')
        if any(float(row['dimensionScore']) != 3 or row.get('levelCode') != 'L3' for row in dimensions):
            raise RuntimeError(f'dimension scores mismatch: {dimensions}')
        if any(float(row['groupScore']) != 3 or row.get('levelCode') != 'L3' for row in groups):
            raise RuntimeError(f'group scores mismatch: {groups}')
        if float(validity['validityScore']) != 10 or validity.get('validityStatus') != 'good':
            raise RuntimeError(f'validity mismatch: {validity}')

        db_counts = mysql(f"""
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_competency_dimension_result WHERE paper_id='{paper_id}'),'|',
  (SELECT COUNT(*) FROM el_competency_group_result WHERE paper_id='{paper_id}'),'|',
  (SELECT COUNT(*) FROM el_competency_validity_result WHERE paper_id='{paper_id}'),'|',
  (SELECT COUNT(*) FROM el_competency_result WHERE paper_id='{paper_id}'))
""")
        if db_counts != '10|2|1|1':
            raise RuntimeError(f'result table counts mismatch: {db_counts}')

        good_page, _ = request('/exam/api/competency/results/paging', admin_token, body={
            'examId': exam_id, 'current': 1, 'size': 20, 'validity': 'good', 'sortBy': 'overallScore',
        })
        questionable_page, _ = request('/exam/api/competency/results/paging', admin_token, body={
            'examId': exam_id, 'current': 1, 'size': 20, 'validity': 'questionable', 'sortBy': 'submittedAt',
        })
        if good_page.get('total') != 1 or questionable_page.get('total') != 0:
            raise RuntimeError('validity paging filter mismatch')

        export_payload, export_headers = request(
            '/exam/api/exam/exam/export-raw-data?examId=' + exam_id,
            admin_token, method='GET', body=None,
        )
        if not isinstance(export_payload, dict):
            raise RuntimeError('unexpected JSON export response')
        _ = export_headers
    except (json.JSONDecodeError, UnicodeDecodeError):
        call = urllib.request.Request(
            BASE + '/exam/api/exam/exam/export-raw-data?examId=' + exam_id,
            headers={'Authorization': 'Bearer ' + admin_token}, method='GET',
        )
        with urllib.request.urlopen(call, timeout=120) as response:
            export_bytes = response.read()
            if 'spreadsheetml.sheet' not in response.headers.get('Content-Type', ''):
                raise RuntimeError('export content type mismatch')
        sheet_count, row_counts = xlsx_summary(export_bytes)
        if sheet_count != 3 or row_counts[0] < 2 or row_counts[1] != 91 or row_counts[2] != 91:
            raise RuntimeError(f'export workbook shape mismatch: sheets={sheet_count}, rows={row_counts}')

        status, raw, _ = request(
            '/exam/api/competency/admin/report-data?paperId=' + paper_id,
            admin_token, method='GET', expect_success=False,
        )
        if status >= 400:
            raise RuntimeError(f'phase-1 report gate returned unexpected HTTP {status}')
        report_payload = json.loads(raw.decode())
        report_approved = os.environ.get('EXPECT_REPORT_APPROVED') == '1'
        if report_approved:
            if report_payload.get('code') not in (0, 200):
                raise RuntimeError(f'approved report-data rejected: {report_payload}')
            report_data = report_payload.get('data') or {}
            if report_data.get('schemaVersion') != 'competency-phase1-report-data-v1' or report_data.get('reportKind') != 'frontline_phase1':
                raise RuntimeError(f'approved report DTO mismatch: {report_data}')
            if len(report_data.get('pages') or []) != 10 or len(report_data.get('groups') or []) != 2 or len(report_data.get('dimensions') or []) != 10:
                raise RuntimeError('approved report DTO cardinality mismatch')
            if 'validityScore' in json.dumps(report_data, ensure_ascii=False):
                raise RuntimeError('approved participant report DTO leaks validityScore')

            generated, _ = request('/exam/api/competency/reports/generate', admin_token, body={'paperId': paper_id})
            if generated.get('status') != 'completed' or generated.get('pdfSize', 0) < 1024:
                raise RuntimeError(f'approved report generation mismatch: {generated}')
            download_status, download_bytes, download_headers = request(
                '/exam/api/competency/reports/download?paperId=' + paper_id,
                admin_token, method='GET', expect_success=False,
            )
            if download_status >= 400 or not download_bytes.startswith(b'%PDF') or 'application/pdf' not in download_headers.get('Content-Type', ''):
                raise RuntimeError('approved report download is not a PDF')
            downloaded_pdf = Path('/tmp') / f'phase1-approved-report-{paper_id}.pdf'
            downloaded_pdf.write_bytes(download_bytes)
            try:
                pdf_info = subprocess.run(['pdfinfo', str(downloaded_pdf)], check=True, capture_output=True, text=True).stdout
                if not re.search(r'^Pages:\s+10$', pdf_info, re.MULTILINE):
                    raise RuntimeError(f'approved report PDF is not 10 pages: {pdf_info}')
                pdf_text = subprocess.run(
                    ['pdftotext', '-f', '1', '-l', '1', '-layout', str(downloaded_pdf), '-'],
                    check=True, capture_output=True, text=True,
                ).stdout
                if '时长：' in pdf_text or '分钟' in pdf_text:
                    raise RuntimeError('approved report PDF still exposes the cover duration')
                complete_pdf_text = subprocess.run(
                    ['pdftotext', '-layout', str(downloaded_pdf), '-'],
                    check=True, capture_output=True, text=True,
                ).stdout
                if re.search(r'(?i)\bpage\s*\d+\b', complete_pdf_text):
                    raise RuntimeError('approved report PDF still exposes the English page prefix')
                bbox_xml = subprocess.run(
                    ['pdftotext', '-bbox-layout', str(downloaded_pdf), '-'],
                    check=True, capture_output=True, text=True,
                ).stdout
                bbox_root = ET.fromstring(bbox_xml)
                pages = [node for node in bbox_root.iter() if node.tag.endswith('page')]
                if len(pages) != 10:
                    raise RuntimeError(f'approved report bbox pages={len(pages)}, want 10')
                cover_page = pages[0]
                cover_height = float(cover_page.attrib['height'])
                cover_footer_numbers = [
                    node for node in cover_page.iter()
                    if node.tag.endswith('word')
                    and re.fullmatch(r'\d+', (node.text or '').strip())
                    and float(node.attrib['yMin']) >= cover_height * 0.85
                ]
                if cover_footer_numbers:
                    raise RuntimeError('approved report cover unexpectedly exposes a numeric footer')
                for page_number, page in enumerate(pages[1:], start=1):
                    page_width = float(page.attrib['width'])
                    page_height = float(page.attrib['height'])
                    footer_numbers = [
                        node for node in page.iter()
                        if node.tag.endswith('word')
                        and (node.text or '').strip() == str(page_number)
                        and float(node.attrib['yMin']) >= page_height * 0.85
                    ]
                    if len(footer_numbers) != 1:
                        raise RuntimeError(f'numbered page {page_number} footer count={len(footer_numbers)}, want 1')
                    footer = footer_numbers[0]
                    footer_center = (float(footer.attrib['xMin']) + float(footer.attrib['xMax'])) / 2
                    if abs(footer_center - page_width / 2) > page_width * 0.12:
                        raise RuntimeError(f'numbered page {page_number} footer is not centered')
                report_xml = subprocess.run(
                    ['pdftohtml', '-xml', '-hidden', '-stdout', str(downloaded_pdf)],
                    check=True, capture_output=True, text=True,
                ).stdout
                report_root = ET.fromstring(report_xml)
                plan_labels = [
                    node for node in report_root.iter()
                    if node.tag == 'text' and '计划执行：' in ''.join(node.itertext())
                ]
                if not plan_labels or not any(any(child.tag == 'b' for child in node.iter()) for node in plan_labels):
                    raise RuntimeError('approved report PDF does not render 计划执行 as a bold label')
            finally:
                downloaded_pdf.unlink(missing_ok=True)
            report_counts = mysql(f"""
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_competency_report_content_package),'|',
  (SELECT COUNT(*) FROM el_competency_report WHERE paper_id='{paper_id}' AND status='completed'),'|',
  (SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id='{paper_id}' AND status=1))
""")
            if report_counts != f'{package_count}|1|2':
                raise RuntimeError(f'approved report persistence mismatch: {report_counts}')
        else:
            if report_payload.get('code') in (0, 200):
                raise RuntimeError('phase-1 report renderer unexpectedly enabled')
            blocked_reason = '一期正式报告内容尚未完成双重批准'
            if report_payload.get('msg') != blocked_reason:
                raise RuntimeError(f'phase-1 report-data gate mismatch: {report_payload}')

            for path, method, body in [
                ('/exam/api/competency/reports/generate', 'POST', {'paperId': paper_id}),
                ('/exam/api/competency/reports/download?paperId=' + paper_id, 'GET', None),
            ]:
                gate_status, gate_raw, gate_headers = request(
                    path, admin_token, method=method, body=body, expect_success=False,
                )
                if gate_status >= 400:
                    raise RuntimeError(f'{path}: unexpected HTTP {gate_status}')
                gate_payload = json.loads(gate_raw.decode())
                if gate_payload.get('code') in (0, 200) or gate_payload.get('msg') != blocked_reason:
                    raise RuntimeError(f'{path}: report gate mismatch: {gate_payload}')
                if method == 'GET' and 'application/pdf' in gate_headers.get('Content-Type', ''):
                    raise RuntimeError('blocked report download returned PDF content type')

            report_gate_counts = mysql(f"""
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_competency_report_content_package),'|',
  (SELECT COUNT(*) FROM el_competency_report WHERE paper_id='{paper_id}'),'|',
  (SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id='{paper_id}'))
""")
            if report_gate_counts != f'{package_count}|0|0':
                raise RuntimeError(f'blocked report actions wrote data: {report_gate_counts}')

        print('STAGING_PHASE1_RUNTIME_PASS')
        print('publish=groups:2|dimensions:10|questions:90|types:80/10|idempotent:true')
        print('paper=questions:90|options:confirmed|restore_same:true')
        print('results=dimension:10|group:2|validity:1|overall:1')
        print('scores=dimensions:3/L3|groups:3/L3|overall:30/weak|validity:10/good')
        print('paging=good:1|questionable:0|ranking_default:complete_good')
        print('export=groups:true|validity:true|question_type:true')
        if report_approved:
            print('report=approved_dto:10-pages/2-groups/10-dimensions|generate:completed|download:10-page-pdf|cover_duration:hidden|page_prefix:absent|page_numbers:centered|plan_execution:bold|audits:2')
        else:
            print(f'report=report_data/generate/download:gated|packages:{package_count}|instances:0|audits:0')
    finally:
        if exam_id:
            request('/exam/api/exam/exam/delete', admin_token, body={'ids': [exam_id]})
        subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)
        remaining = mysql(f"""
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_exam WHERE id='{exam_id}'),'|',
  (SELECT COUNT(*) FROM el_paper WHERE id='{paper_id}'),'|',
  (SELECT COUNT(*) FROM el_candidate WHERE id='{candidate_id}'),'|',
  (SELECT COUNT(*) FROM el_competency_result WHERE paper_id='{paper_id}'))
""")
        print('cleanup_remaining=' + remaining)
        current = mysql("""
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_exam WHERE assessment_type='competency'),'|',
  (SELECT COUNT(*) FROM el_competency_result),'|',
  (SELECT COUNT(*) FROM el_competency_group_result),'|',
  (SELECT COUNT(*) FROM el_competency_validity_result))
""")
        print('baseline_before=' + baseline)
        print('baseline_after=' + current)
        if current != baseline:
            raise RuntimeError(f'competency runtime baseline drift: before={baseline} after={current}')


if __name__ == '__main__':
    main()
