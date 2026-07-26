SELECT 'cand' AS src, c.id, c.name, r.code, c.pdf_path
FROM el_candidate c
JOIN el_exam_repo er ON er.exam_id=c.exam_id
JOIN el_repo r ON r.id=er.repo_id
WHERE c.pdf_path LIKE '%uploadPath%' AND c.pdf_flag=1
  AND r.code LIKE '001%'
ORDER BY c.update_time DESC LIMIT 3;
SELECT 'tester' AS src, t.id, t.name, r.code, t.pdf_path
FROM el_tester t
JOIN el_exam_repo er ON er.exam_id=t.exam_id
JOIN el_repo r ON r.id=er.repo_id
WHERE t.pdf_path LIKE '%uploadPath%' AND t.pdf_flag=1
  AND r.code LIKE '001%'
ORDER BY t.update_time DESC LIMIT 3;
