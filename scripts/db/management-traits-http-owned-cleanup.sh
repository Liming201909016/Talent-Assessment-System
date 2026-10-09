#!/usr/bin/env bash
# Exact four-exam HTTP receipt only; no source writes, DDL or shared settings.
set -euo pipefail
umask 077
mark=${1:?mark}; shift
[[ "$mark" =~ ^MTH[a-f0-9]{12}$ ]] && [ "$#" = 4 ]
[ "$(hostname)" = vm-ubuntu-go-dev ] && [ "$(id -u)" = 0 ]
ids=''
for id in "$@"; do [[ "$id" =~ ^[0-9]{19}$ ]]; ids="${ids:+$ids,}'$id'"; done
root=/opt/talent-assessment
evidence=$root/backups/mng_phase1_20261003_112853_fd24d4ee35a9/http_${mark#MTH}
private=$root/private/management-traits-test-reports
[ "$(cat "$evidence/ownership")" = "$mark" ] && [ "$(stat -c %a "$evidence")" = 700 ]
mysql --batch --raw --skip-column-names element -e "SELECT id,file_key,file_sha FROM el_mng_report_revision WHERE exam_id IN ($ids) ORDER BY id" > "$evidence/owned-pdf-files.tsv"
while IFS=$'\t' read -r id key hash; do
 [[ "$id" =~ ^[a-f0-9-]{36}$ ]] && [ "$key" = "$id.pdf" ] && [[ "$hash" =~ ^[a-f0-9]{64}$ ]]
 [ ! -L "$private/$key" ] && [ "$(sha256sum "$private/$key" | cut -d' ' -f1)" = "$hash" ]
done < "$evidence/owned-pdf-files.tsv"
mysql --batch --raw --skip-column-names element <<SQL
SET NAMES utf8mb4;
SET SESSION SQL_SAFE_UPDATES=1;
SET SESSION group_concat_max_len=10485760;
SET SESSION range_optimizer_max_mem_size=67108864;
START TRANSACTION;
SELECT id FROM el_exam WHERE id IN ($ids) ORDER BY id FOR UPDATE;
SELECT IF((SELECT COUNT(*) FROM el_exam WHERE id IN ($ids) AND title IN ('$mark-00201-candidate','$mark-00201-tester','$mark-00202-candidate','$mark-00202-tester'))=4,1,JSON_EXTRACT('HTTP_EXAM_OWNERSHIP_STOP','\$'));
SELECT IF((SELECT COUNT(*) FROM el_candidate WHERE exam_id IN ($ids) AND name<>'$mark synthetic')=0 AND (SELECT COUNT(*) FROM el_tester WHERE exam_id IN ($ids) AND name<>'$mark synthetic')=0 AND (SELECT COUNT(*) FROM el_user_exam WHERE exam_id IN ($ids))=0 AND (SELECT COUNT(*) FROM el_exam_depart WHERE exam_id IN ($ids))=0,1,JSON_EXTRACT('HTTP_UNKNOWN_OWNER_STOP','\$'));
SELECT IF((SELECT COUNT(*) FROM el_mng_exam_profile p WHERE p.exam_id NOT IN ($ids) AND p.bundle_id IN (SELECT bundle_id FROM el_mng_exam_profile WHERE exam_id IN ($ids)))=0 AND (SELECT COUNT(*) FROM el_mng_paper_snapshot s WHERE s.exam_id NOT IN ($ids) AND s.bundle_id IN (SELECT bundle_id FROM el_mng_exam_profile WHERE exam_id IN ($ids)))=0,1,JSON_EXTRACT('HTTP_SHARED_BUNDLE_STOP','\$'));
SELECT GROUP_CONCAT(QUOTE(bundle_id)) INTO @bundles FROM el_mng_exam_profile WHERE exam_id IN ($ids);
SELECT GROUP_CONCAT(QUOTE(id)) INTO @papers FROM el_paper WHERE exam_id IN ($ids);
SELECT GROUP_CONCAT(QUOTE(id)) INTO @runs FROM el_mng_result_run WHERE exam_id IN ($ids);
SELECT GROUP_CONCAT(QUOTE(id)) INTO @reports FROM el_mng_report_revision WHERE exam_id IN ($ids);
SELECT GROUP_CONCAT(QUOTE(a.id)) INTO @keys FROM el_mng_report_audit a JOIN el_mng_report_revision r ON r.id=a.report_id WHERE r.exam_id IN ($ids);
SET @s=IF(@keys IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_report_audit WHERE id IN (',@keys,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=IF(@papers IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_report_current WHERE paper_id IN (',@papers,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=IF(@reports IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_report_revision WHERE id IN (',@reports,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(d.id)) INTO @keys FROM el_mng_result_dimension d JOIN el_mng_result_run r ON r.id=d.run_id WHERE r.exam_id IN ($ids);
SET @s=IF(@keys IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_result_dimension WHERE id IN (',@keys,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(m.id)) INTO @keys FROM el_mng_result_module m JOIN el_mng_result_run r ON r.id=m.run_id WHERE r.exam_id IN ($ids);
SET @s=IF(@keys IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_result_module WHERE id IN (',@keys,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=IF(@runs IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_runtime_receipt WHERE run_id IN (',@runs,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=IF(@runs IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_result_run WHERE id IN (',@runs,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(q.id)) INTO @keys FROM el_mng_paper_question_snapshot q JOIN el_paper p ON p.id=q.paper_id WHERE p.exam_id IN ($ids);
SET @s=IF(@keys IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_paper_question_snapshot WHERE id IN (',@keys,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=IF(@papers IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_paper_snapshot WHERE paper_id IN (',@papers,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(a.id)) INTO @keys FROM el_paper_qu_answer a JOIN el_paper p ON p.id=a.paper_id WHERE p.exam_id IN ($ids);
SET @s=IF(@keys IS NULL,'SELECT 0',CONCAT('DELETE FROM el_paper_qu_answer WHERE id IN (',@keys,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(q.id)) INTO @keys FROM el_paper_qu q JOIN el_paper p ON p.id=q.paper_id WHERE p.exam_id IN ($ids);
SET @s=IF(@keys IS NULL,'SELECT 0',CONCAT('DELETE FROM el_paper_qu WHERE id IN (',@keys,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(id)) INTO @keys FROM el_candidate WHERE exam_id IN ($ids);
SET @s=IF(@keys IS NULL,'SELECT 0',CONCAT('DELETE FROM el_candidate WHERE id IN (',@keys,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(id)) INTO @keys FROM el_tester WHERE exam_id IN ($ids);
SET @s=IF(@keys IS NULL,'SELECT 0',CONCAT('DELETE FROM el_tester WHERE id IN (',@keys,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SET @s=IF(@papers IS NULL,'SELECT 0',CONCAT('DELETE FROM el_paper WHERE id IN (',@papers,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
DELETE FROM el_mng_exam_profile WHERE exam_id IN ($ids);
SET @s=IF(@bundles IS NULL,'SELECT 0',CONCAT('DELETE FROM el_mng_definition_bundle WHERE id IN (',@bundles,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
SELECT GROUP_CONCAT(QUOTE(id)) INTO @keys FROM el_exam_repo WHERE exam_id IN ($ids);
SET @s=IF(@keys IS NULL,'SELECT 0',CONCAT('DELETE FROM el_exam_repo WHERE id IN (',@keys,')')); PREPARE o FROM @s; EXECUTE o; DEALLOCATE PREPARE o;
DELETE FROM el_exam WHERE id IN ($ids);
COMMIT;
SELECT 'HTTP_EXACT_OWNED_CLEANUP_COMMITTED';
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
sha256sum /etc/systemd/system/talent-assessment.service "$root/server" > "$evidence/runtime-after.sha256"
for name in legacy pdf config runtime; do cmp "$evidence/$name-before.sha256" "$evidence/$name-after.sha256"; done
[ "$(find "$private" -type f | wc -l)" = 0 ]
for s in talent-assessment nginx mysql; do systemctl is-active "$s"; done
curl --fail --silent --max-time 10 http://127.0.0.1:8092/health
printf '\nHTTP_CURRENT_BASELINE_UNCHANGED=1 OLD_PDFS=465 PRIVATE_FILES=0 OWNED_RESIDUAL=0\n'
date -u +%FT%TZ