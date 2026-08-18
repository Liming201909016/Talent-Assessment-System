#!/usr/bin/env python3
import argparse
import re
import tempfile
import zipfile
from decimal import Decimal
from pathlib import Path

# Measured from the real staging LibreOffice 24.2 PDF at 180 DPI.
PIXEL_OFFSETS = {
    3: (Decimal("-1.5"), Decimal("4.5")),
    4: (Decimal("12.5"), Decimal("-3")),
    5: (Decimal("9.5"), Decimal("-3")),
    6: (Decimal("7.5"), Decimal("-2")),
    7: (Decimal("13"), Decimal("-3.5")),
    8: (Decimal("12.5"), Decimal("-2.5")),
    9: (Decimal("15"), Decimal("0.5")),
    10: (Decimal("14.5"), Decimal("0.5")),
    11: (Decimal("18"), Decimal("2.5")),
    12: (Decimal("13"), Decimal("-2.5")),
}
HORIZONTAL_LAYOUT_PIXELS = Decimal("380")
VERTICAL_LAYOUT_PIXELS = Decimal("220")
VISIBLE_LABEL = re.compile(rb"<c:dLbl><c:idx val=\"0\"/>.*?</c:dLbl>")
X_VALUE = re.compile(rb"<c:x val=\"([^\"]+)\"/>")
Y_VALUE = re.compile(rb"<c:y val=\"([^\"]+)\"/>")


def adjusted(value: bytes, pixel_offset: Decimal, layout_pixels: Decimal) -> bytes:
    coordinate = Decimal(value.decode("ascii")) - pixel_offset / layout_pixels
    return format(coordinate, "f").encode("ascii")


def calibrate(data: bytes, chart_index: int) -> bytes:
    match = VISIBLE_LABEL.search(data)
    if match is None:
        raise RuntimeError(f"chart{chart_index} visible label not found")
    label = match.group(0)
    x = X_VALUE.search(label)
    y = Y_VALUE.search(label)
    if x is None or y is None:
        raise RuntimeError(f"chart{chart_index} label coordinates not found")
    offset_x, offset_y = PIXEL_OFFSETS[chart_index]
    calibrated = X_VALUE.sub(b'<c:x val="' + adjusted(x.group(1), offset_x, HORIZONTAL_LAYOUT_PIXELS) + b'"/>', label, count=1)
    calibrated = Y_VALUE.sub(b'<c:y val="' + adjusted(y.group(1), offset_y, VERTICAL_LAYOUT_PIXELS) + b'"/>', calibrated, count=1)
    return data[:match.start()] + calibrated + data[match.end():]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("source", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    if args.output.exists():
        raise RuntimeError(f"output already exists: {args.output}")

    with zipfile.ZipFile(args.source, "r") as archive:
        entries = [(info, archive.read(info.filename)) for info in archive.infolist()]
    with tempfile.NamedTemporaryFile(dir=args.output.parent, suffix=".docx", delete=False) as handle:
        temporary = Path(handle.name)
    try:
        with zipfile.ZipFile(temporary, "w") as archive:
            for info, data in entries:
                match = re.fullmatch(r"word/charts/chart(\d+)\.xml", info.filename)
                if match and int(match.group(1)) in PIXEL_OFFSETS:
                    data = calibrate(data, int(match.group(1)))
                archive.writestr(info, data)
        temporary.replace(args.output)
    finally:
        temporary.unlink(missing_ok=True)


if __name__ == "__main__":
    main()
