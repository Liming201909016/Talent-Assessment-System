-- 输出 JSON 格式：每个 V题的 is_right=1/0 的选项内容
-- 用高分这个 paper 的 qu_id（保证题库版本一致）
SELECT JSON_OBJECTAGG(content, opts) AS data FROM (
  SELECT eq.content,
    JSON_ARRAYAGG(JSON_OBJECT('text', ea.content, 'is_right', ea.is_right)) AS opts
  FROM el_paper_qu pq
  JOIN el_qu eq ON eq.id = pq.qu_id
  LEFT JOIN el_qu_answer ea ON ea.qu_id = pq.qu_id
  WHERE pq.paper_id = '1777355490227515031'
  GROUP BY eq.content
) t;
