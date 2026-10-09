#!/usr/bin/env python3
"""Bounded, offline 002 candidate builder and synthetic fill demonstration.

No runtime renderer, database, source edits, shared-tool calls or activation.
Only complete cached chart data is materialized; never evaluate workbook links.
"""

import argparse
import copy
from decimal import Decimal, ROUND_HALF_UP
from fractions import Fraction
import hashlib
import io
import json
from pathlib import Path
import posixpath
import re
import sys
import xml.etree.ElementTree as ET
import zipfile

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / 'docs/260929管理特质测评-优化/260928管理潜质测评报告模板修改稿V2.8.docx'
WORKBOOK = SOURCE.parent / '260928测评内容+数据图.xlsx'
SOURCE_SHA = 'c82c2dc0cdcaec413866268c557561d5c53abc41ec02d94dc6aea615edca2f48'
WORKBOOK_SHA = 'b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c'
OUT = ROOT / 'docs/generated/management-traits-word-candidate-20261001'
NS = {
    'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main',
    'wp': 'http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing',
    'a': 'http://schemas.openxmlformats.org/drawingml/2006/main',
    'c': 'http://schemas.openxmlformats.org/drawingml/2006/chart',
    'r': 'http://schemas.openxmlformats.org/officeDocument/2006/relationships',
    'wpg': 'http://schemas.microsoft.com/office/word/2010/wordprocessingGroup',
    'wps': 'http://schemas.microsoft.com/office/word/2010/wordprocessingShape',
    'mc': 'http://schemas.openxmlformats.org/markup-compatibility/2006',
    'w14': 'http://schemas.microsoft.com/office/word/2010/wordml',
    's': 'http://schemas.openxmlformats.org/spreadsheetml/2006/main',
}
DIMENSIONS = [
    ('self_confidence', '自信心', 'self', 12, '57.50'),
    ('emotional_stability', '情绪稳定性', 'self', 10, '55.00'),
    ('self_discipline', '自律性', 'self', 10, '56.25'),
    ('sociality', '社会性', 'interpersonal', 10, '51.25'),
    ('leadership', '领导性', 'interpersonal', 10, '52.50'),
    ('interpersonal_sensitivity', '人际敏感性', 'interpersonal', 13, '50.00'),
    ('cooperation', '合作性', 'interpersonal', 10, '57.50'),
    ('planning', '计划性', 'task', 12, '53.75'),
    ('responsibility', '责任心', 'task', 10, '58.75'),
    ('decisiveness', '决断性', 'task', 12, '53.75'),
    ('proactiveness', '进取性', 'development', 10, '55.00'),
    ('learning', '学习力', 'development', 11, '53.75'),
    ('innovation', '创新性', 'development', 10, '50.00'),
]
LABEL_KEYS = {'综合均值': 'chart.overall', '自我管理均值': 'chart.module.self',
              '人际管理均值': 'chart.module.interpersonal', '任务管理均值': 'chart.module.task',
              '发展管理均值': 'chart.module.development'}
# Read-only Word 16 DataLabel/ChartArea bounds from the SHA-locked source,
# in points, before any candidate changes. Negative OOXML dLbl x/y are
# offsets from a score-dependent pie point, not chart-area coordinates.
SOURCE_LABEL_BOUNDS = {
    'chart.overall': ('41.0047244094488', '45.2562204724409', '69.8237795275591', '50.0034645669291', '155.9'),
    'chart.module.self': ('3.5', '24.5799212598425', '69.7', '27.25', '85.05'),
    'chart.module.task': ('5.00007874015748', '24.8429921259842', '66.75', '27.3', '85.05'),
    'chart.module.development': ('5', '25.3985826771654', '70.5', '27.3', '85.05'),
    'chart.module.interpersonal': ('5', '25.7012598425197', '69.7', '28.95', '85.05'),
}
# Source Word Anchor.Information(6): Group 67=180.70 pt, Group 20=364.65 pt.
# Their shared paragraph has two automatically wrapped inline lines.
SOURCE_GROUP_LINE_OFFSETS = {'67': '0', '20': '183.95'}


def q(name):
    prefix, local = name.split(':')
    return '{' + NS[prefix] + '}' + local


def sha(data):
    return hashlib.sha256(data).hexdigest()


def read_package(data):
    with zipfile.ZipFile(io.BytesIO(data)) as archive:
        if len(archive.namelist()) != len(set(archive.namelist())) or archive.testzip():
            raise ValueError('duplicate or damaged ZIP entries')
        return {name: archive.read(name) for name in archive.namelist()}


def pack(parts):
    buffer = io.BytesIO()
    with zipfile.ZipFile(buffer, 'w', compression=zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
        for name in sorted(parts):
            info = zipfile.ZipInfo(name, (1980, 1, 1, 0, 0, 0))
            info.compress_type = zipfile.ZIP_DEFLATED
            info.external_attr = 0o600 << 16
            archive.writestr(info, parts[name])
    return buffer.getvalue()


def serialize(root, original):
    # Preserve namespace prefixes also used only in mc:Ignorable/Requires values.
    declarations = dict(value for event, value in ET.iterparse(io.BytesIO(original), events=['start-ns']))
    for prefix, uri in declarations.items():
        if not re.fullmatch(r'ns\d+', prefix):
            ET.register_namespace(prefix, uri)
    for prefix, uri in NS.items():
        if uri in declarations.values():
            ET.register_namespace(prefix, uri)
    text = ET.tostring(root, encoding='unicode')
    start = text[:text.index('>')]
    missing = []
    for prefix, uri in declarations.items():
        attribute = 'xmlns' + (':' + prefix if prefix else '')
        if attribute + '=' not in start:
            missing.append(' ' + attribute + '="' + uri + '"')
    if missing:
        pos = text.index('>')
        if text[pos-1] == '/':
            pos -= 1
        text = text[:pos] + ''.join(missing) + text[pos:]
    return ('<?xml version="1.0" encoding="UTF-8" standalone="yes"?>\n' + text).encode('utf-8')


def paths(root):
    result = {}
    prefixes = {uri: prefix for prefix, uri in NS.items()}
    def walk(node, path):
        result[node] = path
        counts = {}
        for child in node:
            counts[child.tag] = counts.get(child.tag, 0) + 1
            uri, local = child.tag[1:].split('}')
            name = prefixes.get(uri, '*') + ':' + local
            walk(child, path + '/' + name + '[' + str(counts[child.tag]) + ']')
    walk(root, '/w:document[1]')
    return result


def wtext(node):
    return ''.join(x.text or '' for x in node.iter(q('w:t')))


def run_fragment(run, text, prefix='w'):
    fragment = copy.deepcopy(run)
    for child in list(fragment):
        if child.tag != q(prefix + ':rPr'):
            fragment.remove(child)
    t = ET.SubElement(fragment, q(prefix + ':t'))
    t.set('{http://www.w3.org/XML/1998/namespace}space', 'preserve')
    t.text = text
    return fragment


def bold(run, value):
    props = run.find('w:rPr', NS)
    if props is None:
        props = ET.Element(q('w:rPr'))
        run.insert(0, props)
    for tag in ('w:b', 'w:bCs'):
        node = props.find(tag, NS)
        if node is None:
            node = ET.SubElement(props, q(tag))
        node.set(q('w:val'), '1' if value else '0')


def local_page_flow(doc):
    def flag(p, name, value):
        props = p.find('w:pPr', NS)
        if props is None:
            props = ET.Element(q('w:pPr')); p.insert(0, props)
        node = props.find('w:'+name, NS)
        if node is None:
            node = ET.SubElement(props, q('w:'+name))
        node.set(q('w:val'), '1' if value else '0')

    tables = doc.findall('.//w:tbl', NS)
    for table in tables[5:18]:
        rows = table.findall('w:tr', NS)
        if len(rows) != 3:
            raise ValueError('detail flow requires the verified three-row module')
        for index, row in enumerate(rows):
            props = row.find('w:trPr', NS)
            if props is None:
                props = ET.Element(q('w:trPr')); row.insert(0, props)
            split = props.find('w:cantSplit', NS)
            if split is None:
                split = ET.SubElement(props, q('w:cantSplit'))
            split.set(q('w:val'), '1')
            for p in row.findall('w:tc/w:p', NS):
                flag(p, 'keepLines', True)
                flag(p, 'keepNext', index < len(rows)-1)
    body = list(doc.find('w:body', NS))
    advice_start = next(i for i,p in enumerate(body) if wtext(p) == '★ 各维度发展建议')
    for i,p in enumerate(body):
        text = wtext(p)
        if p.tag != q('w:p'):
            continue
        if text.startswith('管理特质详细分析'):
            flag(p, 'pageBreakBefore', False)
            flag(p, 'keepNext', True)
        if text.startswith(('★ 管理特质综合表现','呈现作答者','◆ ')) or text in (
                '★ 总体发展建议','★ 各维度发展建议'):
            flag(p, 'keepNext', True)
            flag(p, 'keepLines', True)
        if i > advice_start and text in [d[1] for d in DIMENSIONS]:
            flag(p, 'keepNext', True)
            flag(p, 'keepLines', True)
            flag(body[i+1], 'keepLines', True)
            flag(body[i+1], 'keepNext', False)
    row = tables[19].find('w:tr', NS)
    props = row.find('w:trPr', NS)
    if props is None:
        props = ET.Element(q('w:trPr')); row.insert(0, props)
    split = props.find('w:cantSplit', NS)
    if split is None:
        split = ET.SubElement(props, q('w:cantSplit'))
    split.set(q('w:val'), '1')
    paragraphs = row.findall('w:tc/w:p', NS)
    for i,p in enumerate(paragraphs):
        flag(p, 'keepLines', True)
        flag(p, 'keepNext', i < len(paragraphs)-1)


def bind_span(paragraph, start, end, tag, manifest, source_paths, force_bold=None, *, source_layout=False):
    if not 0 <= start < end <= len(wtext(paragraph)) or tag in manifest['fields']:
        raise ValueError('invalid or duplicate field span: ' + tag)
    sdt = ET.Element(q('w:sdt'))
    props = ET.SubElement(sdt, q('w:sdtPr'))
    ET.SubElement(props, q('w:alias')).set(q('w:val'), tag)
    ET.SubElement(props, q('w:tag')).set(q('w:val'), tag)
    ET.SubElement(props, q('w:id')).set(q('w:val'), str(2000 + len(manifest['fields'])))
    content = ET.SubElement(sdt, q('w:sdtContent'))
    offset = 0
    inserted = False
    for run in list(paragraph):
        if run.tag != q('w:r'):
            if source_layout and inserted and start < offset < end:
                if run.tag not in (q('w:bookmarkStart'), q('w:bookmarkEnd')):
                    raise ValueError('unverified non-run node inside source binding: ' + tag)
                paragraph.remove(run)
                content.append(run)
            continue
        value = wtext(run)
        lo, hi = max(start - offset, 0), min(end - offset, len(value))
        offset += len(value)
        if lo >= hi:
            continue
        if any(child.tag not in (q('w:rPr'), q('w:t')) for child in run):
            raise ValueError('mixed non-text run in dynamic span: ' + tag)
        index = list(paragraph).index(run)
        if lo:
            paragraph.insert(index, run_fragment(run, value[:lo])); index += 1
        if not inserted:
            paragraph.insert(index, sdt); inserted = True; index += 1
        piece = run_fragment(run, value[lo:hi])
        if force_bold is not None:
            bold(piece, force_bold)
        content.append(piece)
        if hi < len(value):
            paragraph.insert(index, run_fragment(run, value[hi:]))
        paragraph.remove(run)
    if not inserted:
        raise ValueError('no actual text for ' + tag)
    manifest['fields'][tag] = [{'part': 'word/document.xml',
                              'source_xpath': source_paths[paragraph],
                              'source_char_range': [start, end],
                              'sample': wtext(sdt), 'status': 'candidate_pending_approval'}]
    return sdt


def bind_document(doc, manifest, *, source_layout=False):
    source_paths = paths(doc)
    tables = doc.findall('.//w:tbl', NS)
    if len(tables) != 20 or doc.findall('.//w:sdt', NS):
        raise ValueError('unexpected source table/control structure')
    def cell(table, row, col):
        return tables[table-1].findall('w:tr', NS)[row-1].findall('w:tc', NS)[col-1]
    def whole(p, tag, **kwargs):
        return bind_span(p, 0, len(wtext(p)), tag, manifest, source_paths,
                         source_layout=source_layout, **kwargs)
    for row, col, tag in [(1, 2, 'participant.name'), (1, 4, 'participant.gender'),
                          (2, 2, 'participant.affiliation'), (2, 4, 'participant.post'),
                          (3, 2, 'participant.telephone'), (3, 4, 'result.submittedAt')]:
        whole(cell(1, row, col).find('w:p', NS), tag)
    for group, col in [('high', 1), ('low', 2)]:
        paragraphs = cell(4, 1, col).findall('w:p', NS)
        if len(paragraphs) != 4:
            raise ValueError('summary must have heading plus three slots')
        for index, p in enumerate(paragraphs):
            if source_layout:
                continue
            props = p.find('w:pPr', NS)
            if props is None:
                props = ET.Element(q('w:pPr')); p.insert(0, props)
            if props.find('w:keepLines', NS) is None:
                ET.SubElement(props, q('w:keepLines'))
            props.find('w:keepLines', NS).set(q('w:val'), '1')
            if index < len(paragraphs)-1 and props.find('w:keepNext', NS) is None:
                ET.SubElement(props, q('w:keepNext'))
            if index < len(paragraphs)-1:
                props.find('w:keepNext', NS).set(q('w:val'), '1')
        for i, p in enumerate(paragraphs[1:], 1):
            text = wtext(p); colon = text.index('：')
            bind_span(p, colon+1, len(text), f'overview.{group}.{i}.text', manifest, source_paths,
                      None if source_layout else False, source_layout=source_layout)
            bind_span(p, 0, colon, f'overview.{group}.{i}.name', manifest, source_paths,
                      None if source_layout else True, source_layout=source_layout)
            for run in p.findall('w:r', NS):
                if wtext(run) == '：' and not source_layout:
                    bold(run, True)
    for i, (key, name, module, count, norm) in enumerate(DIMENSIONS):
        if wtext(cell(6+i, 1, 1)) != name:
            raise ValueError('dimension position/name mismatch')
        p = cell(6+i, 1, 2).find('w:p', NS); text = wtext(p)
        m = re.fullmatch(r'得分：(\d+\.\d+)分\s+等级：(.+)', text)
        if not m:
            raise ValueError('unexpected score/level label boundary')
        for suffix, group in [('level', 2), ('score', 1)]:
            bind_span(p, *m.span(group), f'dimension.{key}.{suffix}', manifest, source_paths,
                      source_layout=source_layout)
        p = cell(6+i, 1, 3).find('w:p', NS)
        m = re.fullmatch(r'常模分：(\d+\.\d+)分', wtext(p))
        if not m:
            raise ValueError('unexpected norm boundary')
        bind_span(p, *m.span(1), f'dimension.{key}.norm', manifest, source_paths,
              source_layout=source_layout)
        whole(cell(6+i, 3, 2).find('w:p', NS), f'dimension.{key}.diagnosis')
    body = doc.find('w:body', NS)
    children = list(body)
    for label, tag in [('【等级】', 'overall.level'), ('【诊断】', 'overall.diagnosis')]:
        found = [p for p in children if p.tag == q('w:p') and wtext(p).startswith(label)]
        if len(found) != 1:
            raise ValueError('overall label must identify one actual paragraph')
        p = found[0]; text = wtext(p)
        if tag == 'overall.diagnosis' and not source_layout:
            props = p.find('w:pPr', NS)
            if props is None:
                props = ET.Element(q('w:pPr')); p.insert(0, props)
            if props.find('w:keepLines', NS) is None:
                ET.SubElement(props, q('w:keepLines'))
            props.find('w:keepLines', NS).set(q('w:val'), '1')
        end = text.index('  ') if tag.endswith('level') else len(text)
        bind_span(p, len(label), end, tag, manifest, source_paths, source_layout=source_layout)
    advice_heading = next(i for i, p in enumerate(children) if wtext(p) == '★ 各维度发展建议')
    for key, name, *_ in DIMENSIONS:
        indexes = [i for i in range(advice_heading+1, len(children)) if wtext(children[i]) == name]
        if len(indexes) != 1 or children[indexes[0]+1].tag != q('w:p'):
            raise ValueError('dimension advice position ambiguous')
        whole(children[indexes[0]+1], f'dimension.{key}.advice')
    advice = cell(20, 1, 1).findall('w:p', NS)
    if len(advice) != 3:
        raise ValueError('overall advice must retain three existing paragraphs')
    for i, p in enumerate(advice, 1):
        whole(p, 'overall.advice.' + str(i))
    final_paths = paths(doc)
    for sdt in doc.findall('.//w:sdt', NS):
        tag = sdt.find('w:sdtPr/w:tag', NS).get(q('w:val'))
        manifest['fields'][tag][0]['candidate_xpath'] = final_paths[sdt]


def cache_points(cache, expected=None):
    if cache is None:
        raise ValueError('missing chart cache')
    n = int(cache.find('c:ptCount', NS).get('val'))
    points = cache.findall('c:pt', NS)
    if expected is not None and n != expected:
        raise ValueError('wrong chart point cardinality')
    if [int(x.get('idx')) for x in points] != list(range(n)):
        raise ValueError('incomplete or unordered chart cache')
    for point in points:
        if point.find('c:v', NS) is None or point.find('c:v', NS).text is None:
            raise ValueError('empty chart point')
    return points


def materialize_chart(root, manifest):
    parents = {child: parent for parent in root.iter() for child in parent}
    for node in list(root.iter()):
        if node.tag == q('c:numRef'):
            cache = node.find('c:numCache', NS); cache_points(cache)
            replacement = copy.deepcopy(cache); replacement.tag = q('c:numLit')
        elif node.tag == q('c:strRef'):
            cache = node.find('c:strCache', NS); points = cache_points(cache)
            if parents[node].tag == q('c:tx'):
                if len(points) != 1:
                    raise ValueError('series name must be scalar')
                replacement = ET.Element(q('c:v')); replacement.text = points[0].find('c:v', NS).text
            else:
                replacement = copy.deepcopy(cache); replacement.tag = q('c:strLit')
        elif node.tag == q('c:multiLvlStrRef'):
            cache = node.find('c:multiLvlStrCache', NS)
            if cache is None or cache.find('c:ptCount', NS).get('val') != '13':
                raise ValueError('incomplete multilevel classification')
            levels = [[{'idx': int(p.get('idx')), 'value': p.find('c:v', NS).text}
                       for p in level.findall('c:pt', NS)] for level in cache.findall('c:lvl', NS)]
            if len(levels) != 2 or [p['value'] for p in levels[0]] != [d[1] for d in DIMENSIONS]:
                raise ValueError('classification differs from 13-dimension contract')
            if [p['idx'] for p in levels[1]] != [0, 3, 7, 10]:
                raise ValueError('module category starts differ')
            if 'category_hierarchy' in manifest and manifest['category_hierarchy'] != levels:
                raise ValueError('series classification hierarchy mismatch')
            manifest['category_hierarchy'] = levels
            replacement = ET.Element(q('c:strLit'))
            ET.SubElement(replacement, q('c:ptCount'), {'val': '13'})
            for point in cache.findall('c:lvl', NS)[0].findall('c:pt', NS):
                replacement.append(copy.deepcopy(point))
        elif node.tag == q('c:externalData'):
            parents[node].remove(node); continue
        else:
            continue
        parent = parents[node]; index = list(parent).index(node)
        parent.remove(node); parent.insert(index, replacement)
    forbidden = {q('c:'+name) for name in ('f', 'externalData', 'numRef', 'strRef', 'multiLvlStrRef')}
    if any(node.tag in forbidden for node in root.iter()):
        raise ValueError('unmaterialized chart formula/reference')


def chart_style_fingerprint(root):
    clone = copy.deepcopy(root)
    if clone.find('.//c:barChart', NS) is not None and clone.find('.//c:lineChart', NS) is not None:
        legend = clone.find('c:chart/c:legend', NS)
        layout = legend.find('c:layout', NS)
        if layout is not None:
            legend.remove(layout)
    # Data sources and bound rich text are separate contracts, not styling roots.
    ring = clone.find('.//c:doughnutChart', NS) is not None
    for parent in clone.iter():
        for child in list(parent):
            if child.tag in {q('c:'+x) for x in ('val', 'cat', 'tx', 'externalData')} or (
                    ring and child.tag in {q('c:title'), q('c:dLbls'), q('c:autoTitleDeleted')}):
                parent.remove(child)
    return sha(ET.tostring(clone))


def detach_ring_label(chart, key):
    label = next(x for x in chart.findall('.//c:dLbl', NS)
                 if x.find('c:idx', NS).get('val') == '0')
    title = ET.Element(q('c:title'))
    title.append(copy.deepcopy(label.find('c:tx', NS)))
    layout = ET.SubElement(ET.SubElement(title, q('c:layout')), q('c:manualLayout'))
    for axis in ('x', 'y', 'w', 'h'):
        ET.SubElement(layout, q('c:'+axis+'Mode'), {'val':'edge' if axis in ('x','y') else 'factor'})
    bounds = SOURCE_LABEL_BOUNDS[key]
    for axis, value in zip(('x','y','w','h'), bounds[:4], strict=True):
        ET.SubElement(layout, q('c:'+axis), {'val':str(Decimal(value)/Decimal(bounds[4]))})
    ET.SubElement(title, q('c:overlay'), {'val':'1'})
    for tag in ('c:spPr', 'c:txPr'):
        title.append(copy.deepcopy(label.find(tag, NS)))
    cchart = chart.find('c:chart', NS)
    cchart.insert(0, title)
    deleted = cchart.find('c:autoTitleDeleted', NS)
    if deleted is not None:
        deleted.set('val', '0')
    series = chart.find('.//c:ser', NS)
    labels = series.find('c:dLbls', NS)
    index = list(series).index(labels)
    series.remove(labels)
    # An absent dLbls uses renderer defaults; explicitly suppress native labels.
    labels = ET.Element(q('c:dLbls'))
    ET.SubElement(labels, q('c:delete'), {'val':'1'})
    series.insert(index, labels)


def bind_rich_number(root, field, *, source_layout=False):
    rich = root.find('c:chart/c:title/c:tx/c:rich', NS)
    if rich is None:
        raise ValueError('ring numeric rich text not found')
    paragraphs = rich.findall('a:p', NS)
    if len(paragraphs) != 1:
        raise ValueError('ambiguous ring text paragraphs')
    p = paragraphs[0]; runs = p.findall('a:r', NS)
    text = ''.join(r.findtext('a:t', default='', namespaces=NS) for r in runs)
    matches = list(re.finditer(r'\d+\.\d+', text))
    if len(matches) != 1:
        raise ValueError('ring must contain exactly one numeric slot')
    match = matches[0]; offset = 0; numeric_node = None
    score = root.find('.//c:ser/c:val/c:numLit/c:pt/c:v', NS).text
    for run in runs:
        value = run.findtext('a:t', default='', namespaces=NS)
        lo, hi = max(match.start()-offset, 0), min(match.end()-offset, len(value))
        offset += len(value)
        if lo >= hi:
            continue
        index = list(p).index(run)
        if lo:
            p.insert(index, run_fragment(run, value[:lo], 'a')); index += 1
        if numeric_node is None:
            sample = match.group() if source_layout else format(
                Decimal(score).quantize(Decimal('.01'), rounding=ROUND_HALF_UP), 'f')
            fragment = run_fragment(run, sample, 'a')
            p.insert(index, fragment); index += 1
            numeric_node = fragment.find('a:t', NS)
        if hi < len(value):
            p.insert(index, run_fragment(run, value[hi:], 'a'))
        p.remove(run)
    parents = {child: parent for parent in root.iter() for child in parent}
    chain = []; node = numeric_node
    while node is not root:
        parent = parents[node]
        prefix = next(prefix for prefix, uri in NS.items() if node.tag.startswith('{'+uri+'}'))
        chain.append(prefix+':'+node.tag.split('}')[-1]+'['+str([x for x in parent if x.tag == node.tag].index(node)+1)+']')
        node = parent
    return {'field': field, 'xpath': './'+'/'.join(reversed(chain)),
            'kind': 'per-object-numeric-text-slot', 'format': 'HALF_UP_2',
            'fixed_label_and_unit_outside_slot': True}


def source_inline_origin(doc, paragraph, outer):
    body = doc.find('w:body', NS)
    children = list(body)
    if paragraph not in children or paragraph.find('w:pPr/w:jc', NS).get(q('w:val')) != 'center':
        raise ValueError('source inline must be in its verified centered body paragraph')
    section = next(section for child in children[children.index(paragraph):]
                   for section in ([child] if child.tag == q('w:sectPr') else child.findall('.//w:sectPr', NS)))
    page = section.find('w:pgSz', NS); margins = section.find('w:pgMar', NS)
    column = (int(page.get(q('w:w')))-int(margins.get(q('w:left')))-
              int(margins.get(q('w:right'))))*635
    return {'x':str((column-int(outer.find('wp:extent', NS).get('cx')))//2),
            'y':str(round(Fraction(SOURCE_GROUP_LINE_OFFSETS[outer.find('wp:docPr', NS).get('id')])*12700))}


def native_anchor(key, identifier, size, offset, origin):
    drawing = ET.Element(q('w:drawing'))
    anchor = ET.SubElement(drawing, q('wp:anchor'), {
        'distT':'0','distB':'0','distL':'0','distR':'0', 'simplePos':'0',
        'relativeHeight':str(identifier), 'behindDoc':'0', 'locked':'0',
        'layoutInCell':'1', 'allowOverlap':'1'})
    ET.SubElement(anchor, q('wp:simplePos'), {'x':'0','y':'0'})
    for axis, coord, relative in (('H','x','column'), ('V','y','paragraph')):
        pos = ET.SubElement(anchor, q('wp:position'+axis), {'relativeFrom':relative})
        ET.SubElement(pos, q('wp:posOffset')).text = str(int(origin[coord])+int(offset[coord]))
    ET.SubElement(anchor, q('wp:extent'), size)
    ET.SubElement(anchor, q('wp:effectExtent'), {'l':'0','t':'0','r':'0','b':'0'})
    ET.SubElement(anchor, q('wp:wrapNone'))
    ET.SubElement(anchor, q('wp:docPr'), {'id':str(identifier), 'name':key, 'title':key})
    ET.SubElement(anchor, q('wp:cNvGraphicFramePr'))
    return drawing, anchor


def promote_numeric_label(chart, key, slot, size, offset, origin, identifier, *, source_layout=False,
                          libreoffice_compatible=False):
    title = chart.find('c:chart/c:title', NS)
    rich = title.find('c:tx/c:rich', NS)
    numeric = chart.find(slot['xpath'], NS)
    bounds = SOURCE_LABEL_BOUNDS[key]
    label_size = {axis:str(round(Fraction(bounds[i+2])/Fraction(bounds[4])*int(size[axis])))
                  for i,axis in enumerate(('cx','cy'))}
    label_offset = {coord:str(int(offset[coord])+round(Fraction(bounds[i])/Fraction(bounds[4])*int(size[axis])))
                    for i,(coord,axis) in enumerate((('x','cx'),('y','cy')))}
    # Candidate-only bounds; original chart matrices and text styles stay fixed.
    if not source_layout or libreoffice_compatible:
        for coord, axis, points in (('x','cx', 120),
                                    ('y','cy', 64 if key == 'chart.overall' else 44)):
            extent = round(Fraction(points)/Fraction(bounds[4])*int(size[axis]))
            label_offset[coord] = str(int(label_offset[coord])-(extent-int(label_size[axis]))//2)
            label_size[axis] = str(extent)
    # Local offsets in physical points: modules outside left/right ring edges,
    # overall below its ring, in the gap above the unchanged comparison plot.
    dx, dy = {'chart.overall':(0,108), 'chart.module.self':(-96,0),
              'chart.module.interpersonal':(-96,0), 'chart.module.task':(100,0),
              'chart.module.development':(100,0)}[key]
    for coord, delta in (('x',dx),('y',dy)):
        if not source_layout or libreoffice_compatible:
            label_offset[coord] = str(int(label_offset[coord])+delta*12700)
    drawing, anchor = native_anchor(key+'.numeric-label', identifier, label_size, label_offset, origin)
    graphic = ET.SubElement(anchor, q('a:graphic'))
    data = ET.SubElement(graphic, q('a:graphicData'), {'uri':NS['wps']})
    shape = ET.SubElement(data, q('wps:wsp'))
    ET.SubElement(shape, q('wps:cNvSpPr'), {'txBox':'1'})
    props = ET.SubElement(shape, q('wps:spPr'))
    xfrm = ET.SubElement(props, q('a:xfrm'))
    ET.SubElement(xfrm, q('a:off'), {'x':'0','y':'0'})
    ET.SubElement(xfrm, q('a:ext'), label_size)
    ET.SubElement(ET.SubElement(props, q('a:prstGeom'), {'prst':'rect'}), q('a:avLst'))
    for child in title.find('c:spPr', NS):
        props.append(copy.deepcopy(child))
    content = ET.SubElement(ET.SubElement(shape, q('wps:txbx')), q('w:txbxContent'))
    numeric_text = None
    for source_p in rich.findall('a:p', NS):
        p = ET.SubElement(content, q('w:p'))
        pp = ET.SubElement(p, q('w:pPr'))
        ET.SubElement(pp, q('w:spacing'), {q('w:before'):'0',q('w:after'):'0',q('w:line'):'240',q('w:lineRule'):'auto'})
        ET.SubElement(pp, q('w:jc'), {q('w:val'):'center'})
        default = source_p.find('a:pPr/a:defRPr', NS)
        for source_run in source_p.findall('a:r', NS):
            direct = source_run.find('a:rPr', NS)
            run = ET.SubElement(p, q('w:r')); rp = ET.SubElement(run, q('w:rPr'))
            def style_attr(name, fallback=None):
                return direct.get(name, default.get(name, fallback))
            fonts = {}
            for a_font, w_fonts in [('latin',('ascii','hAnsi')),('ea',('eastAsia',)),('cs',('cs',))]:
                font = direct.find('a:'+a_font, NS)
                if font is None:
                    font = default.find('a:'+a_font, NS)
                if font is not None:
                    fonts.update({q('w:'+name):font.get('typeface') for name in w_fonts})
            ET.SubElement(rp, q('w:rFonts'), fonts)
            for prop in ('b','i'):
                value = style_attr(prop)
                if value is not None:
                    ET.SubElement(rp, q('w:'+prop), {q('w:val'):value})
            font_size = style_attr('sz')
            if font_size is not None:
                for prop in ('sz','szCs'):
                    ET.SubElement(rp, q('w:'+prop), {q('w:val'):str(int(font_size)//50)})
            fill = direct.find('a:solidFill', NS)
            if fill is None:
                fill = default.find('a:solidFill', NS)
            if fill is not None:
                color = fill.find('a:srgbClr', NS)
                if color is not None:
                    ET.SubElement(rp, q('w:color'), {q('w:val'):color.get('val')})
                else:
                    theme = fill.find('a:schemeClr', NS).get('val')
                    ET.SubElement(rp, q('w:color'), {q('w:themeColor'):{'tx1':'text1','tx2':'text2'}.get(theme,theme)})
            language = style_attr('lang')
            if language:
                ET.SubElement(rp, q('w:lang'), {q('w:val'):language})
            text = ET.SubElement(run, q('w:t'), {'{http://www.w3.org/XML/1998/namespace}space':'preserve'})
            source_text = source_run.find('a:t', NS)
            text.text = source_text.text
            if source_text is numeric:
                numeric_text = text
    body = copy.deepcopy(rich.find('a:bodyPr', NS)); body.tag = q('wps:bodyPr')
    if not source_layout or libreoffice_compatible:
        body.attrib.update({inset:'0' for inset in ('lIns','rIns','tIns','bIns')})
    shape.append(body)
    chart.find('c:chart', NS).remove(title)
    deleted = chart.find('c:chart/c:autoTitleDeleted', NS)
    if deleted is not None:
        deleted.set('val', '1')
    if numeric_text is None:
        raise ValueError('numeric text lost during source-style label conversion')
    slot.update({'part':'word/document.xml', 'kind':'independent-source-label-numeric-slot',
                 'source_bounds_points':list(bounds), 'extent_emu':label_size})
    return drawing, numeric_text


def promote_fixed_background(group, parents, origin, run, insertion, next_id):
    for shape in group.findall('.//wps:wsp', NS):
        transform = shape.find('wps:spPr/a:xfrm', NS)
        offset = {axis:Fraction(transform.find('a:off', NS).get(axis)) for axis in ('x','y')}
        size = {axis:Fraction(transform.find('a:ext', NS).get(axis)) for axis in ('cx','cy')}
        node = parents[shape]
        while True:
            transform = node.find('wpg:grpSpPr/a:xfrm', NS)
            if transform.get('rot', '0') != '0':
                raise ValueError('unverified fixed background rotation')
            for coord, axis in (('x','cx'), ('y','cy')):
                scale = Fraction(int(transform.find('a:ext', NS).get(axis)),
                                 int(transform.find('a:chExt', NS).get(axis)))
                offset[coord] = int(transform.find('a:off', NS).get(coord)) + (
                    offset[coord]-int(transform.find('a:chOff', NS).get(coord)))*scale
                size[axis] *= scale
            if node is group:
                break
            node = parents[node]
        offset = {k:str(round(v)) for k,v in offset.items()}
        size = {k:str(round(v)) for k,v in size.items()}
        key = 'fixed.grade.'+shape.find('wps:cNvPr', NS).get('id')
        drawing, anchor = native_anchor(key, next_id, size, offset, origin)
        next_id += 1
        graphic = ET.SubElement(anchor, q('a:graphic'))
        data = ET.SubElement(graphic, q('a:graphicData'), {'uri':NS['wps']})
        clone = copy.deepcopy(shape)
        props = clone.find('wps:spPr/a:xfrm', NS)
        props.find('a:off', NS).attrib.update({'x':'0','y':'0'})
        props.find('a:ext', NS).attrib.update(size)
        data.append(clone)
        run.insert(insertion, drawing); insertion += 1
    return insertion, next_id


def promote_charts(doc, parts, manifest, *, source_layout=False, libreoffice_compatible=False):
    rels = ET.fromstring(parts['word/_rels/document.xml.rels'])
    targets = {r.get('Id'): posixpath.normpath('word/'+r.get('Target')) for r in rels}
    parents = {child: parent for parent in doc.iter() for child in parent}
    groups = doc.findall('.//wpg:wgp', NS)
    next_id = max(int(x.get('id')) for x in doc.findall('.//wp:docPr', NS)) + 1
    for group in groups:
        frames = group.findall('wpg:graphicFrame', NS)
        if not frames:
            continue
        drawing = parents[parents[parents[parents[group]]]]
        if drawing.tag != q('w:drawing'):
            raise ValueError('unexpected group drawing ancestry')
        alt = parents[parents[drawing]]
        if alt.tag != q('mc:AlternateContent'):
            raise ValueError('expected original grouped alternate content')
        run = parents[alt]
        paragraph = parents[run]
        outer = parents[parents[parents[group]]]
        if outer.tag not in (q('wp:inline'), q('wp:anchor')):
            raise ValueError('unexpected original drawing placement')
        insertion = list(run).index(alt)
        origin = source_inline_origin(doc, paragraph, outer) if outer.tag == q('wp:inline') else None
        insertion, next_id = promote_fixed_background(group, parents, origin, run, insertion, next_id)
        group_transform = group.find('wpg:grpSpPr/a:xfrm', NS)
        group_extent = group_transform.find('a:ext', NS)
        group_child_extent = group_transform.find('a:chExt', NS)
        for frame in frames:
            refs = frame.findall('.//c:chart', NS)
            if len(refs) != 1:
                raise ValueError('unexpected non-chart graphicFrame')
            rid = refs[0].get(q('r:id')); part = targets[rid]
            chart = ET.fromstring(parts[part])
            if chart.find('.//c:doughnutChart', NS) is not None:
                rich = chart.find('.//c:dLbl/c:tx/c:rich', NS)
                title = ''.join(t.text or '' for t in rich.findall('.//a:t', NS))
                matches = [key for label, key in LABEL_KEYS.items() if title.startswith(label)]
                if len(matches) != 1:
                    raise ValueError('ambiguous ring business identity')
                key = matches[0]
            elif chart.find('.//c:barChart', NS) is not None and chart.find('.//c:lineChart', NS) is not None:
                key = 'chart.dimension.comparison'
            else:
                raise ValueError('unsupported chart type')
            materialize_chart(chart, manifest)
            series = chart.findall('.//c:ser', NS)
            expected = 13 if key.endswith('comparison') else 2
            if len(series) != (2 if expected == 13 else 1):
                raise ValueError('chart series count mismatch')
            for ser in series:
                cache_points(ser.find('c:val/c:numLit', NS), expected)
            if expected == 13 and (not source_layout or libreoffice_compatible):
                layout = chart.find('c:chart/c:legend/c:layout/c:manualLayout', NS)
                for axis, value in {'x':'0.66', 'y':'0.88', 'w':'0.33', 'h':'0.09'}.items():
                    layout.find('c:'+axis, NS).set('val', value)
            field = 'overall.score' if key == 'chart.overall' else key.removeprefix('chart.')+'.score'
            if expected == 2:
                detach_ring_label(chart, key)
            slots = [] if expected == 13 else [bind_rich_number(chart, field, source_layout=source_layout)]
            transform = frame.find('wpg:xfrm', NS)
            extent = transform.find('a:ext', NS)
            size = {}
            offset = {}
            for axis in ('cx', 'cy'):
                numerator = int(extent.get(axis)) * int(group_extent.get(axis))
                denominator = int(group_child_extent.get(axis))
                size[axis] = str(round(Fraction(numerator, denominator)))
                coordinate = 'x' if axis == 'cx' else 'y'
                offset[coordinate] = str(round(Fraction(
                    (int(transform.find('a:off', NS).get(coordinate))-
                     int(group_transform.find('a:chOff', NS).get(coordinate))) *
                    int(group_extent.get(axis)), denominator)))
            if outer.tag == q('wp:anchor'):
                new_drawing = ET.Element(q('w:drawing'))
                anchor = copy.deepcopy(outer)
                for child in list(anchor):
                    if child.tag == q('a:graphic'):
                        anchor.remove(child)
                for axis, coord in (('H','x'), ('V','y')):
                    position = anchor.find('wp:position'+axis, NS)
                    pos = position.find('wp:posOffset', NS)
                    if pos is None:
                        raise ValueError('aligned outer anchor requires explicit source positioning')
                    pos.text = str(int(pos.text)+int(offset[coord]))
                anchor.find('wp:extent', NS).attrib.update(size)
                anchor.find('wp:docPr', NS).attrib.update({'id':str(next_id), 'name':key, 'title':key})
                new_drawing.append(anchor)
            else:
                new_drawing, anchor = native_anchor(key, next_id, size, offset, origin)
            next_id += 1
            anchor.append(copy.deepcopy(frame.find('a:graphic', NS)))
            run.insert(insertion, new_drawing); insertion += 1
            if slots:
                label_drawing, numeric_text = promote_numeric_label(
                    chart, key, slots[0], size, offset, origin, next_id, source_layout=source_layout,
                    libreoffice_compatible=libreoffice_compatible)
                next_id += 1
                run.insert(insertion, label_drawing); insertion += 1
                node_path = paths(doc)[numeric_text]
                slots[0]['xpath'] = '.'+node_path.removeprefix('/w:document[1]')
            group.remove(frame)
            parts[part] = serialize(chart, parts[part])
            manifest['charts'].append({'key':key, 'part':part, 'relationship_id':rid,
                                       'type':'bar+line' if expected == 13 else 'doughnut',
                                       'series':len(series), 'points_per_series':expected,
                                       'numeric_text_slots':slots,
                                       'source_frame_transform':{x.tag.split('}')[-1]:dict(x.attrib) for x in transform},
                                       'source_group_transform':{x.tag.split('}')[-1]:dict(x.attrib) for x in group_transform},
                                       'anchor_extent_emu':size, 'anchor_offset_emu':offset,
                                       'source_inline_origin_emu':origin,
                                       'source_outer_placement':outer.tag.split('}')[-1],
                                       'source_label_bounds_points':SOURCE_LABEL_BOUNDS.get(key),
                                       'layout_status':'pending_visual_approval'})
        # Keep all fixed children of the original group, including grade background.
        # Obsolete fallback images are not allowed to display stale chart screenshots.
        run.remove(alt)
        if outer.tag == q('wp:inline'):
            # Preserve the original inline flow box even when all group children
            # became anchors. No new grid, paragraph or extra chart-sized line.
            graphic_data = outer.find('a:graphic/a:graphicData', NS)
            graphic_data.clear()
            graphic_data.set('uri', NS['wps'])
            shape = ET.SubElement(graphic_data, q('wps:wsp'))
            ET.SubElement(shape, q('wps:cNvSpPr'))
            props = ET.SubElement(shape, q('wps:spPr'))
            xfrm = ET.SubElement(props, q('a:xfrm'))
            ET.SubElement(xfrm, q('a:off'), {'x':'0','y':'0'})
            ET.SubElement(xfrm, q('a:ext'), dict(outer.find('wp:extent', NS).attrib))
            ET.SubElement(ET.SubElement(props, q('a:prstGeom'), {'prst':'rect'}), q('a:avLst'))
            ET.SubElement(props, q('a:noFill'))
            ET.SubElement(ET.SubElement(props, q('a:ln')), q('a:noFill'))
            ET.SubElement(shape, q('wps:bodyPr'))
            run.insert(insertion, drawing)
        else:
            raise ValueError('unverified grouped outer anchor flow; refuse layout guess')
    if len(manifest['charts']) != 6 or len({x['key'] for x in manifest['charts']}) != 6:
        raise ValueError('six distinct business charts required')


def build_candidate(source=SOURCE, expected_sha=SOURCE_SHA, *, source_layout=False,
                    libreoffice_compatible=False):
    if source_layout and libreoffice_compatible:
        raise ValueError('source-layout and libreoffice-compatible modes are mutually exclusive')
    raw = Path(source).read_bytes()
    if sha(raw) != expected_sha or expected_sha != SOURCE_SHA:
        raise ValueError('source SHA mismatch; do not build from a changed customer source')
    parts = read_package(raw)
    manifest = {'schema':'management-traits-002-word-candidate-v1', 'status':'candidate_not_approved',
                'source_sha256':sha(raw), 'fields':{}, 'charts':[], 'repeat_whitelist':[],
                'overall_advice_protocol':{'logical_key':'overall.advice', 'cardinality':3,
                                           'slots':['overall.advice.1','overall.advice.2','overall.advice.3'],
                                           'layout':'three-existing-numbered-paragraphs', 'approval':'pending'},
                'pending':[
                    {'field':'participant.post','reason':'existing 职务 slot; post semantics require approval'},
                    {'field':'result.submittedAt','reason':'existing 测评日期 slot; submission-time semantics require approval'},
                    {'field':'participant.age','reason':'no approved position; not inserted'},
                    {'field':'result.userTime','reason':'no approved position; not inserted'},
                    {'field':'overall.norm','reason':'no approved position; not inserted'},
                    {'field':'chart.dimension.comparison.categories','reason':'OOXML has no multiLvlStrLit; 13 leaf literals; original two-level values retained here; hierarchical display pending'},
                    {'field':'chart.layout','reason':'native anchors use source frame/group transforms; original inline flow boxes and fixed layers retained; Word/LO placement requires visual approval'},
                    {'field':'overall.level.decorations','reason':'original fixed sample stars retained; conditional star meaning pending'},
                    {'field':'footer','reason':'original PAGE/NUMPAGES and fixed total retained; policy not approved'},
                    {'field':'fixed.content','reason':'definitions/statistical wording/product name kept; not content approval'},
                ]}
    doc = ET.fromstring(parts['word/document.xml'])
    if not source_layout:
        local_page_flow(doc)
    if libreoffice_compatible:
        summary = doc.findall('.//w:tbl', NS)[3].findall('w:tr/w:tc', NS)
        for cell in summary:
            paragraphs = cell.findall('w:p', NS)
            for index, paragraph in enumerate(paragraphs):
                props = paragraph.find('w:pPr', NS)
                for flag in ('keepLines', 'keepNext'):
                    if flag == 'keepNext' and index == len(paragraphs)-1:
                        continue
                    node = props.find('w:'+flag, NS)
                    if node is None:
                        node = ET.SubElement(props, q('w:'+flag))
                    node.set(q('w:val'), '1')
        for paragraph in doc.findall('w:body/w:p', NS):
            if not wtext(paragraph).startswith('【诊断】'):
                continue
            props = paragraph.find('w:pPr', NS)
            node = props.find('w:keepLines', NS)
            if node is None:
                node = ET.SubElement(props, q('w:keepLines'))
            node.set(q('w:val'), '1')
    preserve_source_text = source_layout or libreoffice_compatible
    bind_document(doc, manifest, source_layout=preserve_source_text)
    promote_charts(doc, parts, manifest, source_layout=preserve_source_text,
                   libreoffice_compatible=libreoffice_compatible)
    if preserve_source_text:
        manifest.update({'layout_mode':'source-layout', 'source_path':str(Path(source).resolve()),
                         'source_bytes':len(raw), 'runtime_connected':False,
                         'attachment_byte_equality_verified':False,
                         'sample_values_preserved_not_scoring_gold_standard':True,
                         'output_paths':{
                             'template':str((OUT/'management-traits-002-source-layout-template.docx').resolve()),
                             'manifest':str((OUT/'source-layout-manifest.json').resolve()),
                             'demo':str((OUT/'source-layout-synthetic-demo.docx').resolve())},
                         'layout_policy':'customer source layout first; no label shifts, expanded bounds, inset, legend or pagination optimization; fixed stars/footer retained',
                         'sample_discrepancies':'source overall label 59.45 vs cached score retained without correction; sample is not scoring evidence',
                         'allowed_changes':['88 SDT wrappers/text-run splits preserving source rPr',
                             'six grouped charts/five backgrounds to source-coordinate anchors with original inline flow boxes',
                             'five labels to source-bound/source-font/source-inset Word text boxes with preserved sample numeric slots',
                             'cached data references to literals; accepted 13 leaf category flattening',
                             'six external chart OLE relationships/externalData removed'],
                         'semantic_decisions':{'participant.post':'post',
                             'result.submittedAt':'submittedAt', 'categories':'13 leaf literals accepted'},
                         'pending':[entry for entry in manifest['pending'] if entry['field'] not in (
                             'participant.post','result.submittedAt','chart.dimension.comparison.categories',
                             'overall.level.decorations','footer')]+[
                             {'field':'source.layout','reason':'original bounds/stars/footer retained by current customer priority; local PDF visual acceptance and target server unverified'}]})
    if libreoffice_compatible:
        manifest.update({'layout_mode':'libreoffice-compatible',
                         'output_paths':{
                             'template':str((OUT/'management-traits-002-lo-compatible-template.docx').resolve()),
                             'manifest':str((OUT/'lo-compatible-manifest.json').resolve()),
                             'demo':str((OUT/'lo-compatible-synthetic-demo.docx').resolve())},
                         'layout_policy':'source text/rPr/sample/stars/footer preserved; only numeric-box size/insets/offsets, comparison legend manualLayout and local page-flow compatibility',
                         'allowed_changes':manifest['allowed_changes']+[
                             'five numeric boxes: 120pt width, 64/44pt height, zero insets; left -96pt/right +100pt/overall +108pt offsets only',
                             'comparison legend manualLayout x/y/w/h only',
                             'local summary/detail/advice keepLines/keepNext/cantSplit and detail heading pageBreakBefore only'],
                         'pending':[entry for entry in manifest['pending'] if entry['field']!='source.layout']+[
                             {'field':'compatible.layout','reason':'local PDF contracts require execution; final visual review, Word/PDF parity and target server acceptance not approved'}]})
    for name, data in list(parts.items()):
        if not name.endswith('.rels'):
            continue
        rels = ET.fromstring(data)
        external = [r for r in rels if r.get('TargetMode') == 'External']
        if external:
            if not name.startswith('word/charts/_rels/') or any(not r.get('Type', '').endswith('/oleObject') for r in external):
                raise ValueError('unknown external relationship; refuse candidate')
            for relationship in external:
                rels.remove(relationship)
            parts[name] = serialize(rels, data)
    parts['word/document.xml'] = serialize(doc, parts['word/document.xml'])
    data = pack(parts)
    manifest['candidate_sha256'] = sha(data)
    manifest['media_sha256'] = {name:sha(data) for name, data in parts.items() if name.startswith('word/media/')}
    manifest['changed_parts'] = sorted(name for name in parts if parts[name] != read_package(raw)[name])
    return data, manifest


def workbook_cells():
    raw = WORKBOOK.read_bytes()
    if sha(raw) != WORKBOOK_SHA:
        raise ValueError('customer workbook SHA mismatch')
    parts = read_package(raw)
    strings = [''.join(t.text or '' for t in si.findall('.//s:t', NS))
               for si in ET.fromstring(parts['xl/sharedStrings.xml'])]
    sheets = {}
    for i in range(1, 5):
        cells = {}
        for cell in ET.fromstring(parts[f'xl/worksheets/sheet{i}.xml']).findall('.//s:c', NS):
            value = cell.findtext('s:v', default='', namespaces=NS)
            if cell.get('t') == 's':
                value = strings[int(value)]
            cells[cell.get('r')] = value
        sheets[i] = cells
    return sheets


def number(value, digits=2):
    value = Fraction(value)
    return format((Decimal(value.numerator)/Decimal(value.denominator)).quantize(
        Decimal(1).scaleb(-digits), rounding=ROUND_HALF_UP), 'f')


def make_demo(candidate, manifest):
    if sha(candidate) != manifest['candidate_sha256']:
        raise ValueError('candidate/manifest SHA mismatch')
    sheets = workbook_cells()
    values = {}; origins = {}; scores = []
    for i, (key, name, module, count, norm) in enumerate(DIMENSIONS):
        row = i + 3
        if sheets[2]['B'+str(row)].strip() != name:
            raise ValueError('workbook dimension order/name mismatch')
        # Longest raw customer advice per dimension is stress data, not approval.
        col = max('CDEFG', key=lambda c: len(sheets[3][c+str(row)]))
        level = 'CDEFG'.index(col)
        raw = (5, 4, 3, 2, 1)[level]
        # Synthetic integer final sums produce reachable percentage values.
        score = Fraction(25 * (raw*count-count), count); scores.append(score)
        base = 'dimension.'+key+'.'
        values.update({base+'score':number(score), base+'level':('优秀','良好','合格','欠佳','不足')[level],
                       base+'norm':norm, base+'diagnosis':sheets[2]['DEFGH'[level]+str(row)],
                       base+'advice':sheets[3][col+str(row)]})
        origins[base+'advice'] = {'sheet':3, 'cell':col+str(row), 'selection':'longest raw advice; conflict text unchanged'}
        origins[base+'diagnosis'] = {'sheet':2, 'cell':'DEFGH'[level]+str(row)}
    overall = sum(scores, Fraction()) / 13
    overall_index = next((i for i, boundary in enumerate((90,70,30,10)) if overall >= boundary), 4)
    row = overall_index+2
    values['overall.score'] = number(overall)
    values['overall.level'] = sheets[4]['B'+str(row)]
    values['overall.diagnosis'] = sheets[4]['C'+str(row)]
    advice = sheets[4]['D'+str(row)].split('\n')
    if len(advice) != 3 or not all(advice):
        raise ValueError('customer overall advice is not exactly three paragraphs')
    for i, text in enumerate(advice, 1):
        values['overall.advice.'+str(i)] = text
        origins['overall.advice.'+str(i)] = {'sheet':4,'cell':'D'+str(row),'paragraph':i}
    values.update({'participant.name':'synthetic 演示人员（非正式数据）', 'participant.gender':'—',
                   'participant.affiliation':'synthetic 本地候选审阅', 'participant.post':'synthetic 职务语义待批准',
                   'participant.telephone':'synthetic 非真实联系方式', 'result.submittedAt':'2026-10-01 00:00（synthetic）'})
    for group, ordered in [('high', sorted(range(13), key=lambda i:(-scores[i], i))[:3]),
                           ('low', sorted(range(13), key=lambda i:(scores[i], i))[:3])]:
        for slot, index in enumerate(ordered, 1):
            key, name, *_ = DIMENSIONS[index]
            level = ('优秀','良好','合格','欠佳','不足').index(values['dimension.'+key+'.level'])
            values[f'overview.{group}.{slot}.name'] = name
            values[f'overview.{group}.{slot}.text'] = sheets[1]['CDEFG'[level]+str(index+3)]
    chart_values = {'chart.overall':[[number(overall,12), number(100-overall,12)]]}
    for module in ('self','interpersonal','task','development'):
        members = [scores[i] for i, d in enumerate(DIMENSIONS) if d[2] == module]
        score = sum(members, Fraction())/len(members)
        values['module.'+module+'.score'] = number(score)
        chart_values['chart.module.'+module] = [[number(score,12),number(100-score,12)]]
    chart_values['chart.dimension.comparison'] = [[number(s,12) for s in scores],
                                                 [number(d[4],12) for d in DIMENSIONS]]
    parts = read_package(candidate); doc = ET.fromstring(parts['word/document.xml'])
    for sdt in doc.findall('.//w:sdt', NS):
        tag = sdt.find('w:sdtPr/w:tag', NS).get(q('w:val'))
        if tag not in values:
            raise ValueError('missing demo field: '+tag)
        nodes = sdt.findall('.//w:t', NS)
        nodes[0].text = values[tag]
        for node in nodes[1:]:
            node.text = ''
    for chart in manifest['charts']:
        part = chart['part']; root = ET.fromstring(parts[part])
        for ser, data in zip(root.findall('.//c:ser', NS), chart_values[chart['key']], strict=True):
            points = cache_points(ser.find('c:val/c:numLit', NS), len(data))
            for point, value in zip(points, data, strict=True):
                point.find('c:v', NS).text = value
        for slot in chart['numeric_text_slots']:
            doc.find(slot['xpath'], NS).text = values[slot['field']]
        parts[part] = serialize(root, parts[part])
    parts['word/document.xml'] = serialize(doc, parts['word/document.xml'])
    demo = pack(parts)
    evidence = {'synthetic':True, 'not_formal_report':True, 'no_scoring_gold_standard':True,
                'workbook_sha256':WORKBOOK_SHA, 'values':values, 'chart_values':chart_values,
                'content_origins':origins, 'demo_sha256':sha(demo),
                'synthetic_scores':'reachable final-score sums selected for longest advice stress; not actual answers'}
    return demo, evidence


def write_products(source, candidate_path, manifest_path, demo_path, *, source_layout=False,
                   libreoffice_compatible=False):
    source = Path(source).resolve()
    outputs = [Path(p).resolve() for p in (candidate_path, manifest_path, demo_path)]
    if source in outputs or any(p.exists() and p.samefile(source) for p in outputs) or len(set(outputs)) != 3:
        raise ValueError('source=target or duplicate output paths refused')
    if any(not p.is_relative_to(OUT.resolve()) for p in outputs):
        raise ValueError('outputs restricted to approved local candidate directory')
    if source_layout and [p.name for p in outputs] != [
            'management-traits-002-source-layout-template.docx',
            'source-layout-manifest.json','source-layout-synthetic-demo.docx']:
        raise ValueError('source-layout outputs must use new approved names; historical products protected')
    if libreoffice_compatible and [p.name for p in outputs] != [
            'management-traits-002-lo-compatible-template.docx',
            'lo-compatible-manifest.json','lo-compatible-synthetic-demo.docx']:
        raise ValueError('libreoffice-compatible outputs must use independent approved names')
    data, manifest = build_candidate(source, source_layout=source_layout,
                                     libreoffice_compatible=libreoffice_compatible)
    if source_layout or libreoffice_compatible:
        manifest['output_paths'] = dict(zip(('template','manifest','demo'),map(str,outputs),strict=True))
    demo, evidence = make_demo(data, manifest)
    manifest['synthetic_demo'] = evidence
    raw_manifest = (json.dumps(manifest, ensure_ascii=False, indent=2)+'\n').encode('utf-8')
    for path, content in zip(outputs, (data, raw_manifest, demo), strict=True):
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(content)
    if sha(source.read_bytes()) != SOURCE_SHA:
        raise ValueError('source changed during build')
    return manifest


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--source', type=Path, default=SOURCE)
    modes = parser.add_mutually_exclusive_group()
    modes.add_argument('--source-layout', action='store_true', help='derive using original customer layout; preserve sample values/stars/footer')
    modes.add_argument('--libreoffice-compatible', action='store_true',
                       help='preserve source text/styles/samples/stars/footer with bounded local LibreOffice layout fixes')
    parser.add_argument('--output', type=Path)
    parser.add_argument('--manifest', type=Path)
    parser.add_argument('--demo', type=Path)
    args = parser.parse_args()
    names = ('management-traits-002-lo-compatible-template.docx','lo-compatible-manifest.json',
             'lo-compatible-synthetic-demo.docx') if args.libreoffice_compatible else (
             'management-traits-002-source-layout-template.docx','source-layout-manifest.json',
             'source-layout-synthetic-demo.docx') if args.source_layout else (
             'management-traits-002-candidate.docx','field-position-manifest.json','management-traits-002-synthetic-demo.docx')
    manifest = write_products(args.source, args.output or OUT/names[0], args.manifest or OUT/names[1],
                              args.demo or OUT/names[2], source_layout=args.source_layout,
                              libreoffice_compatible=args.libreoffice_compatible)
    print('SOURCE_SHA256='+manifest['source_sha256'])
    print('CANDIDATE_SHA256='+manifest['candidate_sha256'])
    print('SYNTHETIC_DEMO_SHA256='+manifest['synthetic_demo']['demo_sha256'])
    print('SDT_FIELDS='+str(len(manifest['fields']))+' CHARTS=6 STATUS=candidate_not_approved')
    return 0


if __name__ == '__main__':
    sys.exit(main())