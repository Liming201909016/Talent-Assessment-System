SELECT abc, COUNT(*) AS cnt FROM el_paper_qu_answer pqa
JOIN el_paper p ON p.id=pqa.paper_id
JOIN el_exam_repo er ON er.exam_id=p.exam_id
JOIN el_repo r ON r.id=er.repo_id
WHERE r.code LIKE '003%' GROUP BY abc;
