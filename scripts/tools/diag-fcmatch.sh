#!/bin/bash
echo "=== fc-match 微软雅黑 ==="
fc-match "微软雅黑"
fc-match "微软雅黑:lang=zh"
fc-match "Microsoft YaHei"
fc-match "Microsoft YaHei:lang=zh"

echo
echo "=== 微软雅黑 字体文件实际位置 ==="
fc-list :family="Microsoft YaHei" | head -5
fc-list | grep -i 'yahei' | head -5
echo
echo "=== 模板里 '懂' 字测试 - 用微软雅黑能否渲染 ==="
python3 <<'PY'
import subprocess
# 写一个简单 html 用 chromium-headless 或 PIL 测，不可靠。直接用 fc-match 看实际匹配
for ch in ['懂','女','0','6','9','1','2','5']:
  out = subprocess.run(['fc-match', '-s', f'Microsoft YaHei:charset={hex(ord(ch))[2:]}'],
                       capture_output=True, text=True, timeout=5)
  first = out.stdout.split('\n')[0]
  print(f'{ch!r} → {first}')
PY

echo
echo "=== pdfinfo on diag-39.pdf ==="
pdfinfo /tmp/diag-39.pdf | head -20

echo
echo "=== fc-match -s（首选+回退）'懂' 时的字体链 ==="
fc-match -s "Microsoft YaHei" 2>&1 | head -10
