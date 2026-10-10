#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT=/opt/talent-assessment
SERVICE=talent-assessment
EXPECTED_HOST=iZ0yosjdcen2p4Z
STAMP=${1:?stamp required}
BACKUP="$ROOT/backups/phase1_history_delete_${STAMP}"
CLIENT=/run/phase1-history-delete.cnf
WORK=/tmp/phase1-history-delete-${STAMP}
UPLOAD_ROOT=/opt/talent-assessment/tmp/uploadPath
CURRENT_DIMENSIONS="'competency-a1-01','competency-a1-02','competency-a1-03','competency-a1-04','competency-a1-05','competency-b1-01','competency-b1-02','competency-b1-03','competency-b1-04','competency-b1-05'"
mutating=0
completed=0

cleanup() {
  rm -f "$CLIENT"
  rm -rf "$WORK"
}

rollback() {
  local rc=$?
  if [ "$mutating" -eq 1 ] && [ "$completed" -eq 0 ]; then
    set +e
    systemctl stop "$SERVICE"
    gunzip -c "$BACKUP/element.sql.gz" | mysql --defaults-extra-file="$CLIENT" element
    if [ -d "$BACKUP/report-files" ]; then
      cp -a "$BACKUP/report-files/." /
    fi
    systemctl start "$SERVICE"
    echo 'ROLLBACK_ATTEMPTED=1'
    set -e
  fi
  cleanup
  exit "$rc"
}
trap rollback EXIT HUP INT TERM

[ "$(hostname)" = "$EXPECTED_HOST" ]
[ "$(id -u)" -eq 0 ]
[ "$(systemctl is-active "$SERVICE")" = active ]
[ ! -e "$BACKUP" ]
install -d -m 700 "$WORK"

python3 - "$CLIENT" <<'PY'
from pathlib import Path
import re,sys
out=Path(sys.argv[1]); text='\n'.join(p.read_text(encoding='utf-8-sig').replace('\r','') for p in (Path('/opt/talent-assessment/configs/application.yml'),Path('/opt/talent-assessment/configs/application-production.yml')) if p.exists()); d=''
for line in text.splitlines():
    if line.strip().startswith('dsn:'): d=line.split(':',1)[1].strip().strip('"\'')
m=re.fullmatch(r'([^:]+):([^@]*)@tcp\(([^:)]+):(\d+)\)/([^?]+)(?:\?.*)?',d)
if not m or m.group(5)!='element': raise SystemExit('DB_DSN_INVALID')
u,p,h,port,_=m.groups(); esc=lambda v:v.replace('\\','\\\\').replace('"','\\"')
out.write_text('[client]\nuser="%s"\npassword="%s"\nhost="%s"\nport=%s\n'%(esc(u),esc(p),esc(h),port),encoding='utf-8'); out.chmod(0o600)
PY
mysqlq(){ mysql --defaults-extra-file="$CLIENT" --batch --skip-column-names --raw element -e "$1"; }

E="SELECT id FROM el_exam WHERE assessment_type='competency'"
P="SELECT id FROM el_paper WHERE exam_id IN ($E)"
R="SELECT id FROM el_competency_result_run WHERE exam_id IN ($E) OR paper_id IN ($P)"
CR="SELECT id FROM el_competency_report WHERE exam_id IN ($E) OR paper_id IN ($P)"
Q="SELECT id FROM el_qu WHERE (dimension_id IS NOT NULL OR competency_question_type IS NOT NULL) AND (dimension_id IS NULL OR dimension_id NOT IN ($CURRENT_DIMENSIONS))"

expect(){ local got; got=$(mysqlq "$2"); [ "$got" = "$3" ] || { printf 'PREFLIGHT_MISMATCH=%s:%s:%s\n' "$1" "$got" "$3"; return 1; }; }
expect active_papers 'SELECT COUNT(*) FROM el_paper WHERE state=1' 0
expect exam "SELECT COUNT(*) FROM el_exam WHERE id IN ($E)" 9
expect candidate "SELECT COUNT(*) FROM el_candidate WHERE exam_id IN ($E)" 0
expect paper "SELECT COUNT(*) FROM el_paper WHERE id IN ($P)" 22
expect paper_qu "SELECT COUNT(*) FROM el_paper_qu WHERE paper_id IN ($P)" 1406
expect exam_dimension "SELECT COUNT(*) FROM el_exam_competency_dimension WHERE exam_id IN ($E)" 70
expect exam_question "SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id IN ($E)" 548
expect result "SELECT COUNT(*) FROM el_competency_result WHERE exam_id IN ($E) OR paper_id IN ($P)" 19
expect dimension_result "SELECT COUNT(*) FROM el_competency_dimension_result WHERE paper_id IN ($P)" 147
expect report "SELECT COUNT(*) FROM el_competency_report WHERE id IN ($CR)" 5
expect report_audit "SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id IN ($P) OR report_id IN ($CR)" 11
expect report_files "SELECT COUNT(*) FROM el_competency_report WHERE id IN ($CR) AND pdf_path<>''" 5
expect old_question "SELECT COUNT(*) FROM el_qu WHERE id IN ($Q)" 454
expect old_dimension "SELECT COUNT(*) FROM el_competency_dimension WHERE id NOT IN ($CURRENT_DIMENSIONS)" 48
expect current_question "SELECT COUNT(*) FROM el_qu WHERE dimension_id IN ($CURRENT_DIMENSIONS)" 90
expect old_question_outside_papers "SELECT COUNT(*) FROM el_paper_qu WHERE qu_id IN ($Q) AND paper_id NOT IN ($P)" 0
expect old_question_outside_exams "SELECT COUNT(*) FROM el_exam_competency_question WHERE source_qu_id IN ($Q) AND exam_id NOT IN ($E)" 0
expect old_dimension_outside_exams "SELECT COUNT(*) FROM el_exam_competency_dimension WHERE dimension_id NOT IN ($CURRENT_DIMENSIONS) AND exam_id NOT IN ($E)" 0
expect paper_answers "SELECT COUNT(*) FROM el_paper_qu_answer WHERE paper_id IN ($P)" 0
expect result_runs "SELECT COUNT(*) FROM el_competency_result_run WHERE id IN ($R)" 0
expect mng_cross_links "SELECT (SELECT COUNT(*) FROM el_mng_exam_draft WHERE exam_id IN ($E))+(SELECT COUNT(*) FROM el_mng_exam_profile WHERE exam_id IN ($E))+(SELECT COUNT(*) FROM el_mng_paper_snapshot WHERE exam_id IN ($E) OR paper_id IN ($P))+(SELECT COUNT(*) FROM el_mng_result_run WHERE exam_id IN ($E) OR paper_id IN ($P))+(SELECT COUNT(*) FROM el_mng_report_revision WHERE exam_id IN ($E) OR paper_id IN ($P))+(SELECT COUNT(*) FROM el_mng_report_reissue WHERE exam_id IN ($E) OR paper_id IN ($P))" 0

systemctl stop "$SERVICE"
expect active_papers_after_stop 'SELECT COUNT(*) FROM el_paper WHERE state=1' 0
install -d -m 700 "$BACKUP/report-files"
mysqldump --defaults-extra-file="$CLIENT" --single-transaction --routines --triggers --events --add-drop-table element | gzip -9 > "$BACKUP/element.sql.gz"
gzip -t "$BACKUP/element.sql.gz"
mysqlq "SELECT pdf_path,pdf_sha256,pdf_size FROM el_competency_report WHERE id IN ($CR) ORDER BY id" > "$WORK/report-files.tsv"
python3 - "$WORK/report-files.tsv" "$BACKUP/report-files" "$UPLOAD_ROOT" <<'PY'
from pathlib import Path
import hashlib,shutil,sys
rows=Path(sys.argv[1]).read_text(encoding='utf-8').splitlines(); backup=Path(sys.argv[2]); root=Path(sys.argv[3]).resolve()
if len(rows)!=5: raise SystemExit('REPORT_FILE_COUNT_MISMATCH')
seen=set()
for row in rows:
    parts=row.split('\t')
    if len(parts)!=3: raise SystemExit('REPORT_BINDING_INVALID')
    raw,expected_sha,size_text=parts; path=Path(raw)
    if not path.is_absolute() or path.suffix.lower()!='.pdf' or path.is_symlink(): raise SystemExit('REPORT_PATH_UNSAFE')
    resolved=path.resolve(strict=True)
    if root not in resolved.parents or resolved in seen: raise SystemExit('REPORT_PATH_OUTSIDE_ROOT_OR_DUPLICATE')
    seen.add(resolved); data=resolved.read_bytes()
    if len(data)!=int(size_text) or hashlib.sha256(data).hexdigest()!=expected_sha: raise SystemExit('REPORT_FILE_MISMATCH')
    target=backup / resolved.relative_to(Path('/')); target.parent.mkdir(parents=True,exist_ok=True); shutil.copy2(resolved,target)
PY
sha256sum "$BACKUP/element.sql.gz" > "$BACKUP/SHA256SUMS"
find "$BACKUP" -type d -exec chmod 700 {} +
find "$BACKUP" -type f -exec chmod 600 {} +

hex_ids() {
  local rows
  rows=$(mysqlq "$1")
  if [ -z "$rows" ]; then
    printf "UNHEX('00')"
    return
  fi
  printf '%s\n' "$rows" | awk 'BEGIN{first=1}{if(!first)printf ",";printf "UNHEX(\047%s\047)",$0;first=0}'
}
number_ids() {
  local rows
  rows=$(mysqlq "$1")
  if [ -z "$rows" ]; then
    printf '0'
    return
  fi
  printf '%s\n' "$rows" | awk 'BEGIN{first=1}{if($0 !~ /^[0-9]+$/)exit 2;if(!first)printf ",";printf "%s",$0;first=0}'
}

EXAM_IDS=$(hex_ids "SELECT HEX(id) FROM el_exam WHERE assessment_type='competency' ORDER BY id")
PAPER_IDS=$(hex_ids "SELECT HEX(id) FROM el_paper WHERE exam_id IN ($E) ORDER BY id")
RUN_IDS=$(hex_ids "SELECT HEX(id) FROM el_competency_result_run WHERE exam_id IN ($E) OR paper_id IN ($P) ORDER BY id")
REPORT_IDS=$(hex_ids "SELECT HEX(id) FROM el_competency_report WHERE exam_id IN ($E) OR paper_id IN ($P) ORDER BY id")
QUESTION_IDS=$(hex_ids "SELECT HEX(id) FROM el_qu WHERE (dimension_id IS NOT NULL OR competency_question_type IS NOT NULL) AND (dimension_id IS NULL OR dimension_id NOT IN ($CURRENT_DIMENSIONS)) ORDER BY id")
DIMENSION_IDS=$(hex_ids "SELECT HEX(id) FROM el_competency_dimension WHERE id NOT IN ($CURRENT_DIMENSIONS) ORDER BY id")
OPER_IDS=$(number_ids "SELECT oper_id FROM sys_oper_log WHERE method='ExamHandler.ExportRawAnswers' AND oper_param REGEXP CONCAT('(', (SELECT GROUP_CONCAT(id SEPARATOR '|') FROM el_exam WHERE assessment_type='competency'), ')') ORDER BY oper_id")

cat > "$WORK/delete.sql" <<SQL
SET NAMES utf8mb4;
SET SESSION SQL_SAFE_UPDATES=1;
START TRANSACTION;
DELETE FROM el_competency_report_audit WHERE id IN (SELECT id FROM (SELECT id FROM el_competency_report_audit WHERE paper_id IN ($PAPER_IDS) OR report_id IN ($REPORT_IDS)) x);
DELETE FROM el_competency_report WHERE id IN ($REPORT_IDS);
DELETE FROM el_competency_dimension_result WHERE id IN (SELECT id FROM (SELECT id FROM el_competency_dimension_result WHERE paper_id IN ($PAPER_IDS)) x);
DELETE FROM el_competency_result WHERE paper_id IN ($PAPER_IDS);
DELETE FROM el_paper_qu WHERE id IN (SELECT id FROM (SELECT id FROM el_paper_qu WHERE paper_id IN ($PAPER_IDS)) x);
DELETE FROM el_exam_competency_question WHERE id IN (SELECT id FROM (SELECT id FROM el_exam_competency_question WHERE exam_id IN ($EXAM_IDS)) x);
DELETE FROM el_exam_competency_dimension WHERE id IN (SELECT id FROM (SELECT id FROM el_exam_competency_dimension WHERE exam_id IN ($EXAM_IDS)) x);
DELETE FROM el_paper WHERE id IN ($PAPER_IDS);
DELETE FROM el_exam WHERE id IN ($EXAM_IDS);
DELETE FROM el_qu WHERE id IN ($QUESTION_IDS);
DELETE FROM el_competency_dimension WHERE id IN ($DIMENSION_IDS);
DELETE FROM sys_oper_log WHERE oper_id IN ($OPER_IDS);
COMMIT;
SQL
chmod 600 "$WORK/delete.sql"
mutating=1
mysql --defaults-extra-file="$CLIENT" element < "$WORK/delete.sql"
python3 - "$WORK/report-files.tsv" "$UPLOAD_ROOT" <<'PY'
from pathlib import Path
import sys
root=Path(sys.argv[2]).resolve()
for row in Path(sys.argv[1]).read_text(encoding='utf-8').splitlines():
    path=Path(row.split('\t',1)[0]).resolve(strict=False)
    if root not in path.parents: raise SystemExit('REPORT_PATH_OUTSIDE_ROOT')
    if path.exists(): path.unlink()
PY

expect_final(){ local got; got=$(mysqlq "$2"); [ "$got" = "$3" ] || { printf 'ACCEPTANCE_MISMATCH=%s:%s:%s\n' "$1" "$got" "$3"; return 1; }; }
expect_final active_papers 'SELECT COUNT(*) FROM el_paper WHERE state=1' 0
expect_final competency_exams "SELECT COUNT(*) FROM el_exam WHERE assessment_type='competency'" 0
expect_final competency_papers "SELECT COUNT(*) FROM el_paper WHERE exam_id IN (SELECT id FROM el_exam WHERE assessment_type='competency')" 0
expect_final competency_candidates "SELECT COUNT(*) FROM el_candidate WHERE exam_id IN (SELECT id FROM el_exam WHERE assessment_type='competency')" 0
expect_final competency_results 'SELECT COUNT(*) FROM el_competency_result' 0
expect_final competency_reports 'SELECT COUNT(*) FROM el_competency_report' 0
expect_final old_questions "SELECT COUNT(*) FROM el_qu WHERE (dimension_id IS NOT NULL OR competency_question_type IS NOT NULL) AND (dimension_id IS NULL OR dimension_id NOT IN ($CURRENT_DIMENSIONS))" 0
expect_final current_questions "SELECT COUNT(*) FROM el_qu WHERE dimension_id IN ($CURRENT_DIMENSIONS)" 90
expect_final total_competency_questions "SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NOT NULL OR competency_question_type IS NOT NULL" 90
expect_final old_dimensions "SELECT COUNT(*) FROM el_competency_dimension WHERE id NOT IN ($CURRENT_DIMENSIONS)" 0
expect_final current_dimensions "SELECT COUNT(*) FROM el_competency_dimension WHERE id IN ($CURRENT_DIMENSIONS)" 10
expect_final traditional_questions "SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NULL AND competency_question_type IS NULL" 858
expect_final traditional_repos 'SELECT COUNT(*) FROM el_repo' 8
expect_final v1_text "SELECT COUNT(*) FROM el_competency_report_text WHERE content_version='competency-phase1-content-v1'" 66
expect_final v2_text "SELECT COUNT(*) FROM el_competency_report_text WHERE content_version='competency-phase1-content-v2'" 124
expect_final report_package 'SELECT COUNT(*) FROM el_competency_report_content_package' 2
expect_final remaining_report_files "SELECT COUNT(*) FROM el_competency_report WHERE pdf_path<>''" 0

systemctl start "$SERVICE"
for _ in $(seq 1 30); do [ "$(systemctl is-active "$SERVICE")" = active ] && curl -fsS http://127.0.0.1:8092/health >/dev/null && break; done
[ "$(systemctl is-active "$SERVICE")" = active ]
[ "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8090/prod-api/health)" = 200 ]
completed=1
trap cleanup EXIT HUP INT TERM
printf 'BACKUP=%s\n' "$BACKUP"
printf 'DELETED_EXAMS=9\nDELETED_PAPERS=22\nDELETED_CANDIDATES=0\nDELETED_RESULTS=19\nDELETED_REPORTS=5\nDELETED_QUESTIONS=454\nDELETED_DIMENSIONS=48\n'
printf 'CURRENT_00401_QUESTIONS=90\nTRADITIONAL_QUESTIONS=858\n'
printf 'PHASE1_HISTORY_DELETE=PASS\n'
