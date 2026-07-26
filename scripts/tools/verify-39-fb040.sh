#!/bin/bash
echo "=== B1: isStu 限定 001 ==="
grep -roE 'isStu[^,;}]{1,80}' /opt/talent-assessment/dist/static/js/ 2>/dev/null | grep -i '001' | head -3
echo
echo "=== B1: affiliationLabel computed ==="
grep -roE 'affiliationLabel' /opt/talent-assessment/dist/static/js/ 2>/dev/null | head -3
echo
echo "=== B2: MBTI 强制答完文案 ==="
grep -ro '请全部作答完毕后再交卷' /opt/talent-assessment/dist/static/js/ 2>/dev/null | head -3
grep -ro '定位到未作答' /opt/talent-assessment/dist/static/js/ 2>/dev/null | head -3
echo
echo "=== B3: locateLabelEnd 在二进制 ==="
grep -a -c locateLabelEnd /opt/talent-assessment/server
echo
echo "=== B4: tester list SQL 含 id desc tie-breaker ==="
# 间接验证：找一个 exam 实际调 list API 看返回顺序是否稳定
echo "(see B4 API test below)"
echo
echo "=== captcha + login ==="
CAP=$(curl -s http://127.0.0.1:8092/captchaImage)
UUID=$(echo "$CAP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('uuid',''))")
CODE=$(redis-cli -n 1 GET "captcha_codes:$UUID" 2>/dev/null | tr -d '"')
TOK=$(curl -s -X POST http://127.0.0.1:8092/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"cp@1234\",\"code\":\"$CODE\",\"uuid\":\"$UUID\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")
echo "tok_len=${#TOK}"
if [ "${#TOK}" -lt 100 ]; then echo "LOGIN FAILED, abort"; exit 1; fi

echo
echo "=== B3 e2e: 生成 MBTI 报告（找 003 最近完成的 paper） ==="
PID=$(mysql -upositive_app -p"Positive#2026App!" element -Nse "SELECT c.paper_id FROM el_candidate c JOIN el_exam e ON c.exam_id=e.id JOIN el_exam_repo er ON er.exam_id=e.id JOIN el_repo r ON er.repo_id=r.id WHERE r.code LIKE '003%' AND c.paper_id IS NOT NULL AND c.name IS NOT NULL AND c.end_time IS NOT NULL ORDER BY c.id DESC LIMIT 1" 2>/dev/null)
echo "paperId=$PID"
if [ -n "$PID" ]; then
  RESP=$(curl -s -X POST http://127.0.0.1:8092/exam/api/mbti/generate-report \
    -H "Authorization: Bearer $TOK" -H "Content-Type: application/json" \
    -d "{\"paperId\":\"$PID\",\"type\":\"full\",\"force\":true}")
  echo "$RESP"
  P=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('path',''))" 2>/dev/null)
  if [ -n "$P" ] && [ -f "$P" ]; then
    echo "--- pdftotext cover ---"
    pdftotext -layout -f 1 -l 1 "$P" - | head -15
  fi
fi

echo
echo "=== B5: 查 003 paper.state=2 且 end_time NULL 的（FB-040 部署后应为 0） ==="
mysql -upositive_app -p"Positive#2026App!" element -t -e "
SELECT COUNT(*) AS stale_state2_no_endtime
FROM el_candidate c JOIN el_paper p ON p.id=c.paper_id
JOIN el_exam_repo er ON er.exam_id=c.exam_id
JOIN el_repo r ON er.repo_id=r.id
WHERE r.code LIKE '003%' AND p.state=2 AND c.end_time IS NULL;
" 2>&1 | grep -v Warning

echo
echo "=== B6: 封闭 MBTI 导出 ==="
EID=$(mysql -upositive_app -p"Positive#2026App!" element -Nse "
SELECT e.id FROM el_exam e
JOIN el_exam_repo er ON er.exam_id=e.id
JOIN el_repo r ON er.repo_id=r.id
WHERE e.is_open=2 AND r.code LIKE '003%'
  AND (SELECT COUNT(*) FROM el_tester WHERE exam_id=e.id AND (del_flag IS NULL OR del_flag='0'))>0
ORDER BY (SELECT COUNT(*) FROM el_tester WHERE exam_id=e.id) DESC LIMIT 1" 2>/dev/null)
echo "examId=$EID"
if [ -n "$EID" ]; then
  HTTP=$(curl -s -o /tmp/b6-39.xlsx -w "HTTP %{http_code} size=%{size_download}\n" \
    -X POST "http://127.0.0.1:8092/exam/api/exam/exam/export-raw-data" \
    -H "Authorization: Bearer $TOK" -H "Content-Type: application/x-www-form-urlencoded" \
    -d "examId=$EID")
  echo "$HTTP"
  python3 <<PY
import zipfile, re
z = zipfile.ZipFile("/tmp/b6-39.xlsx")
ss = re.findall(r"<t[^>]*>([^<]*)</t>", z.read("xl/sharedStrings.xml").decode() if "xl/sharedStrings.xml" in z.namelist() else "")
sh = z.read("xl/worksheets/sheet1.xml").decode()
rows = re.findall(r'<row r="(\d+)"', sh)
print("total rows:", len(rows))
m = re.search(r'<row r="3"[^>]*>(.*?)</row>', sh)
if m:
  cells = re.findall(r'<c r="([A-Z]+\d+)"(?: s="\d+")?(?: t="(\w+)")?(?:[^/>]*?/>|>(.*?)</c>)', m.group(1))
  out=[]
  for ref,t,body in cells[:14]:
    v = re.search(r"<v>([^<]*)</v>", body or "")
    val = v.group(1) if v else ""
    if t=="s" and val.isdigit() and int(val)<len(ss): val=ss[int(val)]
    out.append((re.sub(r"\d","",ref), val))
  print("row3:", out)
PY
  rm -f /tmp/b6-39.xlsx
fi

echo
echo "=== B4: 调 tester/list 看是否同秒分组 ==="
RESP=$(curl -s "http://127.0.0.1:8092/exam/api/tester/list?pageNum=1&pageSize=10" \
  -H "Authorization: Bearer $TOK")
echo "$RESP" | python3 -c "
import sys,json
d=json.load(sys.stdin)
rows=d.get('rows',[])
print('rows returned:', len(rows))
last_exam=None
for r in rows[:15]:
  print(r.get('createTime'),'|',r.get('examId'),'|',r.get('name'))
"
