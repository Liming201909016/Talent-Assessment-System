#!/usr/bin/env python3
import json
import re
import urllib.request
import zipfile
from pathlib import Path
from xml.etree import ElementTree as ET

BASE = 'http://127.0.0.1:8092'
STATE = Path('/tmp/competency-results-verify-state.json')
OUT = Path('/tmp/competency-export-verify')
NS = {'m': 'http://schemas.openxmlformats.org/spreadsheetml/2006/main', 'r': 'http://schemas.openxmlformats.org/officeDocument/2006/relationships'}
REL_NS = {'p': 'http://schemas.openxmlformats.org/package/2006/relationships'}


def download(path, token, target):
    request = urllib.request.Request(BASE + path, headers={'Authorization': 'Bearer ' + token}, method='GET')
    with urllib.request.urlopen(request, timeout=60) as response:
        data = response.read()
        content_type = response.headers.get('Content-Type', '')
        disposition = response.headers.get('Content-Disposition', '')
    if 'spreadsheetml.sheet' not in content_type or len(data) < 1000:
        raise RuntimeError(f'invalid download: type={content_type!r} bytes={len(data)}')
    if 'filename*=UTF-8' not in disposition or 'xlsx' not in disposition:
        raise RuntimeError(f'invalid content disposition: {disposition!r}')
    target.write_bytes(data)
    return disposition, len(data)


def cell_column(reference):
    letters = re.match(r'[A-Z]+', reference).group(0)
    result = 0
    for letter in letters:
        result = result * 26 + ord(letter) - 64
    return result


def workbook_rows(path):
    with zipfile.ZipFile(path) as archive:
        shared = []
        if 'xl/sharedStrings.xml' in archive.namelist():
            root = ET.fromstring(archive.read('xl/sharedStrings.xml'))
            for item in root.findall('m:si', NS):
                shared.append(''.join(node.text or '' for node in item.findall('.//m:t', NS)))
        workbook = ET.fromstring(archive.read('xl/workbook.xml'))
        relationships = ET.fromstring(archive.read('xl/_rels/workbook.xml.rels'))
        targets = {item.attrib['Id']: item.attrib['Target'] for item in relationships.findall('p:Relationship', REL_NS)}
        output = {}
        for sheet in workbook.findall('m:sheets/m:sheet', NS):
            name = sheet.attrib['name']
            target = targets[sheet.attrib['{%s}id' % NS['r']]].lstrip('/')
            if not target.startswith('xl/'):
                target = 'xl/' + target
            xml = ET.fromstring(archive.read(target))
            rows = []
            for row in xml.findall('m:sheetData/m:row', NS):
                values = []
                for cell in row.findall('m:c', NS):
                    column = cell_column(cell.attrib['r'])
                    while len(values) < column:
                        values.append('')
                    kind = cell.attrib.get('t')
                    value_node = cell.find('m:v', NS)
                    value = '' if value_node is None else value_node.text or ''
                    if kind == 's' and value:
                        value = shared[int(value)]
                    elif kind == 'inlineStr':
                        value = ''.join(node.text or '' for node in cell.findall('.//m:t', NS))
                    values[column - 1] = value
                rows.append(values)
            output[name] = rows
        return output


def verify(rows):
    if list(rows) != ['结果汇总', '逐题明细', '题目字典']:
        raise RuntimeError(f'sheets mismatch: {list(rows)}')
    summary = rows['结果汇总']
    detail = rows['逐题明细']
    dictionary = rows['题目字典']
    if len(summary) != 4 or len(detail) != 49 or len(dictionary) != 17:
        raise RuntimeError(f'row counts summary/detail/dict={len(summary)}/{len(detail)}/{len(dictionary)}')
    headers = summary[0]
    if headers[17:19] != ['D01 沟通表达', 'D02 人际交往']:
        raise RuntimeError(f'dynamic headers mismatch: {headers[17:19]}')
    names = [row[2] for row in summary[1:]]
    if names != ['Result Low', 'Result Mixed', 'Result High']:
        raise RuntimeError(f'summary names mismatch: {names}')
    overall = [float(row[12]) for row in summary[1:]]
    d01 = [float(row[17]) for row in summary[1:]]
    d02 = [float(row[18]) for row in summary[1:]]
    if overall != [2.0, 6.0, 8.0] or d01 != [1.0, 4.0, 5.0] or d02 != [1.0, 2.0, 3.0]:
        raise RuntimeError(f'scores mismatch overall={overall} d01={d01} d02={d02}')
    if any(row[9] != '100.00%' or row[10] != '完整' for row in summary[1:]):
        raise RuntimeError('completion values mismatch')
    if any(not row[11] or not row[12] or not row[13] for row in detail[1:]):
        raise RuntimeError('raw answer/text/final score missing')
    if any(row[14] != '已作答' for row in detail[1:]):
        raise RuntimeError('answered status mismatch')
    if [int(row[0]) for row in dictionary[1:]] != list(range(1, 17)):
        raise RuntimeError('dictionary snapshot order mismatch')
    if len({row[1] for row in dictionary[1:]}) != 16:
        raise RuntimeError('dictionary question codes not unique')
    return {'overall': overall, 'd01': d01, 'd02': d02, 'detailRows': 48, 'dictionaryRows': 16}


def main():
    state = json.loads(STATE.read_text(encoding='utf-8'))
    OUT.mkdir(exist_ok=True)
    paths = {
        'summary': f'/exam/api/exam/exam/export-raw-data?examId={state["examId"]}',
        'answers': f'/exam/api/exam/exam/export-raw-answers?examId={state["examId"]}',
    }
    normalized = {}
    for name, path in paths.items():
        target = OUT / f'{name}.xlsx'
        disposition, size = download(path, state['adminToken'], target)
        rows = workbook_rows(target)
        evidence = verify(rows)
        normalized[name] = rows
        print(f'{name}_download_bytes={size}')
        print(f'{name}_content_disposition={disposition}')
        print(f'{name}_verified={json.dumps(evidence, ensure_ascii=False, separators=(",", ":"))}')
    if normalized['summary'] != normalized['answers']:
        raise RuntimeError('two existing export endpoints returned different workbook content')
    print('COMPETENCY_EXPORT_VERIFY_OK')
    print('two_endpoints_same_content=true')


if __name__ == '__main__':
    main()
