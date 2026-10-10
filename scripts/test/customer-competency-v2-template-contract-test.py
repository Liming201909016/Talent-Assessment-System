#!/usr/bin/env python3
"""Structural contract for the customer-based 00401 v2 staging candidate."""

from __future__ import annotations

import re
import sys
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TEMPLATE = (
    Path(sys.argv[1]).resolve()
    if len(sys.argv) > 1
    else ROOT / "docs" / "261010胜任力待完善" / "最新胜任力报告模板-修复后-staging候选.docx"
)

MODULES = ("task_management", "interpersonal_management", "self_management")
DIMENSIONS = (
    "logical_reasoning", "plan_execution", "digital_application", "achievement_orientation",
    "continuous_learning", "communication", "cooperation", "truth_pragmatism",
    "self_discipline", "dedication",
)
FIELDS = {
    "participant.name", "participant.age", "participant.gender", "participant.telephone",
    "participant.affiliation", "participant.post", "result.submittedAt", "result.userTime",
    "validity.status", "validity.text", "report.disclaimer", "overall.score", "overall.level",
    "overall.normComparison", "rule.moduleSummary", "rule.overallAdvice", "rule.strength.1",
    "rule.strength.2", "rule.strength.3", "rule.development.1", "rule.development.2",
}
FIELDS.update(f"module.{module}.{suffix}" for module in MODULES for suffix in ("score", "level", "normComparison"))
FIELDS.update(f"dimension.{dimension}.{suffix}" for dimension in DIMENSIONS for suffix in ("score", "level", "performance"))
CHARTS = {"chart.overall.score", "chart.dimension.comparison"}
CHARTS.update(f"chart.dimension.{dimension}" for dimension in DIMENSIONS)


def main() -> None:
    with zipfile.ZipFile(TEMPLATE) as archive:
        bad_part = archive.testzip()
        assert bad_part is None, f"invalid ZIP part: {bad_part}"
        names = archive.namelist()
        for name in names:
            if name.endswith(".xml") or name.endswith(".rels"):
                try:
                    ET.fromstring(archive.read(name))
                except ET.ParseError as exc:
                    raise AssertionError(f"invalid XML part: {name}: {exc}") from exc
        document = archive.read("word/document.xml").decode("utf-8")
        headers = "".join(
            archive.read(name).decode("utf-8")
            for name in names
            if re.fullmatch(r"word/header\d+\.xml", name)
        )
        footers = "".join(
            archive.read(name).decode("utf-8")
            for name in names
            if re.fullmatch(r"word/footer\d+\.xml", name)
        )
        tags = re.findall(r'<w:tag\s+w:val="([a-zA-Z0-9_.-]+)"\s*/>', document + headers)
        unique_tags = set(tags)
        assert unique_tags, "no registered content controls"
        assert unique_tags <= FIELDS, f"unknown fields: {sorted(unique_tags - FIELDS)}"
        assert len(unique_tags) <= 60, f"unique fields={len(unique_tags)}, want <=60"
        assert "validity.text" not in unique_tags, "optional validity.text was added"
        assert "NUMPAGES" not in footers.upper(), "NUMPAGES remains"
        assert "页 共" not in footers, "visible total-page label remains"
        footer_parts = [
            archive.read(name).decode("utf-8")
            for name in names
            if re.fullmatch(r"word/footer\d+\.xml", name)
        ]
        assert all(len(re.findall(r"<w:instrText(?:\s[^>]*)?>\s*PAGE", part, re.I)) == 1 for part in footer_parts), "footer PAGE field count invalid"
        assert ">时长：</w:t>" not in document and ">分钟</w:t>" not in document, "visible duration remains"
        duration = re.search(r'<w:sdt>.*?<w:tag w:val="result\.userTime".*?</w:sdt>', document, re.S)
        assert duration is not None and "<w:vanish" in duration.group(0), "hidden duration contract missing"

        titles = set(re.findall(r'title="(chart\.[a-zA-Z0-9_.-]+)"', document))
        assert CHARTS <= titles, f"missing chart keys: {sorted(CHARTS - titles)}"
        grouped_comparison = [
            match.group(0)
            for match in re.finditer(r"<mc:AlternateContent>.*?</mc:AlternateContent>", document, re.S)
            if 'title="chart.dimension.comparison"' in match.group(0)
        ]
        assert not grouped_comparison, "comparison chart remains grouped"
        assert 'title="chart.dimension.comparison.scale"' in document, "standalone comparison scale missing"
        grouped_overall = [
            match.group(0)
            for match in re.finditer(r"<mc:AlternateContent>.*?</mc:AlternateContent>", document, re.S)
            if 'title="chart.overall.score"' in match.group(0)
        ]
        assert not grouped_overall, "overall score chart remains grouped"
        assert re.search(r'<wp:docPr\b[^>]*title="chart\.overall\.score"', document), "standalone overall score chart missing"
        overall_chart_position = document.index('title="chart.overall.score"')
        overall_region = document[max(0, overall_chart_position - 5000):overall_chart_position + 5000]
        assert ">总体评价</w:t>" in overall_region, "overall chart center label missing"
        assert 'w:val="overall.score"' in overall_region, "overall chart center score control missing"
        assert ">分</w:t>" in overall_region, "overall chart center score unit missing"

        for part in footer_parts:
            assert "<wps:wsp" not in part, "PAGE field remains inside a WPS floating text box"
            assert not re.search(r"<mc:AlternateContent>.*?<w:instrText(?:\s[^>]*)?>\s*PAGE", part, re.I | re.S), "PAGE field remains inside AlternateContent"
        for section in re.findall(r"<w:sectPr>.*?</w:sectPr>", document, re.S):
            if 'w:type="default"' in section and '<w:footerReference' in section:
                assert 'w:type="first"' not in section and "<w:titlePg" not in section, "PAGE section still uses a LibreOffice-suppressed first footer"

        for name in names:
            body = archive.read(name)
            if name.endswith(".rels"):
                assert b'TargetMode="External"' not in body, f"external relationship: {name}"
            if re.fullmatch(r"word/charts/chart\d+\.xml", name):
                for forbidden in (b"<c:numRef", b"<c:strRef", b"<c:externalData", b"<c:f>"):
                    assert forbidden not in body, f"forbidden chart reference {forbidden!r}: {name}"

    print(
        "CUSTOMER_COMPETENCY_V2_TEMPLATE_CONTRACT_PASS "
        f"fields={len(unique_tags)} charts={len(CHARTS)}"
    )


if __name__ == "__main__":
    main()
