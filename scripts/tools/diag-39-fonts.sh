#!/bin/bash
# 39 生产 MBTI PDF 字体诊断
DSN_USER=$(grep -oE 'dsn:.*"[^"]+"' /opt/talent-assessment/configs/application*.yml 2>/dev/null | head -1 | grep -oE '"[^:]+:[^@]+@' | tr -d '"@')
echo "DSN user piece: $DSN_USER"
DBUSER=$(echo "$DSN_USER" | cut -d: -f1)
DBPWD=$(echo "$DSN_USER" | cut -d: -f2)
echo "DB user: $DBUSER"

CAP=$(curl -s http://127.0.0.1:8092/captchaImage)
UUID=$(echo "$CAP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('uuid',''))")
CODE=$(redis-cli -n 1 GET "captcha_codes:$UUID" 2>/dev/null | tr -d '"')
TOK=$(curl -s -X POST http://127.0.0.1:8092/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"cp@1234\",\"code\":\"$CODE\",\"uuid\":\"$UUID\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")
echo "tok_len=${#TOK}"

PID=$(mysql -u"$DBUSER" -p"$DBPWD" element -Nse "
SELECT c.paper_id FROM el_candidate c
JOIN el_exam e ON c.exam_id=e.id
JOIN el_exam_repo er ON er.exam_id=e.id
JOIN el_repo r ON er.repo_id=r.id
WHERE r.code LIKE '003%' AND c.paper_id IS NOT NULL AND c.name IS NOT NULL
  AND EXISTS (SELECT 1 FROM el_mbti_answer WHERE paper_id=c.paper_id)
ORDER BY c.id DESC LIMIT 1" 2>/dev/null)
echo "paperId=$PID"
if [ -z "$PID" ]; then echo "no paper, abort"; exit 1; fi

# 查这个 paper 对应的考生原始数据
echo "=== tester data ==="
mysql -u"$DBUSER" -p"$DBPWD" element -t -e "
SELECT name, age, gender, telephone, affiliation, post FROM el_candidate WHERE paper_id='$PID'" 2>&1 | grep -v Warning

echo
echo "=== generate report ==="
RESP=$(curl -s -X POST http://127.0.0.1:8092/exam/api/mbti/generate-report \
  -H "Authorization: Bearer $TOK" -H "Content-Type: application/json" \
  -d "{\"paperId\":\"$PID\",\"type\":\"full\",\"force\":true}")
echo "$RESP"
P=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('path',''))" 2>/dev/null)
echo "pdf=$P"
if [ -z "$P" ] || [ ! -f "$P" ]; then echo "pdf not found, abort"; exit 1; fi

echo
echo "=== pdffonts page 1 ==="
pdffonts "$P" 2>&1 | head -25

echo
echo "=== pdftotext cover ==="
pdftotext -layout -f 1 -l 1 "$P" - | head -20

echo
echo "=== font installation check ==="
fc-list :lang=zh | head -10
echo "---"
fc-match "FangSong"
fc-match "仿宋"
fc-match "SimSun"
fc-match "Microsoft YaHei"

echo
echo "=== fontconfig fallback conf ==="
ls -la /etc/fonts/conf.d/ | grep -iE 'cjk|fang' | head -5
if [ -f /etc/fonts/conf.d/99-cjk-fallback.conf ]; then
  grep -A1 'FangSong\|仿宋' /etc/fonts/conf.d/99-cjk-fallback.conf | head -10
fi

# 拷贝 PDF 出来给本地看
cp "$P" /tmp/diag-39.pdf
echo "PDF saved to /tmp/diag-39.pdf"
