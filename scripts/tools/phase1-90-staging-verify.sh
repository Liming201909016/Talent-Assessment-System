#!/usr/bin/env bash
set -euo pipefail

BACKUP=/opt/talent-assessment/backups/phase1_90_import_20260810_161855
sudo gzip -t "$BACKUP/element.sql.gz"

current_qu=$(sudo mysql element -Nse "SELECT * FROM el_qu WHERE dimension_id IS NULL AND competency_question_type IS NULL ORDER BY id" | sha256sum | awk '{print $1}')
current_repo=$(sudo mysql element -Nse "SELECT * FROM el_qu_repo ORDER BY id" | sha256sum | awk '{print $1}')
current_answer=$(sudo mysql element -Nse "SELECT * FROM el_qu_answer ORDER BY id" | sha256sum | awk '{print $1}')
test "$current_qu" = "$(sudo cat "$BACKUP/traditional-qu.sha256")"
test "$current_repo" = "$(sudo cat "$BACKUP/traditional-qu-repo.sha256")"
test "$current_answer" = "$(sudo cat "$BACKUP/traditional-qu-answer.sha256")"
echo TRADITIONAL_SIGNATURES=UNCHANGED

sudo mysql element -Nse "
SELECT COUNT(*),COUNT(DISTINCT question_code),
       SUM(competency_question_type='dimension'),SUM(competency_question_type='validity'),
       SUM(competency_question_type='dimension' AND scoring_direction='forward'),
       SUM(competency_question_type='dimension' AND scoring_direction='reverse'),
       SUM(competency_question_type='validity' AND scoring_direction='forward'),
       SUM(question_status=0)
FROM el_qu
WHERE dimension_id IS NOT NULL OR competency_question_type IS NOT NULL;
SELECT COUNT(*) FROM el_exam WHERE assessment_type='competency';
SELECT COUNT(*) FROM el_qu_answer qa INNER JOIN el_qu q ON q.id=qa.qu_id WHERE q.competency_question_type IN ('dimension','validity');
SELECT COUNT(*) FROM el_qu_repo qr INNER JOIN el_qu q ON q.id=qr.qu_id WHERE q.competency_question_type IN ('dimension','validity');
"

printf 'SERVICES='
systemctl is-active talent-assessment nginx mysql | paste -sd, -
printf 'JOURNAL_ERRORS='
sudo journalctl -u talent-assessment --since '2026-08-10 16:20:00' --no-pager | grep -Eci 'panic|level=ERROR|fatal' || true
printf 'TEMP_PHASE1_SESSIONS='
redis-cli -n 1 --scan --pattern 'login_tokens:phase1-*' | wc -l

rm -f /tmp/server-phase1-import-linux /tmp/dist-phase1-import.tar.gz /tmp/dist-phase1-import-v2.tar.gz
rm -f /tmp/competency-phase1-import-20260810.xlsx /tmp/staging-phase1-90-import-e2e.py /tmp/phase1-90-staging-backup.sh
rm -rf /tmp/phase1-ui-extract /tmp/phase1-ui-v2
echo TEMP_ARTIFACTS_CLEANED=true
