#!/usr/bin/env python3
# FB-147/FB-148: docs/regression-tests.md
# Reproduction: a customer-edited template keeps static first-level sample scores and
# displays chart labels with mixed decimal precision.
# Expected: summary and analysis scores share repeatable controls, and every visible
# score chart label uses an explicit two-decimal number format.
import collections
import sys
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TEMPLATE = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / "docs" / "胜任力测评报告模板.docx"
W = "{http://schemas.openxmlformats.org/wordprocessingml/2006/main}"
C = "{http://schemas.openxmlformats.org/drawingml/2006/chart}"
GROUP_SCORE_TAGS = (
    "group.general_ability.score",
    "group.psychological_quality.score",
)
SCORE_CHARTS = (1, *range(3, 13))

with zipfile.ZipFile(TEMPLATE) as archive:
    document = ET.fromstring(archive.read("word/document.xml"))
    tags = collections.Counter(
        item.get(W + "val", "") for item in document.iter(W + "tag")
    )
    for tag in GROUP_SCORE_TAGS:
        assert tags[tag] == 2, f"{tag} controls={tags[tag]}, want 2"

    parents = {child: parent for parent in document.iter() for child in parent}
    for text in document.iter(W + "t"):
        if (text.text or "").strip() not in {"3.75", "3.70"}:
            continue
        ancestor = parents.get(text)
        while ancestor is not None and ancestor.tag != W + "sdt":
            ancestor = parents.get(ancestor)
        assert ancestor is not None, f"static sample score remains outside a content control: {text.text}"

    for index in SCORE_CHARTS:
        chart = ET.fromstring(archive.read(f"word/charts/chart{index}.xml"))
        label_groups = list(chart.iter(C + "dLbls"))
        assert label_groups, f"chart{index} has no data labels"
        formats = [
            item.get("formatCode", "")
            for group in label_groups
            for item in group.iter(C + "numFmt")
        ]
        assert "0.00" in formats, f"chart{index} visible score format={formats or ['missing']}, want 0.00"

print("COMPETENCY_CUSTOMER_TEMPLATE_SCORE_CONTRACT_TEST_PASS")
