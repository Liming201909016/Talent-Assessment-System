#!/usr/bin/env python3
import base64, concurrent.futures, hashlib, hmac, json, re, statistics, subprocess, time, urllib.error, urllib.request, uuid
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'
PAPER_COUNT = 100
WORKERS = 10


def b64url(value):
    return base64.urlsafe_b64encode(value).rstrip(b'=').decode()


def config_text():
    env = subprocess.run(['systemctl', 'show', 'talent-assessment', '--property=Environment', '--value'], check=True, capture_output=True, text=True).stdout
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
    return f'{raw}.{b64url(hmac.new(secret.encode(), raw.encode(), hashlib.sha512).digest())}'


def api(path, body, token=None, headers=None, timeout=120):
    request_headers = {'Content-Type': 'application/json'}
    if token:
        request_headers['Authorization'] = 'Bearer ' + token
    if headers:
        request_headers.update(headers)
    request = urllib.request.Request(BASE + path, data=json.dumps(body, ensure_ascii=False).encode(), headers=request_headers, method='POST')
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            payload = json.loads(response.read().decode())
    except urllib.error.HTTPError as error:
        raise RuntimeError(f'{path} HTTP {error.code}: {error.read().decode(errors="replace")}')
    if payload.get('code') not in (0, 200):
        raise RuntimeError(f'{path}: {payload.get("msg")}')
    return payload.get('data')


def mysql(sql):
    return subprocess.run(['sudo', '-n', 'mysql', 'element', '-Nse', sql], check=True, capture_output=True, text=True).stdout.strip()


def percentile(values, fraction):
    ordered = sorted(values)
    index = min(len(ordered) - 1, max(0, int(len(ordered) * fraction + 0.999999) - 1))
    return ordered[index]


def main():
    suffix = uuid.uuid4().hex[:10]
    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    token_id = 'competency-capacity-' + suffix
    redis_key = 'login_tokens:' + token_id
    now_ms = int(time.time() * 1000)
    login = {'userId': 1, 'token': token_id, 'loginTime': now_ms, 'expireTime': now_ms + 3600000, 'permissions': ['*:*:*'], 'roles': ['admin'], 'user': None}
    subprocess.run(['redis-cli', '-n', redis_db, 'SET', redis_key, json.dumps(login, separators=(',', ':')), 'EX', '3600'], check=True, stdout=subprocess.DEVNULL)
    admin = sign(secret, token_id)
    exam_id = None
    try:
        dimensions = api('/exam/api/competency/dimensions/list', {}, admin)
        enabled = [item for item in dimensions if item['status'] == 0]
        if len(enabled) != 48 or sum(int(item['questionCount']) for item in enabled) != 384:
            raise RuntimeError(f'dimension baseline mismatch: enabled={len(enabled)} questions={sum(int(item["questionCount"]) for item in enabled)}')
        dimension_ids = [item['id'] for item in sorted(enabled, key=lambda item: item['displayOrder'])]
        exam = api('/exam/api/exam/exam/save', {
            'title': 'CAPACITY-48X384-' + suffix, 'content': 'temporary full competency capacity verification',
            'assessmentType': 'competency', 'scoringMode': 'competency_average', 'competencyReportAudience': 'leader',
            'dimensionIds': dimension_ids, 'joinType': 1, 'openType': 1, 'isOpen': 1, 'answerType': 1,
            'state': 0, 'totalTime': 120, 'repoList': [], 'departIds': []
        }, admin)
        exam_id = exam['id']
        publish_start = time.perf_counter()
        published = api('/exam/api/competency/exams/publish', {'examId': exam_id}, admin)
        publish_seconds = time.perf_counter() - publish_start
        if published['dimensionCount'] != 48 or published['questionCount'] != 384:
            raise RuntimeError(f'publish mismatch: {published}')
        snapshot_counts = mysql(f"SELECT COUNT(DISTINCT exam_dimension_id),COUNT(*),COUNT(DISTINCT question_code),COUNT(DISTINCT source_qu_id) FROM el_exam_competency_question WHERE exam_id='{exam_id}'")
        if snapshot_counts != '48\t384\t384\t384':
            raise RuntimeError('snapshot counts mismatch: ' + snapshot_counts)

        def create_one(index):
            started = time.perf_counter()
            telephone = f'139{index:08d}'
            candidate = api('/exam/api/candidate/save', {'examId': exam_id, 'name': f'Capacity {index:03d}', 'telephone': telephone, 'gender': '0'}, admin)
            access = api('/exam/api/competency/participant/create-paper', {
                'examId': exam_id, 'participantId': candidate['id'], 'participantType': 'candidate',
                'participantToken': candidate['participantToken']
            })
            detail = api('/exam/api/competency/participant/paper-detail', {'paperId': access['paperId']}, headers={'X-Competency-Token': access['paperToken']})
            elapsed = time.perf_counter() - started
            ids = [item['id'] for item in detail['questions']]
            codes = [item['code'] for item in detail['questions']]
            if detail['totalCount'] != 384 or len(ids) != 384 or len(set(ids)) != 384 or len(set(codes)) != 384:
                raise RuntimeError(f'paper {index} integrity mismatch total={detail["totalCount"]} ids={len(ids)}/{len(set(ids))} codes={len(set(codes))}')
            return {'index': index, 'paperId': access['paperId'], 'paperToken': access['paperToken'], 'ids': ids, 'hash': hashlib.sha256('|'.join(ids).encode()).hexdigest(), 'seconds': elapsed}

        create_start = time.perf_counter()
        with concurrent.futures.ThreadPoolExecutor(max_workers=WORKERS) as pool:
            papers = list(pool.map(create_one, range(1, PAPER_COUNT + 1)))
        create_total_seconds = time.perf_counter() - create_start
        hashes = {paper['hash'] for paper in papers}
        if len(hashes) != PAPER_COUNT:
            raise RuntimeError(f'independent order hashes={len(hashes)}, want {PAPER_COUNT}')

        def restore_one(paper):
            started = time.perf_counter()
            detail = api('/exam/api/competency/participant/paper-detail', {'paperId': paper['paperId']}, headers={'X-Competency-Token': paper['paperToken']})
            elapsed = time.perf_counter() - started
            ids = [item['id'] for item in detail['questions']]
            if ids != paper['ids']:
                raise RuntimeError(f'paper {paper["index"]} order changed on refresh')
            return elapsed

        with concurrent.futures.ThreadPoolExecutor(max_workers=WORKERS) as pool:
            restore_seconds = list(pool.map(restore_one, papers))

        db_counts = mysql(f"SELECT (SELECT COUNT(*) FROM el_candidate WHERE exam_id='{exam_id}'),(SELECT COUNT(*) FROM el_paper WHERE exam_id='{exam_id}'),(SELECT COUNT(*) FROM el_paper_qu pq INNER JOIN el_paper p ON p.id=pq.paper_id WHERE p.exam_id='{exam_id}'),(SELECT COUNT(DISTINCT pq.paper_id) FROM el_paper_qu pq INNER JOIN el_paper p ON p.id=pq.paper_id WHERE p.exam_id='{exam_id}' AND pq.exam_question_id IS NOT NULL)")
        if db_counts != '100\t100\t38400\t100':
            raise RuntimeError('database capacity counts mismatch: ' + db_counts)
        distribution = mysql(f"SELECT MIN(c),MAX(c),SUM(c) FROM (SELECT COUNT(*) c FROM el_paper_qu pq INNER JOIN el_paper p ON p.id=pq.paper_id WHERE p.exam_id='{exam_id}' GROUP BY pq.paper_id) x")
        if distribution != '384\t384\t38400':
            raise RuntimeError('paper question distribution mismatch: ' + distribution)

        create_latencies = [paper['seconds'] for paper in papers]
        print('COMPETENCY_CAPACITY_VERIFY_OK')
        print('publish=48_dimensions|384_snapshots|seconds=%.3f' % publish_seconds)
        print(f'papers={PAPER_COUNT}|paper_questions=38400|distinct_order_hashes={len(hashes)}|stable_refreshes={len(restore_seconds)}')
        print('create_seconds_total=%.3f|p50=%.3f|p95=%.3f|max=%.3f' % (create_total_seconds, statistics.median(create_latencies), percentile(create_latencies, .95), max(create_latencies)))
        print('refresh_seconds_p50=%.3f|p95=%.3f|max=%.3f' % (statistics.median(restore_seconds), percentile(restore_seconds, .95), max(restore_seconds)))
        print('db_counts=' + db_counts.replace('\t', '|'))
        print('distribution=' + distribution.replace('\t', '|'))
        print('exam_id=' + exam_id)
    finally:
        if exam_id:
            try:
                api('/exam/api/exam/exam/delete', {'ids': [exam_id]}, admin)
            except Exception as error:
                print('cleanup_api_error=' + str(error))
        subprocess.run(['redis-cli', '-n', redis_db, 'DEL', redis_key], check=False, stdout=subprocess.DEVNULL)
        if exam_id:
            remaining = mysql(f"SELECT (SELECT COUNT(*) FROM el_exam WHERE id='{exam_id}')+(SELECT COUNT(*) FROM el_candidate WHERE exam_id='{exam_id}')+(SELECT COUNT(*) FROM el_paper WHERE exam_id='{exam_id}')+(SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id='{exam_id}')+(SELECT COUNT(*) FROM el_exam_competency_dimension WHERE exam_id='{exam_id}')")
            print('cleanup_remaining=' + remaining)
            if remaining != '0':
                raise RuntimeError('cleanup remaining=' + remaining)


if __name__ == '__main__':
    main()
