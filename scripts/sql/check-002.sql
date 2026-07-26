SET @pid = (SELECT paper_id FROM el_candidate WHERE id='1777442919314');
SELECT @pid;
SELECT pq.sort, eq.content, pq.is_right, pq.actual_score, pq.score, pq.answered, LEFT(pq.answer,30) AS answer
FROM el_paper_qu pq JOIN el_qu eq ON eq.id=pq.qu_id
WHERE pq.paper_id=@pid ORDER BY pq.sort LIMIT 15;
