SELECT pq.sort, eq.content, pq.is_right, pq.actual_score, pq.answer, eq.title
FROM el_paper_qu pq JOIN el_qu eq ON eq.id=pq.qu_id
WHERE pq.paper_id='1777355490227515031' ORDER BY pq.sort LIMIT 5;

SELECT '=== 002 sample ===' AS x;
SELECT pq.sort, eq.content, pq.is_right, pq.actual_score, pq.answer
FROM el_paper_qu pq JOIN el_qu eq ON eq.id=pq.qu_id
WHERE pq.paper_id IN (SELECT paper_id FROM el_candidate WHERE id='1777442919314' LIMIT 1)
ORDER BY pq.sort LIMIT 5;
