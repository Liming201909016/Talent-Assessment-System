-- ============================================================
-- FB-002 修复：mbti.FillAnswer 并发 upsert 竞态
-- 对应：docs/regression-tests.md FB-002
-- ============================================================
--
-- 问题：FillAnswer 用 First → Create 两步 upsert，在并发场景下
--       两个请求都查不到记录 → 都尝试 Create → 产生重复行
--
-- 修复方案：在 (paper_id, qu_id) 上加 UNIQUE 索引
--           - 第二个并发 INSERT 会因唯一约束失败
--           - 应用层接到 Duplicate Entry 错误后退化为 UPDATE
--
-- 安全部署步骤：
--   1. 先在 staging 跑 SELECT 检查是否已有重复
--   2. 如有重复，先去重（保留最新 + 最大 score 的记录）
--   3. 再加索引
--
-- ============================================================

-- Step 1: 检查重复（先跑这条，确认无重复后再继续）
SELECT paper_id, qu_id, COUNT(*) AS cnt
FROM el_mbti_answer
GROUP BY paper_id, qu_id
HAVING cnt > 1;

-- Step 2: 如有重复，去重（保留 id 最大的，即最新的）
-- DELETE 必须有主键条件（MySQL safe update mode）
-- DELETE m1
-- FROM el_mbti_answer m1
-- INNER JOIN el_mbti_answer m2
--   ON m1.paper_id = m2.paper_id
--   AND m1.qu_id = m2.qu_id
--   AND m1.id < m2.id;

-- Step 3: 加唯一索引
ALTER TABLE el_mbti_answer
  ADD UNIQUE INDEX uk_mbti_answer_paper_qu (paper_id, qu_id);

-- Step 4: 验证索引存在
SHOW INDEX FROM el_mbti_answer WHERE Key_name = 'uk_mbti_answer_paper_qu';
