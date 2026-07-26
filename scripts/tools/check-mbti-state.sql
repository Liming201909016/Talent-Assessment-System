-- 1) Liming 这条记录的关键字段
SELECT 'el_candidate' AS tbl, c.id, c.paper_id, c.name, c.end_time, c.pdf_flag, c.pdf_path, p.state AS paper_state, p.user_time
FROM el_candidate c LEFT JOIN el_paper p ON p.id=c.paper_id
WHERE c.paper_id='1777217353411783509';

-- 2) 所有 MBTI 已交卷但 end_time 为 null 的 candidate
SELECT c.id, c.paper_id, c.name, c.end_time, c.pdf_flag, p.state, (SELECT COUNT(*) FROM el_mbti_answer WHERE paper_id=p.id AND answered=1) ans
FROM el_candidate c JOIN el_paper p ON p.id=c.paper_id
JOIN el_exam_repo er ON er.exam_id=c.exam_id
JOIN el_repo r ON er.repo_id=r.id
WHERE r.code LIKE '003%' AND p.state=2 AND c.end_time IS NULL
LIMIT 5;
