-- Copy-only synthetic fixture installer.
-- Authorized schema: talent_mng005_local_7081fbec31e3d105 only.
-- Clones 002 content into distinct 005 IDs; never updates/deletes 002 rows.
-- Ownership marker: MNG005_LOCAL_FIXTURE:7081fbec31e3d105:v1
SET NAMES utf8mb4;
SET @fixture_schema_ok=(SELECT DATABASE()='talent_mng005_local_7081fbec31e3d105');
SET @fixture_source_ok=(
  (SELECT COUNT(*) FROM el_repo WHERE code='00201')=1 AND
  (SELECT COUNT(*) FROM el_repo WHERE code='00202')=1 AND
  (SELECT COUNT(*) FROM el_qu_repo qr JOIN el_repo r ON r.id=qr.repo_id WHERE r.code='00201')=140 AND
  (SELECT COUNT(*) FROM el_qu_repo qr JOIN el_repo r ON r.id=qr.repo_id WHERE r.code='00202')=140 AND
  (SELECT COUNT(*) FROM el_qu_answer a JOIN el_qu_repo qr ON qr.qu_id=a.qu_id JOIN el_repo r ON r.id=qr.repo_id WHERE r.code='00201')=700 AND
  (SELECT COUNT(*) FROM el_qu_answer a JOIN el_qu_repo qr ON qr.qu_id=a.qu_id JOIN el_repo r ON r.id=qr.repo_id WHERE r.code='00202')=700
);
SET @fixture_target_empty=(
  (SELECT COUNT(*) FROM el_repo WHERE code IN ('00501','00502') OR id IN ('m00501-fixture-repo-v1','m00502-fixture-repo-v1'))=0 AND
  (SELECT COUNT(*) FROM el_qu WHERE id LIKE 'm00501-q-%' OR id LIKE 'm00502-q-%')=0 AND
  (SELECT COUNT(*) FROM el_qu_answer WHERE id LIKE 'm00501-a-%' OR id LIKE 'm00502-a-%')=0 AND
  (SELECT COUNT(*) FROM el_qu_repo WHERE id LIKE 'm00501-r-%' OR id LIKE 'm00502-r-%')=0
);
SET @fixture_id_lengths_ok=(
  (SELECT COALESCE(MAX(CHAR_LENGTH(q.id)),0) FROM el_qu q JOIN el_qu_repo qr ON qr.qu_id=q.id JOIN el_repo r ON r.id=qr.repo_id WHERE r.code IN ('00201','00202'))<=54 AND
  (SELECT COALESCE(MAX(CHAR_LENGTH(a.id)),0) FROM el_qu_answer a JOIN el_qu_repo qr ON qr.qu_id=a.qu_id JOIN el_repo r ON r.id=qr.repo_id WHERE r.code IN ('00201','00202'))<=54 AND
  (SELECT COALESCE(MAX(CHAR_LENGTH(qr.id)),0) FROM el_qu_repo qr JOIN el_repo r ON r.id=qr.repo_id WHERE r.code IN ('00201','00202'))<=54
);
SET @fixture_sql=IF(@fixture_schema_ok AND @fixture_source_ok AND @fixture_target_empty AND @fixture_id_lengths_ok,
  'SELECT ''MNG005_FIXTURE_PRECONDITION_OK''',
  'SELECT * FROM information_schema.MNG005_FIXTURE_PRECONDITION_FAILED');
PREPARE fixture_stmt FROM @fixture_sql; EXECUTE fixture_stmt; DEALLOCATE PREPARE fixture_stmt;

START TRANSACTION;
INSERT INTO el_repo (id,code,title,radio_count,multi_count,judge_count,remark,create_time,update_time)
SELECT 'm00501-fixture-repo-v1','00501','管理特质测验基层员工新版',radio_count,multi_count,judge_count,
       'MNG005_LOCAL_FIXTURE:7081fbec31e3d105:v1',NOW(),NOW()
FROM el_repo WHERE code='00201';
INSERT INTO el_repo (id,code,title,radio_count,multi_count,judge_count,remark,create_time,update_time)
SELECT 'm00502-fixture-repo-v1','00502','管理特质测验干部新版',radio_count,multi_count,judge_count,
       'MNG005_LOCAL_FIXTURE:7081fbec31e3d105:v1',NOW(),NOW()
FROM el_repo WHERE code='00202';

INSERT INTO el_qu (id,qu_type,level,image,content,create_time,update_time,remark,analysis,title,question_code,dimension_id,dimension_item_no,observation_point,scoring_direction,competency_question_type,question_status)
SELECT CONCAT('m00501-q-',q.id),q.qu_type,q.level,q.image,q.content,NOW(),NOW(),q.remark,q.analysis,q.title,q.question_code,q.dimension_id,q.dimension_item_no,q.observation_point,q.scoring_direction,q.competency_question_type,q.question_status
FROM el_qu q JOIN el_qu_repo qr ON qr.qu_id=q.id JOIN el_repo r ON r.id=qr.repo_id WHERE r.code='00201';
INSERT INTO el_qu (id,qu_type,level,image,content,create_time,update_time,remark,analysis,title,question_code,dimension_id,dimension_item_no,observation_point,scoring_direction,competency_question_type,question_status)
SELECT CONCAT('m00502-q-',q.id),q.qu_type,q.level,q.image,q.content,NOW(),NOW(),q.remark,q.analysis,
       CASE q.content
         WHEN 'V67' THEN '当我接手具有挑战性的工作时，我通常能鼓励大家创新，并提出创新的解决方案。'
         WHEN 'V96' THEN '当下属反对我的某个决定或者工作安排时，我会保持冷静和理性来应对。'
         ELSE q.title END,
       q.question_code,q.dimension_id,q.dimension_item_no,q.observation_point,q.scoring_direction,q.competency_question_type,q.question_status
FROM el_qu q JOIN el_qu_repo qr ON qr.qu_id=q.id JOIN el_repo r ON r.id=qr.repo_id WHERE r.code='00202';

INSERT INTO el_qu_answer (id,qu_id,is_right,image,content,analysis,score)
SELECT CONCAT('m00501-a-',a.id),CONCAT('m00501-q-',a.qu_id),a.is_right,a.image,a.content,a.analysis,a.score
FROM el_qu_answer a JOIN el_qu_repo qr ON qr.qu_id=a.qu_id JOIN el_repo r ON r.id=qr.repo_id WHERE r.code='00201';
INSERT INTO el_qu_answer (id,qu_id,is_right,image,content,analysis,score)
SELECT CONCAT('m00502-a-',a.id),CONCAT('m00502-q-',a.qu_id),a.is_right,a.image,a.content,a.analysis,a.score
FROM el_qu_answer a JOIN el_qu_repo qr ON qr.qu_id=a.qu_id JOIN el_repo r ON r.id=qr.repo_id WHERE r.code='00202';

INSERT INTO el_qu_repo (id,qu_id,repo_id,qu_type,sort)
SELECT CONCAT('m00501-r-',qr.id),CONCAT('m00501-q-',qr.qu_id),'m00501-fixture-repo-v1',qr.qu_type,qr.sort
FROM el_qu_repo qr JOIN el_repo r ON r.id=qr.repo_id WHERE r.code='00201';
INSERT INTO el_qu_repo (id,qu_id,repo_id,qu_type,sort)
SELECT CONCAT('m00502-r-',qr.id),CONCAT('m00502-q-',qr.qu_id),'m00502-fixture-repo-v1',qr.qu_type,qr.sort
FROM el_qu_repo qr JOIN el_repo r ON r.id=qr.repo_id WHERE r.code='00202';

SET @fixture_result_ok=(
  (SELECT COUNT(*) FROM el_repo WHERE code='00501' AND title='管理特质测验基层员工新版' AND id='m00501-fixture-repo-v1' AND remark='MNG005_LOCAL_FIXTURE:7081fbec31e3d105:v1')=1 AND
  (SELECT COUNT(*) FROM el_repo WHERE code='00502' AND title='管理特质测验干部新版' AND id='m00502-fixture-repo-v1' AND remark='MNG005_LOCAL_FIXTURE:7081fbec31e3d105:v1')=1 AND
  (SELECT COUNT(*) FROM el_qu_repo WHERE repo_id='m00501-fixture-repo-v1')=140 AND
  (SELECT COUNT(*) FROM el_qu_repo WHERE repo_id='m00502-fixture-repo-v1')=140 AND
  (SELECT COUNT(*) FROM el_qu_answer WHERE qu_id LIKE 'm00501-q-%')=700 AND
  (SELECT COUNT(*) FROM el_qu_answer WHERE qu_id LIKE 'm00502-q-%')=700 AND
  (SELECT COUNT(*) FROM el_qu q JOIN el_qu_repo qr ON qr.qu_id=q.id WHERE qr.repo_id='m00502-fixture-repo-v1' AND q.content='V67' AND q.title='当我接手具有挑战性的工作时，我通常能鼓励大家创新，并提出创新的解决方案。')=1 AND
  (SELECT COUNT(*) FROM el_qu q JOIN el_qu_repo qr ON qr.qu_id=q.id WHERE qr.repo_id='m00502-fixture-repo-v1' AND q.content='V96' AND q.title='当下属反对我的某个决定或者工作安排时，我会保持冷静和理性来应对。')=1
);
SET @fixture_sql=IF(@fixture_result_ok,
  'SELECT ''MNG005_FIXTURE_SIGNATURE_OK''',
  'SELECT * FROM information_schema.MNG005_FIXTURE_SIGNATURE_FAILED');
PREPARE fixture_stmt FROM @fixture_sql; EXECUTE fixture_stmt; DEALLOCATE PREPARE fixture_stmt;
COMMIT;
