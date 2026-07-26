#!/usr/bin/env bash
# FB-041 round-2 staging verification (read-mostly, only mutates one seeded MBTI paper)
DB_USER=positive_app
DB_PASS='Positive#2026App!'
DB_NAME=element

CAP=$(curl -s http://127.0.0.1:8092/captchaImage)
UUID=$(echo "$CAP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('uuid',''))")
CODE=$(redis-cli -n 1 GET "captcha_codes:$UUID" 2>/dev/null | tr -d '"')
TOK=$(curl -s -X POST http://127.0.0.1:8092/login -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"cp@1234\",\"code\":\"$CODE\",\"uuid\":\"$UUID\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")
echo "tok_len=${#TOK}"
if [ "${#TOK}" -lt 100 ]; then echo "LOGIN_FAILED"; exit 1; fi

dump_rows() {
  local f="$1"
  python3 - <<PY
import zipfile,re
z=zipfile.ZipFile("$f")
ss=re.findall(r'<t[^>]*>([^<]*)</t>', z.read('xl/sharedStrings.xml').decode('utf-8',errors='ignore')) if 'xl/sharedStrings.xml' in z.namelist() else []
sh=z.read('xl/worksheets/sheet1.xml').decode('utf-8',errors='ignore')
rows=re.findall(r'<row r="(\d+)"', sh)
print('  xlsx_rows', len(rows))
PY
}

echo
echo "=== [1] startup migration log (should be idempotent on staging) ==="
journalctl -u talent-assessment --since '15 min ago' --no-pager 2>/dev/null \
  | grep -iE 'added column on el_tester|backfilled el_tester' || echo "  (no migration log => columns already exist, expected)"

echo
echo "=== [2] startup ensureTesterMbtiColumns log absent test ==="
# Confirm no Warning either
journalctl -u talent-assessment --since '15 min ago' --no-pager 2>/dev/null \
  | grep -iE 'add column on el_tester failed|backfill el_tester' || echo "  (no warning, OK)"

echo
echo "=== [3] Export propagation: nonexistent examId ==="
curl -s -X POST http://127.0.0.1:8092/exam/api/exam/exam/export-raw-data \
  -H "Authorization: Bearer $TOK" -H "Content-Type: application/x-www-form-urlencoded" \
  -d "examId=9999999999999999999" | head -c 300
echo

echo
echo "=== [4] Export propagation: empty examId ==="
curl -s -X POST http://127.0.0.1:8092/exam/api/exam/exam/export-raw-data \
  -H "Authorization: Bearer $TOK" -H "Content-Type: application/x-www-form-urlencoded" \
  -d "examId=" | head -c 300
echo

echo
echo "=== [5] Closed MBTI (12 testers) export ==="
out=/tmp/r2-closed-mbti.xlsx
curl -s -o $out -w "  HTTP %{http_code} size=%{size_download}\n" \
  -X POST http://127.0.0.1:8092/exam/api/exam/exam/export-raw-data \
  -H "Authorization: Bearer $TOK" -H "Content-Type: application/x-www-form-urlencoded" \
  -d "examId=1776746714029306200"
dump_rows $out
rm -f $out

echo
echo "=== [6] Open 002 admin export still works ==="
EID=$(mysql -u$DB_USER -p$DB_PASS $DB_NAME -Nse "SELECT e.id FROM el_exam e JOIN el_exam_repo er ON er.exam_id=e.id JOIN el_repo r ON r.id=er.repo_id WHERE e.is_open=1 AND r.code LIKE '002%' ORDER BY e.id DESC LIMIT 1" 2>/dev/null)
echo "  open 002 exam_id=$EID"
if [ -n "$EID" ]; then
  out=/tmp/r2-open-002.xlsx
  curl -s -o $out -w "  HTTP %{http_code} size=%{size_download}\n" \
    -X POST http://127.0.0.1:8092/exam/api/exam/exam/export-raw-data \
    -H "Authorization: Bearer $TOK" -H "Content-Type: application/x-www-form-urlencoded" \
    -d "examId=$EID"
  dump_rows $out
  rm -f $out
fi

echo
echo "=== [7] MBTI submit happy + duplicate guard ==="
PID=1777217353411783509
# Reset paper to state=0 and seed 48 answers
mysql -u$DB_USER -p$DB_PASS $DB_NAME -e "UPDATE el_paper SET state=0, user_time=0 WHERE id='$PID'; UPDATE el_candidate SET end_time=NULL, pdf_flag=0 WHERE paper_id='$PID';" 2>&1 | grep -v Warning
mysql -u$DB_USER -p$DB_PASS $DB_NAME -e "DELETE FROM el_mbti_answer WHERE paper_id='$PID'; SET @row=0; INSERT INTO el_mbti_answer (id, paper_id, qu_id, score_a, score_b, answered, create_time) SELECT CONCAT('R2SEED', LPAD(@row := @row + 1, 4, '0')), '$PID', qr.qu_id, 3, 2, 1, NOW() FROM (SELECT qr.qu_id FROM el_qu_repo qr JOIN el_exam_repo er ON er.repo_id=qr.repo_id JOIN el_paper p ON p.exam_id=er.exam_id WHERE p.id='$PID' ORDER BY qr.sort LIMIT 48) qr;" 2>&1 | grep -v Warning
echo "  --- before submit (expect state=0 end_time=NULL) ---"
mysql -u$DB_USER -p$DB_PASS $DB_NAME -t -e "SELECT 'paper' tbl,state,user_time FROM el_paper WHERE id='$PID'; SELECT 'cand' tbl,end_time,pdf_flag FROM el_candidate WHERE paper_id='$PID';" 2>&1 | grep -v Warning
echo "  --- 1st submit (expect type+scores) ---"
curl -s -X POST http://127.0.0.1:8092/exam/api/mbti/submit -H "Authorization: Bearer $TOK" -H "Content-Type: application/json" -d "{\"paperId\":\"$PID\"}"
echo
echo "  --- after submit (expect state=2 end_time set) ---"
mysql -u$DB_USER -p$DB_PASS $DB_NAME -t -e "SELECT 'paper' tbl,state,user_time FROM el_paper WHERE id='$PID'; SELECT 'cand' tbl,end_time,pdf_flag FROM el_candidate WHERE paper_id='$PID';" 2>&1 | grep -v Warning
echo "  --- 2nd submit (expect 试卷不存在或已交卷) ---"
curl -s -X POST http://127.0.0.1:8092/exam/api/mbti/submit -H "Authorization: Bearer $TOK" -H "Content-Type: application/json" -d "{\"paperId\":\"$PID\"}"
echo

echo
echo "=== [8] tester/list grouping FB-040 regression ==="
curl -s "http://127.0.0.1:8092/exam/api/tester/list?pageNum=1&pageSize=10" -H "Authorization: Bearer $TOK" \
  | python3 -c "import sys,json; d=json.load(sys.stdin); rows=d.get('rows',[]); print('  rows', len(rows), 'total', d.get('total')); [print(' ', r.get('createTime'),'|',r.get('examId'),'|',r.get('name')) for r in rows[:8]]"

echo
echo "=== [9] recent server errors ==="
journalctl -u talent-assessment --since '5 min ago' --no-pager 2>/dev/null \
  | grep -iE 'panic|fatal|level=error|ERR ' | tail -10 || echo "  (clean)"
