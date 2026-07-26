SELECT c.id, c.paper_id, c.name FROM el_candidate c
JOIN el_exam e ON c.exam_id=e.id
JOIN el_exam_repo er ON er.exam_id=e.id
JOIN el_repo r ON er.repo_id=r.id
WHERE r.code LIKE '003%' AND c.paper_id IS NOT NULL
ORDER BY c.id DESC LIMIT 5\G
