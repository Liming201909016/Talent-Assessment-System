#!/usr/bin/env bash
set -euo pipefail

BACKUP=${1:?backup directory required}
EXPECTED_BACKEND=${2:?backend sha256 required}
EXPECTED_INDEX=${3:?frontend index sha256 required}

sudo gzip -t "$BACKUP/element.sql.gz"
current_qu=$(sudo mysql element -Nse "SELECT * FROM el_qu WHERE dimension_id IS NULL AND competency_question_type IS NULL ORDER BY id" | sha256sum | awk '{print $1}')
current_repo=$(sudo mysql element -Nse "SELECT * FROM el_qu_repo ORDER BY id" | sha256sum | awk '{print $1}')
current_answer=$(sudo mysql element -Nse "SELECT * FROM el_qu_answer ORDER BY id" | sha256sum | awk '{print $1}')
test "$current_qu" = "$(sudo cat "$BACKUP/traditional-qu.sha256")"
test "$current_repo" = "$(sudo cat "$BACKUP/traditional-qu-repo.sha256")"
test "$current_answer" = "$(sudo cat "$BACKUP/traditional-qu-answer.sha256")"
echo TRADITIONAL_SIGNATURES=UNCHANGED

test "$(sha256sum /opt/talent-assessment/server | awk '{print $1}')" = "$EXPECTED_BACKEND"
test "$(sha256sum /opt/talent-assessment/dist/index.html | awk '{print $1}')" = "$EXPECTED_INDEX"
echo DEPLOY_HASHES=MATCH

sudo mysql element -Nse "
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_competency_dimension),'|',
  (SELECT COUNT(*) FROM el_qu WHERE competency_question_type='dimension'),'|',
  (SELECT COUNT(*) FROM el_qu WHERE competency_question_type='validity'),'|',
  (SELECT COUNT(*) FROM el_exam WHERE assessment_type='competency'),'|',
  (SELECT COUNT(*) FROM el_competency_result),'|',
  (SELECT COUNT(*) FROM el_competency_group_result),'|',
  (SELECT COUNT(*) FROM el_competency_validity_result));
SELECT CONCAT(COLUMN_TYPE,'|',IS_NULLABLE)
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='overall_score';
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_exam_competency_group g LEFT JOIN el_exam e ON e.id=g.exam_id WHERE e.id IS NULL),'|',
  (SELECT COUNT(*) FROM el_competency_group_result r LEFT JOIN el_paper p ON p.id=r.paper_id WHERE p.id IS NULL),'|',
  (SELECT COUNT(*) FROM el_competency_validity_result r LEFT JOIN el_paper p ON p.id=r.paper_id WHERE p.id IS NULL));
"

printf 'SERVICES='
systemctl is-active talent-assessment nginx mysql | paste -sd, -
printf 'INTERNAL_HEALTH='
curl -fsS http://127.0.0.1:8092/health
printf '\nPUBLIC_HEALTH='
curl -fsS http://20.200.136.133/prod-api/health
printf '\nTEMP_REDIS_SESSIONS='
redis-cli -n 1 --scan --pattern 'login_tokens:phase1-runtime-*' | wc -l
printf 'JOURNAL_ERRORS='
sudo journalctl -u talent-assessment --since '2026-08-11 09:34:00' --no-pager | grep -Eci 'panic|level=ERROR|fatal|unknown column|missing table|segmentation' || true
