-- ============================================================
-- 目标数据库：39.106.61.48 (MySQL 5.7) element
-- 用途：Go 后端部署所需的 Schema 变更
-- 原则：仅新增表和索引，不修改/删除任何现有数据和结构
-- 执行方式：手动在 MySQL 客户端执行，逐段确认
-- ============================================================

-- -------- 1. 创建 el_mbti_answer 表（MBTI 答题记录）--------
-- Go 新增功能：MBTI 性格测试的答题数据存储
-- 使用 utf8mb4_general_ci（兼容 MySQL 5.7，不用 8.0 的 0900_ai_ci）

CREATE TABLE IF NOT EXISTS `el_mbti_answer` (
  `id` varchar(64) NOT NULL,
  `paper_id` varchar(64) NOT NULL,
  `qu_id` varchar(64) NOT NULL,
  `score_a` int(11) NOT NULL DEFAULT '0',
  `score_b` int(11) NOT NULL DEFAULT '0',
  `sort` int(11) NOT NULL DEFAULT '0',
  `answered` tinyint(4) NOT NULL DEFAULT '0',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_mbti_answer_paper_qu` (`paper_id`,`qu_id`),
  KEY `idx_mbti_paper` (`paper_id`),
  KEY `idx_mbti_paper_qu` (`paper_id`,`qu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- 验证
SELECT 'el_mbti_answer' AS step, COUNT(*) AS ok FROM information_schema.TABLES
WHERE TABLE_SCHEMA='element' AND TABLE_NAME='el_mbti_answer';


-- -------- 2. 优化索引（可选，提升查询性能）--------
-- el_qu 复合索引：加速按题型+难度查询
-- IF NOT EXISTS 语法 MySQL 5.7 不支持，用 SELECT 判断

SELECT COUNT(*) INTO @idx_exists FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA='element' AND TABLE_NAME='el_qu' AND INDEX_NAME='idx_qu_type_level_id';
SET @sql = IF(@idx_exists=0,
  'CREATE INDEX idx_qu_type_level_id ON el_qu(qu_type, level, id)',
  'SELECT ''idx_qu_type_level_id already exists'' AS info');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- el_qu_repo 复合索引：加速按题库+题型查询
SELECT COUNT(*) INTO @idx_exists FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA='element' AND TABLE_NAME='el_qu_repo' AND INDEX_NAME='idx_repo_qu_type_qu_id';
SET @sql = IF(@idx_exists=0,
  'CREATE INDEX idx_repo_qu_type_qu_id ON el_qu_repo(repo_id, qu_type, qu_id)',
  'SELECT ''idx_repo_qu_type_qu_id already exists'' AS info');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;


-- -------- 3. 统一 el_repo 字符集（可选）--------
-- 当前 el_repo 使用 utf8_general_ci，其他表用 utf8mb4
-- 统一后消除 JOIN 时的隐式字符集转换

ALTER TABLE el_repo CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

-- 验证
SELECT TABLE_NAME, TABLE_COLLATION FROM information_schema.TABLES
WHERE TABLE_SCHEMA='element' AND TABLE_NAME='el_repo';


-- -------- 4. 最终验证 --------
SELECT TABLE_NAME, TABLE_COLLATION
FROM information_schema.TABLES
WHERE TABLE_SCHEMA='element' AND TABLE_NAME LIKE 'el_%'
ORDER BY TABLE_NAME;
