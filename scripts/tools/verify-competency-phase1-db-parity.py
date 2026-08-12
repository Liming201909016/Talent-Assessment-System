#!/usr/bin/env python3
import argparse
import csv
from pathlib import Path


def parse_args():
    parser = argparse.ArgumentParser(description='Compare staging phase-1 question and dimension dumps with the validated CSV package.')
    parser.add_argument('--csv-dir', type=Path, required=True)
    parser.add_argument('--questions-tsv', type=Path, required=True)
    parser.add_argument('--dimensions-tsv', type=Path, required=True)
    return parser.parse_args()


def rows(path, delimiter=',', encoding='utf-8-sig'):
    with path.open('r', encoding=encoding, newline='') as stream:
        return list(csv.DictReader(stream, delimiter=delimiter))


def text(value):
    return '' if value is None else str(value).strip()


def hex_text(value):
    return str(value).encode('utf-8').hex().upper()


def main():
    args = parse_args()
    questions = rows(args.csv_dir / 'phase1-questions.csv')
    expected_questions = sorted([
        (
            text(row['题目编号']), text(row['题目类型']), text(row['维度ID']), text(row['维度内题号']),
            hex_text(row['题目内容']), hex_text(row['考察点']), text(row['计分方向']),
            '0' if text(row['启用状态']) == 'enabled' else '1',
        )
        for row in questions
    ])
    actual_questions = sorted([
        tuple(text(row[key]) for key in ['question_code', 'question_type', 'dimension_id', 'item_no', 'content_hex', 'observation_hex', 'direction', 'status'])
        for row in rows(args.questions_tsv, delimiter='\t', encoding='utf-8')
    ])
    if actual_questions != expected_questions:
        expected = set(expected_questions)
        actual = set(actual_questions)
        raise RuntimeError(f'question parity mismatch: missing={len(expected - actual)}, unexpected={len(actual - expected)}')

    dimensions = rows(args.csv_dir / 'phase1-dimensions.csv')
    expected_dimensions = sorted([
        (
            text(row['维度ID']), text(row['维度编号']), hex_text(row['维度名称']),
            text(row['顺序']), '0' if text(row['启用状态']) == 'enabled' else '1',
        )
        for row in dimensions
    ])
    actual_dimensions = sorted([
        tuple(text(row[key]) for key in ['dimension_id', 'dimension_code', 'name_hex', 'display_order', 'status'])
        for row in rows(args.dimensions_tsv, delimiter='\t', encoding='utf-8')
    ])
    if actual_dimensions != expected_dimensions:
        expected = set(expected_dimensions)
        actual = set(actual_dimensions)
        raise RuntimeError(f'dimension parity mismatch: missing={len(expected - actual)}, unexpected={len(actual - expected)}')

    print('PHASE1_STAGING_SOURCE_PARITY_PASS')
    print(f'questions={len(actual_questions)}|dimensions={len(actual_dimensions)}')
    print('question_fields=code/type/dimension/item/content/observation/direction/status')
    print('dimension_fields=id/code/name/order/status')


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        print(f'PHASE1_STAGING_SOURCE_PARITY_FAILED: {error}')
        raise SystemExit(1)
