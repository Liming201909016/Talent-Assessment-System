-- ============================================================
-- 胜任力测验迁移 005：历史身份快照、查询索引与关系约束
-- 兼容 MySQL 5.7+；幂等；仅补充字段/索引/外键和既有结果回填
-- ============================================================
SET NAMES utf8mb4;

-- Result participant snapshots (polymorphic candidate/tester; no FK by design).
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='participant_type';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_competency_result` ADD COLUMN `participant_type` varchar(16) NOT NULL DEFAULT '''' AFTER `evaluation_level`','SELECT ''participant_type exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='participant_id';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_competency_result` ADD COLUMN `participant_id` varchar(64) NOT NULL DEFAULT '''' AFTER `participant_type`','SELECT ''participant_id exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='participant_name';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_competency_result` ADD COLUMN `participant_name` varchar(100) NOT NULL DEFAULT '''' AFTER `participant_id`','SELECT ''participant_name exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='participant_telephone';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_competency_result` ADD COLUMN `participant_telephone` varchar(32) NOT NULL DEFAULT '''' AFTER `participant_name`','SELECT ''participant_telephone exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='participant_age';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_competency_result` ADD COLUMN `participant_age` int DEFAULT NULL AFTER `participant_telephone`','SELECT ''participant_age exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='participant_gender';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_competency_result` ADD COLUMN `participant_gender` varchar(16) NOT NULL DEFAULT '''' AFTER `participant_age`','SELECT ''participant_gender exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='participant_affiliation';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_competency_result` ADD COLUMN `participant_affiliation` varchar(255) NOT NULL DEFAULT '''' AFTER `participant_gender`','SELECT ''participant_affiliation exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='participant_post';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_competency_result` ADD COLUMN `participant_post` varchar(100) NOT NULL DEFAULT '''' AFTER `participant_affiliation`','SELECT ''participant_post exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='participant_degree';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_competency_result` ADD COLUMN `participant_degree` varchar(100) NOT NULL DEFAULT '''' AFTER `participant_post`','SELECT ''participant_degree exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @column_exists FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME='participant_major';
SET @sql=IF(@column_exists=0,'ALTER TABLE `el_competency_result` ADD COLUMN `participant_major` varchar(100) NOT NULL DEFAULT '''' AFTER `participant_degree`','SELECT ''participant_major exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT CHARACTER_SET_NAME,COLLATION_NAME INTO @participant_cs,@participant_co FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_candidate' AND COLUMN_NAME='id';
SET @sql=CONCAT('ALTER TABLE `el_competency_result` MODIFY `participant_id` varchar(64) CHARACTER SET ',@participant_cs,' COLLATE ',@participant_co,' NOT NULL DEFAULT '''''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

UPDATE el_competency_result r
LEFT JOIN el_candidate c ON c.paper_id=r.paper_id AND c.exam_id=r.exam_id
LEFT JOIN el_tester t ON t.paper_id=r.paper_id AND t.exam_id=r.exam_id
SET r.participant_type=CASE WHEN c.id IS NOT NULL THEN 'candidate' WHEN t.id IS NOT NULL THEN 'tester' ELSE r.participant_type END,
    r.participant_id=COALESCE(c.id,t.id,r.participant_id),
    r.participant_name=COALESCE(c.name,t.name,r.participant_name),
    r.participant_telephone=COALESCE(c.telephone,t.telephone,r.participant_telephone),
    r.participant_age=COALESCE(c.age,t.age,r.participant_age),
    r.participant_gender=COALESCE(c.gender,t.gender,r.participant_gender),
    r.participant_affiliation=COALESCE(c.affiliation,t.affiliation,r.participant_affiliation),
    r.participant_post=COALESCE(c.post,t.post,r.participant_post),
    r.participant_degree=COALESCE(c.degree,t.degree,r.participant_degree),
    r.participant_major=COALESCE(c.major,t.major,r.participant_major)
WHERE r.participant_id='';

SELECT COUNT(*) INTO @index_exists FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_exam_competency_question' AND INDEX_NAME='uk_exam_competency_question_order';
SET @sql=IF(@index_exists=0,'CREATE UNIQUE INDEX `uk_exam_competency_question_order` ON `el_exam_competency_question` (`exam_id`,`snapshot_order`)','SELECT ''uk_exam_competency_question_order exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @index_exists FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND INDEX_NAME='idx_competency_result_exam_submitted';
SET @sql=IF(@index_exists=0,'CREATE INDEX `idx_competency_result_exam_submitted` ON `el_competency_result` (`exam_id`,`submitted_at`,`paper_id`)','SELECT ''idx_competency_result_exam_submitted exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Add only non-polymorphic foreign keys. Application full-chain deletion already follows child->parent order.
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_ecdim_exam';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_exam_competency_dimension` ADD CONSTRAINT `fk_ecdim_exam` FOREIGN KEY (`exam_id`) REFERENCES `el_exam` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_ecdim_exam exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_ecdim_dimension';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_exam_competency_dimension` ADD CONSTRAINT `fk_ecdim_dimension` FOREIGN KEY (`dimension_id`) REFERENCES `el_competency_dimension` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_ecdim_dimension exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_ecq_exam';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_exam_competency_question` ADD CONSTRAINT `fk_ecq_exam` FOREIGN KEY (`exam_id`) REFERENCES `el_exam` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_ecq_exam exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_ecq_exam_dimension';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_exam_competency_question` ADD CONSTRAINT `fk_ecq_exam_dimension` FOREIGN KEY (`exam_dimension_id`) REFERENCES `el_exam_competency_dimension` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_ecq_exam_dimension exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_ecq_source_qu';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_exam_competency_question` ADD CONSTRAINT `fk_ecq_source_qu` FOREIGN KEY (`source_qu_id`) REFERENCES `el_qu` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_ecq_source_qu exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_pq_exam_question';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_paper_qu` ADD CONSTRAINT `fk_pq_exam_question` FOREIGN KEY (`exam_question_id`) REFERENCES `el_exam_competency_question` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_pq_exam_question exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_cdr_paper';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_competency_dimension_result` ADD CONSTRAINT `fk_cdr_paper` FOREIGN KEY (`paper_id`) REFERENCES `el_paper` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_cdr_paper exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_cdr_exam_dimension';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_competency_dimension_result` ADD CONSTRAINT `fk_cdr_exam_dimension` FOREIGN KEY (`exam_dimension_id`) REFERENCES `el_exam_competency_dimension` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_cdr_exam_dimension exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_cdr_dimension';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_competency_dimension_result` ADD CONSTRAINT `fk_cdr_dimension` FOREIGN KEY (`dimension_id`) REFERENCES `el_competency_dimension` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_cdr_dimension exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_cr_paper';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_competency_result` ADD CONSTRAINT `fk_cr_paper` FOREIGN KEY (`paper_id`) REFERENCES `el_paper` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_cr_paper exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SELECT COUNT(*) INTO @fk_exists FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME='fk_cr_exam';
SET @sql=IF(@fk_exists=0,'ALTER TABLE `el_competency_result` ADD CONSTRAINT `fk_cr_exam` FOREIGN KEY (`exam_id`) REFERENCES `el_exam` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT','SELECT ''fk_cr_exam exists'''); PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SELECT COUNT(*) AS participant_snapshot_columns FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='el_competency_result' AND COLUMN_NAME IN ('participant_type','participant_id','participant_name','participant_telephone','participant_age','participant_gender','participant_affiliation','participant_post','participant_degree','participant_major');
SELECT COUNT(*) AS hardening_indexes FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND INDEX_NAME IN ('uk_exam_competency_question_order','idx_competency_result_exam_submitted');
SELECT COUNT(*) AS competency_foreign_keys FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME IN ('fk_ecdim_exam','fk_ecdim_dimension','fk_ecq_exam','fk_ecq_exam_dimension','fk_ecq_source_qu','fk_pq_exam_question','fk_cdr_paper','fk_cdr_exam_dimension','fk_cdr_dimension','fk_cr_paper','fk_cr_exam');
