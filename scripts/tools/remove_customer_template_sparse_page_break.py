#!/usr/bin/env python3
import argparse
import re
import shutil
import tempfile
import zipfile
from pathlib import Path

SECONDARY_HEADING = "二级维度测评结果及建议".encode("utf-8")
TARGET_PARAGRAPH = re.compile(
    rb'<w:p\b(?:(?!</w:p>).)*<w:br w:type="page"/>(?:(?!</w:p>).)*</w:p>'
    rb'(?=<w:p\b(?:(?!</w:p>).)*<w:t>' + re.escape(SECONDARY_HEADING) + rb'</w:t>)',
    re.DOTALL,
)


def rewrite_template(source: Path, target: Path) -> None:
    with zipfile.ZipFile(source, "r") as archive:
        entries = [(info, archive.read(info.filename)) for info in archive.infolist()]

    changed = 0
    with tempfile.NamedTemporaryFile(dir=target.parent, suffix=".docx", delete=False) as handle:
        temporary = Path(handle.name)

    try:
        with zipfile.ZipFile(temporary, "w") as archive:
            for info, data in entries:
                if info.filename == "word/document.xml":
                    changed = len(TARGET_PARAGRAPH.findall(data))
                    if changed != 1:
                        raise RuntimeError(f"expected one sparse-page break paragraph, found {changed}")
                    data = TARGET_PARAGRAPH.sub(b"", data, count=1)
                archive.writestr(info, data)
        temporary.replace(target)
    finally:
        temporary.unlink(missing_ok=True)


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
