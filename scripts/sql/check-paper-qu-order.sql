-- 检查考生答题顺序是否与题库编号顺序一致
SELECT '=== 003 MBTI: paper_qu.sort vs qu_repo.sort ===' AS x;
SELECT p.id AS paper_id,
  SUM(CASE WHEN pq.sort = qr.sort - 1 THEN 1 ELSE 0 END) AS matched,
  COUNT(*) AS total,
  CASE WHEN SUM(CASE WHEN pq.sort = qr.sort - 1 THEN 1 ELSE 0 END) = COUNT(*) THEN '一致' ELSE '不一致' END AS 结论
FROM el_paper p
JOIN el_exam_repo er ON er.exam_id = p.exam_id
JOIN el_repo r ON r.id = er.repo_id
JOIN el_paper_qu pq ON pq.paper_id = p.id
JOIN el_qu_repo qr ON qr.qu_id = pq.qu_id AND qr.repo_id = er.repo_id
WHERE r.code LIKE '003%'
GROUP BY p.id
LIMIT 5;

SELECT '=== 001 心理特质 ===' AS x;
SELECT p.id AS paper_id,
  SUM(CASE WHEN pq.sort = qr.sort - 1 THEN 1 ELSE 0 END) AS matched,
  COUNT(*) AS total,
  CASE WHEN SUM(CASE WHEN pq.sort = qr.sort - 1 THEN 1 ELSE 0 END) = COUNT(*) THEN '一致' ELSE '不一致' END AS 结论
FROM el_paper p
JOIN el_exam_repo er ON er.exam_id = p.exam_id
JOIN el_repo r ON r.id = er.repo_id
JOIN el_paper_qu pq ON pq.paper_id = p.id
JOIN el_qu_repo qr ON qr.qu_id = pq.qu_id AND qr.repo_id = er.repo_id
WHERE r.code LIKE '001%'
GROUP BY p.id
LIMIT 5;

SELECT '=== 002 管理特质 ===' AS x;
SELECT p.id AS paper_id,
  SUM(CASE WHEN pq.sort = qr.sort - 1 THEN 1 ELSE 0 END) AS matched,
  COUNT(*) AS total,
  CASE WHEN SUM(CASE WHEN pq.sort = qr.sort - 1 THEN 1 ELSE 0 END) = COUNT(*) THEN '一致' ELSE '不一致' END AS 结论
FROM el_paper p
JOIN el_exam_repo er ON er.exam_id = p.exam_id
JOIN el_repo r ON r.id = er.repo_id
JOIN el_paper_qu pq ON pq.paper_id = p.id
JOIN el_qu_repo qr ON qr.qu_id = pq.qu_id AND qr.repo_id = er.repo_id
WHERE r.code LIKE '002%'
GROUP BY p.id
LIMIT 5;
