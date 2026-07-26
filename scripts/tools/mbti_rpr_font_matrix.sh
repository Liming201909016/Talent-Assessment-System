#!/usr/bin/env bash
set -euo pipefail
SRC="/opt/talent-assessment/tmp/uploadPath/20260519/Liming_INFP_20260519180922.docx"
VALUES=("Liming" "22" "男" "13520602871" "2026年5月19日")
ITEMS=(
  "msyh|Microsoft YaHei"
  "noto|Noto Sans CJK SC"
  "simsun|SimSun"
  "simhei|SimHei"
  "wqy|WenQuanYi Micro Hei"
)
for item in "${ITEMS[@]}"; do
  name="${item%%|*}"
  font="${item#*|}"
  out="/tmp/mbti-rpr-${name}"
  rm -f "${out}.docx" "${out}.pdf" "${out}-1.png"
  python3 /tmp/mbti_min_rpr_experiment.py "$SRC" "${out}.docx" "$font" "${VALUES[@]}" >"${out}.log"
  libreoffice --headless --convert-to pdf --outdir /tmp "${out}.docx" >"${out}-lo.log" 2>&1
  pdftoppm -r 100 -f 1 -l 1 "${out}.pdf" "$out" -png
  echo "=== ${name} / ${font} ==="
  cat "${out}.log"
  pdftotext -layout "${out}.pdf" - | grep -E '姓名|年龄|性别|联系方式|报告日期' -A1
  pdffonts "${out}.pdf" | grep -E 'YaHei|Noto|SimSun|SimHei|WenQuan' | head -8 || true
done
ls -lh /tmp/mbti-rpr-*-1.png
