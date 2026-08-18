#!/usr/bin/env python3
import re
import sys
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TEMPLATE = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / "docs" / "胜任力测评报告模板.docx"
TAG = "dimension.competency-a1-04.diagnosis"
LABEL = "【诊断】"

with zipfile.ZipFile(TEMPLATE) as archive:
    document = archive.read("word/document.xml").decode("utf-8")

tag_marker = f'<w:tag w:val="{TAG}"/>'
assert document.count(tag_marker) == 1, f"{TAG} Tag count={document.count(tag_marker)}, want 1"
position = document.index(tag_marker)
starts = [match.start() for match in re.finditer(r"<w:sdt>", document[:position])]
control = ""
control_start = -1
control_end = -1
for start in reversed(starts):
    depth = 0
    for token in re.finditer(r"</?w:sdt>", document[start:]):
        depth += 1 if token.group() == "<w:sdt>" else -1
        if depth == 0:
            end = start + token.end()
            if start <= position < end:
                control = document[start:end]
                control_start, control_end = start, end
            break
    if control:
        break
assert control, "content control not found"
control_text = "".join(re.findall(r"<w:t(?:\s[^>]*)?>(.*?)</w:t>", control, re.S))
assert not control_text.startswith(LABEL), "diagnosis label must be outside the dynamic control"
assert re.search(r'<w:alias\b[^>]*\bw:val="dimension\.competency-a1-04\.diagnosis"', control), "alias does not match Tag"
paragraph_start = max(document.rfind("<w:p>", 0, control_start), document.rfind("<w:p ", 0, control_start))
assert paragraph_start >= 0, "paragraph start not found"
prefix = document[paragraph_start:control_start]
prefix_text = "".join(re.findall(r"<w:t(?:\s[^>]*)?>(.*?)</w:t>", prefix, re.S))
assert prefix_text.endswith(LABEL), "static diagnosis label is not immediately before the dynamic control"

print("COMPETENCY_CUSTOMER_TEMPLATE_TAG_BOUNDARY_TEST_PASS")
