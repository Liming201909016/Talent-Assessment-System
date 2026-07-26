-- 修复 003 MBTI 题库 paper_qu_answer 的 abc/sort 字段
-- 原 bug: 两个选项 is_right 都=0 时，全被赋 abc='B', sort=1
-- 修复: 同一 (paper_id, qu_id) 下两行按 answer_id 排序，第一行 A/0，第二行 B/1
-- 仅影响 003 题库（is_right 都=0 的题）

SET SQL_SAFE_UPDATES = 0;

-- 第一行（answer_id 较小）→ A
UPDATE el_paper_qu_answer pqa
JOIN (
  SELECT paper_id, qu_id, MIN(answer_id) AS min_aid
  FROM el_paper_qu_answer
  GROUP BY paper_id, qu_id
  HAVING COUNT(*) = 2 AND SUM(is_right) = 0
) t ON pqa.paper_id = t.paper_id AND pqa.qu_id = t.qu_id AND pqa.answer_id = t.min_aid
SET pqa.abc = 'A', pqa.sort = 0;

-- 第二行（answer_id 较大）→ B
UPDATE el_paper_qu_answer pqa
JOIN (
  SELECT paper_id, qu_id, MAX(answer_id) AS max_aid
  FROM el_paper_qu_answer
  GROUP BY paper_id, qu_id
  HAVING COUNT(*) = 2 AND SUM(is_right) = 0
) t ON pqa.paper_id = t.paper_id AND pqa.qu_id = t.qu_id AND pqa.answer_id = t.max_aid
SET pqa.abc = 'B', pqa.sort = 1;

SELECT '=== AFTER FIX (sample 003 paper) ===' AS x;
-- 找一个 003 paper 验证
SET @pid = (SELECT c.paper_id FROM el_candidate c
  JOIN el_exam_repo er ON er.exam_id=c.exam_id
  JOIN el_repo r ON r.id=er.repo_id
  WHERE r.code LIKE '003%' AND c.paper_id IS NOT NULL AND c.end_time IS NOT NULL
  ORDER BY c.update_time DESC LIMIT 1);
SELECT @pid AS pid;
SELECT pqa.qu_id, pqa.answer_id, pqa.abc, pqa.sort
FROM el_paper_qu_answer pqa WHERE paper_id=@pid ORDER BY pqa.qu_id, pqa.sort LIMIT 8;

SET SQL_SAFE_UPDATES = 1;
