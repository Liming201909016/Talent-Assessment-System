#!/usr/bin/env bash
# Approved preparation only: current backup, isolated restore, structural DDL.
# Never applies SQL to element or switches binaries/services.
set -Eeuo pipefail
umask 077
stamp="$1"
payload="$2"
[[ "$stamp" =~ ^[a-f0-9]{16}$ ]]
[[ "$(hostname)" = iZ0yosjdcen2p4Z ]]
[[ "$(id -un)" = root ]]
app=/opt/talent-assessment
parent="$app/backups"
root="$parent/competency_only_20261008_$stamp"
schema="competency_verify_$stamp"
[[ -d "$parent" && ! -L "$parent" && ! -e "$root" ]]
[[ -f "$app/configs/application-production.yml" ]]
[[ "$(df -Pk "$app" | tail -1 | awk '{print $4}')" -ge 3145728 ]]
mkdir -m 700 "$root"
printf '%s\n' "$stamp" > "$root/ownership"
phase=preflight
created=0
client_config() {
  perl -0777 -ne 'if (/^\s*dsn:\s*(.+)$/m) { $d=$1; $d =~ s/\r$//; $d =~ s/^\s+|\s+$//g; if (substr($d,0,1) eq chr(34) || substr($d,0,1) eq chr(39)) {$d=substr($d,1,-1);} if ($d =~ /^([^:]+):(.*?)\@tcp\(([^:()]+):(\d+)\)\/(element)(?:\?.*)?$/s) {@v=($1,$2,$3,$4);@k=qw(user password host port); print qq([client]\n);for $i(0..3){$v[$i]=~s/\\/\\\\/g;$v[$i]=~s/"/\\"/g;print $k[$i],qq(="$v[$i]"\n);}} else {die qq(DSN_REJECTED\n);}}else{die qq(DSN_MISSING\n);}' "$app/configs/application-production.yml"
}
sql() { mysql --defaults-extra-file=<(client_config) --batch --raw --skip-column-names --connect-timeout=5 "$@"; }
cleanup() {
  code=$?
  trap - EXIT
  cleanup_code=0
  if [[ "$created" = 1 ]]; then
    if [[ "$(cat "$root/ownership")" = "$stamp" && "$schema" =~ ^competency_verify_[a-f0-9]{16}$ ]]; then
      sql -e "DROP DATABASE \`$schema\`;" 2>>"$root/private-errors.log" || cleanup_code=1
    else cleanup_code=1; fi
  fi
  remaining=$(sql -e "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='$schema';" 2>>"$root/private-errors.log") || cleanup_code=1
  printf 'FINAL_PHASE=%s\nRUN_EXIT=%s\nOWNED_SCHEMA_REMAINING=%s\nCLEANUP_EXIT=%s\nBACKUP_ROOT=%s\n' "$phase" "$code" "${remaining:-unknown}" "$cleanup_code" "$root"
  if [[ "$code" = 0 && "$cleanup_code" = 0 && "$remaining" = 0 ]]; then exit 0; else exit 1; fi
}
trap cleanup EXIT
exec 2>>"$root/private-errors.log"
[[ "$(sql -e 'SELECT VERSION();')" = 5.7.44-log ]]
[[ "$(sql -e "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='$schema';")" = 0 ]]
# Definitions with schema-qualified SQL need a separate reviewed restore path.
definitions=$(sql -e "SELECT (SELECT COUNT(*) FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA='element')+(SELECT COUNT(*) FROM information_schema.TRIGGERS WHERE TRIGGER_SCHEMA='element')+(SELECT COUNT(*) FROM information_schema.EVENTS WHERE EVENT_SCHEMA='element');")
[[ "$definitions" = 0 ]]
systemctl show talent-assessment -p MainPID -p ActiveState -p NRestarts > "$root/service.before"
sha256sum "$app/server" "$app/dist/index.html" > "$root/runtime.before"
phase=backup
mysqldump --defaults-extra-file=<(client_config) --single-transaction --routines --triggers --events --hex-blob --skip-lock-tables element | gzip > "$root/element.sql.gz"
gzip -t "$root/element.sql.gz"
# Restore guard: no cross-schema SQL, database creation or stored definitions.
if gzip -dc "$root/element.sql.gz" | grep -Ei '(^CREATE DATABASE|^USE |`element`\.)' >/dev/null; then exit 1; fi
assets=(server dist configs)
for rel in tmp/uploadPath private; do [[ ! -d "$app/$rel" ]] || assets+=("$rel"); done
tar -C "$app" -czf "$root/application.tar.gz" "${assets[@]}"
gzip -t "$root/application.tar.gz"
tar -tzf "$root/application.tar.gz" >/dev/null
if [[ -d /data/uploadPath && ! -L /data/uploadPath ]]; then
  tar -C /data -czf "$root/data-upload.tar.gz" uploadPath
  gzip -t "$root/data-upload.tar.gz"
fi
stat -c '%n %u %g %a %s %y' "$app/server" "$app/dist" "$app/configs" > "$root/original-metadata"
for f in /etc/systemd/system/talent-assessment.service /www/server/nginx/conf/nginx.conf; do
  [[ ! -f "$f" ]] || cp -p "$f" "$root/$(basename "$f").before"
done
find "$root" -maxdepth 1 -type f -exec chmod 600 {} +
(cd "$root"; sha256sum element.sql.gz application.tar.gz > SHA256SUMS; [[ ! -f data-upload.tar.gz ]] || sha256sum data-upload.tar.gz >> SHA256SUMS; sha256sum --check --status SHA256SUMS)
printf 'CURRENT_BACKUP_VERIFIED=1\n'
# Receive only approved migration bytes from the local driver, without scp.
phase=payload
[[ "$payload" = "$parent/competency_payload_20261008_$stamp.b64" && -f "$payload" && ! -L "$payload" ]]
base64 -d "$payload" > "$root/migrations.json"
command -v node >/dev/null
node - "$root" <<'NODE'
const fs=require('fs'),crypto=require('crypto'),path=require('path');
const root=process.argv[2],j=JSON.parse(fs.readFileSync(path.join(root,'migrations.json')));
const allowed=['competency_007_versions.sql','competency_008_phase1_structures.sql','competency_010_phase1_report_framework.sql','competency_011_result_runs.sql','competency_012_v2_dimension_catalog.sql','competency_013_report_result_run_binding.sql'];
if(JSON.stringify(j.map(x=>x.name))!==JSON.stringify(allowed))throw Error('scope');
for(const x of j){const b=Buffer.from(x.base64,'base64');if(crypto.createHash('sha256').update(b).digest('hex')!==x.sha256)throw Error('SHA');fs.writeFileSync(path.join(root,x.name),b,{mode:0o600,flag:'wx'});}
NODE
phase=restore
sql -e "CREATE DATABASE \`$schema\` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"
created=1
gzip -dc "$root/element.sql.gz" | sql "$schema" > "$root/restore.private"
tables=(el_exam el_qu el_qu_answer el_qu_repo el_repo el_paper el_paper_qu el_paper_qu_answer el_candidate el_tester el_competency_dimension el_competency_result el_competency_report el_competency_report_text)
mkdir -m 700 "$root/facts"
for table in "${tables[@]}"; do
  columns=$(sql -e "SET SESSION group_concat_max_len=65536; SELECT GROUP_CONCAT(CONCAT('\`',COLUMN_NAME,'\`') ORDER BY ORDINAL_POSITION) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='$schema' AND TABLE_NAME='$table';")
  if [[ "$table" = el_competency_result ]]; then
    columns=${columns//\`scoring_version\`/IF(LENGTH(scoring_version)=0,0x636f6d706574656e63792d7631,scoring_version)}
  fi
  printf '%s\n' "$columns" > "$root/facts/$table.columns"
  sql "$schema" -e "SELECT $columns FROM \`$table\` ORDER BY 1;" | sha256sum > "$root/facts/$table.before"
done
schema_signature() {
  sql -e "SELECT TABLE_NAME,COLUMN_NAME,ORDINAL_POSITION,COLUMN_TYPE,IS_NULLABLE,COALESCE(COLLATION_NAME,'-') FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='$schema' ORDER BY TABLE_NAME,ORDINAL_POSITION; SELECT TABLE_NAME,INDEX_NAME,NON_UNIQUE,SEQ_IN_INDEX,COLUMN_NAME FROM information_schema.STATISTICS WHERE TABLE_SCHEMA='$schema' ORDER BY TABLE_NAME,INDEX_NAME,SEQ_IN_INDEX; SELECT k.TABLE_NAME,k.CONSTRAINT_NAME,k.ORDINAL_POSITION,k.COLUMN_NAME,k.REFERENCED_TABLE_NAME,k.REFERENCED_COLUMN_NAME,r.UPDATE_RULE,r.DELETE_RULE FROM information_schema.KEY_COLUMN_USAGE k JOIN information_schema.REFERENTIAL_CONSTRAINTS r ON r.CONSTRAINT_SCHEMA=k.CONSTRAINT_SCHEMA AND r.TABLE_NAME=k.TABLE_NAME AND r.CONSTRAINT_NAME=k.CONSTRAINT_NAME WHERE k.TABLE_SCHEMA='$schema' AND k.REFERENCED_TABLE_NAME IS NOT NULL ORDER BY k.TABLE_NAME,k.CONSTRAINT_NAME,k.ORDINAL_POSITION;" | sha256sum
}
files=(competency_007_versions.sql competency_008_phase1_structures.sql competency_010_phase1_report_framework.sql competency_011_result_runs.sql competency_012_v2_dimension_catalog.sql competency_013_report_result_run_binding.sql)
for round in 1 2; do
  phase="migration-round-$round"
  for f in "${files[@]}"; do
    start=$SECONDS
    sql "$schema" < "$root/$f" > "$root/$f.round-$round.private"
    printf 'MIGRATION=%s ROUND=%s EXIT=0 SECONDS=%s\n' "$f" "$round" "$((SECONDS-start))"
  done
  schema_signature > "$root/schema.round-$round.sha"
done
cmp "$root/schema.round-1.sha" "$root/schema.round-2.sha"
phase=facts
# Compare all original columns. Only 007 scoring_version fills empty labels;
# old nonempty labels and every score/answer/PDF pointer must stay unchanged.
for table in "${tables[@]}"; do
  columns=$(cat "$root/facts/$table.columns")
  sql "$schema" -e "SELECT $columns FROM \`$table\` ORDER BY 1;" | sha256sum > "$root/facts/$table.after"
  cmp "$root/facts/$table.before" "$root/facts/$table.after"
done
[[ "$(sql "$schema" -e 'SELECT COUNT(*) FROM el_competency_dimension;')" = 48 ]]
[[ "$(sql -e "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='$schema' AND LEFT(TABLE_NAME,7)='el_mng_';")" = 0 ]]
systemctl show talent-assessment -p MainPID -p ActiveState -p NRestarts > "$root/service.after"
sha256sum "$app/server" "$app/dist/index.html" > "$root/runtime.after"
cmp "$root/service.before" "$root/service.after"
cmp "$root/runtime.before" "$root/runtime.after"
printf 'RESTORED_OLD_FACTS_SAME=1\nFIRST_REPEAT_SCHEMA_SAME=1\nMAIN_RUNTIME_UNCHANGED=1\n'
phase=completed