SELECT @pid := paper_id FROM el_candidate WHERE name='哈哈' AND exam_id='2026042000040000001' LIMIT 1;
SELECT pq.sort, LEFT(eq.content, 5) AS content, pq.qu_type
FROM el_paper_qu pq JOIN el_qu eq ON pq.qu_id=eq.id
WHERE pq.paper_id = @pid
ORDER BY pq.sort ASC LIMIT 15;
