#!/usr/bin/env python3
import subprocess
import sys
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
TOOL = ROOT / 'scripts' / 'tools' / 'import-competency-phase1-csv.py'
CSV_DIR = ROOT / 'scripts' / 'data' / 'competency-phase1-csv'


def run(*args):
    return subprocess.run(
        [sys.executable, str(TOOL), '--csv-dir', str(CSV_DIR), *args],
        cwd=ROOT,
        capture_output=True,
        text=True,
    )


preview = run('--preview')
assert preview.returncode == 0, preview.stderr or preview.stdout
assert 'PHASE1_CANDIDATE_IMPORT_PREVIEW_PASS' in preview.stdout
assert 'report_texts=66' in preview.stdout
assert 'overall=5|templates=7|groups=2|dimensions=50|validity=2' in preview.stdout
assert 'package_status=draft' in preview.stdout
assert 'rows_status=inactive|is_temporary=1' in preview.stdout

with tempfile.TemporaryDirectory(prefix='phase1-csv-import-') as directory:
    output = Path(directory) / 'candidate-import.sql'
    generated = run('--sql-output', str(output))
    assert generated.returncode == 0, generated.stderr or generated.stdout
    sql = output.read_text(encoding='utf-8')
    assert sql.startswith('SET NAMES utf8mb4;\nSTART TRANSACTION;\n')
    assert sql.rstrip().endswith('COMMIT;')
    content_version_hex = 'competency-phase1-content-v1'.encode('utf-8').hex()
    draft_hex = 'draft'.encode('utf-8').hex()
    approved_hex = 'approved'.encode('utf-8').hex()
    assert sql.count(content_version_hex) >= 67
    assert sql.count("UNHEX('") >= 66
    assert draft_hex in sql
    assert approved_hex not in sql
    assert 'is_temporary' in sql and 'status' in sql
    assert 'ON DUPLICATE KEY UPDATE' in sql
    assert 'PHASE1_CANDIDATE_IMPORT_SQL_GENERATED' in generated.stdout

print('COMPETENCY_PHASE1_CSV_IMPORT_CONTRACT_TEST_PASS')