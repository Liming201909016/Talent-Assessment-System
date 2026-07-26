#!/bin/bash
CAP=$(curl -s http://127.0.0.1:8092/captchaImage)
UUID=$(echo "$CAP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('uuid',''))")
CODE=$(redis-cli -n 1 GET "captcha_codes:$UUID" | tr -d '"')
TOK=$(curl -s -X POST http://127.0.0.1:8092/login -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"cp@1234\",\"code\":\"$CODE\",\"uuid\":\"$UUID\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")
echo "tok=${#TOK}"
echo "=== mbti/submit ==="
curl -s -X POST http://127.0.0.1:8092/exam/api/mbti/submit \
  -H "Authorization: Bearer $TOK" -H "Content-Type: application/json" \
  -d '{"paperId":"1777217353411783509"}'
echo
echo "=== verify DB ==="
mysql -upositive_app -p'Positive#2026App!' -t element <<EOF
SELECT state, user_time FROM el_paper WHERE id='1777217353411783509';
SELECT name, end_time, pdf_flag, update_time FROM el_candidate WHERE paper_id='1777217353411783509';
EOF
