-- 检查 39.106.61.48 的 003 MBTI paper_qu_answer 是否有 abc 全 B 的问题
SELECT '=== 003 MBTI paper_qu_answer ABC 分布 ===' AS x;
SELECT abc, COUNT(*) AS cnt FROM el_paper_qu_answer pqa
JOIN el_paper p ON p.id=pqa.paper_id
JOIN el_exam_repo er ON er.exam_id=p.exam_id
JOIN el_repo r ON r.id=er.repo_id
WHERE r.code LIKE '003%'
GROUP BY abc;

SELECT '=== 同一 paper+qu 两行 abc 是否相同 ===' AS x;
SELECT
  COUNT(*) AS group_cnt,
  SUM(CASE WHEN cnt=2 AND abc_min=abc_max THEN 1 ELSE 0 END) AS same_abc_groups,
  SUM(CASE WHEN cnt=2 AND abc_min!=abc_max THEN 1 ELSE 0 END) AS diff_abc_groups
FROM (
  SELECT pqa.paper_id, pqa.qu_id, COUNT(*) AS cnt, MIN(pqa.abc) AS abc_min, MAX(pqa.abc) AS abc_max
  FROM el_paper_qu_answer pqa
  JOIN el_paper p ON p.id=pqa.paper_id
  JOIN el_exam_repo er ON er.exam_id=p.exam_id
  JOIN el_repo r ON r.id=er.repo_id
  WHERE r.code LIKE '003%'
  GROUP BY pqa.paper_id, pqa.qu_id
) t;

SELECT '=== 001 心理特质（对照） ===' AS x;
SELECT
  COUNT(*) AS group_cnt,
  SUM(CASE WHEN cnt=2 AND abc_min=abc_max THEN 1 ELSE 0 END) AS same_abc_groups,
  SUM(CASE WHEN cnt=2 AND abc_min!=abc_max THEN 1 ELSE 0 END) AS diff_abc_groups
FROM (
  SELECT pqa.paper_id, pqa.qu_id, COUNT(*) AS cnt, MIN(pqa.abc) AS abc_min, MAX(pqa.abc) AS abc_max
  FROM el_paper_qu_answer pqa
  JOIN el_paper p ON p.id=pqa.paper_id
  JOIN el_exam_repo er ON er.exam_id=p.exam_id
  JOIN el_repo r ON r.id=er.repo_id
  WHERE r.code LIKE '001%'
  GROUP BY pqa.paper_id, pqa.qu_id
) t;
