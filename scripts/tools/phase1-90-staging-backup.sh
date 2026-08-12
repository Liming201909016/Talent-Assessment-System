#!/usr/bin/env bash
set -euo pipefail

APP=/opt/talent-assessment
STAMP=$(date +%Y%m%d_%H%M%S)
BACKUP="$APP/backups/phase1_90_import_$STAMP"
sudo mkdir -p "$BACKUP"
sudo chmod 700 "$BACKUP"
sudo mysqldump --single-transaction --routines --triggers --events element | gzip -9 | sudo tee "$BACKUP/element.sql.gz" >/dev/null

EXEC_START=$(systemctl show talent-assessment --property=ExecStart --value | sed -n 's/.*path=\([^ ;]*\).*/\1/p')
if [ -z "$EXEC_START" ]; then
  EXEC_START="$APP/talent-assessment"
fi
sudo cp "$EXEC_START" "$BACKUP/backend.before"

if [ -d "$APP/ruoyi-ui/dist" ]; then
  sudo tar -C "$APP/ruoyi-ui" -czf "$BACKUP/frontend-dist.before.tar.gz" dist
elif [ -d "$APP/dist" ]; then
  sudo tar -C "$APP" -czf "$BACKUP/frontend-dist.before.tar.gz" dist
else
  echo FRONTEND_DIST_NOT_FOUND >&2
  exit 1
fi

sudo mysql element -Nse "SELECT * FROM el_qu WHERE dimension_id IS NULL AND competency_question_type IS NULL ORDER BY id" | sha256sum | awk '{print $1}' | sudo tee "$BACKUP/traditional-qu.sha256" >/dev/null
sudo mysql element -Nse "SELECT * FROM el_qu_repo ORDER BY id" | sha256sum | awk '{print $1}' | sudo tee "$BACKUP/traditional-qu-repo.sha256" >/dev/null
sudo mysql element -Nse "SELECT * FROM el_qu_answer ORDER BY id" | sha256sum | awk '{print $1}' | sudo tee "$BACKUP/traditional-qu-answer.sha256" >/dev/null
sudo sha256sum "$BACKUP/element.sql.gz" "$BACKUP/backend.before" "$BACKUP/frontend-dist.before.tar.gz" | sudo tee "$BACKUP/SHA256SUMS" >/dev/null
sudo find "$BACKUP" -mindepth 1 -maxdepth 1 -type f -exec chmod 600 {} +

sudo mysql element -Nse "SELECT COUNT(*) FROM el_competency_dimension; SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NOT NULL OR competency_question_type IS NOT NULL; SELECT COUNT(*) FROM el_exam WHERE assessment_type='competency'; SELECT COUNT(*) FROM el_competency_migration WHERE migration_key='competency-009-phase1-identity-reset';"
printf 'BACKUP=%s\n' "$BACKUP"
printf 'EXEC_START=%s\n' "$EXEC_START"
printf 'SERVICES='
systemctl is-active talent-assessment nginx mysql | paste -sd, -
printf 'DISK='
df -h "$APP" | tail -1 | awk '{print $4"_free"}'
