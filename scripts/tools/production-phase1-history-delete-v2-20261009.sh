#!/usr/bin/env bash
set -euo pipefail
umask 077
ROOT=/opt/talent-assessment
SERVICE=talent-assessment
STAMP=${1:?stamp required}
BACKUP="$ROOT/backups/phase1_history_delete_v2_${STAMP}"
CLIENT=/run/phase1-history-delete-v2.cnf
WORK=/tmp/phase1-history-delete-v2-${STAMP}
UPLOAD_ROOT=/opt/talent-assessment/tmp/uploadPath
DIMS="'competency-a1-01','competency-a1-02','competency-a1-03','competency-a1-04','competency-a1-05','competency-b1-01','competency-b1-02','competency-b1-03','competency-b1-04','competency-b1-05'"
mutating=0
completed=0
cleanup(){ rm -f "$CLIENT"; rm -rf "$WORK"; }
rollback(){
  rc=$?
  if [ "$mutating" -eq 1 ] && [ "$completed" -eq 0 ]; then
    set +e
    systemctl stop "$SERVICE"
    gunzip -c "$BACKUP/element.sql.gz" | mysql --defaults-extra-file="$CLIENT" element
    find "$BACKUP/report-files" -type f -print0 2>/dev/null | while IFS= read -r -d '' f; do target=/${f#"$BACKUP/report-files/"}; install -D -m 600 "$f" "$target"; done
    systemctl start "$SERVICE"
    chmod 0755 /
    echo ROLLBACK_ATTEMPTED=1
  fi
  cleanup
  exit "$rc"
}
trap rollback EXIT HUP INT TERM
[ "$(hostname)" = iZ0yosjdcen2p4Z ]
[ "$(id -u)" -eq 0 ]
[ "$(stat -c '%a' /)" = 755 ]
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
q(){ mysql --defaults-extra-file="$CLIENT" --batch --skip-column-names --raw element -e "$1"; }
E="SELECT id FROM el_exam WHERE assessment_type='competency'"; P="SELECT id FROM el_paper WHERE exam_id IN ($E)"; CR="SELECT id FROM el_competency_report WHERE exam_id IN ($E) OR paper_id IN ($P)"; Q="SELECT id FROM el_qu WHERE (dimension_id IS NOT NULL OR competency_question_type IS NOT NULL) AND (dimension_id IS NULL OR dimension_id NOT IN ($DIMS))"
expect(){ got=$(q "$2"); [ "$got" = "$3" ] || { echo "MISMATCH=$1:$got:$3"; return 1; }; }
expect active 'SELECT COUNT(*) FROM el_paper WHERE state=1' 0
expect exam "SELECT COUNT(*) FROM el_exam WHERE id IN ($E)" 9
expect paper "SELECT COUNT(*) FROM el_paper WHERE id IN ($P)" 22
expect paper_qu "SELECT COUNT(*) FROM el_paper_qu WHERE paper_id IN ($P)" 1406
expect exam_dim "SELECT COUNT(*) FROM el_exam_competency_dimension WHERE exam_id IN ($E)" 70
expect exam_qu "SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id IN ($E)" 548
expect result "SELECT COUNT(*) FROM el_competency_result WHERE exam_id IN ($E) OR paper_id IN ($P)" 19
expect dim_result "SELECT COUNT(*) FROM el_competency_dimension_result WHERE paper_id IN ($P)" 147
expect report "SELECT COUNT(*) FROM el_competency_report WHERE id IN ($CR)" 5
expect audit "SELECT COUNT(*) FROM el_competency_report_audit WHERE paper_id IN ($P) OR report_id IN ($CR)" 11
expect old_qu "SELECT COUNT(*) FROM el_qu WHERE id IN ($Q)" 454
expect old_dim "SELECT COUNT(*) FROM el_competency_dimension WHERE id NOT IN ($DIMS)" 48
expect current_qu "SELECT COUNT(*) FROM el_qu WHERE dimension_id IN ($DIMS)" 90
expect traditional "SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NULL AND competency_question_type IS NULL" 858
# All omitted ownership tables were verified zero in the prior read-only closure.

systemctl stop "$SERVICE"
expect active_after_stop 'SELECT COUNT(*) FROM el_paper WHERE state=1' 0
install -d -m 700 "$BACKUP/report-files"
mysqldump --defaults-extra-file="$CLIENT" --single-transaction --routines --triggers --events --add-drop-table element | gzip -9 > "$BACKUP/element.sql.gz"
gzip -t "$BACKUP/element.sql.gz"
q "SELECT pdf_path,pdf_sha256,pdf_size FROM el_competency_report WHERE id IN ($CR) ORDER BY id" > "$WORK/reports.tsv"
python3 - "$WORK/reports.tsv" "$BACKUP/report-files" "$UPLOAD_ROOT" <<'PY'
from pathlib import Path
import hashlib,shutil,sys
rows=Path(sys.argv[1]).read_text().splitlines(); backup=Path(sys.argv[2]); root=Path(sys.argv[3]).resolve()
if len(rows)!=5: raise SystemExit('REPORT_COUNT')
for row in rows:
    raw,sha,size=row.split('\t'); p=Path(raw)
    if not p.is_absolute() or p.suffix.lower()!='.pdf' or p.is_symlink(): raise SystemExit('REPORT_PATH')
    p=p.resolve(strict=True)
    if root not in p.parents: raise SystemExit('REPORT_ROOT')
    data=p.read_bytes()
    if len(data)!=int(size) or hashlib.sha256(data).hexdigest()!=sha: raise SystemExit('REPORT_HASH')
    target=backup/p.relative_to('/'); target.parent.mkdir(parents=True,exist_ok=True); shutil.copy2(p,target)
PY
sha256sum "$BACKUP/element.sql.gz" > "$BACKUP/SHA256SUMS"
find "$BACKUP" -type d -exec chmod 700 {} +; find "$BACKUP" -type f -exec chmod 600 {} +

# Emit immutable primary-key literals. Binary IDs are represented as UNHEX.
hexlist(){ rows=$(q "$1"); [ -n "$rows" ] || { echo "UNHEX('00')"; return; }; echo "$rows" | awk 'BEGIN{f=1}{if(!f)printf ",";printf "UNHEX(\047%s\047)",$0;f=0}'; }
numlist(){ rows=$(q "$1"); [ -n "$rows" ] || { echo 0; return; }; echo "$rows" | awk 'BEGIN{f=1}{if($0!~/^[0-9]+$/)exit 2;if(!f)printf ",";printf "%s",$0;f=0}'; }
EXAMS=$(hexlist "SELECT HEX(id) FROM el_exam WHERE assessment_type='competency' ORDER BY id")
PAPERS=$(hexlist "SELECT HEX(id) FROM el_paper WHERE exam_id IN ($E) ORDER BY id")
REPORTS=$(hexlist "SELECT HEX(id) FROM el_competency_report WHERE id IN ($CR) ORDER BY id")
AUDITS=$(hexlist "SELECT HEX(id) FROM el_competency_report_audit WHERE paper_id IN ($P) OR report_id IN ($CR) ORDER BY id")
DIM_RESULTS=$(hexlist "SELECT HEX(id) FROM el_competency_dimension_result WHERE paper_id IN ($P) ORDER BY id")
RESULTS=$(hexlist "SELECT HEX(paper_id) FROM el_competency_result WHERE exam_id IN ($E) OR paper_id IN ($P) ORDER BY paper_id")
PAPER_QUS=$(hexlist "SELECT HEX(id) FROM el_paper_qu WHERE paper_id IN ($P) ORDER BY id")
EXAM_QUS=$(hexlist "SELECT HEX(id) FROM el_exam_competency_question WHERE exam_id IN ($E) ORDER BY id")
EXAM_DIMS=$(hexlist "SELECT HEX(id) FROM el_exam_competency_dimension WHERE exam_id IN ($E) ORDER BY id")
QUESTIONS=$(hexlist "SELECT HEX(id) FROM el_qu WHERE id IN ($Q) ORDER BY id")
DIMENSIONS=$(hexlist "SELECT HEX(id) FROM el_competency_dimension WHERE id NOT IN ($DIMS) ORDER BY id")
OPER=$(numlist "SELECT oper_id FROM sys_oper_log WHERE method='ExamHandler.ExportRawAnswers' AND oper_param REGEXP CONCAT('(', (SELECT GROUP_CONCAT(id SEPARATOR '|') FROM el_exam WHERE assessment_type='competency'), ')') ORDER BY oper_id")
cat > "$WORK/delete.sql" <<SQL
SET NAMES utf8mb4;
SET SESSION SQL_SAFE_UPDATES=1;
START TRANSACTION;
DELETE FROM el_competency_report_audit WHERE id IN ($AUDITS);
DELETE FROM el_competency_report WHERE id IN ($REPORTS);
DELETE FROM el_competency_dimension_result WHERE id IN ($DIM_RESULTS);
DELETE FROM el_competency_result WHERE paper_id IN ($RESULTS);
DELETE FROM el_paper_qu WHERE id IN ($PAPER_QUS);
DELETE FROM el_exam_competency_question WHERE id IN ($EXAM_QUS);
DELETE FROM el_exam_competency_dimension WHERE id IN ($EXAM_DIMS);
DELETE FROM el_paper WHERE id IN ($PAPERS);
DELETE FROM el_exam WHERE id IN ($EXAMS);
DELETE FROM el_qu WHERE id IN ($QUESTIONS);
DELETE FROM el_competency_dimension WHERE id IN ($DIMENSIONS);
DELETE FROM sys_oper_log WHERE oper_id IN ($OPER);
COMMIT;
SQL
chmod 600 "$WORK/delete.sql"
mutating=1
mysql --defaults-extra-file="$CLIENT" element < "$WORK/delete.sql"
python3 - "$WORK/reports.tsv" "$UPLOAD_ROOT" <<'PY'
from pathlib import Path
import sys
root=Path(sys.argv[2]).resolve()
for row in Path(sys.argv[1]).read_text().splitlines():
    p=Path(row.split('\t',1)[0]).resolve(strict=False)
    if root not in p.parents: raise SystemExit('REPORT_ROOT')
    if p.exists(): p.unlink()
PY
expect_final(){ got=$(q "$2"); [ "$got" = "$3" ] || { echo "FINAL_MISMATCH=$1:$got:$3"; return 1; }; }
expect_final exams "SELECT COUNT(*) FROM el_exam WHERE assessment_type='competency'" 0
expect_final papers "SELECT COUNT(*) FROM el_paper WHERE exam_id IN (SELECT id FROM el_exam WHERE assessment_type='competency')" 0
expect_final results 'SELECT COUNT(*) FROM el_competency_result' 0
expect_final reports 'SELECT COUNT(*) FROM el_competency_report' 0
expect_final audits 'SELECT COUNT(*) FROM el_competency_report_audit' 0
expect_final old_questions "SELECT COUNT(*) FROM el_qu WHERE (dimension_id IS NOT NULL OR competency_question_type IS NOT NULL) AND (dimension_id IS NULL OR dimension_id NOT IN ($DIMS))" 0
expect_final current_questions "SELECT COUNT(*) FROM el_qu WHERE dimension_id IN ($DIMS)" 90
expect_final total_competency "SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NOT NULL OR competency_question_type IS NOT NULL" 90
expect_final old_dimensions "SELECT COUNT(*) FROM el_competency_dimension WHERE id NOT IN ($DIMS)" 0
expect_final current_dimensions "SELECT COUNT(*) FROM el_competency_dimension WHERE id IN ($DIMS)" 10
expect_final traditional "SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NULL AND competency_question_type IS NULL" 858
expect_final text_v1 "SELECT COUNT(*) FROM el_competency_report_text WHERE content_version='competency-phase1-content-v1'" 66
expect_final text_v2 "SELECT COUNT(*) FROM el_competency_report_text WHERE content_version='competency-phase1-content-v2'" 124
expect_final packages 'SELECT COUNT(*) FROM el_competency_report_content_package' 2
systemctl start "$SERVICE"
for _ in $(seq 1 30); do systemctl is-active --quiet "$SERVICE" && curl -fsS http://127.0.0.1:8092/health >/dev/null && break; done
systemctl is-active --quiet "$SERVICE"
[ "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8090/prod-api/health)" = 200 ]
completed=1
trap cleanup EXIT HUP INT TERM
printf 'BACKUP=%s\nDELETED_EXAMS=9\nDELETED_PAPERS=22\nDELETED_RESULTS=19\nDELETED_REPORTS=5\nDELETED_QUESTIONS=454\nDELETED_DIMENSIONS=48\nCURRENT_00401_QUESTIONS=90\nTRADITIONAL_QUESTIONS=858\nPHASE1_HISTORY_DELETE_V2=PASS\n' "$BACKUP"
