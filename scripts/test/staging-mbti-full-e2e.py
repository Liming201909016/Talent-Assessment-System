#!/usr/bin/env python3
"""Real staging MBTI lifecycle: setup -> 48 answers -> submit -> score -> reports -> cleanup."""
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


def request(path, token=None, body=None, method='POST', expect_json=True, timeout=240):
    headers = {}
    if token:
        headers['Authorization'] = 'Bearer ' + token
    payload = None
    if body is not None:
        headers['Content-Type'] = 'application/json'
        payload = json.dumps(body, ensure_ascii=False, separators=(',', ':')).encode()
    call = urllib.request.Request(BASE + path, data=payload, headers=headers, method=method)
    try:
        with urllib.request.urlopen(call, timeout=timeout) as response:
            raw = response.read()
            status = response.status
            response_headers = dict(response.headers)
    except urllib.error.HTTPError as error:
        raw = error.read()
        status = error.code
        response_headers = dict(error.headers)
    if not expect_json:
        return status, raw, response_headers
    data = json.loads(raw.decode())
    if status >= 400 or data.get('code') not in (0, 200):
        raise RuntimeError(f'{path}: HTTP {status}: {data.get("msg")}')
    return data.get('data') or {}


def safe_report_path(value):
    path = Path(value).resolve()
    allowed = [
        (APP_DIR / 'tmp' / 'uploadPath').resolve(),
        Path('/data/uploadPath').resolve(),
    ]
    if not any(path == root or root in path.parents for root in allowed):
        raise RuntimeError(f'report path outside allowed roots: {path}')
    return path


def main():
    suffix = uuid.uuid4().hex[:10]
    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    token_id = 'mbti-full-' + suffix
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
    admin = sign(secret, token_id)
    exam_id = candidate_id = paper_id = ''
    generated_paths = set()
    baseline = mysql("SELECT CONCAT((SELECT COUNT(*) FROM el_exam),(SELECT '|'),(SELECT COUNT(*) FROM el_paper),(SELECT '|'),(SELECT COUNT(*) FROM el_candidate),(SELECT '|'),(SELECT COUNT(*) FROM el_mbti_answer))")
    try:
        repo = mysql("SELECT CONCAT(id,'|',radio_count) FROM el_repo WHERE code='00301'")
        repo_id, radio_count = repo.split('|')
        if int(radio_count) != 48:
            raise RuntimeError(f'00301 question count={radio_count}')
        exam = request('/exam/api/exam/exam/save', admin, {
            'title': 'MBTI-FULL-' + suffix,
            'content': 'temporary staging MBTI full-chain verification',
            'assessmentType': 'legacy', 'scoringMode': 'legacy',
            'joinType': 1, 'openType': 1, 'isOpen': 1, 'answerType': 1,
            'state': 0, 'totalTime': 30, 'requiredFields': 'name,gender,telephone',
            'repoList': [{'repoId': repo_id, 'radioCount': 48, 'radioScore': 1, 'multiCount': 0, 'multiScore': 0, 'judgeCount': 0, 'judgeScore': 0}],
            'departIds': [],
        })
        exam_id = exam['id']
        candidate = request('/exam/api/candidate/save', admin, {
            'examId': exam_id, 'name': 'MBTI Verify ' + suffix,
            'telephone': '138' + suffix[:8], 'gender': '0',
        })
        candidate_id = candidate['id']
        paper = request('/exam/api/paper/paper/create-paper', admin, {'examId': exam_id})
        paper_id = paper['id']
        mysql_exec(f"UPDATE el_candidate SET paper_id='{paper_id}' WHERE id='{candidate_id}' AND exam_id='{exam_id}';")

        detail = request('/exam/api/mbti/paper-detail', body={'paperId': paper_id})
        questions = detail.get('quList') or []
        if len(questions) != 48 or len({row['quId'] for row in questions}) != 48:
            raise RuntimeError(f'MBTI paper shape mismatch: {len(questions)}')
        invalid = request('/exam/api/mbti/fill-answer', body={
            'paperId': paper_id, 'quId': questions[0]['quId'], 'scoreA': 4, 'scoreB': 4,
        }, expect_json=False)
        invalid_json = json.loads(invalid[1].decode())
        if invalid_json.get('code') in (0, 200) or '之和必须为 5' not in invalid_json.get('msg', ''):
            raise RuntimeError(f'MBTI invalid-answer guard mismatch: {invalid_json}')
        for question in questions:
            request('/exam/api/mbti/fill-answer', body={
                'paperId': paper_id, 'quId': question['quId'], 'scoreA': 3, 'scoreB': 2,
            })
        restored = request('/exam/api/mbti/paper-detail', body={'paperId': paper_id})
        if sum(1 for row in restored.get('quList') or [] if row.get('answered')) != 48:
            raise RuntimeError('MBTI answer persistence mismatch')
        submitted = request('/exam/api/mbti/submit', body={'paperId': paper_id})
        mbti_type = submitted.get('type', '')
        if not re.fullmatch(r'[EI][NS][FT][JP]', mbti_type):
            raise RuntimeError(f'invalid MBTI type: {mbti_type}')
        scored = request('/exam/api/mbti/score', body={'paperId': paper_id})
        if scored.get('type') != mbti_type or scored.get('scores') != submitted.get('scores'):
            raise RuntimeError('MBTI score/read-back mismatch')

        deadline = time.time() + 120
        while time.time() < deadline:
            if mysql(f"SELECT COUNT(*) FROM el_candidate WHERE id='{candidate_id}' AND pdf_flag=1") == '1':
                break
            time.sleep(2)
        full = request('/exam/api/mbti/generate-report', admin, {'paperId': paper_id, 'type': 'full', 'force': True})
        simple = request('/exam/api/mbti/generate-report', admin, {'paperId': paper_id, 'type': 'simple', 'force': True})
        if full.get('type') != mbti_type or simple.get('type') != mbti_type:
            raise RuntimeError('MBTI report type mismatch')
        generated_paths.update([full.get('path', ''), simple.get('path', '')])
        for report_type in ('full', 'simple'):
            status, payload, headers = request('/exam/api/mbti/download-report', admin, {'paperId': paper_id, 'type': report_type}, expect_json=False)
            if status != 200 or not payload.startswith(b'%PDF') or 'application/pdf' not in headers.get('Content-Type', ''):
                raise RuntimeError(f'MBTI {report_type} download invalid')
            if len(payload) < 10000:
                raise RuntimeError(f'MBTI {report_type} PDF too small: {len(payload)}')
        templates = request('/exam/api/mbti/templates', admin, method='GET', body=None)
        if not isinstance(templates, list) or len(templates) != 16:
            raise RuntimeError(f'MBTI template count mismatch: {len(templates) if isinstance(templates, list) else templates}')
        if sum(1 for row in templates if row.get('exists')) != 16 or sum(1 for row in templates if row.get('simpleExists')) != 16:
            raise RuntimeError('MBTI full/simple templates are incomplete')
        unauth_status, _, _ = request('/exam/api/mbti/templates', method='GET', body=None, expect_json=False)
        if unauth_status not in (401, 403):
            raise RuntimeError(f'MBTI template anonymous boundary={unauth_status}')
        print('STAGING_MBTI_FULL_PASS')
        print(f'paper=48|answers=48|type={mbti_type}|score_readback=true')
        print('reports=full_pdf:true|simple_pdf:true|templates:16/16|anonymous_templates:blocked')
    finally:
        for value in generated_paths:
            if value:
                path = safe_report_path(value)
                subprocess.run(['sudo', '-n', 'rm', '-f', str(path)], check=True)
        if paper_id:
            try:
                request('/exam/api/paper/paper/delete', admin, {'ids': [paper_id]})
            except Exception as error:
                print('paper_cleanup_error=' + str(error))
        if candidate_id:
            try:
                request('/exam/api/candidate/' + urllib.parse.quote(candidate_id), admin, method='DELETE', body=None)
            except Exception as error:
                print('candidate_cleanup_error=' + str(error))
        if exam_id:
            try:
                request('/exam/api/exam/exam/delete', admin, {'ids': [exam_id]})
            except Exception as error:
                print('exam_cleanup_error=' + str(error))
        subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)
        remaining = mysql(f"SELECT CONCAT((SELECT COUNT(*) FROM el_exam WHERE id='{exam_id}'),'|',(SELECT COUNT(*) FROM el_paper WHERE id='{paper_id}'),'|',(SELECT COUNT(*) FROM el_candidate WHERE id='{candidate_id}'),'|',(SELECT COUNT(*) FROM el_mbti_answer WHERE paper_id='{paper_id}'))")
        current = mysql("SELECT CONCAT((SELECT COUNT(*) FROM el_exam),(SELECT '|'),(SELECT COUNT(*) FROM el_paper),(SELECT '|'),(SELECT COUNT(*) FROM el_candidate),(SELECT '|'),(SELECT COUNT(*) FROM el_mbti_answer))")
        print('cleanup_remaining=' + remaining)
        print('baseline_before=' + baseline)
        print('baseline_after=' + current)
        if remaining != '0|0|0|0' or current != baseline:
            raise RuntimeError(f'MBTI cleanup drift: remaining={remaining}, before={baseline}, after={current}')


if __name__ == '__main__':
    main()
