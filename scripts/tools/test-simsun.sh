#!/bin/bash
# 把 produced docx 的 value rPr 字体改为 SimSun，看是否解决豆腐
rm -rf /tmp/test-simsun; mkdir /tmp/test-simsun
cp /opt/talent-assessment/tmp/uploadPath/20260519/Liming_INFP_20260519161941.docx /tmp/test-simsun-src.docx 2>/dev/null || \
  cp /tmp/Liming_INFP_*.docx /tmp/test-simsun-src.docx 2>/dev/null
ls /tmp/test-simsun-src.docx
cd /tmp/test-simsun && unzip -q /tmp/test-simsun-src.docx

python3 <<'PY'
import re
with open('/tmp/test-simsun/word/document.xml','r',encoding='utf-8') as f:
  c = f.read()
# 找含 value 文本的 <w:r> (Liming/22/男/13520602871/dateStr) — 把 rPr 内 <w:rFonts ...> 改为 SimSun
needles = ['Liming', '22', '男', '13520602871', '2026年5月19日']
reRun = re.compile(r'<w:r[^>]*>[\s\S]*?</w:r>')
def fix(run):
  ok = False
  for n in needles:
    if f'>{n}<' in run or f'preserve">{n}<' in run:
      ok = True; break
  if not ok:
    return run
  # 替换 rFonts attr
  run2 = re.sub(r'<w:rFonts[^/]*/>',
                '<w:rFonts w:hint="eastAsia" w:ascii="SimSun" w:hAnsi="SimSun" w:eastAsia="SimSun"/>',
                run)
  return run2
new_c = reRun.sub(fix, c)
with open('/tmp/test-simsun/word/document.xml','w',encoding='utf-8') as f:
  f.write(new_c)
print("rPr fonts switched to SimSun in value runs")
PY

cd /tmp/test-simsun && zip -r -q /tmp/test-simsun.docx . && cd /tmp
rm -f /tmp/test-simsun.pdf
libreoffice --headless --convert-to pdf --outdir /tmp /tmp/test-simsun.docx 2>&1 | tail -3
ls -la /tmp/test-simsun.pdf
pdftotext -layout -f 1 -l 1 /tmp/test-simsun.pdf - 2>/dev/null | head -15