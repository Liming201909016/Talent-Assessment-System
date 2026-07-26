-- 检查每个题库里 qu_repo.sort 与 eq.content 是否完全一一对应
SELECT r.id AS repo_id, r.code AS repo_code, r.title,
  COUNT(*) AS qu_count,
  SUM(CASE WHEN eq.content = CONCAT('V', qr.sort) THEN 1 ELSE 0 END) AS matched,
  SUM(CASE WHEN eq.content != CONCAT('V', qr.sort) THEN 1 ELSE 0 END) AS mismatched
FROM el_qu_repo qr
JOIN el_qu eq ON eq.id = qr.qu_id
JOIN el_repo r ON r.id = qr.repo_id
WHERE r.code LIKE '001%' OR r.code LIKE '002%'
GROUP BY r.id, r.code, r.title
ORDER BY r.code, r.id;
