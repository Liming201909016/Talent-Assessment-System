#!/usr/bin/env python3
import base64
import hashlib
import hmac
import json
import mimetypes
import re
import subprocess
import sys
import urllib.error
import urllib.request
import uuid
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'
INPUT = Path('/tmp/competency-test-questions.xlsx')
EXPECTED_COUNT = 384


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


def multipart(file_path, fields=None):
    boundary = '----competency-' + uuid.uuid4().hex
    chunks = []
    for key, value in (fields or {}).items():
        chunks.extend([
            f'--{boundary}\r\n'.encode(),
            f'Content-Disposition: form-data; name="{key}"\r\n\r\n'.encode(),
            str(value).encode(), b'\r\n',
        ])
    content_type = mimetypes.guess_type(file_path.name)[0] or 'application/octet-stream'
    chunks.extend([
        f'--{boundary}\r\n'.encode(),
        f'Content-Disposition: form-data; name="file"; filename="{file_path.name}"\r\n'.encode(),
        f'Content-Type: {content_type}\r\n\r\n'.encode(),
        file_path.read_bytes(), b'\r\n', f'--{boundary}--\r\n'.encode(),
    ])
    return b''.join(chunks), f'multipart/form-data; boundary={boundary}'


def api(path, token, file_path, fields=None):
    body, content_type = multipart(file_path, fields)
    request = urllib.request.Request(
        BASE + path,
        data=body,
        headers={'Content-Type': content_type, 'Authorization': 'Bearer ' + token},
        method='POST',
    )
    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            payload = json.loads(response.read().decode())
    except urllib.error.HTTPError as error:
        raise RuntimeError(f'{path} HTTP {error.code}: {error.read().decode(errors="replace")}')
    if payload.get('code') not in (0, 200):
        raise RuntimeError(f'{path}: {payload.get("msg")}')
    return payload.get('data')


def mysql(sql):
    return subprocess.run(
        ['sudo', '-n', 'mysql', 'element', '-Nse', sql],
        check=True, capture_output=True, text=True,
    ).stdout.strip()


def main():
    if not INPUT.exists() or INPUT.stat().st_size == 0:
        raise RuntimeError('import file missing or empty')
    current = int(mysql("SELECT COUNT(*) FROM el_qu WHERE question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$';"))
    if current == EXPECTED_COUNT:
        distribution = mysql("SELECT COUNT(*),MIN(c),MAX(c),SUM(c) FROM (SELECT dimension_id,COUNT(*) c FROM el_qu WHERE question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$' GROUP BY dimension_id) x;")
        print('COMPETENCY_QUESTIONS_ALREADY_IMPORTED')
        print(f'questions={current}')
        print(f'distribution={distribution}')
        return
    if current != 0:
        raise RuntimeError(f'partial competency question data exists: {current}; refusing import')

    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    if not secret:
        raise RuntimeError('JWT secret unavailable')
    token_id = 'competency-import-' + uuid.uuid4().hex
    redis_key = 'login_tokens:' + token_id
    now_ms = int(__import__('time').time() * 1000)
    login_user = {
        'userId': 1, 'token': token_id, 'loginTime': now_ms,
        'expireTime': now_ms + 1800000, 'permissions': ['*:*:*'],
        'roles': ['admin'], 'user': None,
    }
    subprocess.run(
        ['redis-cli', '-n', redis_db, 'SET', redis_key, json.dumps(login_user, separators=(',', ':')), 'EX', '1800'],
        check=True, stdout=subprocess.DEVNULL,
    )
    token = sign(secret, token_id)
    try:
        preview = api('/exam/api/competency/questions/import-preview', token, INPUT)
        if preview.get('successCount') != EXPECTED_COUNT or preview.get('errorCount') != 0:
            sample = (preview.get('errorRows') or [])[:5]
            raise RuntimeError(f'preview failed: success={preview.get("successCount")} errors={preview.get("errorCount")} sample={sample}')
        digest = preview.get('sha256')
        imported = api('/exam/api/competency/questions/import', token, INPUT, {'expectedHash': digest})
        if imported.get('importedCount') != EXPECTED_COUNT or imported.get('sha256') != digest:
            raise RuntimeError(f'import response mismatch: {imported}')

        checks = mysql("""
SELECT CONCAT(
  COUNT(*),'|',COUNT(DISTINCT question_code),'|',COUNT(DISTINCT dimension_id),'|',
  SUM(scoring_direction='forward'),'|',SUM(scoring_direction='reverse'),'|',
  SUM(question_status=0),'|',SUM(remark='AI测试题-未信效度验证')
) FROM el_qu WHERE question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$';
SELECT CONCAT(MIN(c),'|',MAX(c),'|',SUM(c)) FROM (
  SELECT dimension_id,COUNT(*) c FROM el_qu
  WHERE question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$'
  GROUP BY dimension_id
) x;
""").splitlines()
        if checks != ['384|384|48|288|96|384|384', '8|8|384']:
            raise RuntimeError(f'database verification mismatch: {checks}')
        print('COMPETENCY_QUESTIONS_IMPORT_OK')
        print('preview_success=384 preview_errors=0')
        print('imported=384 unique_codes=384 dimensions=48')
        print('directions=288_forward_96_reverse')
        print('per_dimension=8 total=384')
        print(f'sha256={digest}')
    finally:
        subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        print(f'COMPETENCY_QUESTIONS_IMPORT_FAILED: {error}', file=sys.stderr)
        sys.exit(1)
