SELECT 'cand' AS src, c.id, c.name, c.telephone, c.paper_id, r.code FROM el_candidate c
  LEFT JOIN el_exam_repo er ON er.exam_id=c.exam_id
  LEFT JOIN el_repo r ON r.id=er.repo_id
WHERE c.name='INFJ' OR c.telephone='19352046391' LIMIT 5;

SET @pid = (SELECT c.paper_id FROM el_candidate c
  JOIN el_exam_repo er ON er.exam_id=c.exam_id
  JOIN el_repo r ON r.id=er.repo_id
  WHERE r.code LIKE '003%' AND c.paper_id IS NOT NULL AND c.end_time IS NOT NULL
  ORDER BY c.update_time DESC LIMIT 1);
SELECT @pid AS pid;
SELECT COUNT(*) AS pqa_cnt FROM el_paper_qu_answer WHERE paper_id=@pid;
SELECT COUNT(*) AS pq_cnt FROM el_paper_qu WHERE paper_id=@pid;
SELECT COUNT(*) AS mbti_cnt FROM el_mbti_answer WHERE paper_id=@pid;
SELECT pqa.qu_id, pqa.answer_id, pqa.abc, pqa.is_right, pqa.checked, pqa.sort
FROM el_paper_qu_answer pqa WHERE paper_id=@pid ORDER BY pqa.sort LIMIT 4;
SELECT 'mbti_answer sample' AS x;
SELECT ma.qu_id, eq.content, ma.score_a, ma.score_b, ma.answered
FROM el_mbti_answer ma JOIN el_qu eq ON eq.id COLLATE utf8mb4_general_ci = ma.qu_id
WHERE ma.paper_id=@pid ORDER BY CAST(SUBSTRING(eq.content,2) AS UNSIGNED) LIMIT 4;
