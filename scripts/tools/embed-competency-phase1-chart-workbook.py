#!/usr/bin/env python3
"""Create a phase-one DOCX candidate with one embedded chart-data workbook."""

import argparse
import html
import re
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DEFAULT_SOURCE = ROOT / 'Go-based Refactored System' / 'configs' / 'export-templates' / 'competency-phase1-report.docx'
DEFAULT_OUTPUT = ROOT / 'Go-based Refactored System' / 'configs' / 'export-templates' / 'competency-phase1-report-embedded.docx'
WORKBOOK_NAME = 'competency-phase1-chart-data.xlsx'
WORKBOOK_PATH = f'word/embeddings/{WORKBOOK_NAME}'
DIMENSIONS = [
    ('逻辑思维', 3.75), ('数字应用', 4.25), ('计划执行', 3.50), ('持续学习', 4.00),
    ('沟通表达', 3.25), ('敬业奉献', 3.75), ('求真务实', 4.125), ('自律性', 4.125),
    ('成就导向', 3.00), ('合作意识', 3.50),
]
FIXED_ZIP_TIME = (1980, 1, 1, 0, 0, 0)


def write_zip_part(archive, name, data, compress_type=zipfile.ZIP_DEFLATED):
    item = zipfile.ZipInfo(name, FIXED_ZIP_TIME)
    item.compress_type = compress_type
    item.external_attr = 0o600 << 16
    archive.writestr(item, data)


def cell(ref, value):
    if isinstance(value, str):
        return f'<c r="{ref}" t="inlineStr"><is><t>{html.escape(value)}</t></is></c>'
    return f'<c r="{ref}"><v>{value}</v></c>'


def worksheet(rows):
    body = []
    for row_index, values in sorted(rows.items()):
        cells = ''.join(cell(ref, value) for ref, value in values)
        body.append(f'<row r="{row_index}">{cells}</row>')
    return (
        '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
        '<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">'
        '<sheetData>' + ''.join(body) + '</sheetData></worksheet>'
    ).encode('utf-8')


def build_workbook():
    sheet1 = {
        1: [('A1', '一级维度'), ('B1', '得分')],
        2: [('A2', '通用能力'), ('B2', 3.75)],
        3: [('A3', '心理素养'), ('B3', 3.50)],
    }
    sheet2 = {3: [('B3', '得分'), ('C3', '辅助系列')]}
    for index, (name, score) in enumerate(DIMENSIONS):
        row = index + 4
        sheet2[row] = [(f'A{row}', name), (f'B{row}', score), (f'C{row}', 0)]
        doughnut_row = index + 33
        sheet2[doughnut_row] = [(f'A{doughnut_row}', name), (f'B{doughnut_row}', score), (f'C{doughnut_row}', 5 - score)]

    parts = {
        '[Content_Types].xml': b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/><Override PartName="/xl/worksheets/sheet2.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/><Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/></Types>''',
        '_rels/.rels': b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>''',
        'xl/workbook.xml': b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Sheet1" sheetId="1" r:id="rId1"/><sheet name="Sheet2" sheetId="2" r:id="rId2"/></sheets></workbook>''',
        'xl/_rels/workbook.xml.rels': b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>''',
        'xl/styles.xml': b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><fonts count="1"><font><sz val="11"/><name val="Calibri"/></font></fonts><fills count="2"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill></fills><borders count="1"><border/></borders><cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs><cellXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/></cellXfs></styleSheet>''',
        'xl/worksheets/sheet1.xml': worksheet(sheet1),
        'xl/worksheets/sheet2.xml': worksheet(sheet2),
    }
    output = bytearray()
    from io import BytesIO
    stream = BytesIO()
    with zipfile.ZipFile(stream, 'w', zipfile.ZIP_DEFLATED) as archive:
        for name, data in parts.items():
            write_zip_part(archive, name, data)
    output.extend(stream.getvalue())
    return bytes(output)


def replace_chart_relationship(xml):
    pattern = re.compile(r'<Relationship\b[^>]*Type="[^"]*/oleObject"[^>]*/>')
    matches = pattern.findall(xml)
    if len(matches) != 1:
        raise RuntimeError(f'chart external workbook relationship count={len(matches)}, want 1')
    relationship_id = re.search(r'Id="([^"]+)"', matches[0]).group(1)
    replacement = (
        f'<Relationship Id="{relationship_id}" '
        'Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/package" '
        f'Target="../embeddings/{WORKBOOK_NAME}"/>'
    )
    return pattern.sub(replacement, xml, count=1)


def replace_chart_formulas(chart_index, xml):
    xml = re.sub(r'\[[^\]]+\.xlsx\]', f'[{WORKBOOK_NAME}]', xml)
    if chart_index >= 3:
        row = chart_index + 30
        formulas = re.findall(r'<c:f>.*?</c:f>', xml)
        if len(formulas) != 1:
            raise RuntimeError(f'chart{chart_index} formula count={len(formulas)}, want 1')
        xml = re.sub(r'<c:f>.*?</c:f>', f'<c:f>[{WORKBOOK_NAME}]Sheet2!$B${row}:$C${row}</c:f>', xml, count=1)
    return xml


def add_xlsx_content_type(xml):
    if 'Extension="xlsx"' in xml:
        return xml
    marker = '</Types>'
    default = '<Default Extension="xlsx" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"/>'
    if marker not in xml:
        raise RuntimeError('DOCX content types are invalid')
    return xml.replace(marker, default + marker)


def build(source, output):
    workbook = build_workbook()
    output.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(source, 'r') as incoming, zipfile.ZipFile(output, 'w') as outgoing:
        for item in incoming.infolist():
            if item.filename == WORKBOOK_PATH:
                continue
            data = incoming.read(item.filename)
            if item.filename == '[Content_Types].xml':
                data = add_xlsx_content_type(data.decode('utf-8')).encode('utf-8')
            chart_match = re.fullmatch(r'word/charts/chart(\d+)\.xml', item.filename)
            if chart_match:
                data = replace_chart_formulas(int(chart_match.group(1)), data.decode('utf-8')).encode('utf-8')
            rel_match = re.fullmatch(r'word/charts/_rels/chart(\d+)\.xml\.rels', item.filename)
            if rel_match:
                data = replace_chart_relationship(data.decode('utf-8')).encode('utf-8')
            outgoing.writestr(item, data)
        write_zip_part(outgoing, WORKBOOK_PATH, workbook)

    with zipfile.ZipFile(output) as check:
        names = check.namelist()
        if names.count(WORKBOOK_PATH) != 1:
            raise RuntimeError('embedded workbook missing or duplicated')
        for index in range(1, 13):
            rels = check.read(f'word/charts/_rels/chart{index}.xml.rels').decode('utf-8')
            if 'TargetMode="External"' in rels or f'../embeddings/{WORKBOOK_NAME}' not in rels:
                raise RuntimeError(f'chart{index} relationship is not embedded')
    print('PHASE1_EMBEDDED_CHART_TEMPLATE_BUILT')
    print(f'output={output}')
    print('workbooks=1|charts=12|external_links=0')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--source', type=Path, default=DEFAULT_SOURCE)
    parser.add_argument('--output', type=Path, default=DEFAULT_OUTPUT)
    args = parser.parse_args()
    build(args.source.resolve(), args.output.resolve())


if __name__ == '__main__':
    main()
