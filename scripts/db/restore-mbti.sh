#!/bin/bash
# Restore MBTI data that was lost during per-table sync
DB="element"
BDIR="/tmp/db-sync-v2"

echo "=== Restoring MBTI data ==="

# 0. Add missing columns that source DB doesn't have
echo "0. Ensuring el_exam has required columns..."
sudo mysql $DB -e "ALTER TABLE el_exam ADD COLUMN content TEXT DEFAULT NULL AFTER title" 2>/dev/null || true
sudo mysql $DB -e "ALTER TABLE el_exam ADD COLUMN required_fields VARCHAR(500) DEFAULT NULL" 2>/dev/null || true
ECOLS=$(sudo mysql $DB -N -e 'SHOW COLUMNS FROM el_exam' | wc -l)
echo "   el_exam columns: $ECOLS"

# Ensure el_qu has all needed columns too
sudo mysql $DB -e "ALTER TABLE el_qu ADD COLUMN qu_type_text VARCHAR(50) DEFAULT NULL AFTER qu_type" 2>/dev/null || true

# 1. Import MBTI-specific tables directly (these don't conflict with source)
echo "1. Restoring el_mbti_answer..."
sudo mysql $DB < $BDIR/mbti_answer.sql 2>/dev/null
echo "   $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_mbti_answer') rows"

# 2. Insert MBTI exam (use the backup INSERT line with IGNORE)
echo "2. Inserting MBTI exam..."
grep '^INSERT' $BDIR/mbti_exam.sql | sed 's/^INSERT INTO/INSERT IGNORE INTO/' | sudo mysql $DB 2>/dev/null
echo "   MBTI exam: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_exam WHERE id=2026042000040000001')"

# 3. Insert MBTI exam_repo
echo "3. Inserting exam_repo..."
grep '^INSERT' $BDIR/mbti_exam_repo.sql | sed 's/^INSERT INTO/INSERT IGNORE INTO/' | sudo mysql $DB 2>/dev/null

# 4. Insert MBTI repo (00301)
echo "4. Checking repo..."
REPO_EXISTS=$(sudo mysql $DB -N -e "SELECT COUNT(*) FROM el_repo WHERE code='00301'")
if [ "$REPO_EXISTS" = "0" ]; then
  # Need to find the repo INSERT from original backup
  # Check if there's an older backup with the repo
  for D in /tmp/db-sync-*/; do
    if [ -f "$D/mbti_repo.sql" ]; then
      grep '^INSERT' "$D/mbti_repo.sql" | sed 's/^INSERT INTO/INSERT IGNORE INTO/' | sudo mysql $DB 2>/dev/null
      echo "   repo restored from $D"
      break
    fi
  done
fi
echo "   MBTI repo: $(sudo mysql $DB -N -e "SELECT code FROM el_repo WHERE code='00301'")"

# If repo still missing, insert manually
REPO_EXISTS=$(sudo mysql $DB -N -e "SELECT COUNT(*) FROM el_repo WHERE code='00301'")
if [ "$REPO_EXISTS" = "0" ]; then
  sudo mysql $DB -e "INSERT INTO el_repo (id,code,title) VALUES ('1776654740683017062','00301','职业性格测验(MBTI)')"
  echo "   repo inserted manually"
fi

# 5. Insert MBTI qu_repo links
echo "5. Checking qu_repo..."
QR_EXISTS=$(sudo mysql $DB -N -e "SELECT COUNT(*) FROM el_qu_repo WHERE repo_id='1776654740683017062'")
if [ "$QR_EXISTS" = "0" ]; then
  for D in /tmp/db-sync-*/; do
    if [ -f "$D/mbti_qu_repo.sql" ]; then
      grep '^INSERT' "$D/mbti_qu_repo.sql" | sed 's/^INSERT INTO/INSERT IGNORE INTO/' | sudo mysql $DB 2>/dev/null
      echo "   qu_repo restored from $D"
      break
    fi
  done
fi

# 6. MBTI questions (should already be in el_qu from source sync if they exist)
echo "6. MBTI questions: $(sudo mysql $DB -N -e "SELECT COUNT(*) FROM el_qu WHERE id LIKE '2026042000010%'")"

# 7. Restore tester/candidate/paper
echo "7. Restoring tester/candidate/paper..."
grep '^INSERT' $BDIR/mbti_tester.sql 2>/dev/null | sed 's/^INSERT INTO/INSERT IGNORE INTO/' | sudo mysql $DB 2>/dev/null
grep '^INSERT' $BDIR/mbti_candidate.sql 2>/dev/null | sed 's/^INSERT INTO/INSERT IGNORE INTO/' | sudo mysql $DB 2>/dev/null
grep '^INSERT' $BDIR/mbti_paper.sql 2>/dev/null | sed 's/^INSERT INTO/INSERT IGNORE INTO/' | sudo mysql $DB 2>/dev/null
grep '^INSERT' $BDIR/mbti_pq.sql 2>/dev/null | sed 's/^INSERT INTO/INSERT IGNORE INTO/' | sudo mysql $DB 2>/dev/null
grep '^INSERT' $BDIR/mbti_pqa.sql 2>/dev/null | sed 's/^INSERT INTO/INSERT IGNORE INTO/' | sudo mysql $DB 2>/dev/null

# 8. Add mbti columns if missing
sudo mysql $DB -e "ALTER TABLE el_tester ADD COLUMN mbti_type VARCHAR(4) DEFAULT NULL AFTER pdf_flag, ADD COLUMN mbti_scores VARCHAR(200) DEFAULT NULL AFTER mbti_type" 2>/dev/null

# 9. Restore mbti_type data
if [ -f "$BDIR/mbti_tester_data.tsv" ]; then
  while IFS=$'\t' read -r tid mt ms; do
    if [ -n "$mt" ]; then
      sudo mysql $DB -e "UPDATE el_tester SET mbti_type='$mt', mbti_scores='$ms' WHERE id='$tid'" 2>/dev/null
    fi
  done < $BDIR/mbti_tester_data.tsv
fi

# Verify
echo ""
echo "=== Verification ==="
echo "  exams: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_exam') (should be 40)"
echo "  MBTI exam: $(sudo mysql $DB -N -e 'SELECT title FROM el_exam WHERE id=2026042000040000001')"
echo "  testers: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_tester')"
echo "  repos: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_repo')"
echo "  MBTI repo: $(sudo mysql $DB -N -e "SELECT code FROM el_repo WHERE code='00301'")"
echo "  questions: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_qu')"
echo "  MBTI qu: $(sudo mysql $DB -N -e "SELECT COUNT(*) FROM el_qu WHERE id LIKE '2026042000010%'")"
echo "  papers: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_paper')"
echo "  mbti_answers: $(sudo mysql $DB -N -e 'SELECT COUNT(*) FROM el_mbti_answer')"

sudo systemctl restart talent-assessment
sleep 2
echo "  service: $(sudo systemctl is-active talent-assessment)"
echo "=== DONE ==="
