#!/usr/bin/env bash
# Only the exact MTRa05eb3af5e49 receipt. Never DDL, FK disabling, or legacy PDFs.
set -euo pipefail
umask 077
[ "$(hostname)" = vm-ubuntu-go-dev ] && [ "$(id -u)" = 0 ]
root=/opt/talent-assessment
evidence=$root/backups/mng_phase1_20261003_112853_fd24d4ee35a9/four_20261003_145100_b5bc510880c1
private=$root/private/management-traits-test-reports
[ "$(cat "$evidence/ownership")" = mng-four-current-baseline-v1 ]
[ "$(stat -c %a "$evidence")" = 700 ]
mysql --batch --raw --skip-column-names element -e "SELECT id,file_key,file_sha FROM el_mng_report_revision WHERE exam_id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205') ORDER BY id" > "$evidence/owned-pdf-files.tsv"
[ "$(wc -l < "$evidence/owned-pdf-files.tsv")" = 4 ]
while IFS=$'\t' read -r id key hash; do
 [[ "$id" =~ ^[a-f0-9-]{36}$ ]] && [ "$key" = "$id.pdf" ] && [[ "$hash" =~ ^[a-f0-9]{64}$ ]]
 [ ! -L "$private/$key" ] && [ "$(sha256sum "$private/$key" | cut -d' ' -f1)" = "$hash" ]
done < "$evidence/owned-pdf-files.tsv"
mysql --batch --raw --skip-column-names element <<'SQL'
SET NAMES utf8mb4; SET SESSION SQL_SAFE_UPDATES=1; SET SESSION group_concat_max_len=10485760;
SET SESSION range_optimizer_max_mem_size=67108864;
START TRANSACTION;
SELECT id FROM el_exam WHERE id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205') ORDER BY id FOR UPDATE;
SELECT IF((SELECT COUNT(*) FROM el_exam WHERE
 (id='1791039089847823693' AND title='MTRa05eb3af5e49-00201-candidate') OR
 (id='1791039090118371541' AND title='MTRa05eb3af5e49-00201-tester') OR
 (id='1791039117022179538' AND title='MTRa05eb3af5e49-00202-candidate') OR
 (id='1791039117279750205' AND title='MTRa05eb3af5e49-00202-tester'))=4,1,JSON_EXTRACT('OWNERSHIP_EXAM_STOP','$'));
SELECT GROUP_CONCAT(QUOTE(id)) INTO @exam_ids FROM el_exam WHERE id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205');
SELECT id FROM el_mng_definition_bundle WHERE id IN ('5c72dd3d-7e35-4e54-88be-8d92ca3aaa78','253d2ef0-ff8d-4b14-b79d-933a00da8a6c') ORDER BY id FOR UPDATE;
SELECT IF((SELECT COUNT(*) FROM el_mng_exam_profile WHERE bundle_id IN ('5c72dd3d-7e35-4e54-88be-8d92ca3aaa78','253d2ef0-ff8d-4b14-b79d-933a00da8a6c') AND exam_id NOT IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205'))=0 AND
 (SELECT COUNT(*) FROM el_mng_paper_snapshot WHERE bundle_id IN ('5c72dd3d-7e35-4e54-88be-8d92ca3aaa78','253d2ef0-ff8d-4b14-b79d-933a00da8a6c') AND exam_id NOT IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205'))=0,1,JSON_EXTRACT('SHARED_BUNDLE_REFERENCE_STOP','$'));
SELECT IF((SELECT COUNT(*) FROM el_paper WHERE exam_id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205') AND id NOT IN ('57af1302-b5a6-413c-ac0b-54a1e586c873','0281004e-5638-428b-8d3b-b3712071ce3e','ab3ed6b8-1f6a-493d-8723-0c6d89a57467','8480da2c-52b3-4c98-a77c-4a81015af2f6','babedb5c-ebb5-495b-994f-9ceedb8952b6','0846eda9-42c2-451c-a61d-dea3ed9253d6'))=0,1,JSON_EXTRACT('UNKNOWN_PAPER_STOP','$'));
SELECT IF((SELECT COUNT(*) FROM el_candidate WHERE exam_id IN ('1791039089847823693','1791039117022179538') AND LEFT(name,15)<>'MTRa05eb3af5e49')=0 AND
 (SELECT COUNT(*) FROM el_tester WHERE exam_id IN ('1791039090118371541','1791039117279750205') AND id NOT IN ('1791039116754649070','1791039117369562530'))=0 AND
 (SELECT COUNT(*) FROM el_user_exam WHERE exam_id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205'))=0 AND
 (SELECT COUNT(*) FROM el_exam_depart WHERE exam_id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205'))=0,1,JSON_EXTRACT('UNKNOWN_DEPENDENCY_STOP','$'));
SELECT GROUP_CONCAT(QUOTE(id)) INTO @papers FROM el_paper WHERE id IN ('57af1302-b5a6-413c-ac0b-54a1e586c873','0281004e-5638-428b-8d3b-b3712071ce3e','ab3ed6b8-1f6a-493d-8723-0c6d89a57467','8480da2c-52b3-4c98-a77c-4a81015af2f6','babedb5c-ebb5-495b-994f-9ceedb8952b6','0846eda9-42c2-451c-a61d-dea3ed9253d6');
SELECT GROUP_CONCAT(QUOTE(id)) INTO @runs FROM el_mng_result_run WHERE exam_id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205');
SELECT GROUP_CONCAT(QUOTE(id)) INTO @reports FROM el_mng_report_revision WHERE exam_id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205');
SELECT GROUP_CONCAT(QUOTE(a.id)) INTO @audits FROM el_mng_report_audit a JOIN el_mng_report_revision r ON r.id=a.report_id WHERE r.exam_id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205');
SET @s=IF(@audits IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_report_audit WHERE id IN (',@audits,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=CONCAT('DELETE FROM el_mng_report_current WHERE paper_id IN (',@papers,')'); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=IF(@reports IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_report_revision WHERE id IN (',@reports,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(d.id)) INTO @ids FROM el_mng_result_dimension d JOIN el_mng_result_run r ON r.id=d.run_id WHERE r.exam_id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205');
SET @s=IF(@ids IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_result_dimension WHERE id IN (',@ids,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(m.id)) INTO @ids FROM el_mng_result_module m JOIN el_mng_result_run r ON r.id=m.run_id WHERE r.exam_id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205');
SET @s=IF(@ids IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_result_module WHERE id IN (',@ids,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=IF(@runs IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_runtime_receipt WHERE run_id IN (',@runs,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=IF(@runs IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_result_run WHERE id IN (',@runs,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(id)) INTO @ids FROM el_mng_paper_question_snapshot WHERE paper_id IN ('57af1302-b5a6-413c-ac0b-54a1e586c873','0281004e-5638-428b-8d3b-b3712071ce3e','ab3ed6b8-1f6a-493d-8723-0c6d89a57467','8480da2c-52b3-4c98-a77c-4a81015af2f6','babedb5c-ebb5-495b-994f-9ceedb8952b6','0846eda9-42c2-451c-a61d-dea3ed9253d6');
SET @s=CONCAT('DELETE FROM el_mng_paper_question_snapshot WHERE id IN (',@ids,')'); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=CONCAT('DELETE FROM el_mng_paper_snapshot WHERE paper_id IN (',@papers,')'); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(id)) INTO @ids FROM el_paper_qu_answer WHERE paper_id IN ('57af1302-b5a6-413c-ac0b-54a1e586c873','0281004e-5638-428b-8d3b-b3712071ce3e','ab3ed6b8-1f6a-493d-8723-0c6d89a57467','8480da2c-52b3-4c98-a77c-4a81015af2f6','babedb5c-ebb5-495b-994f-9ceedb8952b6','0846eda9-42c2-451c-a61d-dea3ed9253d6');
SET @s=CONCAT('DELETE FROM el_paper_qu_answer WHERE id IN (',@ids,')'); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(id)) INTO @ids FROM el_paper_qu WHERE paper_id IN ('57af1302-b5a6-413c-ac0b-54a1e586c873','0281004e-5638-428b-8d3b-b3712071ce3e','ab3ed6b8-1f6a-493d-8723-0c6d89a57467','8480da2c-52b3-4c98-a77c-4a81015af2f6','babedb5c-ebb5-495b-994f-9ceedb8952b6','0846eda9-42c2-451c-a61d-dea3ed9253d6');
SET @s=CONCAT('DELETE FROM el_paper_qu WHERE id IN (',@ids,')'); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(id)) INTO @ids FROM el_candidate WHERE exam_id IN ('1791039089847823693','1791039117022179538');
SET @s=CONCAT('DELETE FROM el_candidate WHERE id IN (',@ids,')'); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
DELETE FROM el_tester WHERE id IN ('1791039116754649070','1791039117369562530');
SET @s=CONCAT('DELETE FROM el_paper WHERE id IN (',@papers,')'); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=CONCAT('DELETE FROM el_mng_exam_profile WHERE exam_id IN (',@exam_ids,')'); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
DELETE FROM el_mng_definition_bundle WHERE id IN ('5c72dd3d-7e35-4e54-88be-8d92ca3aaa78','253d2ef0-ff8d-4b14-b79d-933a00da8a6c');
SELECT GROUP_CONCAT(QUOTE(id)) INTO @ids FROM el_exam_repo WHERE exam_id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205');
SET @s=CONCAT('DELETE FROM el_exam_repo WHERE id IN (',@ids,')'); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=CONCAT('DELETE FROM el_exam WHERE id IN (',@exam_ids,')'); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
COMMIT;
SELECT 'EXACT_OWNED_CLEANUP_COMMITTED';
SELECT 'OWNED_LEGACY_RESIDUAL',(SELECT COUNT(*) FROM el_exam WHERE id IN ('1791039089847823693','1791039090118371541','1791039117022179538','1791039117279750205'))+(SELECT COUNT(*) FROM el_paper WHERE id IN ('57af1302-b5a6-413c-ac0b-54a1e586c873','0281004e-5638-428b-8d3b-b3712071ce3e','ab3ed6b8-1f6a-493d-8723-0c6d89a57467','8480da2c-52b3-4c98-a77c-4a81015af2f6','babedb5c-ebb5-495b-994f-9ceedb8952b6','0846eda9-42c2-451c-a61d-dea3ed9253d6'));
SQL
while IFS=$'\t' read -r id key hash; do
 [ "$(sha256sum "$private/$key" | cut -d' ' -f1)" = "$hash" ]; rm -- "$private/$key"
done < "$evidence/owned-pdf-files.tsv"
for t in $(mysql --batch --skip-column-names element -e "SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name"); do
 [[ "$t" =~ ^el_mng_[a-z_]+$ ]]; n=$(mysql --batch --skip-column-names element -e "SELECT COUNT(*) FROM $t"); [ "$n" = 0 ]; printf 'FINAL_SIDECAR=%s ROWS=%s\n' "$t" "$n"
done
mysqldump --single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary element el_exam el_exam_repo el_repo el_qu el_qu_repo el_qu_answer el_paper el_paper_qu el_paper_qu_answer el_candidate el_tester el_mbti_answer | sha256sum | cut -d' ' -f1 > "$evidence/legacy-after.sha256"
find "$root/tmp" /data/uploadPath -type f -iname '*.pdf' -print0 | sort -z | xargs -0 -r sha256sum > "$evidence/pdf-after.sha256"
find "$root/dist" "$root/configs" /etc/systemd/system/talent-assessment.service.d -type f -print0 | sort -z | xargs -0 sha256sum > "$evidence/config-after.sha256"
cmp "$evidence/legacy-before.sha256" "$evidence/legacy-after.sha256"
cmp "$evidence/pdf-before.sha256" "$evidence/pdf-after.sha256"
cmp "$evidence/config-before.sha256" "$evidence/config-after.sha256"
[ "$(find "$private" -type f | wc -l)" = 0 ]
[ "$(sha256sum "$root/server" | cut -d' ' -f1)" = abea56b32f5159211b44f065facfc740ee961b313c4a51b377f915284fc3a2cf ]
for s in talent-assessment nginx mysql; do systemctl is-active "$s"; done
curl --fail --silent --max-time 10 http://127.0.0.1:8092/health
printf '\nCURRENT_BASELINE_UNCHANGED=1 OLD_PDFS=465 PRIVATE_FILES=0 OWNED_RESIDUAL=0 BACKEND_UNCHANGED=1\n'
date -u +%FT%TZ