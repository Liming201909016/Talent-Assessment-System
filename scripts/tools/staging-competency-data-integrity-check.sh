#!/usr/bin/env bash
set -euo pipefail

DB=element
failures=0
check_eq() {
  local name="$1" actual="$2" expected="$3"
  if [[ "$actual" == "$expected" ]]; then
    echo "PASS $name=$actual"
  else
    echo "FAIL $name actual=$actual expected=$expected" >&2
    failures=$((failures + 1))
  fi
}
query() { sudo -n mysql "$DB" -Nse "$1"; }

# Required schema columns and exact core types/nullability.
check_eq dimension_columns "$(query "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='$DB' AND TABLE_NAME='el_competency_dimension' AND COLUMN_NAME IN ('id','code','name','vird_level','applicable_category','core_meaning','display_order','status','create_time','update_time');")" 10
check_eq exam_dimension_columns "$(query "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='$DB' AND TABLE_NAME='el_exam_competency_dimension' AND COLUMN_NAME IN ('id','exam_id','dimension_id','dimension_code','dimension_name','vird_level','applicable_category','core_meaning','display_order','question_count','create_time','snapshot_time');")" 12
check_eq question_extension_definition "$(query "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='$DB' AND TABLE_NAME='el_qu' AND ((COLUMN_NAME='question_code' AND COLUMN_TYPE='varchar(32)' AND IS_NULLABLE='YES') OR (COLUMN_NAME='dimension_id' AND COLUMN_TYPE='varchar(64)' AND IS_NULLABLE='YES') OR (COLUMN_NAME='dimension_item_no' AND COLUMN_TYPE='int' AND IS_NULLABLE='YES') OR (COLUMN_NAME='observation_point' AND COLUMN_TYPE='varchar(255)' AND IS_NULLABLE='YES') OR (COLUMN_NAME='scoring_direction' AND COLUMN_TYPE='varchar(16)' AND IS_NULLABLE='YES') OR (COLUMN_NAME='question_status' AND COLUMN_TYPE='tinyint' AND IS_NULLABLE='NO')); ")" 6
check_eq exam_dispatch_definition "$(query "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='$DB' AND TABLE_NAME='el_exam' AND ((COLUMN_NAME='assessment_type' AND COLUMN_TYPE='varchar(32)' AND IS_NULLABLE='NO') OR (COLUMN_NAME='scoring_mode' AND COLUMN_TYPE='varchar(32)' AND IS_NULLABLE='NO') OR (COLUMN_NAME='competency_report_audience' AND COLUMN_TYPE='varchar(32)' AND IS_NULLABLE='YES') OR (COLUMN_NAME='publish_status' AND COLUMN_TYPE='tinyint' AND IS_NULLABLE='NO') OR (COLUMN_NAME='published_at' AND COLUMN_TYPE='datetime' AND IS_NULLABLE='YES') OR (COLUMN_NAME='published_by' AND COLUMN_TYPE='bigint' AND IS_NULLABLE='YES'));")" 6
check_eq runtime_tables "$(query "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='$DB' AND TABLE_NAME IN ('el_exam_competency_question','el_competency_dimension_result','el_competency_result');")" 3
check_eq paper_question_columns "$(query "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='$DB' AND TABLE_NAME='el_paper_qu' AND ((COLUMN_NAME='exam_question_id' AND COLUMN_TYPE='varchar(64)' AND IS_NULLABLE='YES') OR (COLUMN_NAME='raw_answer' AND COLUMN_TYPE='tinyint' AND IS_NULLABLE='YES') OR (COLUMN_NAME='final_score' AND COLUMN_TYPE='tinyint' AND IS_NULLABLE='YES'));")" 3
check_eq runtime_decimal_columns "$(query "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='$DB' AND ((TABLE_NAME='el_competency_dimension_result' AND COLUMN_NAME='dimension_score' AND COLUMN_TYPE='decimal(18,6)' AND IS_NULLABLE='YES') OR (TABLE_NAME='el_competency_result' AND COLUMN_NAME='overall_score' AND COLUMN_TYPE='decimal(18,6)' AND IS_NULLABLE='NO') OR (TABLE_NAME='el_competency_result' AND COLUMN_NAME='evaluation_average' AND COLUMN_TYPE='decimal(18,6)' AND IS_NULLABLE='YES'));")" 3
check_eq required_indexes "$(query "SELECT COUNT(DISTINCT CONCAT(TABLE_NAME,'.',INDEX_NAME)) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA='$DB' AND CONCAT(TABLE_NAME,'.',INDEX_NAME) IN ('el_competency_dimension.uk_competency_dimension_code','el_competency_dimension.uk_competency_dimension_name','el_competency_dimension.uk_competency_dimension_order','el_competency_dimension.idx_competency_dimension_status_order','el_exam_competency_dimension.uk_exam_competency_dimension','el_exam_competency_dimension.idx_exam_competency_dimension_order','el_qu.uk_qu_question_code','el_qu.uk_qu_dimension_item','el_qu.idx_qu_dimension_status','el_exam_competency_question.uk_exam_competency_question_source','el_exam_competency_question.uk_exam_competency_question_code','el_exam_competency_question.idx_exam_competency_question_dimension','el_paper_qu.uk_paper_exam_question','el_paper_qu.idx_paper_question_answered','el_paper.idx_paper_state_limit_time','el_competency_dimension_result.uk_competency_dimension_result','el_competency_dimension_result.idx_competency_dimension_score','el_competency_result.idx_competency_result_exam_score');")" 18
check_eq malformed_required_indexes "$(query "SELECT COUNT(*) FROM (SELECT TABLE_NAME,INDEX_NAME,NON_UNIQUE,GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX) cols FROM information_schema.STATISTICS WHERE TABLE_SCHEMA='$DB' GROUP BY TABLE_NAME,INDEX_NAME,NON_UNIQUE) x WHERE (TABLE_NAME='el_qu' AND INDEX_NAME='uk_qu_question_code' AND NOT (NON_UNIQUE=0 AND cols='question_code')) OR (TABLE_NAME='el_qu' AND INDEX_NAME='uk_qu_dimension_item' AND NOT (NON_UNIQUE=0 AND cols='dimension_id,dimension_item_no')) OR (TABLE_NAME='el_qu' AND INDEX_NAME='idx_qu_dimension_status' AND NOT (NON_UNIQUE=1 AND cols='dimension_id,question_status')) OR (TABLE_NAME='el_exam_competency_question' AND INDEX_NAME='uk_exam_competency_question_source' AND NOT (NON_UNIQUE=0 AND cols='exam_id,source_qu_id')) OR (TABLE_NAME='el_exam_competency_question' AND INDEX_NAME='uk_exam_competency_question_code' AND NOT (NON_UNIQUE=0 AND cols='exam_id,question_code')) OR (TABLE_NAME='el_paper_qu' AND INDEX_NAME='uk_paper_exam_question' AND NOT (NON_UNIQUE=0 AND cols='paper_id,exam_question_id')) OR (TABLE_NAME='el_competency_dimension_result' AND INDEX_NAME='uk_competency_dimension_result' AND NOT (NON_UNIQUE=0 AND cols='paper_id,exam_dimension_id'));")" 0

# Join key collations must match.
check_eq question_dimension_collation_mismatch "$(query "SELECT COUNT(*) FROM information_schema.COLUMNS a JOIN information_schema.COLUMNS b ON b.TABLE_SCHEMA=a.TABLE_SCHEMA WHERE a.TABLE_SCHEMA='$DB' AND a.TABLE_NAME='el_qu' AND a.COLUMN_NAME='dimension_id' AND b.TABLE_NAME='el_competency_dimension' AND b.COLUMN_NAME='id' AND (a.CHARACTER_SET_NAME<>b.CHARACTER_SET_NAME OR a.COLLATION_NAME<>b.COLLATION_NAME);")" 0
check_eq exam_dimension_exam_collation_mismatch "$(query "SELECT COUNT(*) FROM information_schema.COLUMNS a JOIN information_schema.COLUMNS b ON b.TABLE_SCHEMA=a.TABLE_SCHEMA WHERE a.TABLE_SCHEMA='$DB' AND a.TABLE_NAME='el_exam_competency_dimension' AND a.COLUMN_NAME='exam_id' AND b.TABLE_NAME='el_exam' AND b.COLUMN_NAME='id' AND (a.CHARACTER_SET_NAME<>b.CHARACTER_SET_NAME OR a.COLLATION_NAME<>b.COLLATION_NAME);")" 0

# Dimension master integrity.
dimension_summary="$(query "SELECT CONCAT(COUNT(*),'|',COUNT(DISTINCT id),'|',COUNT(DISTINCT code),'|',COUNT(DISTINCT name),'|',COUNT(DISTINCT display_order),'|',MIN(display_order),'|',MAX(display_order),'|',SUM(status=0)) FROM el_competency_dimension;")"
check_eq dimension_summary "$dimension_summary" '48|48|48|48|48|1|48|48'
check_eq invalid_dimension_codes "$(query "SELECT COUNT(*) FROM el_competency_dimension WHERE code<>CONCAT('D',LPAD(display_order,2,'0')) OR id<>CONCAT('competency-d',LPAD(display_order,2,'0'));")" 0
check_eq d42_name "$(query "SELECT name FROM el_competency_dimension WHERE code='D42';")" '权力动机'
check_eq blank_dimension_fields "$(query "SELECT COUNT(*) FROM el_competency_dimension WHERE id='' OR code='' OR name='' OR vird_level='' OR applicable_category='' OR core_meaning='';")" 0
check_eq vird_distribution "$(query "SELECT GROUP_CONCAT(CONCAT(vird_level,':',c) ORDER BY first_order SEPARATOR '|') FROM (SELECT vird_level,COUNT(*) c,MIN(display_order) first_order FROM el_competency_dimension GROUP BY vird_level) x;")" 'Versatility 胜任力:20|Integrity 信念力:10|Resilience 人格力:10|Drive 内驱力:8'
check_eq category_distribution "$(query "SELECT CONCAT(SUM(applicable_category='基层通用'),'|',SUM(applicable_category='管理通用')) FROM el_competency_dimension;")" '10|38'

# Imported question integrity.
question_summary="$(query "SELECT CONCAT(COUNT(*),'|',COUNT(DISTINCT question_code),'|',COUNT(DISTINCT dimension_id),'|',SUM(scoring_direction='forward'),'|',SUM(scoring_direction='reverse'),'|',SUM(question_status=0),'|',SUM(remark='AI测试题-未信效度验证')) FROM el_qu WHERE question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$';")"
check_eq question_summary "$question_summary" '384|384|48|288|96|384|384'
check_eq question_distribution "$(query "SELECT CONCAT(COUNT(*),'|',MIN(c),'|',MAX(c),'|',SUM(c)) FROM (SELECT dimension_id,COUNT(*) c FROM el_qu WHERE question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$' GROUP BY dimension_id) x;")" '48|8|8|384'
check_eq invalid_required_question_fields "$(query "SELECT COUNT(*) FROM el_qu WHERE question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$' AND (dimension_id IS NULL OR dimension_item_no IS NULL OR TRIM(content)='' OR observation_point IS NULL OR TRIM(observation_point)='' OR scoring_direction NOT IN ('forward','reverse') OR question_status NOT IN (0,1));")" 0
check_eq invalid_question_mapping "$(query "SELECT COUNT(*) FROM el_qu q JOIN el_competency_dimension d ON d.id=q.dimension_id WHERE q.question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$' AND (BINARY LEFT(q.question_code,3)<>BINARY d.code OR CAST(RIGHT(q.question_code,2) AS UNSIGNED)<>q.dimension_item_no OR q.dimension_item_no NOT BETWEEN 1 AND 8);")" 0
check_eq orphan_questions "$(query "SELECT COUNT(*) FROM el_qu q LEFT JOIN el_competency_dimension d ON d.id=q.dimension_id WHERE q.question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$' AND d.id IS NULL;")" 0
check_eq duplicate_dimension_items "$(query "SELECT COUNT(*) FROM (SELECT dimension_id,dimension_item_no,COUNT(*) c FROM el_qu WHERE question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$' GROUP BY dimension_id,dimension_item_no HAVING c<>1) x;")" 0
check_eq duplicate_contents "$(query "SELECT COUNT(*) FROM (SELECT content,COUNT(*) c FROM el_qu WHERE question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$' GROUP BY content HAVING c>1) x;")" 0
check_eq competency_answer_rows "$(query "SELECT COUNT(*) FROM el_qu_answer a JOIN el_qu q ON q.id=a.qu_id WHERE q.question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$';")" 0
check_eq competency_repo_rows "$(query "SELECT COUNT(*) FROM el_qu_repo r JOIN el_qu q ON q.id=r.qu_id WHERE q.question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$';")" 0

# Existing legacy data must remain outside competency metadata.
check_eq legacy_rows_with_competency_metadata "$(query "SELECT COUNT(*) FROM el_qu WHERE question_code IS NULL AND (dimension_id IS NOT NULL OR dimension_item_no IS NOT NULL OR observation_point IS NOT NULL OR scoring_direction IS NOT NULL);")" 0
check_eq total_question_rows "$(query "SELECT COUNT(*) FROM el_qu;")" 1239

# Runtime referential consistency (whether empty or populated).
check_eq orphan_exam_dimensions "$(query "SELECT COUNT(*) FROM el_exam_competency_dimension x LEFT JOIN el_exam e ON e.id=x.exam_id LEFT JOIN el_competency_dimension d ON d.id=x.dimension_id WHERE e.id IS NULL OR d.id IS NULL;")" 0
check_eq orphan_exam_questions "$(query "SELECT COUNT(*) FROM el_exam_competency_question q LEFT JOIN el_exam e ON e.id=q.exam_id LEFT JOIN el_exam_competency_dimension d ON d.id=q.exam_dimension_id LEFT JOIN el_qu s ON s.id=q.source_qu_id WHERE e.id IS NULL OR d.id IS NULL OR s.id IS NULL;")" 0
check_eq orphan_dimension_results "$(query "SELECT COUNT(*) FROM el_competency_dimension_result r LEFT JOIN el_paper p ON p.id=r.paper_id LEFT JOIN el_exam_competency_dimension d ON d.id=r.exam_dimension_id WHERE p.id IS NULL OR d.id IS NULL;")" 0
check_eq orphan_results "$(query "SELECT COUNT(*) FROM el_competency_result r LEFT JOIN el_paper p ON p.id=r.paper_id LEFT JOIN el_exam e ON e.id=r.exam_id WHERE p.id IS NULL OR e.id IS NULL;")" 0
check_eq invalid_exam_modes "$(query "SELECT COUNT(*) FROM el_exam WHERE NOT ((assessment_type='legacy' AND scoring_mode='legacy' AND publish_status=1) OR (assessment_type='competency' AND scoring_mode='competency_average' AND publish_status IN (0,1) AND competency_report_audience IN ('frontline_employee','leader')));")" 0
check_eq invalid_runtime_answers "$(query "SELECT COUNT(*) FROM el_paper_qu WHERE exam_question_id IS NOT NULL AND ((raw_answer IS NOT NULL AND raw_answer NOT BETWEEN 1 AND 5) OR (final_score IS NOT NULL AND final_score NOT BETWEEN 1 AND 5));")" 0
check_eq invalid_dimension_result_values "$(query "SELECT COUNT(*) FROM el_competency_dimension_result WHERE answered_question_count<0 OR answered_question_count>total_question_count OR score_sum<0 OR (dimension_score IS NOT NULL AND dimension_score NOT BETWEEN 1 AND 5) OR is_complete NOT IN (0,1);")" 0
check_eq invalid_overall_result_values "$(query "SELECT COUNT(*) FROM el_competency_result WHERE answered_question_count<0 OR answered_question_count>total_question_count OR effective_dimension_count<0 OR overall_score<0 OR (evaluation_average IS NOT NULL AND evaluation_average NOT BETWEEN 1 AND 5) OR evaluation_level NOT IN ('low','average','good','high') OR report_audience NOT IN ('frontline_employee','leader') OR is_complete NOT IN (0,1) OR submit_type NOT IN ('manual','timeout');")" 0

# Produce a canonical database hash matching the local XLSX verifier.
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
query "SELECT CONCAT(HEX(CAST(d.display_order AS CHAR)),'|',HEX(d.name),'|',HEX(q.question_code),'|',HEX(CAST(q.dimension_item_no AS CHAR)),'|',HEX(q.content),'|',HEX(q.observation_point),'|',HEX(q.scoring_direction),'|',HEX(CAST(q.question_status AS CHAR)),'|',HEX(q.remark)) FROM el_qu q JOIN el_competency_dimension d ON d.id=q.dimension_id WHERE q.question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$' ORDER BY q.question_code;" > "$tmp"
echo "canonical_sha256=$(sha256sum "$tmp" | cut -d' ' -f1)"

if [[ "$failures" -ne 0 ]]; then
  echo "COMPETENCY_DATA_INTEGRITY_FAILED failures=$failures" >&2
  exit 1
fi
echo "COMPETENCY_DATA_INTEGRITY_OK"
