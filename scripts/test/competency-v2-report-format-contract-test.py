#!/usr/bin/env python3
"""FB-190: contract for the repaired phase-1 v2 Word template."""

from __future__ import annotations

import re
import struct
import sys
import zipfile
import xml.etree.ElementTree as ET
import importlib.util
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TEMPLATE = (
    Path(sys.argv[1]).resolve()
    if len(sys.argv) > 1
    else ROOT / "Go-based Refactored System" / "configs" / "export-templates" / "competency-phase1-report-v2.docx"
)
W = "{http://schemas.openxmlformats.org/wordprocessingml/2006/main}"
A = "{http://schemas.openxmlformats.org/drawingml/2006/main}"
R = "{http://schemas.openxmlformats.org/officeDocument/2006/relationships}"
PR = "{http://schemas.openxmlformats.org/package/2006/relationships}"


def test_bug_fb191_customer_grade_scale_relationship() -> None:
    """FB-191: a customer Word re-save may renumber image17 away from rId27."""
    repair_path = ROOT / "scripts" / "tools" / "repair-competency-v2-report-template.py"
    spec = importlib.util.spec_from_file_location("repair_competency_v2_report_template", repair_path)
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    document = (
        '<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" '
        'xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" '
        'xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture" '
        'xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">'
        '<pic:pic><pic:blipFill><a:blip r:embed="rId28"/></pic:blipFill><pic:spPr><a:xfrm>'
        '<a:off x="5796" y="54639"/><a:ext cx="580" cy="3637"/>'
        '</a:xfrm></pic:spPr></pic:pic></w:document>'
    )
    relationships = (
        b'<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">'
        b'<Relationship Id="rId28" Target="media/image17.png" '
        b'Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"/>'
        b'</Relationships>'
    )
    repaired = module.resize_scale_drawing(document, relationships)
    assert '<a:off x="5796" y="56300"/><a:ext cx="3500" cy="700"/>' in repaired


def test_bug_fb192_comparison_chart_is_not_grouped() -> None:
    """FB-192: LibreOffice drops chart2 when it remains inside a Word wpg group."""
    with zipfile.ZipFile(TEMPLATE) as archive:
        document = archive.read("word/document.xml").decode("utf-8")
    grouped = [
        match.group(0)
        for match in re.finditer(r"<mc:AlternateContent>.*?</mc:AlternateContent>", document, re.S)
        if 'title="chart.dimension.comparison"' in match.group(0)
    ]
    assert not grouped, "comparison chart remains in mc:AlternateContent/wpg group"
    assert 'title="chart.dimension.comparison.scale"' in document, "standalone grade scale missing"
    assert 'title="chart.dimension.comparison"' in document, "standalone comparison chart missing"


def test_fb194_overview_visual_hierarchy() -> None:
    """FB-194: overview card, doughnut centre and vertical scale have stable visual hierarchy."""
    with zipfile.ZipFile(TEMPLATE) as archive:
        document = ET.fromstring(archive.read("word/document.xml"))
        relationships = ET.fromstring(archive.read("word/_rels/document.xml.rels"))
        relation_targets = {item.get("Id"): item.get("Target", "") for item in relationships.findall(PR + "Relationship")}

        summary = next(table for table in document.iter(W + "tbl") if "rule.moduleSummary" in tags(table))
        summary_cell = next(summary.iter(W + "tc"))
        properties = summary_cell.find(W + "tcPr")
        assert properties is not None
        borders = properties.find(W + "tcBorders")
        left = borders.find(W + "left") if borders is not None else None
        assert left is not None and left.get(W + "color") == "00B050" and int(left.get(W + "sz", "0")) >= 24
        assert properties.find(W + "tcMar") is not None, "summary card cell margins missing"

        center_paragraph = next(
            paragraph
            for paragraph in document.iter(W + "p")
            if "总体得分" in text(paragraph) and "overall.score" in tags(paragraph)
        )
        overall_score_control = next(control for control in center_paragraph.iter(W + "sdt") if "overall.score" in tags(control))
        score_sizes = [int(item.get(W + "val")) for item in overall_score_control.iter(W + "sz")]
        score_colors = [item.get(W + "val") for item in overall_score_control.iter(W + "color")]
        assert score_sizes and max(score_sizes) >= 32 and "00B050" in score_colors

        scale_relation = next(
            item.get(R + "embed")
            for item in document.iter(A + "blip")
            if relation_targets.get(item.get(R + "embed"), "").endswith("image17.png")
        )
        scale_bytes = archive.read("word/" + relation_targets[scale_relation])
        width, height = struct.unpack(">II", scale_bytes[16:24])
        assert width >= 100 and height >= 550, f"overview scale pixels={width}x{height}"


def text(element: ET.Element) -> str:
    return "".join(item.text or "" for item in element.iter(W + "t"))


def tags(element: ET.Element) -> list[str]:
    return [item.get(W + "val", "") for item in element.iter(W + "tag")]


def main() -> None:
    test_bug_fb191_customer_grade_scale_relationship()
    test_bug_fb192_comparison_chart_is_not_grouped()
    test_fb194_overview_visual_hierarchy()
    with zipfile.ZipFile(TEMPLATE) as archive:
        document = ET.fromstring(archive.read("word/document.xml"))
        relationships = ET.fromstring(archive.read("word/_rels/document.xml.rels"))
        all_tags = tags(document)
        for name in archive.namelist():
            if re.fullmatch(r"word/header\d+\.xml", name):
                all_tags.extend(tags(ET.fromstring(archive.read(name))))
        assert len(set(all_tags)) == 60, f"unique fields={len(set(all_tags))}, want 60"
        for key in ("validity.text", "report.disclaimer"):
            assert key in all_tags, f"missing approved-text binding: {key}"

        profile = next(table for table in document.iter(W + "tbl") if "result.userTime" in tags(table))
        grid = [int(column.get(W + "w")) for column in profile.iter(W + "gridCol")]
        assert len(grid) == 3, f"profile grid columns={grid}, want 3"
        duration_cell = next(cell for cell in profile.iter(W + "tc") if "result.userTime" in tags(cell))
        duration_width = int(next(duration_cell.iter(W + "tcW")).get(W + "w"))
        assert duration_width >= 1500, f"duration cell width={duration_width}, want >=1500"
        assert next(duration_cell.iter(W + "gridSpan"), None) is None, "duration cell still spans ghost columns"
        phone_cell = next(cell for cell in profile.iter(W + "tc") if "participant.telephone" in tags(cell))
        phone_paragraph = next(paragraph for paragraph in phone_cell.iter(W + "p") if "participant.telephone" in tags(paragraph))
        justify = phone_paragraph.find(".//" + W + "jc")
        assert justify is not None and justify.get(W + "val") == "left", "telephone row is not left aligned"
        sizes = [int(item.get(W + "val")) for item in phone_paragraph.iter(W + "sz")]
        assert sizes and max(sizes) <= 18, f"telephone font sizes={sizes}"

        relation_targets = {
            item.get("Id"): item.get("Target", "")
            for item in relationships.findall(PR + "Relationship")
        }
        scale_relation = next(
            item.get(R + "embed")
            for item in document.iter(A + "blip")
            if relation_targets.get(item.get(R + "embed"), "").endswith("image17.png")
        )
        scale_target = relation_targets[scale_relation]
        scale_bytes = archive.read("word/" + scale_target)
        width, height = struct.unpack(">II", scale_bytes[16:24])
        assert 100 <= width <= 220 and height >= 550, f"grade scale pixels={width}x{height}"
        scale_blip = next(item for item in document.iter(A + "blip") if item.get(R + "embed") == scale_relation)
        parents = {child: parent for parent in document.iter() for child in parent}
        picture = parents[parents[scale_blip]]
        extent = next(picture.iter(A + "ext"))
        assert 550000 <= int(extent.get("cx")) <= 750000, f"grade scale extent={extent.get('cx')}"

        separator_ids = {
            paragraph.get("{http://schemas.microsoft.com/office/word/2010/wordml}paraId")
            for paragraph in document.iter(W + "p")
            if paragraph.find(".//" + W + "br") is not None
        }
        assert {"FB190001", "FB190002", "FB190003", "FB190004"}.issubset(separator_ids), separator_ids
        detail_tables = [table for table in document.iter(W + "tbl") if any(tag.startswith("dimension.") and tag.endswith(".score") for tag in tags(table))]
        dimensions_per_table = [sum(tag.startswith("dimension.") and tag.endswith(".score") for tag in tags(table)) for table in detail_tables]
        assert dimensions_per_table == [2, 2, 1, 1, 1, 1, 2], dimensions_per_table

    print("COMPETENCY_V2_REPORT_FORMAT_CONTRACT_TEST_PASS")


if __name__ == "__main__":
    main()
