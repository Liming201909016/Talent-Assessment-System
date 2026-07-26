-- 题库里"题目编号"= el_qu.sort（题目管理界面看到的 1,2,3...）
-- 取该题库前 12 题：题目编号 vs V内容编码
SELECT eq.sort+1 AS 题目编号, eq.content AS V编码, LEFT(ea.content,30) AS 选项A
FROM el_qu eq
LEFT JOIN el_qu_answer ea ON ea.qu_id=eq.id AND ea.id LIKE '%V%'
WHERE eq.repo_id='1797154092426985473'
GROUP BY eq.id, eq.sort, eq.content
ORDER BY eq.sort
LIMIT 15;
