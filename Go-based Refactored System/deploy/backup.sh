#!/bin/bash
# Daily MySQL backup - called by cron at 3am
BACKUP_DIR=/data/backup
mkdir -p "$BACKUP_DIR"
FILENAME="element-$(date +%Y%m%d).sql"
sudo mysqldump element --single-transaction --result-file="$BACKUP_DIR/$FILENAME"
# Keep only last 7 days
find "$BACKUP_DIR" -name 'element-*.sql' -mtime +7 -delete
echo "$(date): Backup $FILENAME done" >> "$BACKUP_DIR/backup.log"
