#!/usr/bin/env python3
"""
Small MBTI DOCX experiment: replace only the run properties (w:rPr) of generated cover
value runs with a minimal Microsoft YaHei style, then save a new DOCX.

This is a diagnostic tool under scripts/tools/; it does not affect production code.
"""
from __future__ import annotations

import re
import sys
import zipfile
from pathlib import Path

def min_rpr(font: str) -> str:
    return (
        '<w:rPr>'
        f'<w:rFonts w:hint="eastAsia" w:ascii="{font}" '
        f'w:hAnsi="{font}" w:eastAsia="{font}"/>'
        '<w:color w:val="75BD42"/>'
        '<w:sz w:val="30"/><w:szCs w:val="30"/>'
        '</w:rPr>'
    )

RUN_RE = re.compile(r'<w:r[^>]*>[\s\S]*?</w:r>')
RPR_RE = re.compile(r'<w:rPr>[\s\S]*?</w:rPr>')
WT_RE = re.compile(r'<w:t[^>]*>([^<]*)</w:t>')
W14_TEXT_FILL_RE = re.compile(r'<w14:textFill>[\s\S]*?</w14:textFill>')
W14_PROPS3D_RE = re.compile(r'<w14:props3d[^>]*>[\s\S]*?</w14:props3d>|<w14:props3d[^>]*/>')


def run_text(run: str) -> str:
    return ''.join(WT_RE.findall(run))


def normalize_run(run: str, font: str) -> str:
    run = W14_TEXT_FILL_RE.sub('', run)
    run = W14_PROPS3D_RE.sub('', run)
    if RPR_RE.search(run):
        return RPR_RE.sub(min_rpr(font), run, count=1)
    insert_at = run.find('>') + 1
    return run[:insert_at] + min_rpr(font) + run[insert_at:]


def main() -> int:
    if len(sys.argv) < 5:
        print('usage: mbti_min_rpr_experiment.py input.docx output.docx font value1 [value2 ...]', file=sys.stderr)
        return 2

    src = Path(sys.argv[1])
    dst = Path(sys.argv[2])
    font = sys.argv[3]
    values = set(sys.argv[4:])

    with zipfile.ZipFile(src, 'r') as zin:
        document = zin.read('word/document.xml').decode('utf-8')
        changed = 0

        def repl(match: re.Match[str]) -> str:
            nonlocal changed
            run = match.group(0)
            text = run_text(run)
            if text in values:
                changed += 1
                return normalize_run(run, font)
            return run

        new_document = RUN_RE.sub(repl, document)

        with zipfile.ZipFile(dst, 'w', zipfile.ZIP_DEFLATED) as zout:
            for item in zin.infolist():
                data = zin.read(item.filename)
                if item.filename == 'word/document.xml':
                    data = new_document.encode('utf-8')
                zout.writestr(item, data)

    print(f'changed_runs={changed}')
    return 0 if changed > 0 else 1


if __name__ == '__main__':
    raise SystemExit(main())
