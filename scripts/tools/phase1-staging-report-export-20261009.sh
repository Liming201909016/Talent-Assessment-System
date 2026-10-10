#!/usr/bin/env bash
set -euo pipefail
umask 077

EXPECTED_HOST=${1:-vm-ubuntu-go-dev}
OUTPUT_DIR=${2:-/tmp/phase1-staging-baseline-20261009}

if [ "$(hostname)" != "$EXPECTED_HOST" ]; then
  echo 'EXPORT_ERROR=unexpected_host'
  exit 1
fi

rm -rf "$OUTPUT_DIR"
install -d -m 700 "$OUTPUT_DIR"
sudo -n mysqldump --compact --skip-extended-insert --no-create-info --skip-triggers \
  --where="content_version='competency-phase1-content-v1'" \
  element el_competency_report_text > "$OUTPUT_DIR/v1-text.sql"
sudo -n mysqldump --compact --skip-extended-insert --no-create-info --skip-triggers \
  --where="content_version='competency-phase1-content-v1'" \
  element el_competency_report_content_package > "$OUTPUT_DIR/v1-package.sql"
cp /opt/talent-assessment/configs/export-templates/competency-phase1-report-v2.docx \
  "$OUTPUT_DIR/v2-template.docx"
chmod 600 "$OUTPUT_DIR"/*
sha256sum "$OUTPUT_DIR"/*
printf 'EXPORT=PASS\n'
