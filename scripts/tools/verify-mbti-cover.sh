#!/bin/bash
set -e
CAP=$(curl -s http://127.0.0.1:8092/captchaImage)
UUID=$(echo "$CAP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('uuid',''))")
CODE=$(redis-cli -n 1 GET "captcha_codes:$UUID" | tr -d '"')
echo "captcha uuid=$UUID code=$CODE"
TOK=$(curl -s -X POST http://127.0.0.1:8092/login -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"cp@1234\",\"code\":\"$CODE\",\"uuid\":\"$UUID\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")
echo "token len=${#TOK}"
echo "=== generate-report ==="
RESP=$(curl -s -X POST http://127.0.0.1:8092/exam/api/mbti/generate-report \
  -H "Authorization: Bearer $TOK" -H "Content-Type: application/json" \
  -d "{\"paperId\":\"1777217353411783509\",\"type\":\"full\",\"force\":true}")
echo "$RESP"
PATH_=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('path',''))")
echo "pdf path=$PATH_"
if [ -n "$PATH_" ] && [ -f "$PATH_" ]; then
  echo "=== pdf size ==="
  ls -la "$PATH_"
  echo "=== cover text (page 1) ==="
  pdftotext -layout -f 1 -l 1 "$PATH_" - | head -30
fi
