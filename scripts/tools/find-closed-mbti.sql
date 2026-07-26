SELECT e.id, e.title, e.is_open, r.code,
  (SELECT COUNT(*) FROM el_tester WHERE exam_id=e.id AND (del_flag IS NULL OR del_flag=0)) tester_cnt,
  (SELECT COUNT(*) FROM el_candidate WHERE exam_id=e.id AND (del_flag IS NULL OR del_flag='0')) candidate_cnt
FROM el_exam e
JOIN el_exam_repo er ON er.exam_id=e.id
JOIN el_repo r ON er.repo_id=r.id
WHERE e.is_open=2 AND r.code LIKE '003%'
ORDER BY tester_cnt DESC LIMIT 5;
