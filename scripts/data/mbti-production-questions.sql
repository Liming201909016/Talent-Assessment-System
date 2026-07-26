-- MBTI 题库（生产版）
-- 生成时间: 2026-04-21
-- 修正: 2026-04-23 修复编码/事务/幂等/统计/清理

SET NAMES utf8mb4;
BEGIN;

-- 0. 清理旧数据（保证每次可重复执行）
DELETE FROM el_qu_repo WHERE id BETWEEN '2026042100040000001' AND '2026042100040000048';
DELETE FROM el_qu_answer WHERE id BETWEEN '2026042100030000001' AND '2026042100030000096';
DELETE FROM el_qu WHERE id BETWEEN '2026042100020000001' AND '2026042100020000048';
DELETE FROM el_repo WHERE id = '2026042100010000001';

-- 1. 创建题库
INSERT INTO el_repo (id, code, title, create_time, update_time) VALUES ('2026042100010000001', '00301', '职业性格测验(生产)', '2026-04-21 12:00:00', '2026-04-21 12:00:00');

-- 2. 创建48道题目 + 96个选项 + 题库关联
-- Q1 (E-I)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000001', 1, 1, '', 'V1', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'E-I', '', '当你遇到新朋友时，你');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000001', '2026042100020000001', 0, '', '说话的时间与聆听的时间差不多', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000002', '2026042100020000001', 0, '', '聆听的时间会比说话的时间多', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000001', '2026042100020000001', '2026042100010000001', 1, 0);

-- Q2 (S-N)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000002', 1, 1, '', 'V2', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'S-N', '', '下列哪一种做法是你的一般生活取向？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000003', '2026042100020000002', 0, '', '只管做吧', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000004', '2026042100020000002', 0, '', '先找出多种不同选择再去做', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000002', '2026042100020000002', '2026042100010000001', 1, 1);

-- Q3 (T-F)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000003', 1, 1, '', 'V3', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'T-F', '', '你喜欢自己的哪种性格？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000005', '2026042100020000003', 0, '', '冷静而理性', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000006', '2026042100020000003', 0, '', '热情而体谅', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000003', '2026042100020000003', '2026042100010000001', 1, 2);

-- Q4 (J-P)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000004', 1, 1, '', 'V4', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'J-P', '', '你擅长');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000007', '2026042100020000004', 0, '', '同时协调进行多项工作', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000008', '2026042100020000004', 0, '', '专注在某一项工作上，直至把它完成为止', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000004', '2026042100020000004', '2026042100010000001', 1, 3);

-- Q5 (E-I)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000005', 1, 1, '', 'V5', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'E-I', '', '你参与社交聚会时');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000009', '2026042100020000005', 0, '', '总是能认识新朋友', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000010', '2026042100020000005', 0, '', '只跟几个亲密挚友待在一起', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000005', '2026042100020000005', '2026042100010000001', 1, 4);

-- Q6 (S-N)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000006', 1, 1, '', 'V6', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'S-N', '', '当你尝试了解某些事情时，一般你会');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000011', '2026042100020000006', 0, '', '先了解细节', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000012', '2026042100020000006', 0, '', '先了解整体情况，细节容后再谈', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000006', '2026042100020000006', '2026042100010000001', 1, 5);

-- Q7 (T-F)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000007', 1, 1, '', 'V7', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'T-F', '', '你对下列哪方面较感兴趣？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000013', '2026042100020000007', 0, '', '知道别人的想法', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000014', '2026042100020000007', 0, '', '了解别人的感受', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000007', '2026042100020000007', '2026042100010000001', 1, 6);

-- Q8 (J-P)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000008', 1, 1, '', 'V8', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'J-P', '', '你比较喜欢下列哪个工作？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000015', '2026042100020000008', 0, '', '让你能迅速和及时做出反应的', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000016', '2026042100020000008', 0, '', '让你能定出目标，然后逐步达成目标的', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000008', '2026042100020000008', '2026042100010000001', 1, 7);

-- Q9 (E-I)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000009', 1, 1, '', 'V9', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'E-I', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000017', '2026042100020000009', 0, '', '当我与朋友狂欢后，我会感到精力充沛，并会继续追求这种欢娱', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000018', '2026042100020000009', 0, '', '当我与朋友狂欢后，我会感到疲累，觉得需要一些空间', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000009', '2026042100020000009', '2026042100010000001', 1, 8);

-- Q10 (S-N)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000010', 1, 1, '', 'V10', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'S-N', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000019', '2026042100020000010', 0, '', '我较有兴趣知道别人的经历，例如他们做过什么？认识什么人？', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000020', '2026042100020000010', 0, '', '我较有兴趣知道别人的计划和梦想，例如他们会往哪里去？憧憬什么？', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000010', '2026042100020000010', '2026042100010000001', 1, 9);

-- Q11 (T-F)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000011', 1, 1, '', 'V11', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'T-F', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000021', '2026042100020000011', 0, '', '我擅长制定一些可行的计划', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000022', '2026042100020000011', 0, '', '我擅长促成别人同意一些计划，并衷心合作', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000011', '2026042100020000011', '2026042100010000001', 1, 10);

-- Q12 (J-P)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000012', 1, 1, '', 'V12', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'J-P', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000023', '2026042100020000012', 0, '', '我会突然尝试做某些事，看看会有什么事情发生', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000024', '2026042100020000012', 0, '', '我尝试做任何事情前，都想事先知道可能有什么事情发生', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000012', '2026042100020000012', '2026042100010000001', 1, 11);

-- Q13 (E-I)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000013', 1, 1, '', 'V13', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'E-I', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000025', '2026042100020000013', 0, '', '我经常边说话，边思考', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000026', '2026042100020000013', 0, '', '我在说话前，通常会先想好要说的话', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000013', '2026042100020000013', '2026042100010000001', 1, 12);

-- Q14 (S-N)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000014', 1, 1, '', 'V14', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'S-N', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000027', '2026042100020000014', 0, '', '周围的实际环境对我很重要，而且会影响我的感受', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000028', '2026042100020000014', 0, '', '如果我喜欢所做的事情，气氛对我而言并不是那么重要', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000014', '2026042100020000014', '2026042100010000001', 1, 13);

-- Q15 (T-F)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000015', 1, 1, '', 'V15', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'T-F', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000029', '2026042100020000015', 0, '', '我喜欢分析，心思缜密', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000030', '2026042100020000015', 0, '', '我对人感兴趣，关心他们所发生的事', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000015', '2026042100020000015', '2026042100010000001', 1, 14);

-- Q16 (J-P)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000016', 1, 1, '', 'V16', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'J-P', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000031', '2026042100020000016', 0, '', '即使已经定出计划，我也喜欢探讨其他新的方案', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000032', '2026042100020000016', 0, '', '一旦定出计划，我便希望能依计行事', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000016', '2026042100020000016', '2026042100010000001', 1, 15);

-- Q17 (E-I)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000017', 1, 1, '', 'V17', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'E-I', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000033', '2026042100020000017', 0, '', '认识我的人，一般都知道什么对我来说是重要的', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000034', '2026042100020000017', 0, '', '除了我感觉亲近的人，我不会对人说出什么对我来说是重要的', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000017', '2026042100020000017', '2026042100010000001', 1, 16);

-- Q18 (S-N)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000018', 1, 1, '', 'V18', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'S-N', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000035', '2026042100020000018', 0, '', '如果我喜欢某种活动，我会经常进行这种活动', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000036', '2026042100020000018', 0, '', '我一旦熟悉某种活动后，便希望转而尝试其他新的活动', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000018', '2026042100020000018', '2026042100010000001', 1, 17);

-- Q19 (T-F)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000019', 1, 1, '', 'V19', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'T-F', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000037', '2026042100020000019', 0, '', '当我作决定的时候，我更多地考虑正反两面的观点，并且做推理与质证', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000038', '2026042100020000019', 0, '', '当我作决定的时候，我会更多地了解其他人的想法，并希望能够达成共识', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000019', '2026042100020000019', '2026042100010000001', 1, 18);

-- Q20 (J-P)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000020', 1, 1, '', 'V20', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'J-P', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000039', '2026042100020000020', 0, '', '当我专注做某件事情时，需要不时停下来休息', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000040', '2026042100020000020', 0, '', '当我专注做某件事情时，不希望受到任何干扰', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000020', '2026042100020000020', '2026042100010000001', 1, 19);

-- Q21 (E-I)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000021', 1, 1, '', 'V21', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'E-I', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000041', '2026042100020000021', 0, '', '我独处太久，便会感到不安', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000042', '2026042100020000021', 0, '', '若没有足够的自处时间，我便会感到烦躁不安', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000021', '2026042100020000021', '2026042100010000001', 1, 20);

-- Q22 (S-N)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000022', 1, 1, '', 'V22', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'S-N', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000043', '2026042100020000022', 0, '', '我对一些没有实际用途的意念不感兴趣', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000044', '2026042100020000022', 0, '', '我喜欢意念本身，并享受想象意念的过程', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000022', '2026042100020000022', '2026042100010000001', 1, 21);

-- Q23 (T-F)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000023', 1, 1, '', 'V23', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'T-F', '', '下列哪一种说法较适合你？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000045', '2026042100020000023', 0, '', '当进行谈判时，我依靠自己的知识和技巧', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000046', '2026042100020000023', 0, '', '当进行谈判时，我会团结其他人至同一阵线', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000023', '2026042100020000023', '2026042100010000001', 1, 22);

-- Q24 (J-P)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000024', 1, 1, '', 'V24', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'J-P', '', '当你放假时，你多数会');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000047', '2026042100020000024', 0, '', '随遇而安，做当时想做的事', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000048', '2026042100020000024', 0, '', '为想做的事情订出时间表', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000024', '2026042100020000024', '2026042100010000001', 1, 23);

-- Q25 (E-I)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000025', 1, 1, '', 'V25', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'E-I', '', '当你放假时，你多数会');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000049', '2026042100020000025', 0, '', '多花些时间与别人共度', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000050', '2026042100020000025', 0, '', '多花些时间自己阅读、散步或者做白日梦', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000025', '2026042100020000025', '2026042100010000001', 1, 24);

-- Q26 (S-N)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000026', 1, 1, '', 'V26', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'S-N', '', '当你放假时，你多数会');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000051', '2026042100020000026', 0, '', '返回你喜欢的地方度假', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000052', '2026042100020000026', 0, '', '选择前往一些你从未到过的地方', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000026', '2026042100020000026', '2026042100010000001', 1, 25);

-- Q27 (T-F)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000027', 1, 1, '', 'V27', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'T-F', '', '当你放假时，你多数会');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000053', '2026042100020000027', 0, '', '做一些与工作有关的事情', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000054', '2026042100020000027', 0, '', '处理一些对你重要的人际关系', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000027', '2026042100020000027', '2026042100010000001', 1, 26);

-- Q28 (J-P)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000028', 1, 1, '', 'V28', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'J-P', '', '当你放假时，你多数会');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000055', '2026042100020000028', 0, '', '忘记平时发生的事情，专心享乐', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000056', '2026042100020000028', 0, '', '想着假期过后要准备的事情', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000028', '2026042100020000028', '2026042100010000001', 1, 27);

-- Q29 (E-I)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000029', 1, 1, '', 'V29', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'E-I', '', '当你放假时，你多数会');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000057', '2026042100020000029', 0, '', '参观著名景点', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000058', '2026042100020000029', 0, '', '花时间逛博物馆和一些较为幽静的地方', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000029', '2026042100020000029', '2026042100010000001', 1, 28);

-- Q30 (S-N)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000030', 1, 1, '', 'V30', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'S-N', '', '当你放假时，你多数会');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000059', '2026042100020000030', 0, '', '在喜欢的餐厅用餐', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000060', '2026042100020000030', 0, '', '尝试新的菜式', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000030', '2026042100020000030', '2026042100010000001', 1, 29);

-- Q31 (T-F)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000031', 1, 1, '', 'V31', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'T-F', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000061', '2026042100020000031', 0, '', '别人认为我会公正处事，并且尊重他人', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000062', '2026042100020000031', 0, '', '别人相信在他们有需要时，我会在他们身边', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000031', '2026042100020000031', '2026042100010000001', 1, 30);

-- Q32 (J-P)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000032', 1, 1, '', 'V32', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'J-P', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000063', '2026042100020000032', 0, '', '随机应变', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000064', '2026042100020000032', 0, '', '按照计划行事', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000032', '2026042100020000032', '2026042100010000001', 1, 31);

-- Q33 (E-I)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000033', 1, 1, '', 'V33', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'E-I', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000065', '2026042100020000033', 0, '', '坦率', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000066', '2026042100020000033', 0, '', '深沉', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000033', '2026042100020000033', '2026042100010000001', 1, 32);

-- Q34 (S-N)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000034', 1, 1, '', 'V34', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'S-N', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000067', '2026042100020000034', 0, '', '留意事实', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000068', '2026042100020000034', 0, '', '注重事实', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000034', '2026042100020000034', '2026042100010000001', 1, 33);

-- Q35 (T-F)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000035', 1, 1, '', 'V35', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'T-F', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000069', '2026042100020000035', 0, '', '知识广博', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000070', '2026042100020000035', 0, '', '善解人意', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000035', '2026042100020000035', '2026042100010000001', 1, 34);

-- Q36 (J-P)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000036', 1, 1, '', 'V36', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'J-P', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000071', '2026042100020000036', 0, '', '容易适应转变', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000072', '2026042100020000036', 0, '', '处事井井有条', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000036', '2026042100020000036', '2026042100010000001', 1, 35);

-- Q37 (E-I)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000037', 1, 1, '', 'V37', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'E-I', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000073', '2026042100020000037', 0, '', '爽朗', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000074', '2026042100020000037', 0, '', '沉稳', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000037', '2026042100020000037', '2026042100010000001', 1, 36);

-- Q38 (S-N)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000038', 1, 1, '', 'V38', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'S-N', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000075', '2026042100020000038', 0, '', '实事求是', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000076', '2026042100020000038', 0, '', '富有想象力', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000038', '2026042100020000038', '2026042100010000001', 1, 37);

-- Q39 (T-F)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000039', 1, 1, '', 'V39', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'T-F', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000077', '2026042100020000039', 0, '', '喜欢询问实情', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000078', '2026042100020000039', 0, '', '喜欢探索感受', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000039', '2026042100020000039', '2026042100010000001', 1, 38);

-- Q40 (J-P)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000040', 1, 1, '', 'V40', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'J-P', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000079', '2026042100020000040', 0, '', '不断接受新意见', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000080', '2026042100020000040', 0, '', '着眼达成目标', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000040', '2026042100020000040', '2026042100010000001', 1, 39);

-- Q41 (E-I)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000041', 1, 1, '', 'V41', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'E-I', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000081', '2026042100020000041', 0, '', '率直', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000082', '2026042100020000041', 0, '', '内敛', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000041', '2026042100020000041', '2026042100010000001', 1, 40);

-- Q42 (S-N)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000042', 1, 1, '', 'V42', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'S-N', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000083', '2026042100020000042', 0, '', '实事求是', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000084', '2026042100020000042', 0, '', '具备远大目光', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000042', '2026042100020000042', '2026042100010000001', 1, 41);

-- Q43 (T-F)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000043', 1, 1, '', 'V43', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'T-F', '', '下列哪个说法最能贴切形容你对自己的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000085', '2026042100020000043', 0, '', '公正', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000086', '2026042100020000043', 0, '', '宽容', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000043', '2026042100020000043', '2026042100010000001', 1, 42);

-- Q44 (J-P)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000044', 1, 1, '', 'V44', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'J-P', '', '你会倾向于');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000087', '2026042100020000044', 0, '', '暂时放下不愉快的事情，直至想处理时才处理', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000088', '2026042100020000044', 0, '', '及时处理不愉快的事情，力求把它们抛诸脑后', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000044', '2026042100020000044', '2026042100010000001', 1, 43);

-- Q45 (E-I)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000045', 1, 1, '', 'V45', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'E-I', '', '你会倾向于');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000089', '2026042100020000045', 0, '', '自己的工作被欣赏，即使你自己并不满意', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000090', '2026042100020000045', 0, '', '创造一些有长远价值的东西，但不一定需要别人知道是你做的', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000045', '2026042100020000045', '2026042100010000001', 1, 44);

-- Q46 (S-N)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000046', 1, 1, '', 'V46', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'S-N', '', '你会倾向于');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000091', '2026042100020000046', 0, '', '在自己有兴趣的范畴，积累丰富的经验', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000092', '2026042100020000046', 0, '', '有各式各样不同的经验', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000046', '2026042100020000046', '2026042100010000001', 1, 45);

-- Q47 (T-F)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000047', 1, 1, '', 'V47', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'T-F', '', '哪一句较能表达你的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000093', '2026042100020000047', 0, '', '感情用事的人较容易犯错', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000094', '2026042100020000047', 0, '', '逻辑思维会令人自以为是，因而容易犯错', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000047', '2026042100020000047', '2026042100010000001', 1, 46);

-- Q48 (J-P)
INSERT INTO el_qu (id, qu_type, level, image, content, create_time, update_time, analysis, remark, title) VALUES ('2026042100020000048', 1, 1, '', 'V48', '2026-04-21 12:00:00', '2026-04-21 12:00:00', 'J-P', '', '哪一句较能表达你的看法？');
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000095', '2026042100020000048', 0, '', '犹豫不决必失败', '', 0);
INSERT INTO el_qu_answer (id, qu_id, is_right, image, content, analysis, score) VALUES ('2026042100030000096', '2026042100020000048', 0, '', '三思而后行', '', 0);
INSERT INTO el_qu_repo (id, qu_id, repo_id, qu_type, sort) VALUES ('2026042100040000048', '2026042100020000048', '2026042100010000001', 1, 47);

-- 3. 刷新题库统计
UPDATE el_repo SET radio_count = (SELECT COUNT(*) FROM el_qu_repo WHERE repo_id = '2026042100010000001' AND qu_type = 1) WHERE id = '2026042100010000001';

COMMIT;
