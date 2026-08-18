#!/usr/bin/env python3
"""Create the phase-one V2 DOCX with transparent fields and business-keyed charts."""

import argparse
import html
import re
import zipfile
from io import BytesIO
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DEFAULT_SOURCE = ROOT / 'Go-based Refactored System' / 'configs' / 'export-templates' / 'competency-phase1-report.docx'
DEFAULT_OUTPUT = ROOT / 'Go-based Refactored System' / 'configs' / 'export-templates' / 'competency-phase1-report-embedded.docx'
WORKBOOK_NAME = 'competency-phase1-chart-data.xlsx'
WORKBOOK_PATH = f'word/embeddings/{WORKBOOK_NAME}'
SCHEMA_VERSION = 'competency-phase1-template-schema-v2'
FIXED_ZIP_TIME = (1980, 1, 1, 0, 0, 0)

DIMENSIONS = [
    ('competency-a1-01', '逻辑思维', 3.75), ('competency-a1-02', '数字应用', 4.25),
    ('competency-a1-03', '计划执行', 3.50), ('competency-a1-04', '持续学习', 4.00),
    ('competency-a1-05', '沟通表达', 3.25), ('competency-b1-01', '敬业奉献', 3.75),
    ('competency-b1-02', '求真务实', 4.125), ('competency-b1-03', '自律性', 4.125),
    ('competency-b1-04', '成就导向', 3.00), ('competency-b1-05', '合作意识', 3.50),
]
CHART_KEYS = [
    ('chart.group.overview', '一级维度概览图'),
    ('chart.dimension.radar', '十个二级维度雷达图'),
] + [(f'chart.dimension.{dimension_id}', f'{name}环形图') for dimension_id, name, _ in DIMENSIONS]

FIELDS = [
    ('report.date', '报告日期', 'text', True, True, '2026年8月13日', '报告生成日期'),
    ('report.generatedAt', '报告生成时间', 'text', False, True, '2026年8月13日', '报告生成日期，可选字段'),
    ('participant.name', '姓名', 'text', True, True, '张三', '提交时冻结的人员姓名'),
    ('participant.age', '年龄', 'text', True, True, '30', '提交时冻结的人员年龄'),
    ('participant.gender', '性别', 'text', True, True, '男', '提交时冻结的人员性别'),
    ('participant.telephone', '手机号', 'text', True, True, '13800000000', '提交时冻结的人员手机号'),
    ('participant.affiliation', '单位', 'text', True, True, '示例单位', '提交时冻结的人员单位'),
    ('participant.post', '岗位', 'text', True, True, '示例岗位', '提交时冻结的人员岗位'),
    ('participant.degree', '学历', 'text', False, True, '本科', '提交时冻结的人员学历，可选字段'),
    ('participant.major', '专业', 'text', False, True, '计算机科学', '提交时冻结的人员专业，可选字段'),
    ('result.submittedAt', '提交日期', 'text', True, True, '2026年8月13日', '答卷提交日期'),
    ('result.userTime', '作答时长', 'text', True, True, '20', '报告数据中的作答分钟数'),
    ('overall.score', '总体得分', 'decimal', False, True, '35.00', '十个二级维度得分之和，满分50'),
    ('overall.maxScore', '总体满分', 'decimal', False, True, '50', '一期总体固定满分'),
    ('overall.percentage', '总体百分比', 'decimal', False, True, '70.00', '总体得分占50分满分的百分比'),
    ('overall.level', '总体等级', 'text', True, True, '合格胜任', '一期总体五档等级'),
    ('overall.diagnosis', '总体诊断', 'text', True, False, '此处显示总体诊断。', '正式内容快照中的总体诊断'),
    ('validity.notice', '效度说明', 'text', True, True, '本次作答效度良好。', '正式内容快照中的效度提示'),
    ('report.disclaimer', '免责声明', 'text', True, False, '此处显示正式免责声明。', '经批准的正式免责声明'),
]
for code, label in [('general_ability', '通用能力'), ('psychological_quality', '心理素养')]:
    FIELDS.extend([
        (f'group.{code}.score', f'{label}得分', 'decimal', True, True, '3.50', f'{label}下五个二级维度的平均分'),
        (f'group.{code}.level', f'{label}等级', 'text', True, True, '较高分', f'{label}五档等级'),
        (f'group.{code}.description', f'{label}说明', 'text', True, False, '此处显示一级维度说明。', '正式内容快照中的一级维度说明'),
    ])
for dimension_id, dimension_name, _ in DIMENSIONS:
    label = dimension_name
    FIELDS.extend([
        (f'dimension.{dimension_id}.score', f'{label}得分', 'decimal', True, True, '3.50', '该二级维度8题平均分，满分5'),
        (f'dimension.{dimension_id}.remainingScore', f'{label}距满分差值', 'decimal', False, True, '1.50', '5减去该二级维度得分，用于环形图'),
        (f'dimension.{dimension_id}.percentage', f'{label}百分比', 'decimal', False, True, '70.00', '该二级维度得分占5分满分的百分比'),
        (f'dimension.{dimension_id}.level', f'{label}等级', 'text', True, True, '合格', '该二级维度五档等级'),
        (f'dimension.{dimension_id}.diagnosis', f'{label}诊断', 'text', True, False, '此处显示诊断与发展建议。', '正式内容快照中的维度诊断与建议'),
    ])


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
    dictionary = {
        1: [('A1', 'schemaVersion'), ('B1', SCHEMA_VERSION)],
        3: [('A3', 'BusinessKey'), ('B3', '中文名称'), ('C3', '类型'), ('D3', '必需'),
            ('E3', '可重复'), ('F3', '示例值'), ('G3', '说明')],
    }
    for index, field in enumerate(FIELDS, start=4):
        key, label, value_type, required, repeatable, example, description = field
        dictionary[index] = [
            (f'A{index}', key), (f'B{index}', label), (f'C{index}', value_type),
            (f'D{index}', '是' if required else '否'), (f'E{index}', '是' if repeatable else '否'),
            (f'F{index}', example), (f'G{index}', description),
        ]
    chart_data = {
        1: [('A1', 'BusinessKey'), ('B1', '名称'), ('C1', 'Score'), ('D1', 'Remainder')],
        2: [('A2', 'group.general_ability'), ('B2', '通用能力'), ('C2', 3.75), ('D2', 0)],
        3: [('A3', 'group.psychological_quality'), ('B3', '心理素养'), ('C3', 3.50), ('D3', 0)],
    }
    for index, (dimension_id, name, score) in enumerate(DIMENSIONS, start=4):
        chart_data[index] = [
            (f'A{index}', dimension_id), (f'B{index}', name),
            (f'C{index}', score), (f'D{index}', 5 - score),
        ]
    parts = {
        '[Content_Types].xml': b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/><Override PartName="/xl/worksheets/sheet2.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/><Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/></Types>''',
        '_rels/.rels': b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>''',
        'xl/workbook.xml': b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="FieldDictionary" sheetId="1" r:id="rId1"/><sheet name="ChartData" sheetId="2" r:id="rId2"/></sheets></workbook>''',
        'xl/_rels/workbook.xml.rels': b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>''',
        'xl/styles.xml': b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><fonts count="1"><font><sz val="11"/><name val="Calibri"/></font></fonts><fills count="2"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill></fills><borders count="1"><border/></borders><cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs><cellXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/></cellXfs></styleSheet>''',
        'xl/worksheets/sheet1.xml': worksheet(dictionary),
        'xl/worksheets/sheet2.xml': worksheet(chart_data),
    }
    stream = BytesIO()
    with zipfile.ZipFile(stream, 'w', zipfile.ZIP_DEFLATED) as archive:
        for name, data in parts.items():
            write_zip_part(archive, name, data)
    return stream.getvalue()


def add_business_chart_titles(document):
    chart_index = 0
    drawing_pattern = re.compile(r'<wp:(?:anchor|inline)\b.*?</wp:(?:anchor|inline)>', re.S)

    def replace_drawing(match):
        nonlocal chart_index
        drawing = match.group(0)
        if not re.search(r'<c:chart\b[^>]*\br:id="[^"]+"[^>]*/>', drawing):
            return drawing
        if chart_index >= len(CHART_KEYS):
            raise RuntimeError('document contains more than 12 charts')
        key = CHART_KEYS[chart_index][0]
        chart_index += 1
        doc_pr = re.search(r'<wp:docPr\b[^>]*/>', drawing)
        if not doc_pr:
            raise RuntimeError(f'chart object has no wp:docPr: {key}')
        updated = re.sub(r'\s+title="[^"]*"', '', doc_pr.group(0))
        updated = updated[:-2] + f' title="{key}"/>'
        return drawing[:doc_pr.start()] + updated + drawing[doc_pr.end():]

    result = drawing_pattern.sub(replace_drawing, document)
    if chart_index != len(CHART_KEYS):
        raise RuntimeError(f'document chart count={chart_index}, want {len(CHART_KEYS)}')
    return result


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
    prefix = f'[{WORKBOOK_NAME}]ChartData!'
    if chart_index == 1:
        formulas = [prefix + '$B$2:$B$3', prefix + '$C$2:$C$3']
    elif chart_index == 2:
        formulas = [
            prefix + '$C$1', prefix + '$B$4:$B$13', prefix + '$C$4:$C$13',
            prefix + '$D$1', prefix + '$B$4:$B$13', prefix + '$D$4:$D$13',
        ]
    else:
        row = chart_index + 1
        formulas = [prefix + f'$C${row}:$D${row}']
    matches = list(re.finditer(r'<c:f>.*?</c:f>', xml))
    if len(matches) != len(formulas):
        raise RuntimeError(f'chart{chart_index} formula count={len(matches)}, want {len(formulas)}')
    for match, formula in reversed(list(zip(matches, formulas))):
        xml = xml[:match.start()] + f'<c:f>{formula}</c:f>' + xml[match.end():]
    return xml


def add_xlsx_content_type(xml):
    part_name = f'/{WORKBOOK_PATH}'
    if f'PartName="{part_name}"' in xml:
        return xml
    marker = '</Types>'
    override = (
        f'<Override PartName="{part_name}" '
        'ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"/>'
    )
    if marker not in xml:
        raise RuntimeError('DOCX content types are invalid')
    return xml.replace(marker, override + marker)


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
            elif item.filename == 'word/document.xml':
                data = add_business_chart_titles(data.decode('utf-8')).encode('utf-8')
            chart_match = re.fullmatch(r'word/charts/chart(\d+)\.xml', item.filename)
            if chart_match:
                data = replace_chart_formulas(int(chart_match.group(1)), data.decode('utf-8')).encode('utf-8')
            rel_match = re.fullmatch(r'word/charts/_rels/chart(\d+)\.xml\.rels', item.filename)
            if rel_match:
                data = replace_chart_relationship(data.decode('utf-8')).encode('utf-8')
            outgoing.writestr(item, data)
        write_zip_part(outgoing, WORKBOOK_PATH, workbook)

    with zipfile.ZipFile(output) as check:
        document = check.read('word/document.xml').decode('utf-8')
        titles = re.findall(r'<wp:docPr\b[^>]*\btitle="(chart\.[a-zA-Z0-9_.-]+)"[^>]*/>', document)
        if titles != [key for key, _ in CHART_KEYS]:
            raise RuntimeError(f'business chart keys are invalid: {titles}')
        if check.namelist().count(WORKBOOK_PATH) != 1:
            raise RuntimeError('embedded workbook missing or duplicated')
        for index in range(1, 13):
            rels = check.read(f'word/charts/_rels/chart{index}.xml.rels').decode('utf-8')
            if 'TargetMode="External"' in rels or f'../embeddings/{WORKBOOK_NAME}' not in rels:
                raise RuntimeError(f'chart{index} relationship is not embedded')
    print('PHASE1_TEMPLATE_V2_BUILT')
    print(f'output={output}')
    print(f'schema={SCHEMA_VERSION}|fields={len(FIELDS)}|business_charts=12|workbooks=1|external_links=0')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--source', type=Path, default=DEFAULT_SOURCE)
    parser.add_argument('--output', type=Path, default=DEFAULT_OUTPUT)
    args = parser.parse_args()
    build(args.source.resolve(), args.output.resolve())


if __name__ == '__main__':
    main()
