-- 1) 总题目数 + 关联题库数
SELECT COUNT(*) AS total_qu FROM el_qu;
SELECT COUNT(DISTINCT qu_id) AS qu_with_repo FROM el_qu_repo;

-- 2) 没关联任何题库的题目（"孤儿题"）
SELECT '=== 孤儿题（el_qu 中未在 el_qu_repo 出现）===' AS x;
SELECT eq.id, eq.content, eq.create_time, eq.update_time
FROM el_qu eq
LEFT JOIN el_qu_repo qr ON qr.qu_id=eq.id
WHERE qr.qu_id IS NULL
ORDER BY eq.update_time DESC LIMIT 20;

-- 3) 此类孤儿题在试卷里有没有被引用
SELECT '=== 孤儿题被引用情况 ===' AS x;
SELECT COUNT(DISTINCT pq.qu_id) AS orphan_qu_in_paper
FROM el_paper_qu pq
WHERE NOT EXISTS (SELECT 1 FROM el_qu_repo qr WHERE qr.qu_id=pq.qu_id);
