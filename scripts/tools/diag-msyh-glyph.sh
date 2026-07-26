#!/bin/bash
echo "=== MSYH.TTC info ==="
ls -la /usr/share/fonts/truetype/msfonts/MSYH.TTC
file /usr/share/fonts/truetype/msfonts/MSYH.TTC
md5sum /usr/share/fonts/truetype/msfonts/MSYH.TTC

echo
echo "=== Python fontTools test '懂' in Microsoft YaHei ==="
python3 <<'PY' 2>&1
try:
  from fontTools.ttLib import TTCollection, TTFont
  ttc = TTCollection('/usr/share/fonts/truetype/msfonts/MSYH.TTC')
  for i, f in enumerate(ttc.fonts):
    cmap = f.getBestCmap()
    name = f['name'].getDebugName(1)
    test = {ch: (ord(ch) in cmap) for ch in '懂女0691259年月日'}
    print(f'font[{i}] name={name!r} chars={test}')
except ImportError:
  print('fontTools not installed, try pip3 install fontTools')
except Exception as e:
  print(f'err: {e}')
PY

echo
echo "=== 在 39 上独立用 LibreOffice 转换一个 docx 看效果 ==="
# 用 server 内置的 chromium 渲染一段 html 看是否同样 □
cat > /tmp/test-render.html <<HTML
<html><head><meta charset="utf-8"><style>
body{font-family:"Microsoft YaHei","微软雅黑",sans-serif;font-size:30px;color:#75BD42}
</style></head><body>
姓名：懂<br>
年龄：30<br>
性别：女<br>
联系方式：12345678992<br>
报告日期：2026 年 5 月 19 日
</body></html>
HTML

google-chrome --no-sandbox --headless --disable-gpu --print-to-pdf=/tmp/test-render.pdf /tmp/test-render.html 2>&1 | tail -3
echo "test-render.pdf size: $(stat -c%s /tmp/test-render.pdf 2>/dev/null)"
pdftotext /tmp/test-render.pdf - | head -10

echo
echo "=== 测试：直接拿模板用 libreoffice 转 PDF 看封面（不经过 server 的 replace） ==="
cp /opt/talent-assessment/mbti-templates/MBTI-INFP.docx /tmp/test-tpl.docx
libreoffice --headless --convert-to pdf --outdir /tmp /tmp/test-tpl.docx 2>&1 | tail -3
ls -la /tmp/test-tpl.pdf 2>/dev/null
pdftotext -layout -f 1 -l 1 /tmp/test-tpl.pdf - 2>/dev/null | head -15
