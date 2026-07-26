#!/bin/bash
echo "=== fonts in document.xml ==="
grep -oE 'w:ascii="[^"]+"' /tmp/sim.xml | sort | uniq -c | sort -rn | head -10
echo
echo "=== fc-match 汉仪雅酷黑 75W ==="
fc-match "汉仪雅酷黑 75W"
fc-list | grep -iE 'hanyi|yaku|汉仪' | head -3
echo
echo "=== fc-match Liming run rPr font (Microsoft YaHei) ==="
# 联系方式 value run rPr 用的是 Microsoft YaHei
grep -c '微软雅黑' /tmp/sim.xml
echo
echo "=== which font is missing on linux ==="
fc-match -s "汉仪雅酷黑 75W" | head -5
