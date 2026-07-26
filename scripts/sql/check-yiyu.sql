SET @pid = '1777355490227515031';
-- 抑郁公式: V12+V26+V38+1-V49+V70+1-V77+1-V86+V21
SELECT '=== 抑郁涉及题 ===' AS x;
SELECT pq.sort, eq.content, pq.is_right, pq.actual_score, pq.score, pq.answered, pq.answer
FROM el_paper_qu pq JOIN el_qu eq ON eq.id=pq.qu_id
WHERE pq.paper_id=@pid AND eq.content IN ('V12','V21','V26','V38','V49','V70','V77','V86')
ORDER BY CAST(SUBSTRING(eq.content,2) AS UNSIGNED);

-- 看题目内容（el_qu 中真实的题干）和答案
SELECT '=== 题目+答案选项 ===' AS x;
SELECT eq.content AS qu_v, eq.id AS qu_id, ea.id AS ans_id, ea.content AS ans_text, ea.is_right AS ans_is_right
FROM el_qu eq JOIN el_qu_answer ea ON ea.qu_id = eq.id
WHERE eq.content IN ('V12','V21','V26','V38','V49','V70','V77','V86')
ORDER BY CAST(SUBSTRING(eq.content,2) AS UNSIGNED), ea.id;

-- 看 paper_qu.answer 是哪个 ans_id
SELECT '=== 实际作答 ===' AS x;
SELECT pq.sort, eq.content AS qu_v, pq.is_right, pq.answer, pq.actual_score
FROM el_paper_qu pq JOIN el_qu eq ON eq.id=pq.qu_id
WHERE pq.paper_id=@pid AND eq.content IN ('V12','V21','V26','V38','V49','V70','V77','V86');
