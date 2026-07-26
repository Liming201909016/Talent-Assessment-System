-- MBTI 职业性格测验答题记录表
-- 一题一行，score_a + score_b = 5
-- 与现有 el_paper_qu_answer 表独立，避免 checked/is_right 语义冲突

CREATE TABLE IF NOT EXISTS el_mbti_answer (
    id          VARCHAR(64) NOT NULL PRIMARY KEY,
    paper_id    VARCHAR(64) NOT NULL,
    qu_id       VARCHAR(64) NOT NULL,
    score_a     INT NOT NULL DEFAULT 0 COMMENT 'A选项赋分 0-5',
    score_b     INT NOT NULL DEFAULT 0 COMMENT 'B选项赋分 0-5',
    sort        INT NOT NULL DEFAULT 0 COMMENT '题目排序',
    answered    TINYINT NOT NULL DEFAULT 0 COMMENT '是否已答 0/1',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_mbti_paper (paper_id),
    INDEX idx_mbti_paper_qu (paper_id, qu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
