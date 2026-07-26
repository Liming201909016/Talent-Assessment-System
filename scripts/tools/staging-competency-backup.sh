#!/usr/bin/env bash
set -euo pipefail

backup_dir=/opt/talent-assessment/backups
stamp=$(date +%Y%m%d_%H%M%S)
backup_file="$backup_dir/element_before_competency_${stamp}.sql.gz"

sudo -n mkdir -p "$backup_dir"
sudo -n mysqldump --single-transaction --routines --triggers --events --set-gtid-purged=OFF element | gzip -9 | sudo -n tee "$backup_file" >/dev/null
sudo -n chmod 600 "$backup_file"
sudo -n gzip -t "$backup_file"

size=$(sudo -n stat -c %s "$backup_file")
hash=$(sudo -n sha256sum "$backup_file" | cut -d' ' -f1)
echo "BACKUP_OK"
echo "path=$backup_file"
echo "bytes=$size"
echo "sha256=$hash"
