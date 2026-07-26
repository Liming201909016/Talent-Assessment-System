#!/bin/bash
# Schema consistency check: prod (39.106.61.48) vs test (localhost)
# Prod DB is READ-ONLY

PROD="mysql -h 39.106.61.48 -u root -pYFkdkiP3r1RWbsWz element -N"
TEST="sudo mysql element -N"

echo "=== Schema Consistency Check ==="
echo "Prod: 39.106.61.48 | Test: localhost (20.200.136.133)"
echo ""

echo "--- 1. Table Count ---"
P=$($PROD -e 'SELECT COUNT(*) FROM information_schema.tables WHERE table_schema="element"' 2>/dev/null)
T=$($TEST -e 'SELECT COUNT(*) FROM information_schema.tables WHERE table_schema="element"')
echo "Prod=$P Test=$T $([ "$P" = "$T" ] && echo '✓' || echo '✗')"

echo ""
echo "--- 2. Table Names ---"
diff <($PROD -e 'SELECT table_name FROM information_schema.tables WHERE table_schema="element" ORDER BY table_name' 2>/dev/null) \
     <($TEST -e 'SELECT table_name FROM information_schema.tables WHERE table_schema="element" ORDER BY table_name') \
  && echo "Tables: MATCH" || echo "Tables: DIFF (see above)"

echo ""
echo "--- 3. Row Counts ---"
for tbl in el_exam el_tester el_qu el_qu_answer el_qu_repo el_paper el_paper_qu el_paper_qu_answer el_candidate el_repo el_exam_repo el_exam_depart el_user_exam el_user_book sys_user sys_role sys_menu sys_dept sys_dict_type sys_dict_data sys_config sys_notice sys_post sys_role_menu sys_user_role; do
  P=$($PROD -e "SELECT COUNT(*) FROM $tbl" 2>/dev/null)
  T=$($TEST -e "SELECT COUNT(*) FROM $tbl" 2>/dev/null)
  if [ "$P" = "$T" ]; then
    echo "✓ $tbl: $P"
  else
    echo "✗ $tbl: prod=$P test=$T"
  fi
done

echo ""
echo "--- 4. Column Count per Table ---"
diff <($PROD -e 'SELECT table_name, COUNT(*) FROM information_schema.columns WHERE table_schema="element" GROUP BY table_name ORDER BY table_name' 2>/dev/null) \
     <($TEST -e 'SELECT table_name, COUNT(*) FROM information_schema.columns WHERE table_schema="element" GROUP BY table_name ORDER BY table_name') \
  && echo "Columns: MATCH" || echo "Columns: DIFF"

echo ""
echo "--- 5. Key Business Data Summary ---"
echo "Exams by is_open:"
$TEST -e 'SELECT is_open, COUNT(*) AS cnt FROM el_exam GROUP BY is_open'
echo "Exams by open_type:"
$TEST -e 'SELECT open_type, COUNT(*) AS cnt FROM el_exam GROUP BY open_type'
echo "Repos:"
$TEST -e 'SELECT id, code, title FROM el_repo ORDER BY id'
echo "Users:"
$TEST -e 'SELECT user_id, user_name, real_name FROM sys_user ORDER BY user_id'
echo "Testers by exam:"
$TEST -e 'SELECT exam_id, COUNT(*) AS cnt FROM el_tester GROUP BY exam_id ORDER BY cnt DESC LIMIT 10'
echo "Papers by state:"
$TEST -e 'SELECT state, COUNT(*) AS cnt FROM el_paper GROUP BY state'
echo "Candidates by exam:"
$TEST -e 'SELECT exam_id, COUNT(*) AS cnt FROM el_candidate GROUP BY exam_id ORDER BY cnt DESC LIMIT 10'

echo ""
echo "CHECK_DONE"
