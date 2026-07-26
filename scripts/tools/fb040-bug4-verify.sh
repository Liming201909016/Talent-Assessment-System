#!/bin/bash
CAP=$(curl -s http://127.0.0.1:8092/captchaImage)
UUID=$(echo "$CAP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('uuid',''))")
CODE=$(redis-cli -n 1 GET "captcha_codes:$UUID" | tr -d '"')
TOK=$(curl -s -X POST http://127.0.0.1:8092/login -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"cp@1234\",\"code\":\"$CODE\",\"uuid\":\"$UUID\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")
echo "tok len=${#TOK}"
echo "=== tester/list ==="
curl -s "http://127.0.0.1:8092/exam/api/tester/list?pageNum=1&pageSize=20&idNumber=FB040" \
  -H "Authorization: Bearer $TOK" \
  | python3 -c "
import sys,json
d=json.load(sys.stdin)
rows=d.get('rows',[])
print('total:', d.get('total'))
for r in rows:
  print(r.get('id'), r.get('examId'), r.get('name'))
"
