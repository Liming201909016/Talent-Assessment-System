#!/bin/bash
# Per-table sync from 39.106.61.48 → localhost, preserving MBTI
# Run on 20.200.136.133

SRC_H="39.106.61.48"
SRC_U="root"
SRC_P="YFkdkiP3r1RWbsWz"
DB="element"
BDIR="/tmp/db-sync-v2"
mkdir -p $BDIR

echo "=== 1. Backup MBTI data ==="
sudo mysqldump $DB el_mbti_answer > $BDIR/mbti_answer.sql 2>/dev/null
sudo mysqldump $DB el_exam --where="id='2026042000040000001'" > $BDIR/mbti_exam.sql 2>/dev/null
sudo mysqldump $DB el_exam_repo --where="exam_id='2026042000040000001'" > $BDIR/mbti_exam_repo.sql 2>/dev/null
sudo mysqldump $DB el_tester --where="exam_id='2026042000040000001'" > $BDIR/mbti_tester.sql 2>/dev/null
sudo mysqldump $DB el_candidate --where="exam_id='2026042000040000001'" > $BDIR/mbti_candidate.sql 2>/dev/null
PIDS=$(sudo mysql $DB -N -e "SELECT GROUP_CONCAT(QUOTE(id)) FROM el_paper WHERE exam_id='2026042000040000001'")
sudo mysqldump $DB el_paper --where="exam_id='2026042000040000001'" > $BDIR/mbti_paper.sql 2>/dev/null
if [ -n "$PIDS" ] && [ "$PIDS" != "NULL" ]; then
  sudo mysqldump $DB el_paper_qu --where="paper_id IN ($PIDS)" > $BDIR/mbti_pq.sql 2>/dev/null
  sudo mysqldump $DB el_paper_qu_answer --where="paper_id IN ($PIDS)" > $BDIR/mbti_pqa.sql 2>/dev/null
fi
sudo mysql $DB -N -e "SELECT id,mbti_type,mbti_scores FROM el_tester WHERE mbti_type IS NOT NULL" > $BDIR/mbti_tester_data.tsv
echo "  MBTI backup done ($(ls $BDIR/*.sql 2>/dev/null | wc -l) files)"

echo ""
echo "=== 2. Per-table sync from source ==="
# Tables to sync from source (production data)
SYNC_TABLES="el_exam el_exam_depart el_exam_repo el_repo el_qu el_qu_answer el_qu_repo el_tester el_candidate el_paper el_paper_qu el_paper_qu_answer el_user_book el_user_exam"

for T in $SYNC_TABLES; do
  echo -n "  $T: "
  # Dump from source, pipe to local
  mysqldump -h $SRC_H -u $SRC_U -p"$SRC_P" \
    --single-transaction --column-statistics=0 \
    --set-gtid-purged=OFF --default-character-set=utf8 \
    $DB $T 2>/dev/null > $BDIR/table_${T}.sql
  SIZE=$(du -h $BDIR/table_${T}.sql | cut -f1)
  sudo mysql $DB < $BDIR/table_${T}.sql 2>/dev/null
  COUNT=$(sudo mysql $DB -N -e "SELECT COUNT(*) FROM $T" 2>/dev/null)
  echo "$COUNT rows ($SIZE)"
done

echo ""
echo "=== 3. Add MBTI columns ==="
sudo mysql $DB -e "ALTER TABLE el_tester ADD COLUMN mbti_type VARCHAR(4) DEFAULT NULL AFTER pdf_flag, ADD COLUMN mbti_scores VARCHAR(200) DEFAULT NULL AFTER mbti_type" 2>/dev/null && echo "  columns added" || echo "  columns already exist"

echo ""
echo "=== 4. Restore MBTI ==="
sudo mysql $DB -e "CREATE TABLE IF NOT EXISTS el_mbti_answer (id varchar(20) NOT NULL, paper_id varchar(20), qu_id varchar(20), score_a int DEFAULT 0, score_b int DEFAULT 0, sort int DEFAULT 0, answered tinyint DEFAULT 0, create_time datetime, PRIMARY KEY(id), KEY idx_paper(paper_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci"

# Use sed to change DROP TABLE + CREATE TABLE to INSERT IGNORE only
for F in mbti_answer mbti_exam mbti_exam_repo mbti_tester mbti_candidate mbti_paper mbti_pq mbti_pqa; do
  if [ -f "$BDIR/${F}.sql" ]; then
    # Strip DROP TABLE and CREATE TABLE, keep INSERT only, change to INSERT IGNORE
    grep '^INSERT ' $BDIR/${F}.sql | sed 's/^INSERT /INSERT IGNORE /' | sudo mysql $DB 2>/dev/null
    echo "  restored $F (INSERT IGNORE)"
  fi
done

# Restore mbti_type/mbti_scores
if [ -f "$BDIR/mbti_tester_data.tsv" ]; then
  while IFS=$'\t' read -r tid mt ms; do
    if [ -n "$mt" ]; then
      sudo mysql $DB -e "UPDATE el_tester SET mbti_type='$mt', mbti_scores='$ms' WHERE id='$tid'" 2>/dev/null
    fi
  done < $BDIR/mbti_tester_data.tsv
  echo "  tester mbti data restored"
fi

echo ""
echo "=== 5. Verify ==="
echo "  exams: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_exam')"
echo "  testers: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_tester')"
echo "  repos: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_repo')"
echo "  questions: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_qu')"
echo "  papers: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_paper')"
echo "  mbti_answers: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_mbti_answer')"
echo "  mbti_exam: $(sudo mysql $DB -N -e "SELECT title FROM el_exam WHERE id='2026042000040000001'")"

echo ""
echo "=== 6. Restart ==="
sudo systemctl restart talent-assessment
sleep 2
sudo systemctl is-active talent-assessment
echo ""
echo "=== SYNC COMPLETE ==="
