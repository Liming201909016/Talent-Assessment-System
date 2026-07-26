#!/usr/bin/env bash
set -euo pipefail

sudo -n mysql element < /tmp/competency_001_schema.sql
sudo -n mysql element < /tmp/competency_002_dimensions.sql
sudo -n mysql element < /tmp/competency_003_questions.sql
sudo -n mysql element < /tmp/competency_004_runtime.sql

sudo -n mysql element -Nse "
SELECT CONCAT('dispatch_columns=', COUNT(*)) FROM information_schema.columns WHERE table_schema='element' AND table_name='el_exam' AND column_name IN ('assessment_type','scoring_mode','competency_report_audience','publish_status','published_at','published_by');
SELECT CONCAT('dispatch_index_columns=', COUNT(*)) FROM information_schema.statistics WHERE table_schema='element' AND table_name='el_exam' AND index_name='idx_exam_assessment_publish';
SELECT CONCAT('competency_tables=', COUNT(*)) FROM information_schema.tables WHERE table_schema='element' AND table_name IN ('el_competency_dimension','el_exam_competency_dimension');
SELECT CONCAT('dimension_count=', COUNT(*), ',min=', MIN(display_order), ',max=', MAX(display_order), ',codes=', COUNT(DISTINCT code), ',names=', COUNT(DISTINCT name)) FROM el_competency_dimension;
SELECT CONCAT('d42=', code, ':', name, ':', display_order) FROM el_competency_dimension WHERE id='competency-d42';
SELECT CONCAT('legacy_exam_count=', COUNT(*)) FROM el_exam WHERE assessment_type='legacy' AND scoring_mode='legacy' AND publish_status=1;
SELECT CONCAT('invalid_exam_count=', COUNT(*)) FROM el_exam WHERE NOT (assessment_type='legacy' AND scoring_mode='legacy' AND publish_status=1);
SELECT CONCAT('question_columns=', COUNT(*)) FROM information_schema.columns WHERE table_schema='element' AND table_name='el_qu' AND column_name IN ('question_code','dimension_id','dimension_item_no','observation_point','scoring_direction','question_status');
SELECT CONCAT('question_index_columns=', COUNT(*)) FROM information_schema.statistics WHERE table_schema='element' AND table_name='el_qu' AND index_name IN ('uk_qu_question_code','uk_qu_dimension_item','idx_qu_dimension_status');
SELECT CONCAT('question_dimension_collation=', COLLATION_NAME) FROM information_schema.columns WHERE table_schema='element' AND table_name='el_qu' AND column_name='dimension_id';
SELECT CONCAT('dimension_id_collation=', COLLATION_NAME) FROM information_schema.columns WHERE table_schema='element' AND table_name='el_competency_dimension' AND column_name='id';
SELECT CONCAT('legacy_questions_with_competency_metadata=', COUNT(*)) FROM el_qu WHERE question_code IS NOT NULL OR dimension_id IS NOT NULL OR dimension_item_no IS NOT NULL OR observation_point IS NOT NULL OR scoring_direction IS NOT NULL;
SELECT CONCAT('runtime_tables=', COUNT(*)) FROM information_schema.tables WHERE table_schema='element' AND table_name IN ('el_exam_competency_question','el_competency_dimension_result','el_competency_result');
SELECT CONCAT('paper_question_columns=', COUNT(*)) FROM information_schema.columns WHERE table_schema='element' AND table_name='el_paper_qu' AND column_name IN ('exam_question_id','raw_answer','final_score');
SELECT CONCAT('runtime_indexes=', COUNT(DISTINCT index_name)) FROM information_schema.statistics WHERE table_schema='element' AND ((table_name='el_exam_competency_question' AND index_name IN ('uk_exam_competency_question_source','uk_exam_competency_question_code','idx_exam_competency_question_dimension')) OR (table_name='el_paper_qu' AND index_name IN ('uk_paper_exam_question','idx_paper_question_answered')) OR (table_name='el_paper' AND index_name='idx_paper_state_limit_time') OR (table_name='el_competency_dimension_result' AND index_name IN ('uk_competency_dimension_result','idx_competency_dimension_score')) OR (table_name='el_competency_result' AND index_name='idx_competency_result_exam_score'));
"
