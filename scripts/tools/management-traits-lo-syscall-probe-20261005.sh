#!/usr/bin/env bash
# Diagnostic-only: two dedicated synthetic LO subprocesses, never the app PID.
set -euo pipefail
umask 077
[ "$(hostname)" = vm-ubuntu-go-dev ] && [ "$(id -u)" = 0 ]
command -v strace >/dev/null
root=/opt/talent-assessment
parent=$root/backups/mng_phase1_20261003_112853_fd24d4ee35a9
prior=$parent/lo_diag_20261005_8BorrQx7
source=$prior/payload-receipt/synthetic.docx
[ "$(sha256sum "$source" | cut -d' ' -f1)" = dadc9610b59be8deb4eabb497c4a7ec5bc5f2b09de0c81d77d3992e4d2583049 ]
[ "$(stat -c '%U:%G:%a' "$parent")" = root:root:700 ]
pid=$(systemctl show talent-assessment -p MainPID --value)
[ "$(sha256sum /proc/"$pid"/exe | cut -d' ' -f1)" = 8fb264e669b9667c9672905dadf43fbd4654a669b285643db938cb0d7b33ae93 ]
child_env=()
while IFS= read -r -d '' item; do
 case "${item%%=*}" in
  HOME|PATH|TMPDIR|TMP|TEMP|XDG_RUNTIME_DIR|XDG_CONFIG_HOME|XDG_CACHE_HOME|LANG|LC_ALL) child_env+=("$item");;
  REPORT_EFFECTIVE_ENV|MNG_TEST_REPORT_ENV|APP_ENV) export "$item";;
 esac
done < /proc/"$pid"/environ
unset item
[ "${APP_ENV:-}" = production ] && [ "${REPORT_EFFECTIVE_ENV:-}" = staging ] && [ "${MNG_TEST_REPORT_ENV:-}" = staging ]
evidence=$(mktemp -d "$parent/lo_syscall_20261005_XXXXXXXX")
printf 'mng-synthetic-only-syscall-v1\n' > "$evidence/ownership"
workspace=''
finish() {
 code=$?; trap - EXIT
 if [ -n "$workspace" ] && [[ "$workspace" = /tmp/phase1-word-pdf-* ]] && [ -f "$workspace/owned-syscall" ] && [ "$(cat "$workspace/owned-syscall")" = "$evidence" ]; then
  rm -rf -- "$workspace"
 fi
 find "$evidence" -type d -exec chmod 0700 {} +
 find "$evidence" -type f -exec chmod 0600 {} +
 printf 'PROBE_EXIT=%s WORKSPACE_REMAINING=%s EVIDENCE=%s\n' "$code" "$([ -n "$workspace" ] && [ -e "$workspace" ] && echo 1 || echo 0)" "$evidence"
 exit "$code"
}
trap finish EXIT
trap 'exit 130' HUP INT TERM
dump_args=(--single-transaction --quick --skip-lock-tables --set-gtid-purged=OFF --no-tablespaces --skip-comments --compact --hex-blob --order-by-primary)
legacy=(el_exam el_exam_repo el_repo el_qu el_qu_repo el_qu_answer el_paper el_paper_qu el_paper_qu_answer el_candidate el_tester el_mbti_answer)
baseline() {
 local suffix=$1
 mysqldump "${dump_args[@]}" element "${legacy[@]}" 2> "$evidence/dump-$suffix.stderr" | sha256sum | cut -d' ' -f1 > "$evidence/legacy-$suffix.sha256"
 mysqldump "${dump_args[@]}" element el_repo el_qu el_qu_repo el_qu_answer 2> "$evidence/source-$suffix.stderr" | sha256sum > "$evidence/source-$suffix.sha256"
 find "$root/tmp" /data/uploadPath -type f -iname '*.pdf' -print0 | sort -z | xargs -0 -r sha256sum > "$evidence/pdf-$suffix.sha256"
 { find "$root/dist" "$root/configs" /etc/systemd/system/talent-assessment.service.d -type f -print0 | sort -z | xargs -0 sha256sum; sha256sum /etc/systemd/system/talent-assessment.service "$root/server"; } > "$evidence/assets-$suffix.sha256"
}
baseline before
[ "$(wc -l < "$evidence/pdf-before.sha256")" = 465 ]
cp "$source" "$evidence/synthetic.docx"
for label in root liming; do
 workspace=$(mktemp -d -t phase1-word-pdf-XXXXXXXX)
 printf '%s\n' "$evidence" > "$workspace/owned-syscall"
 cp "$source" "$workspace/synthetic.docx"
 chmod 0600 "$workspace"/*
 if [ "$label" = liming ]; then
  chown liming:liming "$workspace" "$workspace"/*
  sudo -u liming test -w "$workspace"
  sudo -u liming test -r "$workspace/synthetic.docx"
  envargs=("${child_env[@]}"); userargs=(-u liming)
 else
  envargs=(HOME=/root PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/snap/bin LANG=C.UTF-8); userargs=()
 fi
 printf '%s\n' "$workspace" > "$evidence/$label.workspace"
 cd "$root"
 begin=$(date -u +%FT%TZ)
 set +e
 (ulimit -c 0; ulimit -f 16384; timeout --signal=TERM --kill-after=5s 75s strace -f "${userargs[@]}" -s 256 -e trace=openat,access,newfstatat,stat,lstat,readlink,readlinkat,mkdir,mkdirat,execve,clone,clone3,chdir -o "$evidence/$label.trace" /usr/bin/env -i "${envargs[@]}" /usr/bin/libreoffice "-env:UserInstallation=file://$workspace/profile" --headless --convert-to pdf --outdir "$workspace" "$workspace/synthetic.docx") > "$evidence/$label.stdout" 2> "$evidence/$label.stderr"
 rc=$?
 set -e
 valid=0; bytes=0
 if [ -f "$workspace/synthetic.pdf" ]; then
  bytes=$(stat -c %s "$workspace/synthetic.pdf")
  [ "$(head -c 5 "$workspace/synthetic.pdf")" != %PDF- ] || valid=1
 fi
 printf 'CASE=%s TRACE_EXIT=%s PDF_VALID=%s PDF_BYTES=%s BEGIN=%s END=%s\n' "$label" "$rc" "$valid" "$bytes" "$begin" "$(date -u +%FT%TZ)" | tee "$evidence/$label.summary"
 for stream in trace stdout stderr; do
  printf 'CASE=%s STREAM=%s BYTES=%s SHA=%s\n' "$label" "$stream" "$(wc -c < "$evidence/$label.$stream")" "$(sha256sum "$evidence/$label.$stream" | cut -d' ' -f1)"
 done
 printf 'CASE=%s DEPLOYMENT_CLASS=%s GENERIC_APP_CLASS=%s\n' "$label" "$(grep -Fc DeploymentException "$evidence/$label.stderr" || true)" "$(grep -Fc 'Unspecified Application Error' "$evidence/$label.stderr" || true)"
 [ "$rc" != 124 ] && [ "$rc" != 137 ] && [ "$rc" != 153 ]
 [ "$(stat -c %s "$evidence/$label.trace")" -lt 16777216 ]
 rm -rf -- "$workspace"; [ ! -e "$workspace" ]; workspace=''
done
# No raw stderr/trace/environment output. Strict known-system path allowlist only.
perl - "$evidence/root.trace" "$evidence/liming.trace" <<'PERL' | tee "$evidence/safe-comparison.tsv"
use strict; use warnings;
my (%root, %fail, %other); my $index=0;
for my $file (@ARGV) {
 open my $fh, '<', $file or die "trace unavailable\n";
 while (<$fh>) {
  next unless /\b(openat|access|newfstatat|stat|lstat|readlink|readlinkat|mkdir|mkdirat)\(/;
  my $call=$1; next unless /"([^"\\]*)"/; my $p=$1;
  next unless $p =~ m{^/(?:usr/lib/libreoffice/|usr/share/libreoffice/|etc/libreoffice/)} && $p =~ m{^/[A-Za-z0-9_./+:-]+$};
  if ($index==0 && /\)\s+=\s+([0-9]+)/) { $root{"$call $p"}=$1; }
  if ($index==1 && /\)\s+=\s+-1\s+(EACCES|EPERM|ENOENT)/) { $fail{"$call $p $1"}++; }
 }
 close $fh; $index++;
}
for my $key (sort keys %fail) {
 my ($call,$path,$errno)=split / /,$key; my $ok=$root{"$call $path"};
 next if $errno eq 'ENOENT' && !defined $ok;
 printf "RESOURCE call=%s path=%s liming_errno=%s failures=%d root_success=%s root_result=%s\n",$call,$path,$errno,$fail{$key},defined($ok)?1:0,defined($ok)?$ok:'NA';
}
PERL
baseline after
for category in legacy source pdf assets; do
 if cmp -s "$evidence/$category-before.sha256" "$evidence/$category-after.sha256"; then printf 'BASELINE_%s_UNCHANGED=1\n' "$category"; else printf 'BASELINE_%s_UNCHANGED=0\n' "$category"; fi
done
mysql --batch --raw --skip-column-names element -e "SELECT 'COUNTS',(SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()),(SELECT COUNT(*) FROM el_exam),(SELECT COUNT(*) FROM el_paper),(SELECT COUNT(*) FROM el_paper_qu),(SELECT COUNT(*) FROM el_paper_qu_answer),(SELECT COUNT(*) FROM el_candidate),(SELECT COUNT(*) FROM el_tester); SELECT 'EXACT_OLD_OWNER',COUNT(*) FROM el_exam WHERE id='1791105048344426522';" | tee "$evidence/counts-final.tsv"
while IFS= read -r table; do
 [[ "$table" =~ ^el_mng_[a-z_]+$ ]]
 printf 'SIDECAR=%s ROWS=%s\n' "$table" "$(mysql --batch --skip-column-names element -e "SELECT COUNT(*) FROM $table")"
done < <(mysql --batch --skip-column-names element -e "SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND LEFT(table_name,7)='el_mng_' ORDER BY table_name")
printf 'PRIVATE_FILES=%s OLD_PDFS=%s PID_BEFORE=%s PID_AFTER=%s\n' "$(find "$root/private/management-traits-test-reports" -type f | wc -l)" "$(wc -l < "$evidence/pdf-after.sha256")" "$pid" "$(systemctl show talent-assessment -p MainPID --value)"
(cd "$parent/cleanup_20261005_6jpcfVUo"; sha256sum -c current-db-SHA256SUMS; gzip -t element-current.sql.gz)
for service in talent-assessment nginx mysql; do printf 'SERVICE=%s ' "$service"; systemctl is-active "$service"; done
curl --fail --silent --max-time 10 http://127.0.0.1:8092/health; printf '\n'
date -u +%FT%TZ