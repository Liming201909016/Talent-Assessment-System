-- 测试数据清理 - 查看阶段
SELECT 'el_exam' AS tbl, COUNT(*) AS cnt FROM el_exam
UNION ALL SELECT 'el_paper', COUNT(*) FROM el_paper
UNION ALL SELECT 'el_paper_qu', COUNT(*) FROM el_paper_qu
UNION ALL SELECT 'el_paper_qu_answer', COUNT(*) FROM el_paper_qu_answer
UNION ALL SELECT 'el_mbti_answer', COUNT(*) FROM el_mbti_answer
UNION ALL SELECT 'el_tester', COUNT(*) FROM el_tester
UNION ALL SELECT 'el_candidate', COUNT(*) FROM el_candidate
UNION ALL SELECT 'el_repo', COUNT(*) FROM el_repo
UNION ALL SELECT 'el_qu', COUNT(*) FROM el_qu
UNION ALL SELECT 'el_qu_repo', COUNT(*) FROM el_qu_repo;

-- 列出所有 exam（含创建时间，便于辨识哪些是测试产生的）
SELECT id, title, state, is_open, total_time,
       DATE_FORMAT(create_time, '%Y-%m-%d %H:%i') AS created
FROM el_exam
ORDER BY create_time DESC
LIMIT 30;
