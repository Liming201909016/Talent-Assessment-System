#!/bin/bash
rm -rf /tmp/w14work; mkdir /tmp/w14work
cd /tmp/w14work && unzip -q /tmp/diag-replaced.docx
python3 <<'PY'
import re
with open('/tmp/w14work/word/document.xml','r',encoding='utf-8') as f:
  c = f.read()
# 只移除 props3d (其他保留)
c = re.sub(r'<w14:props3d[^/]*/>', '', c)
c = re.sub(r'<w14:props3d[^>]*>.*?</w14:props3d>', '', c, flags=re.S)
with open('/tmp/w14work/word/document.xml','w',encoding='utf-8') as f:
  f.write(c)
print("only props3d removed")
PY
cd /tmp/w14work && zip -r -q /tmp/diag-no-w14.docx . -x ".*" && cd /tmp
ls -la /tmp/diag-no-w14.docx
rm -f /tmp/diag-no-w14.pdf
libreoffice --headless --convert-to pdf --outdir /tmp /tmp/diag-no-w14.docx 2>&1 | tail -3
ls -la /tmp/diag-no-w14.pdf 2>/dev/null
echo "=== pdftotext ==="
pdftotext -layout -f 1 -l 1 /tmp/diag-no-w14.pdf - 2>/dev/null | head -15
