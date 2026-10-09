#!/usr/bin/env bash
# Runtime-only synthetic conversion, at most root/liming/matched-unit once each.
set -euo pipefail
umask 077
if [ "${1:-}" = --case ]; then
 label=${2:?}; payload=${3:?}; executable=${4:?}
 [[ "$label" =~ ^(root|liming|matched)$ ]] && [[ "$payload" =~ ^/tmp/mng_lo_diag_20261005_[a-f0-9]{16}$ ]]
 [[ "$executable" = /usr/bin/libreoffice ]]
 out=$payload/$label
 mkdir -m 0700 "$out"
 workspace=$(mktemp -d -t phase1-word-pdf-XXXXXXXX)
 [[ "$workspace" = /tmp/phase1-word-pdf-* ]]
 finish_case() { code=$?; trap - EXIT; rm -rf -- "$workspace"; printf 'WORKSPACE_REMOVED=%s CASE_SCRIPT_EXIT=%s\n' "$([ ! -e "$workspace" ] && echo 1 || echo 0)" "$code"; exit "$code"; }
 trap finish_case EXIT
 cp "$payload/synthetic.docx" "$workspace/synthetic.docx"; chmod 0600 "$workspace/synthetic.docx"
 cd /opt/talent-assessment
 begin=$(date -u +%FT%TZ)
 set +e
 timeout --signal=TERM --kill-after=5s 90s "$executable" "-env:UserInstallation=file://$workspace/profile" --headless --convert-to pdf --outdir "$workspace" "$workspace/synthetic.docx" > "$out/stdout" 2> "$out/stderr"
 loexit=$?
 set -e
 valid=0; bytes=0
 if [ -f "$workspace/synthetic.pdf" ]; then
  bytes=$(stat -c %s "$workspace/synthetic.pdf")
  if [ "$bytes" -ge 1024 ] && [ "$(head -c 5 "$workspace/synthetic.pdf")" = %PDF- ]; then valid=1; fi
 fi
 printf 'CASE=%s UID=%s LO_EXIT=%s PDF_VALID=%s PDF_BYTES=%s BEGIN=%s END=%s\n' "$label" "$(id -u)" "$loexit" "$valid" "$bytes" "$begin" "$(date -u +%FT%TZ)"
 for stream in stdout stderr; do
  printf 'CASE=%s STREAM=%s BYTES=%s SHA=%s\n' "$label" "$stream" "$(wc -c < "$out/$stream")" "$(sha256sum "$out/$stream" | cut -d' ' -f1)"
 done
 categories=(permission_denied user_installation profile_lock javaldx java_warning font_error crash_signal namespace_denied source_load_error)
 patterns=('permission denied|access denied|operation not permitted' 'user installation.*(could not|failed|error)|cannot create.*(profile|config)' 'lock.*(failed|denied)|another instance' 'javaldx failed|failed to launch javaldx' 'could not find.*java|java.*(warning|runtime)' 'font.*(error|failed|denied)' 'segmentation fault|aborted|signal [0-9]+|terminate called' 'apparmor="DENIED"|userns|namespace.*(failed|denied)' 'source file could not be loaded|Error:.*(load|open)')
 for i in "${!categories[@]}"; do
  n=$(grep -Eic "${patterns[$i]}" "$out/stderr" || true)
  printf 'CASE=%s CATEGORY=%s COUNT=%s\n' "$label" "${categories[$i]}" "$n"
 done
 chmod 0600 "$out"/*
 exit 0
fi
payload=${1:?}; expected_sha=${2:?}
[[ "$payload" =~ ^/tmp/mng_lo_diag_20261005_[a-f0-9]{16}$ ]] && [[ "$expected_sha" =~ ^[a-f0-9]{64}$ ]]
[ "$(hostname)" = vm-ubuntu-go-dev ] && [ "$(id -u)" = 0 ]
[ "$(stat -c '%U:%G:%a' "$payload")" = liming:liming:700 ]
[ "$(sha256sum "$payload/synthetic.docx" | cut -d' ' -f1)" = "$expected_sha" ]
parent=/opt/talent-assessment/backups/mng_phase1_20261003_112853_fd24d4ee35a9
evidence=$(mktemp -d "$parent/lo_diag_20261005_XXXXXXXX")
printf '%s\n' "$payload" mng-synthetic-only-lo-probe-v1 > "$evidence/ownership"
unit=mng-lo-diag-${payload##*_}
finish() {
 code=$?; trap - EXIT
 if systemctl show "$unit.service" -p LoadState --value 2>/dev/null | grep -qx loaded; then systemctl stop "$unit.service" > "$evidence/unit-stop.out" 2> "$evidence/unit-stop.err" || true; fi
 if [ -d "$payload" ] && [ "$(cat "$payload/ownership")" = mng-synthetic-only-lo-probe-v1 ]; then
  cp -a "$payload" "$evidence/payload-receipt"
  find "$evidence" -type d -exec chmod 0700 {} +
  find "$evidence" -type f -exec chmod 0600 {} +
  rm -rf -- "$payload"
 fi
 printf 'PROBE_EXIT=%s PAYLOAD_REMAINING=%s EVIDENCE=%s\n' "$code" "$([ -e "$payload" ] && echo 1 || echo 0)" "$evidence"
 exit "$code"
}
trap finish EXIT
pid=$(systemctl show talent-assessment -p MainPID --value)
[ "$(systemctl show talent-assessment -p User --value)" = liming ]
[ "$(systemctl show talent-assessment -p WorkingDirectory --value)" = /opt/talent-assessment ]
[ "$(sha256sum /proc/"$pid"/exe | cut -d' ' -f1)" = 8fb264e669b9667c9672905dadf43fbd4654a669b285643db938cb0d7b33ae93 ]
safe_env=(); child_env=(); executable=/usr/bin/libreoffice
while IFS= read -r -d '' item; do
 case "${item%%=*}" in
  HOME|PATH|TMPDIR|TMP|TEMP|XDG_RUNTIME_DIR|XDG_CONFIG_HOME|XDG_CACHE_HOME|LANG|LC_ALL)
   safe_env+=("--setenv=$item"); child_env+=("$item");;
  LIBREOFFICE_PATH) executable=${item#*=};;
  REPORT_EFFECTIVE_ENV|MNG_TEST_REPORT_ENV) export "$item";;
 esac
done < /proc/"$pid"/environ
unset item
[ "${REPORT_EFFECTIVE_ENV:-}" = staging ] && [ "${MNG_TEST_REPORT_ENV:-}" = staging ]
[ "$executable" = /usr/bin/libreoffice ]
props=(User Group WorkingDirectory ProtectHome ProtectSystem PrivateTmp PrivateDevices NoNewPrivileges RestrictNamespaces UMask LimitNOFILE MemoryMax TasksMax RootDirectory AppArmorProfile SELinuxContext ReadWritePaths ReadOnlyPaths InaccessiblePaths SystemCallFilter SystemCallArchitectures RestrictAddressFamilies CapabilityBoundingSet AmbientCapabilities)
show=(); for p in "${props[@]}"; do show+=("-p" "$p"); done
systemctl show talent-assessment "${show[@]}" -p MainPID -p ActiveEnterTimestamp > "$evidence/service-props"
cat "$evidence/service-props"
# Refuse unsupported confinement rather than claim a matched environment.
for p in RootDirectory AppArmorProfile SELinuxContext ReadWritePaths ReadOnlyPaths InaccessiblePaths SystemCallArchitectures AmbientCapabilities; do [ -z "$(systemctl show talent-assessment -p "$p" --value)" ]; done
[ "$(systemctl show talent-assessment -p SystemCallFilter --value)" = '~' ]
[ "$(systemctl show talent-assessment -p RestrictAddressFamilies --value)" = '~' ]
sandbox=()
for p in ProtectHome ProtectSystem PrivateTmp PrivateDevices NoNewPrivileges RestrictNamespaces UMask LimitNOFILE MemoryMax TasksMax CapabilityBoundingSet; do sandbox+=("--property=$p=$(systemctl show talent-assessment -p "$p" --value)"); done
for spec in 'service:talent-assessment:2026-10-04 09:12:20 UTC:2026-10-04 09:12:24 UTC' 'kernel:kernel:2026-10-04 09:12:20 UTC:2026-10-04 09:12:24 UTC'; do
 IFS=: read -r kind target begin end <<< "$spec"
 # UTC timestamps contain colons: use fixed times, not split timestamp text.
 if [ "$kind" = service ]; then jargs=(-u talent-assessment); else jargs=(-k); fi
 journalctl "${jargs[@]}" --since '2026-10-04 09:12:20 UTC' --until '2026-10-04 09:12:24 UTC' --no-pager -o cat > "$evidence/journal-$kind.private"
 for pattern in 'lo_command_exec_exit' 'apparmor="DENIED"' 'segfault|segmentation fault' 'permission denied' 'javaldx|java.*failed' 'oom-kill|out of memory'; do
  n=$(grep -Eic "$pattern" "$evidence/journal-$kind.private" || true)
  printf 'HISTORICAL_WINDOW=%s CATEGORY=%s COUNT=%s\n' "$kind" "$pattern" "$n"
 done
done
env -i HOME=/root PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/snap/bin LANG=C.UTF-8 bash "$payload/probe.sh" --case root "$payload" "$executable" | tee "$evidence/root-summary"
sudo -u liming -H env -i HOME=/home/liming PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/snap/bin LANG=C.UTF-8 bash "$payload/probe.sh" --case liming "$payload" "$executable" | tee "$evidence/liming-summary"
systemd-run --unit "$unit" --wait --pipe --collect --service-type=exec --uid=liming --gid=liming --working-directory=/opt/talent-assessment --property=RuntimeMaxSec=100s "${sandbox[@]}" "${safe_env[@]}" /usr/bin/env -i "${child_env[@]}" /usr/bin/bash "$payload/probe.sh" --case matched "$payload" "$executable" > "$evidence/matched-summary" 2> "$evidence/transient.stderr"
cat "$evidence/matched-summary"
printf 'MATCHED_TRANSIENT_STDERR_BYTES=%s SHA=%s\n' "$(wc -c < "$evidence/transient.stderr")" "$(sha256sum "$evidence/transient.stderr" | cut -d' ' -f1)"
printf 'SERVICE_PID_BEFORE=%s AFTER=%s MATCHED_SAFE_ENV_FROM_PROC=1\n' "$pid" "$(systemctl show talent-assessment -p MainPID --value)"
for s in talent-assessment nginx mysql; do printf 'SERVICE=%s ' "$s"; systemctl is-active "$s"; done
curl --fail --silent --max-time 10 http://127.0.0.1:8092/health
printf '\n'