#!/bin/bash
# 用最简 rPr 替换所有 value run
rm -rf /tmp/w14work; mkdir /tmp/w14work
cd /tmp/w14work && unzip -q /tmp/diag-replaced.docx
python3 <<'PY'
import re
with open('/tmp/w14work/word/document.xml','r',encoding='utf-8') as f:
  c = f.read()

# 策略：替换 value 的 <w:rPr>...含 w14...</w:rPr> 为最简版本
# 简化为只含字体+颜色+大小
simple_rpr = '<w:rPr><w:rFonts w:hint="eastAsia" w:ascii="Microsoft YaHei" w:hAnsi="Microsoft YaHei" w:eastAsia="Microsoft YaHei"/><w:color w:val="75BD42"/><w:sz w:val="30"/><w:szCs w:val="30"/></w:rPr>'

# 只替换那些含 w14:textFill 或 w14:props3d 的 rPr
def repl(m):
  return simple_rpr if ('w14:textFill' in m.group(0) or 'w14:props3d' in m.group(0)) else m.group(0)

new_c = re.sub(r'<w:rPr>.*?</w:rPr>', repl, c, flags=re.S)
with open('/tmp/w14work/word/document.xml','w',encoding='utf-8') as f:
  f.write(new_c)
print('replaced', new_c.count(simple_rpr), 'rPr blocks')
PY

cd /tmp/w14work && zip -r -q /tmp/diag-simple-rpr.docx . && cd /tmp
ls -la /tmp/diag-simple-rpr.docx
rm -f /tmp/diag-simple-rpr.pdf
libreoffice --headless --convert-to pdf --outdir /tmp /tmp/diag-simple-rpr.docx 2>&1 | tail -3
ls -la /tmp/diag-simple-rpr.pdf 2>/dev/null
pdftotext -layout -f 1 -l 1 /tmp/diag-simple-rpr.pdf - 2>/dev/null | head -15