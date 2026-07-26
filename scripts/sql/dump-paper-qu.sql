-- 提取指定 paper 的题目作答（按 sort 升序，输出 V格式）
SET @pid = '1777424935150068442';  -- 李沫 001
SELECT 'sort,content,is_right,actual_score,answered' AS h;
SELECT pq.sort, eq.content, pq.is_right, pq.actual_score, pq.answered
FROM el_paper_qu pq JOIN el_qu eq ON eq.id = pq.qu_id
WHERE pq.paper_id = @pid ORDER BY pq.sort;
