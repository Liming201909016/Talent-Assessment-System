#!/bin/bash
# 逐页对比两份 PDF 的视觉差异
# 用法: bash compare-pdfs.sh /tmp/baseline.pdf /tmp/new.pdf
set -e
B=$1
N=$2
OUT=/tmp/pdfdiff
rm -rf $OUT && mkdir -p $OUT
echo "=== 渲染 baseline ==="
pdftoppm -r 100 "$B" $OUT/base -png
BCNT=$(ls $OUT/base-*.png | wc -l)
echo "  $BCNT 页"
echo "=== 渲染 new ==="
pdftoppm -r 100 "$N" $OUT/new -png
NCNT=$(ls $OUT/new-*.png | wc -l)
echo "  $NCNT 页"

echo
echo "=== 逐页统计 ==="
echo "页号 | baseline 尺寸 | new 尺寸 | 像素差异"
echo "-----+---------------+----------+----------"
MAX=$(( BCNT > NCNT ? BCNT : NCNT ))
for i in $(seq 1 $MAX); do
  pi=$(printf "%02d" $i)
  bf=$OUT/base-$pi.png
  # new 用 1 起，无 0 填充（pdftoppm -r 100 9页时是 new-1..new-9）
  nf=$OUT/new-$i.png
  [ ! -f "$nf" ] && nf=$OUT/new-$pi.png
  bsize="-"
  nsize="-"
  diff="-"
  [ -f "$bf" ] && bsize=$(identify -format "%wx%h" "$bf" 2>/dev/null || echo "?")
  [ -f "$nf" ] && nsize=$(identify -format "%wx%h" "$nf" 2>/dev/null || echo "?")
  if [ -f "$bf" ] && [ -f "$nf" ]; then
    # 统一尺寸到 baseline 的尺寸再 diff（解决 chromedp/jsPDF 1px 渲染差异）
    convert "$nf" -resize "${bsize}!" $OUT/new-resized-$pi.png 2>/dev/null
    if [ -f "$OUT/new-resized-$pi.png" ]; then
      diff=$(compare -metric AE -fuzz 5% "$bf" "$OUT/new-resized-$pi.png" $OUT/diff-$pi.png 2>&1 || true)
      total_pixels=$(echo "$bsize" | awk -Fx '{print $1*$2}')
      pct=$(awk -v d="$diff" -v t="$total_pixels" 'BEGIN{printf "%.1f%%", (d/t)*100}' 2>/dev/null || echo "?")
      diff="$diff ($pct)"
    fi
  fi
  printf "%4d | %-13s | %-8s | %s\n" $i "$bsize" "$nsize" "$diff"
done
echo
echo "diff PNG 已保存到 $OUT/diff-*.png （红色=差异）"
ls $OUT/diff-*.png 2>/dev/null
