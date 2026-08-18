#!/usr/bin/env python3
# FB-139: docs/regression-tests.md
# Reproduction: LibreOffice renders the visible doughnut value away from the hole centre.
# Expected: the visible label box centre equals the chart centre on both axes.
import zipfile
import xml.etree.ElementTree as ET
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TEMPLATE = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / "docs" / "胜任力测评报告模板.docx"
C = "{http://schemas.openxmlformats.org/drawingml/2006/chart}"
C15 = "{http://schemas.microsoft.com/office/drawing/2012/chart}"

with zipfile.ZipFile(TEMPLATE) as archive:
    for index in range(3, 13):
        root = ET.fromstring(archive.read(f"word/charts/chart{index}.xml"))
        visible_label = next(
            label for label in root.iter(C + "dLbl")
            if label.find(C + "delete") is None
        )
        manual_layout = visible_label.find(f"{C}layout/{C}manualLayout")
        extension_layout = visible_label.find(f"{C}extLst/{C}ext/{C15}layout/{C}manualLayout")
        assert manual_layout is not None, f"chart{index} visible label has no manual layout"
        assert extension_layout is not None, f"chart{index} visible label has no size layout"

        x = float(manual_layout.find(C + "x").get("val"))
        y = float(manual_layout.find(C + "y").get("val"))
        width = float(extension_layout.find(C + "w").get("val"))
        height = float(extension_layout.find(C + "h").get("val"))
        assert abs(x + width / 2) < 1e-12, f"chart{index} label horizontal centre offset={x + width / 2}"
        assert abs(y + height / 2) < 1e-12, f"chart{index} label vertical centre offset={y + height / 2}"

print("COMPETENCY_CUSTOMER_TEMPLATE_CHART_LABEL_TEST_PASS")
