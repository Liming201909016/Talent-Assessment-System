#!/usr/bin/env python3
"""Authorized production-host 00401 phase-1 runtime E2E with exact cleanup.

Run on the application host as a user that can inspect the systemd service and
invoke mysql, redis-cli and pdfinfo. Stdout is reserved for one JSON receipt.
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
import zipfile
from io import BytesIO
from pathlib import Path
from xml.etree import ElementTree

APP_DIR = Path("/opt/talent-assessment")
SERVICE = "talent-assessment"
HTTP_TIMEOUT = 120
REPORT_TIMEOUT = 240
ID_RE = re.compile(r"^[A-Za-z0-9-]{1,64}$")
SHA_RE = re.compile(r"^[a-f0-9]{64}$")
BLOCKED_REASON = "一期正式报告内容尚未完成双重批准"
TOUCHED_TABLES = (
    "el_exam", "el_exam_repo", "el_exam_depart", "el_user_exam", "el_candidate",
    "el_tester", "el_paper", "el_paper_qu", "el_paper_qu_answer", "el_mbti_answer",
    "el_exam_competency_group", "el_exam_competency_dimension", "el_exam_competency_question",
    "el_competency_result", "el_competency_dimension_result", "el_competency_group_result",
    "el_competency_validity_result", "el_competency_result_run",
    "el_competency_result_run_overall", "el_competency_result_run_module",
    "el_competency_result_run_dimension", "el_competency_result_run_validity",
    "el_competency_report", "el_competency_report_current", "el_competency_report_audit",
    "el_competency_report_content_package", "sys_oper_log",
)


def fail(message):
    raise RuntimeError(message)


def run_checked(argv, *, env=None, input_text=None, timeout=60):
    try:
        result = subprocess.run(
            argv, input=input_text, text=True, capture_output=True, timeout=timeout,
            env=env, check=False,
        )
    except (OSError, subprocess.SubprocessError) as exc:
        raise RuntimeError("required host command failed") from exc
    if result.returncode != 0:
        raise RuntimeError("required host command failed")
    return result.stdout


def systemctl_value(prop):
    return run_checked(
        ["systemctl", "show", SERVICE, "--property=" + prop, "--value"], timeout=15,
    ).strip()


def parse_env_assignments(text):
    try:
        words = shlex.split(text, comments=False, posix=True)
    except ValueError as exc:
        raise RuntimeError("invalid systemd environment syntax") from exc
    values = {}
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


def effective_service_environment():
    values = {}
    for match in re.finditer(r"(?:^|\s)(/[^\s;]+)", systemctl_value("EnvironmentFiles")):
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
            ["sudo", "-n", "cat", str(proc_path)], capture_output=True, check=True,
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
    return values, pid


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
    env, pid = effective_service_environment()
    app_env = env.get("APP_ENV", "").strip().lower()
    if app_env != "production":
        fail("APP_ENV is not production")

    config_paths = (
        APP_DIR / "configs" / "application.yml",
        APP_DIR / "configs" / "application-production.yml",
    )
    existing_paths = [path for path in config_paths if path.is_file()]
    if not existing_paths:
        fail("application configuration is missing")
    text = "\n".join(
        path.read_text(encoding="utf-8-sig").replace("\r", "") for path in existing_paths
    )

    # MYSQL_DSN is trusted only when inherited by the effective service process.
    dsn_source = "process:MYSQL_DSN" if env.get("MYSQL_DSN") else "config:database.dsn"
    dsn = env.get("MYSQL_DSN") or yaml_scalar(text, "database", "dsn")
    # The deployed Go service historically names this section mysql; retain that
    # non-secret compatibility only when database.dsn is absent.
    if not dsn:
        dsn = yaml_scalar(text, "mysql", "dsn")
        dsn_source = "config:mysql.dsn"
    settings = {
        "dsn": dsn,
        "dsn_source": dsn_source,
        "redis_addr": env.get("REDIS_ADDR") or yaml_scalar(text, "redis", "addr", "127.0.0.1:6379"),
        "redis_db": env.get("REDIS_DB") or yaml_scalar(text, "redis", "db", "1"),
        "redis_password": env.get("REDIS_PASSWORD") or yaml_scalar(text, "redis", "password"),
        "jwt_secret": env.get("JWT_SECRET") or yaml_scalar(text, "jwt", "secret"),
        "login_user_key": env.get("JWT_LOGIN_USER_KEY") or yaml_scalar(text, "jwt", "loginUserKey", "login_user_key"),
        "server_port": env.get("SERVER_PORT") or yaml_scalar(text, "server", "port", "8092"),
        "upload_root": env.get("UPLOAD_PATH") or yaml_scalar(text, "upload", "path"),
        "report_root": env.get("REPORT_PATH") or yaml_scalar(text, "report", "path"),
        "app_env": app_env,
        "pid": pid,
        "config_paths": existing_paths,
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
    upload_root = Path(settings["upload_root"])
    if not upload_root.is_absolute():
        fail("upload root is not absolute")
    if settings["report_root"] and not Path(settings["report_root"]).is_absolute():
        fail("report root is not absolute")
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
        address_match = re.fullmatch(r"\[([^]]+)\]:(\d+)", address)
    else:
        address_match = re.fullmatch(r"(.+):(\d+)", address)
    if not address_match:
        fail("application MySQL address format is unsupported")
    host, port = address_match.groups()
    if not user or not host or not port.isdigit():
        fail("application MySQL DSN is incomplete")

    fd, name = tempfile.mkstemp(prefix="phase1-production-mysql-", suffix=".cnf", dir="/tmp")
    try:
        os.fchmod(fd, 0o600)
        escape = lambda value: value.replace("\\", "\\\\").replace('"', '\\"')
        content = '[client]\nuser="{0}"\npassword="{1}"\nhost="{2}"\nport={3}\n'.format(
            escape(user), escape(password), escape(host), port
        )
        os.write(fd, content.encode("utf-8"))
    finally:
        os.close(fd)
    path = Path(name)
    if stat.S_IMODE(path.stat().st_mode) != 0o600:
        path.unlink(missing_ok=True)
        fail("temporary MySQL client file mode is not 0600")
    return path, database


class MySQL:
    def __init__(self, client_file, database):
        self.client_file = client_file
        self.database = database

    def argv(self):
        return [
            "mysql", "--defaults-extra-file=" + str(self.client_file), "--batch",
            "--skip-column-names", "--raw", self.database,
        ]

    def query(self, sql):
        return run_checked(self.argv() + ["--execute", sql], timeout=90).strip()

    def execute(self, sql):
        run_checked(self.argv(), input_text=sql, timeout=90)

    def scalar_int(self, sql):
        value = self.query(sql)
        if not re.fullmatch(r"\d+", value):
            fail("database count query returned invalid data")
        return int(value)


def table_exists(db, table):
    if table not in TOUCHED_TABLES:
        fail("unapproved baseline table")
    return db.scalar_int(
        "SELECT COUNT(*) FROM information_schema.tables "
        f"WHERE table_schema=DATABASE() AND table_name='{table}'"
    ) == 1


def baseline_counts(db):
    counts = {}
    for table in TOUCHED_TABLES:
        if table_exists(db, table):
            counts[table] = db.scalar_int("SELECT COUNT(*) FROM " + table)
    required = {
        "el_exam", "el_candidate", "el_paper", "el_paper_qu", "el_paper_qu_answer",
        "el_exam_competency_group", "el_exam_competency_dimension",
        "el_exam_competency_question", "el_competency_result",
        "el_competency_dimension_result", "el_competency_group_result",
        "el_competency_validity_result", "el_competency_report",
        "el_competency_report_audit", "sys_oper_log",
    }
    if not required.issubset(counts):
        fail("required 00401 runtime tables are missing")
    return counts


def sql_literal(value):
    value = str(value or "")
    if not ID_RE.fullmatch(value):
        fail("unsafe owned identifier")
    return "'" + value + "'"


def require_id(value, label):
    value = str(value or "")
    if not ID_RE.fullmatch(value):
        fail(label + " is invalid")
    return value


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
    return match.group(1) or match.group(3), match.group(2) or match.group(4)


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
        env=env, timeout=20,
    ).strip()
    if expect is not None and stdout != expect:
        fail("Redis command returned an unexpected result")
    return stdout


def request_raw(base_url, path, *, token=None, participant_token=None, body=None,
                method="POST", timeout=HTTP_TIMEOUT):
    headers = {"Accept": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    if participant_token:
        headers["X-Competency-Token"] = participant_token
    payload = None
    if body is not None:
        headers["Content-Type"] = "application/json"
        payload = json.dumps(body, ensure_ascii=False, separators=(",", ":")).encode("utf-8")
    req = urllib.request.Request(base_url + path, data=payload, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as response:
            return response.status, response.read(), dict(response.headers)
    except urllib.error.HTTPError as exc:
        return exc.code, exc.read(), dict(exc.headers)
    except urllib.error.URLError as exc:
        raise RuntimeError("application HTTP request failed") from exc


def decode_json(raw):
    try:
        document = json.loads(raw.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise RuntimeError("application endpoint returned invalid JSON") from exc
    if not isinstance(document, dict):
        fail("application endpoint returned a non-object JSON response")
    return document


def request_json(base_url, path, *, token=None, participant_token=None, body=None,
                 method="POST", timeout=HTTP_TIMEOUT, require_success=True):
    status, raw, headers = request_raw(
        base_url, path, token=token, participant_token=participant_token,
        body=body, method=method, timeout=timeout,
    )
    document = decode_json(raw)
    success = status == 200 and document.get("code") in (0, 200) and document.get("success") is not False
    if require_success and not success:
        fail(
            "application endpoint rejected: " + path
            + " code=" + str(document.get("code"))
            + " msg=" + str(document.get("msg", ""))
        )
    return document, headers, status, success


def sha256_file(path):
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def read_only_smoke(settings):
    health_status, health_raw, _ = request_raw(settings["base_url"], "/health", method="GET")
    if health_status != 200:
        fail("health endpoint is not healthy")
    try:
        health = json.loads(health_raw.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise RuntimeError("health endpoint returned invalid JSON") from exc
    if not isinstance(health, dict) or health.get("status") != "ok":
        fail("health endpoint payload is not healthy")

    disk_server = APP_DIR / "server"
    process_exe = Path("/proc") / settings["pid"] / "exe"
    index_file = APP_DIR / "dist" / "index.html"
    for path in (disk_server, process_exe, index_file):
        if not path.exists():
            fail("release marker input is missing")
    server_sha = sha256_file(disk_server)
    process_sha = sha256_file(process_exe)
    index_sha = sha256_file(index_file)
    if not all(SHA_RE.fullmatch(value) for value in (server_sha, process_sha, index_sha)):
        fail("release marker SHA is invalid")
    if server_sha != process_sha:
        fail("running process does not match deployed server")
    return {
        "health": {"path": "/health", "httpStatus": health_status, "status": health["status"]},
        "mainPid": int(settings["pid"]),
        "serverSha256": server_sha,
        "processSha256": process_sha,
        "indexSha256": index_sha,
        "diskProcessEqual": True,
    }, {
        "service": SERVICE,
        "appEnvironment": settings["app_env"],
        "appDirectory": str(APP_DIR),
        "serverPath": str(disk_server),
        "processExePath": str(process_exe),
        "indexPath": str(index_file),
        "configPaths": [str(path) for path in settings["config_paths"]],
        "databaseDSNSource": settings["dsn_source"],
    }


def workbook_summary(payload):
    if not payload.startswith(b"PK"):
        fail("export is not an XLSX package")
    namespace = "{http://schemas.openxmlformats.org/spreadsheetml/2006/main}"
    with zipfile.ZipFile(BytesIO(payload)) as archive:
        names = set(archive.namelist())
        if "xl/workbook.xml" not in names:
            fail("XLSX workbook metadata is missing")
        workbook = ElementTree.fromstring(archive.read("xl/workbook.xml"))
        sheet_names = [node.attrib.get("name", "") for node in workbook.findall(".//" + namespace + "sheet")]
        if sheet_names != ["结果汇总", "逐题明细", "题目字典"]:
            fail("XLSX sheet names or order are invalid")
        row_counts = []
        for index in range(1, 4):
            member = f"xl/worksheets/sheet{index}.xml"
            if member not in names:
                fail("XLSX worksheet is missing")
            worksheet = ElementTree.fromstring(archive.read(member))
            row_counts.append(len(worksheet.findall(".//" + namespace + "row")))
    if row_counts[0] < 2 or row_counts[1:] != [91, 91]:
        fail("XLSX worksheet row counts are invalid")
    return {"sheets": sheet_names, "rowCounts": row_counts}


def report_file_binding(db, paper_id):
    if not table_exists(db, "el_competency_report"):
        return None
    row = db.query(
        "SELECT CONCAT(id,'\\t',pdf_path,'\\t',pdf_sha256,'\\t',pdf_size) "
        "FROM el_competency_report "
        f"WHERE paper_id={sql_literal(paper_id)} AND status='completed' ORDER BY create_time DESC LIMIT 1"
    )
    if not row:
        return None
    parts = row.split("\t")
    if len(parts) != 4 or not ID_RE.fullmatch(parts[0]) or not SHA_RE.fullmatch(parts[2]) or not parts[3].isdigit():
        fail("report file database binding is invalid")
    return {"reportId": parts[0], "path": parts[1], "sha256": parts[2], "size": int(parts[3])}


def validate_bound_report_path(settings, binding):
    path = Path(binding["path"])
    if not path.is_absolute() or path.suffix.lower() != ".pdf" or path.is_symlink():
        fail("bound report path is unsafe")
    resolved = path.resolve(strict=False)
    roots = [Path(settings["upload_root"]).resolve()]
    if settings["report_root"]:
        roots.append(Path(settings["report_root"]).resolve())
    if not any(resolved == root or root in resolved.parents for root in roots):
        fail("bound report path escapes configured roots")
    if resolved.is_file():
        if resolved.stat().st_size != binding["size"] or sha256_file(resolved) != binding["sha256"]:
            fail("bound report file does not match database metadata")
    return resolved


def owned_counts(db, exam_id, candidate_id, paper_id):
    exam = sql_literal(exam_id) if exam_id else "''"
    candidate = sql_literal(candidate_id) if candidate_id else "''"
    paper = sql_literal(paper_id) if paper_id else "''"
    queries = {
        "exam": f"SELECT COUNT(*) FROM el_exam WHERE id={exam}",
        "candidate": f"SELECT COUNT(*) FROM el_candidate WHERE id={candidate} OR exam_id={exam}",
        "paper": f"SELECT COUNT(*) FROM el_paper WHERE id={paper} OR exam_id={exam}",
        "result": f"SELECT COUNT(*) FROM el_competency_result WHERE paper_id={paper} OR exam_id={exam}",
        "dimensionResult": f"SELECT COUNT(*) FROM el_competency_dimension_result WHERE paper_id={paper}",
        "groupResult": f"SELECT COUNT(*) FROM el_competency_group_result WHERE paper_id={paper}",
        "validityResult": f"SELECT COUNT(*) FROM el_competency_validity_result WHERE paper_id={paper}",
        "report": f"SELECT COUNT(*) FROM el_competency_report WHERE paper_id={paper} OR exam_id={exam}",
        "reportAudit": f"SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id={paper}",
    }
    if table_exists(db, "el_competency_result_run"):
        queries["resultRun"] = f"SELECT COUNT(*) FROM el_competency_result_run WHERE paper_id={paper} OR exam_id={exam}"
    if table_exists(db, "el_competency_report_current"):
        queries["reportCurrent"] = f"SELECT COUNT(*) FROM el_competency_report_current WHERE paper_id={paper}"
    return {name: db.scalar_int(sql) for name, sql in queries.items()}


def cleanup_owned_export_logs(db, exam_id):
    if not exam_id or not table_exists(db, "sys_oper_log"):
        return 0
    exam_literal = sql_literal(exam_id)
    ids = db.query(
        "SELECT oper_id FROM sys_oper_log "
        "WHERE method='ExamHandler.ExportRawAnswers' "
        f"AND oper_param LIKE CONCAT('{{\"examId\":\"', {exam_literal}, '\",%') ORDER BY oper_id"
    ).splitlines()
    if not ids:
        return 0
    if any(not re.fullmatch(r"\d+", value) for value in ids):
        fail("owned export audit identifier is invalid")
    db.execute(
        "SET SESSION SQL_SAFE_UPDATES=1;\n"
        "DELETE FROM sys_oper_log WHERE oper_id IN (" + ",".join(ids) + ");\n"
    )
    return len(ids)


def main():
    receipt = {
        "schema": "production-phase1-runtime-e2e-receipt-v1",
        "status": "failed",
        "environment": "production",
        "assessment": "00401-phase1",
        "authorized": True,
    }
    settings = None
    db = None
    mysql_file = None
    baseline = None
    admin_token = ""
    redis_key = ""
    redis_created = False
    exam_id = ""
    candidate_id = ""
    paper_id = ""
    report_binding = None
    cleanup_errors = []
    started_at = time.time()
    marker = "PROD-PHASE1-E2E-" + uuid.uuid4().hex[:16]
    phone = "198" + str(int(time.time() * 1000))[-8:]
    try:
        settings = load_settings()
        smoke, marker_inputs = read_only_smoke(settings)
        receipt["readOnlySmoke"] = smoke
        receipt["releaseMarkerInputs"] = marker_inputs

        mysql_file, database = mysql_client_from_dsn(settings["dsn"])
        db = MySQL(mysql_file, database)
        if db.query("SELECT DATABASE()") != "element":
            fail("MySQL client is not using the authorized database")
        baseline = baseline_counts(db)
        receipt["baselineBefore"] = baseline
        if db.scalar_int(
            "SELECT COUNT(*) FROM el_exam WHERE title=" + sql_literal(marker)
        ) != 0 or db.scalar_int(
            "SELECT COUNT(*) FROM el_candidate WHERE name=" + sql_literal(marker) +
            " OR telephone=" + sql_literal(phone)
        ) != 0:
            fail("synthetic ownership marker already exists")

        token_id = "prod-phase1-e2e-" + uuid.uuid4().hex
        redis_key = "login_tokens:" + token_id
        redis_call(settings, "EXISTS", redis_key, expect="0")
        now_ms = int(time.time() * 1000)
        login_user = {
            "userId": 1, "token": token_id, "loginTime": now_ms,
            "expireTime": now_ms + 1800000, "permissions": ["*:*:*"],
            "roles": ["admin"], "user": None,
        }
        redis_call(
            settings, "SET", redis_key,
            json.dumps(login_user, separators=(",", ":")), "NX", "EX", "1800", expect="OK",
        )
        redis_created = True
        admin_token = make_admin_token(settings["jwt_secret"], settings["login_user_key"], token_id)

        exam_doc, _, _, _ = request_json(
            settings["base_url"], "/exam/api/exam/exam/save", token=admin_token,
            body={
                "title": marker,
                "content": "authorized production phase-1 runtime verification",
                "assessmentType": "competency", "scoringMode": "competency_average",
                "joinType": 1, "openType": 1, "isOpen": 1, "answerType": 1,
                "state": 0, "totalTime": 20, "repoList": [], "departIds": [],
            },
        )
        exam = exam_doc.get("data") or {}
        exam_id = require_id(exam.get("id"), "exam id")
        if db.scalar_int(
            "SELECT COUNT(*) FROM el_exam WHERE id=" + sql_literal(exam_id) +
            " AND title=" + sql_literal(marker) +
            " AND assessment_type='competency' AND scoring_mode='competency_average'"
        ) != 1:
            fail("synthetic exam ownership or mode was not persisted")

        publish_doc, _, _, _ = request_json(
            settings["base_url"], "/exam/api/competency/exams/publish",
            token=admin_token, body={"examId": exam_id},
        )
        published = publish_doc.get("data") or {}
        if published.get("dimensionCount") != 10 or published.get("questionCount") != 90:
            fail("published inventory is not 10 dimensions and 90 questions")
        repeated_doc, _, _, _ = request_json(
            settings["base_url"], "/exam/api/competency/exams/publish",
            token=admin_token, body={"examId": exam_id},
        )
        repeated = repeated_doc.get("data") or {}
        if repeated.get("alreadyPublished") is not True or repeated.get("dimensionCount") != 10 or repeated.get("questionCount") != 90:
            fail("republish was not idempotent")
        snapshot = db.query(
            "SELECT CONCAT("
            f"(SELECT COUNT(*) FROM el_exam_competency_group WHERE exam_id={sql_literal(exam_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_exam_competency_dimension WHERE exam_id={sql_literal(exam_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_exam_competency_dimension WHERE exam_id={sql_literal(exam_id)} AND group_id IS NOT NULL),'|',"
            f"(SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id={sql_literal(exam_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id={sql_literal(exam_id)} AND competency_question_type='dimension'),'|',"
            f"(SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id={sql_literal(exam_id)} AND competency_question_type='validity'))"
        )
        if snapshot != "2|10|10|90|80|10":
            fail("published database snapshot inventory is invalid")

        candidate_doc, _, _, _ = request_json(
            settings["base_url"], "/exam/api/candidate/save", token=admin_token,
            body={"examId": exam_id, "name": marker, "telephone": phone, "gender": "0"},
        )
        candidate = candidate_doc.get("data") or {}
        candidate_id = require_id(candidate.get("id"), "candidate id")
        participant_token = candidate.get("participantToken")
        if not isinstance(participant_token, str) or not participant_token:
            fail("participant token was not issued")

        create_body = {
            "examId": exam_id, "participantId": candidate_id,
            "participantType": "candidate", "participantToken": participant_token,
        }
        paper_doc, _, _, _ = request_json(
            settings["base_url"], "/exam/api/competency/participant/create-paper", body=create_body,
        )
        paper = paper_doc.get("data") or {}
        paper_id = require_id(paper.get("paperId"), "paper id")
        paper_token = paper.get("paperToken")
        if not isinstance(paper_token, str) or not paper_token:
            fail("paper token was not issued")
        restored_doc, _, _, _ = request_json(
            settings["base_url"], "/exam/api/competency/participant/create-paper", body=create_body,
        )
        restored = restored_doc.get("data") or {}
        if restored.get("paperId") != paper_id:
            fail("second paper creation did not restore the same paper")

        detail_doc, _, _, _ = request_json(
            settings["base_url"], "/exam/api/competency/participant/paper-detail",
            participant_token=paper_token, body={"paperId": paper_id},
        )
        detail = detail_doc.get("data") or {}
        questions = detail.get("questions") or []
        if detail.get("totalCount") != 90 or len(questions) != 90:
            fail("paper does not contain exactly 90 questions")
        question_ids = [require_id(row.get("id"), "paper question id") for row in questions]
        if len(set(question_ids)) != 90:
            fail("paper question identities are not unique")
        labels = ["完全不符合", "比较不符合", "不确定", "比较符合", "完全符合"]
        for question, question_id in zip(questions, question_ids):
            if [option.get("label") for option in question.get("options") or []] != labels:
                fail("paper option labels do not match the frozen contract")
            raw_value = 1 if str(question.get("code", "")).startswith("P1-VAL-Q") else 3
            request_json(
                settings["base_url"], "/exam/api/competency/participant/fill-answer",
                participant_token=paper_token,
                body={"paperId": paper_id, "paperQuestionId": question_id, "rawValue": raw_value},
            )

        submitted_doc, _, _, _ = request_json(
            settings["base_url"], "/exam/api/competency/participant/submit",
            participant_token=paper_token, body={"paperId": paper_id, "submitType": "manual"},
        )
        submitted = submitted_doc.get("data") or {}
        if submitted.get("isComplete") is not True:
            fail("manual submission was not complete")
        resubmit_doc, _, _, _ = request_json(
            settings["base_url"], "/exam/api/competency/participant/submit",
            participant_token=paper_token, body={"paperId": paper_id, "submitType": "manual"},
        )
        if (resubmit_doc.get("data") or {}).get("alreadySubmitted") is not True:
            fail("manual resubmission was not idempotent")

        result_doc, _, _, _ = request_json(
            settings["base_url"], "/exam/api/competency/results/detail",
            token=admin_token, body={"paperId": paper_id},
        )
        result_data = result_doc.get("data") or {}
        result = result_data.get("result") or {}
        dimensions = result_data.get("dimensions") or []
        groups = result_data.get("groups") or []
        validity = result_data.get("validity") or {}
        if len(dimensions) != 10 or len(groups) != 2:
            fail("admin detail cardinality is invalid")
        if float(result.get("overallScore")) != 30 or result.get("evaluationLevel") != "weak" or result.get("isComplete") != 1:
            fail("admin detail overall values are invalid")
        if result.get("totalQuestionCount") != 90 or result.get("dimensionQuestionCount") != 80:
            fail("admin detail question counts are invalid")
        if any(float(row.get("dimensionScore")) != 3 or row.get("levelCode") != "L3" for row in dimensions):
            fail("admin detail dimension values are invalid")
        if any(float(row.get("groupScore")) != 3 or row.get("levelCode") != "L3" for row in groups):
            fail("admin detail group values are invalid")
        if float(validity.get("validityScore")) != 10 or validity.get("validityStatus") != "good":
            fail("admin detail validity values are invalid")

        legacy_counts = db.query(
            "SELECT CONCAT("
            f"(SELECT COUNT(*) FROM el_competency_dimension_result WHERE paper_id={sql_literal(paper_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_group_result WHERE paper_id={sql_literal(paper_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_validity_result WHERE paper_id={sql_literal(paper_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_result WHERE paper_id={sql_literal(paper_id)}))"
        )
        if legacy_counts != "10|2|1|1":
            fail("persisted v1 result counts are invalid")
        v2_counts = None
        if table_exists(db, "el_competency_result_run"):
            v2_counts = db.query(
                "SELECT CONCAT("
                f"(SELECT COUNT(*) FROM el_competency_result_run WHERE paper_id={sql_literal(paper_id)}),'|',"
                f"(SELECT COUNT(*) FROM el_competency_result_run_overall o JOIN el_competency_result_run r ON r.id=o.result_run_id WHERE r.paper_id={sql_literal(paper_id)}),'|',"
                f"(SELECT COUNT(*) FROM el_competency_result_run_module m JOIN el_competency_result_run r ON r.id=m.result_run_id WHERE r.paper_id={sql_literal(paper_id)}),'|',"
                f"(SELECT COUNT(*) FROM el_competency_result_run_dimension d JOIN el_competency_result_run r ON r.id=d.result_run_id WHERE r.paper_id={sql_literal(paper_id)}),'|',"
                f"(SELECT COUNT(*) FROM el_competency_result_run_validity v JOIN el_competency_result_run r ON r.id=v.result_run_id WHERE r.paper_id={sql_literal(paper_id)}))"
            )
            if v2_counts != "1|1|3|10|1":
                fail("persisted v2 result-run counts are invalid")

        good_doc, _, _, _ = request_json(
            settings["base_url"], "/exam/api/competency/results/paging", token=admin_token,
            body={"examId": exam_id, "current": 1, "size": 20, "validity": "good", "sortBy": "overallScore"},
        )
        questionable_doc, _, _, _ = request_json(
            settings["base_url"], "/exam/api/competency/results/paging", token=admin_token,
            body={"examId": exam_id, "current": 1, "size": 20, "validity": "questionable", "sortBy": "submittedAt"},
        )
        if (good_doc.get("data") or {}).get("total") != 1 or (questionable_doc.get("data") or {}).get("total") != 0:
            fail("paging validity filters are invalid")

        export_status, export_bytes, export_headers = request_raw(
            settings["base_url"], "/exam/api/exam/exam/export-raw-data?examId=" + urllib.parse.quote(exam_id),
            token=admin_token, method="GET", timeout=REPORT_TIMEOUT,
        )
        if export_status != 200 or "spreadsheetml.sheet" not in export_headers.get("Content-Type", ""):
            fail("raw export did not return an XLSX file")
        export_summary = workbook_summary(export_bytes)

        report_doc, _, report_http, report_approved = request_json(
            settings["base_url"], "/exam/api/competency/admin/report-data?paperId=" + urllib.parse.quote(paper_id),
            token=admin_token, method="GET", require_success=False,
        )
        if report_http != 200:
            fail("report-data returned an unexpected HTTP status")
        report_summary = {"approved": report_approved}
        if report_approved:
            report_data = report_doc.get("data") or {}
            schema = report_data.get("schemaVersion")
            if schema != "competency-phase1-report-data-v1" or report_data.get("reportKind") != "frontline_phase1":
                fail("approved report DTO identity is invalid")
            if len(report_data.get("pages") or []) != 10 or len(report_data.get("groups") or []) != 2 or len(report_data.get("dimensions") or []) != 10:
                fail("approved report DTO cardinality is invalid")
            if "validityScore" in json.dumps(report_data, ensure_ascii=False):
                fail("approved report DTO exposes the private validity score")

            generated_doc, _, _, _ = request_json(
                settings["base_url"], "/exam/api/competency/reports/generate",
                token=admin_token, body={"paperId": paper_id}, timeout=REPORT_TIMEOUT,
            )
            generated = generated_doc.get("data") or {}
            if generated.get("status") != "completed" or int(generated.get("pdfSize", 0)) < 1024:
                fail("approved report generation did not complete")
            report_binding = report_file_binding(db, paper_id)
            if report_binding is None:
                fail("completed report has no database file binding")
            bound_path = validate_bound_report_path(settings, report_binding)
            if not bound_path.is_file():
                fail("completed report file is missing")

            download_status, download_bytes, download_headers = request_raw(
                settings["base_url"], "/exam/api/competency/reports/download?paperId=" + urllib.parse.quote(paper_id),
                token=admin_token, method="GET", timeout=REPORT_TIMEOUT,
            )
            if download_status != 200 or not download_bytes.startswith(b"%PDF-") or "application/pdf" not in download_headers.get("Content-Type", ""):
                fail("approved report download is not a valid PDF response")
            if len(download_bytes) != report_binding["size"] or hashlib.sha256(download_bytes).hexdigest() != report_binding["sha256"]:
                fail("downloaded PDF does not match the database-bound file")
            fd, pdf_name = tempfile.mkstemp(prefix="phase1-production-report-", suffix=".pdf", dir="/tmp")
            try:
                with os.fdopen(fd, "wb") as stream:
                    stream.write(download_bytes)
                pdf_info = run_checked(["pdfinfo", pdf_name], timeout=30)
                page_match = re.search(r"^Pages:\s+(\d+)$", pdf_info, re.MULTILINE)
                if not page_match or int(page_match.group(1)) != 10:
                    fail("approved report PDF is not exactly 10 pages")
            finally:
                Path(pdf_name).unlink(missing_ok=True)
            report_counts = db.query(
                "SELECT CONCAT("
                f"(SELECT COUNT(*) FROM el_competency_report WHERE paper_id={sql_literal(paper_id)} AND status='completed'),'|',"
                f"(SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id={sql_literal(paper_id)} AND status=1))"
            )
            if report_counts != "1|2":
                fail("approved report persistence or audit counts are invalid")
            report_summary.update({"schemaVersion": schema, "pages": 10, "reports": 1, "audits": 2})
        else:
            if report_doc.get("msg") != BLOCKED_REASON:
                fail("report-data did not fail closed for the approval gate")
            for path, method, body in (
                ("/exam/api/competency/reports/generate", "POST", {"paperId": paper_id}),
                ("/exam/api/competency/reports/download?paperId=" + urllib.parse.quote(paper_id), "GET", None),
            ):
                gate_doc, gate_headers, gate_http, gate_success = request_json(
                    settings["base_url"], path, token=admin_token, method=method,
                    body=body, timeout=REPORT_TIMEOUT, require_success=False,
                )
                if gate_http != 200 or gate_success or not str(gate_doc.get("msg", "")).strip():
                    fail("report endpoint did not fail closed for the approval gate")
                if "application/pdf" in gate_headers.get("Content-Type", ""):
                    fail("gated report download returned a PDF content type")
            gated_counts = db.query(
                "SELECT CONCAT("
                f"(SELECT COUNT(*) FROM el_competency_report WHERE paper_id={sql_literal(paper_id)}),'|',"
                f"(SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id={sql_literal(paper_id)}))"
            )
            if gated_counts != "0|0":
                fail("gated report operations wrote report or audit rows")
            report_summary.update({"reason": "dual_approval_required", "reports": 0, "audits": 0})

        receipt["runtime"] = {
            "publish": {"groups": 2, "dimensions": 10, "questions": 90, "dimensionQuestions": 80, "validityQuestions": 10, "republishIdempotent": True},
            "paper": {"restoredSame": True, "questions": 90, "labelsVerified": True},
            "submission": {"manual": True, "resubmitIdempotent": True},
            "result": {"dimensions": 10, "groups": 2, "validity": "good", "validityScore": 10, "overallScore": 30, "overallLevel": "weak", "v1Counts": legacy_counts, "v2Counts": v2_counts},
            "paging": {"good": 1, "questionable": 0},
            "export": export_summary,
            "report": report_summary,
        }
        receipt["status"] = "passed"
    except Exception as exc:
        receipt["failure"] = type(exc).__name__
        receipt["failureDetail"] = str(exc)
    finally:
        if exam_id and admin_token and settings:
            try:
                delete_doc, _, _, _ = request_json(
                    settings["base_url"], "/exam/api/exam/exam/delete",
                    token=admin_token, body={"ids": [exam_id]},
                )
                if delete_doc.get("data") is not True:
                    fail("full-chain exam deletion was not confirmed")
            except Exception as exc:
                cleanup_errors.append("exam_delete:" + type(exc).__name__)

        if db is not None and exam_id:
            try:
                removed_logs = cleanup_owned_export_logs(db, exam_id)
                receipt.setdefault("cleanup", {})["ownedExportLogs"] = removed_logs
            except Exception as exc:
                cleanup_errors.append("export_log_cleanup:" + type(exc).__name__)

        if report_binding is not None and settings is not None:
            try:
                bound_path = validate_bound_report_path(settings, report_binding)
                if bound_path.exists():
                    bound_path.unlink()
                    receipt.setdefault("cleanup", {})["orphanedBoundPdfRemoved"] = True
                else:
                    receipt.setdefault("cleanup", {})["orphanedBoundPdfRemoved"] = False
            except Exception as exc:
                cleanup_errors.append("report_file_cleanup:" + type(exc).__name__)

        if redis_created and redis_key and settings is not None:
            try:
                redis_call(settings, "DEL", redis_key, expect="1")
            except Exception as exc:
                cleanup_errors.append("redis_cleanup:" + type(exc).__name__)

        if db is not None and baseline is not None:
            try:
                owned_after = owned_counts(db, exam_id, candidate_id, paper_id)
                after = baseline_counts(db)
                redis_absent = True
                if redis_key and settings is not None:
                    redis_absent = redis_call(settings, "EXISTS", redis_key, expect="0") == "0"
                receipt.setdefault("cleanup", {}).update({
                    "ownedRemaining": owned_after,
                    "redisKeyRemoved": redis_absent,
                    "baselineAfter": after,
                    "baselineRestored": after == baseline,
                })
                if any(owned_after.values()) or not redis_absent or after != baseline:
                    cleanup_errors.append("exact_restoration:failed")
            except Exception as exc:
                cleanup_errors.append("exact_restoration:" + type(exc).__name__)

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
        if receipt["status"] != "passed":
            sys.stderr.write(receipt.get("failure", "CleanupError") + "\n")

    return 0 if receipt["status"] == "passed" else 1


if __name__ == "__main__":
    sys.exit(main())
