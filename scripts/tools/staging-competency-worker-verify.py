#!/usr/bin/env python3
import base64, concurrent.futures, hashlib, hmac, json, re, subprocess, time, urllib.error, urllib.request, uuid
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'


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


def sign(secret, claims):
    header = b64url(json.dumps({'alg': 'HS512', 'typ': 'JWT'}, separators=(',', ':')).encode())
    payload = b64url(json.dumps(claims, separators=(',', ':')).encode())
    raw = f'{header}.{payload}'
    return f'{raw}.{b64url(hmac.new(secret.encode(), raw.encode(), hashlib.sha512).digest())}'


def admin_token(secret, redis_db, token_id):
    now_ms = int(time.time() * 1000)
    login = {'userId': 1, 'token': token_id, 'loginTime': now_ms, 'expireTime': now_ms + 1800000, 'permissions': ['*:*:*'], 'roles': ['admin'], 'user': None}
    subprocess.run(['redis-cli', '-n', redis_db, 'SET', 'login_tokens:' + token_id, json.dumps(login, separators=(',', ':')), 'EX', '1800'], check=True, stdout=subprocess.DEVNULL)
    return sign(secret, {'login_user_key': token_id})


def participant_token(secret, participant_type, participant_id, exam_id):
    now = int(time.time())
    return sign(secret, {'purpose': 'competency_participant', 'participant_type': participant_type, 'participant_id': participant_id, 'exam_id': exam_id, 'paper_id': '', 'iat': now, 'exp': now + 1800})


def api(path, body, token=None, headers=None):
    request_headers = {'Content-Type': 'application/json'}
    if token:
        request_headers['Authorization'] = 'Bearer ' + token
    if headers:
        request_headers.update(headers)
    request = urllib.request.Request(BASE + path, data=json.dumps(body, ensure_ascii=False).encode(), headers=request_headers, method='POST')
    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            payload = json.loads(response.read().decode())
    except urllib.error.HTTPError as error:
        raise RuntimeError(f'{path} HTTP {error.code}: {error.read().decode(errors="replace")}')
    if payload.get('code') not in (0, 200):
        raise RuntimeError(f'{path}: {payload.get("msg")}')
    return payload.get('data')


def mysql(sql):
    return subprocess.run(['sudo', '-n', 'mysql', 'element', '-Nse', sql], check=True, capture_output=True, text=True).stdout.strip()


def create_candidate_paper(admin, exam_id, name, telephone):
    candidate = api('/exam/api/candidate/save', {'examId': exam_id, 'name': name, 'telephone': telephone, 'gender': '0'}, admin)
    access = api('/exam/api/competency/participant/create-paper', {'examId': exam_id, 'participantId': candidate['id'], 'participantType': 'candidate', 'participantToken': candidate['participantToken']})
    detail = api('/exam/api/competency/participant/paper-detail', {'paperId': access['paperId']}, headers={'X-Competency-Token': access['paperToken']})
    return candidate['id'], access, detail


def fill(access, question, value=3):
    api('/exam/api/competency/participant/fill-answer', {'paperId': access['paperId'], 'paperQuestionId': question['id'], 'rawValue': value}, headers={'X-Competency-Token': access['paperToken']})


def main():
    suffix = uuid.uuid4().hex[:10]
    text = config_text()
    secret = scalar(text, 'jwt', 'secret')
    redis_db = scalar(text, 'redis', 'db', '1')
    token_id = 'competency-worker-' + suffix
    admin = admin_token(secret, redis_db, token_id)
    exam_id = None
    tester_id = 'worker-tester-' + suffix
    try:
        dimensions = api('/exam/api/competency/dimensions/list', {}, admin)
        selected = [item['id'] for item in dimensions if item['code'] in ('D01', 'D02')]
        exam = api('/exam/api/exam/exam/save', {
            'title': 'WORKER-' + suffix, 'content': 'temporary worker verification',
            'assessmentType': 'competency', 'scoringMode': 'competency_average', 'competencyReportAudience': 'leader',
            'dimensionIds': selected, 'joinType': 1, 'openType': 1, 'isOpen': 1, 'answerType': 1,
            'state': 0, 'totalTime': 30, 'repoList': [], 'departIds': []
        }, admin)
        exam_id = exam['id']
        api('/exam/api/competency/exams/publish', {'examId': exam_id}, admin)

        _, concurrent_access, concurrent_detail = create_candidate_paper(admin, exam_id, 'Worker Concurrent', '13810000001')
        for question in concurrent_detail['questions']:
            fill(concurrent_access, question, 3)
        mysql(f"UPDATE el_paper SET limit_time=DATE_SUB(NOW(), INTERVAL 1 MINUTE) WHERE id='{concurrent_access['paperId']}'")
        def submit_once(_):
            return api('/exam/api/competency/participant/submit', {'paperId': concurrent_access['paperId'], 'submitType': 'manual'}, headers={'X-Competency-Token': concurrent_access['paperToken']})
        with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
            submits = list(pool.map(submit_once, range(2)))
        concurrent_counts = mysql(f"SELECT (SELECT COUNT(*) FROM el_competency_result WHERE paper_id='{concurrent_access['paperId']}'),(SELECT COUNT(*) FROM el_competency_dimension_result WHERE paper_id='{concurrent_access['paperId']}')")
        if concurrent_counts != '1\t2' or not all(item.get('isComplete') for item in submits):
            raise RuntimeError(f'concurrent submit mismatch: responses={submits} counts={concurrent_counts}')

        _, partial_access, partial_detail = create_candidate_paper(admin, exam_id, 'Worker Partial', '13810000002')
        fill(partial_access, partial_detail['questions'][0], 3)

        now = time.strftime('%Y-%m-%d %H:%M:%S')
        mysql("INSERT INTO el_tester (id,exam_id,id_number,name,password,telephone,del_flag,create_time,update_time) VALUES "
              f"('{tester_id}','{exam_id}','ID-{suffix}','Worker Zero','0000','13810000003',0,'{now}','{now}')")
        tester_access = api('/exam/api/competency/participant/create-paper', {
            'examId': exam_id, 'participantId': tester_id, 'participantType': 'tester',
            'participantToken': participant_token(secret, 'tester', tester_id, exam_id)
        })
        api('/exam/api/competency/participant/paper-detail', {'paperId': tester_access['paperId']}, headers={'X-Competency-Token': tester_access['paperToken']})

        mysql(f"UPDATE el_paper SET limit_time=DATE_SUB(NOW(), INTERVAL 1 MINUTE) WHERE id IN ('{partial_access['paperId']}','{tester_access['paperId']}')")
        subprocess.run(['sudo', '-n', 'systemctl', 'restart', 'talent-assessment'], check=True)
        deadline = time.time() + 15
        while time.time() < deadline:
            count = mysql(f"SELECT COUNT(*) FROM el_competency_result WHERE paper_id IN ('{partial_access['paperId']}','{tester_access['paperId']}')")
            if count == '2':
                break
        else:
            raise RuntimeError('initial startup scan did not submit two expired papers within 15 seconds')

        partial = mysql(f"SELECT answered_question_count,total_question_count,effective_dimension_count,overall_score,is_complete,submit_type FROM el_competency_result WHERE paper_id='{partial_access['paperId']}'")
        zero = mysql(f"SELECT answered_question_count,total_question_count,effective_dimension_count,overall_score,is_complete,submit_type FROM el_competency_result WHERE paper_id='{tester_access['paperId']}'")
        partial_dims = mysql(f"SELECT COUNT(*),SUM(dimension_score IS NULL) FROM el_competency_dimension_result WHERE paper_id='{partial_access['paperId']}'")
        zero_dims = mysql(f"SELECT COUNT(*),SUM(dimension_score IS NULL) FROM el_competency_dimension_result WHERE paper_id='{tester_access['paperId']}'")
        if partial != '1\t16\t1\t3.000000\t0\ttimeout' or partial_dims != '2\t1':
            raise RuntimeError(f'partial result mismatch: result={partial} dims={partial_dims}')
        if zero != '0\t16\t0\t0.000000\t0\ttimeout' or zero_dims != '2\t2':
            raise RuntimeError(f'zero result mismatch: result={zero} dims={zero_dims}')
        print('COMPETENCY_WORKER_VERIFY_OK')
        print('concurrent_submit=1_overall|2_dimensions')
        print('partial=' + partial.replace('\t', '|') + '|nil_dimensions=1')
        print('zero=' + zero.replace('\t', '|') + '|nil_dimensions=2')
        print('startup_scan_seconds_lt=15')
        print('exam_id=' + exam_id)
    finally:
        if exam_id:
            try:
                api('/exam/api/exam/exam/delete', {'ids': [exam_id]}, admin)
            except Exception as error:
                print('cleanup_api_error=' + str(error))
        subprocess.run(['redis-cli', '-n', redis_db, 'DEL', 'login_tokens:' + token_id], check=False, stdout=subprocess.DEVNULL)
        if exam_id:
            remaining = mysql(f"SELECT (SELECT COUNT(*) FROM el_exam WHERE id='{exam_id}')+(SELECT COUNT(*) FROM el_paper WHERE exam_id='{exam_id}')+(SELECT COUNT(*) FROM el_competency_result WHERE exam_id='{exam_id}')+(SELECT COUNT(*) FROM el_candidate WHERE exam_id='{exam_id}')+(SELECT COUNT(*) FROM el_tester WHERE exam_id='{exam_id}')")
            print('cleanup_remaining=' + remaining)
            if remaining != '0':
                raise RuntimeError('cleanup remaining=' + remaining)


if __name__ == '__main__':
    main()
