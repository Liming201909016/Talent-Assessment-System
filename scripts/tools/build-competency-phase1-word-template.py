#!/usr/bin/env python3
"""Build the customer-maintainable phase-one DOCX template with stable placeholders."""

import argparse
import html
import re
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DEFAULT_SOURCE = ROOT / 'docs' / '260807胜任力开发资料' / '胜任力测评报告样例.docx'
DEFAULT_OUTPUT = ROOT / 'Go-based Refactored System' / 'configs' / 'export-templates' / 'competency-phase1-report.docx'

DIMENSIONS = [
    ('逻辑思维', 'competency-a1-01'), ('数字应用', 'competency-a1-02'),
    ('计划执行', 'competency-a1-03'), ('持续学习', 'competency-a1-04'),
    ('沟通表达', 'competency-a1-05'), ('敬业奉献', 'competency-b1-01'),
    ('求真务实', 'competency-b1-02'), ('自律性', 'competency-b1-03'),
    ('成就导向', 'competency-b1-04'), ('合作意识', 'competency-b1-05'),
]

PARAGRAPH_RE = re.compile(r'<w:p(?:\s[^>]*)?>.*?</w:p>', re.S)
TEXT_RE = re.compile(r'(<w:t(?:\s[^>]*)?>)(.*?)(</w:t>)', re.S)


def paragraph_text(paragraph):
    return ''.join(html.unescape(match.group(2)) for match in TEXT_RE.finditer(paragraph))


def set_runs(paragraph, values):
    index = 0

    def replace(match):
        nonlocal index
        value = values[index] if index < len(values) else ''
        index += 1
        return match.group(1) + html.escape(value, quote=False) + match.group(3)

    result = TEXT_RE.sub(replace, paragraph)
    if index < len(values):
        raise RuntimeError(f'paragraph has {index} text runs, needs {len(values)}')
    return result


def set_whole(paragraph, value):
    return set_runs(paragraph, [value])


def transform_document(document):
    dimension_by_name = dict(DIMENSIONS)
    title_counts = {name: 0 for name, _ in DIMENSIONS}
    score_index = 0
    level_index = 0
    diagnosis_index = -1  # overall diagnosis is index -1, dimensions start at 0
    group_score_index = 0
    group_level_index = 0
    group_description_index = 0
    replacements = 0

    def transform(match):
        nonlocal score_index, level_index, diagnosis_index
        nonlocal group_score_index, group_level_index, group_description_index, replacements
        paragraph = match.group(0)
        text = paragraph_text(paragraph)
        stripped = text.strip()

        if stripped in dimension_by_name:
            title_counts[stripped] += 1
            return paragraph

        if stripped == '2026年8月5日':
            replacements += 1
            return set_whole(paragraph, '{{report.date}}')
        if stripped == '姓名：年龄：':
            replacements += 2
            return set_runs(paragraph, ['姓名：{{participant.name}}', '年龄：{{participant.age}}'])
        if stripped == '性别：手机号：':
            replacements += 2
            return set_runs(paragraph, ['性别：{{participant.gender}}', '手机号：{{participant.telephone}}'])
        if stripped == '单位：':
            replacements += 1
            return set_whole(paragraph, '单位：{{participant.affiliation}}')
        if stripped == '岗位：':
            replacements += 1
            return set_whole(paragraph, '岗位：{{participant.post}}')
        if stripped == '时间：2026年8月5日时长：20分钟':
            replacements += 2
            return set_runs(paragraph, ['时间：{{result.submittedAt}}', '', '', '', '', '时长：{{result.userTime}}分钟'])
        if stripped == '您的胜任力等级为合格胜任':
            replacements += 1
            return set_runs(paragraph, ['您的胜任力等级为', '{{overall.level}}'])
        if stripped.startswith('特别说明：'):
            replacements += 1
            return set_whole(paragraph, '特别说明：{{report.disclaimer}}')
        if stripped.startswith('提示：'):
            replacements += 1
            return set_whole(paragraph, '提示：{{validity.notice}}')
        if stripped.startswith('【得分】') and group_score_index < 2:
            code = ('general_ability', 'psychological_quality')[group_score_index]
            group_score_index += 1
            replacements += 1
            return set_whole(paragraph, f'【得分】 {{{{group.{code}.score}}}}')
        if stripped.startswith('【评价等级】') and group_level_index < 2:
            code = ('general_ability', 'psychological_quality')[group_level_index]
            group_level_index += 1
            replacements += 1
            return set_whole(paragraph, f'【评价等级】{{{{group.{code}.level}}}}')
        if stripped.startswith('【维度说明】') and group_description_index < 2:
            code = ('general_ability', 'psychological_quality')[group_description_index]
            group_description_index += 1
            replacements += 1
            return set_whole(paragraph, f'【维度说明】{{{{group.{code}.description}}}}')
        if stripped.startswith('测评得分：') and score_index < len(DIMENSIONS):
            dimension_id = DIMENSIONS[score_index][1]
            score_index += 1
            replacements += 1
            return set_whole(paragraph, f'测评得分：{{{{dimension.{dimension_id}.score}}}}分')
        if stripped.startswith('评价等级：') and level_index < len(DIMENSIONS):
            dimension_id = DIMENSIONS[level_index][1]
            level_index += 1
            replacements += 1
            return set_whole(paragraph, f'评价等级：{{{{dimension.{dimension_id}.level}}}}')
        if stripped.startswith('【诊断】'):
            if diagnosis_index == -1:
                token = '{{overall.diagnosis}}'
            elif diagnosis_index < len(DIMENSIONS):
                token = f'{{{{dimension.{DIMENSIONS[diagnosis_index][1]}.diagnosis}}}}'
            else:
                return paragraph
            diagnosis_index += 1
            replacements += 1
            return set_whole(paragraph, '【诊断】' + token)
        return paragraph

    transformed = PARAGRAPH_RE.sub(transform, document)
    required_tokens = 49  # dimension definitions remain customer-maintained fixed Word content
    if replacements != required_tokens:
        raise RuntimeError(f'placeholder replacements={replacements}, want {required_tokens}')
    unresolved_dimensions = [name for name, count in title_counts.items() if count < 2]
    if unresolved_dimensions:
        raise RuntimeError('dimension titles missing: ' + ','.join(unresolved_dimensions))
    return transformed


def build(source, output):
    output.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(source) as archive, zipfile.ZipFile(output, 'w') as generated:
        for item in archive.infolist():
            data = archive.read(item.filename)
            if item.filename == 'word/document.xml':
                data = transform_document(data.decode('utf-8')).encode('utf-8')
            compression = zipfile.ZIP_STORED if item.is_dir() or not data else zipfile.ZIP_DEFLATED
            generated.writestr(item.filename, data, compress_type=compression)
    with zipfile.ZipFile(output) as check:
        document = check.read('word/document.xml').decode('utf-8')
        tokens = sorted(set(re.findall(r'\{\{[a-zA-Z0-9_.-]+\}\}', document)))
        if len(tokens) != 49:
            raise RuntimeError(f'unique placeholder count={len(tokens)}, want 49')
    print('PHASE1_WORD_TEMPLATE_BUILT')
    print(f'output={output}')
    print(f'placeholders={len(tokens)}|charts=12')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--source', type=Path, default=DEFAULT_SOURCE)
    parser.add_argument('--output', type=Path, default=DEFAULT_OUTPUT)
    args = parser.parse_args()
    build(args.source.resolve(), args.output.resolve())


if __name__ == '__main__':
    main()
