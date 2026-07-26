-- 查一道 002 题的所有选项 (likert)
SELECT eq.content AS V, ea.id AS ans_id, ea.content, ea.is_right
FROM el_qu eq JOIN el_qu_answer ea ON ea.qu_id=eq.id
WHERE eq.content='V21' AND eq.id IN (
  SELECT pq.qu_id FROM el_paper_qu pq
  JOIN el_candidate c ON c.paper_id=pq.paper_id
  WHERE c.id='1777442919314' AND pq.sort < 30
)
ORDER BY ea.id;
