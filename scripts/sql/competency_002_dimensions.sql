-- ============================================================
-- 胜任力测验迁移 002：48 维度主数据与测评维度选择表
-- 兼容：MySQL 5.7+
-- 特性：可重复执行；不删除现有数据；不覆盖管理员维护的维度主数据
-- ============================================================

SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `el_competency_dimension` (
  `id` varchar(64) NOT NULL COMMENT '稳定ID',
  `code` varchar(16) NOT NULL COMMENT 'D01-D48',
  `name` varchar(100) NOT NULL COMMENT '维度名称',
  `vird_level` varchar(100) NOT NULL COMMENT 'VIRD层级',
  `applicable_category` varchar(50) NOT NULL COMMENT '基层通用/管理通用',
  `core_meaning` varchar(500) NOT NULL COMMENT '核心含义',
  `display_order` int NOT NULL COMMENT '默认展示顺序1-48',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0启用1停用',
  `create_time` datetime DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_competency_dimension_code` (`code`),
  UNIQUE KEY `uk_competency_dimension_name` (`name`),
  UNIQUE KEY `uk_competency_dimension_order` (`display_order`),
  KEY `idx_competency_dimension_status_order` (`status`,`display_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力维度主数据';

CREATE TABLE IF NOT EXISTS `el_exam_competency_dimension` (
  `id` varchar(64) NOT NULL COMMENT 'ID',
  `exam_id` varchar(64) NOT NULL COMMENT '测评ID',
  `dimension_id` varchar(64) NOT NULL COMMENT '维度ID',
  `dimension_code` varchar(16) NOT NULL DEFAULT '' COMMENT '发布快照：维度编号',
  `dimension_name` varchar(100) NOT NULL DEFAULT '' COMMENT '发布快照：维度名称',
  `vird_level` varchar(100) NOT NULL DEFAULT '' COMMENT '发布快照：VIRD层级',
  `applicable_category` varchar(50) NOT NULL DEFAULT '' COMMENT '发布快照：适用类别',
  `core_meaning` varchar(500) NOT NULL DEFAULT '' COMMENT '发布快照：核心含义',
  `display_order` int NOT NULL DEFAULT 0 COMMENT '发布快照：默认顺序',
  `question_count` int NOT NULL DEFAULT 0 COMMENT '发布时启用题数',
  `create_time` datetime DEFAULT NULL,
  `snapshot_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_exam_competency_dimension` (`exam_id`,`dimension_id`),
  KEY `idx_exam_competency_dimension_order` (`exam_id`,`display_order`),
  KEY `idx_exam_competency_dimension_dimension` (`dimension_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='测评选择的胜任力维度及发布快照';

-- 关联列必须继承现有 el_exam.id 的字符集和排序规则。
-- staging MySQL 8 使用 utf8mb4_0900_ai_ci，MySQL 5.7 环境使用 utf8mb4_general_ci；
-- 动态读取可同时兼容两类环境，并确保重复执行时修复已存在表。
SELECT CHARACTER_SET_NAME, COLLATION_NAME
INTO @exam_id_charset, @exam_id_collation
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'el_exam'
  AND COLUMN_NAME = 'id';
SET @sql = CONCAT(
  'ALTER TABLE `el_exam_competency_dimension` MODIFY COLUMN `exam_id` varchar(64) CHARACTER SET ',
  @exam_id_charset,
  ' COLLATE ',
  @exam_id_collation,
  ' NOT NULL COMMENT ''测评ID'''
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

INSERT INTO `el_competency_dimension`
(`id`,`code`,`name`,`vird_level`,`applicable_category`,`core_meaning`,`display_order`,`status`,`create_time`,`update_time`)
VALUES
('competency-d01','D01','沟通表达','Versatility 胜任力','基层通用','清晰传递与接收信息',1,0,NOW(),NOW()),
('competency-d02','D02','人际交往','Versatility 胜任力','基层通用','建立和维护积极的工作关系',2,0,NOW(),NOW()),
('competency-d03','D03','数字应用','Versatility 胜任力','基层通用','熟练运用数据工具处理工作',3,0,NOW(),NOW()),
('competency-d04','D04','计划执行','Versatility 胜任力','基层通用','高效推进计划并达成预期结果',4,0,NOW(),NOW()),
('competency-d05','D05','逻辑思维','Versatility 胜任力','基层通用','推理有条理、有依据',5,0,NOW(),NOW()),
('competency-d06','D06','持续学习','Versatility 胜任力','基层通用','主动更新知识储备与技能',6,0,NOW(),NOW()),
('competency-d07','D07','组织协调','Versatility 胜任力','基层通用','推动跨部门协作与流程顺畅衔接',7,0,NOW(),NOW()),
('competency-d08','D08','信息整合','Versatility 胜任力','基层通用','从碎片信息中提炼关键要点',8,0,NOW(),NOW()),
('competency-d09','D09','问题解决','Versatility 胜任力','基层通用','找到问题根源并拿出有效方案',9,0,NOW(),NOW()),
('competency-d10','D10','归纳总结','Versatility 胜任力','基层通用','从具体信息中提炼规律与结论',10,0,NOW(),NOW()),
('competency-d11','D11','战略思维','Versatility 胜任力','管理通用','从全局和长远视角把握方向',11,0,NOW(),NOW()),
('competency-d12','D12','团队领导','Versatility 胜任力','管理通用','凝聚团队力量实现共同目标',12,0,NOW(),NOW()),
('competency-d13','D13','商业洞察','Versatility 胜任力','管理通用','捕捉市场变化并预判机会与风险',13,0,NOW(),NOW()),
('competency-d14','D14','统筹规划','Versatility 胜任力','管理通用','合理配置资源、确定优先次序',14,0,NOW(),NOW()),
('competency-d15','D15','人才培养','Versatility 胜任力','管理通用','帮助下属成长、能够独当一面',15,0,NOW(),NOW()),
('competency-d16','D16','目标管理','Versatility 胜任力','管理通用','科学分解目标并持续追踪达成',16,0,NOW(),NOW()),
('competency-d17','D17','资源整合','Versatility 胜任力','管理通用','识别并最大化利用内外部资源',17,0,NOW(),NOW()),
('competency-d18','D18','风险防范','Versatility 胜任力','管理通用','提前预判风险并做好应对准备',18,0,NOW(),NOW()),
('competency-d19','D19','绩效管理','Versatility 胜任力','管理通用','通过考核与反馈推动团队进步',19,0,NOW(),NOW()),
('competency-d20','D20','成本管控','Versatility 胜任力','管理通用','在保证质量前提下优化资源使用',20,0,NOW(),NOW()),
('competency-d21','D21','敬业奉献','Integrity 信念力','管理通用','视工作为责任，全心投入',21,0,NOW(),NOW()),
('competency-d22','D22','客户至上','Integrity 信念力','管理通用','一切工作以客户价值为出发点',22,0,NOW(),NOW()),
('competency-d23','D23','规则意识','Integrity 信念力','管理通用','视制度与流程为不可突破的底线',23,0,NOW(),NOW()),
('competency-d24','D24','全局观念','Integrity 信念力','管理通用','超越本位，以组织整体利益为重',24,0,NOW(),NOW()),
('competency-d25','D25','利他精神','Integrity 信念力','管理通用','乐于助人，不计较个人得失',25,0,NOW(),NOW()),
('competency-d26','D26','公平公正','Integrity 信念力','管理通用','处理事务客观公道、不偏不倚',26,0,NOW(),NOW()),
('competency-d27','D27','共生共赢','Integrity 信念力','管理通用','追求各方共同受益、共同成长',27,0,NOW(),NOW()),
('competency-d28','D28','求真务实','Integrity 信念力','管理通用','尊重事实、注重实效、不搞形式',28,0,NOW(),NOW()),
('competency-d29','D29','担当作为','Integrity 信念力','管理通用','面对困难主动挺身、不推诿退缩',29,0,NOW(),NOW()),
('competency-d30','D30','组织忠诚','Integrity 信念力','管理通用','维护组织利益与声誉，始终如一',30,0,NOW(),NOW()),
('competency-d31','D31','自信心','Resilience 人格力','管理通用','确信自己的判断与能力',31,0,NOW(),NOW()),
('competency-d32','D32','严谨性','Resilience 人格力','管理通用','做事细致可靠、力求准确',32,0,NOW(),NOW()),
('competency-d33','D33','自律性','Resilience 人格力','管理通用','自我约束、不松懈、不逾矩',33,0,NOW(),NOW()),
('competency-d34','D34','条理性','Resilience 人格力','管理通用','工作安排井然有序、有章法',34,0,NOW(),NOW()),
('competency-d35','D35','合作性','Resilience 人格力','管理通用','乐于配合、不固执己见',35,0,NOW(),NOW()),
('competency-d36','D36','果断性','Resilience 人格力','管理通用','关键时刻不犹豫、快速拍板',36,0,NOW(),NOW()),
('competency-d37','D37','抗压性','Resilience 人格力','管理通用','承受压力时保持稳定发挥',37,0,NOW(),NOW()),
('competency-d38','D38','适应性','Resilience 人格力','管理通用','面对变化能快速自我调整',38,0,NOW(),NOW()),
('competency-d39','D39','主动性','Resilience 人格力','管理通用','不等安排即主动采取行动',39,0,NOW(),NOW()),
('competency-d40','D40','创新性','Resilience 人格力','管理通用','不墨守成规、乐于探索新思路',40,0,NOW(),NOW()),
('competency-d41','D41','成就动机','Drive 内驱力','管理通用','渴望挑战并达成高难度目标',41,0,NOW(),NOW()),
('competency-d42','D42','权力动机','Drive 内驱力','管理通用','渴望拥有影响力和主导权',42,0,NOW(),NOW()),
('competency-d43','D43','关系动机','Drive 内驱力','管理通用','渴望与人建立温暖的联结',43,0,NOW(),NOW()),
('competency-d44','D44','认知动机','Drive 内驱力','管理通用','渴望探索未知、追求真理',44,0,NOW(),NOW()),
('competency-d45','D45','超越动机','Drive 内驱力','管理通用','渴望不断突破自身极限',45,0,NOW(),NOW()),
('competency-d46','D46','自主动机','Drive 内驱力','管理通用','渴望按自己的节奏和方式行事',46,0,NOW(),NOW()),
('competency-d47','D47','胜任动机','Drive 内驱力','管理通用','渴望证明自己具备高水平能力',47,0,NOW(),NOW()),
('competency-d48','D48','社会动机','Drive 内驱力','管理通用','渴望为群体或社会创造价值',48,0,NOW(),NOW())
ON DUPLICATE KEY UPDATE
  `id`=`id`;

-- 只读验证：应为 48 条、序号 1-48、D42=权力动机。
SELECT COUNT(*) AS dimension_count,
       MIN(display_order) AS min_order,
       MAX(display_order) AS max_order,
       COUNT(DISTINCT code) AS unique_codes,
       COUNT(DISTINCT name) AS unique_names
FROM el_competency_dimension;

SELECT id, code, name, display_order, status
FROM el_competency_dimension
WHERE id = 'competency-d42';
