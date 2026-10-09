#!/usr/bin/env bash
# Approved A only: one empty root-owned cache directory 0700 -> 0755.
set -euo pipefail
umask 077
if [ "${1:-}" = --case ]; then
 label=${2:?}; payload=${3:?}
 [[ "$label" =~ ^(normal|matched)$ ]]
 [[ "$payload" =~ ^/tmp/mng_lo_permission_20261005_[a-zA-Z0-9]{8}$ ]]
 [ "$(id -un)" = liming ]
 [ "$(cat "$payload/ownership")" = mng-lo-permission-synthetic-v1 ]
 [ "$(sha256sum "$payload/synthetic.docx" | cut -d' ' -f1)" = dadc9610b59be8deb4eabb497c4a7ec5bc5f2b09de0c81d77d3992e4d2583049 ]
 out=$payload/$label; mkdir -m 0700 "$out"
 workspace=$(mktemp -d -t phase1-word-pdf-XXXXXXXX)
 printf '%s\n' "$payload" > "$workspace/owned-permission"
 finish_case() {
  rc=$?; trap - EXIT
  if [[ "$workspace" = /tmp/phase1-word-pdf-* ]] && [ "$(cat "$workspace/owned-permission")" = "$payload" ]; then rm -rf -- "$workspace"; fi
  printf 'CASE=%s WORKSPACE_REMAINING=%s CASE_EXIT=%s\n' "$label" "$([ -e "$workspace" ] && echo 1 || echo 0)" "$rc"
  exit "$rc"
 }
 trap finish_case EXIT
 trap 'exit 130' HUP INT TERM
 cp "$payload/synthetic.docx" "$workspace/synthetic.docx"
 [ "$(stat -c '%u:%g:%a' "$workspace")" = "$(id -u):$(id -g):700" ]
 [ -r "$workspace/synthetic.docx" ] && [ -w "$workspace" ]
 cd /opt/talent-assessment
 begin=$(date -u +%FT%TZ)
 set +e
 (ulimit -c 0; ulimit -f 16384; timeout --signal=TERM --kill-after=5s 90s /usr/bin/libreoffice "-env:UserInstallation=file://$workspace/profile" --headless --convert-to pdf --outdir "$workspace" "$workspace/synthetic.docx") > "$out/stdout.private" 2> "$out/stderr.private"
 loexit=$?
 set -e
 valid=0; bytes=0
 if [ -f "$workspace/synthetic.pdf" ]; then
  bytes=$(stat -c %s "$workspace/synthetic.pdf")
  if [ "$bytes" -ge 1024 ] && [ "$(head -c 5 "$workspace/synthetic.pdf")" = %PDF- ]; then valid=1; fi
 fi
 printf 'CASE=%s LO_EXIT=%s PDF_VALID=%s PDF_BYTES=%s BEGIN=%s END=%s DEFAULT_SHARED_CACHE=1\n' "$label" "$loexit" "$valid" "$bytes" "$begin" "$(date -u +%FT%TZ)" | tee "$out/summary"
 printf 'CASE=%s STDERR_BYTES=%s STDERR_SHA=%s DEPLOYMENT_CLASS=%s PERMISSION_CLASS=%s GENERIC_APP_CLASS=%s\n' "$label" "$(wc -c < "$out/stderr.private")" "$(sha256sum "$out/stderr.private" | cut -d' ' -f1)" "$(grep -Fc DeploymentException "$out/stderr.private" || true)" "$(grep -Eic 'permission denied|access denied|operation not permitted' "$out/stderr.private" || true)" "$(grep -Fc 'Unspecified Application Error' "$out/stderr.private" || true)"
 [ "$loexit" = 0 ] && [ "$valid" = 1 ]
 cp "$workspace/synthetic.pdf" "$out/synthetic.pdf"
 pdfinfo "$out/synthetic.pdf" > "$out/pdfinfo.private" 2> "$out/pdfinfo.stderr.private"
 pdffonts "$out/synthetic.pdf" > "$out/pdffonts.private" 2> "$out/pdffonts.stderr.private"
 pdftotext -enc UTF-8 "$out/synthetic.pdf" "$out/text.private" 2> "$out/pdftotext.stderr.private"
 pages=$(awk '/^Pages:/ {print $2}' "$out/pdfinfo.private")
 [ "$pages" = 9 ] && [ -s "$out/text.private" ]
 grep -q TEST "$out/text.private"
 nonempty=0
 for ((i=1;i<=pages;i++)); do
  pdftotext -f "$i" -l "$i" -enc UTF-8 "$out/synthetic.pdf" - 2>> "$out/pdftotext.stderr.private" | grep -q '[^[:space:]]'
  nonempty=$((nonempty+1))
 done
 read -r fonts unembedded < <(awk 'NR>2 && NF {n++; if ($(NF-4)!="yes") bad++} END {print n+0,bad+0}' "$out/pdffonts.private")
 [ "$fonts" -gt 0 ] && [ "$unembedded" = 0 ]
 printf 'CASE=%s PAGES=%s NONEMPTY_PAGES=%s TEST_TEXT=1 FONTS=%s UNEMBEDDED=%s PDF_SHA=%s TEXT_SHA=%s\n' "$label" "$pages" "$nonempty" "$fonts" "$unembedded" "$(sha256sum "$out/synthetic.pdf" | cut -d' ' -f1)" "$(sha256sum "$out/text.private" | cut -d' ' -f1)"
 exit 0
fi
[ "$(hostname)" = vm-ubuntu-go-dev ] && [ "$(id -u)" = 0 ]
root=/opt/talent-assessment
parent=$root/backups/mng_phase1_20261003_112853_fd24d4ee35a9
prior=$parent/lo_syscall_20261005_roBjxbvP
source=$parent/lo_diag_20261005_8BorrQx7/payload-receipt/synthetic.docx
c=/var/spool/libreoffice/uno_packages/cache/uno_packages
cache=${c%/*}
parents=(/ /var /var/spool /var/spool/libreoffice /var/spool/libreoffice/uno_packages "$cache")
pid=$(systemctl show talent-assessment -p MainPID --value)
child_env=(); safe_env=()
while IFS= read -r -d '' item; do
 case "${item%%=*}" in
  HOME|PATH|TMPDIR|TMP|TEMP|XDG_RUNTIME_DIR|XDG_CONFIG_HOME|XDG_CACHE_HOME|LANG|LC_ALL) child_env+=("$item"); safe_env+=("--setenv=$item");;
  REPORT_EFFECTIVE_ENV|MNG_TEST_REPORT_ENV|APP_ENV) export "$item";;
 esac
done < /proc/"$pid"/environ
unset item
[ "$APP_ENV" = production ] && [ "$REPORT_EFFECTIVE_ENV" = staging ] && [ "$MNG_TEST_REPORT_ENV" = staging ]
[ "$(systemctl show talent-assessment -p User --value)" = liming ]
[ "$(systemctl show talent-assessment -p WorkingDirectory --value)" = "$root" ]
[ "$(sha256sum /proc/"$pid"/exe | cut -d' ' -f1)" = 8fb264e669b9667c9672905dadf43fbd4654a669b285643db938cb0d7b33ae93 ]
[ "$(stat -c '%u:%g:%a' "$parent")" = 0:0:700 ]
[ "$(sha256sum "$source" | cut -d' ' -f1)" = dadc9610b59be8deb4eabb497c4a7ec5bc5f2b09de0c81d77d3992e4d2583049 ]
acl() {
 perl - "$c" "$cache" <<'ACL'
use strict; use warnings; use Errno qw(ENODATA); require 'syscall.ph';
for my $arg (@ARGV) { my $p="$arg"; for my $name ('system.posix_acl_access','system.posix_acl_default') { my $n="$name"; my $buf="\0" x 4096; my $r=syscall(&SYS_getxattr,$p,$n,$buf,4096); die "ACL_DRIFT\n" unless $r==-1 && $! == ENODATA; } }
print "ACL_ACCESS_DEFAULT_ENODATA=4\n";
ACL
}
cache_snapshot() { find "$cache" -printf '%P %y %U %G %m %s %T@ %C@\n' | sort; find "$cache" -type f -print0 | sort -z | xargs -0 -r sha256sum; }
precondition() {
 [ "$(realpath -e "$c")" = "$c" ] && [ ! -L "$c" ]
 [ "$(stat -c '%u:%g:%a' "$c")" = 0:0:700 ]
 [ -z "$(find "$c" -mindepth 1 -print -quit)" ]
 for x in "${parents[@]}"; do [ ! -L "$x" ]; [ "$(stat -c '%u:%g:%a' "$x")" = 0:0:755 ]; done
 acl
 cache_snapshot | cmp - "$prior/cache-after-targeted.private"
 [ "$(pgrep -c -f '^(/usr/lib/libreoffice/program/(soffice.bin|oosplash)|/usr/bin/unopkg)' || true)" = 0 ]
}
precondition
evidence=$(mktemp -d "$parent/lo_permission_20261005_XXXXXXXX")
printf 'mng-lo-permission-approved-A-v1\n' > "$evidence/ownership"
payload=''; unit=''; changed=0; verified=0
finish() {
 rc=$?; trap - EXIT
 if [ "$changed" = 1 ] && [ "$verified" = 0 ]; then
  if [ ! -L "$c" ] && [ "$(stat -c '%u:%g' "$c")" = 0:0 ]; then
   chmod 0700 -- "$c"
   printf 'PERMISSION=ROLLEDBACK MODE=%s\n' "$(stat -c %a "$c")" | tee "$evidence/rollback.receipt"
  else printf 'ROLLBACK_BLOCKED_METADATA_DRIFT=1\n' | tee "$evidence/rollback.receipt"; fi
 fi
 if [ -n "$unit" ] && [ "$(systemctl show "$unit.service" -p LoadState --value 2>/dev/null)" = loaded ]; then systemctl stop "$unit.service" > "$evidence/unit-stop.private" 2>&1 || true; fi
 if [ -n "$payload" ] && [[ "$payload" =~ ^/tmp/mng_lo_permission_20261005_[a-zA-Z0-9]{8}$ ]] && [ -f "$payload/ownership" ] && [ "$(cat "$payload/ownership")" = mng-lo-permission-synthetic-v1 ]; then
  cp -r -- "$payload" "$evidence/payload-evidence"
  rm -rf -- "$payload"
 fi
 printf 'SCRIPT_EXIT=%s PERMISSION_VERIFIED=%s MODE=%s PAYLOAD_REMAINING=%s EVIDENCE=%s\n' "$rc" "$verified" "$(stat -c %a "$c")" "$([ -n "$payload" ] && [ -e "$payload" ] && echo 1 || echo 0)" "$evidence" | tee "$evidence/final.receipt"
 (cd "$evidence"; find . -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum > SHA256SUMS; sha256sum -c SHA256SUMS > manifest-check.private)
 printf 'MANIFEST_SHA=%s EVIDENCE_DIR_MODE=%s\n' "$(sha256sum "$evidence/SHA256SUMS" | cut -d' ' -f1)" "$(stat -c '%u:%g:%a' "$evidence")"
 exit "$rc"
}
trap finish EXIT
trap 'exit 130' HUP INT TERM
stat -c '%n %u %g %a %s %y %z %w' "$c" > "$evidence/target-before.private"
stat -c '%n %u %g %a %s %y %z %w' "${parents[@]}" > "$evidence/parents-before.private"
cache_snapshot > "$evidence/cache-before.private"
acl > "$evidence/acl-before.private"
sha256sum "$evidence/target-before.private" "$evidence/parents-before.private" "$evidence/cache-before.private" "$evidence/acl-before.private" > "$evidence/original-metadata.sha256"
props=(User Group WorkingDirectory ProtectHome ProtectSystem PrivateTmp PrivateDevices NoNewPrivileges RestrictNamespaces UMask LimitNOFILE MemoryMax TasksMax RootDirectory AppArmorProfile SELinuxContext ReadWritePaths ReadOnlyPaths InaccessiblePaths SystemCallFilter SystemCallArchitectures RestrictAddressFamilies CapabilityBoundingSet AmbientCapabilities)
show=(); for p in "${props[@]}"; do show+=("-p" "$p"); done
systemctl show talent-assessment "${show[@]}" > "$evidence/service-props-before.private"
for p in RootDirectory AppArmorProfile SELinuxContext ReadWritePaths ReadOnlyPaths InaccessiblePaths SystemCallArchitectures AmbientCapabilities; do [ -z "$(systemctl show talent-assessment -p "$p" --value)" ]; done
[ "$(systemctl show talent-assessment -p SystemCallFilter --value)" = '~' ]
[ "$(systemctl show talent-assessment -p RestrictAddressFamilies --value)" = '~' ]
sandbox=()
for p in ProtectHome ProtectSystem PrivateTmp PrivateDevices NoNewPrivileges RestrictNamespaces UMask LimitNOFILE MemoryMax TasksMax CapabilityBoundingSet; do sandbox+=("--property=$p=$(systemctl show talent-assessment -p "$p" --value)"); done
dump_args=(--single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary)
legacy=(el_exam el_exam_repo el_repo el_qu el_qu_repo el_qu_answer el_paper el_paper_qu el_paper_qu_answer el_candidate el_tester el_mbti_answer)
baseline() {
 suffix=$1
 mysqldump "${dump_args[@]}" element "${legacy[@]}" 2> "$evidence/dump-$suffix.stderr.private" | sha256sum | cut -d' ' -f1 > "$evidence/legacy-$suffix.sha256"
 mysqldump "${dump_args[@]}" element el_repo el_qu el_qu_repo el_qu_answer 2> "$evidence/source-$suffix.stderr.private" | sha256sum > "$evidence/source-$suffix.sha256"
 find "$root/tmp" /data/uploadPath -type f -iname '*.pdf' -print0 | sort -z | xargs -0 -r sha256sum > "$evidence/pdf-$suffix.sha256"
 { find "$root/dist" "$root/configs" /etc/systemd/system/talent-assessment.service.d -type f -print0 | sort -z | xargs -0 sha256sum; sha256sum /etc/systemd/system/talent-assessment.service "$root/server"; } > "$evidence/assets-$suffix.sha256"
}
state() {
 suffix=$1
 mysql --batch --skip-column-names element -e "SELECT 'COUNTS',(SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()),(SELECT COUNT(*) FROM el_exam),(SELECT COUNT(*) FROM el_paper),(SELECT COUNT(*) FROM el_paper_qu),(SELECT COUNT(*) FROM el_paper_qu_answer),(SELECT COUNT(*) FROM el_candidate),(SELECT COUNT(*) FROM el_tester); SELECT 'RESTRICT_FKS',COUNT(*) FROM information_schema.referential_constraints WHERE constraint_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' AND update_rule='RESTRICT' AND delete_rule='RESTRICT';" > "$evidence/state-$suffix.tsv"
 n=0
 while IFS= read -r table; do
  [[ "$table" =~ ^el_mng_[a-z_]+$ ]]
  rows=$(mysql --batch --skip-column-names element -e "SELECT COUNT(*) FROM $table")
  [ "$rows" = 0 ]; n=$((n+1)); printf 'SIDECAR=%s ROWS=%s\n' "$table" "$rows" >> "$evidence/state-$suffix.tsv"
 done < <(mysql --batch --skip-column-names element -e "SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name")
 [ "$n" = 11 ]
 [ "$(find "$root/private/management-traits-test-reports" -type f | wc -l)" = 0 ]
}
baseline before; state before
[ "$(cat "$evidence/legacy-before.sha256")" = 051795bd969650d0e55da676083078255331c66405b5fff0ee4ca9d0a7641d0b ]
[ "$(wc -l < "$evidence/pdf-before.sha256")" = 465 ]
for category in legacy source pdf assets; do cmp "$evidence/$category-before.sha256" "$prior/$category-after.sha256"; done
(cd "$parent/cleanup_20261005_6jpcfVUo"; sha256sum -c current-db-SHA256SUMS; gzip -t element-current.sql.gz) > "$evidence/backup-check.private" 2>&1
printf 'PREFLIGHT_CURRENT_BASELINE_MATCH=1 SIDECAR_ZERO=11 PRIVATE_FILES=0 BACKUP_GZIP_SHA_PASS=1\n'
# Inputs are created by liming, not placed behind the root-only evidence path.
payload=$(runuser -u liming -- mktemp -d -t mng_lo_permission_20261005_XXXXXXXX)
runuser -u liming -- bash -c 'umask 077; printf "mng-lo-permission-synthetic-v1\n" > "$1/ownership"; cat > "$1/synthetic.docx"' _ "$payload" < "$source"
runuser -u liming -- bash -c 'umask 077; cat > "$1/probe.sh"' _ "$payload" < "${1:?executed-probe-path-required}"
cp "${1}" "$evidence/executed-probe.sh"
unit=mng-lo-permission-${payload##*_}
precondition
stat -c '%n %u %g %a %s %y %z %w' "${parents[@]}" | cmp - "$evidence/parents-before.private"
# The only shared chmod. Mark first, so any failure runs the exact rollback.
changed=1
chmod 0755 -- "$c"
[ "$(stat -c '%u:%g:%a' "$c")" = 0:0:755 ]
runuser -u liming -- test -r "$c"
runuser -u liming -- test -x "$c"
if runuser -u liming -- test -w "$c"; then exit 1; fi
if runuser -u liming -- test -w "$cache"; then exit 1; fi
runuser -u liming -- perl -e 'opendir(my $d,"/usr/lib/libreoffice/share/uno_packages/cache/uno_packages") or die "DEFAULT_CACHE_OPEN_FAILED\n"; closedir($d); print "DEFAULT_SHARED_PATH_OPEN_SUCCESS=1\n";'
runuser -u liming -- env -i "${child_env[@]}" /usr/bin/bash "$payload/probe.sh" --case normal "$payload" > "$evidence/normal-summary" 2> "$evidence/normal-launch.stderr.private"
cat "$evidence/normal-summary"
systemd-run --unit "$unit" --wait --pipe --collect --service-type=exec --uid=liming --gid=liming --working-directory="$root" --property=RuntimeMaxSec=100s "${sandbox[@]}" "${safe_env[@]}" /usr/bin/env -i "${child_env[@]}" /usr/bin/bash "$payload/probe.sh" --case matched "$payload" > "$evidence/matched-summary" 2> "$evidence/transient.stderr.private"
cat "$evidence/matched-summary"
cmp "$payload/normal/text.private" "$payload/matched/text.private"
printf 'TWO_DEFAULT_CONVERSIONS_TEXT_SAME=1 ROOT_LO_RUNS=0 CACHE_OVERRIDE=0\n'
baseline after; state after
for category in legacy source pdf assets; do cmp "$evidence/$category-before.sha256" "$evidence/$category-after.sha256"; printf 'BASELINE_%s_UNCHANGED=1\n' "$category"; done
cmp "$evidence/state-before.tsv" "$evidence/state-after.tsv"
cat "$evidence/state-after.tsv"
stat -c '%n %u %g %a %s %y %z %w' "${parents[@]}" > "$evidence/parents-after.private"
cmp "$evidence/parents-before.private" "$evidence/parents-after.private"
acl > "$evidence/acl-after.private"; cmp "$evidence/acl-before.private" "$evidence/acl-after.private"
cache_snapshot > "$evidence/cache-after.private"
[ "$(find "$cache" -mindepth 1 | wc -l)" = 1 ]
[ -z "$(find "$c" -mindepth 1 -print -quit)" ]
[ "$(stat -c '%u:%g:%a' "$c")" = 0:0:755 ]
# Apart from target mode/ctime, the cache snapshot must remain identical.
sed -E '/^uno_packages /s/ 700 / 755 /; /^uno_packages /s/ [^ ]+$/ CTIME_CHANGED/' "$evidence/cache-before.private" > "$evidence/cache-normalized-before.private"
sed -E '/^uno_packages /s/ [^ ]+$/ CTIME_CHANGED/' "$evidence/cache-after.private" > "$evidence/cache-normalized-after.private"
cmp "$evidence/cache-normalized-before.private" "$evidence/cache-normalized-after.private"
systemctl show talent-assessment "${show[@]}" > "$evidence/service-props-after.private"
cmp "$evidence/service-props-before.private" "$evidence/service-props-after.private"
[ "$(systemctl show talent-assessment -p MainPID --value)" = "$pid" ]
[ "$(systemctl show "$unit.service" -p LoadState --value 2>/dev/null)" = not-found ]
for s in talent-assessment nginx mysql; do printf 'SERVICE=%s ' "$s"; systemctl is-active "$s"; done
curl --fail --silent --max-time 10 http://127.0.0.1:8092/health; printf '\n'
printf 'PID_BEFORE=%s PID_AFTER=%s SHARED_DIRECTORY_CHANGES=1 SHARED_FILE_CHMOD=0 SHARED_CHOWN=0 PARENTS_SAME=1 SHARED_NEW_FILES=0 SIDECAR_ZERO=11 PRIVATE_FILES=0 HTTP_WRITES=0\n' "$pid" "$(systemctl show talent-assessment -p MainPID --value)"
verified=1
printf 'PERMISSION=APPLIED_VERIFIED OWNER=0:0 MODE=755 ROLLBACK_NOT_NEEDED=1\n'
date -u +%FT%TZ