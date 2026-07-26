#!/bin/bash
# Deep business data analysis (READ-ONLY on prod data)
DB="sudo mysql element -N"

echo "╔══════════════════════════════════════════════════════╗"
echo "║  生产数据深度分析 — 业务逻辑理解                      ║"
echo "╚══════════════════════════════════════════════════════╝"

echo ""
echo "━━━ 1. 考试完整信息（含题库关联） ━━━"
$DB -e "
SELECT e.id, e.title, e.is_open,
  CASE e.is_open WHEN 1 THEN '开放' WHEN 2 THEN '封闭' END AS open_mode,
  e.open_type, e.join_type, e.state, e.time_limit, e.total_time, e.total_score,
  r.code AS repo_code, r.title AS repo_title,
  (SELECT COUNT(*) FROM el_tester t WHERE t.exam_id=e.id) AS tester_cnt,
  (SELECT COUNT(*) FROM el_candidate c WHERE c.exam_id=e.id) AS candidate_cnt,
  (SELECT COUNT(*) FROM el_paper p WHERE p.exam_id=e.id) AS paper_cnt
FROM el_exam e
LEFT JOIN el_exam_repo er ON er.exam_id = e.id
LEFT JOIN el_repo r ON r.id = er.repo_id
ORDER BY e.create_time DESC
"

echo ""
echo "━━━ 2. 题库统计 ━━━"
$DB -e "
SELECT r.id, r.code, r.title, r.radio_count, r.multi_count, r.judge_count,
  (SELECT COUNT(*) FROM el_qu_repo qr WHERE qr.repo_id=r.id) AS actual_qu_count,
  (SELECT COUNT(*) FROM el_exam_repo er WHERE er.repo_id=r.id) AS used_in_exams
FROM el_repo r ORDER BY r.id
"

echo ""
echo "━━━ 3. 测评人员分布 ━━━"
echo "--- 按考试分布 ---"
$DB -e "
SELECT t.exam_id, e.title AS exam_title,
  COUNT(*) AS tester_count,
  SUM(CASE WHEN t.paper_id IS NOT NULL THEN 1 ELSE 0 END) AS has_paper,
  SUM(CASE WHEN t.end_time IS NOT NULL THEN 1 ELSE 0 END) AS completed,
  SUM(CASE WHEN t.pdf_flag=1 THEN 1 ELSE 0 END) AS has_pdf
FROM el_tester t
LEFT JOIN el_exam e ON e.id = t.exam_id
WHERE t.del_flag IS NULL OR t.del_flag = '0'
GROUP BY t.exam_id, e.title
ORDER BY tester_count DESC
"

echo ""
echo "━━━ 4. 候选人（开放测评）分布 ━━━"
$DB -e "
SELECT c.exam_id, e.title,
  COUNT(*) AS total,
  SUM(CASE WHEN c.paper_id IS NOT NULL THEN 1 ELSE 0 END) AS has_paper,
  SUM(CASE WHEN c.end_time IS NOT NULL THEN 1 ELSE 0 END) AS completed,
  SUM(CASE WHEN c.pdf_flag=1 THEN 1 ELSE 0 END) AS has_pdf
FROM el_candidate c
LEFT JOIN el_exam e ON e.id = c.exam_id
WHERE c.del_flag = '0'
GROUP BY c.exam_id, e.title
ORDER BY total DESC LIMIT 10
"

echo ""
echo "━━━ 5. 试卷状态分析 ━━━"
$DB -e "
SELECT
  CASE p.state WHEN 0 THEN '进行中' WHEN 1 THEN '待批改' WHEN 2 THEN '已完成' WHEN 3 THEN '中断' ELSE CONCAT('未知:',p.state) END AS state_name,
  COUNT(*) AS cnt,
  ROUND(AVG(p.user_score),1) AS avg_score,
  ROUND(AVG(p.user_time),1) AS avg_time_min
FROM el_paper p
GROUP BY p.state
ORDER BY p.state
"

echo ""
echo "━━━ 6. 答题数据分析 ━━━"
$DB -e "
SELECT
  COUNT(DISTINCT pq.paper_id) AS papers_with_answers,
  COUNT(*) AS total_paper_qu,
  SUM(pq.answered) AS answered,
  ROUND(SUM(pq.answered)*100.0/COUNT(*),1) AS answer_rate,
  SUM(pq.is_right) AS correct,
  ROUND(SUM(pq.is_right)*100.0/NULLIF(SUM(pq.answered),0),1) AS correct_rate
FROM el_paper_qu pq
"

echo ""
echo "━━━ 7. 用户角色映射 ━━━"
$DB -e "
SELECT u.user_id, u.user_name, u.real_name,
  GROUP_CONCAT(r.role_name) AS roles,
  d.dept_name
FROM sys_user u
LEFT JOIN sys_user_role ur ON ur.user_id = u.user_id
LEFT JOIN sys_role r ON r.role_id = ur.role_id
LEFT JOIN sys_dept d ON d.dept_id = u.dept_id
GROUP BY u.user_id, u.user_name, u.real_name, d.dept_name
ORDER BY u.user_id
"

echo ""
echo "━━━ 8. 考试时间线 ━━━"
$DB -e "
SELECT e.title,
  DATE(e.create_time) AS created,
  e.start_time, e.end_time,
  CASE e.state WHEN 0 THEN '启用' WHEN 1 THEN '禁用' ELSE CONCAT('其他:',e.state) END AS state,
  e.is_open, e.total_time
FROM el_exam e
ORDER BY e.create_time DESC LIMIT 15
"

echo ""
echo "━━━ 9. 数据外键完整性 ━━━"
echo "孤儿 paper (exam_id 不存在):"
$DB -e "SELECT COUNT(*) FROM el_paper p LEFT JOIN el_exam e ON p.exam_id=e.id WHERE e.id IS NULL"
echo "孤儿 paper_qu (paper_id 不存在):"
$DB -e "SELECT COUNT(*) FROM el_paper_qu pq LEFT JOIN el_paper p ON pq.paper_id=p.id WHERE p.id IS NULL"
echo "孤儿 tester (exam_id 不存在):"
$DB -e "SELECT COUNT(*) FROM el_tester t LEFT JOIN el_exam e ON t.exam_id=e.id WHERE e.id IS NULL AND t.exam_id IS NOT NULL"
echo "孤儿 candidate (exam_id 不存在):"
$DB -e "SELECT COUNT(*) FROM el_candidate c LEFT JOIN el_exam e ON c.exam_id=e.id WHERE e.id IS NULL"

echo ""
echo "ANALYSIS_DONE"
