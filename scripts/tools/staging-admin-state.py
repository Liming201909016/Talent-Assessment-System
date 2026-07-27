#!/usr/bin/env python3
import base64
import hashlib
import hmac
import json
import re
import subprocess
import sys
import time
import uuid
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')


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


def main():
    if len(sys.argv) != 4:
        raise RuntimeError('usage: staging-admin-state.py OUTPUT EXAM_ID PAPER_ID')
    output, exam_id, paper_id = sys.argv[1:]
    metadata = subprocess.run(
        ['sudo', '-n', 'mysql', 'element', '--batch', '--skip-column-names', '-e',
         "SELECT e.title,r.participant_name FROM el_competency_result r "
         "INNER JOIN el_exam e ON e.id=r.exam_id WHERE r.exam_id='" + exam_id.replace("'", "''") +
         "' AND r.paper_id='" + paper_id.replace("'", "''") + "'"],
        check=True, capture_output=True, text=True,
    ).stdout.strip().split('\t')
    if len(metadata) != 2:
        raise RuntimeError('exam/paper metadata not found')
    exam_title, participant_name = metadata
    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    token_id = 'browser-buttons-' + uuid.uuid4().hex[:12]
    redis_key = 'login_tokens:' + token_id
    now_ms = int(time.time() * 1000)
    login_user = {
        'userId': 1, 'token': token_id, 'loginTime': now_ms,
        'expireTime': now_ms + 1800000, 'permissions': ['*:*:*'], 'roles': ['admin'],
        'user': {'userId': 1, 'userName': 'admin', 'nickName': 'admin', 'avatar': '', 'admin': True},
    }
    subprocess.run(
        ['redis-cli', '-n', redis_db, 'SET', redis_key, json.dumps(login_user, separators=(',', ':')), 'EX', '1800'],
        check=True, stdout=subprocess.DEVNULL,
    )
    state = {
        'adminToken': sign(secret, token_id), 'redisDb': redis_db, 'redisKey': redis_key,
        'examId': exam_id, 'paperId': paper_id, 'examTitle': exam_title, 'participantName': participant_name,
    }
    path = Path(output)
    path.write_text(json.dumps(state, ensure_ascii=False), encoding='utf-8')
    path.chmod(0o600)


if __name__ == '__main__':
    main()
