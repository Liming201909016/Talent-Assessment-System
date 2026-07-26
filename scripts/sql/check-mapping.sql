-- 高分 paper=1777355490227515031, 取所有 paper_qu, JOIN qu_repo 拿到题库编号
-- 比较 paper_qu.sort（答题序）vs qu_repo.sort（题库编号） vs eq.content（V编码）
SELECT pq.sort AS 答题序, qr.sort AS 题库编号, eq.content AS V编码, pq.is_right
FROM el_paper_qu pq
JOIN el_qu eq ON eq.id=pq.qu_id
JOIN el_qu_repo qr ON qr.qu_id=pq.qu_id
WHERE pq.paper_id='1777355490227515031'
  AND qr.repo_id='1797154092426985473'
ORDER BY qr.sort
LIMIT 15;
