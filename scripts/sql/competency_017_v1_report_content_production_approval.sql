-- Production approval for the exact staging-tested v1 content candidate.
-- Authorized on 2026-10-09 to reuse the named staging approver and immutable SHA values.
SET NAMES utf8mb4;
BEGIN;

UPDATE `el_competency_report_content_package`
SET `approval_status`='approved',
    `content_approved_by`='Liming',
    `content_approved_at`=NOW(),
    `psychometric_approved_by`='Liming',
    `psychometric_approved_at`=NOW(),
    `effective_environment`='production',
    `update_time`=NOW()
WHERE `id`='phase1-candidate-draft-v1'
  AND `product_version`='competency-frontline-phase1-v1'
  AND `scoring_version`='competency-phase1-scoring-v1'
  AND `content_version`='competency-phase1-content-v1'
  AND `template_version`='competency-phase1-report-v1'
  AND `audience`='frontline_employee'
  AND `approval_status`='draft'
  AND `question_source_sha256`='f33b878e6fa3f3b8496a838c1a8e648dda29b1e75a41a21a73e90362a978b42f'
  AND `content_source_sha256`='3060bf06f3f52715c7cf9b05f277e4ccd723a7f571785c5a26918cc98d8dbb42';
SELECT ROW_COUNT() INTO @v1_package_approved;
SET @sql=IF(@v1_package_approved=1,
  'SELECT ''v1 production package approval valid''',
  'SELECT * FROM `__v1_production_package_approval_invalid__`');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

UPDATE `el_competency_report_text`
SET `status`=0,
    `is_temporary`=0,
    `disclaimer`=CONVERT(UNHEX('e7a791e5ada6e6b58be8af84e79a84e58e9fe58899e698afe5a49ae8b4a8e5a49ae6b395efbc8ce58db3e5afb9e6af8fe4b8aae7b4a0e8b4a8e79a84e8af84e4bbb7e98787e794a8e5a49ae7a78de6b58be8af95e6898be6aeb5efbc8ce69cace6b58be8af84e4bd9ce4b8bae4b880e7a78de6b58be8af95e6898be6aeb5efbc8ce899bde8af81e6988ee69c89e69588efbc8ce585b6e7bb93e69e9ce59fbae4ba8ee58f97e6b58be88085e887aae99988e58f8de5ba94e7bb93e69e9cefbc8ce58f97e585b6e6b58be8af84e78ab6e68081e38081e887aae68891e8aea4e79fa5e5818fe5b7aee7ad89e5bdb1e5938defbc8ce4b88de883bde4bd9ce4b8bae7b2bee58786e588a4e696ade4b8aae4bd93e5b297e4bd8de8839ce4bbbbe58a9be79a84e594afe4b880e4be9de68daee38082e69bb4e7b2bee58786e79a84e6b58be8af95e5bbbae8aeaee6a0b9e68daee5ae9ee99985e68385e586b5efbc8ce7bb93e59088e585b6e4bb96e6b58be8af95e7bb93e69e9cefbc8ce5a682e99da2e8af95e38081e7bba9e69588e8a1a8e78eb0e38081e68385e699afe6a8a1e68b9fe38081333630e8af84e4bcb0e7ad89e8bf9be8a18ce7bbbce59088e588a4e696ade38082') USING utf8mb4),
    `update_time`=NOW()
WHERE `content_version`='competency-phase1-content-v1'
  AND `audience`='frontline_employee'
  AND `is_temporary`=1
  AND `status`=1;
SELECT ROW_COUNT() INTO @v1_text_activated;
SET @sql=IF(@v1_text_activated=66,
  'SELECT ''v1 production text activation valid''',
  'SELECT * FROM `__v1_production_text_activation_invalid__`');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

COMMIT;