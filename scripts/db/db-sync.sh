#!/bin/bash
# Data Sync: 39.106.61.48 (source, READ-ONLY) → 20.200.136.133 (target)
# Preserves MBTI data on target
# Run on: 20.200.136.133

set +e  # Don't exit on error, handle manually

SOURCE_HOST="39.106.61.48"
SOURCE_USER="root"
SOURCE_PWD="YFkdkiP3r1RWbsWz"
TARGET_DB="element"
BACKUP_DIR="/tmp/db-sync-$(date +%Y%m%d%H%M%S)"
mkdir -p $BACKUP_DIR

echo "============================================"
echo "  DB Sync: $SOURCE_HOST → localhost"
echo "  Backup dir: $BACKUP_DIR"
echo "============================================"

# Step 1: Backup MBTI data from target (BEFORE overwrite)
echo ""
echo "=== Step 1: Backup MBTI data ==="
sudo mysqldump $TARGET_DB el_mbti_answer > $BACKUP_DIR/mbti_answer.sql 2>/dev/null
echo "  el_mbti_answer: $(sudo mysql $TARGET_DB -N -e 'SELECT COUNT(*) FROM el_mbti_answer') rows"

# Backup MBTI-specific records
sudo mysql $TARGET_DB -N -e "SELECT id,title FROM el_exam WHERE id='2026042000040000001'" > $BACKUP_DIR/mbti_exam_check.txt
sudo mysqldump $TARGET_DB el_exam --where="id='2026042000040000001'" > $BACKUP_DIR/mbti_exam.sql 2>/dev/null
sudo mysqldump $TARGET_DB el_exam_repo --where="exam_id='2026042000040000001'" > $BACKUP_DIR/mbti_exam_repo.sql 2>/dev/null
sudo mysqldump $TARGET_DB el_repo --where="code='00301'" > $BACKUP_DIR/mbti_repo.sql 2>/dev/null
sudo mysqldump $TARGET_DB el_qu --where="id LIKE '2026042000010%'" > $BACKUP_DIR/mbti_qu.sql 2>/dev/null
sudo mysqldump $TARGET_DB el_qu_answer --where="qu_id LIKE '2026042000010%'" > $BACKUP_DIR/mbti_qu_answer.sql 2>/dev/null
sudo mysqldump $TARGET_DB el_qu_repo --where="repo_id='1776654740683017062'" > $BACKUP_DIR/mbti_qu_repo.sql 2>/dev/null
# Get MBTI paper IDs for subqueries
MBTI_PAPER_IDS=$(sudo mysql $TARGET_DB -N -e "SELECT GROUP_CONCAT(QUOTE(id)) FROM el_paper WHERE exam_id='2026042000040000001'")
echo "  MBTI paper IDs: $MBTI_PAPER_IDS"
if [ -n "$MBTI_PAPER_IDS" ] && [ "$MBTI_PAPER_IDS" != "NULL" ]; then
  sudo mysqldump $TARGET_DB el_paper --where="exam_id='2026042000040000001'" > $BACKUP_DIR/mbti_paper.sql 2>/dev/null
  sudo mysqldump $TARGET_DB el_paper_qu --where="paper_id IN ($MBTI_PAPER_IDS)" > $BACKUP_DIR/mbti_paper_qu.sql 2>/dev/null
  sudo mysqldump $TARGET_DB el_paper_qu_answer --where="paper_id IN ($MBTI_PAPER_IDS)" > $BACKUP_DIR/mbti_paper_qu_answer.sql 2>/dev/null
fi
# Backup tester MBTI columns
sudo mysql $TARGET_DB -N -e "SELECT id, mbti_type, mbti_scores FROM el_tester WHERE mbti_type IS NOT NULL" > $BACKUP_DIR/mbti_tester_data.tsv
# Backup MBTI tester records (those linked to MBTI exam)
sudo mysqldump $TARGET_DB el_tester --where="exam_id='2026042000040000001'" > $BACKUP_DIR/mbti_tester.sql 2>/dev/null
# Backup candidate records for MBTI
sudo mysqldump $TARGET_DB el_candidate --where="exam_id='2026042000040000001'" > $BACKUP_DIR/mbti_candidate.sql 2>/dev/null

echo "  MBTI backup complete: $(ls -1 $BACKUP_DIR/*.sql | wc -l) files"

# Step 2: Dump from source (READ-ONLY)
echo ""
echo "=== Step 2: Dump from source ($SOURCE_HOST) ==="
mysqldump -h $SOURCE_HOST -u $SOURCE_USER -p"$SOURCE_PWD" \
  --single-transaction --quick --set-gtid-purged=OFF \
  $TARGET_DB > $BACKUP_DIR/source_dump.sql 2>/dev/null
SOURCE_SIZE=$(du -h $BACKUP_DIR/source_dump.sql | cut -f1)
echo "  Source dump: $SOURCE_SIZE"

# Step 3: Import source dump to target (overwrite)
echo ""
echo "=== Step 3: Import source dump ==="
sudo mysql -e "DROP DATABASE IF EXISTS ${TARGET_DB}_backup"
sudo mysql -e "CREATE DATABASE ${TARGET_DB}_backup"
sudo mysqldump $TARGET_DB | sudo mysql ${TARGET_DB}_backup 2>/dev/null
echo "  Full backup saved to ${TARGET_DB}_backup database"

sudo mysql $TARGET_DB < $BACKUP_DIR/source_dump.sql
echo "  Source data imported"

# Step 4: Restore MBTI data
echo ""
echo "=== Step 4: Restore MBTI data ==="

# Add MBTI columns back to el_tester (source doesn't have them)
sudo mysql $TARGET_DB -e "ALTER TABLE el_tester ADD COLUMN mbti_type VARCHAR(4) DEFAULT NULL COMMENT 'MBTI类型' AFTER pdf_flag, ADD COLUMN mbti_scores VARCHAR(200) DEFAULT NULL COMMENT 'MBTI八维度JSON' AFTER mbti_type" 2>/dev/null || echo "  mbti columns may already exist"

# Restore MBTI tables (el_mbti_answer is new, source doesn't have it)
sudo mysql $TARGET_DB -e "CREATE TABLE IF NOT EXISTS el_mbti_answer (
  id varchar(20) NOT NULL,
  paper_id varchar(20) DEFAULT NULL,
  qu_id varchar(20) DEFAULT NULL,
  score_a int DEFAULT 0,
  score_b int DEFAULT 0,
  sort int DEFAULT 0,
  answered tinyint DEFAULT 0,
  create_time datetime DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_paper_id (paper_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci"
sudo mysql $TARGET_DB < $BACKUP_DIR/mbti_answer.sql
echo "  el_mbti_answer restored: $(sudo mysql $TARGET_DB -N -e 'SELECT COUNT(*) FROM el_mbti_answer') rows"

# Restore MBTI repo, questions, answers
sudo mysql $TARGET_DB < $BACKUP_DIR/mbti_repo.sql 2>/dev/null || true
sudo mysql $TARGET_DB < $BACKUP_DIR/mbti_qu.sql 2>/dev/null || true
sudo mysql $TARGET_DB < $BACKUP_DIR/mbti_qu_answer.sql 2>/dev/null || true
sudo mysql $TARGET_DB < $BACKUP_DIR/mbti_qu_repo.sql 2>/dev/null || true
echo "  MBTI repo+questions restored"

# Restore MBTI exam
sudo mysql $TARGET_DB < $BACKUP_DIR/mbti_exam.sql 2>/dev/null || true
sudo mysql $TARGET_DB < $BACKUP_DIR/mbti_exam_repo.sql 2>/dev/null || true
echo "  MBTI exam restored"

# Restore MBTI papers
sudo mysql $TARGET_DB < $BACKUP_DIR/mbti_paper.sql 2>/dev/null || true
sudo mysql $TARGET_DB < $BACKUP_DIR/mbti_paper_qu.sql 2>/dev/null || true
sudo mysql $TARGET_DB < $BACKUP_DIR/mbti_paper_qu_answer.sql 2>/dev/null || true
echo "  MBTI papers restored"

# Restore MBTI testers
sudo mysql $TARGET_DB < $BACKUP_DIR/mbti_tester.sql 2>/dev/null || true
sudo mysql $TARGET_DB < $BACKUP_DIR/mbti_candidate.sql 2>/dev/null || true
echo "  MBTI testers restored"

# Restore mbti_type/mbti_scores for testers
while IFS=$'\t' read -r tid mtype mscores; do
  if [ -n "$mtype" ]; then
    sudo mysql $TARGET_DB -e "UPDATE el_tester SET mbti_type='$mtype', mbti_scores='$mscores' WHERE id='$tid'" 2>/dev/null
  fi
done < $BACKUP_DIR/mbti_tester_data.tsv
echo "  MBTI tester scores restored"

# Step 5: Verify
echo ""
echo "=== Step 5: Verification ==="
echo "  el_exam count: $(sudo mysql $TARGET_DB -N -e 'SELECT COUNT(*) FROM el_exam')"
echo "  el_tester count: $(sudo mysql $TARGET_DB -N -e 'SELECT COUNT(*) FROM el_tester')"
echo "  el_qu count: $(sudo mysql $TARGET_DB -N -e 'SELECT COUNT(*) FROM el_qu')"
echo "  el_mbti_answer count: $(sudo mysql $TARGET_DB -N -e 'SELECT COUNT(*) FROM el_mbti_answer')"
echo "  MBTI exam: $(sudo mysql $TARGET_DB -N -e "SELECT id,title FROM el_exam WHERE id='2026042000040000001'")"
echo "  MBTI repo: $(sudo mysql $TARGET_DB -N -e "SELECT id,code FROM el_repo WHERE code='00301'")"
echo "  MBTI papers: $(sudo mysql $TARGET_DB -N -e "SELECT COUNT(*) FROM el_paper WHERE exam_id='2026042000040000001'")"
echo "  mbti_type column: $(sudo mysql $TARGET_DB -N -e "SHOW COLUMNS FROM el_tester LIKE 'mbti_type'" | head -1)"

# Restart Go service
echo ""
echo "=== Step 6: Restart service ==="
sudo systemctl restart talent-assessment
sleep 2
sudo systemctl is-active talent-assessment

echo ""
echo "=== SYNC COMPLETE ==="
echo "Backup dir: $BACKUP_DIR"
