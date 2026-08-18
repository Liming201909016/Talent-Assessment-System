#!/usr/bin/env python3
import sys
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TEMPLATE = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / "docs" / "胜任力测评报告模板.docx"
RELATIONSHIPS = "{http://schemas.openxmlformats.org/package/2006/relationships}"
CHART = "{http://schemas.openxmlformats.org/drawingml/2006/chart}"

external = []
dangling_external_data = []
with zipfile.ZipFile(TEMPLATE) as archive:
    for index in range(1, 13):
        relation_name = f"word/charts/_rels/chart{index}.xml.rels"
        relation_ids = set()
        if relation_name in archive.namelist():
            root = ET.fromstring(archive.read(relation_name))
            for relation in root.findall(RELATIONSHIPS + "Relationship"):
                relation_ids.add(relation.get("Id"))
                if relation.get("TargetMode") == "External":
                    external.append((index, relation.get("Id"), relation.get("Target")))
        chart = ET.fromstring(archive.read(f"word/charts/chart{index}.xml"))
        for external_data in chart.iter(CHART + "externalData"):
            dangling_external_data.append(index)

assert not external, f"external chart relationships remain: {external}"
assert not dangling_external_data, f"externalData nodes remain in charts: {dangling_external_data}"
print("COMPETENCY_CUSTOMER_TEMPLATE_CHART_LINKS_TEST_PASS")
