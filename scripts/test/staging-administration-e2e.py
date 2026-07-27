#!/usr/bin/env python3
import base64
import hashlib
import hmac
import json
import os
import re
import secrets
import shlex
import subprocess
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
APP_BASE = 'http://127.0.0.1:8092'
PUBLIC_BASE = 'http://127.0.0.1'
DATABASE = 'element'
REGISTER_KEY = 'sys.account.registerUser'
RETIRED_MENU_IDS = [109, 110, 113, 115, 1041, 1042, 1044, 1045, 1046, 1047, 1048,
                    1049, 1050, 1051, 1052, 1053, 1054, 1055, 1056, 1057, 1058, 1059, 1060]
TINY_PNG = base64.b64decode(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII='
)


def b64url(value):
    return base64.urlsafe_b64encode(value).rstrip(b'=').decode()


def parse_bool(value):
    return str(value).strip().lower() in ('1', 'true', 'yes', 'on')


def systemd_environment():
    raw = subprocess.run(
        ['systemctl', 'show', 'talent-assessment', '--property=Environment', '--value'],
        check=True, capture_output=True, text=True,
    ).stdout.strip()
    result = {}
    for item in shlex.split(raw):
        if '=' in item:
            key, value = item.split('=', 1)
            result[key] = value
    return result


def config_text(environment):
    app_env = environment.get('APP_ENV', 'local')
    files = [APP_DIR / 'configs' / 'application.yml', APP_DIR / 'configs' / f'application-{app_env}.yml']
    return '\n'.join(path.read_text(encoding='utf-8') for path in files if path.exists())


def scalar(text, section, key, default=''):
    value = default
    for section_match in re.finditer(rf'(?ms)^{re.escape(section)}:\s*\n(.*?)(?=^[A-Za-z][\w-]*:\s*$|\Z)', text):
        key_match = re.search(rf'(?m)^\s+{re.escape(key)}:\s*(.+?)\s*$', section_match.group(1))
        if key_match:
            value = key_match.group(1).strip().strip('"\'')
    return value


def runtime_config():
    environment = systemd_environment()
    text = config_text(environment)
    values = {
        'jwt_secret': environment.get('JWT_SECRET', scalar(text, 'jwt', 'secret')),
        'redis_db': environment.get('REDIS_DB', scalar(text, 'redis', 'db', '1')),
        'redis_addr': environment.get('REDIS_ADDR', scalar(text, 'redis', 'addr', '127.0.0.1:6379')),
        'redis_password': environment.get('REDIS_PASSWORD', scalar(text, 'redis', 'password', '')),
        'captcha_enabled': environment.get('CAPTCHA_ENABLED', scalar(text, 'captcha', 'enabled', 'true')),
        'upload_path': environment.get('UPLOAD_PATH', scalar(text, 'upload', 'path', str(APP_DIR / 'tmp/uploadPath'))),
        'profile_url': scalar(text, 'upload', 'profile', '/profile'),
    }
    if not values['jwt_secret']:
        raise RuntimeError('deployed JWT secret is missing')
    return values


def sign(secret, token_id):
    header = b64url(json.dumps({'alg': 'HS512', 'typ': 'JWT'}, separators=(',', ':')).encode())
    payload = b64url(json.dumps({'login_user_key': token_id}, separators=(',', ':')).encode())
    raw = f'{header}.{payload}'
    signature = b64url(hmac.new(secret.encode(), raw.encode(), hashlib.sha512).digest())
    return f'{raw}.{signature}'


def token_id(token):
    payload = token.split('.')[1]
    payload += '=' * (-len(payload) % 4)
    return json.loads(base64.urlsafe_b64decode(payload.encode()))['login_user_key']


def mysql_raw(sql):
    return subprocess.run(
        ['sudo', '-n', 'mysql', DATABASE, '--batch', '--skip-column-names', '--raw', '-e', sql],
        check=True, capture_output=True, text=True,
    ).stdout


def mysql(sql):
    return mysql_raw(sql).strip()


def sql_quote(value):
    if value is None:
        return 'NULL'
    return "'" + str(value).replace('\\', '\\\\').replace("'", "''") + "'"


def redis_command(config, *args, capture=True, check=True):
    host, separator, port = config['redis_addr'].rpartition(':')
    if not separator or not port.isdigit():
        host, port = config['redis_addr'], '6379'
    command = ['redis-cli', '--raw', '-h', host or '127.0.0.1', '-p', port, '-n', str(config['redis_db']), *map(str, args)]
    environment = os.environ.copy()
    if config['redis_password']:
        environment['REDISCLI_AUTH'] = config['redis_password']
    result = subprocess.run(
        command, check=check, env=environment,
        stdout=subprocess.PIPE if capture else subprocess.DEVNULL,
        stderr=subprocess.PIPE if capture else subprocess.DEVNULL,
        text=True,
    )
    return result.stdout.rstrip('\r\n') if capture else ''


def redis_set(config, key, value, ttl=600):
    redis_command(config, 'SET', key, value, 'EX', ttl, capture=False)


def request(path, token=None, method='GET', body=None, content_type='application/json', base=APP_BASE):
    headers = {}
    if token:
        headers['Authorization'] = 'Bearer ' + token
    data = body
    if body is not None:
        headers['Content-Type'] = content_type
    call = urllib.request.Request(base + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(call, timeout=120) as response:
            return response.status, response.read(), dict(response.headers)
    except urllib.error.HTTPError as error:
        return error.code, error.read(), dict(error.headers)


def json_request(path, token=None, method='GET', body=None):
    encoded = None if body is None else json.dumps(body, ensure_ascii=False, separators=(',', ':')).encode()
    status, raw, headers = request(path, token, method, encoded)
    try:
        payload = json.loads(raw.decode())
    except (UnicodeDecodeError, json.JSONDecodeError):
        raise RuntimeError(f'{method} {path}: non-JSON response status={status}')
    return status, payload, headers


def api_success(path, token=None, method='GET', body=None):
    status, payload, _ = json_request(path, token, method, body)
    if status != 200 or payload.get('code') not in (0, 200):
        raise RuntimeError(f'{method} {path}: status={status} code={payload.get("code")} msg={payload.get("msg")}')
    return payload


def api_failure(path, token=None, method='POST', body=None, message_part=None):
    status, payload, _ = json_request(path, token, method, body)
    if status != 200 or payload.get('code') in (0, 200):
        raise RuntimeError(f'{method} {path}: expected HTTP-200 business failure, got status={status} code={payload.get("code")}')
    if message_part and message_part not in str(payload.get('msg', '')):
        raise RuntimeError(f'{method} {path}: unexpected failure message={payload.get("msg")}')
    return payload


def assert_http_404(path, token, method='GET'):
    status, _, _ = request(path, token, method)
    if status != 404:
        raise RuntimeError(f'{method} {path}: status={status}, want 404')


def multipart(field, filename, content, content_type):
    boundary = '----admin-e2e-' + uuid.uuid4().hex
    body = (
        f'--{boundary}\r\nContent-Disposition: form-data; name="{field}"; filename="{filename}"\r\n'
        f'Content-Type: {content_type}\r\n\r\n'
    ).encode() + content + f'\r\n--{boundary}--\r\n'.encode()
    return body, 'multipart/form-data; boundary=' + boundary


def upload_avatar(token, filename, content):
    body, content_type = multipart('avatarfile', filename, content, 'image/png')
    status, raw, _ = request('/system/user/profile/avatar', token, 'POST', body, content_type)
    payload = json.loads(raw.decode())
    return status, payload


def seed_captcha(config, tracked_keys, prefix):
    captcha_uuid = prefix + '-' + uuid.uuid4().hex
    answer = str(secrets.randbelow(9000) + 1000)
    key = 'captcha_codes:' + captcha_uuid
    redis_set(config, key, answer, 300)
    tracked_keys.add(key)
    return captcha_uuid, answer, key


def login(config, tracked_keys, username, password):
    if parse_bool(config['captcha_enabled']):
        captcha_uuid, answer, _ = seed_captcha(config, tracked_keys, 'login')
    else:
        captcha_uuid, answer = '', ''
    payload = api_success('/login', method='POST', body={
        'username': username, 'password': password, 'code': answer, 'uuid': captcha_uuid,
    })
    token = payload.get('token')
    if not token:
        raise RuntimeError('login response did not contain a token')
    tracked_keys.add('login_tokens:' + token_id(token))
    return token


def register(config, tracked_keys, username, password, captcha=None, expect_success=True):
    if captcha is None and parse_bool(config['captcha_enabled']):
        captcha_uuid, answer, captcha_key = seed_captcha(config, tracked_keys, 'register')
    elif captcha is None:
        captcha_uuid, answer, captcha_key = '', '', ''
    else:
        captcha_uuid, answer, captcha_key = captcha
    body = {
        'username': username, 'password': password, 'confirmPassword': password,
        'code': answer, 'uuid': captcha_uuid,
    }
    if expect_success:
        api_success('/register', method='POST', body=body)
    else:
        api_failure('/register', method='POST', body=body)
    return body, (captcha_uuid, answer, captcha_key)


def hex_decode(value):
    return bytes.fromhex(value).decode('utf-8') if value else ''


def read_register_config():
    row = mysql(
        "SELECT config_id,HEX(config_name),HEX(config_key),HEX(config_value),HEX(config_type),"
        "HEX(COALESCE(remark,'')),IF(remark IS NULL,1,0) FROM sys_config "
        f"WHERE config_key={sql_quote(REGISTER_KEY)} ORDER BY config_id"
    )
    if not row:
        return None
    lines = row.splitlines()
    if len(lines) != 1:
        raise RuntimeError('sys.account.registerUser is not unique')
    fields = lines[0].split('\t')
    return {
        'configId': int(fields[0]), 'configName': hex_decode(fields[1]), 'configKey': hex_decode(fields[2]),
        'configValue': hex_decode(fields[3]), 'configType': hex_decode(fields[4]),
        'remark': None if fields[6] == '1' else hex_decode(fields[5]),
    }


def edit_config(admin, snapshot, value):
    body = dict(snapshot)
    body['configValue'] = value
    api_success('/system/config', admin, 'PUT', body)


def config_row_digest(config_id):
    raw = mysql_raw(f'SELECT * FROM sys_config WHERE config_id={int(config_id)}')
    return hashlib.sha256(raw.encode()).hexdigest()


def admin_digest():
    raw = mysql_raw(
        "SELECT * FROM sys_user WHERE user_id=1;"
        "SELECT user_id,role_id FROM sys_user_role WHERE user_id=1 ORDER BY role_id;"
        "SELECT role_id,dept_id FROM sys_role_dept WHERE role_id=1 ORDER BY dept_id;"
        "SELECT role_id,menu_id FROM sys_role_menu WHERE role_id=1 ORDER BY menu_id"
    )
    return hashlib.sha256(raw.encode()).hexdigest()


def ensure_username_index():
    duplicates = int(mysql('SELECT COUNT(*) FROM (SELECT user_name FROM sys_user GROUP BY user_name HAVING COUNT(*)>1) d') or '0')
    definition = mysql(
        "SELECT CONCAT(non_unique,'|',GROUP_CONCAT(column_name ORDER BY seq_in_index SEPARATOR ',')) "
        "FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='sys_user' "
        "AND index_name='uk_sys_user_user_name' GROUP BY non_unique"
    )
    if duplicates:
        print(f'SKIP unique_username_index duplicate_groups={duplicates}')
        return False
    if definition and definition != '0|user_name':
        print('SKIP unique_username_index conflicting_index_definition')
        return False
    if not definition:
        mysql_raw('ALTER TABLE sys_user ADD UNIQUE INDEX uk_sys_user_user_name (user_name)')
    verified = mysql(
        "SELECT CONCAT(non_unique,'|',GROUP_CONCAT(column_name ORDER BY seq_in_index SEPARATOR ',')) "
        "FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='sys_user' "
        "AND index_name='uk_sys_user_user_name' GROUP BY non_unique"
    )
    if verified != '0|user_name':
        raise RuntimeError('unique username index verification failed')
    return True


def create_admin_session(config, tracked_keys):
    session_id = 'administration-e2e-' + uuid.uuid4().hex
    key = 'login_tokens:' + session_id
    now_ms = int(time.time() * 1000)
    login_user = {
        'userId': 1, 'token': session_id, 'loginTime': now_ms, 'expireTime': now_ms + 3600000,
        'permissions': ['*:*:*'], 'roles': ['admin'],
        'user': {'userId': 1, 'userName': 'admin', 'nickName': 'admin', 'avatar': '', 'admin': True},
    }
    redis_set(config, key, json.dumps(login_user, separators=(',', ':')), 3600)
    tracked_keys.add(key)
    return sign(config['jwt_secret'], session_id)


def row_ids(payload):
    return {str(row.get('userId')) for row in payload.get('rows') or []}


def flatten_router_fields(routes):
    fields = {'name': set(), 'path': set(), 'component': set(), 'title': set()}
    for route in routes or []:
        for key in ('name', 'path', 'component'):
            value = route.get(key)
            if value:
                fields[key].add(str(value).lower())
        title = (route.get('meta') or {}).get('title')
        if title:
            fields['title'].add(str(title).lower())
        child = flatten_router_fields(route.get('children') or [])
        for key in fields:
            fields[key].update(child[key])
    return fields


def cache_snapshot(config, key):
    exists = redis_command(config, 'EXISTS', key) == '1'
    if not exists:
        return {'exists': False}
    return {
        'exists': True,
        'value': redis_command(config, 'GET', key),
        'pttl': int(redis_command(config, 'PTTL', key) or '-2'),
    }


def restore_cache_snapshot(config, key, snapshot):
    if not snapshot or not snapshot.get('exists'):
        redis_command(config, 'DEL', key, capture=False, check=False)
        return
    if snapshot['pttl'] > 0:
        redis_command(config, 'SET', key, snapshot['value'], 'PX', snapshot['pttl'], capture=False)
    else:
        redis_command(config, 'SET', key, snapshot['value'], capture=False)


def main():
    suffix = uuid.uuid4().hex[:10]
    username = 'adm_' + suffix
    helper_username = 'batch_' + suffix
    old_password = 'A1' + secrets.token_urlsafe(12)
    new_password = 'B2' + secrets.token_urlsafe(12)
    helper_password = 'C3' + secrets.token_urlsafe(12)
    role_key = 'e2e_role_' + suffix
    dict_type = 'e2e_dict_' + suffix
    config = runtime_config()
    tracked_redis_keys = set()
    created_user_ids = []
    created_role_id = None
    dict_id = None
    dict_code = None
    avatar_path = None
    register_snapshot = None
    register_digest = None
    register_cache = None
    admin_before = admin_digest()
    admin = create_admin_session(config, tracked_redis_keys)
    cleanup_errors = []
    evidence = []

    try:
        if ensure_username_index():
            evidence.append('unique_username_index=PASS')

        register_snapshot = read_register_config()
        if register_snapshot is None:
            raise RuntimeError('sys.account.registerUser config is missing; refusing to invent permanent system config')
        register_digest = config_row_digest(register_snapshot['configId'])
        register_cache_key = 'sys_config:' + REGISTER_KEY
        register_cache = cache_snapshot(config, register_cache_key)

        edit_config(admin, register_snapshot, 'false')
        if mysql(f"SELECT HEX(config_value) FROM sys_config WHERE config_id={register_snapshot['configId']}") != '66616C7365':
            raise RuntimeError('registration disabled value was not persisted')
        if redis_command(config, 'EXISTS', register_cache_key) != '0':
            raise RuntimeError('registration config exact cache invalidation failed')

        if parse_bool(config['captcha_enabled']):
            reg_captcha = seed_captcha(config, tracked_redis_keys, 'register-once')
        else:
            reg_captcha = ('', '', '')
        disabled_body = {
            'username': username, 'password': old_password, 'confirmPassword': old_password,
            'code': reg_captcha[1], 'uuid': reg_captcha[0],
        }
        api_failure('/register', method='POST', body=disabled_body, message_part='未开放')
        if reg_captcha[2] and redis_command(config, 'EXISTS', reg_captcha[2]) != '1':
            raise RuntimeError('disabled registration consumed captcha before feature gate')

        miss = api_success('/system/config/configKey/' + urllib.parse.quote(REGISTER_KEY, safe=''), admin)
        if miss.get('data') != 'false' or redis_command(config, 'GET', register_cache_key) != 'false':
            raise RuntimeError('config cache miss/read-through mismatch')
        redis_set(config, register_cache_key, 'cache-hit-sentinel', 600)
        hit = api_success('/system/config/configKey/' + urllib.parse.quote(REGISTER_KEY, safe=''), admin)
        if hit.get('data') != 'cache-hit-sentinel':
            raise RuntimeError('config cache hit did not return the cached value')

        unrelated_config_key = 'sys_config:e2e-unrelated-' + suffix
        tracked_redis_keys.add(unrelated_config_key)
        redis_set(config, unrelated_config_key, 'keep-until-refresh', 600)
        edit_config(admin, register_snapshot, 'true')
        if redis_command(config, 'EXISTS', register_cache_key) != '0':
            raise RuntimeError('config edit did not invalidate the exact key')
        if redis_command(config, 'GET', unrelated_config_key) != 'keep-until-refresh':
            raise RuntimeError('config edit invalidated an unrelated key')

        if parse_bool(config['captcha_enabled']):
            api_success('/register', method='POST', body=disabled_body)
            api_failure('/register', method='POST', body=disabled_body, message_part='验证码已失效')
            if redis_command(config, 'EXISTS', reg_captcha[2]) != '0':
                raise RuntimeError('registration captcha was not consumed atomically')
            evidence.append('registration_gate_and_captcha_replay=PASS')
        else:
            register(config, tracked_redis_keys, username, old_password)
            print('SKIP one_time_registration_captcha deployed_captcha_disabled')
            evidence.append('registration_gate=PASS')

        user_id = int(mysql(f"SELECT user_id FROM sys_user WHERE user_name={sql_quote(username)}"))
        created_user_ids.append(user_id)
        if user_id == 1:
            raise RuntimeError('temporary registration resolved to protected admin')

        register(config, tracked_redis_keys, helper_username, helper_password)
        helper_user_id = int(mysql(f"SELECT user_id FROM sys_user WHERE user_name={sql_quote(helper_username)}"))
        created_user_ids.append(helper_user_id)

        first_token = login(config, tracked_redis_keys, username, old_password)
        info = api_success('/getInfo', first_token)
        if str((info.get('user') or {}).get('userId')) != str(user_id):
            raise RuntimeError('login/getInfo user mismatch')

        profile_body = {
            'userId': 1, 'nickName': 'E2E-' + suffix,
            'email': suffix + '@example.test', 'phonenumber': '139' + str(int(suffix[:8], 16) % 100000000).zfill(8), 'sex': '2',
        }
        api_success('/system/user/profile', first_token, 'PUT', profile_body)
        profile = api_success('/system/user/profile', first_token).get('data') or {}
        if str(profile.get('userId')) != str(user_id) or profile.get('nickName') != profile_body['nickName']:
            raise RuntimeError('profile update was not scoped to current user')
        if admin_digest() != admin_before:
            raise RuntimeError('forged profile userId modified protected admin state')
        evidence.append('profile_current_user_scope=PASS')

        second_token = login(config, tracked_redis_keys, username, old_password)
        api_failure('/system/user/profile/updatePwd', first_token, 'PUT', {
            'oldPassword': new_password, 'newPassword': 'Wrong9' + secrets.token_hex(4),
        }, '旧密码不正确')
        api_success('/system/user/profile', second_token)
        api_success('/system/user/profile/updatePwd', first_token, 'PUT', {
            'oldPassword': old_password, 'newPassword': new_password,
        })
        for session_token in (first_token, second_token):
            if redis_command(config, 'EXISTS', 'login_tokens:' + token_id(session_token)) != '0':
                raise RuntimeError('password change did not invalidate every user session')
            status, payload, _ = json_request('/system/user/profile', session_token)
            if status != 401 or payload.get('code') != 401:
                raise RuntimeError('invalidated session still accessed the profile')
        current_token = login(config, tracked_redis_keys, username, new_password)
        evidence.append('password_wrong_correct_two_session_invalidation=PASS')

        status, avatar = upload_avatar(current_token, 'tiny.png', TINY_PNG)
        if status != 200 or avatar.get('code') not in (0, 200) or not avatar.get('imgUrl'):
            raise RuntimeError(f'valid avatar upload failed: status={status} msg={avatar.get("msg")}')
        image_url = avatar['imgUrl']
        if not image_url.startswith(config['profile_url'].rstrip('/') + '/avatar/'):
            raise RuntimeError('avatar URL is outside configured profile path')
        relative_url = image_url[len(config['profile_url'].rstrip('/')):].lstrip('/')
        avatar_path = Path(config['upload_path']).resolve() / config['profile_url'].strip('/') / relative_url
        avatar_path = avatar_path.resolve()
        profile_root = (Path(config['upload_path']).resolve() / config['profile_url'].strip('/')).resolve()
        if profile_root not in avatar_path.parents:
            raise RuntimeError('avatar file path escaped configured profile directory')
        public_status, public_bytes, _ = request(image_url, base=PUBLIC_BASE)
        if public_status != 200 or public_bytes != TINY_PNG:
            raise RuntimeError(f'anonymous avatar retrieval mismatch: status={public_status} bytes={len(public_bytes)}')
        fake_status, fake = upload_avatar(current_token, 'fake.png', b'not-a-real-image')
        if fake_status != 200 or fake.get('code') in (0, 200):
            raise RuntimeError('fake image upload was not rejected as an HTTP-200 business failure')
        if mysql(f"SELECT HEX(avatar) FROM sys_user WHERE user_id={user_id}") != image_url.encode().hex().upper():
            raise RuntimeError('fake image rejection changed the stored avatar')
        evidence.append('avatar_png_anonymous_bytes_and_fake_rejection=PASS')

        api_success('/system/role', admin, 'POST', {
            'roleName': 'E2E Role ' + suffix, 'roleKey': role_key, 'roleSort': 998, 'status': '0', 'menuIds': [],
        })
        role_row = mysql(f"SELECT role_id,COALESCE(del_flag,'') FROM sys_role WHERE role_key={sql_quote(role_key)}")
        if not role_row:
            raise RuntimeError('temporary role was not created')
        role_fields = role_row.split('\t')
        created_role_id = int(role_fields[0])
        if role_fields[1] != '0':
            mysql_raw(f"UPDATE sys_role SET del_flag='0' WHERE role_id={created_role_id}")

        api_success('/system/user/authRole?' + urllib.parse.urlencode({
            'userId': user_id, 'roleIds': created_role_id,
        }), admin, 'PUT')
        assigned = mysql(f'SELECT GROUP_CONCAT(role_id ORDER BY role_id) FROM sys_user_role WHERE user_id={user_id}')
        if assigned != str(created_role_id):
            raise RuntimeError('user-role replacement did not replace existing roles')

        allocated_path = '/system/role/authUser/allocatedList?' + urllib.parse.urlencode({
            'roleId': created_role_id, 'userName': username, 'pageNum': 1, 'pageSize': 20,
        })
        if str(user_id) not in row_ids(api_success(allocated_path, admin)):
            raise RuntimeError('allocated list omitted the assigned user')
        api_success('/system/role/authUser/cancel', admin, 'PUT', {'roleId': created_role_id, 'userId': user_id})
        unallocated_path = '/system/role/authUser/unallocatedList?' + urllib.parse.urlencode({
            'roleId': created_role_id, 'userName': username, 'pageNum': 1, 'pageSize': 20,
        })
        if str(user_id) not in row_ids(api_success(unallocated_path, admin)):
            raise RuntimeError('unallocated list omitted the cancelled user')

        batch_ids = f'{user_id},{helper_user_id}'
        api_success('/system/role/authUser/selectAll?' + urllib.parse.urlencode({
            'roleId': created_role_id, 'userIds': batch_ids,
        }), admin, 'PUT')
        if mysql(f'SELECT COUNT(*) FROM sys_user_role WHERE role_id={created_role_id} AND user_id IN ({user_id},{helper_user_id})') != '2':
            raise RuntimeError('batch role authorization did not create two relations')
        api_success('/system/role/authUser/cancelAll?' + urllib.parse.urlencode({
            'roleId': created_role_id, 'userIds': batch_ids,
        }), admin, 'PUT')
        if mysql(f'SELECT COUNT(*) FROM sys_user_role WHERE role_id={created_role_id} AND user_id IN ({user_id},{helper_user_id})') != '0':
            raise RuntimeError('batch role cancellation left relations')

        dept_id = mysql("SELECT dept_id FROM sys_dept WHERE status='0' AND del_flag='0' ORDER BY dept_id LIMIT 1")
        if dept_id:
            api_success('/system/role/dataScope', admin, 'PUT', {
                'roleId': created_role_id, 'dataScope': '2', 'deptIds': [int(dept_id)],
            })
            if mysql(f'SELECT CONCAT(data_scope,"|",(SELECT COUNT(*) FROM sys_role_dept WHERE role_id={created_role_id} AND dept_id={int(dept_id)})) FROM sys_role WHERE role_id={created_role_id}') != '2|1':
                raise RuntimeError('role custom data scope was not persisted')
        else:
            print('SKIP role_custom_data_scope no_active_department')
        api_success('/system/role/dataScope', admin, 'PUT', {
            'roleId': created_role_id, 'dataScope': '5', 'deptIds': [int(dept_id)] if dept_id else [],
        })
        if mysql(f'SELECT CONCAT(data_scope,"|",(SELECT COUNT(*) FROM sys_role_dept WHERE role_id={created_role_id})) FROM sys_role WHERE role_id={created_role_id}') != '5|0':
            raise RuntimeError('role self-only data scope did not clear department relations')
        api_success('/system/role/changeStatus', admin, 'PUT', {'roleId': created_role_id, 'status': '1'})
        if mysql(f'SELECT status FROM sys_role WHERE role_id={created_role_id}') != '1':
            raise RuntimeError('role disable status was not persisted')
        api_success('/system/role/changeStatus', admin, 'PUT', {'roleId': created_role_id, 'status': '0'})
        evidence.append('role_replace_lists_single_batch_status_scope=PASS')

        config_sentinel = 'sys_config:e2e-refresh-' + suffix
        login_sentinel = 'login_tokens:e2e-refresh-' + suffix
        captcha_sentinel = 'captcha_codes:e2e-refresh-' + suffix
        tracked_redis_keys.update((config_sentinel, login_sentinel, captcha_sentinel))
        for key in (config_sentinel, login_sentinel, captcha_sentinel):
            redis_set(config, key, 'sentinel', 600)
        api_success('/system/config/refreshCache', admin, 'DELETE')
        if redis_command(config, 'EXISTS', config_sentinel) != '0':
            raise RuntimeError('config refresh left a sys_config key')
        if redis_command(config, 'EXISTS', login_sentinel) != '1' or redis_command(config, 'EXISTS', captcha_sentinel) != '1':
            raise RuntimeError('config refresh deleted non-config cache keys')
        evidence.append('config_cache_miss_hit_exact_invalidation_refresh=PASS')

        api_success('/system/dict/type', admin, 'POST', {
            'dictName': 'E2E Dict ' + suffix, 'dictType': dict_type, 'status': '0', 'remark': 'temporary e2e',
        })
        dict_id_text = mysql(f"SELECT dict_id FROM sys_dict_type WHERE dict_type={sql_quote(dict_type)}")
        if not dict_id_text:
            raise RuntimeError('temporary dict type was not created')
        dict_id = int(dict_id_text)
        api_success('/system/dict/data', admin, 'POST', {
            'dictSort': 1, 'dictLabel': 'before', 'dictValue': 'v1', 'dictType': dict_type,
            'status': '0', 'remark': 'temporary e2e',
        })
        dict_code = int(mysql(f"SELECT dict_code FROM sys_dict_data WHERE dict_type={sql_quote(dict_type)} ORDER BY dict_code DESC LIMIT 1"))
        dict_cache_key = 'sys_dict:' + dict_type
        tracked_redis_keys.add(dict_cache_key)
        first_dict = api_success('/system/dict/data/type/' + urllib.parse.quote(dict_type, safe='')).get('data') or []
        if len(first_dict) != 1 or first_dict[0].get('dictLabel') != 'before' or redis_command(config, 'EXISTS', dict_cache_key) != '1':
            raise RuntimeError('dict cache miss/read-through mismatch')
        cached_row = dict(first_dict[0])
        cached_row['dictLabel'] = 'cached-hit'
        redis_set(config, dict_cache_key, json.dumps([cached_row], ensure_ascii=False, separators=(',', ':')), 600)
        cache_hit = api_success('/system/dict/data/type/' + urllib.parse.quote(dict_type, safe='')).get('data') or []
        if len(cache_hit) != 1 or cache_hit[0].get('dictLabel') != 'cached-hit':
            raise RuntimeError('dict cache hit did not return cached data')
        api_success('/system/dict/data', admin, 'PUT', {
            'dictCode': dict_code, 'dictSort': 1, 'dictLabel': 'after', 'dictValue': 'v1',
            'dictType': dict_type, 'status': '0', 'remark': 'temporary e2e',
        })
        if redis_command(config, 'EXISTS', dict_cache_key) != '0':
            raise RuntimeError('dict edit did not invalidate exact cache key')
        after_dict = api_success('/system/dict/data/type/' + urllib.parse.quote(dict_type, safe='')).get('data') or []
        if len(after_dict) != 1 or after_dict[0].get('dictLabel') != 'after':
            raise RuntimeError('dict refresh after exact invalidation returned stale data')
        unrelated_dict_key = 'sys_dict:e2e-unrelated-' + suffix
        tracked_redis_keys.add(unrelated_dict_key)
        redis_set(config, unrelated_dict_key, 'sentinel', 600)
        api_success('/system/dict/type/refreshCache', admin, 'DELETE')
        if redis_command(config, 'EXISTS', dict_cache_key) != '0' or redis_command(config, 'EXISTS', unrelated_dict_key) != '0':
            raise RuntimeError('dict refresh left sys_dict keys')
        if redis_command(config, 'EXISTS', login_sentinel) != '1' or redis_command(config, 'EXISTS', captcha_sentinel) != '1':
            raise RuntimeError('dict refresh deleted non-dict cache keys')
        evidence.append('dict_cache_miss_hit_exact_invalidation_refresh=PASS')

        retired_id_list = ','.join(map(str, RETIRED_MENU_IDS))
        retired_rows = mysql(
            "SELECT menu_id,HEX(COALESCE(menu_name,'')),HEX(COALESCE(path,'')),HEX(COALESCE(component,'')),status,visible "
            f'FROM sys_menu WHERE menu_id IN ({retired_id_list}) ORDER BY menu_id'
        ).splitlines()
        if len(retired_rows) != len(RETIRED_MENU_IDS):
            raise RuntimeError('retired menu migration target rows are missing')
        retired_markers = []
        for row in retired_rows:
            fields = row.split('\t')
            if fields[4:] != ['1', '1']:
                raise RuntimeError(f'retired menu {fields[0]} is not disabled and hidden')
            retired_markers.append((hex_decode(fields[1]), hex_decode(fields[2]), hex_decode(fields[3])))
        routers = api_success('/getRouters', admin).get('data') or []
        router_fields = flatten_router_fields(routers)
        for menu_name, path, component in retired_markers:
            if menu_name and menu_name.lower() in router_fields['title']:
                raise RuntimeError(f'getRouters exposed retired menu title: {menu_name}')
            if path and path.lower() in router_fields['path']:
                raise RuntimeError(f'getRouters exposed retired path: {path}')
            if component and component.lower() in router_fields['component']:
                raise RuntimeError(f'getRouters exposed retired component: {component}')

        retired_requests = [
            ('GET', '/monitor/online/list'), ('GET', '/monitor/job/list'),
            ('GET', '/monitor/jobLog/list'), ('GET', '/monitor/cache'), ('GET', '/tool/gen/list'),
            ('POST', '/exam/api/user/repo/paging'), ('POST', '/exam/api/user/wrong-book/list'),
            ('DELETE', '/monitor/operlog/1'), ('DELETE', '/monitor/operlog/clean'),
            ('DELETE', '/monitor/logininfor/1'), ('DELETE', '/monitor/logininfor/clean'),
        ]
        for method, path in retired_requests:
            assert_http_404(path, admin, method)
        for path in ('/monitor/operlog/list?pageNum=1&pageSize=1', '/monitor/logininfor/list?pageNum=1&pageSize=1'):
            payload = api_success(path, admin)
            if not isinstance(payload.get('rows'), list) or not isinstance(payload.get('total'), (int, float)):
                raise RuntimeError(f'{path} did not return a RuoYi table response')
        evidence.append('retired_routes_404_routers_clean_audit_lists=PASS')

        if admin_digest() != admin_before:
            raise RuntimeError('protected admin digest changed during administration E2E')
        evidence.append('protected_admin_digest=PASS')

        print('STAGING_ADMINISTRATION_E2E_PASS')
        for line in evidence:
            print(line)
    finally:
        try:
            if register_snapshot is not None:
                try:
                    edit_config(admin, register_snapshot, register_snapshot['configValue'])
                except Exception:
                    pass
                mysql_raw(
                    "UPDATE sys_config SET "
                    f"config_name={sql_quote(register_snapshot['configName'])},"
                    f"config_key={sql_quote(register_snapshot['configKey'])},"
                    f"config_value={sql_quote(register_snapshot['configValue'])},"
                    f"config_type={sql_quote(register_snapshot['configType'])},"
                    f"remark={sql_quote(register_snapshot['remark'])} "
                    f"WHERE config_id={register_snapshot['configId']}"
                )
        except Exception as error:
            cleanup_errors.append('register_config_restore=' + str(error))

        try:
            if dict_code is not None:
                mysql_raw(f'DELETE FROM sys_dict_data WHERE dict_code={int(dict_code)}')
            if dict_id is not None:
                mysql_raw(f'DELETE FROM sys_dict_type WHERE dict_id={int(dict_id)}')
        except Exception as error:
            cleanup_errors.append('dict_cleanup=' + str(error))

        try:
            if created_role_id is not None:
                mysql_raw(
                    'START TRANSACTION;'
                    f'DELETE FROM sys_user_role WHERE role_id={int(created_role_id)};'
                    f'DELETE FROM sys_role_dept WHERE role_id={int(created_role_id)};'
                    f'DELETE FROM sys_role_menu WHERE role_id={int(created_role_id)};'
                    f'DELETE FROM sys_role WHERE role_id={int(created_role_id)};'
                    'COMMIT;'
                )
        except Exception as error:
            cleanup_errors.append('role_cleanup=' + str(error))

        try:
            if created_user_ids:
                ids = ','.join(str(int(value)) for value in created_user_ids)
                mysql_raw(
                    'START TRANSACTION;'
                    f'DELETE FROM sys_user_role WHERE user_id IN ({ids});'
                    f'DELETE FROM sys_user_post WHERE user_id IN ({ids});'
                    f'DELETE FROM sys_user WHERE user_id IN ({ids});'
                    'COMMIT;'
                )
        except Exception as error:
            cleanup_errors.append('user_cleanup=' + str(error))

        try:
            if avatar_path is not None and avatar_path.exists():
                avatar_path.unlink()
        except Exception as error:
            cleanup_errors.append('avatar_cleanup=' + str(error))

        for key in sorted(tracked_redis_keys):
            redis_command(config, 'DEL', key, capture=False, check=False)
        for username_value in (username, helper_username):
            redis_command(config, 'DEL', 'login_fail:' + username_value, capture=False, check=False)
        if register_snapshot is not None:
            restore_cache_snapshot(config, 'sys_config:' + REGISTER_KEY, register_cache)

        try:
            user_remaining = '0'
            if created_user_ids:
                ids = ','.join(str(int(value)) for value in created_user_ids)
                user_remaining = mysql(f'SELECT COUNT(*) FROM sys_user WHERE user_id IN ({ids})')
            role_remaining = '0'
            relation_remaining = '0'
            if created_role_id is not None:
                role_remaining = mysql(f'SELECT COUNT(*) FROM sys_role WHERE role_id={int(created_role_id)}')
                relation_remaining = mysql(
                    'SELECT ('
                    f'(SELECT COUNT(*) FROM sys_user_role WHERE role_id={int(created_role_id)})+'
                    f'(SELECT COUNT(*) FROM sys_role_dept WHERE role_id={int(created_role_id)})+'
                    f'(SELECT COUNT(*) FROM sys_role_menu WHERE role_id={int(created_role_id)}))'
                )
            dict_remaining = '0'
            if dict_id is not None or dict_code is not None:
                dict_remaining = mysql(
                    'SELECT ('
                    f'(SELECT COUNT(*) FROM sys_dict_type WHERE dict_id={int(dict_id or 0)})+'
                    f'(SELECT COUNT(*) FROM sys_dict_data WHERE dict_code={int(dict_code or 0)}))'
                )
            if user_remaining != '0' or role_remaining != '0' or relation_remaining != '0' or dict_remaining != '0':
                cleanup_errors.append(
                    f'database_remaining users={user_remaining} role={role_remaining} relations={relation_remaining} dict={dict_remaining}'
                )
            if avatar_path is not None and avatar_path.exists():
                cleanup_errors.append('avatar_file_remaining=1')
            if register_snapshot is not None and register_digest is not None:
                if config_row_digest(register_snapshot['configId']) != register_digest:
                    cleanup_errors.append('registration_config_digest_changed')
            if admin_digest() != admin_before:
                cleanup_errors.append('protected_admin_digest_changed')
        except Exception as error:
            cleanup_errors.append('cleanup_verification=' + str(error))

        if cleanup_errors:
            print('cleanup=FAIL')
            for error in cleanup_errors:
                print('cleanup_error=' + error)
            raise RuntimeError('administration E2E cleanup failed')
        print('cleanup=PASS|users=0|role=0|relations=0|dict=0|avatar=0|sessions=0')
        print('register_config_restored_exactly=true')
        print('protected_admin_digest_unchanged=true')


if __name__ == '__main__':
    main()
