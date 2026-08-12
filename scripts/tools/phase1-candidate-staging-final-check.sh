#!/usr/bin/env bash
set -euo pipefail

BACKUP=${1:?backup directory required}
EXPECTED_CONTENT_SHA=${2:?content source sha256 required}
EXPECTED_INDEX_SHA=${3:?frontend index sha256 required}
APP=/opt/talent-assessment

sudo gzip -t "$BACKUP/element.sql.gz"
sudo sha256sum -c "$BACKUP/SHA256SUMS"
[ "$(sudo stat -c '%a' "$BACKUP")" = "700" ]
while IFS= read -r mode; do
  [ "$mode" = "600" ]
done < <(sudo find "$BACKUP" -mindepth 1 -maxdepth 1 -type f -printf '%m\n')

test "$(sudo mysql element -Nse "SELECT * FROM el_qu WHERE dimension_id IS NULL AND competency_question_type IS NULL ORDER BY id" | sha256sum | cut -d' ' -f1)" = "$(sudo cat "$BACKUP/traditional-qu.sha256")"
test "$(sudo mysql element -Nse "SELECT * FROM el_qu_repo ORDER BY id" | sha256sum | cut -d' ' -f1)" = "$(sudo cat "$BACKUP/traditional-qu-repo.sha256")"
test "$(sudo mysql element -Nse "SELECT * FROM el_qu_answer ORDER BY id" | sha256sum | cut -d' ' -f1)" = "$(sudo cat "$BACKUP/traditional-qu-answer.sha256")"
echo TRADITIONAL_SIGNATURES_UNCHANGED

STATE=$(sudo mysql element -Nse "
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_competency_dimension),'|',
  (SELECT COUNT(*) FROM el_qu WHERE competency_question_type='dimension'),'|',
  (SELECT COUNT(*) FROM el_qu WHERE competency_question_type='validity'),'|',
  (SELECT COUNT(*) FROM el_competency_report_text WHERE content_version='competency-phase1-content-v1'),'|',
  (SELECT SUM(content_type='template') FROM el_competency_report_text WHERE content_version='competency-phase1-content-v1'),'|',
  (SELECT SUM(is_temporary=1 AND status=1) FROM el_competency_report_text WHERE content_version='competency-phase1-content-v1'),'|',
  (SELECT COUNT(*) FROM el_competency_report_content_package WHERE id='phase1-candidate-draft-v1' AND approval_status='draft' AND content_approved_at IS NULL AND psychometric_approved_at IS NULL AND effective_environment='' AND disclaimer<>'' AND content_source_sha256='$EXPECTED_CONTENT_SHA'),'|',
  (SELECT COUNT(*) FROM el_exam WHERE assessment_type='competency'),'|',
  (SELECT COUNT(*) FROM el_competency_result),'|',
  (SELECT COUNT(*) FROM el_competency_report),'|',
  (SELECT COUNT(*) FROM el_competency_report_audit))")
[ "$STATE" = "10|80|10|66|7|66|1|0|0|0|0" ]
echo "CANDIDATE_STATE=$STATE"

[ "$(sha256sum "$APP/dist/index.html" | cut -d' ' -f1)" = "$EXPECTED_INDEX_SHA" ]
[ "$(systemctl is-active talent-assessment)" = "active" ]
[ "$(systemctl is-active nginx)" = "active" ]
[ "$(systemctl is-active mysql)" = "active" ]
[ "$(curl -fsS http://127.0.0.1:8092/health)" = '{"status":"ok"}' ]

SESSION_COUNT=$(redis-cli -n 1 --scan --pattern 'login_tokens:phase1-*' | wc -l)
[ "$SESSION_COUNT" = "0" ]
ERROR_COUNT=$(sudo journalctl -u talent-assessment --since '2026-08-12 11:33:00' --no-pager | grep -Eci 'panic|level=ERROR|fatal|unknown column|missing table|segmentation' || true)
[ "$ERROR_COUNT" = "0" ]

rm -f /tmp/phase1-candidate-staging-backup.sh /tmp/phase1-candidate-import.sql /tmp/phase1-candidate-import-v2.sql
rm -f /tmp/staging-phase1-runtime-e2e.py /tmp/staging-phase1-runtime-negative-e2e.py
rm -f /tmp/phase1-report-fixed-text-dist.tar.gz /tmp/staging-phase1-frontend-deploy.sh
rm -rf /tmp/phase1-report-fixed-text-dist

echo PHASE1_CANDIDATE_STAGING_FINAL_CHECK_PASS
echo "frontend_index_sha256=$EXPECTED_INDEX_SHA"
echo "services=active,active,active|sessions=0|journal_errors=0"
