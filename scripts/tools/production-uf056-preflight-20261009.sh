#!/usr/bin/env bash
set -euo pipefail

ROOT=/opt/talent-assessment
TEMPLATE="$ROOT/configs/export-templates/competency-phase1-report-v2.docx"
EXPECTED_HOST=iZ0yosjdcen2p4Z
EXPECTED_OLD_SHA=a814c36e3759c8ff2cf5c17148e0939b477f17d4d3f430bb6a22f530022f9f5a

[ "$(hostname)" = "$EXPECTED_HOST" ]
[ -f "$TEMPLATE" ]
[ "$(sha256sum "$TEMPLATE" | cut -d' ' -f1)" = "$EXPECTED_OLD_SHA" ]
[ "$(systemctl is-active talent-assessment)" = active ]
[ "$(curl -fsS http://127.0.0.1:8092/health)" = '{"status":"ok"}' ]
[ "$(systemctl show talent-assessment -p NRestarts --value)" = 0 ]
command -v python3 >/dev/null
command -v libreoffice >/dev/null
command -v pdfinfo >/dev/null
command -v pdftotext >/dev/null
[ "$(df -Pk "$ROOT" | awk 'NR==2 {print $4}')" -gt 1048576 ]
printf '{"status":"passed","environment":"production","stage":"preflight","host":"%s","templateSha256":"%s","templateBytes":%s,"mainPid":%s,"nRestarts":%s,"databaseAccess":false}\n' \
  "$(hostname)" "$(sha256sum "$TEMPLATE" | cut -d' ' -f1)" "$(stat -c %s "$TEMPLATE")" \
  "$(systemctl show talent-assessment -p MainPID --value)" \
  "$(systemctl show talent-assessment -p NRestarts --value)"
