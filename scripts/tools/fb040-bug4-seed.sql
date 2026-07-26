-- FB-040 Bug4 验证：构造同 create_time 的 2 个测评 × 3 人，看 list 是否分组
SET NAMES utf8mb4;
DELETE FROM el_tester WHERE id_number LIKE 'FB040-%';

-- 共用同一时间戳 NOW() (秒级)
SET @t = NOW();

-- 测评 A (1776822816300709851 = 002 基层员工开放版) 3 人
INSERT INTO el_tester (id, exam_id, id_number, name, age, gender, password, telephone, stu_flag, create_time, update_time)
VALUES
  ('FB040ID000A1', '1776822816300709851', 'FB040-A1', 'A_001', 22, '0', '0001', '13800000001', 1, @t, @t),
  ('FB040ID000A2', '1776822816300709851', 'FB040-A2', 'A_002', 23, '0', '0002', '13800000002', 1, @t, @t),
  ('FB040ID000A3', '1776822816300709851', 'FB040-A3', 'A_003', 24, '0', '0003', '13800000003', 1, @t, @t);

-- 测评 B (1776820387219694716 = 003 MBTI 开放) 3 人 — id 字典序在 A 之后
INSERT INTO el_tester (id, exam_id, id_number, name, age, gender, password, telephone, stu_flag, create_time, update_time)
VALUES
  ('FB040ID000B1', '1776820387219694716', 'FB040-B1', 'B_001', 25, '0', '0004', '13800000004', 0, @t, @t),
  ('FB040ID000B2', '1776820387219694716', 'FB040-B2', 'B_002', 26, '0', '0005', '13800000005', 0, @t, @t),
  ('FB040ID000B3', '1776820387219694716', 'FB040-B3', 'B_003', 27, '0', '0006', '13800000006', 0, @t, @t);

SELECT id, exam_id, name, create_time FROM el_tester WHERE id_number LIKE 'FB040-%' ORDER BY id;
