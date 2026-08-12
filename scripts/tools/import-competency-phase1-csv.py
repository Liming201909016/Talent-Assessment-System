#!/usr/bin/env python3
import argparse
import csv
import hashlib
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
VALIDATOR = ROOT / 'scripts' / 'tools' / 'validate-competency-phase1-csv.js'
CONTENT_VERSION = 'competency-phase1-content-v1'
AUDIENCE = 'frontline_employee'


def parse_args():
    parser = argparse.ArgumentParser(description='Generate an inactive staging candidate import from the validated phase-1 CSV package.')
    parser.add_argument('--csv-dir', type=Path, required=True)
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument('--preview', action='store_true')
    mode.add_argument('--sql-output', type=Path)
    return parser.parse_args()


def read_rows(directory, name):
    path = directory / name
    with path.open('r', encoding='utf-8-sig', newline='') as stream:
        return list(csv.DictReader(stream))


def text(value):
    return '' if value is None else str(value).strip()


def sql_text(value):
    return "UNHEX('%s')" % str(value).encode('utf-8').hex()


def content_sha(directory):
    digest = hashlib.sha256()
    for path in sorted(directory.glob('phase1-*.csv')):
        if path.name == 'phase1-package.csv':
            continue
        digest.update(path.name.encode('utf-8'))
        digest.update(b'\n')
        digest.update(path.read_bytes())
        digest.update(b'\n')
    return digest.hexdigest()


def validate(directory):
    result = subprocess.run(
        ['node', str(VALIDATOR), str(directory)],
        cwd=ROOT,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or result.stdout.strip() or 'CSV validation failed')


def report_rows(directory):
    dimensions = read_rows(directory, 'phase1-dimension-levels.csv')
    overall = read_rows(directory, 'phase1-overall-levels.csv')
    static = read_rows(directory, 'phase1-report-static-texts.csv')
    rows = []
    for row in overall:
        rows.append(('overall', '', text(row['总体等级编号']), text(row['诊断'])))
    for row in static:
        content_type = text(row['内容类型'])
        rows.append((content_type, text(row['匹配键']), text(row['等级编号']), text(row['内容'])))
    for row in dimensions:
        rows.append(('dimension', text(row['维度ID']), text(row['等级编号']), text(row['诊断与建议'])))
    if len(rows) != 66:
        raise RuntimeError(f'report text row count={len(rows)}, want 66')
    if any(not content for _, _, _, content in rows):
        raise RuntimeError('candidate report text contains blank content')
    keys = [(kind, dimension, level) for kind, dimension, level, _ in rows]
    if len(set(keys)) != len(keys):
        raise RuntimeError('candidate report text lookup keys are not unique')
    return rows


def stable_id(content_type, dimension_id, level_code):
    key = f'{content_type}|{dimension_id}|{level_code}'.encode('utf-8')
    return 'phase1-candidate-' + hashlib.sha256(key).hexdigest()[:32]


def insert_report_text(row):
    content_type, dimension_id, level_code, content = row
    values = [
        sql_text(stable_id(content_type, dimension_id, level_code)),
        sql_text(CONTENT_VERSION), sql_text(AUDIENCE), sql_text(content_type),
        sql_text(dimension_id), sql_text(level_code), sql_text(content), sql_text(''),
        '1', '1', 'NOW()', 'NOW()',
    ]
    return (
        'INSERT INTO `el_competency_report_text` '
        '(`id`,`content_version`,`audience`,`content_type`,`dimension_id`,`level_code`,`content`,`disclaimer`,`is_temporary`,`status`,`create_time`,`update_time`) '
        f"VALUES ({','.join(values)}) ON DUPLICATE KEY UPDATE "
        '`content_version`=VALUES(`content_version`),`audience`=VALUES(`audience`),'
        '`content_type`=VALUES(`content_type`),`dimension_id`=VALUES(`dimension_id`),'
        '`level_code`=VALUES(`level_code`),`content`=VALUES(`content`),'
        '`disclaimer`=VALUES(`disclaimer`),`is_temporary`=VALUES(`is_temporary`),'
        '`status`=VALUES(`status`),`update_time`=VALUES(`update_time`);'
    )


def insert_package(package, source_sha):
    values = [
        sql_text('phase1-candidate-draft-v1'), sql_text(text(package['产品版本'])),
        sql_text(text(package['评分版本'])), sql_text(text(package['内容版本'])),
        sql_text(text(package['报告模板版本'])), sql_text(text(package['报告对象'])),
        sql_text('draft'), sql_text(''), 'NULL', sql_text(''), 'NULL',
        sql_text(text(package['原始题本SHA256'])), sql_text(source_sha), sql_text(''),
        sql_text(text(package['最终免责声明'])), 'NOW()', 'NOW()',
    ]
    return (
        'INSERT INTO `el_competency_report_content_package` '
        '(`id`,`product_version`,`scoring_version`,`content_version`,`template_version`,`audience`,`approval_status`,'
        '`content_approved_by`,`content_approved_at`,`psychometric_approved_by`,`psychometric_approved_at`,'
        '`question_source_sha256`,`content_source_sha256`,`effective_environment`,`disclaimer`,`create_time`,`update_time`) '
        f"VALUES ({','.join(values)}) ON DUPLICATE KEY UPDATE "
        '`product_version`=VALUES(`product_version`),`scoring_version`=VALUES(`scoring_version`),'
        '`content_version`=VALUES(`content_version`),`template_version`=VALUES(`template_version`),'
        '`audience`=VALUES(`audience`),`approval_status`=VALUES(`approval_status`),'
        '`content_approved_by`=VALUES(`content_approved_by`),`content_approved_at`=VALUES(`content_approved_at`),'
        '`psychometric_approved_by`=VALUES(`psychometric_approved_by`),`psychometric_approved_at`=VALUES(`psychometric_approved_at`),'
        '`question_source_sha256`=VALUES(`question_source_sha256`),`content_source_sha256`=VALUES(`content_source_sha256`),'
        '`effective_environment`=VALUES(`effective_environment`),`disclaimer`=VALUES(`disclaimer`),'
        '`update_time`=VALUES(`update_time`);'
    )


def build_sql(directory, rows, package):
    statements = ['SET NAMES utf8mb4;', 'START TRANSACTION;']
    statements.extend(insert_report_text(row) for row in rows)
    statements.append(insert_package(package, content_sha(directory)))
    statements.append('COMMIT;')
    return '\n'.join(statements) + '\n'


def main():
    args = parse_args()
    directory = args.csv_dir.resolve()
    validate(directory)
    packages = read_rows(directory, 'phase1-package.csv')
    if len(packages) != 1 or text(packages[0]['批准状态']) != 'draft':
        raise RuntimeError('candidate import requires exactly one draft package')
    rows = report_rows(directory)
    counts = {
        'overall': sum(kind == 'overall' for kind, _, _, _ in rows),
        'templates': sum(kind == 'template' for kind, _, _, _ in rows),
        'groups': sum(kind == 'group' for kind, _, _, _ in rows),
        'dimensions': sum(kind == 'dimension' for kind, _, _, _ in rows),
        'validity': sum(kind == 'validity' for kind, _, _, _ in rows),
    }
    if args.preview:
        print('PHASE1_CANDIDATE_IMPORT_PREVIEW_PASS')
        print(f'report_texts={len(rows)}')
        print(f"overall={counts['overall']}|templates={counts['templates']}|groups={counts['groups']}|dimensions={counts['dimensions']}|validity={counts['validity']}")
        print('package_status=draft')
        print('rows_status=inactive|is_temporary=1')
        print(f'content_source_sha256={content_sha(directory)}')
        return
    args.sql_output.parent.mkdir(parents=True, exist_ok=True)
    args.sql_output.write_text(build_sql(directory, rows, packages[0]), encoding='utf-8', newline='\n')
    print('PHASE1_CANDIDATE_IMPORT_SQL_GENERATED')
    print(f'output={args.sql_output.resolve()}')
    print(f'report_texts={len(rows)}|package_status=draft|rows_status=inactive')


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        print(f'PHASE1_CANDIDATE_IMPORT_FAILED: {error}', file=sys.stderr)
        raise SystemExit(1)
