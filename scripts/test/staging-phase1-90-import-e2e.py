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
from xml.etree import ElementTree as ET

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'
IMPORT_FILE = Path('/tmp/competency-phase1-import-20260810.xlsx')
QUESTION_HEADERS = [
    '维度序号', '维度名称', '题目类型', '题目编号', '维度内题号',
    '题目内容', '考察点', '计分方向', '启用状态', '备注',
]
NS = {'m': 'http://schemas.openxmlformats.org/spreadsheetml/2006/main', 'r': 'http://schemas.openxmlformats.org/officeDocument/2006/relationships'}
REL_NS = {'p': 'http://schemas.openxmlformats.org/package/2006/relationships'}


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
    signature = b64url(hmac.new(secret.encode(), raw.encode(), hashlib.sha512).digest())
    return f'{raw}.{signature}'


def mysql(sql):
    return subprocess.run(
        ['sudo', '-n', 'mysql', 'element', '-Nse', sql],
        check=True, capture_output=True, text=True,
    ).stdout.strip()


def request(path, token, method='GET', body=None, content_type='application/json'):
    headers = {'Authorization': 'Bearer ' + token}
    if body is not None:
        headers['Content-Type'] = content_type
    call = urllib.request.Request(BASE + path, data=body, headers=headers, method=method)
    try:
        with urllib.request.urlopen(call, timeout=120) as response:
            return response.read(), dict(response.headers)
    except urllib.error.HTTPError as error:
        raise RuntimeError(f'{path} HTTP {error.code}: {error.read().decode(errors="replace")}')


def multipart(file_bytes, expected_hash=None):
    boundary = '----phase1-90-' + uuid.uuid4().hex
    parts = [
        f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="competency-phase1-import-20260810.xlsx"\r\nContent-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet\r\n\r\n'.encode(),
        file_bytes,
        b'\r\n',
    ]
    if expected_hash:
        parts.append(f'--{boundary}\r\nContent-Disposition: form-data; name="expectedHash"\r\n\r\n{expected_hash}\r\n'.encode())
    parts.append(f'--{boundary}--\r\n'.encode())
    return b''.join(parts), 'multipart/form-data; boundary=' + boundary


def api_upload(path, token, file_bytes, expected_hash=None, expect_success=True):
    body, content_type = multipart(file_bytes, expected_hash)
    raw, _ = request(path, token, method='POST', body=body, content_type=content_type)
    payload = json.loads(raw.decode())
    succeeded = payload.get('code') in (0, 200)
    if expect_success and not succeeded:
        raise RuntimeError(f'{path}: {payload.get("msg")}')
    if not expect_success and succeeded:
        raise RuntimeError(f'{path}: expected rejection, got {payload}')
    return payload.get('data') or {}, payload


def api_json(path, token, body):
    raw, _ = request(
        path, token, method='POST',
        body=json.dumps(body, ensure_ascii=False, separators=(',', ':')).encode(),
    )
    payload = json.loads(raw.decode())
    if payload.get('code') not in (0, 200):
        raise RuntimeError(f'{path}: {payload.get("msg")}')
    return payload.get('data') or {}


def cell_column(reference):
    letters = re.match(r'[A-Z]+', reference).group(0)
    number = 0
    for letter in letters:
        number = number * 26 + ord(letter) - 64
    return number


def workbook_rows(data):
    with zipfile.ZipFile(BytesIO(data)) as archive:
        shared = []
        if 'xl/sharedStrings.xml' in archive.namelist():
            root = ET.fromstring(archive.read('xl/sharedStrings.xml'))
            for item in root.findall('m:si', NS):
                shared.append(''.join(node.text or '' for node in item.findall('.//m:t', NS)))
        workbook = ET.fromstring(archive.read('xl/workbook.xml'))
        relationships = ET.fromstring(archive.read('xl/_rels/workbook.xml.rels'))
        targets = {item.attrib['Id']: item.attrib['Target'] for item in relationships.findall('p:Relationship', REL_NS)}
        output = {}
        for sheet in workbook.findall('m:sheets/m:sheet', NS):
            name = sheet.attrib['name']
            target = targets[sheet.attrib['{%s}id' % NS['r']]].lstrip('/')
            if not target.startswith('xl/'):
                target = 'xl/' + target
            xml = ET.fromstring(archive.read(target))
            rows = []
            for row in xml.findall('m:sheetData/m:row', NS):
                values = []
                for cell in row.findall('m:c', NS):
                    column = cell_column(cell.attrib['r'])
                    while len(values) < column:
                        values.append('')
                    kind = cell.attrib.get('t')
                    value_node = cell.find('m:v', NS)
                    value = '' if value_node is None else value_node.text or ''
                    if kind == 's' and value:
                        value = shared[int(value)]
                    elif kind == 'inlineStr':
                        value = ''.join(node.text or '' for node in cell.findall('.//m:t', NS))
                    values[column - 1] = value
                rows.append(values)
            output[name] = rows
        return output


def normalize_rows(rows):
    return {row[3]: row[:10] for row in rows[1:] if len(row) >= 10}


def assert_xlsx(headers, data):
    if 'spreadsheetml.sheet' not in headers.get('Content-Type', '') or not data.startswith(b'PK'):
        raise RuntimeError(f'invalid xlsx response: type={headers.get("Content-Type")} bytes={len(data)}')
    if 'filename*=UTF-8' not in headers.get('Content-Disposition', ''):
        raise RuntimeError('missing RFC5987 filename')


def main():
    if not IMPORT_FILE.exists():
        raise RuntimeError(f'import file missing: {IMPORT_FILE}')
    file_bytes = IMPORT_FILE.read_bytes()
    file_sha256 = hashlib.sha256(file_bytes).hexdigest()
    import_rows = workbook_rows(file_bytes).get('胜任力题目') or []
    if len(import_rows) != 91 or import_rows[0] != QUESTION_HEADERS:
        raise RuntimeError(f'import workbook shape mismatch: rows={len(import_rows)} header={import_rows[:1]}')

    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    token_id = 'phase1-90-' + uuid.uuid4().hex[:12]
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
    token = sign(secret, token_id)
    try:
        baseline = int(mysql("SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NOT NULL OR competency_question_type IS NOT NULL"))
        if baseline not in (0, 90):
            raise RuntimeError(f'refusing mixed/unknown competency source baseline: {baseline}')

        if baseline == 0:
            preview, _ = api_upload('/exam/api/competency/questions/import-preview', token, file_bytes)
            if preview.get('successCount') != 90 or preview.get('errorCount') != 0 or preview.get('sha256') != file_sha256:
                raise RuntimeError(f'preview mismatch: {preview}')
            imported, _ = api_upload('/exam/api/competency/questions/import', token, file_bytes, preview['sha256'])
            if imported.get('importedCount') != 90 or imported.get('sha256') != file_sha256:
                raise RuntimeError(f'import mismatch: {imported}')
        else:
            preview = {'successCount': 0, 'errorCount': 90, 'sha256': file_sha256}
            imported = {'importedCount': 0, 'sha256': file_sha256}

        counts = mysql("""
SELECT COUNT(*),COUNT(DISTINCT question_code),
       SUM(competency_question_type='dimension'),SUM(competency_question_type='validity'),
       SUM(competency_question_type='dimension' AND scoring_direction='forward'),
       SUM(competency_question_type='dimension' AND scoring_direction='reverse'),
       SUM(competency_question_type='validity' AND scoring_direction='forward'),
       SUM(question_status=0)
FROM el_qu
WHERE dimension_id IS NOT NULL OR competency_question_type IS NOT NULL
""").split('\t')
        if counts != ['90', '90', '80', '10', '62', '18', '10', '90']:
            raise RuntimeError(f'database counts mismatch: {counts}')

        distribution = mysql("""
SELECT COUNT(*),MIN(dimension_count),MAX(dimension_count),SUM(dimension_count),
       MIN(validity_count),MAX(validity_count),SUM(validity_count)
FROM (
  SELECT d.id,
         SUM(q.competency_question_type='dimension') AS dimension_count,
         SUM(q.competency_question_type='validity') AS validity_count
  FROM el_competency_dimension d
  LEFT JOIN el_qu q ON q.dimension_id=d.id
  GROUP BY d.id
) x
""").split('\t')
        if distribution != ['10', '8', '8', '80', '1', '1', '10']:
            raise RuntimeError(f'dimension/type distribution mismatch: {distribution}')

        relation_counts = mysql("""
SELECT
  (SELECT COUNT(*) FROM el_qu_answer qa INNER JOIN el_qu q ON q.id=qa.qu_id WHERE q.competency_question_type IN ('dimension','validity')),
  (SELECT COUNT(*) FROM el_qu_repo qr INNER JOIN el_qu q ON q.id=qr.qu_id WHERE q.competency_question_type IN ('dimension','validity'))
""").split('\t')
        if relation_counts != ['0', '0']:
            raise RuntimeError(f'unexpected answer/repository rows: {relation_counts}')

        dimensions = api_json('/exam/api/competency/dimensions/list', token, {})
        if len(dimensions) != 10 or sum(item.get('questionCount', 0) for item in dimensions) != 80:
            raise RuntimeError('dimension API must count only 80 enabled dimension questions')

        paging = api_json('/exam/api/competency/questions/paging', token, {'current': 1, 'size': 200, 'params': {}})
        records = paging.get('records') or []
        if paging.get('total') != 90 or len(records) != 90:
            raise RuntimeError(f'question paging mismatch: total={paging.get("total")} rows={len(records)}')
        paging_types = {kind: sum(row.get('competencyQuestionType') == kind for row in records) for kind in ('dimension', 'validity')}
        if paging_types != {'dimension': 80, 'validity': 10}:
            raise RuntimeError(f'question paging types mismatch: {paging_types}')

        exported, export_headers = request('/exam/api/competency/questions/export', token)
        assert_xlsx(export_headers, exported)
        export_rows = workbook_rows(exported).get('胜任力题目') or []
        if export_rows[0] != QUESTION_HEADERS or len(export_rows) != 91:
            raise RuntimeError(f'export shape mismatch: rows={len(export_rows)}')
        if normalize_rows(export_rows) != normalize_rows(import_rows):
            raise RuntimeError('exported ten-column content differs from imported workbook')

        rejected_preview, _ = api_upload('/exam/api/competency/questions/import-preview', token, file_bytes)
        if rejected_preview.get('successCount') != 0 or rejected_preview.get('errorCount') != 90:
            raise RuntimeError(f'repeated preview must expose 90 existing-row errors: {rejected_preview}')
        _, repeated_import = api_upload(
            '/exam/api/competency/questions/import', token, file_bytes, file_sha256,
            expect_success=False,
        )
        if '导入数据存在错误' not in repeated_import.get('msg', ''):
            raise RuntimeError(f'repeated import rejection mismatch: {repeated_import}')
        if mysql("SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NOT NULL OR competency_question_type IS NOT NULL") != '90':
            raise RuntimeError('repeated import changed database row count')

        print('STAGING_PHASE1_90_IMPORT_PASS')
        print(f'import_sha256={file_sha256}')
        print(f'baseline={baseline}|preview_success={preview["successCount"]}|preview_errors={preview["errorCount"]}|imported={imported["importedCount"]}')
        print('database=90|unique=90|dimension=80|validity=10|dimension_forward=62|dimension_reverse=18|validity_forward=10|enabled=90')
        print('distribution=dimensions:10|dimension_each:8|validity_each:1')
        print('relations=answers:0|repositories:0')
        print('api=dimensions:10|dimension_question_count:80|paging:90|paging_types:80/10')
        print('export=rows:90|headers:10|content_match=true')
        print('repeat=preview_errors:90|import_rejected:true|rows_unchanged:90')
    finally:
        subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)
        print('cleanup_redis_session=' + subprocess.run(
            ['redis-cli', '-n', redis_db, 'EXISTS', redis_key], check=True, capture_output=True, text=True,
        ).stdout.strip())


if __name__ == '__main__':
    main()
