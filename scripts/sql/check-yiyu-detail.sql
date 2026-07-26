SET @pid = '1777355490227515031';

-- 拼出每道题的"考生选项内容"（通过 paper_qu.is_right 找到对应 ans 的 content）
SELECT
  pq.sort + 1 AS 题序,
  eq.content AS V编号,
  eq.id AS qu_id_long,
  pq.is_right AS db_is_right,
  pq.actual_score AS db_score,
  GROUP_CONCAT(DISTINCT
    CASE WHEN ea.is_right = pq.is_right THEN CONCAT('[选中] ', ea.content) ELSE NULL END
    SEPARATOR ' / ') AS 考生选了,
  GROUP_CONCAT(DISTINCT
    CASE WHEN ea.is_right != pq.is_right THEN CONCAT('[未选] ', ea.content) ELSE NULL END
    SEPARATOR ' / ') AS 另一选项
FROM el_paper_qu pq
JOIN el_qu eq ON eq.id = pq.qu_id
LEFT JOIN el_qu_answer ea ON ea.qu_id = pq.qu_id
WHERE pq.paper_id = @pid
  AND eq.content IN ('V12','V21','V26','V38','V49','V70','V77','V86')
GROUP BY pq.sort, eq.content, eq.id, pq.is_right, pq.actual_score
ORDER BY CAST(SUBSTRING(eq.content,2) AS UNSIGNED);
