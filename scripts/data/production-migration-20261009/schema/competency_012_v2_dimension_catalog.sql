-- ============================================================
-- 胜任力测验迁移 012：v2产品版本维度目录及v1到v2身份映射
-- 兼容 MySQL 5.7+；幂等；不修改旧维度、题目或发布快照
-- ============================================================
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `el_competency_version_dimension` (
  `id` varchar(64) NOT NULL,
  `product_version` varchar(32) NOT NULL,
  `stable_key` varchar(64) NOT NULL,
  `display_code` varchar(16) NOT NULL,
  `name` varchar(100) NOT NULL,
  `module_code` varchar(32) NOT NULL,
  `display_order` int NOT NULL,
  `create_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_version_dimension_key` (`product_version`,`stable_key`),
  UNIQUE KEY `uk_version_dimension_code` (`product_version`,`display_code`),
  UNIQUE KEY `uk_version_dimension_order` (`product_version`,`display_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力产品版本维度目录';

CREATE TABLE IF NOT EXISTS `el_competency_dimension_mapping` (
  `id` varchar(64) NOT NULL,
  `source_product_version` varchar(32) NOT NULL,
  `source_dimension_id` varchar(64) NOT NULL,
  `target_product_version` varchar(32) NOT NULL,
  `target_dimension_id` varchar(64) NOT NULL,
  `create_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dimension_mapping_source` (`source_product_version`,`source_dimension_id`,`target_product_version`),
  UNIQUE KEY `uk_dimension_mapping_target` (`source_product_version`,`target_product_version`,`target_dimension_id`),
  CONSTRAINT `fk_dimension_mapping_target` FOREIGN KEY (`target_dimension_id`) REFERENCES `el_competency_version_dimension` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='胜任力跨产品版本维度身份映射';

INSERT INTO `el_competency_version_dimension`
  (`id`,`product_version`,`stable_key`,`display_code`,`name`,`module_code`,`display_order`,`create_time`)
VALUES
  ('competency-logical-reasoning','competency-frontline-phase1-v2','logical_reasoning','A1-01','逻辑思维','task_management',1,NOW()),
  ('competency-plan-execution','competency-frontline-phase1-v2','plan_execution','A1-02','计划执行','task_management',2,NOW()),
  ('competency-digital-application','competency-frontline-phase1-v2','digital_application','A1-03','数字应用','task_management',3,NOW()),
  ('competency-achievement-orientation','competency-frontline-phase1-v2','achievement_orientation','A1-04','成就导向','task_management',4,NOW()),
  ('competency-continuous-learning','competency-frontline-phase1-v2','continuous_learning','A1-05','持续学习','task_management',5,NOW()),
  ('competency-communication','competency-frontline-phase1-v2','communication','B1-01','沟通表达','interpersonal_management',6,NOW()),
  ('competency-cooperation','competency-frontline-phase1-v2','cooperation','B1-02','合作意识','interpersonal_management',7,NOW()),
  ('competency-truth-pragmatism','competency-frontline-phase1-v2','truth_pragmatism','C1-01','求真务实','self_management',8,NOW()),
  ('competency-self-discipline','competency-frontline-phase1-v2','self_discipline','C1-02','自律性','self_management',9,NOW()),
  ('competency-dedication','competency-frontline-phase1-v2','dedication','C1-03','敬业奉献','self_management',10,NOW())
ON DUPLICATE KEY UPDATE `id`=`id`;

-- No-op reruns must not silently preserve drifted immutable catalog rows.
SELECT COUNT(*) INTO @v2_dimension_signature_count
FROM `el_competency_version_dimension`
WHERE (`id`,`product_version`,`stable_key`,`display_code`,`name`,`module_code`,`display_order`) IN (
  ('competency-logical-reasoning','competency-frontline-phase1-v2','logical_reasoning','A1-01','逻辑思维','task_management',1),
  ('competency-plan-execution','competency-frontline-phase1-v2','plan_execution','A1-02','计划执行','task_management',2),
  ('competency-digital-application','competency-frontline-phase1-v2','digital_application','A1-03','数字应用','task_management',3),
  ('competency-achievement-orientation','competency-frontline-phase1-v2','achievement_orientation','A1-04','成就导向','task_management',4),
  ('competency-continuous-learning','competency-frontline-phase1-v2','continuous_learning','A1-05','持续学习','task_management',5),
  ('competency-communication','competency-frontline-phase1-v2','communication','B1-01','沟通表达','interpersonal_management',6),
  ('competency-cooperation','competency-frontline-phase1-v2','cooperation','B1-02','合作意识','interpersonal_management',7),
  ('competency-truth-pragmatism','competency-frontline-phase1-v2','truth_pragmatism','C1-01','求真务实','self_management',8),
  ('competency-self-discipline','competency-frontline-phase1-v2','self_discipline','C1-02','自律性','self_management',9),
  ('competency-dedication','competency-frontline-phase1-v2','dedication','C1-03','敬业奉献','self_management',10)
);
SET @sql=IF(@v2_dimension_signature_count=10,
  'SELECT ''v2 dimension catalog signature valid''',
  'SELECT * FROM `__v2_dimension_catalog_drift__`');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @v2_dimension_scope_count
FROM `el_competency_version_dimension`
WHERE `product_version`='competency-frontline-phase1-v2';
SET @sql=IF(@v2_dimension_scope_count=10,
  'SELECT ''v2 dimension catalog scope valid''',
  'SELECT * FROM `__v2_dimension_catalog_extra_rows__`');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

INSERT INTO `el_competency_dimension_mapping`
  (`id`,`source_product_version`,`source_dimension_id`,`target_product_version`,`target_dimension_id`,`create_time`)
VALUES
  ('v1-to-v2-logical-reasoning','competency-frontline-phase1-v1','competency-a1-01','competency-frontline-phase1-v2','competency-logical-reasoning',NOW()),
  ('v1-to-v2-plan-execution','competency-frontline-phase1-v1','competency-a1-03','competency-frontline-phase1-v2','competency-plan-execution',NOW()),
  ('v1-to-v2-digital-application','competency-frontline-phase1-v1','competency-a1-02','competency-frontline-phase1-v2','competency-digital-application',NOW()),
  ('v1-to-v2-achievement-orientation','competency-frontline-phase1-v1','competency-b1-04','competency-frontline-phase1-v2','competency-achievement-orientation',NOW()),
  ('v1-to-v2-continuous-learning','competency-frontline-phase1-v1','competency-a1-04','competency-frontline-phase1-v2','competency-continuous-learning',NOW()),
  ('v1-to-v2-communication','competency-frontline-phase1-v1','competency-a1-05','competency-frontline-phase1-v2','competency-communication',NOW()),
  ('v1-to-v2-cooperation','competency-frontline-phase1-v1','competency-b1-05','competency-frontline-phase1-v2','competency-cooperation',NOW()),
  ('v1-to-v2-truth-pragmatism','competency-frontline-phase1-v1','competency-b1-02','competency-frontline-phase1-v2','competency-truth-pragmatism',NOW()),
  ('v1-to-v2-self-discipline','competency-frontline-phase1-v1','competency-b1-03','competency-frontline-phase1-v2','competency-self-discipline',NOW()),
  ('v1-to-v2-dedication','competency-frontline-phase1-v1','competency-b1-01','competency-frontline-phase1-v2','competency-dedication',NOW())
ON DUPLICATE KEY UPDATE `id`=`id`;

SELECT COUNT(*) INTO @v2_mapping_signature_count
FROM `el_competency_dimension_mapping`
WHERE (`id`,`source_product_version`,`source_dimension_id`,`target_product_version`,`target_dimension_id`) IN (
  ('v1-to-v2-logical-reasoning','competency-frontline-phase1-v1','competency-a1-01','competency-frontline-phase1-v2','competency-logical-reasoning'),
  ('v1-to-v2-plan-execution','competency-frontline-phase1-v1','competency-a1-03','competency-frontline-phase1-v2','competency-plan-execution'),
  ('v1-to-v2-digital-application','competency-frontline-phase1-v1','competency-a1-02','competency-frontline-phase1-v2','competency-digital-application'),
  ('v1-to-v2-achievement-orientation','competency-frontline-phase1-v1','competency-b1-04','competency-frontline-phase1-v2','competency-achievement-orientation'),
  ('v1-to-v2-continuous-learning','competency-frontline-phase1-v1','competency-a1-04','competency-frontline-phase1-v2','competency-continuous-learning'),
  ('v1-to-v2-communication','competency-frontline-phase1-v1','competency-a1-05','competency-frontline-phase1-v2','competency-communication'),
  ('v1-to-v2-cooperation','competency-frontline-phase1-v1','competency-b1-05','competency-frontline-phase1-v2','competency-cooperation'),
  ('v1-to-v2-truth-pragmatism','competency-frontline-phase1-v1','competency-b1-02','competency-frontline-phase1-v2','competency-truth-pragmatism'),
  ('v1-to-v2-self-discipline','competency-frontline-phase1-v1','competency-b1-03','competency-frontline-phase1-v2','competency-self-discipline'),
  ('v1-to-v2-dedication','competency-frontline-phase1-v1','competency-b1-01','competency-frontline-phase1-v2','competency-dedication')
);
SET @sql=IF(@v2_mapping_signature_count=10,
  'SELECT ''v1-to-v2 dimension mapping signature valid''',
  'SELECT * FROM `__v1_to_v2_dimension_mapping_drift__`');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @v2_mapping_scope_count
FROM `el_competency_dimension_mapping`
WHERE `source_product_version`='competency-frontline-phase1-v1'
  AND `target_product_version`='competency-frontline-phase1-v2';
SET @sql=IF(@v2_mapping_scope_count=10,
  'SELECT ''v1-to-v2 dimension mapping scope valid''',
  'SELECT * FROM `__v1_to_v2_dimension_mapping_extra_rows__`');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
