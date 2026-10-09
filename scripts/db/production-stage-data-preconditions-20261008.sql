-- Read-only, aggregate-only migration preconditions; no personal values.
SET SESSION MAX_EXECUTION_TIME=10000;
START TRANSACTION READ ONLY;
SELECT 'P','mbti-type-over4',COUNT(*) FROM el_tester WHERE CHAR_LENGTH(mbti_type)>4;
SELECT 'P','mbti-scores-over200',COUNT(*) FROM el_tester WHERE CHAR_LENGTH(mbti_scores)>200;
SELECT 'P','source-question-id-over32',COUNT(*) FROM el_qu WHERE OCTET_LENGTH(id)>32;
SELECT 'P','source-option-id-over32',COUNT(*) FROM el_qu_answer WHERE OCTET_LENGTH(id)>32;
SELECT 'P','source-question-id-nonascii',COUNT(*) FROM el_qu WHERE OCTET_LENGTH(id)<>CHAR_LENGTH(id);
SELECT 'P','source-option-id-nonascii',COUNT(*) FROM el_qu_answer WHERE OCTET_LENGTH(id)<>CHAR_LENGTH(id);
SELECT 'P','competency-source-count',COUNT(*) FROM el_qu WHERE dimension_id IS NOT NULL;
SELECT 'P','source-duplicate-dimension-item',COUNT(*) FROM (SELECT dimension_id,dimension_item_no FROM el_qu WHERE dimension_id IS NOT NULL GROUP BY dimension_id,dimension_item_no HAVING COUNT(*)>1) d;
SELECT 'P','report-duplicate-paper-content',COUNT(*) FROM (SELECT paper_id,content_version FROM el_competency_report GROUP BY paper_id,content_version HAVING COUNT(*)>1) d;
SELECT 'P','competency-result-missing-paper',COUNT(*) FROM el_competency_result r LEFT JOIN el_paper p ON p.id=r.paper_id WHERE p.id IS NULL;
SELECT 'P','competency-report-missing-paper',COUNT(*) FROM el_competency_report r LEFT JOIN el_paper p ON p.id=r.paper_id WHERE p.id IS NULL;
SELECT 'P','competency-snapshot-missing-exam',COUNT(*) FROM el_exam_competency_question q LEFT JOIN el_exam e ON e.id=q.exam_id WHERE e.id IS NULL;
SELECT 'P','pending-competency-state0',COUNT(*) FROM el_paper p JOIN el_exam e ON e.id=p.exam_id WHERE p.state=0 AND e.assessment_type='competency' AND p.limit_time<=NOW();
SELECT 'P','mng-parent-precondition',COUNT(*) FROM information_schema.COLUMNS c JOIN information_schema.COLUMNS p ON p.TABLE_SCHEMA=c.TABLE_SCHEMA AND p.TABLE_NAME='el_exam' AND p.COLUMN_NAME='id'
WHERE c.TABLE_SCHEMA=DATABASE() AND c.TABLE_NAME IN ('el_exam','el_paper','el_paper_qu') AND c.COLUMN_NAME='id' AND c.DATA_TYPE='varchar' AND c.CHARACTER_MAXIMUM_LENGTH BETWEEN 36 AND 64 AND c.IS_NULLABLE='NO' AND c.CHARACTER_SET_NAME='utf8mb4' AND c.COLLATION_NAME=p.COLLATION_NAME;
COMMIT;