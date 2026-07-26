DESC el_mbti_answer;
SELECT '===sample (橡皮 paper) ===' AS x;
SELECT ma.qu_id, eq.content, ma.score_a, ma.score_b, ma.answered
FROM el_mbti_answer ma
JOIN el_qu eq ON eq.id COLLATE utf8mb4_general_ci = ma.qu_id
WHERE ma.paper_id IN (SELECT paper_id FROM el_candidate WHERE name='橡皮' AND telephone='17352955195' LIMIT 1)
ORDER BY CAST(SUBSTRING(eq.content,2) AS UNSIGNED) LIMIT 10;
