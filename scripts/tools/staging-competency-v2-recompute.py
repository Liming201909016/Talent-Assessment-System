#!/usr/bin/env python3
import argparse
import base64
import hashlib
import hmac
import json
import re
import subprocess
import time
import urllib.error
import urllib.request
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'


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


def mysql(sql):
    return subprocess.run(
        ['sudo', '-n', 'mysql', 'element', '-Nse', sql],
        check=True, capture_output=True, text=True,
    ).stdout.strip()


def api(path, body, token):
    request = urllib.request.Request(
        BASE + path,
        data=json.dumps(body, separators=(',', ':')).encode(),
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


def download_report(paper_id, token, expected_sha):
    request = urllib.request.Request(
        BASE + '/exam/api/competency/reports/download?paperId=' + paper_id,
        headers={'Authorization': 'Bearer ' + token}, method='GET',
    )
    with urllib.request.urlopen(request, timeout=120) as response:
        content = response.read()
        content_type = response.headers.get_content_type()
    actual_sha = hashlib.sha256(content).hexdigest()
    if content_type != 'application/pdf' or actual_sha != expected_sha:
        raise RuntimeError(f'download mismatch: type={content_type}, sha={actual_sha}')
    Path('/tmp/v2-report-download.pdf').write_bytes(content)
    print(f'STAGING_V2_REPORT_DOWNLOAD_PASS size={len(content)} sha256={actual_sha}')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--generate-report', action='store_true')
    args = parser.parse_args()
    paper_ids = [value for value in mysql(
        "SELECT r.paper_id FROM el_competency_result r "
        "INNER JOIN el_paper p ON p.id=r.paper_id "
        "WHERE r.is_complete=1 "
        "AND r.product_version='competency-frontline-phase1-v1' "
        "AND r.scoring_version='competency-phase1-scoring-v1' "
        "AND r.content_version='competency-phase1-content-v1' "
        "AND r.report_template_version='competency-phase1-report-v1' "
        "AND r.report_audience='frontline_employee' ORDER BY r.paper_id"
    ).splitlines() if value]
    if not paper_ids:
        raise RuntimeError('no eligible completed v1 paper found')

    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    if not secret:
        raise RuntimeError('JWT secret is unavailable')
    token_id = 'v2-recompute-' + str(int(time.time()))
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
    created = reused = 0
    try:
        token = sign(secret, token_id)
        for paper_id in paper_ids:
            result = api('/exam/api/competency/results/recompute-v2', {'paperId': paper_id}, token)
            if result.get('reused'):
                reused += 1
            else:
                created += 1
        if args.generate_report:
            report = api(
                '/exam/api/competency/reports/generate',
                {'paperId': paper_ids[0], 'force': True}, token,
            )
            print('STAGING_V2_REPORT_GENERATE_PASS')
            print(json.dumps(report, ensure_ascii=False, sort_keys=True))
            download_report(paper_ids[0], token, report['pdfSha256'])
    finally:
        subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)

    counts = mysql(
        "SELECT CONCAT(COUNT(DISTINCT r.id),'|',COUNT(DISTINCT o.result_run_id),'|',"
        "COUNT(DISTINCT m.id),'|',COUNT(DISTINCT d.id),'|',COUNT(DISTINCT v.result_run_id)) "
        "FROM el_competency_result_run r "
        "LEFT JOIN el_competency_result_run_overall o ON o.result_run_id=r.id "
        "LEFT JOIN el_competency_result_run_module m ON m.result_run_id=r.id "
        "LEFT JOIN el_competency_result_run_dimension d ON d.result_run_id=r.id "
        "LEFT JOIN el_competency_result_run_validity v ON v.result_run_id=r.id "
        "WHERE r.scoring_version='competency-phase1-scoring-v2'"
    )
    expected = f'{len(paper_ids)}|{len(paper_ids)}|{len(paper_ids) * 3}|{len(paper_ids) * 10}|{len(paper_ids)}'
    if counts != expected:
        raise RuntimeError(f'v2 result-run counts={counts}, want={expected}')
    print('STAGING_V2_RECOMPUTE_PASS')
    print(f'eligible={len(paper_ids)}|created={created}|reused={reused}|counts={counts}')


if __name__ == '__main__':
    main()
