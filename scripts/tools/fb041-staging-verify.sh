#!/usr/bin/env bash
# Do not use `set -e` because subsequent diagnostic blocks must run even if
# a single mysql/curl command returns non-zero.
echo "=== A. Export propagation: pass nonexistent examId ==="
CAP=$(curl -s http://127.0.0.1:8092/captchaImage)
UUID=$(echo "$CAP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('uuid',''))")
CODE=$(redis-cli -n 1 GET "captcha_codes:$UUID" 2>/dev/null | tr -d '"')
TOK=$(curl -s -X POST http://127.0.0.1:8092/login -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"cp@1234\",\"code\":\"$CODE\",\"uuid\":\"$UUID\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")
echo "tok_len=${#TOK}"
echo "--- non-existent examId (should return business error, NOT 200 empty xlsx) ---"
curl -s -X POST "http://127.0.0.1:8092/exam/api/exam/exam/export-raw-data" \
  -H "Authorization: Bearer $TOK" -H "Content-Type: application/x-www-form-urlencoded" \
  -d "examId=9999999999999999999" | head -c 500
echo
echo

echo "=== B. Normal closed MBTI exam export still works ==="
EID=1776746714029306200
out=/tmp/exp-norm.xlsx
curl -s -o $out -w "HTTP %{http_code} size=%{size_download}\n" \
  -X POST "http://127.0.0.1:8092/exam/api/exam/exam/export-raw-data" \
  -H "Authorization: Bearer $TOK" -H "Content-Type: application/x-www-form-urlencoded" \
  -d "examId=$EID"
python3 - <<'PY'
import zipfile,re
z=zipfile.ZipFile('/tmp/exp-norm.xlsx')
sh=z.read('xl/worksheets/sheet1.xml').decode('utf-8',errors='ignore')
rows=re.findall(r'<row r="(\d+)"', sh)
print('xlsx_rows', len(rows))
PY
rm -f $out

echo
echo "=== C. Submit fallback path: try a real seed (paperId 1777217353411783509) ==="
mysql -upositive_app -p'Positive#2026App!' element -e "UPDATE el_paper SET state=0, user_time=0 WHERE id='1777217353411783509'; UPDATE el_candidate SET end_time=NULL, pdf_flag=0 WHERE paper_id='1777217353411783509';" 2>&1 | grep -v Warning
mysql -upositive_app -p'Positive#2026App!' element -e "DELETE FROM el_mbti_answer WHERE paper_id='1777217353411783509'; SET @row=0; INSERT INTO el_mbti_answer (id, paper_id, qu_id, score_a, score_b, answered, create_time) SELECT CONCAT('SEED', LPAD(@row := @row + 1, 4, '0')), '1777217353411783509', qr.qu_id, 3, 2, 1, NOW() FROM (SELECT qr.qu_id FROM el_qu_repo qr JOIN el_exam_repo er ON er.repo_id=qr.repo_id JOIN el_paper p ON p.exam_id=er.exam_id WHERE p.id='1777217353411783509' ORDER BY qr.sort LIMIT 48) qr;" 2>&1 | grep -v Warning
echo "--- before submit ---"
mysql -upositive_app -p'Positive#2026App!' element -t -e "SELECT 'paper' tbl,state,user_time FROM el_paper WHERE id='1777217353411783509'; SELECT 'cand' tbl,name,end_time FROM el_candidate WHERE paper_id='1777217353411783509';" 2>&1 | grep -v Warning
curl -s -X POST http://127.0.0.1:8092/exam/api/mbti/submit -H "Authorization: Bearer $TOK" -H "Content-Type: application/json" -d "{\"paperId\":\"1777217353411783509\"}"
echo
echo "--- after submit ---"
mysql -upositive_app -p'Positive#2026App!' element -t -e "SELECT 'paper' tbl,state,user_time FROM el_paper WHERE id='1777217353411783509'; SELECT 'cand' tbl,name,end_time,pdf_flag FROM el_candidate WHERE paper_id='1777217353411783509';" 2>&1 | grep -v Warning

echo
echo "=== D. recent server logs ==="
journalctl -u talent-assessment --since '5 min ago' --no-pager | grep -iE 'mbti.submit|update el_tester|export:|panic|error' | tail -20 || true
