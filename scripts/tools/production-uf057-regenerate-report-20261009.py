#!/usr/bin/env python3
"""Regenerate one authorized production 00401 report with exact backup and rollback."""

from __future__ import annotations

import importlib.util
import json
import os
import re
import shutil
import subprocess
import sys
import time
import uuid
import xml.etree.ElementTree as ET
from pathlib import Path

PAPER_ID = "9174f98e-181f-486e-bbc7-0118bc39ced1"
REPORT_ID = "103d9a0d-0113-4aef-8962-434645783477"
TEMPLATE_SHA = "52e0020c5a6f39535d020bf41b18d9412c096ab43ce6608964b0101b955f30b9"
ROOT = Path("/opt/talent-assessment")


def load_base(path: Path):
    spec = importlib.util.spec_from_file_location("phase1_base", path)
    if spec is None or spec.loader is None:
        raise RuntimeError("runtime helper import failed")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def dump_rows(mysql_file: Path, database: str, table: str, where: str) -> bytes:
    result = subprocess.run(
        [
            "mysqldump", "--defaults-extra-file=" + str(mysql_file), "--compact",
            "--skip-extended-insert", "--no-create-info", "--skip-triggers",
            "--where=" + where, database, table,
        ],
        check=True, capture_output=True,
    )
    return result.stdout


def pdf_assertions(path: Path, base) -> tuple[int, int]:
    info = base.run_checked(["pdfinfo", str(path)], timeout=30)
    page_match = re.search(r"^Pages:\s+(\d+)$", info, re.MULTILINE)
    page_count = int(page_match.group(1)) if page_match else 0
    if page_count != 10:
        raise RuntimeError("regenerated report is not 10 pages")
    cover = base.run_checked(["pdftotext", "-f", "1", "-l", "1", "-layout", str(path), "-"], timeout=30)
    if "时长：" in cover or "分钟" in cover:
        raise RuntimeError("regenerated cover still exposes duration label or unit")
    bbox = base.run_checked(
        ["pdftotext", "-f", "1", "-l", "1", "-bbox-layout", str(path), "-"], timeout=30,
    )
    bbox_root = ET.fromstring(bbox)
    page = next((node for node in bbox_root.iter() if node.tag.endswith("page")), None)
    if page is None:
        raise RuntimeError("regenerated cover bbox is invalid")
    standalone_ones = [
        node for node in page.iter()
        if node.tag.endswith("word") and (node.text or "").strip() == "1"
    ]
    if standalone_ones:
        raise RuntimeError("regenerated cover still exposes an orphan duration value")
    complete = base.run_checked(["pdftotext", "-layout", str(path), "-"], timeout=30)
    page_prefixes = len(re.findall(r"(?i)\bpage\s*\d+\b", complete))
    if page_prefixes:
        raise RuntimeError("regenerated report still exposes the English page prefix")
    complete_bbox = base.run_checked(["pdftotext", "-bbox-layout", str(path), "-"], timeout=60)
    complete_bbox_root = ET.fromstring(complete_bbox)
    page_nodes = [node for node in complete_bbox_root.iter() if node.tag.endswith("page")]
    if len(page_nodes) != 10:
        raise RuntimeError("regenerated report bbox does not contain 10 pages")
    cover_footer_numbers = [
        node for node in page_nodes[0].iter()
        if node.tag.endswith("word")
        and re.fullmatch(r"\d+", (node.text or "").strip())
        and float(node.attrib["yMin"]) >= float(page_nodes[0].attrib["height"]) * 0.85
    ]
    if cover_footer_numbers:
        raise RuntimeError("regenerated cover unexpectedly exposes a numeric footer")
    for expected_page_number, current_page in enumerate(page_nodes[1:], start=1):
        current_width = float(current_page.attrib["width"])
        current_height = float(current_page.attrib["height"])
        footer_numbers = [
            node for node in current_page.iter()
            if node.tag.endswith("word")
            and (node.text or "").strip() == str(expected_page_number)
            and float(node.attrib["yMin"]) >= current_height * 0.85
        ]
        if len(footer_numbers) != 1:
            raise RuntimeError(f"page {expected_page_number} numeric footer count is invalid")
        footer_number = footer_numbers[0]
        footer_center = (float(footer_number.attrib["xMin"]) + float(footer_number.attrib["xMax"])) / 2
        if abs(footer_center - current_width / 2) > current_width * 0.12:
            raise RuntimeError(f"page {expected_page_number} footer is not centered")
    report_xml = base.run_checked(
        ["pdftohtml", "-xml", "-hidden", "-stdout", str(path)], timeout=60,
    )
    report_root = ET.fromstring(report_xml)
    plan_labels = [
        node for node in report_root.iter()
        if node.tag == "text" and "计划执行：" in "".join(node.itertext())
    ]
    if not plan_labels or not any(any(child.tag == "b" for child in node.iter()) for node in plan_labels):
        raise RuntimeError("regenerated report does not render 计划执行 as a bold label")
    return page_count, page_prefixes


def main() -> None:
    helper_path = Path(sys.argv[1]).resolve()
    stamp = sys.argv[2]
    base = load_base(helper_path)
    settings = base.load_settings()
    mysql_file, database = base.mysql_client_from_dsn(settings["dsn"])
    db = base.MySQL(mysql_file, database)
    backup = ROOT / "backups" / ("uf057-report-" + stamp)
    token_id = "prod-uf057-regenerate-" + uuid.uuid4().hex
    redis_key = "login_tokens:" + token_id
    old_binding = None
    new_binding = None
    old_path = None
    redis_created = False
    committed = False
    try:
        if subprocess.run(["hostname"], check=True, capture_output=True, text=True).stdout.strip() != "iZ0yosjdcen2p4Z":
            raise RuntimeError("unexpected production host")
        template = ROOT / "configs" / "export-templates" / "competency-phase1-report-v2.docx"
        if base.sha256_file(template) != TEMPLATE_SHA:
            raise RuntimeError("production template SHA mismatch")
        old_binding = base.report_file_binding(db, PAPER_ID)
        if old_binding is None or old_binding["reportId"] != REPORT_ID:
            raise RuntimeError("authorized report binding mismatch")
        old_path = base.validate_bound_report_path(settings, old_binding)
        if not old_path.is_file():
            raise RuntimeError("authorized old report file missing")
        before_report_count = db.scalar_int("SELECT COUNT(*) FROM el_competency_report WHERE paper_id='" + PAPER_ID + "'")
        before_audit_count = db.scalar_int("SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id='" + PAPER_ID + "'")
        before_current_count = db.scalar_int("SELECT COUNT(*) FROM el_competency_report_current WHERE paper_id='" + PAPER_ID + "'")
        if before_report_count != 1 or before_current_count != 1:
            raise RuntimeError("authorized report closure cardinality mismatch")

        backup.mkdir(mode=0o700, parents=True, exist_ok=False)
        shutil.copy2(old_path, backup / "report-before.pdf")
        for table in ("el_competency_report", "el_competency_report_current", "el_competency_report_audit"):
            payload = dump_rows(mysql_file, database, table, "paper_id='" + PAPER_ID + "'")
            (backup / (table + ".sql")).write_bytes(payload)
        (backup / "receipt-before.json").write_text(
            json.dumps({
                "paperId": PAPER_ID, "reportId": REPORT_ID,
                "reportSha256": old_binding["sha256"], "reportSize": old_binding["size"],
                "reportPath": str(old_path), "reportCount": before_report_count,
                "currentCount": before_current_count, "auditCount": before_audit_count,
            }, sort_keys=True, separators=(",", ":")), encoding="utf-8",
        )
        for path in backup.iterdir():
            path.chmod(0o600)

        base.redis_call(settings, "EXISTS", redis_key, expect="0")
        now_ms = int(time.time() * 1000)
        login_user = {
            "userId": 1, "token": token_id, "loginTime": now_ms,
            "expireTime": now_ms + 600000, "permissions": ["*:*:*"],
            "roles": ["admin"], "user": None,
        }
        base.redis_call(
            settings, "SET", redis_key, json.dumps(login_user, separators=(",", ":")),
            "NX", "EX", "600", expect="OK",
        )
        redis_created = True
        token = base.make_admin_token(settings["jwt_secret"], settings["login_user_key"], token_id)
        document, _, _, _ = base.request_json(
            settings["base_url"], "/exam/api/competency/reports/generate",
            token=token, body={"paperId": PAPER_ID, "force": True}, timeout=180,
        )
        generated = document.get("data") or {}
        if generated.get("status") != "completed" or generated.get("id") != REPORT_ID:
            raise RuntimeError("forced report regeneration did not complete")

        new_binding = base.report_file_binding(db, PAPER_ID)
        if new_binding is None or new_binding["reportId"] != REPORT_ID:
            raise RuntimeError("regenerated report binding mismatch")
        new_path = base.validate_bound_report_path(settings, new_binding)
        pages, page_prefixes = pdf_assertions(new_path, base)
        after_report_count = db.scalar_int("SELECT COUNT(*) FROM el_competency_report WHERE paper_id='" + PAPER_ID + "'")
        after_current_count = db.scalar_int("SELECT COUNT(*) FROM el_competency_report_current WHERE paper_id='" + PAPER_ID + "'")
        after_audit_count = db.scalar_int("SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id='" + PAPER_ID + "'")
        regenerate_count = db.scalar_int(
            "SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id='" + PAPER_ID
            + "' AND action='regenerate' AND status=1"
        )
        if after_report_count != 1 or after_current_count != 1 or after_audit_count != before_audit_count + 1 or regenerate_count < 1:
            raise RuntimeError("regenerated report persistence mismatch")
        status, download, headers = base.request_raw(
            settings["base_url"], "/exam/api/competency/reports/download?paperId=" + PAPER_ID,
            token=token, method="GET", timeout=180,
        )
        if status != 200 or not download.startswith(b"%PDF-") or "application/pdf" not in headers.get("Content-Type", ""):
            raise RuntimeError("regenerated report download response invalid")
        if len(download) != new_binding["size"] or base.hashlib.sha256(download).hexdigest() != new_binding["sha256"]:
            raise RuntimeError("regenerated download does not match database binding")
        committed = True
        print(json.dumps({
            "status": "passed", "environment": "production", "paperId": PAPER_ID,
            "reportId": REPORT_ID, "oldSha256": old_binding["sha256"],
            "newSha256": new_binding["sha256"], "newSize": new_binding["size"],
            "pages": pages, "coverDurationHidden": True, "pagePrefixCount": page_prefixes,
            "reportCount": after_report_count, "currentCount": after_current_count,
            "auditBefore": before_audit_count, "auditAfterGenerate": after_audit_count,
            "downloadValidated": True, "backup": str(backup),
        }, sort_keys=True, separators=(",", ":")))
    finally:
        if redis_created:
            base.redis_call(settings, "DEL", redis_key)
        mysql_file.unlink(missing_ok=True)
        if not committed and backup.exists() and old_binding is not None and old_path is not None:
            try:
                if new_binding is None:
                    client_file, db_name = base.mysql_client_from_dsn(settings["dsn"])
                    rollback_db = base.MySQL(client_file, db_name)
                else:
                    client_file, db_name = base.mysql_client_from_dsn(settings["dsn"])
                    rollback_db = base.MySQL(client_file, db_name)
                rollback_db.execute(
                    "START TRANSACTION;"
                    "DELETE FROM el_competency_report_audit WHERE paper_id='" + PAPER_ID + "' AND id<>'';"
                    "DELETE FROM el_competency_report_current WHERE paper_id='" + PAPER_ID + "' AND report_id<>'';"
                    "DELETE FROM el_competency_report WHERE paper_id='" + PAPER_ID + "' AND id<>'';"
                    + (backup / "el_competency_report.sql").read_text(encoding="utf-8")
                    + (backup / "el_competency_report_current.sql").read_text(encoding="utf-8")
                    + (backup / "el_competency_report_audit.sql").read_text(encoding="utf-8")
                    + "COMMIT;"
                )
                client_file.unlink(missing_ok=True)
                if new_binding is not None and new_binding.get("path") != str(old_path):
                    Path(new_binding["path"]).unlink(missing_ok=True)
                shutil.copy2(backup / "report-before.pdf", old_path)
                print("ROLLBACK=completed", file=sys.stderr)
            except Exception as rollback_error:
                print("ROLLBACK=failed:" + type(rollback_error).__name__, file=sys.stderr)
                raise


if __name__ == "__main__":
    main()
