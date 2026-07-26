#!/bin/bash
CAP=$(curl -s http://127.0.0.1:8092/captchaImage)
UUID=$(echo "$CAP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('uuid',''))")
CODE=$(redis-cli -n 1 GET "captcha_codes:$UUID" | tr -d '"')
TOK=$(curl -s -X POST http://127.0.0.1:8092/login -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"cp@1234\",\"code\":\"$CODE\",\"uuid\":\"$UUID\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")
EXAM_ID='1776746714029306200'
echo "tok=${#TOK}"
echo "=== export-raw-data (closed MBTI, 12 testers) ==="
curl -s -o /tmp/closed-mbti.xlsx -w "HTTP %{http_code} size=%{size_download}\n" \
  -X POST "http://127.0.0.1:8092/exam/api/exam/exam/export-raw-data" \
  -H "Authorization: Bearer $TOK" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "examId=$EXAM_ID"
ls -la /tmp/closed-mbti.xlsx
file /tmp/closed-mbti.xlsx
# 转换 xlsx 看数据行数
python3 <<'PY'
import zipfile, re
try:
  z = zipfile.ZipFile('/tmp/closed-mbti.xlsx')
  for n in z.namelist():
    if 'sheet1.xml' in n:
      x = z.read(n).decode('utf-8', errors='ignore')
      rows = re.findall(r'<row r="(\d+)"', x)
      print('rows:', rows[:30], '... total:', len(rows))
      # 显示前 3 行内容
      cells = re.findall(r'<c r="([A-Z]+\d+)"[^>]*>(?:<v>([^<]*)</v>)?(?:<is><t[^>]*>([^<]*)</t></is>)?', x)
      from collections import defaultdict
      bymrow = defaultdict(dict)
      for ref, v, t in cells[:60]:
        r = re.sub(r'\D','',ref); col = re.sub(r'\d','',ref)
        bymrow[int(r)][col] = v or t
      for rk in sorted(bymrow)[:5]:
        print(f'row{rk}:', bymrow[rk])
except Exception as e:
  print('err:', e)
PY

