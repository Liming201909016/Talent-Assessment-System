#!/usr/bin/env python3
import argparse
import re
import shutil
import tempfile
import zipfile
from pathlib import Path

VISIBLE_LABEL = re.compile(rb'<c:dLbl><c:idx val="0"/>.*?</c:dLbl>')
NUMBER_FORMAT = re.compile(rb'<c:numFmt\b[^>]*/>')
LAYOUT_END = b'</c:layout>'
TWO_DECIMALS = b'<c:numFmt formatCode="0.00" sourceLinked="0"/>'


def set_two_decimal_format(data: bytes, chart_index: int) -> bytes:
    match = VISIBLE_LABEL.search(data)
    if match is None:
        raise RuntimeError(f"chart{chart_index} visible label not found")
    label = match.group(0)
    if NUMBER_FORMAT.search(label):
        updated = NUMBER_FORMAT.sub(TWO_DECIMALS, label, count=1)
    else:
        if label.count(LAYOUT_END) != 1:
            raise RuntimeError(f"chart{chart_index} visible label layout is ambiguous")
        updated = label.replace(LAYOUT_END, LAYOUT_END + TWO_DECIMALS, 1)
    return data[:match.start()] + updated + data[match.end():]


def rewrite_template(source: Path, target: Path) -> None:
    with zipfile.ZipFile(source, "r") as archive:
        entries = [(info, archive.read(info.filename)) for info in archive.infolist()]

    changed = set()
    with tempfile.NamedTemporaryFile(dir=target.parent, suffix=".docx", delete=False) as handle:
        temporary = Path(handle.name)

    try:
        with zipfile.ZipFile(temporary, "w") as archive:
            for info, data in entries:
                match = re.fullmatch(r"word/charts/chart(\d+)\.xml", info.filename)
                if match and 3 <= int(match.group(1)) <= 12:
                    index = int(match.group(1))
                    updated = set_two_decimal_format(data, index)
                    if updated != data:
                        changed.add(index)
                    data = updated
                archive.writestr(info, data)
        if changed != set(range(3, 13)):
            raise RuntimeError(f"expected chart3-chart12 changes, changed {sorted(changed)}")
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
