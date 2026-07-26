#!/bin/bash
# 用 inotifywait 抓取生成的 docx
OUTDIR=/opt/talent-assessment/tmp/uploadPath/$(date +%Y%m%d)
mkdir -p "$OUTDIR"
echo "watching $OUTDIR"

# 后台启动监听
( inotifywait -m -e create --format '%f' "$OUTDIR" 2>/dev/null | while read fn; do
    if [[ "$fn" == *.docx ]]; then
      cp "$OUTDIR/$fn" /tmp/captured.docx
      echo "captured: $fn"
      exit 0
    fi
  done
) &
WATCHER=$!
sleep 1

# 触发 generate-report
CAP=$(curl -s http://127.0.0.1:8092/captchaImage)
UUID=$(echo "$CAP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('uuid',''))")
CODE=$(redis-cli -n 1 GET "captcha_codes:$UUID" 2>/dev/null | tr -d '"')
TOK=$(curl -s -X POST http://127.0.0.1:8092/login -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"cp@1234\",\"code\":\"$CODE\",\"uuid\":\"$UUID\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")
PID=$(mysql -uroot -pYFkdkiP3r1RWbsWz element -Nse "SELECT c.paper_id FROM el_candidate c JOIN el_exam e ON c.exam_id=e.id JOIN el_exam_repo er ON er.exam_id=e.id JOIN el_repo r ON er.repo_id=r.id WHERE r.code LIKE '003%' AND c.paper_id IS NOT NULL AND c.name IS NOT NULL AND EXISTS(SELECT 1 FROM el_mbti_answer WHERE paper_id=c.paper_id) ORDER BY c.id DESC LIMIT 1" 2>/dev/null)
echo "paperId=$PID"
curl -s -X POST http://127.0.0.1:8092/exam/api/mbti/generate-report \
  -H "Authorization: Bearer $TOK" -H "Content-Type: application/json" \
  -d "{\"paperId\":\"$PID\",\"type\":\"full\",\"force\":true}" > /tmp/genresp.json
cat /tmp/genresp.json
echo

# 等待最多 5s 让 watcher 抓
sleep 5
kill $WATCHER 2>/dev/null

ls -la /tmp/captured.docx 2>/dev/null
if [ -f /tmp/captured.docx ]; then
  echo "=== document.xml around 姓名 ==="
  unzip -p /tmp/captured.docx word/document.xml > /tmp/captured-doc.xml
  python3 <<'PY'
import re
with open('/tmp/captured-doc.xml','r',encoding='utf-8') as f:
  c = f.read()
for label in ['姓名：','年龄：','报告日期：']:
  idx = c.find(label+'</w:t>')
  if idx < 0:
    # try cross-run
    pass
  if idx < 0: continue
  # 截取该 label 末尾后 800 字符
  segment = c[idx:idx+800]
  print(f'--- {label} segment ---')
  print(segment[:600])
  print()
PY
fi
