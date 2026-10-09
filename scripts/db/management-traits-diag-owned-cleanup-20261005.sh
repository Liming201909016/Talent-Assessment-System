#!/usr/bin/env bash
# One previously receipted staging exam; no business/configuration changes.
set -euo pipefail
umask 077
[ "$(hostname)" = vm-ubuntu-go-dev ] && [ "$(id -u)" = 0 ]
root=/opt/talent-assessment
parent=$root/backups/mng_phase1_20261003_112853_fd24d4ee35a9
old=$parent/http_1455f3ab7923
exam=1791105048344426522
paper=c6d52955-aeec-4abe-8468-c5f350f1df84
candidate=a3889750-1307-4fed-bc27-e89613d2f518
run=ea660428-f2b5-4445-b6e7-9d8b2351e00d
bundle=7a20628b-d061-4606-9874-db38aa198f80
q() { mysql --batch --raw --skip-column-names element "$@"; }
pid=$(systemctl show talent-assessment -p MainPID --value)
while IFS= read -r -d '' item; do
 case "${item%%=*}" in REPORT_EFFECTIVE_ENV|MNG_TEST_REPORT_ENV) export "$item";; esac
done < /proc/"$pid"/environ
unset item
[ "${REPORT_EFFECTIVE_ENV:-}" = staging ] && [ "${MNG_TEST_REPORT_ENV:-}" = staging ]
[ "$(sha256sum "$root/server" | cut -d' ' -f1)" = 8fb264e669b9667c9672905dadf43fbd4654a669b285643db938cb0d7b33ae93 ]
[ "$(stat -c '%U:%G:%a' "$parent")" = root:root:700 ]
[ "$(cat "$old/ownership")" = MTH1455f3ab7923 ]
evidence=$(mktemp -d "$parent/cleanup_20261005_XXXXXXXX")
printf '%s\n' mng-diag-exact-exam-1791105048344426522 > "$evidence/ownership"
finish() { code=$?; trap - EXIT; printf 'CLEANUP_EXIT=%s EVIDENCE=%s\n' "$code" "$evidence"; exit "$code"; }
trap finish EXIT
mysqldump --single-transaction --quick --routines --triggers --events --set-gtid-purged=OFF --no-tablespaces element 2> "$evidence/dump.stderr" | gzip > "$evidence/element-current.sql.gz"
gzip -t "$evidence/element-current.sql.gz"
(cd "$evidence"; sha256sum element-current.sql.gz > current-db-SHA256SUMS; sha256sum -c current-db-SHA256SUMS)
chmod 0600 "$evidence"/*
printf 'BACKUP_SHA=%s PERMISSIONS=%s\n' "$(cut -d' ' -f1 "$evidence/current-db-SHA256SUMS")" "$(stat -c '%U:%G:%a' "$evidence")"
dump_args=(--single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary)
legacy=(el_exam el_exam_repo el_repo el_qu el_qu_repo el_qu_answer el_paper el_paper_qu el_paper_qu_answer el_candidate el_tester el_mbti_answer)
nonowned() {
 local t where
 for t in "${legacy[@]}"; do
  where=1
  case "$t" in
   el_exam) where="id<>'$exam'";;
   el_exam_repo|el_tester) where="exam_id<>'$exam' OR exam_id IS NULL";;
   el_candidate) where="id<>'$candidate'";;
   el_paper) where="id<>'$paper'";;
   el_paper_qu|el_paper_qu_answer) where="paper_id<>'$paper' OR paper_id IS NULL";;
  esac
  mysqldump "${dump_args[@]}" "--where=$where" element "$t" | sha256sum | awk -v t="$t" '{print t" "$1}'
 done
}
pdfs() { find "$root/tmp" /data/uploadPath -type f -iname '*.pdf' -print0 | sort -z | xargs -0 -r sha256sum; }
assets() { find "$root/dist" "$root/configs" /etc/systemd/system/talent-assessment.service.d -type f -print0 | sort -z | xargs -0 sha256sum; sha256sum /etc/systemd/system/talent-assessment.service "$root/server"; }
nonowned > "$evidence/nonowned-before.sha256"
pdfs > "$evidence/pdf-before.sha256"
assets > "$evidence/assets-before.sha256"
[ "$(wc -l < "$evidence/pdf-before.sha256")" = 465 ]
[ "$(find "$root/private/management-traits-test-reports" -type f | wc -l)" = 0 ]
mysqldump "${dump_args[@]}" element el_repo el_qu el_qu_repo el_qu_answer | sha256sum > "$evidence/source-before.sha256"
q -e "SELECT 'COUNTS_BEFORE',(SELECT COUNT(*) FROM el_exam),(SELECT COUNT(*) FROM el_paper),(SELECT COUNT(*) FROM el_paper_qu),(SELECT COUNT(*) FROM el_paper_qu_answer),(SELECT COUNT(*) FROM el_candidate),(SELECT COUNT(*) FROM el_tester);" | tee "$evidence/counts-before.tsv"
# Save exact child PK receipt privately, then pin every DELETE to those PKs.
declare -a tables conditions expected pk extra
tables=(el_mng_result_dimension el_mng_result_module el_mng_runtime_receipt el_mng_result_run el_mng_paper_question_snapshot el_mng_paper_snapshot el_paper_qu_answer el_paper_qu el_candidate el_paper el_mng_exam_profile el_exam_repo el_exam)
conditions=("run_id='$run'" "run_id='$run'" "run_id='$run' AND paper_id='$paper' AND exam_id='$exam'" "id='$run' AND paper_id='$paper' AND exam_id='$exam' AND participant_id='$candidate'" "paper_id='$paper'" "paper_id='$paper' AND exam_id='$exam' AND participant_id='$candidate' AND bundle_id='$bundle'" "paper_id='$paper'" "paper_id='$paper'" "id='$candidate' AND exam_id='$exam' AND paper_id='$paper'" "id='$paper' AND exam_id='$exam'" "exam_id='$exam' AND bundle_id='$bundle'" "exam_id='$exam'" "id='$exam'")
expected=(13 4 1 1 140 1 700 140 1 1 1 1 1)
pk=(id id run_id id id paper_id id id id id exam_id id id)
extra=("run_id='$run'" "run_id='$run'" "paper_id='$paper' AND exam_id='$exam'" "paper_id='$paper' AND exam_id='$exam' AND participant_id='$candidate'" "paper_id='$paper'" "exam_id='$exam' AND participant_id='$candidate' AND bundle_id='$bundle'" "paper_id='$paper'" "paper_id='$paper'" "exam_id='$exam' AND paper_id='$paper'" "exam_id='$exam'" "bundle_id='$bundle'" "exam_id='$exam'" "title='MTH1455f3ab7923-00201-candidate' AND create_time='2026-10-04 17:10:48'")
for i in "${!tables[@]}"; do
 t=${tables[$i]}; p=${pk[$i]}
 q -e "SELECT $p FROM $t WHERE ${conditions[$i]} ORDER BY $p" > "$evidence/$t.ids"
 [ "$(wc -l < "$evidence/$t.ids")" = "${expected[$i]}" ]
done
shared=$(q -e "SELECT (SELECT COUNT(*) FROM el_mng_exam_profile WHERE bundle_id='$bundle' AND exam_id<>'$exam')+(SELECT COUNT(*) FROM el_mng_paper_snapshot WHERE bundle_id='$bundle' AND paper_id<>'$paper');")
[[ "$shared" =~ ^[0-9]+$ ]]
{
 printf 'SET NAMES utf8mb4; SET SESSION SQL_SAFE_UPDATES=1; SET SESSION range_optimizer_max_mem_size=67108864; START TRANSACTION;\n'
 printf "SELECT id FROM el_exam WHERE id='%s' FOR UPDATE;\n" "$exam"
 printf "SELECT IF((SELECT COUNT(*) FROM el_exam WHERE id='$exam' AND title='MTH1455f3ab7923-00201-candidate' AND create_time='2026-10-04 17:10:48')=1 AND (SELECT COUNT(*) FROM el_candidate WHERE id='$candidate' AND exam_id='$exam' AND paper_id='$paper' AND name='MTH1455f3ab7923 synthetic' AND create_time='2026-10-04 17:10:49' AND IFNULL(pdf_path,'')='')=1,1,JSON_EXTRACT('EXACT_OWNER_STOP','\$'));\n"
 printf "SELECT IF((SELECT COUNT(*) FROM el_paper WHERE exam_id='$exam')=1 AND (SELECT COUNT(*) FROM el_candidate WHERE exam_id='$exam')=1 AND (SELECT COUNT(*) FROM el_tester WHERE exam_id='$exam')=0 AND (SELECT COUNT(*) FROM el_user_exam WHERE exam_id='$exam')=0 AND (SELECT COUNT(*) FROM el_exam_depart WHERE exam_id='$exam')=0 AND (SELECT COUNT(*) FROM el_mng_result_run WHERE exam_id='$exam')=1 AND (SELECT COUNT(*) FROM el_mng_paper_snapshot WHERE exam_id='$exam')=1,1,JSON_EXTRACT('EXTRA_OWNER_STOP','\$'));\n"
 printf "SELECT IF((SELECT COUNT(*) FROM el_mng_report_revision)=0 AND (SELECT COUNT(*) FROM el_mng_report_current)=0 AND (SELECT COUNT(*) FROM el_mng_report_audit)=0,1,JSON_EXTRACT('REPORT_ZERO_STOP','\$'));\n"
 for i in "${!tables[@]}"; do
  t=${tables[$i]}; p=${pk[$i]}; list=''
    while IFS= read -r id; do
     case "$t" in
        el_mng_result_dimension) [[ "$id" = "$run-d-"* && "$id" =~ ^[a-z0-9_-]{1,64}$ ]];;
        el_mng_result_module) [[ "$id" = "$run-m-"* && "$id" =~ ^[a-z0-9_-]{1,64}$ ]];;
        *) [[ "$id" =~ ^[a-f0-9-]{36}$ || "$id" =~ ^[0-9]{19}$ ]];;
     esac
     list="${list:+$list,}'$id'"
    done < "$evidence/$t.ids"
  printf "SELECT IF((SELECT COUNT(*) FROM $t WHERE ${conditions[$i]})=${expected[$i]},1,JSON_EXTRACT('COUNT_STOP_$t','\$'));\n"
  printf "DELETE FROM $t WHERE $p IN ($list) AND ${extra[$i]};\n"
  printf "SET @n=ROW_COUNT(); SELECT 'DELETED_$t',@n; SELECT IF(@n=${expected[$i]},1,JSON_EXTRACT('DELETE_COUNT_STOP_$t','\$'));\n"
  if [ "$t" = el_mng_exam_profile ] && [ "$shared" = 0 ]; then
   printf "SELECT IF((SELECT COUNT(*) FROM el_mng_exam_profile WHERE bundle_id='$bundle')=0 AND (SELECT COUNT(*) FROM el_mng_paper_snapshot WHERE bundle_id='$bundle')=0,1,JSON_EXTRACT('SHARED_BUNDLE_STOP','\$')); DELETE FROM el_mng_definition_bundle WHERE id='$bundle'; SET @n=ROW_COUNT(); SELECT 'DELETED_BUNDLE',@n; SELECT IF(@n=1,1,JSON_EXTRACT('BUNDLE_COUNT_STOP','\$'));\n"
  fi
 done
 printf "COMMIT; SELECT 'EXACT_CLEANUP_COMMITTED';\n"
} > "$evidence/cleanup.sql"
mysql --batch --raw --skip-column-names element < "$evidence/cleanup.sql" > "$evidence/cleanup-output.tsv" 2> "$evidence/cleanup.stderr"
cat "$evidence/cleanup-output.tsv"
printf 'BUNDLE_PRESERVED_EXTERNAL_REFS=%s\n' "$shared"
for i in "${!tables[@]}"; do
 t=${tables[$i]}; n=$(q -e "SELECT COUNT(*) FROM $t WHERE ${conditions[$i]}"); [ "$n" = 0 ]; printf 'OWNED_FINAL=%s ROWS=%s\n' "$t" "$n"
done
for t in $(q -e "SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name"); do
 [[ "$t" =~ ^el_mng_[a-z_]+$ ]]; n=$(q -e "SELECT COUNT(*) FROM $t"); printf 'SIDECAR_FINAL=%s ROWS=%s\n' "$t" "$n"
 if [ "$shared" = 0 ]; then [ "$n" = 0 ]; fi
done
nonowned > "$evidence/nonowned-after.sha256"; cmp "$evidence/nonowned-before.sha256" "$evidence/nonowned-after.sha256"
pdfs > "$evidence/pdf-after.sha256"; cmp "$evidence/pdf-before.sha256" "$evidence/pdf-after.sha256"
assets > "$evidence/assets-after.sha256"; cmp "$evidence/assets-before.sha256" "$evidence/assets-after.sha256"
mysqldump "${dump_args[@]}" element el_repo el_qu el_qu_repo el_qu_answer | sha256sum > "$evidence/source-after.sha256"; cmp "$evidence/source-before.sha256" "$evidence/source-after.sha256"
mysqldump "${dump_args[@]}" element "${legacy[@]}" | sha256sum | cut -d' ' -f1 > "$evidence/legacy-after.sha256"
if cmp -s "$old/legacy-before.sha256" "$evidence/legacy-after.sha256"; then printf 'OCT04_BASELINE_MATCH=1\n'; else printf 'OCT04_BASELINE_MATCH=0 NONOWNED_CURRENT_UNCHANGED=1\n'; fi
q -e "SELECT 'COUNTS_AFTER',(SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()),(SELECT COUNT(*) FROM el_exam),(SELECT COUNT(*) FROM el_paper),(SELECT COUNT(*) FROM el_paper_qu),(SELECT COUNT(*) FROM el_paper_qu_answer),(SELECT COUNT(*) FROM el_candidate),(SELECT COUNT(*) FROM el_tester); SELECT 'MNG_RESTRICT_FKS',COUNT(*) FROM information_schema.referential_constraints WHERE constraint_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' AND update_rule='RESTRICT' AND delete_rule='RESTRICT';" | tee "$evidence/counts-after.tsv"
[ "$(find "$root/private/management-traits-test-reports" -type f | wc -l)" = 0 ]
chmod 0600 "$evidence"/*
for s in talent-assessment nginx mysql; do printf 'SERVICE=%s ' "$s"; systemctl is-active "$s"; done
curl --fail --silent --max-time 10 http://127.0.0.1:8092/health
printf '\nEXACT_OWNED_RESIDUAL=0 PRIVATE_FILES=0 OLD_PDFS=465 CURRENT_NONOWNED_SOURCE_PDF_ASSETS_UNCHANGED=1\n'
date -u +%FT%TZ