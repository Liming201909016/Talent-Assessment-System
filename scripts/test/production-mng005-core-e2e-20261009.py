#!/usr/bin/env python3
"""Authorized production-host MNG00501 core E2E with exact rollback.

Run on the application host as a user that can read the service configuration,
invoke mysql/redis-cli, and delete only this run's private reissue PDF.
Stdout is reserved for one non-secret JSON receipt.
"""

import base64
import hashlib
import hmac
import json
import os
import re
import shlex
import stat
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid
from pathlib import Path

APP_DIR = Path("/opt/talent-assessment")
SERVICE = "talent-assessment"
EXAM_ID = "9051103000000000501"
REPO_CODE = "00501"
HTTP_TIMEOUT = 30
REPORT_TIMEOUT = 150
ID_RE = re.compile(r"^[A-Za-z0-9_-]{1,64}$")
SHA_RE = re.compile(r"^[a-f0-9]{64}$")


def fail(message):
    raise RuntimeError(message)


def run_checked(argv, *, env=None, input_text=None, timeout=30):
    try:
        result = subprocess.run(
            argv,
            input=input_text,
            text=True,
            capture_output=True,
            timeout=timeout,
            env=env,
            check=False,
        )
    except (OSError, subprocess.SubprocessError) as exc:
        raise RuntimeError("required system command failed") from exc
    if result.returncode != 0:
        raise RuntimeError("required system command failed")
    return result.stdout


def systemctl_value(prop):
    return run_checked(
        ["systemctl", "show", SERVICE, "--property=" + prop, "--value"],
        timeout=15,
    ).strip()


def parse_env_assignments(text):
    values = {}
    try:
        words = shlex.split(text, comments=False, posix=True)
    except ValueError as exc:
        raise RuntimeError("invalid systemd environment syntax") from exc
    for word in words:
        if "=" not in word:
            continue
        key, value = word.split("=", 1)
        if re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", key):
            values[key] = value
    return values


def parse_environment_file(path):
    values = {}
    for raw in path.read_text(encoding="utf-8-sig").replace("\r", "").splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        if line.startswith("export "):
            line = line[7:].lstrip()
        if "=" not in line:
            continue
        key, raw_value = line.split("=", 1)
        key = key.strip()
        if not re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", key):
            continue
        try:
            parts = shlex.split(raw_value, comments=True, posix=True)
        except ValueError as exc:
            raise RuntimeError("invalid service environment file") from exc
        values[key] = "" if not parts else parts[0]
    return values


def service_environment():
    values = {}
    raw_files = systemctl_value("EnvironmentFiles")
    for match in re.finditer(r"(?:^|\s)(/[^\s;]+)", raw_files):
        path = Path(match.group(1))
        if path.is_file():
            values.update(parse_environment_file(path))
    values.update(parse_env_assignments(systemctl_value("Environment")))

    pid = systemctl_value("MainPID")
    if not pid.isdigit() or int(pid) <= 0:
        fail("application service is not running")
    proc_path = Path("/proc") / pid / "environ"
    try:
        raw = proc_path.read_bytes()
    except OSError:
        raw = subprocess.run(
            ["sudo", "-n", "cat", str(proc_path)],
            capture_output=True,
            check=True,
        ).stdout
    for item in raw.split(b"\0"):
        if b"=" not in item:
            continue
        key_raw, value_raw = item.split(b"=", 1)
        try:
            key = key_raw.decode("ascii")
            value = value_raw.decode("utf-8")
        except UnicodeDecodeError:
            continue
        if re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", key):
            values[key] = value
    return values


def yaml_scalar(text, section, key, default=""):
    value = default
    pattern = rf"(?ms)^{re.escape(section)}:\s*\n(.*?)(?=^[A-Za-z][\w-]*:\s*$|\Z)"
    for section_match in re.finditer(pattern, text):
        key_match = re.search(
            rf"(?m)^\s+{re.escape(key)}:\s*(.+?)\s*$", section_match.group(1)
        )
        if key_match:
            value = key_match.group(1).strip().strip("\"'")
    return value


def load_settings():
    env = service_environment()
    app_env = env.get("APP_ENV", "").strip().lower()
    if app_env != "production":
        fail("APP_ENV is not production")
    for key in ("REPORT_EFFECTIVE_ENV", "MNG_TEST_REPORT_ENV"):
        if env.get(key, "").strip().lower() != "production":
            fail(key + " is not production")

    texts = []
    for path in (
        APP_DIR / "configs" / "application.yml",
        APP_DIR / "configs" / "application-production.yml",
    ):
        if path.is_file():
            texts.append(path.read_text(encoding="utf-8-sig").replace("\r", ""))
    if not texts:
        fail("application configuration is missing")
    text = "\n".join(texts)

    settings = {
        "dsn": env.get("MYSQL_DSN") or yaml_scalar(text, "database", "dsn") or yaml_scalar(text, "mysql", "dsn"),
        "redis_addr": env.get("REDIS_ADDR") or yaml_scalar(text, "redis", "addr", "127.0.0.1:6379"),
        "redis_db": env.get("REDIS_DB") or yaml_scalar(text, "redis", "db", "1"),
        "redis_password": env.get("REDIS_PASSWORD") or yaml_scalar(text, "redis", "password"),
        "jwt_secret": env.get("JWT_SECRET") or yaml_scalar(text, "jwt", "secret"),
        "login_user_key": env.get("JWT_LOGIN_USER_KEY") or yaml_scalar(text, "jwt", "loginUserKey", "login_user_key"),
        "server_port": env.get("SERVER_PORT") or yaml_scalar(text, "server", "port", "8092"),
        "report_root": env.get("MNG_TEST_REPORT_DIR", ""),
    }
    if not settings["dsn"] or not settings["jwt_secret"]:
        fail("required application credentials are not configured")
    if not re.fullmatch(r"\d{1,5}", settings["server_port"]):
        fail("invalid application port")
    port = int(settings["server_port"])
    if not 1 <= port <= 65535:
        fail("invalid application port")
    if not re.fullmatch(r"\d+", settings["redis_db"]):
        fail("invalid Redis database")
    if not re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", settings["login_user_key"]):
        fail("invalid JWT login user key")
    root = Path(settings["report_root"])
    if not root.is_absolute():
        fail("management report root is not absolute")
    settings["app_env"] = app_env
    settings["base_url"] = f"http://127.0.0.1:{port}"
    return settings


def mysql_client_from_dsn(dsn):
    match = re.fullmatch(r"([^:]+):(.*)@tcp\(([^)]+)\)/([^?]+)(?:\?.*)?", dsn)
    if not match:
        fail("application MySQL DSN format is unsupported")
    user, password, address, database = match.groups()
    if database != "element" or any("\n" in value or "\r" in value for value in match.groups()):
        fail("application MySQL DSN is outside the authorized database")
    if address.startswith("["):
        host_match = re.fullmatch(r"\[([^]]+)\]:(\d+)", address)
    else:
        host_match = re.fullmatch(r"(.+):(\d+)", address)
    if not host_match:
        fail("application MySQL address format is unsupported")
    host, port = host_match.groups()
    if not user or not host or not port.isdigit():
        fail("application MySQL DSN is incomplete")
    fd, name = tempfile.mkstemp(prefix="mng005-mysql-", suffix=".cnf", dir="/tmp")
    try:
        os.fchmod(fd, 0o600)
        escape = lambda value: value.replace("\\", "\\\\").replace('"', '\\"')
        content = '[client]\nuser="{0}"\npassword="{1}"\nhost="{2}"\nport={3}\n'.format(
            escape(user), escape(password), escape(host), port
        )
        os.write(fd, content.encode("utf-8"))
    finally:
        os.close(fd)
    if stat.S_IMODE(os.stat(name).st_mode) != 0o600:
        Path(name).unlink(missing_ok=True)
        fail("temporary MySQL client file mode is not 0600")
    return Path(name), database


class MySQL:
    def __init__(self, client_file, database):
        self.client_file = client_file
        self.database = database

    def _argv(self):
        return [
            "mysql",
            "--defaults-extra-file=" + str(self.client_file),
            "--batch",
            "--skip-column-names",
            "--raw",
            self.database,
        ]

    def query(self, sql):
        return run_checked(self._argv() + ["--execute", sql], timeout=60).strip()

    def execute(self, sql):
        run_checked(self._argv(), input_text=sql, timeout=90)

    def scalar_int(self, sql):
        value = self.query(sql)
        if not re.fullmatch(r"\d+", value):
            fail("database count query returned an invalid value")
        return int(value)


def sql_literal(value):
    if not ID_RE.fullmatch(value):
        fail("unsafe owned identifier")
    return "'" + value + "'"


def sql_in(values):
    if not values:
        return "NULL"
    return ",".join(sql_literal(value) for value in values)


def select_ids(db, sql):
    raw = db.query(sql)
    if not raw:
        return []
    values = raw.splitlines()
    if len(values) != len(set(values)) or any(not ID_RE.fullmatch(value) for value in values):
        fail("owned identifier query returned invalid data")
    return values


def b64url(raw):
    return base64.urlsafe_b64encode(raw).rstrip(b"=").decode("ascii")


def make_admin_token(secret, claim_key, token_id):
    header = b64url(json.dumps({"alg": "HS512", "typ": "JWT"}, separators=(",", ":")).encode())
    payload = b64url(json.dumps({claim_key: token_id}, separators=(",", ":")).encode())
    signing_input = header + "." + payload
    signature = hmac.new(secret.encode(), signing_input.encode(), hashlib.sha512).digest()
    return signing_input + "." + b64url(signature)


def split_host_port(address):
    match = re.fullmatch(r"\[([^]]+)\]:(\d+)|([^:]+):(\d+)", address)
    if not match:
        fail("unsupported Redis address")
    host = match.group(1) or match.group(3)
    port = match.group(2) or match.group(4)
    return host, port


def redis_call(settings, *arguments, expect=None):
    host, port = split_host_port(settings["redis_addr"])
    env = os.environ.copy()
    password = settings["redis_password"]
    if password:
        env["REDISCLI_AUTH"] = password
    else:
        env.pop("REDISCLI_AUTH", None)
    stdout = run_checked(
        ["redis-cli", "--raw", "-h", host, "-p", port, "-n", settings["redis_db"], *arguments],
        env=env,
        timeout=20,
    ).strip()
    if expect is not None and stdout != expect:
        fail("Redis command returned an unexpected result")
    return stdout


def request(base_url, path, *, token=None, participant_token=None, body=None, method="POST", binary=False, timeout=HTTP_TIMEOUT):
    headers = {"Accept": "application/pdf" if binary else "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    if participant_token:
        headers["X-Management-Traits-Token"] = participant_token
    payload = None
    if body is not None:
        headers["Content-Type"] = "application/json"
        payload = json.dumps(body, ensure_ascii=False, separators=(",", ":")).encode("utf-8")
    req = urllib.request.Request(base_url + path, data=payload, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as response:
            raw = response.read()
            status = response.status
            response_headers = dict(response.headers)
    except urllib.error.HTTPError as exc:
        raw = exc.read()
        status = exc.code
        response_headers = dict(exc.headers)
    except urllib.error.URLError as exc:
        raise RuntimeError("application HTTP request failed") from exc
    if binary:
        if status != 200:
            fail("PDF endpoint returned a non-success status")
        return raw, response_headers
    try:
        document = json.loads(raw.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise RuntimeError("application endpoint returned invalid JSON") from exc
    if status != 200 or document.get("code") not in (0, 200) or document.get("success") is False:
        fail("application endpoint rejected the E2E operation")
    return document.get("data") if document.get("data") is not None else {}


def baseline_counts(db):
    queries = {
        "candidate": f"SELECT COUNT(*) FROM el_candidate WHERE exam_id='{EXAM_ID}'",
        "paper": f"SELECT COUNT(*) FROM el_paper WHERE exam_id='{EXAM_ID}'",
        "paperQuestion": f"SELECT COUNT(*) FROM el_paper_qu q JOIN el_paper p ON p.id=q.paper_id WHERE p.exam_id='{EXAM_ID}'",
        "paperAnswer": f"SELECT COUNT(*) FROM el_paper_qu_answer a JOIN el_paper p ON p.id=a.paper_id WHERE p.exam_id='{EXAM_ID}'",
        "snapshot": f"SELECT COUNT(*) FROM el_mng_paper_snapshot WHERE exam_id='{EXAM_ID}'",
        "questionSnapshot": f"SELECT COUNT(*) FROM el_mng_paper_question_snapshot q JOIN el_mng_paper_snapshot s ON s.paper_id=q.paper_id WHERE s.exam_id='{EXAM_ID}'",
        "run": f"SELECT COUNT(*) FROM el_mng_result_run WHERE exam_id='{EXAM_ID}'",
        "dimension": f"SELECT COUNT(*) FROM el_mng_result_dimension d JOIN el_mng_result_run r ON r.id=d.run_id WHERE r.exam_id='{EXAM_ID}'",
        "module": f"SELECT COUNT(*) FROM el_mng_result_module m JOIN el_mng_result_run r ON r.id=m.run_id WHERE r.exam_id='{EXAM_ID}'",
        "receipt": f"SELECT COUNT(*) FROM el_mng_runtime_receipt WHERE exam_id='{EXAM_ID}'",
        "revision": f"SELECT COUNT(*) FROM el_mng_report_revision WHERE exam_id='{EXAM_ID}'",
        "current": f"SELECT COUNT(*) FROM el_mng_report_current c JOIN el_paper p ON p.id=c.paper_id WHERE p.exam_id='{EXAM_ID}'",
        "reportAudit": f"SELECT COUNT(*) FROM el_mng_report_audit a JOIN el_mng_report_revision r ON r.id=a.report_id WHERE r.exam_id='{EXAM_ID}'",
        "reissue": f"SELECT COUNT(*) FROM el_mng_report_reissue WHERE exam_id='{EXAM_ID}'",
        "reissueAudit": f"SELECT COUNT(*) FROM el_mng_reissue_audit WHERE exam_id='{EXAM_ID}'",
    }
    return {name: db.scalar_int(sql) for name, sql in queries.items()}


def require_id(value, label):
    value = str(value or "")
    if not ID_RE.fullmatch(value):
        fail(label + " is invalid")
    return value


def make_identity(required_fields, marker, phone):
    values = {
        "name": marker,
        "gender": "0",
        "telephone": phone,
        "affiliation": "PROD-E2E",
        "post": "Verifier",
        "age": 30,
        "stuFlag": 0,
        "degree": "Synthetic",
        "major": "E2E",
    }
    supported = set(values)
    if not isinstance(required_fields, list) or not required_fields or any(field not in supported for field in required_fields):
        fail("00501 field contract is unsupported")
    body = {"examId": EXAM_ID}
    for field in required_fields:
        body[field] = values[field]
    return body


def result_cardinality(detail):
    result = detail.get("Result") or detail.get("result") or {}
    dimensions = result.get("Dimensions") or result.get("dimensions") or []
    modules = result.get("Modules") or result.get("modules") or []
    answered = result.get("AnsweredQuestionCount", result.get("answeredQuestionCount"))
    total = result.get("TotalQuestionCount", result.get("totalQuestionCount"))
    if answered != 140 or total != 140 or len(dimensions) != 13 or len(modules) != 4:
        fail("management result detail cardinality mismatch")
    return len(dimensions), len(modules)


def readonly_smoke(db, settings, admin_token):
    details = {}
    for code in ("00101", "00201", "00301"):
        exam_id = db.query(
            "SELECT e.id FROM el_exam e "
            "JOIN el_exam_repo er ON er.exam_id=e.id "
            "JOIN el_repo r ON r.id=er.repo_id "
            f"WHERE r.code='{code}' ORDER BY e.id LIMIT 1"
        )
        require_id(exam_id, code + " exam")
        detail = request(
            settings["base_url"],
            "/exam/api/exam/exam/detail",
            token=admin_token,
            body={"id": exam_id},
        )
        if detail.get("repoCode") != code:
            fail(code + " detail smoke returned another repository")
        details[code] = True

    table_count = db.scalar_int(
        "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() "
        "AND table_name IN ('el_competency_dimension','el_qu')"
    )
    dimension_count = db.scalar_int("SELECT COUNT(*) FROM el_competency_dimension")
    question_count = db.scalar_int(
        "SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NOT NULL "
        "OR competency_question_type IS NOT NULL"
    )
    if table_count != 2 or dimension_count < 1 or question_count < 1:
        fail("00401 virtual competency backing data is absent")
    return {
        "legacyDetail": details,
        "competency00401": {"dimensions": dimension_count, "questions": question_count},
    }


def owned_closure(db, marker, phone, tracked):
    marker_sql = marker.replace("'", "''")
    phone_sql = phone.replace("'", "''")
    candidate_ids = select_ids(
        db,
        "SELECT id FROM el_candidate "
        f"WHERE exam_id='{EXAM_ID}' AND name='{marker_sql}' AND telephone='{phone_sql}' ORDER BY id",
    )
    if tracked.get("candidate_id") and tracked["candidate_id"] not in candidate_ids:
        fail("tracked candidate is outside the exact owned marker")
    if len(candidate_ids) > 1:
        fail("owned marker matched multiple candidates")
    paper_ids = []
    if candidate_ids:
        paper_ids = select_ids(
            db,
            "SELECT id FROM el_paper "
            f"WHERE exam_id='{EXAM_ID}' AND user_id IN ({sql_in(candidate_ids)}) ORDER BY id",
        )
    if tracked.get("paper_id") and tracked["paper_id"] not in paper_ids:
        fail("tracked paper is outside the exact owned candidate")
    if len(paper_ids) > 1:
        fail("owned candidate matched multiple papers")
    run_ids = select_ids(
        db,
        f"SELECT id FROM el_mng_result_run WHERE exam_id='{EXAM_ID}' AND paper_id IN ({sql_in(paper_ids)}) ORDER BY id",
    ) if paper_ids else []
    report_ids = select_ids(
        db,
        f"SELECT id FROM el_mng_report_revision WHERE exam_id='{EXAM_ID}' AND run_id IN ({sql_in(run_ids)}) ORDER BY id",
    ) if run_ids else []
    reissue_ids = select_ids(
        db,
        f"SELECT id FROM el_mng_report_reissue WHERE exam_id='{EXAM_ID}' AND run_id IN ({sql_in(run_ids)}) ORDER BY id",
    ) if run_ids else []
    if tracked.get("run_id") and tracked["run_id"] not in run_ids:
        fail("tracked run is outside the exact owned paper")
    if tracked.get("reissue_id") and tracked["reissue_id"] not in reissue_ids:
        fail("tracked reissue is outside the exact owned run")
    return {
        "candidate": candidate_ids,
        "paper": paper_ids,
        "run": run_ids,
        "revision": report_ids,
        "reissue": reissue_ids,
    }


def report_files(db, settings, reissue_ids):
    files = []
    root = Path(settings["report_root"]).resolve()
    reissue_root = (root / "reissues").resolve()
    for report_id in reissue_ids:
        row = db.query(
            "SELECT CONCAT(file_key,'|',file_sha,'|',file_bytes) "
            f"FROM el_mng_report_reissue WHERE id={sql_literal(report_id)}"
        )
        parts = row.split("|")
        if len(parts) != 3 or parts[0] != report_id + ".pdf" or not SHA_RE.fullmatch(parts[1]) or not parts[2].isdigit():
            fail("owned report file metadata is invalid")
        path = (reissue_root / parts[0]).resolve()
        if path.parent != reissue_root or path.is_symlink() or not path.is_file():
            fail("owned report file path is unsafe")
        raw = path.read_bytes()
        if len(raw) != int(parts[2]) or hashlib.sha256(raw).hexdigest() != parts[1]:
            fail("owned report file bytes do not match database metadata")
        files.append((path, parts[1], len(raw)))
    return files


def exact_cleanup(db, settings, marker, phone, tracked):
    closure = owned_closure(db, marker, phone, tracked)
    files = report_files(db, settings, closure["reissue"])
    papers, runs = closure["paper"], closure["run"]
    revisions, reissues = closure["revision"], closure["reissue"]

    reissue_audits = select_ids(db, f"SELECT id FROM el_mng_reissue_audit WHERE report_id IN ({sql_in(reissues)}) ORDER BY id") if reissues else []
    report_audits = select_ids(db, f"SELECT id FROM el_mng_report_audit WHERE report_id IN ({sql_in(revisions)}) ORDER BY id") if revisions else []
    dimensions = select_ids(db, f"SELECT id FROM el_mng_result_dimension WHERE run_id IN ({sql_in(runs)}) ORDER BY id") if runs else []
    modules = select_ids(db, f"SELECT id FROM el_mng_result_module WHERE run_id IN ({sql_in(runs)}) ORDER BY id") if runs else []
    snapshots = select_ids(db, f"SELECT id FROM el_mng_paper_question_snapshot WHERE paper_id IN ({sql_in(papers)}) ORDER BY id") if papers else []
    paper_answers = select_ids(db, f"SELECT id FROM el_paper_qu_answer WHERE paper_id IN ({sql_in(papers)}) ORDER BY id") if papers else []
    paper_questions = select_ids(db, f"SELECT id FROM el_paper_qu WHERE paper_id IN ({sql_in(papers)}) ORDER BY id") if papers else []

    delete_plan = [
        ("el_mng_reissue_audit", reissue_audits),
        ("el_mng_report_reissue", reissues),
        ("el_mng_report_audit", report_audits),
        ("el_mng_report_current", papers),
        ("el_mng_report_revision", revisions),
        ("el_mng_runtime_receipt", runs),
        ("el_mng_result_dimension", dimensions),
        ("el_mng_result_module", modules),
        ("el_mng_result_run", runs),
        ("el_mng_paper_question_snapshot", snapshots),
        ("el_mng_paper_snapshot", papers),
        ("el_paper_qu_answer", paper_answers),
        ("el_paper_qu", paper_questions),
        ("el_candidate", closure["candidate"]),
        ("el_paper", papers),
    ]
    statements = [
        "SET NAMES utf8mb4",
        "SET SESSION SQL_SAFE_UPDATES=1",
        "START TRANSACTION",
        f"SELECT id FROM el_exam WHERE id='{EXAM_ID}' AND state=1 FOR UPDATE",
    ]
    for table, ids in delete_plan:
        if not re.fullmatch(r"el_[a-z_]+", table):
            fail("unsafe cleanup table")
        if ids:
            key = "paper_id" if table in ("el_mng_report_current", "el_mng_paper_snapshot") else "run_id" if table == "el_mng_runtime_receipt" else "id"
            statements.append(f"DELETE FROM {table} WHERE {key} IN ({sql_in(ids)})")
    statements.extend(["COMMIT", "SELECT 'MNG005_EXACT_CLEANUP_OK'"])
    db.execute(";\n".join(statements) + ";\n")

    for path, expected_sha, expected_size in files:
        if path.is_symlink() or not path.is_file():
            fail("owned report file disappeared before cleanup")
        raw = path.read_bytes()
        if len(raw) != expected_size or hashlib.sha256(raw).hexdigest() != expected_sha:
            fail("owned report file changed before cleanup")
        path.unlink()
    summary = {name: len(values) for name, values in closure.items()}
    summary["files"] = len(files)
    return summary


def main():
    receipt = {
        "schema": "production-mng005-core-e2e-receipt-v1",
        "status": "failed",
        "environment": "production",
        "exam": REPO_CODE,
        "authorized": True,
    }
    settings = None
    db = None
    mysql_file = None
    redis_key = ""
    admin_token = ""
    baseline = None
    marker = "PROD-MNG005-E2E-" + uuid.uuid4().hex[:16]
    phone = "199" + str(int(time.time() * 1000))[-8:]
    tracked = {}
    state_verified = False
    cleanup_errors = []
    started_at = time.time()
    try:
        settings = load_settings()
        mysql_file, database = mysql_client_from_dsn(settings["dsn"])
        db = MySQL(mysql_file, database)
        if db.query("SELECT DATABASE()") != "element":
            fail("MySQL client is not using the authorized database")
        if db.scalar_int("SELECT COUNT(*) FROM el_exam WHERE id='9051103000000000501' AND state=1") != 1:
            fail("00501 initial exam state is not 1")
        if db.scalar_int(f"SELECT COUNT(*) FROM el_candidate WHERE name='{marker}' OR telephone='{phone}'") != 0:
            fail("synthetic ownership marker already exists")
        baseline = baseline_counts(db)
        receipt["baselineBefore"] = baseline
        state_verified = True

        token_id = "prod-mng005-e2e-" + uuid.uuid4().hex
        redis_key = "login_tokens:" + token_id
        if redis_call(settings, "EXISTS", redis_key, expect="0") != "0":
            fail("synthetic Redis key already exists")
        now_ms = int(time.time() * 1000)
        login_user = {
            "userId": 1,
            "token": token_id,
            "loginTime": now_ms,
            "expireTime": now_ms + 1800000,
            "permissions": ["*:*:*"],
            "roles": ["admin"],
            "user": None,
        }
        redis_call(
            settings,
            "SET",
            redis_key,
            json.dumps(login_user, separators=(",", ":")),
            "NX",
            "EX",
            "1800",
            expect="OK",
        )
        admin_token = make_admin_token(settings["jwt_secret"], settings["login_user_key"], token_id)

        smoke = readonly_smoke(db, settings, admin_token)
        receipt["readOnlySmoke"] = smoke
        profile = request(
            settings["base_url"],
            "/exam/api/management-traits/profile/detail?examId=" + EXAM_ID,
            token=admin_token,
            method="GET",
            body=None,
        )
        field_contract = json.loads(profile.get("fieldContract", ""))
        candidate_body = make_identity(field_contract.get("requiredFields"), marker, phone)

        state = request(
            settings["base_url"],
            "/exam/api/management-traits/admin/exam/state",
            token=admin_token,
            body={"examId": EXAM_ID, "state": "0"},
        )
        if str(state.get("examId")) != EXAM_ID or state.get("state") != 0:
            fail("00501 state transition to 0 was not confirmed")

        candidate = request(
            settings["base_url"],
            "/exam/api/candidate/save",
            body=candidate_body,
        )
        tracked["candidate_id"] = require_id(candidate.get("id"), "candidate id")
        participant_token = candidate.get("participantToken")
        if not isinstance(participant_token, str) or not participant_token:
            fail("participant token was not issued")

        paper = request(
            settings["base_url"],
            "/exam/api/management-traits/participant/create-paper",
            participant_token=participant_token,
            body={"examId": EXAM_ID},
        )
        tracked["paper_id"] = require_id(paper.get("paperId"), "paper id")
        paper_token = paper.get("paperToken")
        questions = paper.get("questions") or []
        if not isinstance(paper_token, str) or not paper_token or len(questions) != 140:
            fail("00501 paper creation did not return 140 questions")
        if len({str(question.get("id")) for question in questions}) != 140:
            fail("00501 paper question identities are not unique")

        for index, question in enumerate(questions, 1):
            question_id = require_id(question.get("id"), "paper question id")
            options = question.get("options") or []
            selected = [option for option in options if option.get("displayOrder") == 3]
            if len(selected) != 1:
                fail("a paper question has no unique displayOrder=3 option")
            option_id = require_id(selected[0].get("id"), "option id")
            answer = request(
                settings["base_url"],
                "/exam/api/management-traits/participant/fill-answer",
                participant_token=paper_token,
                body={
                    "paperId": tracked["paper_id"],
                    "paperQuestionId": question_id,
                    "optionId": option_id,
                },
            )
            if answer.get("answered") != index:
                fail("answer persistence count mismatch")

        submitted = request(
            settings["base_url"],
            "/exam/api/management-traits/participant/submit",
            participant_token=paper_token,
            body={"paperId": tracked["paper_id"], "submitType": "manual"},
        )
        tracked["run_id"] = require_id(submitted.get("runId"), "result run id")
        if submitted.get("status") != "completed" or submitted.get("answered") != 140:
            fail("00501 submission was not complete")

        runs = request(
            settings["base_url"],
            "/exam/api/management-traits/results/list",
            token=admin_token,
            body={"examId": EXAM_ID},
        )
        owned_runs = [row for row in runs if str(row.get("id")) == tracked["run_id"]]
        if len(owned_runs) != 1 or owned_runs[0].get("answeredQuestionCount") != 140 or owned_runs[0].get("totalQuestionCount") != 140:
            fail("admin results list did not contain the exact completed run")
        detail = request(
            settings["base_url"],
            "/exam/api/management-traits/results/detail?runId=" + urllib.parse.quote(tracked["run_id"]),
            token=admin_token,
            method="GET",
            body=None,
        )
        dimensions, modules = result_cardinality(detail)

        generated = request(
            settings["base_url"],
            "/exam/api/management-traits/report-reissues/generate",
            token=admin_token,
            body={"runId": tracked["run_id"]},
            timeout=REPORT_TIMEOUT,
        )
        report = generated.get("report") or {}
        tracked["reissue_id"] = require_id(report.get("id"), "reissue report id")
        if generated.get("reused") is not False or report.get("fileBytes", 0) < 1024:
            fail("new report reissue was not generated exactly once")

        query_id = urllib.parse.quote(tracked["reissue_id"])
        view_bytes, view_headers = request(
            settings["base_url"],
            "/exam/api/management-traits/report-reissues/view?reportId=" + query_id,
            token=admin_token,
            method="GET",
            body=None,
            binary=True,
            timeout=REPORT_TIMEOUT,
        )
        download_bytes, download_headers = request(
            settings["base_url"],
            "/exam/api/management-traits/report-reissues/download?reportId=" + query_id,
            token=admin_token,
            method="GET",
            body=None,
            binary=True,
            timeout=REPORT_TIMEOUT,
        )
        view_sha = hashlib.sha256(view_bytes).hexdigest()
        download_sha = hashlib.sha256(download_bytes).hexdigest()
        if (
            not view_bytes.startswith(b"%PDF-")
            or view_bytes != download_bytes
            or view_sha != download_sha
            or len(view_bytes) != report.get("fileBytes")
            or view_sha != report.get("fileSha")
            or "application/pdf" not in view_headers.get("Content-Type", "")
            or "application/pdf" not in download_headers.get("Content-Type", "")
        ):
            fail("report view/download PDF bytes do not match")

        receipt["core"] = {
            "answers": 140,
            "dimensions": dimensions,
            "modules": modules,
            "reportBytes": len(view_bytes),
            "reportSha256": view_sha,
            "viewDownloadEqual": True,
        }
        receipt["status"] = "passed"
    except Exception as exc:
        receipt["failure"] = type(exc).__name__
    finally:
        if state_verified and admin_token and settings:
            try:
                restored = request(
                    settings["base_url"],
                    "/exam/api/management-traits/admin/exam/state",
                    token=admin_token,
                    body={"examId": EXAM_ID, "state": "1"},
                )
                if str(restored.get("examId")) != EXAM_ID or restored.get("state") != 1:
                    fail("00501 state restoration was not confirmed")
            except Exception as exc:
                cleanup_errors.append("state_restore:" + type(exc).__name__)

        if db is not None:
            try:
                cleanup_summary = exact_cleanup(db, settings, marker, phone, tracked)
                receipt["cleanup"] = cleanup_summary
            except Exception as exc:
                cleanup_errors.append("owned_cleanup:" + type(exc).__name__)

        if redis_key and settings:
            try:
                redis_call(settings, "DEL", redis_key, expect="1")
            except Exception as exc:
                cleanup_errors.append("redis_cleanup:" + type(exc).__name__)

        if db is not None and baseline is not None:
            try:
                after = baseline_counts(db)
                state_count = db.scalar_int(
                    "SELECT COUNT(*) FROM el_exam WHERE id='9051103000000000501' AND state=1"
                )
                marker_count = db.scalar_int(
                    "SELECT COUNT(*) FROM el_candidate "
                    f"WHERE name='{marker}' OR telephone='{phone}'"
                )
                redis_absent = True
                if settings and redis_key:
                    redis_absent = redis_call(settings, "EXISTS", redis_key, expect="0") == "0"
                receipt["baselineAfter"] = after
                receipt["baselineRestored"] = after == baseline
                receipt["ownedMarkerRemaining"] = marker_count
                receipt["examStateRestored"] = state_count == 1
                receipt["redisKeyRemoved"] = redis_absent
                if after != baseline or marker_count != 0 or state_count != 1 or not redis_absent:
                    cleanup_errors.append("post_cleanup_verification:failed")
            except Exception as exc:
                cleanup_errors.append("post_cleanup_verification:" + type(exc).__name__)

        if mysql_file is not None:
            try:
                mysql_file.unlink(missing_ok=True)
            except OSError:
                cleanup_errors.append("mysql_client_cleanup:OSError")
            if mysql_file.exists():
                cleanup_errors.append("mysql_client_cleanup:remaining")

        if cleanup_errors:
            receipt["status"] = "failed"
            receipt["cleanupErrors"] = cleanup_errors
        receipt["durationSeconds"] = round(time.time() - started_at, 3)
        receipt["finishedAtEpoch"] = int(time.time())
        print(json.dumps(receipt, ensure_ascii=True, sort_keys=True, separators=(",", ":")))

    return 0 if receipt["status"] == "passed" else 1


if __name__ == "__main__":
    sys.exit(main())
