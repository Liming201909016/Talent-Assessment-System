-- 重置一份 MBTI paper 用于测试（清空所有答案）
DELETE FROM el_mbti_answer WHERE paper_id=1777205263459584611;
SELECT COUNT(*) remaining FROM el_mbti_answer WHERE paper_id=1777205263459584611;
