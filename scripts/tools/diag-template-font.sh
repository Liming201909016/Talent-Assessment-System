#!/bin/bash
# 看 INFP 模板里 value 占位符的字体设置
T=/opt/talent-assessment/configs/export-templates/MBTI-INFP.docx
[ ! -f "$T" ] && T=$(find /opt/talent-assessment -name 'MBTI-INFP.docx' | head -1)
echo "template: $T"
unzip -p "$T" word/document.xml > /tmp/doc.xml
echo "doc.xml size: $(wc -c < /tmp/doc.xml)"

# 找到"姓名：" 后面的紧邻 <w:t> 占位符 run，看它的 <w:rPr>
python3 <<'PY'
import re
with open('/tmp/doc.xml','r',encoding='utf-8') as f:
  c = f.read()

# 找 "姓名：" 出现的所有位置
for label in ['姓名：','年龄：','性别：','联系方式：','报告日期：']:
  # 找 label 后第一个完整的 <w:r>...</w:r> 节
  idx = c.find(label + '</w:t>')
  if idx < 0:
    print(f'[{label}] NOT FOUND directly')
    continue
  # 看 label 之前的 <w:r>...<w:t> 里的 <w:rPr>
  # 找 label 后的下一个 <w:r ...>...<w:rPr>...</w:rPr>...<w:t>占位</w:t>
  tail = c[idx+len(label)+len('</w:t>'):idx+len(label)+1500]
  m = re.search(r'<w:r[^>]*>(.*?)<w:t[^>]*>([^<]*)</w:t>', tail, re.S)
  if m:
    rpr = m.group(1)
    placeholder = m.group(2)
    # 提取 rFonts
    fontm = re.search(r'<w:rFonts[^/]*/>', rpr)
    print(f'[{label}] placeholder="{placeholder!r}"')
    print(f'   rPr: {rpr[:300]}')
    print(f'   rFonts: {fontm.group(0) if fontm else "(none)"}')
  else:
    print(f'[{label}] no following run found')
  print()
PY
