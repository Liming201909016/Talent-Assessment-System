#!/usr/bin/env python3
import argparse
import re
import shutil
import tempfile
import zipfile
from decimal import Decimal
from pathlib import Path

VISIBLE_LABEL = re.compile(rb"<c:dLbl><c:idx val=\"0\"/>.*?</c:dLbl>")
X_VALUE = re.compile(rb"<c:x val=\"[^\"]+\"/>")
Y_VALUE = re.compile(rb"<c:y val=\"[^\"]+\"/>")
WIDTH_VALUE = re.compile(rb"<c:w val=\"([^\"]+)\"/>")
HEIGHT_VALUE = re.compile(rb"<c:h val=\"([^\"]+)\"/>")


def centred_coordinate(size: bytes) -> bytes:
    value = -(Decimal(size.decode("ascii")) / Decimal(2))
    return format(value, "f").encode("ascii")


def centre_label(data: bytes, chart_index: int) -> bytes:
    match = VISIBLE_LABEL.search(data)
    if match is None:
        raise RuntimeError(f"chart{chart_index} visible label not found")
    label = match.group(0)
    width = WIDTH_VALUE.search(label)
    height = HEIGHT_VALUE.search(label)
    if width is None or height is None:
        raise RuntimeError(f"chart{chart_index} label size not found")
    if len(X_VALUE.findall(label)) != 1 or len(Y_VALUE.findall(label)) != 1:
        raise RuntimeError(f"chart{chart_index} label position is ambiguous")

    x = centred_coordinate(width.group(1))
    y = centred_coordinate(height.group(1))
    centred = X_VALUE.sub(b'<c:x val="' + x + b'"/>', label, count=1)
    centred = Y_VALUE.sub(b'<c:y val="' + y + b'"/>', centred, count=1)
    return data[:match.start()] + centred + data[match.end():]


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
                    data = centre_label(data, index)
                    changed.add(index)
                archive.writestr(info, data)
        if changed != set(range(3, 13)):
            raise RuntimeError(f"expected chart3-chart12, changed {sorted(changed)}")
        try:
            temporary.replace(target)
        except PermissionError:
            target.write_bytes(temporary.read_bytes())
    finally:
        temporary.unlink(missing_ok=True)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("template", type=Path)
    parser.add_argument("--backup", type=Path)
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()

    target = args.output or args.template
    if target == args.template:
        if args.backup is None:
            raise RuntimeError("--backup is required for in-place updates")
        if args.backup.exists():
            raise RuntimeError(f"backup already exists: {args.backup}")
        shutil.copy2(args.template, args.backup)
    elif target.exists():
        raise RuntimeError(f"output already exists: {target}")
    rewrite_template(args.template, target)


if __name__ == "__main__":
    main()
