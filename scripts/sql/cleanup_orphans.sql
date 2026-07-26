-- ====================================================================
-- 深度清理：去除所有孤儿数据
-- ====================================================================
-- 1. el_paper 关联的 exam 已不存在 → 删 paper
-- 2. el_paper_qu 关联的 paper 已不存在 → 删 paper_qu
-- 3. el_paper_qu_answer 同上
-- 4. el_mbti_answer 同上
-- 5. el_tester 关联的 exam 已不存在 → 删 tester
-- 6. el_candidate 同上

START TRANSACTION;

-- 孤儿 paper（exam 不存在）
DELETE FROM el_paper WHERE exam_id NOT IN (SELECT id FROM el_exam);

-- 孤儿 paper_qu_answer / paper_qu / mbti_answer
DELETE FROM el_paper_qu_answer WHERE paper_id NOT IN (SELECT id FROM el_paper);
DELETE FROM el_paper_qu WHERE paper_id NOT IN (SELECT id FROM el_paper);
DELETE FROM el_mbti_answer WHERE paper_id NOT IN (SELECT id FROM el_paper);

-- 孤儿 tester / candidate
DELETE FROM el_tester WHERE exam_id NOT IN (SELECT id FROM el_exam);
DELETE FROM el_candidate WHERE exam_id NOT IN (SELECT id FROM el_exam);

-- 老旧测试 paper（创建时间 > 7 天前但 exam 还在 — 可能是真实业务，保留）
-- 这里不删 — 仅删孤儿

SELECT 'AFTER DEEP CLEAN' AS step;
SELECT 'el_exam' AS tbl, COUNT(*) AS cnt FROM el_exam
UNION ALL SELECT 'el_paper', COUNT(*) FROM el_paper
UNION ALL SELECT 'el_paper_qu', COUNT(*) FROM el_paper_qu
UNION ALL SELECT 'el_paper_qu_answer', COUNT(*) FROM el_paper_qu_answer
UNION ALL SELECT 'el_mbti_answer', COUNT(*) FROM el_mbti_answer
UNION ALL SELECT 'el_tester', COUNT(*) FROM el_tester
UNION ALL SELECT 'el_candidate', COUNT(*) FROM el_candidate;

COMMIT;
