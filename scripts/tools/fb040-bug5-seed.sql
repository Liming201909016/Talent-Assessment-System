-- 重置 Liming paper 用于端到端 submit 测试
UPDATE el_paper SET state=0, user_time=0 WHERE id='1777217353411783509';
UPDATE el_candidate SET end_time=NULL, pdf_flag=0 WHERE paper_id='1777217353411783509';
DELETE FROM el_mbti_answer WHERE paper_id='1777217353411783509';
-- 灌入与该 exam 关联的 qu_id 的 48 题答案
INSERT INTO el_mbti_answer (id, paper_id, qu_id, score_a, score_b, answered, create_time)
SELECT
  CONCAT('FB040E2E', LPAD(@row := @row + 1, 4, '0')) AS id,
  '1777217353411783509',
  qr.qu_id,
  3, 2, 1, NOW()
FROM (SELECT @row := 0) r,
     (SELECT qr.qu_id FROM el_qu_repo qr
      JOIN el_exam_repo er ON er.repo_id = qr.repo_id
      JOIN el_paper p ON p.exam_id = er.exam_id
      WHERE p.id = '1777217353411783509'
      ORDER BY qr.sort LIMIT 48) qr;
SELECT 'paper' AS tbl, state, user_time FROM el_paper WHERE id='1777217353411783509';
SELECT 'candidate' AS tbl, end_time, pdf_flag FROM el_candidate WHERE paper_id='1777217353411783509';
SELECT 'answer_count' AS tbl, COUNT(*) FROM el_mbti_answer WHERE paper_id='1777217353411783509';
