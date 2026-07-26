#!/bin/bash
# 一次性清理所有 MBTI 旧简版（让下次下载触发重生成）
# 客户机 39.106.61.48 专用
echo "=== 找到的旧简版文件 ==="
find /opt/talent-assessment/tmp/uploadPath -name "*_simple.*" -type f | wc -l
echo
echo "=== 列出（前20）==="
find /opt/talent-assessment/tmp/uploadPath -name "*_simple.*" -type f | head -20
echo
read -p "确认删除全部旧简版？输入 YES 继续: " CONFIRM
if [ "$CONFIRM" = "YES" ]; then
  find /opt/talent-assessment/tmp/uploadPath -name "*_simple.*" -type f -delete
  echo "已删除"
else
  echo "取消"
fi
