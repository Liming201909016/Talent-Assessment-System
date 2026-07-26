#!/usr/bin/env bash
set -euo pipefail

echo "HOST=$(hostname)"
if sudo -n true >/dev/null 2>&1; then
  echo "SUDO=OK"
else
  echo "SUDO=NO"
  exit 1
fi

echo "SERVICE=$(systemctl is-active talent-assessment 2>/dev/null || true)"
echo "NGINX=$(systemctl is-active nginx 2>/dev/null || true)"
echo "MYSQL_SERVICE=$(systemctl is-active mysql 2>/dev/null || systemctl is-active mariadb 2>/dev/null || true)"
mysql --version | head -1
df -h / /opt | tail -n +2

echo "DB_CHECK"
sudo -n mysql element -Nse "
SELECT CONCAT('table_count=', COUNT(*)) FROM information_schema.tables WHERE table_schema='element';
SELECT CONCAT('exam_count=', COUNT(*)) FROM el_exam;
SELECT CONCAT('dispatch_columns=', COUNT(*)) FROM information_schema.columns WHERE table_schema='element' AND table_name='el_exam' AND column_name IN ('assessment_type','scoring_mode','competency_report_audience','publish_status','published_at','published_by');
SELECT CONCAT('competency_tables=', COUNT(*)) FROM information_schema.tables WHERE table_schema='element' AND table_name IN ('el_competency_dimension','el_exam_competency_dimension');
"
