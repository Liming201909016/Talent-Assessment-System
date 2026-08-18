#!/usr/bin/env python3
import sys
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TEMPLATE = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / "docs" / "胜任力测评报告模板.docx"
W = "{http://schemas.openxmlformats.org/wordprocessingml/2006/main}"
MODULES = {
    "dimension.competency-a1-01": (0, 1, 2),
    "dimension.competency-a1-02": (3, 4, 5),
    "dimension.competency-a1-03": (6, 7, 8),
    "dimension.competency-a1-04": (9, 10, 11),
    "dimension.competency-a1-05": (12, 13, 14),
    "dimension.competency-b1-01": (0, 1, 2),
    "dimension.competency-b1-02": (3, 4, 5),
    "dimension.competency-b1-03": (6, 7, 8),
    "dimension.competency-b1-04": (9, 10, 11),
    "dimension.competency-b1-05": (12, 13, 14),
}

with zipfile.ZipFile(TEMPLATE) as archive:
    root = ET.fromstring(archive.read("word/document.xml"))

body = root.find(W + "body")
assert body is not None
body_children = list(body)


def visible_text(element):
    return "".join(item.text or "" for item in element.iter(W + "t")).strip()


secondary_heading_index = next(
    index for index, element in enumerate(body_children)
    if visible_text(element) == "二级维度测评结果及建议"
)
previous_element = body_children[secondary_heading_index - 1]
assert not any(
    item.get(W + "type") == "page" for item in previous_element.iter(W + "br")
), "explicit page break before secondary dimensions creates a sparse LibreOffice page"

def table_for(prefix: str):
    for table in root.iter(W + "tbl"):
        tags = [item.get(W + "val", "") for item in table.iter(W + "tag")]
        if any(tag.startswith(prefix) for tag in tags):
            return table
    raise AssertionError(f"table not found for {prefix}")

for prefix, row_indexes in MODULES.items():
    table = table_for(prefix)
    rows = table.findall(W + "tr")
    for row_index in row_indexes:
        row = rows[row_index]
        properties = row.find(W + "trPr")
        assert properties is not None and properties.find(W + "cantSplit") is not None, f"{prefix} row {row_index} may split"
    for row_index in row_indexes[:2]:
        paragraphs = [paragraph for cell in rows[row_index].findall(W + "tc") for paragraph in cell.findall(W + "p")]
        assert paragraphs, f"{prefix} row {row_index} has no paragraph"
        for paragraph in paragraphs:
            properties = paragraph.find(W + "pPr")
            assert properties is not None and properties.find(W + "keepNext") is not None, f"{prefix} row {row_index} is not kept with next"

print("COMPETENCY_CUSTOMER_TEMPLATE_MODULE_FLOW_TEST_PASS")
