#!/usr/bin/env python3
"""Hide the v2 00401 cover duration while preserving its internal field contract."""

from __future__ import annotations

import argparse
import re
import tempfile
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DEFAULT_TEMPLATE = (
    ROOT
    / "Go-based Refactored System"
    / "configs"
    / "export-templates"
    / "competency-phase1-report-v2.docx"
)


def hide_duration(document: str) -> str:
    tag = 'w:val="result.userTime"'
    at = document.find(tag)
    if at < 0:
        raise RuntimeError("duration control not found")
    row_start = document.rfind("<w:tr", 0, at)
    row_end = document.find("</w:tr>", at)
    if row_start < 0 or row_end < 0:
        raise RuntimeError("duration row not found")
    row_end += len("</w:tr>")
    row = document[row_start:row_end]
    cells = list(re.finditer(r"<w:tc\b.*?</w:tc>", row, re.S))
    duration_cell = next((match for match in cells if tag in match.group(0)), None)
    if duration_cell is None:
        raise RuntimeError("duration cell not found")
    cell = duration_cell.group(0)
    run_pattern = re.compile(r"<w:r(?:\s[^>]*)?>.*?</w:r>", re.S)

    def remove_run_with_text(value: str, text: str) -> tuple[str, int]:
        matches = [match for match in run_pattern.finditer(value) if f">{text}</w:t>" in match.group(0)]
        if len(matches) != 1:
            return value, len(matches)
        match = matches[0]
        return value[: match.start()] + value[match.end() :], 1

    cell, label_count = remove_run_with_text(cell, "时长：")
    cell, unit_count = remove_run_with_text(cell, "分钟")
    if label_count != 1 or unit_count != 1:
        raise RuntimeError("duration label or unit structure changed")
    control_pattern = re.compile(r"<w:sdt>.*?<w:tag w:val=\"result\.userTime\".*?</w:sdt>", re.S)
    control_match = control_pattern.search(cell)
    if control_match is None:
        raise RuntimeError("duration control structure changed")
    control = control_match.group(0)
    content_start = control.find("<w:sdtContent>")
    content_end = control.rfind("</w:sdtContent>")
    if content_start < 0 or content_end <= content_start:
        raise RuntimeError("duration control content changed")
    content_start += len("<w:sdtContent>")
    body = control[content_start:content_end]
    if "<w:vanish" not in body:
        if "<w:rPr>" in body:
            body = body.replace("<w:rPr>", "<w:rPr><w:vanish/>", 1)
        else:
            run_end = body.find(">")
            if run_end < 0 or not body.startswith("<w:r"):
                raise RuntimeError("duration value run changed")
            body = body[: run_end + 1] + "<w:rPr><w:vanish/></w:rPr>" + body[run_end + 1 :]
    control = control[:content_start] + body + control[content_end:]
    cell = cell[: control_match.start()] + control + cell[control_match.end() :]
    row = row[: duration_cell.start()] + cell + row[duration_cell.end() :]
    updated = document[:row_start] + row + document[row_end:]
    if ">时长：</w:t>" in updated or ">分钟</w:t>" in updated:
        raise RuntimeError("visible duration text remains")
    return updated


def rewrite_template(source_path: Path, output_path: Path) -> None:
    if not source_path.is_file():
        raise RuntimeError(f"template not found: {source_path}")
    output_path.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(source_path) as source:
        infos = source.infolist()
        parts = {info.filename: source.read(info.filename) for info in infos}
    document_name = "word/document.xml"
    if document_name not in parts:
        raise RuntimeError("document.xml not found")
    before = parts[document_name].decode("utf-8")
    after = hide_duration(before)
    if before == after:
        raise RuntimeError("template was not changed")
    parts[document_name] = after.encode("utf-8")
    with tempfile.NamedTemporaryFile(dir=output_path.parent, suffix=".docx", delete=False) as temporary:
        output = Path(temporary.name)
    try:
        with zipfile.ZipFile(output, "w") as target:
            for info in infos:
                method = zipfile.ZIP_STORED if info.is_dir() or not parts[info.filename] else zipfile.ZIP_DEFLATED
                clone = zipfile.ZipInfo(info.filename, info.date_time)
                clone.external_attr = info.external_attr
                clone.internal_attr = info.internal_attr
                clone.create_system = info.create_system
                clone.flag_bits = info.flag_bits
                clone.compress_type = method
                target.writestr(clone, parts[info.filename])
        with zipfile.ZipFile(output) as verification:
            verification.testzip()
            rendered = verification.read(document_name).decode("utf-8")
            if 'w:val="result.userTime"' not in rendered or "<w:vanish" not in rendered:
                raise RuntimeError("duration contract verification failed")
        output.replace(output_path)
    finally:
        output.unlink(missing_ok=True)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", type=Path, default=DEFAULT_TEMPLATE)
    parser.add_argument("--output", type=Path, default=DEFAULT_TEMPLATE)
    args = parser.parse_args()
    rewrite_template(args.source.resolve(), args.output.resolve())


if __name__ == "__main__":
    main()
