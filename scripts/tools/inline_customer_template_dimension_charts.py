#!/usr/bin/env python3
import argparse
import re
import shutil
import tempfile
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path

RELATIONSHIPS = "{http://schemas.openxmlformats.org/package/2006/relationships}"
ANCHOR = re.compile(r"<wp:anchor\b[^>]*>(.*?)</wp:anchor>", re.S)


def inline_chart(document: str, relationship_id: str, chart_index: int) -> tuple[str, bool]:
    marker = f'r:id="{relationship_id}"'
    anchors = [match for match in ANCHOR.finditer(document) if marker in match.group(0)]
    if not anchors:
        return document, False
    if len(anchors) != 1:
        raise RuntimeError(f"chart{chart_index} anchor count={len(anchors)}")

    match = anchors[0]
    content = match.group(1)
    content = re.sub(r"<wp:simplePos\b[^>]*/>", "", content)
    content = re.sub(r"<wp:position[HV]\b.*?</wp:position[HV]>", "", content, flags=re.S)
    content = re.sub(
        r"<wp:wrap(?:None|Square|Tight|Through|TopAndBottom)\b.*?(?:/>|</wp:wrap(?:Square|Tight|Through|TopAndBottom)>)",
        "",
        content,
        flags=re.S,
    )
    inline = '<wp:inline distT="0" distB="0" distL="0" distR="0">' + content + "</wp:inline>"
    return document[:match.start()] + inline + document[match.end():], True


def rewrite_template(source: Path, target: Path) -> None:
    with zipfile.ZipFile(source, "r") as incoming:
        relationships = ET.fromstring(incoming.read("word/_rels/document.xml.rels"))
        relationship_ids = {
            relation.get("Target", "").replace("\\", "/"): relation.get("Id", "")
            for relation in relationships.findall(RELATIONSHIPS + "Relationship")
        }
        document = incoming.read("word/document.xml").decode("utf-8")
        changed = []
        for index in range(3, 13):
            target_name = f"charts/chart{index}.xml"
            relationship_id = relationship_ids.get(target_name, "")
            if not relationship_id:
                raise RuntimeError(f"chart{index} relationship missing")
            document, converted = inline_chart(document, relationship_id, index)
            if converted:
                changed.append(index)
        if not changed:
            raise RuntimeError("no anchored dimension chart found")

        entries = [(item, incoming.read(item.filename)) for item in incoming.infolist()]

    with tempfile.NamedTemporaryFile(dir=target.parent, suffix=".docx", delete=False) as handle:
        temporary = Path(handle.name)
    try:
        with zipfile.ZipFile(temporary, "w") as outgoing:
            for item, body in entries:
                if item.filename == "word/document.xml":
                    body = document.encode("utf-8")
                outgoing.writestr(item, body)
        temporary.replace(target)
    finally:
        temporary.unlink(missing_ok=True)
    print("INLINED_CHARTS=" + ",".join(map(str, changed)))


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("template", type=Path)
    parser.add_argument("--backup", type=Path, required=True)
    args = parser.parse_args()

    if args.backup.exists():
        raise RuntimeError(f"backup already exists: {args.backup}")
    shutil.copy2(args.template, args.backup)
    rewrite_template(args.template, args.template)


if __name__ == "__main__":
    main()
