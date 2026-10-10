#!/usr/bin/env python3
"""Prepare and verify one retained production 00401 browser-test closure.

This script is intentionally split into prepare/finalize phases. The browser
performs candidate registration, paper creation, 90 answer saves, submission,
result inspection, report generation, viewing, and native download between the
phases. Successful runs retain all business data and remove only the ephemeral
administrator Redis session.
"""

import importlib.util
import json
import re
import sys
import time
import uuid
from pathlib import Path

BASE_PATH = Path(__file__).with_name("production-phase1-runtime-e2e-20261009.py")
ID_RE = re.compile(r"^[A-Za-z0-9-]{1,64}$")
MARKER_RE = re.compile(r"^TEST-BROWSER-00401-[0-9]{14}$")
PHONE_RE = re.compile(r"^199[0-9]{8}$")


def load_base():
    spec = importlib.util.spec_from_file_location("production_phase1_runtime", BASE_PATH)
    if spec is None or spec.loader is None:
        raise RuntimeError("production 00401 runtime helper is unavailable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def emit(document, exit_code=0):
    print(json.dumps(document, ensure_ascii=True, sort_keys=True, separators=(",", ":")))
    return exit_code


def prepare(base, marker, phone):
    if not MARKER_RE.fullmatch(marker) or not PHONE_RE.fullmatch(phone):
        raise RuntimeError("invalid retained-test ownership marker")
    settings = base.load_settings()
    mysql_file, database = base.mysql_client_from_dsn(settings["dsn"])
    db = base.MySQL(mysql_file, database)
    exam_id = ""
    redis_key = ""
    admin_token = ""
    try:
        if db.scalar_int("SELECT COUNT(*) FROM el_paper WHERE state=1") != 0:
            raise RuntimeError("active paper preflight failed")
        if db.scalar_int("SELECT COUNT(*) FROM el_exam WHERE title=" + base.sql_literal(marker)) != 0:
            raise RuntimeError("retained-test marker already exists")
        token_id = "prod-phase1-browser-" + uuid.uuid4().hex
        redis_key = "login_tokens:" + token_id
        base.redis_call(settings, "EXISTS", redis_key, expect="0")
        now_ms = int(time.time() * 1000)
        login_user = {
            "userId": 1,
            "token": token_id,
            "loginTime": now_ms,
            "expireTime": now_ms + 3600000,
            "permissions": ["*:*:*"],
            "roles": ["admin"],
            "user": None,
        }
        base.redis_call(
            settings,
            "SET",
            redis_key,
            json.dumps(login_user, separators=(",", ":")),
            "NX",
            "EX",
            "3600",
            expect="OK",
        )
        admin_token = base.make_admin_token(settings["jwt_secret"], settings["login_user_key"], token_id)
        exam_doc, _, _, _ = base.request_json(
            settings["base_url"],
            "/exam/api/exam/exam/save",
            token=admin_token,
            body={
                "title": marker,
                "content": "TEST retained production browser verification; not business data",
                "assessmentType": "competency",
                "scoringMode": "competency_average",
                "joinType": 1,
                "openType": 1,
                "isOpen": 1,
                "answerType": 1,
                "state": 0,
                "totalTime": 30,
                "repoList": [],
                "departIds": [],
                "requiredFields": "name,telephone,age,gender",
            },
        )
        exam_id = base.require_id((exam_doc.get("data") or {}).get("id"), "exam id")
        publish_doc, _, _, _ = base.request_json(
            settings["base_url"],
            "/exam/api/competency/exams/publish",
            token=admin_token,
            body={"examId": exam_id},
        )
        published = publish_doc.get("data") or {}
        if published.get("dimensionCount") != 10 or published.get("questionCount") != 90:
            raise RuntimeError("production publish did not freeze 10 dimensions and 90 questions")
        snapshot = db.query(
            "SELECT CONCAT("
            f"(SELECT COUNT(*) FROM el_exam_competency_group WHERE exam_id={base.sql_literal(exam_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_exam_competency_dimension WHERE exam_id={base.sql_literal(exam_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id={base.sql_literal(exam_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id={base.sql_literal(exam_id)} AND competency_question_type='dimension'),'|',"
            f"(SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id={base.sql_literal(exam_id)} AND competency_question_type='validity'))"
        )
        if snapshot != "2|10|90|80|10":
            raise RuntimeError("production frozen snapshot is invalid")
        return emit({
            "schema": "production-phase1-browser-retained-state-v1",
            "status": "prepared",
            "environment": "production",
            "marker": marker,
            "phone": phone,
            "examId": exam_id,
            "adminToken": admin_token,
            "redisKey": redis_key,
            "publish": {"groups": 2, "dimensions": 10, "questions": 90},
            "retainBusinessData": True,
        })
    except Exception:
        if exam_id and admin_token:
            try:
                base.request_json(
                    settings["base_url"],
                    "/exam/api/exam/exam/delete",
                    token=admin_token,
                    body={"ids": [exam_id]},
                )
            except Exception:
                pass
        if redis_key:
            try:
                base.redis_call(settings, "DEL", redis_key)
            except Exception:
                pass
        raise
    finally:
        mysql_file.unlink(missing_ok=True)


def finalize(base, state_path):
    state = json.loads(Path(state_path).read_text(encoding="utf-8"))
    marker = state.get("marker", "")
    phone = state.get("phone", "")
    exam_id = state.get("examId", "")
    redis_key = state.get("redisKey", "")
    if (
        not MARKER_RE.fullmatch(marker)
        or not PHONE_RE.fullmatch(phone)
        or not ID_RE.fullmatch(exam_id)
        or not redis_key.startswith("login_tokens:prod-phase1-browser-")
    ):
        raise RuntimeError("retained browser state is invalid")
    settings = base.load_settings()
    mysql_file, database = base.mysql_client_from_dsn(settings["dsn"])
    db = base.MySQL(mysql_file, database)
    try:
        row = db.query(
            "SELECT CONCAT(c.id,'\\t',p.id) FROM el_candidate c "
            "JOIN el_paper p ON p.exam_id=c.exam_id AND p.user_id=c.id "
            f"WHERE c.exam_id={base.sql_literal(exam_id)} "
            f"AND c.telephone={base.sql_literal(phone)} ORDER BY p.create_time DESC LIMIT 1"
        )
        parts = row.split("\t") if row else []
        if len(parts) != 2 or not all(ID_RE.fullmatch(value) for value in parts):
            raise RuntimeError("retained candidate or paper was not found")
        candidate_id, paper_id = parts
        counts = db.query(
            "SELECT CONCAT("
            f"(SELECT COUNT(*) FROM el_candidate WHERE id={base.sql_literal(candidate_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_paper WHERE id={base.sql_literal(paper_id)} AND state=2),'|',"
            f"(SELECT COUNT(*) FROM el_paper_qu WHERE paper_id={base.sql_literal(paper_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_paper_qu WHERE paper_id={base.sql_literal(paper_id)} AND answered=1),'|',"
            f"(SELECT COUNT(*) FROM el_competency_dimension_result WHERE paper_id={base.sql_literal(paper_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_group_result WHERE paper_id={base.sql_literal(paper_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_validity_result WHERE paper_id={base.sql_literal(paper_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_result WHERE paper_id={base.sql_literal(paper_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_report WHERE paper_id={base.sql_literal(paper_id)} AND status='completed'),'|',"
            f"(SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id={base.sql_literal(paper_id)} AND status=1))"
        )
        count_parts = counts.split("|")
        if len(count_parts) != 10 or count_parts[:9] != ["1", "1", "90", "90", "10", "2", "1", "1", "1"]:
            raise RuntimeError("retained 00401 closure counts are invalid: " + counts)
        audit_count = int(count_parts[9])
        if audit_count < 2:
            raise RuntimeError("retained 00401 report audit count is invalid: " + counts)
        binding = base.report_file_binding(db, paper_id)
        if binding is None:
            raise RuntimeError("retained report binding is absent")
        report_path = base.validate_bound_report_path(settings, binding)
        if not report_path.is_file():
            raise RuntimeError("retained report file is absent")
        pdf_info = base.run_checked(["pdfinfo", str(report_path)], timeout=30)
        page_match = re.search(r"^Pages:\s+(\d+)$", pdf_info, re.MULTILINE)
        if not page_match or int(page_match.group(1)) != 10:
            raise RuntimeError("retained report is not a 10-page PDF")
        v2_counts = db.query(
            "SELECT CONCAT("
            f"(SELECT COUNT(*) FROM el_competency_result_run WHERE paper_id={base.sql_literal(paper_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_result_run_overall o JOIN el_competency_result_run r ON r.id=o.result_run_id WHERE r.paper_id={base.sql_literal(paper_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_result_run_module m JOIN el_competency_result_run r ON r.id=m.result_run_id WHERE r.paper_id={base.sql_literal(paper_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_result_run_dimension d JOIN el_competency_result_run r ON r.id=d.result_run_id WHERE r.paper_id={base.sql_literal(paper_id)}),'|',"
            f"(SELECT COUNT(*) FROM el_competency_result_run_validity v JOIN el_competency_result_run r ON r.id=v.result_run_id WHERE r.paper_id={base.sql_literal(paper_id)}))"
        )
        if v2_counts != "1|1|3|10|1":
            raise RuntimeError("retained v2 result counts are invalid")
        base.redis_call(settings, "DEL", redis_key, expect="1")
        return emit({
            "schema": "production-phase1-browser-retained-receipt-v1",
            "status": "passed",
            "environment": "production",
            "marker": marker,
            "examId": exam_id,
            "candidateId": candidate_id,
            "paperId": paper_id,
            "counts": counts,
            "reportAuditCount": audit_count,
            "v2Counts": v2_counts,
            "report": {
                "reportId": binding["reportId"],
                "pages": 10,
                "size": binding["size"],
                "sha256": binding["sha256"],
            },
            "retained": True,
            "ephemeralAdminSessionRemoved": True,
            "health": base.read_only_smoke(settings)[0]["health"],
        })
    finally:
        mysql_file.unlink(missing_ok=True)


def main():
    try:
        if len(sys.argv) == 4 and sys.argv[1] == "prepare":
            return prepare(load_base(), sys.argv[2], sys.argv[3])
        if len(sys.argv) == 3 and sys.argv[1] == "finalize":
            return finalize(load_base(), sys.argv[2])
        raise RuntimeError("usage: prepare MARKER PHONE | finalize STATE_JSON")
    except Exception as exc:
        return emit({"status": "failed", "failedStage": sys.argv[1] if len(sys.argv) > 1 else "arguments", "error": str(exc)}, 1)


if __name__ == "__main__":
    sys.exit(main())
