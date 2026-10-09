#!/usr/bin/env python3
"""FB-172: validate the cleaned, customer-maintainable 260915 report draft."""

from __future__ import annotations

import collections
import re
import sys
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SOURCE = (
    Path(sys.argv[2]).resolve()
    if len(sys.argv) > 2
    else ROOT / "docs" / "260915" / "260915胜任力测评报告样例-基层员工版-表格版.docx"
)
TEMPLATE = (
    Path(sys.argv[1]).resolve()
    if len(sys.argv) > 1
    else ROOT / "docs" / "260915" / "competency-frontline-report-template-draft-v2.docx"
)

W = "{http://schemas.openxmlformats.org/wordprocessingml/2006/main}"
WP = "{http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing}"
C = "{http://schemas.openxmlformats.org/drawingml/2006/chart}"
R = "{http://schemas.openxmlformats.org/officeDocument/2006/relationships}"
PR = "{http://schemas.openxmlformats.org/package/2006/relationships}"

DIMENSIONS = [
    ("logical_reasoning", "逻辑思维"),
    ("plan_execution", "计划执行"),
    ("digital_application", "数字应用"),
    ("achievement_orientation", "成就导向"),
    ("continuous_learning", "持续学习"),
    ("communication", "沟通表达"),
    ("cooperation", "合作意识"),
    ("truth_pragmatism", "求真务实"),
    ("self_discipline", "自律性"),
    ("dedication", "敬业奉献"),
]
MODULES = ["task_management", "interpersonal_management", "self_management"]
EXPECTED_CHART_KEYS = [
    "chart.overall.score",
    "chart.dimension.comparison",
    *[f"chart.dimension.{key}" for key, _ in DIMENSIONS],
]
EXPECTED_TAG_COUNTS = collections.Counter(
    {
        "participant.name": 2,
        "participant.age": 1,
        "participant.gender": 2,
        "participant.telephone": 2,
        "participant.affiliation": 1,
        "participant.post": 1,
        "result.submittedAt": 1,
        "result.userTime": 1,
        "validity.status": 1,
        "overall.score": 5,
        "overall.level": 2,
        "overall.normComparison": 1,
        "rule.moduleSummary": 1,
        "rule.overallAdvice": 1,
        "rule.strength.1": 1,
        "rule.strength.2": 1,
        "rule.strength.3": 1,
        "rule.development.1": 1,
        "rule.development.2": 1,
    }
)
for module in MODULES:
    EXPECTED_TAG_COUNTS[f"module.{module}.score"] = 2
    EXPECTED_TAG_COUNTS[f"module.{module}.level"] = 2
    EXPECTED_TAG_COUNTS[f"module.{module}.normComparison"] = 2
for dimension, _ in DIMENSIONS:
    EXPECTED_TAG_COUNTS[f"dimension.{dimension}.score"] = 1
    EXPECTED_TAG_COUNTS[f"dimension.{dimension}.level"] = 1
    EXPECTED_TAG_COUNTS[f"dimension.{dimension}.performance"] = 1

assert TEMPLATE.is_file(), f"FB-172 RED: cleaned template is missing: {TEMPLATE}"
assert SOURCE.is_file(), f"authoritative customer source is missing: {SOURCE}"

with zipfile.ZipFile(TEMPLATE) as archive:
    names = archive.namelist()
    assert len(names) == len(set(names)), "duplicate ZIP members"
    assert "word/document.xml" in names, "document.xml missing"
    document = ET.fromstring(archive.read("word/document.xml"))
    tags = collections.Counter(
        element.get(W + "val", "") for element in document.iter(W + "tag")
    )
    for header_name in sorted(name for name in names if re.fullmatch(r"word/header\d+\.xml", name)):
        header = ET.fromstring(archive.read(header_name))
        tags.update(element.get(W + "val", "") for element in header.iter(W + "tag"))

    assert tags == EXPECTED_TAG_COUNTS, f"content-control contract mismatch: {tags - EXPECTED_TAG_COUNTS}; missing={EXPECTED_TAG_COUNTS - tags}"
    assert sum(tags.values()) == 75 and len(tags) == 58, f"controls={sum(tags.values())}/{len(tags)}, want 75/58"

    document_xml = archive.read("word/document.xml").decode("utf-8")
    header_xml = "".join(
        archive.read(name).decode("utf-8")
        for name in names
        if re.fullmatch(r"word/header\d+\.xml", name)
    )
    assert "{{" not in document_xml + header_xml, "visible placeholder remains"
    for old_value in ("68.15", "75.08", "78.25"):
        assert old_value not in document_xml, f"rounded sample value remains: {old_value}"
    for value in ("68.13", "75.00", "78.13"):
        assert value in document_xml, f"corrected exact-score sample missing: {value}"

    for paragraph_id in ("1B10269B", "7953F03A", "087D950F", "35898E47"):
        assert paragraph_id not in document_xml, f"drift-prone blank cover paragraph remains: {paragraph_id}"
    assert document_xml.count("<w:sectPr") == 4, "customer section structure changed"

    tables = list(document.iter(W + "tbl"))
    assert len(tables) == 9, f"table count={len(tables)}, want 9"
    detail_tables = tables[6:9]
    row_counts = [15, 6, 9]
    for table, expected_rows in zip(detail_tables, row_counts):
        rows = table.findall(W + "tr")
        assert len(rows) == expected_rows
        for row_index, row in enumerate(rows):
            properties = row.find(W + "trPr")
            assert properties is not None and properties.find(W + "cantSplit") is not None, f"detail row {row_index} may split"
        for row_index in range(0, expected_rows, 3):
            for keep_index in (row_index, row_index + 1):
                paragraphs = [paragraph for cell in rows[keep_index].findall(W + "tc") for paragraph in cell.findall(W + "p")]
                assert paragraphs, f"detail row {keep_index} has no paragraph"
                for paragraph in paragraphs:
                    properties = paragraph.find(W + "pPr")
                    assert properties is not None and properties.find(W + "keepNext") is not None, f"detail row {keep_index} is not kept with next"

    relationships = ET.fromstring(archive.read("word/_rels/document.xml.rels"))
    relationship_targets = {
        element.get("Id"): element.get("Target")
        for element in relationships.findall(PR + "Relationship")
    }
    chart_keys_by_part = {}
    chart_drawings = 0
    for drawing in list(document.iter(WP + "inline")) + list(document.iter(WP + "anchor")):
        chart = drawing.find(".//" + C + "chart")
        if chart is None:
            continue
        chart_drawings += 1
        relation_id = chart.get(R + "id")
        target = relationship_targets.get(relation_id, "")
        part = "word/" + target.replace("\\", "/")
        document_properties = drawing.find(WP + "docPr")
        key = document_properties.get("title", "") if document_properties is not None else ""
        chart_keys_by_part[part] = key
    assert chart_drawings == 12, f"chart drawings={chart_drawings}, want 12"
    assert sorted(chart_keys_by_part.values()) == sorted(EXPECTED_CHART_KEYS), f"chart business keys invalid: {chart_keys_by_part}"

    for name in names:
        if not name.endswith(".rels"):
            continue
        root = ET.fromstring(archive.read(name))
        external = [item for item in root.findall(PR + "Relationship") if item.get("TargetMode") == "External"]
        assert not external, f"external relationship remains in {name}"

    for index in range(1, 13):
        part = f"word/charts/chart{index}.xml"
        chart = ET.fromstring(archive.read(part))
        assert not list(chart.iter(C + "externalData")), f"externalData remains in chart{index}"
        assert not list(chart.iter(C + "numRef")), f"numeric reference remains in chart{index}"
        assert not list(chart.iter(C + "strRef")), f"string reference remains in chart{index}"
        assert not list(chart.iter(C + "f")), f"formula remains in chart{index}"
        series = list(chart.iter(C + "ser"))
        if index == 1:
            assert chart.find(".//" + C + "doughnutChart") is not None and len(series) == 1
            counts = [int(item.get("val")) for item in chart.iter(C + "ptCount")]
            assert 2 in counts, f"overall chart data points invalid: {counts}"
        elif index == 2:
            assert chart.find(".//" + C + "barChart") is not None
            assert chart.find(".//" + C + "lineChart") is not None
            assert len(series) == 2, f"comparison chart series={len(series)}, want 2"
            for item in series:
                counts = [int(count.get("val")) for count in item.iter(C + "ptCount")]
                assert 10 in counts, f"comparison series data points invalid: {counts}"
        else:
            assert chart.find(".//" + C + "doughnutChart") is not None and len(series) == 1
            counts = [int(item.get("val")) for item in chart.iter(C + "ptCount")]
            assert 2 in counts, f"dimension chart{index} data points invalid: {counts}"

    assert "68.125" in archive.read("word/charts/chart1.xml").decode("utf-8")
    comparison_xml = archive.read("word/charts/chart2.xml").decode("utf-8")
    assert comparison_xml.count("78.125") >= 2, "exact 78.125 scores missing from comparison chart"

with zipfile.ZipFile(SOURCE) as source_archive, zipfile.ZipFile(TEMPLATE) as template_archive:
    source_media = sorted(name for name in source_archive.namelist() if name.startswith("word/media/"))
    template_media = sorted(name for name in template_archive.namelist() if name.startswith("word/media/"))
    assert source_media == template_media, "customer media inventory changed"
    for name in source_media:
        assert source_archive.read(name) == template_archive.read(name), f"customer media bytes changed: {name}"

print("COMPETENCY_260915_TEMPLATE_CONTRACT_TEST_PASS")
print(f"template={TEMPLATE}")
print(f"content_controls={sum(EXPECTED_TAG_COUNTS.values())}|unique_tags={len(EXPECTED_TAG_COUNTS)}|charts=12|external_links=0")
