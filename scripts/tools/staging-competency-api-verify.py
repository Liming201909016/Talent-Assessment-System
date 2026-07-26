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
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'
TITLE = 'AUTO-COMPETENCY-STAGING-VERIFY'


def b64url(value):
    return base64.urlsafe_b64encode(value).rstrip(b'=').decode('ascii')


def config_text():
    env_line = subprocess.run(
        ['systemctl', 'show', 'talent-assessment', '--property=Environment', '--value'],
        check=True, capture_output=True, text=True,
    ).stdout.strip()
    match = re.search(r'(?:^|\s)APP_ENV=([^\s]+)', env_line)
    app_env = match.group(1) if match else 'local'
    files = [APP_DIR / 'configs' / 'application.yml', APP_DIR / 'configs' / f'application-{app_env}.yml']
    return '\n'.join(path.read_text(encoding='utf-8') for path in files if path.exists())


def scalar(text, section, key, default=''):
    value = default
    for section_match in re.finditer(rf'(?ms)^{re.escape(section)}:\s*\n(.*?)(?=^[A-Za-z][\w-]*:\s*$|\Z)', text):
        key_match = re.search(rf'(?m)^\s+{re.escape(key)}:\s*(.+?)\s*$', section_match.group(1))
        if key_match:
            value = key_match.group(1).strip().strip('"\'')
    return value


def sign_token(secret, token_id):
    header = b64url(json.dumps({'alg': 'HS512', 'typ': 'JWT'}, separators=(',', ':')).encode())
    payload = b64url(json.dumps({'login_user_key': token_id}, separators=(',', ':')).encode())
    signing_input = f'{header}.{payload}'
    signature = b64url(hmac.new(secret.encode(), signing_input.encode(), hashlib.sha512).digest())
    return f'{signing_input}.{signature}'


def api(path, token, body):
    request = urllib.request.Request(
        BASE + path,
        data=json.dumps(body, ensure_ascii=False).encode('utf-8'),
        headers={'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token},
        method='POST',
    )
    try:
        with urllib.request.urlopen(request, timeout=20) as response:
            payload = json.loads(response.read().decode('utf-8'))
    except urllib.error.HTTPError as error:
        raise RuntimeError(f'{path} HTTP {error.code}: {error.read().decode("utf-8", errors="replace")}')
    if payload.get('code') != 0:
        raise RuntimeError(f'{path} API code={payload.get("code")}: {payload.get("msg")}')
    return payload.get('data')


def mysql_scalar(sql):
    result = subprocess.run(['sudo', '-n', 'mysql', 'element', '-Nse', sql], check=True, capture_output=True, text=True)
    return result.stdout.strip()


def main():
    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    if not secret:
        raise RuntimeError('jwt secret missing')

    token_id = 'competency-verify-' + uuid.uuid4().hex
    now_ms = int(time.time() * 1000)
    login_user = {
        'userId': 1,
        'token': token_id,
        'loginTime': now_ms,
        'expireTime': now_ms + 15 * 60 * 1000,
        'permissions': ['*:*:*'],
        'roles': ['admin'],
        'user': None,
    }
    redis_key = 'login_tokens:' + token_id
    subprocess.run(['redis-cli', '-n', redis_db, 'SET', redis_key, json.dumps(login_user, separators=(',', ':')), 'EX', '900'], check=True, stdout=subprocess.DEVNULL)
    token = sign_token(secret, token_id)
    exam_id = None

    try:
        dimensions = api('/exam/api/competency/dimensions/list', token, {})
        if len(dimensions) != 48:
            raise RuntimeError(f'dimension count={len(dimensions)}, want 48')
        selected = [dimensions[0]['id'], dimensions[1]['id']]

        create_body = {
            'title': TITLE,
            'content': 'staging automated verification',
            'assessmentType': 'competency',
            'scoringMode': 'competency_average',
            'competencyReportAudience': 'frontline_employee',
            'dimensionIds': selected,
            'joinType': 1,
            'openType': 1,
            'isOpen': 1,
            'answerType': 1,
            'state': 0,
            'totalScore': 10,
            'totalTime': 30,
            'qualifyScore': 0,
            'showPdf': False,
            'timeLimit': False,
            'requiredFields': 'name,gender,age,telephone',
            'repoList': [],
            'departIds': [],
        }
        created = api('/exam/api/exam/exam/save', token, create_body)
        exam_id = created['id']
        detail = api('/exam/api/exam/exam/detail', token, {'id': exam_id})
        if detail.get('assessmentType') != 'competency':
            raise RuntimeError('created assessment type mismatch')
        if detail.get('competencyReportAudience') != 'frontline_employee':
            raise RuntimeError('frontline report audience mismatch')
        if detail.get('dimensionIds') != selected:
            raise RuntimeError(f'dimension round-trip mismatch: {detail.get("dimensionIds")}')
        if detail.get('publishStatus') != 0:
            raise RuntimeError('competency draft publish status must be 0')

        create_body.update({
            'id': exam_id,
            'competencyReportAudience': 'leader',
            'dimensionIds': [selected[1]],
        })
        api('/exam/api/exam/exam/save', token, create_body)
        edited = api('/exam/api/exam/exam/detail', token, {'id': exam_id})
        if edited.get('competencyReportAudience') != 'leader' or edited.get('dimensionIds') != [selected[1]]:
            raise RuntimeError('leader/dimension edit round-trip mismatch')

        row = mysql_scalar(
            "SELECT CONCAT(assessment_type,'|',scoring_mode,'|',competency_report_audience,'|',publish_status) "
            f"FROM el_exam WHERE id='{exam_id}';"
        )
        assoc_count = mysql_scalar(f"SELECT COUNT(*) FROM el_exam_competency_dimension WHERE exam_id='{exam_id}';")
        if row != 'competency|competency_average|leader|0' or assoc_count != '1':
            raise RuntimeError(f'DB verification mismatch row={row} associations={assoc_count}')

        api('/exam/api/exam/exam/delete', token, {'ids': [exam_id]})
        remaining = mysql_scalar(
            f"SELECT (SELECT COUNT(*) FROM el_exam WHERE id='{exam_id}') + "
            f"(SELECT COUNT(*) FROM el_exam_competency_dimension WHERE exam_id='{exam_id}');"
        )
        if remaining != '0':
            raise RuntimeError(f'cleanup failed, remaining={remaining}')
        exam_id = None

        print('API_VERIFY_OK')
        print('dimension_count=48')
        print('create_audience=frontline_employee')
        print('edit_audience=leader')
        print('create_dimensions=2')
        print('edit_dimensions=1')
        print('cleanup_remaining=0')
    finally:
        if exam_id:
            subprocess.run(['sudo', '-n', 'mysql', 'element', '-e', f"DELETE FROM el_exam_competency_dimension WHERE exam_id='{exam_id}'; DELETE FROM el_exam WHERE id='{exam_id}';"], check=False)
        subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)


if __name__ == '__main__':
    main()
