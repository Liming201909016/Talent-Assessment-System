#!/usr/bin/env python3
import base64
import hashlib
import hmac
import json
import re
import subprocess
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'
OUTPUT = Path('/tmp/00401-competency-report.pdf')


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


def api(path, body, token):
    request = urllib.request.Request(
        BASE + path,
        data=json.dumps(body, ensure_ascii=False).encode(),
        headers={'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token},
        method='POST',
    )
    try:
        with urllib.request.urlopen(request, timeout=120) as response:
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
        ['sudo', '-n', 'mysql', 'element', '-Nse', sql],
        check=True, capture_output=True, text=True,
    ).stdout.strip()


def main():
    paper_id = sys.argv[1] if len(sys.argv) > 1 else ''
    if not re.fullmatch(r'[0-9a-fA-F-]{36}', paper_id):
        raise RuntimeError('valid paper id argument required')
    row = mysql(
        "SELECT CONCAT(e.title,'|',r.is_complete,'|',r.answered_question_count,'|',r.total_question_count) "
        "FROM el_competency_result r INNER JOIN el_exam e ON e.id=r.exam_id "
        f"WHERE r.paper_id='{paper_id}'"
    )
    if not row:
        raise RuntimeError('competency result not found')
    title, is_complete, answered, total = row.split('|')
    if '00401' not in title or is_complete != '1' or answered != total:
        raise RuntimeError(f'paper is not a complete 00401 result: {title}|{is_complete}|{answered}|{total}')

    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    token_id = '00401-report-' + str(int(time.time()))
    redis_key = 'login_tokens:' + token_id
    now_ms = int(time.time() * 1000)
    login_user = {
        'userId': 1, 'token': token_id, 'loginTime': now_ms,
        'expireTime': now_ms + 900000, 'permissions': ['*:*:*'], 'roles': ['admin'], 'user': None,
    }
    subprocess.run(
        ['redis-cli', '-n', redis_db, 'SET', redis_key, json.dumps(login_user, separators=(',', ':')), 'EX', '900'],
        check=True, stdout=subprocess.DEVNULL,
    )
    try:
        token = sign(secret, token_id)
        report = api('/exam/api/competency/reports/generate', {'paperId': paper_id, 'force': True}, token)
        if report.get('status') != 'completed' or report.get('contentVersion') != 'temp-v1':
            raise RuntimeError('report generation did not complete')
        pdf = download('/exam/api/competency/reports/download?paperId=' + paper_id, token)
        if not pdf.startswith(b'%PDF') or len(pdf) < 1024:
            raise RuntimeError(f'invalid PDF: {len(pdf)} bytes')
        digest = hashlib.sha256(pdf).hexdigest()
        if digest != report.get('pdfSha256'):
            raise RuntimeError('downloaded PDF hash differs from report instance')
        OUTPUT.write_bytes(pdf)
        extracted = subprocess.run(['pdftotext', str(OUTPUT), '-'], check=True, capture_output=True, text=True).stdout
        required = [title, '临时测试报告', '报告阅读说明', '总体评价', '测评结果分析', '核心含义', '发展提示', '不可作为人才决策依据']
        missing = [item for item in required if item not in extracted]
        if missing:
            raise RuntimeError('PDF text missing: ' + ','.join(missing))
        pages = subprocess.run(['pdfinfo', str(OUTPUT)], check=True, capture_output=True, text=True).stdout
        page_match = re.search(r'^Pages:\s+(\d+)', pages, re.MULTILINE)
        page_count = int(page_match.group(1)) if page_match else 0
        dimension_count = int(mysql(f"SELECT COUNT(*) FROM el_competency_dimension_result WHERE paper_id='{paper_id}'"))
        expected_pages = 4 + dimension_count
        if page_count != expected_pages:
            raise RuntimeError(f'PDF page count mismatch: {page_count}, want {expected_pages}')
        print('REPORT_00401_GENERATE_OK')
        print('paper_id=' + paper_id)
        print('exam_title=' + title)
        print('answered=' + answered + '/' + total)
        print('bytes=' + str(len(pdf)))
        print('sha256=' + digest)
        print('pages=' + str(page_count))
        print('remote_file=' + str(OUTPUT))
    finally:
        subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)


if __name__ == '__main__':
    main()
