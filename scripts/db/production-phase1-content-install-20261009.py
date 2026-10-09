#!/usr/bin/env python3
"""Install the approved 00401 phase-1 workbook on the production host.

The installer is additive, fail-closed, and uses only the Python standard
library plus the mysql/mysqldump and redis-cli host clients. Stdout is reserved
for one non-secret JSON receipt.
"""

import base64
import datetime as dt
import gzip
import hashlib
import hmac
import json
import os
import re
import shlex
import shutil
import stat
import subprocess
import sys
import tempfile
import urllib.error
import urllib.request
import uuid
import zipfile
from io import BytesIO
from pathlib import Path
from xml.etree import ElementTree as ET

APP_DIR = Path("/opt/talent-assessment")
BACKUP_DIR = APP_DIR / "backups"
SERVICE = "talent-assessment"
DEFAULT_WORKBOOK = Path("/tmp/competency-phase1-import-20260810.xlsx")
EXPECTED_WORKBOOK_SHA256 = "828c4267e6c7ad387a73ddb0e923b461d5a336ef225bd7414216c0def814de9f"
EXPECTED_HEADERS = [
    "维度序号", "维度名称", "题目类型", "题目编号", "维度内题号",
    "题目内容", "考察点", "计分方向", "启用状态", "备注",
]
DIMENSIONS = [
    ("competency-a1-01", "A1-01", "逻辑思维", "通用能力", "基层员工", "逻辑分析严谨，推理判断有据", 1, 0),
    ("competency-a1-02", "A1-02", "数字应用", "通用能力", "基层员工", "善用数字化工具与AI技术，具备数据思维", 2, 0),
    ("competency-a1-03", "A1-03", "计划执行", "通用能力", "基层员工", "高效推进计划并达成预期结果", 3, 0),
    ("competency-a1-04", "A1-04", "持续学习", "通用能力", "基层员工", "主动学习，多渠道获取知识并学以致用", 4, 0),
    ("competency-a1-05", "A1-05", "沟通表达", "通用能力", "基层员工", "清晰传递信息，重视倾听与反馈", 5, 0),
    ("competency-b1-01", "B1-01", "敬业奉献", "心理素养", "基层员工", "视工作为使命，全心投入，甘于奉献", 6, 0),
    ("competency-b1-02", "B1-02", "求真务实", "心理素养", "基层员工", "追求真理，尊重事实，注重实效", 7, 0),
    ("competency-b1-03", "B1-03", "自律性", "心理素养", "基层员工", "自我约束，规划在先，言行一致", 8, 0),
    ("competency-b1-04", "B1-04", "成就导向", "心理素养", "基层员工", "追求工作成功，不断挑战更高目标", 9, 0),
    ("competency-b1-05", "B1-05", "合作意识", "心理素养", "基层员工", "主动协作，乐于分享，促成共赢", 10, 0),
]
DIMENSION_IDS = tuple(row[0] for row in DIMENSIONS)
DIMENSION_CODES = tuple(row[1] for row in DIMENSIONS)
DIMENSION_NAMES = tuple(row[2] for row in DIMENSIONS)
ARCHIVAL_SUFFIX = "（历史）"
LEGACY_COLLISIONS = (
    ("competency-d01", "D01", "沟通表达", 1),
    ("competency-d03", "D03", "数字应用", 3),
    ("competency-d04", "D04", "计划执行", 4),
    ("competency-d05", "D05", "逻辑思维", 5),
    ("competency-d06", "D06", "持续学习", 6),
    ("competency-d21", "D21", "敬业奉献", 21),
    ("competency-d28", "D28", "求真务实", 28),
    ("competency-d33", "D33", "自律性", 33),
)
FROZEN_ASSOCIATION_RESULT_TABLES = (
    "el_exam_competency_group",
    "el_exam_competency_dimension",
    "el_exam_competency_question",
    "el_paper_qu",
    "el_paper_qu_answer",
    "el_competency_result",
    "el_competency_group_result",
    "el_competency_dimension_result",
    "el_competency_validity_result",
    "el_competency_result_run",
    "el_competency_result_run_overall",
    "el_competency_result_run_module",
    "el_competency_result_run_dimension",
    "el_competency_result_run_validity",
)
SHA_RE = re.compile(r"^[a-f0-9]{64}$")
SAFE_ID_RE = re.compile(r"^[A-Za-z0-9-]{1,128}$")
NS = {
    "m": "http://schemas.openxmlformats.org/spreadsheetml/2006/main",
    "r": "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
}
REL_NS = {"p": "http://schemas.openxmlformats.org/package/2006/relationships"}


class InstallError(RuntimeError):
    pass


def fail(message):
    raise InstallError(message)


def run_checked(argv, *, env=None, input_text=None, timeout=120):
    try:
        result = subprocess.run(
            argv, input=input_text, text=True, capture_output=True, env=env,
            timeout=timeout, check=False,
        )
    except (OSError, subprocess.SubprocessError) as exc:
        raise InstallError("required host command failed") from exc
    if result.returncode != 0:
        raise InstallError("required host command failed")
    return result.stdout


def sha256_bytes(data):
    return hashlib.sha256(data).hexdigest()


def sha256_file(path):
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def sql_string(value):
    data = str(value).encode("utf-8")
    return "CONVERT(UNHEX('" + data.hex() + "') USING utf8mb4)"


def sql_list(values):
    return ",".join(sql_string(value) for value in values)


def systemctl_value(prop):
    return run_checked(
        ["systemctl", "show", SERVICE, "--property=" + prop, "--value"], timeout=20,
    ).strip()


def parse_env_assignments(text):
    try:
        words = shlex.split(text, comments=False, posix=True)
    except ValueError as exc:
        raise InstallError("invalid systemd environment syntax") from exc
    output = {}
    for word in words:
        if "=" not in word:
            continue
        key, value = word.split("=", 1)
        if re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", key):
            output[key] = value
    return output


def parse_environment_file(path):
    output = {}
    text = path.read_text(encoding="utf-8-sig").replace("\r", "")
    for raw in text.splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        if line.startswith("export "):
            line = line[7:].lstrip()
        if "=" not in line:
            continue
        key, value = line.split("=", 1)
        key = key.strip()
        if not re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", key):
            continue
        try:
            parts = shlex.split(value, comments=True, posix=True)
        except ValueError as exc:
            raise InstallError("invalid service environment file") from exc
        output[key] = "" if not parts else parts[0]
    return output


def effective_service_environment():
    if systemctl_value("ActiveState") != "active":
        fail("application service is not active")
    values = {}
    for match in re.finditer(r"(?:^|\s)(/[^\s;]+)", systemctl_value("EnvironmentFiles")):
        path = Path(match.group(1))
        if path.is_file():
            values.update(parse_environment_file(path))
    values.update(parse_env_assignments(systemctl_value("Environment")))
    pid = systemctl_value("MainPID")
    if not pid.isdigit() or int(pid) <= 0:
        fail("application process is not running")
    proc_env = Path("/proc") / pid / "environ"
    try:
        raw = proc_env.read_bytes()
    except OSError:
        result = subprocess.run(
            ["sudo", "-n", "cat", str(proc_env)], capture_output=True, check=False,
        )
        if result.returncode != 0:
            fail("effective process environment is unavailable")
        raw = result.stdout
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
    if env.get("APP_ENV", "").strip().lower() != "production":
        fail("APP_ENV is not production")
    paths = (
        APP_DIR / "configs" / "application.yml",
        APP_DIR / "configs" / "application-production.yml",
    )
    existing = [path for path in paths if path.is_file()]
    if not existing:
        fail("application configuration is missing")
    text = "\n".join(
        path.read_text(encoding="utf-8-sig").replace("\r", "") for path in existing
    )
    dsn = env.get("MYSQL_DSN")
    dsn_source = "process:MYSQL_DSN"
    if not dsn:
        dsn = yaml_scalar(text, "mysql", "dsn")
        dsn_source = "config:mysql.dsn"
    if not dsn:
        dsn = yaml_scalar(text, "database", "dsn")
        dsn_source = "config:database.dsn"
    settings = {
        "dsn": dsn,
        "dsnSource": dsn_source,
        "redisAddr": env.get("REDIS_ADDR") or yaml_scalar(text, "redis", "addr", "127.0.0.1:6379"),
        "redisDB": env.get("REDIS_DB") or yaml_scalar(text, "redis", "db", "1"),
        "redisPassword": env.get("REDIS_PASSWORD") or yaml_scalar(text, "redis", "password"),
        "jwtSecret": env.get("JWT_SECRET") or yaml_scalar(text, "jwt", "secret"),
        "jwtHeader": env.get("JWT_HEADER") or yaml_scalar(text, "jwt", "header", "Authorization"),
        "jwtPrefix": env.get("JWT_PREFIX") or yaml_scalar(text, "jwt", "prefix", "Bearer "),
        "jwtClaim": env.get("JWT_LOGIN_USER_KEY") or yaml_scalar(text, "jwt", "loginUserKey", "login_user_key"),
        "serverPort": env.get("SERVER_PORT") or yaml_scalar(text, "server", "port", "8092"),
        "pid": pid,
    }
    if not settings["dsn"] or not settings["jwtSecret"]:
        fail("required application credentials are not configured")
    if not re.fullmatch(r"\d+", settings["redisDB"]):
        fail("invalid Redis database")
    if not re.fullmatch(r"\d{1,5}", settings["serverPort"]):
        fail("invalid application port")
    port = int(settings["serverPort"])
    if not 1 <= port <= 65535:
        fail("invalid application port")
    if not re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", settings["jwtClaim"]):
        fail("invalid JWT claim name")
    if not re.fullmatch(r"[A-Za-z0-9-]+", settings["jwtHeader"]):
        fail("invalid JWT header name")
    if settings["jwtPrefix"] and not settings["jwtPrefix"].endswith((" ", "\t")):
        settings["jwtPrefix"] += " "
    settings["baseURL"] = f"http://127.0.0.1:{port}"
    return settings


def mysql_client_from_dsn(dsn):
    match = re.fullmatch(r"([^:]+):(.*)@tcp\(([^)]+)\)/([^?]+)(?:\?.*)?", dsn)
    if not match:
        fail("application MySQL DSN format is unsupported")
    user, password, address, database = match.groups()
    if database != "element" or any("\n" in value or "\r" in value for value in match.groups()):
        fail("application MySQL DSN is outside the authorized database")
    address_match = (
        re.fullmatch(r"\[([^]]+)\]:(\d+)", address)
        if address.startswith("[") else re.fullmatch(r"(.+):(\d+)", address)
    )
    if not address_match or not user:
        fail("application MySQL DSN is incomplete")
    host, port = address_match.groups()
    fd, name = tempfile.mkstemp(prefix="phase1-content-", suffix=".cnf", dir="/tmp")
    try:
        os.fchmod(fd, 0o600)
        escape = lambda value: value.replace("\\", "\\\\").replace('"', '\\"')
        body = '[client]\nuser="{}"\npassword="{}"\nhost="{}"\nport={}\n'.format(
            escape(user), escape(password), escape(host), port
        )
        os.write(fd, body.encode("utf-8"))
    finally:
        os.close(fd)
    path = Path(name)
    if stat.S_IMODE(path.stat().st_mode) != 0o600:
        path.unlink(missing_ok=True)
        fail("temporary MySQL client mode is not 0600")
    return path, database


class MySQL:
    def __init__(self, client_file, database):
        self.client_file = client_file
        self.database = database

    def base(self, executable="mysql"):
        return [
            executable, "--defaults-extra-file=" + str(self.client_file),
            "--batch", "--skip-column-names", "--raw", self.database,
        ]

    def query(self, sql):
        return run_checked(self.base() + ["--execute", sql], timeout=120).strip()

    def execute(self, sql):
        run_checked(self.base(), input_text=sql, timeout=120)

    def scalar_int(self, sql):
        value = self.query(sql)
        if not re.fullmatch(r"\d+", value):
            fail("database count query returned invalid data")
        return int(value)

    def table_exists(self, table):
        if not re.fullmatch(r"[a-z0-9_]+", table):
            fail("invalid table identifier")
        return self.scalar_int(
            "SELECT COUNT(*) FROM information_schema.tables "
            f"WHERE table_schema=DATABASE() AND table_name={sql_string(table)}"
        ) == 1

    def column_exists(self, table, column):
        if not re.fullmatch(r"[a-z0-9_]+", table + column):
            fail("invalid column identifier")
        return self.scalar_int(
            "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() "
            f"AND table_name={sql_string(table)} AND column_name={sql_string(column)}"
        ) == 1


def split_host_port(address):
    match = re.fullmatch(r"\[([^]]+)\]:(\d+)|([^:]+):(\d+)", address)
    if not match:
        fail("unsupported Redis address")
    return match.group(1) or match.group(3), match.group(2) or match.group(4)


def redis_call(settings, *arguments, expect=None):
    host, port = split_host_port(settings["redisAddr"])
    env = os.environ.copy()
    if settings["redisPassword"]:
        env["REDISCLI_AUTH"] = settings["redisPassword"]
    else:
        env.pop("REDISCLI_AUTH", None)
    output = run_checked(
        ["redis-cli", "--raw", "-h", host, "-p", port, "-n", settings["redisDB"], *arguments],
        env=env, timeout=30,
    ).strip()
    if expect is not None and output != expect:
        fail("Redis command returned an unexpected result")
    return output


def b64url(data):
    return base64.urlsafe_b64encode(data).rstrip(b"=").decode("ascii")


def make_admin_token(secret, claim, token_id):
    header = b64url(json.dumps({"alg": "HS512", "typ": "JWT"}, separators=(",", ":")).encode())
    payload = b64url(json.dumps({claim: token_id}, separators=(",", ":")).encode())
    signing = header + "." + payload
    signature = hmac.new(secret.encode(), signing.encode(), hashlib.sha512).digest()
    return signing + "." + b64url(signature)


def request_raw(settings, path, *, token=None, data=None, content_type=None, method="GET", timeout=180):
    headers = {"Accept": "application/json"}
    if token:
        headers[settings["jwtHeader"]] = settings["jwtPrefix"] + token
    if content_type:
        headers["Content-Type"] = content_type
    request = urllib.request.Request(settings["baseURL"] + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            return response.status, response.read(), dict(response.headers)
    except urllib.error.HTTPError as exc:
        return exc.code, exc.read(), dict(exc.headers)
    except urllib.error.URLError as exc:
        raise InstallError("application HTTP request failed") from exc


def decode_response(raw):
    try:
        document = json.loads(raw.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise InstallError("application returned invalid JSON") from exc
    if not isinstance(document, dict):
        fail("application returned non-object JSON")
    return document


def api_json(settings, path, token, body):
    raw_body = json.dumps(body, ensure_ascii=False, separators=(",", ":")).encode("utf-8")
    status, raw, _ = request_raw(
        settings, path, token=token, data=raw_body, content_type="application/json", method="POST"
    )
    document = decode_response(raw)
    if status != 200 or document.get("code") not in (0, 200) or document.get("success") is False:
        fail("application API rejected a verification request")
    return document.get("data") or {}


def multipart_payload(file_bytes, expected_hash=None):
    boundary = "----phase1-production-" + uuid.uuid4().hex
    chunks = [
        (f"--{boundary}\r\nContent-Disposition: form-data; name=\"file\"; "
         "filename=\"competency-phase1-import-20260810.xlsx\"\r\n"
         "Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet\r\n\r\n").encode(),
        file_bytes,
        b"\r\n",
    ]
    if expected_hash is not None:
        chunks.append(
            (f"--{boundary}\r\nContent-Disposition: form-data; name=\"expectedHash\"\r\n\r\n"
             f"{expected_hash}\r\n").encode("ascii")
        )
    chunks.append(f"--{boundary}--\r\n".encode("ascii"))
    return b"".join(chunks), "multipart/form-data; boundary=" + boundary


def api_upload(settings, path, token, file_bytes, expected_hash=None):
    body, content_type = multipart_payload(file_bytes, expected_hash)
    status, raw, _ = request_raw(
        settings, path, token=token, data=body, content_type=content_type, method="POST"
    )
    document = decode_response(raw)
    success = status == 200 and document.get("code") in (0, 200) and document.get("success") is not False
    return success, document.get("data") or {}, str(document.get("msg", ""))


def column_number(reference):
    match = re.match(r"^[A-Z]+", reference)
    if not match:
        fail("workbook cell reference is invalid")
    value = 0
    for letter in match.group(0):
        value = value * 26 + ord(letter) - 64
    return value


def workbook_rows(data, *, exact_contract=True):
    if not data.startswith(b"PK"):
        fail("workbook is not an XLSX package")
    try:
        with zipfile.ZipFile(BytesIO(data)) as archive:
            names = set(archive.namelist())
            required = {"xl/workbook.xml", "xl/_rels/workbook.xml.rels"}
            if not required.issubset(names):
                fail("workbook metadata is incomplete")
            shared = []
            if "xl/sharedStrings.xml" in names:
                root = ET.fromstring(archive.read("xl/sharedStrings.xml"))
                shared = ["".join(node.text or "" for node in item.findall(".//m:t", NS)) for item in root.findall("m:si", NS)]
            workbook = ET.fromstring(archive.read("xl/workbook.xml"))
            sheets = workbook.findall("m:sheets/m:sheet", NS)
            if len(sheets) != 1 or sheets[0].attrib.get("name") != "胜任力题目":
                fail("workbook must contain exactly the approved worksheet")
            relationships = ET.fromstring(archive.read("xl/_rels/workbook.xml.rels"))
            targets = {node.attrib["Id"]: node.attrib["Target"] for node in relationships.findall("p:Relationship", REL_NS)}
            rel_id = sheets[0].attrib.get("{" + NS["r"] + "}id")
            target = targets.get(rel_id, "").lstrip("/")
            if not target.startswith("xl/"):
                target = "xl/" + target
            if target not in names:
                fail("approved worksheet payload is missing")
            root = ET.fromstring(archive.read(target))
            dimension = root.find("m:dimension", NS)
            dimension_ref = "" if dimension is None else dimension.attrib.get("ref", "")
            if exact_contract and dimension_ref != "A1:J91":
                fail("workbook physical range is not A1:J91")
            output = []
            for row in root.findall("m:sheetData/m:row", NS):
                values = [""] * 10
                cells = row.findall("m:c", NS)
                if len(cells) != 10:
                    fail("every workbook row must contain exactly ten cells")
                for cell in cells:
                    column = column_number(cell.attrib.get("r", ""))
                    if not 1 <= column <= 10:
                        fail("workbook contains a cell outside the ten-column contract")
                    kind = cell.attrib.get("t")
                    value_node = cell.find("m:v", NS)
                    value = "" if value_node is None else value_node.text or ""
                    if kind == "s" and value:
                        value = shared[int(value)]
                    elif kind == "inlineStr":
                        value = "".join(node.text or "" for node in cell.findall(".//m:t", NS))
                    values[column - 1] = value
                output.append(values)
    except (zipfile.BadZipFile, ET.ParseError, IndexError, KeyError, ValueError) as exc:
        raise InstallError("workbook XML is invalid") from exc
    if (exact_contract and len(output) != 91) or len(output) < 1 or output[0] != EXPECTED_HEADERS:
        fail("workbook row count or header contract is invalid")
    return output


def validate_workbook(data):
    digest = sha256_bytes(data)
    if digest != EXPECTED_WORKBOOK_SHA256:
        fail("workbook SHA-256 is not approved")
    rows = workbook_rows(data)
    names_by_order = {str(row[6]): row[2] for row in DIMENSIONS}
    ids_by_order = {str(row[6]): row[0] for row in DIMENSIONS}
    normalized = {}
    totals = {"dimension": 0, "validity": 0, "dimensionForward": 0, "dimensionReverse": 0, "validityForward": 0}
    per_dimension = {row[0]: {"dimension": 0, "validity": 0} for row in DIMENSIONS}
    for number, raw in enumerate(rows[1:], 2):
        row = [str(value).strip() for value in raw]
        order, name, type_text, code, item_no, content, observation, direction_text, status_text, remark = row
        if order not in names_by_order or name != names_by_order[order]:
            fail("workbook dimension identity is invalid")
        if not SAFE_ID_RE.fullmatch(code) or code in normalized:
            fail("workbook question code is invalid or duplicated")
        if not re.fullmatch(r"[1-9]\d*", item_no) or not content or not observation:
            fail("workbook question payload is incomplete")
        question_type = {"维度题": "dimension", "效度题": "validity"}.get(type_text)
        direction = {"正向": "forward", "反向": "reverse"}.get(direction_text)
        status = {"启用": "0", "停用": "1"}.get(status_text)
        if question_type is None or direction is None or status is None:
            fail("workbook enum value is invalid")
        if status != "0" or (question_type == "validity" and direction != "forward"):
            fail("workbook phase-1 status or validity direction is invalid")
        dimension_id = ids_by_order[order]
        normalized[code] = [order, name, type_text, code, item_no, content, observation, direction_text, status_text, remark]
        totals[question_type] += 1
        totals[question_type + ("Forward" if direction == "forward" else "Reverse")] += 1
        per_dimension[dimension_id][question_type] += 1
    if totals != {"dimension": 80, "validity": 10, "dimensionForward": 62, "dimensionReverse": 18, "validityForward": 10}:
        fail("workbook type or direction distribution is invalid")
    if any(value != {"dimension": 8, "validity": 1} for value in per_dimension.values()):
        fail("workbook per-dimension distribution is invalid")
    return rows, normalized, totals


def runtime_marker(settings):
    status, raw, _ = request_raw(settings, "/health", method="GET", timeout=20)
    try:
        health = json.loads(raw.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise InstallError("health endpoint returned invalid JSON") from exc
    if status != 200 or not isinstance(health, dict) or health.get("status") != "ok":
        fail("application health endpoint is not healthy")
    disk = APP_DIR / "server"
    process = Path("/proc") / settings["pid"] / "exe"
    if not disk.is_file() or not process.exists():
        fail("server or process executable is missing")
    server_sha = sha256_file(disk)
    process_sha = sha256_file(process)
    if not SHA_RE.fullmatch(server_sha) or server_sha != process_sha:
        fail("deployed server and running process do not match")
    return {"serverSha256": server_sha, "processSha256": process_sha, "mainPid": int(settings["pid"])}


def database_snapshot(db, question_codes):
    target_codes = sql_list(question_codes)
    target_ids = sql_list(DIMENSION_IDS)
    return {
        "dimensions": db.scalar_int("SELECT COUNT(*) FROM el_competency_dimension"),
        "questions": db.scalar_int("SELECT COUNT(*) FROM el_qu"),
        "answers": db.scalar_int("SELECT COUNT(*) FROM el_qu_answer"),
        "repoRelations": db.scalar_int("SELECT COUNT(*) FROM el_qu_repo"),
        "activePapers": db.scalar_int("SELECT COUNT(*) FROM el_paper WHERE state=1"),
        "scopedDimensions": db.scalar_int(f"SELECT COUNT(*) FROM el_competency_dimension WHERE id IN ({target_ids})"),
        "scopedQuestions": db.scalar_int(f"SELECT COUNT(*) FROM el_qu WHERE question_code IN ({target_codes})"),
    }


def quoted_identifier(value):
    if not re.fullmatch(r"[a-z0-9_]+", value):
        fail("invalid database identifier")
    return "`" + value + "`"


def canonical_sql_value(column):
    identifier = quoted_identifier(column)
    return f"IF({identifier} IS NULL,'N',CONCAT('V',HEX({identifier})))"


def table_data_signature(db, table):
    if not db.table_exists(table):
        fail("required frozen association or result table is missing")
    columns = db.query(
        "SELECT column_name FROM information_schema.columns WHERE table_schema=DATABASE() "
        f"AND table_name={sql_string(table)} ORDER BY ordinal_position"
    ).splitlines()
    primary_key = db.query(
        "SELECT column_name FROM information_schema.statistics WHERE table_schema=DATABASE() "
        f"AND table_name={sql_string(table)} AND index_name='PRIMARY' ORDER BY seq_in_index"
    ).splitlines()
    if not columns or not primary_key or any(column not in columns for column in primary_key):
        fail("frozen association or result table lacks a stable primary key")
    values = ",".join(canonical_sql_value(column) for column in columns)
    ordering = ",".join(quoted_identifier(column) for column in primary_key)
    rows = db.query(
        f"SELECT CONCAT_WS('|',{values}) FROM {quoted_identifier(table)} ORDER BY {ordering}"
    )
    return {
        "rows": 0 if not rows else len(rows.splitlines()),
        "sha256": sha256_bytes(rows.encode("utf-8")),
    }


def frozen_association_result_signatures(db):
    return {table: table_data_signature(db, table) for table in FROZEN_ASSOCIATION_RESULT_TABLES}


def legacy_dimension_immutable_sha(db, legacy_ids):
    if len(legacy_ids) != 48 or len(set(legacy_ids)) != 48:
        fail("legacy dimension identity set is invalid")
    columns = (
        "id", "code", "vird_level", "applicable_category", "core_meaning",
        "status", "create_time", "update_time",
    )
    values = ",".join(canonical_sql_value(column) for column in columns)
    rows = db.query(
        f"SELECT CONCAT_WS('|',{values}) FROM el_competency_dimension WHERE id IN ("
        + sql_list(legacy_ids) + ") ORDER BY id"
    )
    if not rows or len(rows.splitlines()) != 48:
        fail("legacy dimension immutable receipt is incomplete")
    return sha256_bytes(rows.encode("utf-8"))


def legacy_dimension_receipt(db):
    rows = db.query(
        "SELECT CONCAT_WS('\\t',HEX(id),HEX(code),HEX(name),display_order) "
        "FROM el_competency_dimension ORDER BY display_order,id"
    ).splitlines()
    receipt = []
    observed_collisions = []
    for line in rows:
        parts = line.split("\t")
        if len(parts) != 4:
            fail("legacy dimension receipt row is invalid")
        try:
            identity = bytes.fromhex(parts[0]).decode("utf-8")
            code = bytes.fromhex(parts[1]).decode("utf-8")
            name = bytes.fromhex(parts[2]).decode("utf-8")
            order = int(parts[3])
        except (UnicodeDecodeError, ValueError) as exc:
            raise InstallError("legacy dimension receipt row is invalid") from exc
        receipt.append({"id": identity, "name": name, "displayOrder": order})
        if name in DIMENSION_NAMES:
            observed_collisions.append((identity, code, name, order))
    ids = tuple(row["id"] for row in receipt)
    orders = tuple(row["displayOrder"] for row in receipt)
    if len(receipt) != 48 or len(set(ids)) != 48 or orders != tuple(range(1, 49)):
        fail("legacy dimension order signature is not exact 48/48/min1/max48")
    if any(row["name"].endswith(ARCHIVAL_SUFFIX) for row in receipt):
        fail("a legacy dimension name already uses the archival suffix")
    if tuple(observed_collisions) != LEGACY_COLLISIONS:
        fail("legacy dimension name collisions are not the exact approved eight rows")
    return {
        "rows": receipt,
        "immutableSha256": legacy_dimension_immutable_sha(db, ids),
    }


def old_signatures(db, question_codes):
    target_codes = sql_list(question_codes)
    question_rows = db.query(
        "SELECT CONCAT_WS('|',HEX(id),qu_type,level,HEX(image),HEX(content),HEX(remark),HEX(analysis),HEX(title),"
        "HEX(COALESCE(question_code,'')),HEX(COALESCE(dimension_id,'')),COALESCE(dimension_item_no,-1),"
        "HEX(COALESCE(observation_point,'')),HEX(COALESCE(scoring_direction,'')),"
        "HEX(COALESCE(competency_question_type,'')),question_status,HEX(COALESCE(create_time,'')),"
        "HEX(COALESCE(update_time,''))) FROM el_qu "
        f"WHERE question_code IS NULL OR question_code NOT IN ({target_codes}) ORDER BY id"
    )
    return {
        "oldQuestionsSha256": sha256_bytes(question_rows.encode()),
        "frozenExamAssociationsResults": frozen_association_result_signatures(db),
    }


def preflight_database(db, question_codes):
    required = [
        "el_competency_dimension", "el_qu", "el_qu_answer", "el_qu_repo", "el_paper",
        "el_exam_competency_dimension", "el_exam_competency_question", "el_paper_qu", "el_paper_qu_answer",
        *FROZEN_ASSOCIATION_RESULT_TABLES,
    ]
    if any(not db.table_exists(table) for table in required):
        fail("required production table is missing")
    if db.query("SELECT DATABASE()") != "element":
        fail("MySQL client is not using element")
    target_ids = sql_list(DIMENSION_IDS)
    target_codes = sql_list(DIMENSION_CODES)
    question_codes_sql = sql_list(question_codes)
    if db.scalar_int(f"SELECT COUNT(*) FROM el_competency_dimension WHERE id IN ({target_ids}) OR code IN ({target_codes})") != 0:
        fail("phase-1 dimension IDs or codes already exist")
    if db.scalar_int(
        f"SELECT COUNT(*) FROM el_qu WHERE dimension_id IN ({target_ids}) OR question_code IN ({target_codes}) OR question_code IN ({question_codes_sql})"
    ) != 0:
        fail("questions already use phase-1 identities or workbook codes")
    if db.scalar_int("SELECT COUNT(*) FROM el_paper WHERE state=1") != 0:
        fail("active papers exist")
    legacy = legacy_dimension_receipt(db)
    marker_count = 0
    if db.table_exists("el_competency_migration"):
        marker_count = db.scalar_int(
            "SELECT COUNT(*) FROM el_competency_migration WHERE migration_key="
            + sql_string("competency-009-phase1-identity-reset")
        )
    if marker_count != 0:
        fail("migration 009 was already executed or marked")
    before = database_snapshot(db, question_codes)
    signatures = old_signatures(db, question_codes)
    return before, signatures, legacy, {"existingDimensions": before["dimensions"], "existingQuestions": before["questions"], "migration009Marker": marker_count}


def create_backup(db, workbook_path, workbook_bytes, before_receipt):
    stamp = uuid.uuid4().hex[:16]
    root = BACKUP_DIR / ("phase1_content_20261009_" + stamp)
    BACKUP_DIR.mkdir(mode=0o700, parents=True, exist_ok=True)
    os.chmod(BACKUP_DIR, stat.S_IMODE(BACKUP_DIR.stat().st_mode))
    root.mkdir(mode=0o700)
    os.chmod(root, 0o700)
    if stat.S_IMODE(root.stat().st_mode) != 0o700:
        fail("backup directory mode is not 0700")
    dump_path = root / "element.sql.gz"
    dump_argv = [
        "mysqldump", "--defaults-extra-file=" + str(db.client_file), "--single-transaction", "--quick",
        "--skip-lock-tables", "--routines", "--triggers", "--events", "--hex-blob",
        "--set-gtid-purged=OFF", "--no-tablespaces", db.database,
    ]
    try:
        process = subprocess.Popen(dump_argv, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
    except OSError as exc:
        raise InstallError("mysqldump could not start") from exc
    assert process.stdout is not None
    with gzip.open(dump_path, "wb", compresslevel=9) as output:
        shutil.copyfileobj(process.stdout, output, length=1024 * 1024)
    process.wait(timeout=600)
    if process.returncode != 0:
        fail("mysqldump failed")
    workbook_copy = root / "competency-phase1-import-20260810.xlsx"
    workbook_copy.write_bytes(workbook_bytes)
    receipt_path = root / "before-receipt.json"
    receipt_path.write_text(json.dumps(before_receipt, ensure_ascii=True, sort_keys=True, separators=(",", ":")) + "\n", encoding="utf-8")
    files = [dump_path, workbook_copy, receipt_path]
    for path in files:
        os.chmod(path, 0o600)
    manifest = root / "SHA256SUMS"
    manifest.write_text("".join(f"{sha256_file(path)}  {path.name}\n" for path in files), encoding="ascii")
    os.chmod(manifest, 0o600)
    with gzip.open(dump_path, "rb") as stream:
        while stream.read(1024 * 1024):
            pass
    if sha256_file(workbook_copy) != EXPECTED_WORKBOOK_SHA256:
        fail("backup workbook hash verification failed")
    for line in manifest.read_text(encoding="ascii").splitlines():
        digest, name = line.split("  ", 1)
        if not SHA_RE.fullmatch(digest) or sha256_file(root / name) != digest:
            fail("backup hash verification failed")
    if any(stat.S_IMODE(path.stat().st_mode) != 0o600 for path in files + [manifest]):
        fail("backup file mode is not 0600")
    return root


def install_compatibility_dimensions(db, legacy):
    legacy_ids = tuple(row["id"] for row in legacy["rows"])
    values = []
    for row in DIMENSIONS:
        values.append("(" + ",".join(sql_string(value) for value in row[:6]) + f",{row[6]},{row[7]},NOW(),NOW())")
    exact_collisions = " OR ".join(
        "(id={0} AND code={1} AND name={2} AND display_order={3})".format(
            sql_string(row[0]), sql_string(row[1]), sql_string(row[2]), row[3]
        )
        for row in LEGACY_COLLISIONS
    )
    rename_cases = " ".join(
        "WHEN " + sql_string(row[0]) + " THEN " + sql_string(row[2] + ARCHIVAL_SUFFIX)
        for row in LEGACY_COLLISIONS
    )
    guard = (
        "(SELECT COUNT(*) FROM el_competency_dimension)=48 AND "
        "(SELECT COUNT(DISTINCT display_order) FROM el_competency_dimension)=48 AND "
        "(SELECT MIN(display_order) FROM el_competency_dimension)=1 AND "
        "(SELECT MAX(display_order) FROM el_competency_dimension)=48 AND "
        "(SELECT COUNT(*) FROM el_competency_dimension WHERE id IN (" + sql_list(legacy_ids) + "))=48 AND "
        "(SELECT COUNT(*) FROM el_competency_dimension WHERE name IN (" + sql_list(DIMENSION_NAMES) + "))=8 AND "
        f"(SELECT COUNT(*) FROM el_competency_dimension WHERE {exact_collisions})=8 AND "
        "(SELECT COUNT(*) FROM el_competency_dimension WHERE RIGHT(name,CHAR_LENGTH("
        + sql_string(ARCHIVAL_SUFFIX) + "))=" + sql_string(ARCHIVAL_SUFFIX) + ")=0"
    )
    sql = (
        "SET NAMES utf8mb4;\nSET SESSION SQL_SAFE_UPDATES=1;\n"
        "SET SESSION sql_mode=CONCAT_WS(',',@@SESSION.sql_mode,'STRICT_ALL_TABLES');\n"
        "START TRANSACTION;\n"
        "SELECT id FROM el_competency_dimension ORDER BY id FOR UPDATE;\n"
        "CREATE TEMPORARY TABLE phase1_compatibility_guard (ok TINYINT NOT NULL);\n"
        "INSERT INTO phase1_compatibility_guard(ok) SELECT IF(" + guard + ",1,NULL);\n"
        "UPDATE el_competency_dimension SET display_order=display_order+100,"
        "name=CASE id " + rename_cases + " ELSE name END,update_time=update_time "
        "WHERE id IN (" + sql_list(legacy_ids) + ");\n"
        "INSERT INTO el_competency_dimension "
        "(id,code,name,vird_level,applicable_category,core_meaning,display_order,status,create_time,update_time) VALUES\n"
        + ",\n".join(values)
        + ";\nCOMMIT;\n"
    )
    db.execute(sql)


def expected_question_rows(normalized):
    by_order = {str(row[6]): row[0] for row in DIMENSIONS}
    output = {}
    for code, row in normalized.items():
        output[code] = {
            "dimensionID": by_order[row[0]],
            "questionType": {"维度题": "dimension", "效度题": "validity"}[row[2]],
            "itemNo": int(row[4]),
            "content": row[5],
            "observation": row[6],
            "direction": {"正向": "forward", "反向": "reverse"}[row[7]],
            "status": {"启用": 0, "停用": 1}[row[8]],
            "remark": row[9],
        }
    return output


def verify_dimensions(db):
    rows = db.query(
        "SELECT CONCAT_WS('\\t',id,code,HEX(name),HEX(vird_level),HEX(applicable_category),HEX(core_meaning),display_order,status) "
        "FROM el_competency_dimension WHERE id IN (" + sql_list(DIMENSION_IDS) + ") ORDER BY display_order"
    ).splitlines()
    actual = []
    for line in rows:
        parts = line.split("\t")
        if len(parts) != 8:
            fail("dimension verification row is invalid")
        actual.append((parts[0], parts[1], *(bytes.fromhex(value).decode("utf-8") for value in parts[2:6]), int(parts[6]), int(parts[7])))
    if actual != DIMENSIONS:
        fail("scoped dimension rows differ from migration 009")


def archived_legacy_rows(legacy):
    collision_names = {row[0]: row[2] + ARCHIVAL_SUFFIX for row in LEGACY_COLLISIONS}
    return [
        {
            "id": row["id"],
            "name": collision_names.get(row["id"], row["name"]),
            "displayOrder": row["displayOrder"] + 100,
        }
        for row in legacy["rows"]
    ]


def load_legacy_name_order_rows(db, legacy_ids):
    rows = db.query(
        "SELECT CONCAT_WS('\\t',HEX(id),HEX(name),display_order) FROM el_competency_dimension WHERE id IN ("
        + sql_list(legacy_ids) + ") ORDER BY display_order,id"
    ).splitlines()
    actual = []
    for line in rows:
        parts = line.split("\t")
        if len(parts) != 3:
            fail("legacy dimension name/order row is invalid")
        try:
            actual.append({
                "id": bytes.fromhex(parts[0]).decode("utf-8"),
                "name": bytes.fromhex(parts[1]).decode("utf-8"),
                "displayOrder": int(parts[2]),
            })
        except (UnicodeDecodeError, ValueError) as exc:
            raise InstallError("legacy dimension name/order row is invalid") from exc
    return actual


def verify_archived_legacy_dimensions(db, legacy):
    legacy_ids = tuple(row["id"] for row in legacy["rows"])
    if load_legacy_name_order_rows(db, legacy_ids) != archived_legacy_rows(legacy):
        fail("legacy dimensions do not have exact archived names and orders 101..148")
    if legacy_dimension_immutable_sha(db, legacy_ids) != legacy["immutableSha256"]:
        fail("legacy dimension immutable columns changed")
    return {
        "archivedLegacyDimensions": 48,
        "renamedLegacyNames": 8,
        "immutableSha256Unchanged": True,
    }


def load_scoped_questions(db, expected):
    codes = tuple(expected)
    rows = db.query(
        "SELECT CONCAT_WS('\\t',id,question_code,dimension_id,competency_question_type,dimension_item_no,"
        "HEX(content),HEX(observation_point),scoring_direction,question_status,HEX(remark)) FROM el_qu "
        "WHERE question_code IN (" + sql_list(codes) + ") ORDER BY question_code"
    ).splitlines()
    actual = {}
    ids = []
    for line in rows:
        parts = line.split("\t")
        if len(parts) != 10 or not SAFE_ID_RE.fullmatch(parts[0]):
            fail("scoped question row is invalid")
        code = parts[1]
        ids.append(parts[0])
        actual[code] = {
            "dimensionID": parts[2], "questionType": parts[3], "itemNo": int(parts[4]),
            "content": bytes.fromhex(parts[5]).decode("utf-8"),
            "observation": bytes.fromhex(parts[6]).decode("utf-8"),
            "direction": parts[7], "status": int(parts[8]),
            "remark": bytes.fromhex(parts[9]).decode("utf-8"),
        }
    if actual != expected or len(ids) != 90 or len(set(ids)) != 90:
        fail("scoped question rows differ from approved workbook")
    return ids


def scoped_relations(db, question_ids):
    if not question_ids:
        return {"answers": 0, "repoRelations": 0, "examAssociations": 0, "paperReferences": 0, "bookReferences": 0, "dimensionResults": 0}
    ids = sql_list(question_ids)
    dimension_ids = sql_list(DIMENSION_IDS)
    counts = {
        "answers": db.scalar_int(f"SELECT COUNT(*) FROM el_qu_answer WHERE qu_id IN ({ids})"),
        "repoRelations": db.scalar_int(f"SELECT COUNT(*) FROM el_qu_repo WHERE qu_id IN ({ids})"),
        "examAssociations": db.scalar_int(
            f"SELECT (SELECT COUNT(*) FROM el_exam_competency_dimension WHERE dimension_id IN ({dimension_ids})) + "
            f"(SELECT COUNT(*) FROM el_exam_competency_question WHERE source_qu_id IN ({ids}))"
        ),
        "paperReferences": db.scalar_int(
            f"SELECT (SELECT COUNT(*) FROM el_paper_qu WHERE qu_id IN ({ids})) + "
            f"(SELECT COUNT(*) FROM el_paper_qu_answer WHERE qu_id IN ({ids}))"
        ),
        "bookReferences": 0,
        "dimensionResults": 0,
    }
    if db.table_exists("el_user_book") and db.column_exists("el_user_book", "qu_id"):
        counts["bookReferences"] = db.scalar_int(f"SELECT COUNT(*) FROM el_user_book WHERE qu_id IN ({ids})")
    if db.table_exists("el_competency_dimension_result") and db.column_exists("el_competency_dimension_result", "dimension_id"):
        counts["dimensionResults"] += db.scalar_int(
            f"SELECT COUNT(*) FROM el_competency_dimension_result WHERE dimension_id IN ({dimension_ids})"
        )
    if db.table_exists("el_competency_result_run_dimension") and db.column_exists("el_competency_result_run_dimension", "source_dimension_id"):
        counts["dimensionResults"] += db.scalar_int(
            f"SELECT COUNT(*) FROM el_competency_result_run_dimension WHERE source_dimension_id IN ({dimension_ids})"
        )
    return counts


def verify_database(db, expected, before, signatures, legacy):
    verify_dimensions(db)
    legacy_evidence = verify_archived_legacy_dimensions(db, legacy)
    question_ids = load_scoped_questions(db, expected)
    relations = scoped_relations(db, question_ids)
    if any(relations.values()):
        fail("scoped questions have unexpected relations or runtime references")
    counts = db.query(
        "SELECT CONCAT_WS('|',COUNT(*),COUNT(DISTINCT question_code),"
        "SUM(competency_question_type='dimension'),SUM(competency_question_type='validity'),"
        "SUM(competency_question_type='dimension' AND scoring_direction='forward'),"
        "SUM(competency_question_type='dimension' AND scoring_direction='reverse'),"
        "SUM(competency_question_type='validity' AND scoring_direction='forward'),SUM(question_status=0)) "
        "FROM el_qu WHERE question_code IN (" + sql_list(tuple(expected)) + ")"
    )
    if counts != "90|90|80|10|62|18|10|90":
        fail("scoped question distribution is invalid")
    distribution = db.query(
        "SELECT CONCAT_WS('|',COUNT(*),MIN(dimension_count),MAX(dimension_count),SUM(dimension_count),"
        "MIN(validity_count),MAX(validity_count),SUM(validity_count)) FROM ("
        "SELECT d.id,SUM(q.competency_question_type='dimension') dimension_count,"
        "SUM(q.competency_question_type='validity') validity_count FROM el_competency_dimension d "
        "LEFT JOIN el_qu q ON q.dimension_id=d.id WHERE d.id IN (" + sql_list(DIMENSION_IDS) + ") GROUP BY d.id) scoped"
    )
    if distribution != "10|8|8|80|1|1|10":
        fail("scoped per-dimension distribution is invalid")
    after = database_snapshot(db, tuple(expected))
    if after["dimensions"] != before["dimensions"] + 10 or after["questions"] != before["questions"] + 90:
        fail("global dimension or question count drifted unexpectedly")
    for key in ("answers", "repoRelations"):
        if after[key] != before[key]:
            fail("global relation count drifted unexpectedly")
    if after["activePapers"] != 0 or after["scopedDimensions"] != 10 or after["scopedQuestions"] != 90:
        fail("scoped or active-paper counts are invalid")
    if old_signatures(db, tuple(expected)) != signatures:
        fail("old questions or frozen exam associations/results changed")
    return after, question_ids, relations, legacy_evidence


def normalize_export(data):
    rows = workbook_rows(data, exact_contract=False)
    return {row[3].strip(): [str(value).strip() for value in row] for row in rows[1:] if row[3].strip()}


def verify_apis(settings, token, normalized, file_bytes):
    dimensions = api_json(settings, "/exam/api/competency/dimensions/list", token, {})
    by_id = {str(row.get("id", "")): row for row in dimensions if isinstance(row, dict)}
    for expected in DIMENSIONS:
        row = by_id.get(expected[0])
        if row is None:
            fail("dimension API omitted an installed dimension")
        actual = (
            row.get("id"), row.get("code"), row.get("name"), row.get("virdLevel"),
            row.get("applicableCategory"), row.get("coreMeaning"), row.get("displayOrder"), row.get("status"),
        )
        if actual != expected or row.get("questionCount") != 8:
            fail("dimension API scoped evidence is invalid")
    expected_by_dimension = {dimension_id: set() for dimension_id in DIMENSION_IDS}
    order_to_id = {str(row[6]): row[0] for row in DIMENSIONS}
    for code, row in normalized.items():
        expected_by_dimension[order_to_id[row[0]]].add(code)
    paging_evidence = {}
    for dimension_id in DIMENSION_IDS:
        page = api_json(
            settings, "/exam/api/competency/questions/paging", token,
            {"current": 1, "size": 20, "dimensionId": dimension_id, "params": {}},
        )
        records = page.get("records") or []
        codes = {str(row.get("questionCode", "")) for row in records if isinstance(row, dict)}
        if page.get("total") != 9 or len(records) != 9 or codes != expected_by_dimension[dimension_id]:
            fail("question paging scoped evidence is invalid")
        paging_evidence[dimension_id] = len(records)
    status, exported, headers = request_raw(settings, "/exam/api/competency/questions/export", token=token, method="GET")
    if status != 200 or "spreadsheetml.sheet" not in headers.get("Content-Type", "") or not exported.startswith(b"PK"):
        fail("question export did not return XLSX bytes")
    exported_rows = normalize_export(exported)
    for code, expected_row in normalized.items():
        if exported_rows.get(code) != expected_row:
            fail("exported workbook does not contain exact normalized source content")
    return {"dimensions": 10, "questionPagingByDimension": paging_evidence, "exportScopedRows": 90}


def rollback(db, expected, before, signatures, legacy):
    result = {"attempted": True, "completed": False, "refused": False}
    target_codes = tuple(expected)
    current_dimensions = db.scalar_int(
        "SELECT COUNT(*) FROM el_competency_dimension WHERE id IN (" + sql_list(DIMENSION_IDS) + ")"
    )
    current_questions = db.scalar_int(
        "SELECT COUNT(*) FROM el_qu WHERE question_code IN (" + sql_list(target_codes) + ")"
    )
    if current_dimensions not in (0, 10) or current_questions not in (0, 90):
        result.update({"refused": True, "reason": "unexpected_scoped_row_drift"})
        return result
    question_ids = []
    if current_dimensions == 10:
        try:
            verify_dimensions(db)
        except InstallError:
            result.update({"refused": True, "reason": "unexpected_scoped_dimension_drift"})
            return result
    if current_questions == 90:
        try:
            question_ids = load_scoped_questions(db, expected)
        except InstallError:
            result.update({"refused": True, "reason": "unexpected_scoped_question_drift"})
            return result
    current = database_snapshot(db, target_codes)
    allowed_dimension_delta = current_dimensions
    allowed_question_delta = current_questions
    if current["dimensions"] != before["dimensions"] + allowed_dimension_delta or current["questions"] != before["questions"] + allowed_question_delta:
        result.update({"refused": True, "reason": "unexpected_global_row_drift"})
        return result
    if old_signatures(db, target_codes) != signatures:
        result.update({"refused": True, "reason": "old_question_or_frozen_row_drift"})
        return result
    legacy_ids = tuple(row["id"] for row in legacy["rows"])
    if legacy_dimension_immutable_sha(db, legacy_ids) != legacy["immutableSha256"]:
        result.update({"refused": True, "reason": "legacy_immutable_row_drift"})
        return result
    current_legacy = load_legacy_name_order_rows(db, legacy_ids)
    archived_legacy = archived_legacy_rows(legacy)
    if current_legacy not in (legacy["rows"], archived_legacy):
        result.update({"refused": True, "reason": "unexpected_legacy_archive_drift"})
        return result
    relations = scoped_relations(db, question_ids)
    protected = relations["examAssociations"] + relations["paperReferences"] + relations["bookReferences"] + relations["dimensionResults"]
    if protected:
        result.update({"refused": True, "reason": "runtime_or_exam_references_exist", "referenceCounts": relations})
        return result
    if relations["answers"] or relations["repoRelations"]:
        result.update({"refused": True, "reason": "unexpected_question_relations_exist", "referenceCounts": relations})
        return result
    if current["answers"] != before["answers"] + relations["answers"] or current["repoRelations"] != before["repoRelations"] + relations["repoRelations"]:
        result.update({"refused": True, "reason": "unexpected_relation_drift"})
        return result
    ids_sql = sql_list(question_ids) if question_ids else "''"
    restore_legacy = current_legacy == archived_legacy
    restore_names = " ".join(
        "WHEN " + sql_string(row[0]) + " THEN " + sql_string(row[2])
        for row in LEGACY_COLLISIONS
    )
    restore_sql = ""
    if restore_legacy:
        restore_sql = (
            "UPDATE el_competency_dimension SET name=CASE id " + restore_names
            + " ELSE name END,display_order=display_order-100,update_time=update_time WHERE id IN ("
            + sql_list(legacy_ids) + ");\n"
        )
    db.execute(
        "SET SESSION SQL_SAFE_UPDATES=1;\nSTART TRANSACTION;\n"
        f"DELETE FROM el_qu_answer WHERE qu_id IN ({ids_sql}) AND id IS NOT NULL;\n"
        f"DELETE FROM el_qu_repo WHERE qu_id IN ({ids_sql}) AND id IS NOT NULL;\n"
        f"DELETE FROM el_qu WHERE id IN ({ids_sql}) AND id IS NOT NULL;\n"
        "DELETE FROM el_competency_dimension WHERE id IN (" + sql_list(DIMENSION_IDS) + ") AND id IS NOT NULL;\n"
        + restore_sql
        + "COMMIT;\n"
    )
    after = database_snapshot(db, target_codes)
    restored_legacy = legacy_dimension_receipt(db)
    if after != before or restored_legacy != legacy or old_signatures(db, target_codes) != signatures:
        fail("rollback verification failed")
    result.update({
        "completed": True,
        "scopedQuestionsRemoved": current_questions,
        "scopedDimensionsRemoved": current_dimensions,
        "legacyDimensionsRestored": 48 if restore_legacy else 0,
        "legacyNamesRestored": 8 if restore_legacy else 0,
        "beforeSnapshotAndSignaturesRestored": True,
    })
    return result


def main():
    receipt = {
        "status": "failed",
        "workbookSha256": EXPECTED_WORKBOOK_SHA256,
        "backupRoot": None,
    }
    stage = "arguments"
    settings = None
    db = None
    mysql_file = None
    token_key = ""
    mutated = False
    expected = None
    before = None
    signatures = None
    legacy = None
    try:
        if len(sys.argv) > 2:
            fail("usage: production-phase1-content-install-20261009.py [absolute-workbook-path]")
        workbook_path = Path(sys.argv[1]) if len(sys.argv) == 2 else DEFAULT_WORKBOOK
        if not workbook_path.is_absolute() or workbook_path.is_symlink() or not workbook_path.is_file():
            fail("workbook path must be an absolute regular non-symlink file")
        if os.geteuid() != 0:
            fail("installer must run as root")
        for executable in ("mysql", "mysqldump", "redis-cli", "systemctl"):
            if shutil.which(executable) is None:
                fail("required host executable is missing")
        stage = "workbook"
        workbook_bytes = workbook_path.read_bytes()
        _, normalized, workbook_totals = validate_workbook(workbook_bytes)
        expected = expected_question_rows(normalized)

        stage = "runtime_preflight"
        settings = load_settings()
        runtime_before = runtime_marker(settings)
        mysql_file, database = mysql_client_from_dsn(settings["dsn"])
        db = MySQL(mysql_file, database)

        stage = "database_preflight"
        before, signatures, legacy, production_baseline = preflight_database(db, tuple(expected))
        before_receipt = {
            "schema": "production-phase1-content-before-v2",
            "environment": "production",
            "workbookSha256": EXPECTED_WORKBOOK_SHA256,
            "workbookContract": {"rows": 91, "columns": 10, **workbook_totals},
            "globalCounts": before,
            "oldRowSignatures": signatures,
            "legacyDimensions": legacy,
            "productionBaseline": production_baseline,
            "runtime": runtime_before,
            "dsnSource": settings["dsnSource"],
        }

        stage = "backup"
        backup_root = create_backup(db, workbook_path, workbook_bytes, before_receipt)
        receipt["backupRoot"] = str(backup_root)

        stage = "dimension_insert"
        mutated = True
        install_compatibility_dimensions(db, legacy)
        verify_dimensions(db)
        verify_archived_legacy_dimensions(db, legacy)
        if db.table_exists("el_competency_migration") and db.scalar_int(
            "SELECT COUNT(*) FROM el_competency_migration WHERE migration_key="
            + sql_string("competency-009-phase1-identity-reset")
        ) != production_baseline["migration009Marker"]:
            fail("migration 009 marker changed")

        stage = "admin_session"
        token_id = "phase1-content-production-" + uuid.uuid4().hex
        token_key = "login_tokens:" + token_id
        redis_call(settings, "EXISTS", token_key, expect="0")
        now_ms = int(dt.datetime.now(dt.timezone.utc).timestamp() * 1000)
        login_user = {
            "userId": 1, "token": token_id, "loginTime": now_ms,
            "expireTime": now_ms + 1800000, "permissions": ["*:*:*"],
            "roles": ["admin"], "user": None,
        }
        redis_call(
            settings, "SET", token_key,
            json.dumps(login_user, separators=(",", ":")), "NX", "EX", "1800", expect="OK",
        )
        token = make_admin_token(settings["jwtSecret"], settings["jwtClaim"], token_id)

        stage = "import_preview"
        preview_ok, preview, _ = api_upload(
            settings, "/exam/api/competency/questions/import-preview", token, workbook_bytes
        )
        if not preview_ok or preview.get("successCount") != 90 or preview.get("errorCount") != 0 or preview.get("sha256") != EXPECTED_WORKBOOK_SHA256:
            fail("official import preview did not approve exactly 90 rows")

        stage = "import"
        import_ok, imported, _ = api_upload(
            settings, "/exam/api/competency/questions/import", token, workbook_bytes, EXPECTED_WORKBOOK_SHA256
        )
        if not import_ok or imported.get("importedCount") != 90 or imported.get("sha256") != EXPECTED_WORKBOOK_SHA256:
            fail("official import did not insert exactly 90 rows")

        stage = "database_verification"
        after, question_ids, relation_counts, legacy_evidence = verify_database(
            db, expected, before, signatures, legacy
        )

        stage = "api_verification"
        api_evidence = verify_apis(settings, token, normalized, workbook_bytes)

        stage = "repeat_fail_closed"
        counts_before_repeat = database_snapshot(db, tuple(expected))
        repeat_preview_ok, repeat_preview, _ = api_upload(
            settings, "/exam/api/competency/questions/import-preview", token, workbook_bytes
        )
        if not repeat_preview_ok or repeat_preview.get("successCount") != 0 or repeat_preview.get("errorCount") != 90:
            fail("repeated preview did not expose exactly 90 existing-row errors")
        repeat_import_ok, _, repeat_message = api_upload(
            settings, "/exam/api/competency/questions/import", token, workbook_bytes, EXPECTED_WORKBOOK_SHA256
        )
        if repeat_import_ok or "导入数据存在错误" not in repeat_message:
            fail("repeated import did not fail closed")
        if database_snapshot(db, tuple(expected)) != counts_before_repeat:
            fail("repeated preview/import changed scoped or global counts")

        stage = "final_runtime"
        runtime_after = runtime_marker(settings)
        if runtime_after["serverSha256"] != runtime_before["serverSha256"] or runtime_after["processSha256"] != runtime_before["processSha256"]:
            fail("server or process SHA changed during installation")
        verify_archived_legacy_dimensions(db, legacy)
        if old_signatures(db, tuple(expected)) != signatures:
            fail("old questions or frozen exam associations/results changed during final verification")

        receipt = {
            "status": "success",
            "backupRoot": str(backup_root),
            "workbookSha256": EXPECTED_WORKBOOK_SHA256,
            "before": {"scopedDimensions": before["scopedDimensions"], "scopedQuestions": before["scopedQuestions"]},
            "after": {"scopedDimensions": after["scopedDimensions"], "scopedQuestions": after["scopedQuestions"]},
            "archivedLegacyDimensions": 48,
            "renamedLegacyNames": 8,
            "oldExamSnapshotsUnchanged": True,
            "globalCounts": {
                "before": before,
                "after": after,
                "verifiedDelta": {"dimensions": 10, "questions": 90, "answers": 0, "repoRelations": 0},
            },
            "verification": {
                "questions": {"total": 90, "dimension": 80, "validity": 10, "dimensionForward": 62, "dimensionReverse": 18, "validityForward": 10, "enabled": 90},
                "relations": relation_counts,
                "api": api_evidence,
                "repeatPreviewErrors": 90,
                "repeatImportRejected": True,
                **legacy_evidence,
                "oldQuestionsUnchanged": True,
                "oldExamSnapshotsUnchanged": True,
                "migration009Executed": False,
                "migration009MarkerCreated": False,
            },
            "serverSha256": runtime_after["serverSha256"],
            "processSha256": runtime_after["processSha256"],
        }
    except Exception as exc:
        receipt.update({
            "status": "failed",
            "failedStage": stage,
            "error": type(exc).__name__,
            "errorDetail": str(exc),
        })
        if mutated and db is not None and expected is not None and before is not None and signatures is not None and legacy is not None:
            try:
                receipt["rollback"] = rollback(db, expected, before, signatures, legacy)
            except Exception as rollback_exc:
                receipt["rollback"] = {"attempted": True, "completed": False, "refused": False, "error": type(rollback_exc).__name__}
    finally:
        if token_key and settings is not None:
            try:
                if redis_call(settings, "EXISTS", token_key) == "1":
                    redis_call(settings, "DEL", token_key, expect="1")
                redis_call(settings, "EXISTS", token_key, expect="0")
                receipt["syntheticAdminSessionCleaned"] = True
            except Exception:
                receipt["syntheticAdminSessionCleaned"] = False
                receipt["status"] = "failed"
        if mysql_file is not None:
            try:
                mysql_file.unlink(missing_ok=True)
            except OSError:
                receipt["status"] = "failed"
                receipt["mysqlClientCleaned"] = False
            else:
                receipt["mysqlClientCleaned"] = not mysql_file.exists()
        print(json.dumps(receipt, ensure_ascii=True, sort_keys=True, separators=(",", ":")))
    return 0 if receipt.get("status") == "success" else 1


if __name__ == "__main__":
    sys.exit(main())
