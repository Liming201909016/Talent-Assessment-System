#!/usr/bin/env python3
"""FB-192: verify the dynamic ten-dimension comparison chart is visible in a PDF."""

from __future__ import annotations

import argparse
import re
import shutil
import subprocess
import tempfile
from pathlib import Path

DIMENSIONS = [
    "逻辑思维",
    "计划执行",
    "数字应用",
    "成就导向",
    "持续学习",
    "沟通表达",
    "合作意识",
    "求真务实",
    "自律性",
    "敬业奉献",
]


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("pdf", type=Path)
    args = parser.parse_args()
    pdf = args.pdf.resolve()
    if not pdf.is_file():
        raise SystemExit(f"PDF does not exist: {pdf}")
    converter = shutil.which("pdftotext")
    if not converter:
        raise SystemExit("pdftotext is required")
    with tempfile.TemporaryDirectory(prefix="competency-v2-chart-") as directory:
        text_path = Path(directory) / "report.txt"
        subprocess.run([converter, "-layout", str(pdf), str(text_path)], check=True)
        text = text_path.read_text(encoding="utf-8")
    pages = text.split("\f")
    if pages and not pages[-1].strip():
        pages.pop()
    matching = []
    for index, page in enumerate(pages, 1):
        compact = re.sub(r"\s+", "", page)
        if "常模分" in compact and all(dimension in compact for dimension in DIMENSIONS):
            matching.append(index)
    assert len(matching) == 1, f"visible comparison chart pages={matching}, want exactly one"
    print(f"COMPETENCY_V2_COMPARISON_CHART_PDF_PASS page={matching[0]} dimensions=10 normal_line=visible")


if __name__ == "__main__":
    main()
