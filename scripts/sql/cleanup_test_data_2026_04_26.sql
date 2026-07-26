-- ====================================================================
-- 测试数据清理脚本 (2026-04-26 当日测试残留)
-- ====================================================================
-- 策略：删除 2026-04-26 创建的测试 exam 及所有关联数据
-- 保留：4-22 之前的真实业务数据
--
-- 测试 exam 特征：
--   - 标题：'Zero' / 'DCF01测评' / 'TypeTest' / '管理干部能力测评' / '2026年春季心理健康筛查'
--   - create_time >= '2026-04-26 00:00:00'
--
-- 安全：
--   1. SET autocommit = 0; 进入事务
--   2. 先 SELECT 看到将删什么
--   3. 满意后 COMMIT，否则 ROLLBACK

SET autocommit = 0;
START TRANSACTION;

-- Step 1: 找出所有测试 exam ID
CREATE TEMPORARY TABLE _trash_exam_ids AS
SELECT id FROM el_exam
WHERE create_time >= '2026-04-26 00:00:00'
  AND (
    title IN ('Zero', 'TypeTest', 'DCF01测评',
              '2026年春季心理健康筛查', '管理干部能力测评')
    OR title LIKE 'TypeTest%'
    OR title LIKE '%test%'
    OR title LIKE '%Test%'
  );

SELECT 'trash exams' AS step, COUNT(*) AS cnt FROM _trash_exam_ids;
SELECT id, title, create_time FROM el_exam WHERE id IN (SELECT id FROM _trash_exam_ids);

-- Step 2: 找出所有关联 paper ID
CREATE TEMPORARY TABLE _trash_paper_ids AS
SELECT id FROM el_paper WHERE exam_id IN (SELECT id FROM _trash_exam_ids);

SELECT 'trash papers' AS step, COUNT(*) AS cnt FROM _trash_paper_ids;

-- Step 3: 找出所有关联 tester ID
CREATE TEMPORARY TABLE _trash_tester_ids AS
SELECT id FROM el_tester WHERE exam_id IN (SELECT id FROM _trash_exam_ids);

SELECT 'trash testers' AS step, COUNT(*) AS cnt FROM _trash_tester_ids;

-- Step 4: 删除（按依赖顺序）
DELETE FROM el_paper_qu_answer WHERE paper_id IN (SELECT id FROM _trash_paper_ids);
DELETE FROM el_paper_qu WHERE paper_id IN (SELECT id FROM _trash_paper_ids);
DELETE FROM el_mbti_answer WHERE paper_id IN (SELECT id FROM _trash_paper_ids);
DELETE FROM el_paper WHERE id IN (SELECT id FROM _trash_paper_ids);
DELETE FROM el_tester WHERE id IN (SELECT id FROM _trash_tester_ids);
DELETE FROM el_exam_repo WHERE exam_id IN (SELECT id FROM _trash_exam_ids);
DELETE FROM el_exam_depart WHERE exam_id IN (SELECT id FROM _trash_exam_ids);
DELETE FROM el_exam WHERE id IN (SELECT id FROM _trash_exam_ids);

-- Step 5: 清理孤儿 paper（exam 不存在的 paper）
DELETE FROM el_paper_qu_answer
  WHERE paper_id NOT IN (SELECT id FROM el_paper);
DELETE FROM el_paper_qu
  WHERE paper_id NOT IN (SELECT id FROM el_paper);
DELETE FROM el_mbti_answer
  WHERE paper_id NOT IN (SELECT id FROM el_paper);

-- Step 6: 清理测试 repo（如果有）
CREATE TEMPORARY TABLE _trash_repo_ids AS
SELECT id FROM el_repo
WHERE create_time >= '2026-04-26 00:00:00'
  AND (
    code LIKE '002M%'
    OR code LIKE 'TEST%'
    OR title LIKE '%test%'
    OR title LIKE '%TypeTest%'
  );

SELECT 'trash repos' AS step, COUNT(*) AS cnt FROM _trash_repo_ids;

DELETE FROM el_qu_repo WHERE repo_id IN (SELECT id FROM _trash_repo_ids);
DELETE FROM el_qu WHERE id IN (
  SELECT qu_id FROM el_qu_repo WHERE repo_id IN (SELECT id FROM _trash_repo_ids)
);
DELETE FROM el_repo WHERE id IN (SELECT id FROM _trash_repo_ids);

-- Step 7: 报告
SELECT 'AFTER' AS step;
SELECT 'el_exam' AS tbl, COUNT(*) AS cnt FROM el_exam
UNION ALL SELECT 'el_paper', COUNT(*) FROM el_paper
UNION ALL SELECT 'el_paper_qu', COUNT(*) FROM el_paper_qu
UNION ALL SELECT 'el_paper_qu_answer', COUNT(*) FROM el_paper_qu_answer
UNION ALL SELECT 'el_mbti_answer', COUNT(*) FROM el_mbti_answer
UNION ALL SELECT 'el_tester', COUNT(*) FROM el_tester
UNION ALL SELECT 'el_repo', COUNT(*) FROM el_repo
UNION ALL SELECT 'el_qu', COUNT(*) FROM el_qu
UNION ALL SELECT 'el_qu_repo', COUNT(*) FROM el_qu_repo;

-- 提交
COMMIT;
