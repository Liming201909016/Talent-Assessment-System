#!/usr/bin/env bash
set -euo pipefail

echo "PUBLIC_HEALTH=$(curl -fsS http://127.0.0.1/prod-api/health)"
echo "LOGIN_HTTP=$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1/)"
temp_exams=$(sudo -n mysql element -Nse "SELECT COUNT(*) FROM el_exam WHERE title='AUTO-COMPETENCY-STAGING-VERIFY';")
temp_associations=$(sudo -n mysql element -Nse "SELECT COUNT(*) FROM el_exam_competency_dimension ecd LEFT JOIN el_exam e ON e.id=ecd.exam_id WHERE e.id IS NULL;")
dimensions=$(sudo -n mysql element -Nse "SELECT COUNT(*) FROM el_competency_dimension;")
dispatch_columns=$(sudo -n mysql element -Nse "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='element' AND table_name='el_exam' AND column_name IN ('assessment_type','scoring_mode','competency_report_audience','publish_status','published_at','published_by');")
echo "TEMP_EXAMS=$temp_exams"
echo "TEMP_ASSOCIATIONS=$temp_associations"
echo "DIMENSIONS=$dimensions"
echo "DISPATCH_COLUMNS=$dispatch_columns"
echo "RECENT_ERRORS_BEGIN"
sudo -n journalctl -u talent-assessment --since '2026-07-24 20:23:00' --no-pager | grep -Ei 'panic|fatal|unknown column|table .* doesn.t exist|error' || true
echo "RECENT_ERRORS_END"
