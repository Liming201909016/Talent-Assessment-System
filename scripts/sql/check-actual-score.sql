SELECT pq.sort, eq.content, pq.is_right, pq.actual_score, pq.score, pq.answered, pq.answer
FROM el_paper_qu pq JOIN el_qu eq ON eq.id=pq.qu_id
WHERE pq.paper_id='1777424935150068442' ORDER BY pq.sort LIMIT 12;

SELECT '---002 sample---' AS x;
SELECT pq.sort, eq.content, pq.is_right, pq.actual_score, pq.score, pq.answered, LEFT(pq.answer,50) AS answer
FROM el_paper_qu pq JOIN el_qu eq ON eq.id=pq.qu_id
WHERE pq.paper_id IN (SELECT paper_id FROM el_candidate WHERE id='1777442919314' LIMIT 1)
ORDER BY pq.sort LIMIT 12;
