-- 2026-04-20: 新增 required_fields 列
-- 用于存储考试配置中勾选的考生必填信息项（逗号分隔）
-- 例：name,gender,age,telephone
-- 前端 form.vue 通过 requiredFieldsList checkbox 生成此字段
ALTER TABLE el_exam ADD COLUMN required_fields VARCHAR(500) DEFAULT NULL AFTER pdf_path;
