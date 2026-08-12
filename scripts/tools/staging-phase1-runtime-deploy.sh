#!/usr/bin/env bash
set -euo pipefail

EXPECTED_BACKEND=${1:?expected backend sha256 required}
EXPECTED_INDEX=${2:?expected frontend index sha256 required}

sudo mysql element < /tmp/competency_008_phase1_structures.sql > /tmp/phase1-runtime-migration-1.log
sudo mysql element < /tmp/competency_008_phase1_structures.sql > /tmp/phase1-runtime-migration-2.log
echo MIGRATION_RERUN_OK
sudo mysql element -Nse "
SELECT CONCAT(COLUMN_TYPE,'|',IS_NULLABLE)
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA=DATABASE()
  AND TABLE_NAME='el_competency_result'
  AND COLUMN_NAME='overall_score';
SELECT CONCAT(
  (SELECT COUNT(*) FROM el_qu WHERE competency_question_type='dimension'),'|',
  (SELECT COUNT(*) FROM el_qu WHERE competency_question_type='validity'),'|',
  (SELECT COUNT(*) FROM el_exam WHERE assessment_type='competency'));
"

bash /tmp/staging-competency-deploy.sh "$EXPECTED_BACKEND" "$EXPECTED_INDEX"
