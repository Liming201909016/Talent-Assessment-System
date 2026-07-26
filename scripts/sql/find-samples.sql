SELECT 'tester 001' AS lbl;
SELECT t.id, t.name, t.paper_id, r.code FROM el_tester t
  JOIN el_exam_repo er ON er.exam_id=t.exam_id
  JOIN el_repo r ON r.id=er.repo_id
WHERE t.paper_id IS NOT NULL AND t.end_time IS NOT NULL AND r.code LIKE '001%'
ORDER BY t.end_time DESC LIMIT 3;

SELECT 'candidate 001' AS lbl;
SELECT c.id, c.name, c.paper_id, r.code FROM el_candidate c
  JOIN el_exam_repo er ON er.exam_id=c.exam_id
  JOIN el_repo r ON r.id=er.repo_id
WHERE c.paper_id IS NOT NULL AND c.end_time IS NOT NULL AND r.code LIKE '001%'
ORDER BY c.end_time DESC LIMIT 3;

SELECT 'candidate 002' AS lbl;
SELECT c.id, c.name, c.paper_id, r.code FROM el_candidate c
  JOIN el_exam_repo er ON er.exam_id=c.exam_id
  JOIN el_repo r ON r.id=er.repo_id
WHERE c.paper_id IS NOT NULL AND c.end_time IS NOT NULL AND r.code LIKE '002%'
ORDER BY c.end_time DESC LIMIT 3;
