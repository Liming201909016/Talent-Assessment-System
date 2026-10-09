#!/usr/bin/env bash
# Third and final subprocess: command-local cache redirect, no shared repair.
set -euo pipefail
umask 077
evidence=${1:?}
[[ "$evidence" =~ ^/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9/lo_syscall_20261005_[a-zA-Z0-9]{8}$ ]]
[ "$(id -u)" = 0 ] && [ "$(hostname)" = vm-ubuntu-go-dev ]
[ "$(cat "$evidence/ownership")" = mng-synthetic-only-syscall-v1 ]
[ ! -e "$evidence/targeted.trace" ]
[ "$(sha256sum "$evidence/synthetic.docx" | cut -d' ' -f1)" = dadc9610b59be8deb4eabb497c4a7ec5bc5f2b09de0c81d77d3992e4d2583049 ]
pid=$(systemctl show talent-assessment -p MainPID --value)
child_env=()
while IFS= read -r -d '' item; do
 case "${item%%=*}" in HOME|PATH|TMPDIR|TMP|TEMP|XDG_RUNTIME_DIR|XDG_CONFIG_HOME|XDG_CACHE_HOME|LANG|LC_ALL) child_env+=("$item");; esac
done < /proc/"$pid"/environ
unset item
workspace=$(mktemp -d -t phase1-word-pdf-XXXXXXXX)
printf '%s\n' "$evidence" > "$workspace/owned-targeted"
finish() {
 code=$?; trap - EXIT
 if [[ "$workspace" = /tmp/phase1-word-pdf-* ]] && [ "$(cat "$workspace/owned-targeted")" = "$evidence" ]; then rm -rf -- "$workspace"; fi
 find "$evidence" -type d -exec chmod 0700 {} +
 find "$evidence" -type f -exec chmod 0600 {} +
 printf 'TARGETED_SCRIPT_EXIT=%s WORKSPACE_REMAINING=%s\n' "$code" "$([ -e "$workspace" ] && echo 1 || echo 0)"
 exit "$code"
}
trap finish EXIT
trap 'exit 130' HUP INT TERM
cache=/var/spool/libreoffice/uno_packages/cache
snapshot_cache() { find "$cache" -printf '%P %y %U %G %m %s %T@ %C@\n' | sort; find "$cache" -type f -print0 | sort -z | xargs -0 -r sha256sum; }
snapshot_cache > "$evidence/cache-before-targeted.private"
cp "$evidence/synthetic.docx" "$workspace/synthetic.docx"
mkdir -m 0700 "$workspace/shared-cache"
chmod 0600 "$workspace/synthetic.docx" "$workspace/owned-targeted"
chown liming:liming "$workspace" "$workspace/synthetic.docx" "$workspace/owned-targeted" "$workspace/shared-cache"
sudo -u liming test -w "$workspace/shared-cache"
sudo -u liming test -r "$workspace/synthetic.docx"
printf '%s\n' "$workspace" > "$evidence/targeted.workspace"
cd /opt/talent-assessment
begin=$(date -u +%FT%TZ)
set +e
(ulimit -c 0; ulimit -f 16384; timeout --signal=TERM --kill-after=5s 75s strace -f -u liming -s 256 -e trace=openat,access,newfstatat,stat,lstat,readlink,readlinkat,mkdir,mkdirat,execve,clone,clone3,chdir -o "$evidence/targeted.trace" /usr/bin/env -i "${child_env[@]}" /usr/bin/libreoffice "-env:UserInstallation=file://$workspace/profile" "-env:UNO_SHARED_PACKAGES_CACHE=file://$workspace/shared-cache" --headless --convert-to pdf --outdir "$workspace" "$workspace/synthetic.docx") > "$evidence/targeted.stdout" 2> "$evidence/targeted.stderr"
rc=$?
set -e
valid=0; bytes=0
if [ -f "$workspace/synthetic.pdf" ]; then bytes=$(stat -c %s "$workspace/synthetic.pdf"); [ "$(head -c 5 "$workspace/synthetic.pdf")" != %PDF- ] || valid=1; fi
printf 'CASE=targeted TRACE_EXIT=%s PDF_VALID=%s PDF_BYTES=%s BEGIN=%s END=%s\n' "$rc" "$valid" "$bytes" "$begin" "$(date -u +%FT%TZ)" | tee "$evidence/targeted.summary"
for stream in trace stdout stderr; do printf 'CASE=targeted STREAM=%s BYTES=%s SHA=%s\n' "$stream" "$(wc -c < "$evidence/targeted.$stream")" "$(sha256sum "$evidence/targeted.$stream" | cut -d' ' -f1)"; done
printf 'CASE=targeted DEPLOYMENT_CLASS=%s GENERIC_APP_CLASS=%s ORIGINAL_SHARED_DIRECTORY_CALLS=%s\n' "$(grep -Fc DeploymentException "$evidence/targeted.stderr" || true)" "$(grep -Fc 'Unspecified Application Error' "$evidence/targeted.stderr" || true)" "$(grep -Fc '"/usr/lib/libreoffice/share/uno_packages/cache/uno_packages"' "$evidence/targeted.trace" || true)"
snapshot_cache > "$evidence/cache-after-targeted.private"
if cmp -s "$evidence/cache-before-targeted.private" "$evidence/cache-after-targeted.private"; then printf 'TARGETED_SHARED_CACHE_BYTES_METADATA_UNCHANGED=1\n'; else printf 'TARGETED_SHARED_CACHE_BYTES_METADATA_UNCHANGED=0\n'; fi
printf 'PID_BEFORE=%s PID_AFTER=%s\n' "$pid" "$(systemctl show talent-assessment -p MainPID --value)"
[ "$rc" != 124 ] && [ "$rc" != 137 ] && [ "$rc" != 153 ]
[ "$(stat -c %s "$evidence/targeted.trace")" -lt 16777216 ]
date -u +%FT%TZ