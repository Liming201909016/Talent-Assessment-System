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
EXAM_ID = '1785060929295657251'
QUESTION_HEADERS = ['维度序号', '维度名称', '题目编号', '维度内题号', '题目内容', '考察点', '计分方向', '启用状态', '备注']
EXAMPLE_CODES = ['D01-EXAMPLE-F', 'D01-EXAMPLE-R']
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


def mysql(sql):
    return subprocess.run(
        ['sudo', '-n', 'mysql', 'element', '-Nse', sql],
        check=True, capture_output=True, text=True,
    ).stdout.strip()


def request(path, token, method='GET', body=None, content_type='application/json'):
    headers = {'Authorization': 'Bearer ' + token}
    data = body
    if body is not None:
        headers['Content-Type'] = content_type
    call = urllib.request.Request(BASE + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(call, timeout=120) as response:
            return response.read(), dict(response.headers)
    except urllib.error.HTTPError as error:
        raise RuntimeError(f'{path} HTTP {error.code}: {error.read().decode(errors="replace")}')


def multipart(file_bytes, expected_hash=None):
    boundary = '----00401-' + uuid.uuid4().hex
    parts = []
    parts.append(f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="00401-template.xlsx"\r\nContent-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet\r\n\r\n'.encode())
    parts.append(file_bytes)
    parts.append(b'\r\n')
    if expected_hash:
        parts.append(f'--{boundary}\r\nContent-Disposition: form-data; name="expectedHash"\r\n\r\n{expected_hash}\r\n'.encode())
    parts.append(f'--{boundary}--\r\n'.encode())
    return b''.join(parts), 'multipart/form-data; boundary=' + boundary


def api_upload(path, token, file_bytes, expected_hash=None):
    body, content_type = multipart(file_bytes, expected_hash)
    raw, _ = request(path, token, method='POST', body=body, content_type=content_type)
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


def assert_xlsx(headers, data):
    if 'spreadsheetml.sheet' not in headers.get('Content-Type', '') or not data.startswith(b'PK'):
        raise RuntimeError(f'invalid xlsx response: type={headers.get("Content-Type")} bytes={len(data)}')
    if 'filename*=UTF-8' not in headers.get('Content-Disposition', ''):
        raise RuntimeError('missing RFC5987 filename')


def main():
    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    token_id = '00401-io-' + uuid.uuid4().hex[:12]
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
    imported_ids = []
    try:
        if mysql("SELECT COUNT(*) FROM el_qu WHERE question_code IN ('D01-EXAMPLE-F','D01-EXAMPLE-R')") != '0':
            raise RuntimeError('example question codes already exist; refusing to delete unknown data')
        baseline = int(mysql('SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NOT NULL'))

        template, template_headers = request('/exam/api/competency/questions/import-template', token)
        assert_xlsx(template_headers, template)
        template_book = workbook_rows(template)
        template_rows = template_book.get('胜任力题目') or []
        if len(template_rows) != 4 or template_rows[0] != QUESTION_HEADERS:
            raise RuntimeError(f'template rows/header mismatch: rows={len(template_rows)} header={template_rows[:1]}')
        if [template_rows[2][2], template_rows[3][2]] != EXAMPLE_CODES or [template_rows[2][6], template_rows[3][6]] != ['正向', '反向']:
            raise RuntimeError('template does not contain the required forward/reverse examples')

        preview = api_upload('/exam/api/competency/questions/import-preview', token, template)
        if preview.get('successCount') != 2 or preview.get('errorCount') != 0 or len(preview.get('sha256', '')) != 64:
            raise RuntimeError(f'template preview mismatch: {preview}')
        imported = api_upload('/exam/api/competency/questions/import', token, template, preview['sha256'])
        if imported.get('importedCount') != 2:
            raise RuntimeError(f'template import mismatch: {imported}')
        imported_ids = mysql("SELECT id FROM el_qu WHERE question_code IN ('D01-EXAMPLE-F','D01-EXAMPLE-R') ORDER BY question_code").splitlines()
        if len(imported_ids) != 2 or int(mysql('SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NOT NULL')) != baseline + 2:
            raise RuntimeError('imported database rows mismatch')

        questions, question_headers = request('/exam/api/competency/questions/export', token)
        assert_xlsx(question_headers, questions)
        question_rows = workbook_rows(questions).get('胜任力题目') or []
        if len(question_rows) != baseline + 3 or question_rows[0] != QUESTION_HEADERS:
            raise RuntimeError(f'question export row count/header mismatch: {len(question_rows)}')
        exported_examples = {row[2]: row for row in question_rows[1:] if len(row) >= 9 and row[2] in EXAMPLE_CODES}
        if set(exported_examples) != set(EXAMPLE_CODES):
            raise RuntimeError('question export is missing imported examples')
        if exported_examples['D01-EXAMPLE-F'][6:8] != ['正向', '启用'] or exported_examples['D01-EXAMPLE-R'][6:8] != ['反向', '启用']:
            raise RuntimeError('question export direction/status mismatch')

        result_exports = []
        for endpoint in ('export-raw-data', 'export-raw-answers'):
            data, headers = request(f'/exam/api/exam/exam/{endpoint}?examId={EXAM_ID}', token)
            assert_xlsx(headers, data)
            result_exports.append(workbook_rows(data))
        if result_exports[0] != result_exports[1]:
            raise RuntimeError('result export endpoints returned different content')
        result_book = result_exports[0]
        if list(result_book) != ['结果汇总', '逐题明细', '题目字典']:
            raise RuntimeError(f'result export sheets mismatch: {list(result_book)}')
        if len(result_book['结果汇总']) != 2 or len(result_book['逐题明细']) != 41 or len(result_book['题目字典']) != 41:
            raise RuntimeError('result export row counts mismatch')
        summary = result_book['结果汇总']
        if summary[1][12] != '5.000000' or summary[1][13] != '1.000000' or summary[1][8] != '40/40':
            raise RuntimeError(f'result export persisted score mismatch: {summary[1][8:14]}')

        print('STAGING_00401_IMPORT_EXPORT_PASS')
        print(f'template_rows={len(template_rows)}')
        print('template_examples=2|forward=1|reverse=1')
        print('preview_success=2|preview_errors=0')
        print('imported_count=2')
        print(f'question_export_rows={len(question_rows)-1}')
        print('result_export_sheets=3|summary_rows=1|detail_rows=40|dictionary_rows=40')
        print('result_export_endpoints_same=true')
    finally:
        if imported_ids:
            quoted = ','.join("'" + value.replace("'", "''") + "'" for value in imported_ids)
            mysql(f'DELETE FROM el_qu WHERE id IN ({quoted})')
        remaining = mysql("SELECT COUNT(*) FROM el_qu WHERE question_code IN ('D01-EXAMPLE-F','D01-EXAMPLE-R')")
        subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)
        print('cleanup_example_questions=' + remaining)


if __name__ == '__main__':
    main()
