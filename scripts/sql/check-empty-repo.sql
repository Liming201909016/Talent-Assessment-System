-- 截图3行空题库 创建时间: 2026-02-03 15:11, 2026-01-29 15:46, 2026-01-27 16:41
SELECT eq.id, eq.content, eq.create_time, qr.repo_id, r.title AS repo_title
FROM el_qu eq
LEFT JOIN el_qu_repo qr ON qr.qu_id=eq.id
LEFT JOIN el_repo r ON r.id=qr.repo_id
WHERE eq.create_time IN ('2026-02-03 15:11:18','2026-01-29 15:46:23','2026-01-27 16:41:36');

SELECT '=== 关联了但题库不存在的"挂空"题 ===' AS x;
SELECT eq.id, eq.content, eq.create_time, qr.repo_id
FROM el_qu eq
JOIN el_qu_repo qr ON qr.qu_id=eq.id
LEFT JOIN el_repo r ON r.id=qr.repo_id
WHERE r.id IS NULL
LIMIT 20;

SELECT '=== 这些题在试卷中是否被引用过 ===' AS x;
SELECT COUNT(DISTINCT pq.qu_id) AS used_count
FROM el_paper_qu pq
JOIN el_qu_repo qr ON qr.qu_id=pq.qu_id
LEFT JOIN el_repo r ON r.id=qr.repo_id
WHERE r.id IS NULL;
