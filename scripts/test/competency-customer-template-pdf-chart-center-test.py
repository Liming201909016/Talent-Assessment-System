#!/usr/bin/env python3
# FB-140: docs/regression-tests.md
# Reproduction: Word centres chart values, but LibreOffice 24.2 shifts them in PDF.
# Expected: every rendered score centre is within 2 pixels of its doughnut-hole centre at 180 DPI.
import argparse
import math
import shutil
import subprocess
import tempfile
from collections import deque
from pathlib import Path


def read_ppm(path: Path):
    data = path.read_bytes()
    position = 0
    tokens = []
    while len(tokens) < 4:
        while data[position:position + 1].isspace():
            position += 1
        if data[position:position + 1] == b"#":
            position = data.index(b"\n", position) + 1
            continue
        end = position
        while not data[end:end + 1].isspace():
            end += 1
        tokens.append(data[position:end])
        position = end
    while data[position:position + 1].isspace():
        position += 1
    assert tokens[0] == b"P6" and tokens[3] == b"255", f"unsupported PPM header: {tokens}"
    width, height = int(tokens[1]), int(tokens[2])
    pixels = data[position:]
    assert len(pixels) == width * height * 3
    return width, height, pixels


def colour_mask(width, height, pixels):
    mask = set()
    for y in range(height):
        row = y * width * 3
        for x in range(width):
            offset = row + x * 3
            red, green, blue = pixels[offset:offset + 3]
            if max(red, green, blue) >= 175 and min(red, green, blue) >= 90 and max(red, green, blue) - min(red, green, blue) >= 25:
                mask.add((x, y))
    dilated = set(mask)
    for x, y in mask:
        for dy in range(-2, 3):
            for dx in range(-2, 3):
                dilated.add((x + dx, y + dy))
    return dilated


def doughnuts(width, height, pixels):
    remaining = colour_mask(width, height, pixels)
    circles = []
    while remaining:
        start = remaining.pop()
        queue = deque([start])
        min_x = max_x = start[0]
        min_y = max_y = start[1]
        count = 0
        while queue:
            x, y = queue.popleft()
            count += 1
            min_x, max_x = min(min_x, x), max(max_x, x)
            min_y, max_y = min(min_y, y), max(max_y, y)
            for neighbour in ((x - 1, y), (x + 1, y), (x, y - 1), (x, y + 1)):
                if neighbour in remaining:
                    remaining.remove(neighbour)
                    queue.append(neighbour)
        box_width, box_height = max_x - min_x + 1, max_y - min_y + 1
        if count >= 5000 and 140 <= box_width <= 260 and 140 <= box_height <= 260 and abs(box_width - box_height) <= 12:
            circles.append((min_x, min_y, max_x, max_y))
    return sorted(circles, key=lambda box: (box[1], box[0]))


def dark_text_centre(width, pixels, box):
    min_x, min_y, max_x, max_y = box
    centre_x = (min_x + max_x) / 2
    centre_y = (min_y + max_y) / 2
    radius = min(max_x - min_x + 1, max_y - min_y + 1) * 0.27
    points = []
    for y in range(math.ceil(centre_y - radius), math.floor(centre_y + radius) + 1):
        for x in range(math.ceil(centre_x - radius), math.floor(centre_x + radius) + 1):
            if (x - centre_x) ** 2 + (y - centre_y) ** 2 > radius ** 2:
                continue
            offset = (y * width + x) * 3
            red, green, blue = pixels[offset:offset + 3]
            if max(red, green, blue) <= 145:
                points.append((x, y))
    assert len(points) >= 40, f"score text not found inside doughnut {box}"
    text_min_x = min(point[0] for point in points)
    text_max_x = max(point[0] for point in points)
    text_min_y = min(point[1] for point in points)
    text_max_y = max(point[1] for point in points)
    return centre_x, centre_y, (text_min_x + text_max_x) / 2, (text_min_y + text_max_y) / 2


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("pdf", type=Path)
    parser.add_argument("--first-page", type=int, default=6)
    parser.add_argument("--last-page", type=int, default=11)
    parser.add_argument("--tolerance", type=float, default=2.0)
    args = parser.parse_args()

    pdftoppm = shutil.which("pdftoppm")
    if not pdftoppm:
        raise RuntimeError("pdftoppm is required")
    with tempfile.TemporaryDirectory() as directory:
        prefix = Path(directory) / "page"
        subprocess.run([
            pdftoppm, "-r", "180", "-f", str(args.first_page), "-l", str(args.last_page),
            str(args.pdf), str(prefix),
        ], check=True, stdout=subprocess.DEVNULL)
        results = []
        for page in sorted(Path(directory).glob("page-*.ppm")):
            width, height, pixels = read_ppm(page)
            for box in doughnuts(width, height, pixels):
                ring_x, ring_y, text_x, text_y = dark_text_centre(width, pixels, box)
                results.append((page.name, ring_x, ring_y, text_x, text_y, text_x - ring_x, text_y - ring_y))

    assert len(results) == 10, f"rendered doughnut count={len(results)}, want 10; results={results}"
    failures = [result for result in results if abs(result[5]) > args.tolerance or abs(result[6]) > args.tolerance]
    assert not failures, f"rendered score labels are off-centre: {failures}; all={results}"
    print("COMPETENCY_CUSTOMER_TEMPLATE_PDF_CHART_CENTER_TEST_PASS")


if __name__ == "__main__":
    main()
