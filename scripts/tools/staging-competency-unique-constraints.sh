#!/usr/bin/env bash
set -euo pipefail

DB=element
PREFIX=competency-constraint-check
CODE_A=ZZ-CONSTRAINT-CODE-A
CODE_B=ZZ-CONSTRAINT-CODE-B
CODE_C=ZZ-CONSTRAINT-ITEM-A
CODE_D=ZZ-CONSTRAINT-ITEM-B
DIMENSION_ID=competency-d01

cleanup() {
  sudo -n mysql "$DB" -e "
    DELETE FROM el_qu
    WHERE id IN ('${PREFIX}-code-a','${PREFIX}-code-b','${PREFIX}-item-a','${PREFIX}-item-b');
  " >/dev/null
}
trap cleanup EXIT
cleanup

before_count=$(sudo -n mysql "$DB" -Nse "SELECT COUNT(*) FROM el_qu;")
index_count=$(sudo -n mysql "$DB" -Nse "
  SELECT COUNT(DISTINCT INDEX_NAME)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA='${DB}' AND TABLE_NAME='el_qu'
    AND INDEX_NAME IN ('uk_qu_question_code','uk_qu_dimension_item')
    AND NON_UNIQUE=0;
")
if [[ "$index_count" != "2" ]]; then
  echo "UNIQUE_INDEX_PREFLIGHT_FAILED count=$index_count" >&2
  exit 1
fi

set +e
sudo -n mysql "$DB" >/tmp/competency-duplicate-code.out 2>&1 <<SQL
START TRANSACTION;
INSERT INTO el_qu
(id,qu_type,level,content,remark,question_code,dimension_id,dimension_item_no,observation_point,scoring_direction,question_status,create_time,update_time)
VALUES
('${PREFIX}-code-a',1,0,'constraint check A','temporary','${CODE_A}','${DIMENSION_ID}',900001,'temporary','forward',0,NOW(),NOW());
INSERT INTO el_qu
(id,qu_type,level,content,remark,question_code,dimension_id,dimension_item_no,observation_point,scoring_direction,question_status,create_time,update_time)
VALUES
('${PREFIX}-code-b',1,0,'constraint check B','temporary','${CODE_A}','${DIMENSION_ID}',900002,'temporary','forward',0,NOW(),NOW());
ROLLBACK;
SQL
code_status=$?

sudo -n mysql "$DB" >/tmp/competency-duplicate-item.out 2>&1 <<SQL
START TRANSACTION;
INSERT INTO el_qu
(id,qu_type,level,content,remark,question_code,dimension_id,dimension_item_no,observation_point,scoring_direction,question_status,create_time,update_time)
VALUES
('${PREFIX}-item-a',1,0,'constraint check C','temporary','${CODE_C}','${DIMENSION_ID}',900003,'temporary','forward',0,NOW(),NOW());
INSERT INTO el_qu
(id,qu_type,level,content,remark,question_code,dimension_id,dimension_item_no,observation_point,scoring_direction,question_status,create_time,update_time)
VALUES
('${PREFIX}-item-b',1,0,'constraint check D','temporary','${CODE_D}','${DIMENSION_ID}',900003,'temporary','reverse',0,NOW(),NOW());
ROLLBACK;
SQL
item_status=$?
set -e

if [[ "$code_status" -eq 0 ]] || ! grep -q "Duplicate entry" /tmp/competency-duplicate-code.out || ! grep -q "uk_qu_question_code" /tmp/competency-duplicate-code.out; then
  echo "QUESTION_CODE_UNIQUE_FAILED status=$code_status" >&2
  sed -E 's/(password|token|secret)=[^ ]+/${1}=***/Ig' /tmp/competency-duplicate-code.out >&2
  exit 1
fi
if [[ "$item_status" -eq 0 ]] || ! grep -q "Duplicate entry" /tmp/competency-duplicate-item.out || ! grep -q "uk_qu_dimension_item" /tmp/competency-duplicate-item.out; then
  echo "DIMENSION_ITEM_UNIQUE_FAILED status=$item_status" >&2
  sed -E 's/(password|token|secret)=[^ ]+/${1}=***/Ig' /tmp/competency-duplicate-item.out >&2
  exit 1
fi

cleanup
trap - EXIT
after_count=$(sudo -n mysql "$DB" -Nse "SELECT COUNT(*) FROM el_qu;")
remaining=$(sudo -n mysql "$DB" -Nse "
  SELECT COUNT(*) FROM el_qu
  WHERE id IN ('${PREFIX}-code-a','${PREFIX}-code-b','${PREFIX}-item-a','${PREFIX}-item-b')
     OR question_code IN ('${CODE_A}','${CODE_B}','${CODE_C}','${CODE_D}');
")
if [[ "$before_count" != "$after_count" ]] || [[ "$remaining" != "0" ]]; then
  echo "CONSTRAINT_CHECK_CLEANUP_FAILED before=$before_count after=$after_count remaining=$remaining" >&2
  exit 1
fi

echo "UNIQUE_CONSTRAINTS_OK"
echo "question_code_duplicate_rejected=1"
echo "dimension_item_duplicate_rejected=1"
echo "before_count=$before_count"
echo "after_count=$after_count"
echo "remaining=$remaining"
