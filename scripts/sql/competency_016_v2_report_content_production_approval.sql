-- Production approval for the exact staging-tested v2 content candidate.
-- Authorized on 2026-10-09 to reuse the named staging approvers and immutable SHA values.
SET NAMES utf8mb4;
BEGIN;

UPDATE `el_competency_report_content_package`
SET `approval_status`='approved',
    `content_approved_by`='LIming',
    `content_approved_at`=NOW(),
    `psychometric_approved_by`='Ruiling',
    `psychometric_approved_at`=NOW(),
    `effective_environment`='production',
    `update_time`=NOW()
WHERE `id`='phase1-v2-content-draft'
  AND `product_version`='competency-frontline-phase1-v2'
  AND `scoring_version`='competency-phase1-scoring-v2'
  AND `content_version`='competency-phase1-content-v2'
  AND `template_version`='competency-phase1-report-v2'
  AND `audience`='frontline_employee'
  AND `approval_status`='draft'
  AND `question_source_sha256`='edb9efd27ec86bc34db3a796c2022a99495fd9ec52e8a6580cd7404b2ab933b5'
  AND `content_source_sha256`='329409e408f10ec7f048a757e83c963b7736f66e41e8d601ff4397fd8545b48c';
SELECT ROW_COUNT() INTO @v2_package_approved;
SET @sql=IF(@v2_package_approved=1,
  'SELECT ''v2 production package approval valid''',
  'SELECT * FROM `__v2_production_package_approval_invalid__`');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

UPDATE `el_competency_report_text`
SET `status`=0,`update_time`=NOW()
WHERE `content_version`='competency-phase1-content-v2'
  AND `audience`='frontline_employee'
  AND `is_temporary`=0
  AND `status`=1;
SELECT ROW_COUNT() INTO @v2_text_activated;
SET @sql=IF(@v2_text_activated=124,
  'SELECT ''v2 production text activation valid''',
  'SELECT * FROM `__v2_production_text_activation_invalid__`');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

COMMIT;
