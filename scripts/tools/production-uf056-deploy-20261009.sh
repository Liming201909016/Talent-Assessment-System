#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT=/opt/talent-assessment
PAYLOAD=${1:?payload directory required}
STAMP=${2:?release stamp required}
TEMPLATE="$ROOT/configs/export-templates/competency-phase1-report-v2.docx"
CANDIDATE="$PAYLOAD/competency-phase1-report-v2.docx"
BACKUP="$ROOT/backups/uf056-template-$STAMP"
EXPECTED_HOST=iZ0yosjdcen2p4Z
OLD_SHA=a814c36e3759c8ff2cf5c17148e0939b477f17d4d3f430bb6a22f530022f9f5a
NEW_SHA=52e0020c5a6f39535d020bf41b18d9412c096ab43ce6608964b0101b955f30b9
PID_BEFORE=$(systemctl show talent-assessment -p MainPID --value)
RESTARTS_BEFORE=$(systemctl show talent-assessment -p NRestarts --value)

[ "$(hostname)" = "$EXPECTED_HOST" ]
[ "$RESTARTS_BEFORE" = 0 ]
[ "$(sha256sum "$TEMPLATE" | cut -d' ' -f1)" = "$OLD_SHA" ]
[ "$(sha256sum "$CANDIDATE" | cut -d' ' -f1)" = "$NEW_SHA" ]

python3 - "$TEMPLATE" "$CANDIDATE" <<'PY'
import hashlib
import sys
import zipfile

old, new = sys.argv[1:]
with zipfile.ZipFile(old) as baseline, zipfile.ZipFile(new) as candidate:
    baseline_names = set(baseline.namelist())
    candidate_names = set(candidate.namelist())
    if baseline_names != candidate_names:
        raise SystemExit('DOCX_PART_SET_MISMATCH')
    changed = [
        name for name in sorted(baseline_names)
        if hashlib.sha256(baseline.read(name)).digest() != hashlib.sha256(candidate.read(name)).digest()
    ]
    if changed != ['word/document.xml']:
        raise SystemExit('UNEXPECTED_CHANGED_PARTS:' + ','.join(changed))
    document = candidate.read('word/document.xml').decode('utf-8')
    if '>时长：</w:t>' in document or '>分钟</w:t>' in document:
        raise SystemExit('VISIBLE_DURATION_REMAINS')
    if 'w:val="result.userTime"' not in document or '<w:vanish' not in document:
        raise SystemExit('HIDDEN_CONTRACT_MISSING')
    if candidate.testzip() is not None:
        raise SystemExit('DOCX_ZIP_INVALID')
print('SEMANTIC_DELTA=word/document.xml-only')
PY

install -d -m 700 "$BACKUP"
cp -a "$TEMPLATE" "$BACKUP/competency-phase1-report-v2.docx"
[ "$(sha256sum "$BACKUP/competency-phase1-report-v2.docx" | cut -d' ' -f1)" = "$OLD_SHA" ]

rollback() {
  install -m 644 "$BACKUP/competency-phase1-report-v2.docx" "$TEMPLATE"
  printf 'PRODUCTION_ROLLBACK=completed\n'
}
trap rollback ERR
install -m 644 "$CANDIDATE" "$TEMPLATE"
[ "$(sha256sum "$TEMPLATE" | cut -d' ' -f1)" = "$NEW_SHA" ]
python3 - "$TEMPLATE" <<'PY'
import sys
import zipfile

with zipfile.ZipFile(sys.argv[1]) as archive:
    if archive.testzip() is not None:
        raise SystemExit('INSTALLED_DOCX_ZIP_INVALID')
    document = archive.read('word/document.xml').decode('utf-8')
    if '>时长：</w:t>' in document or '>分钟</w:t>' in document:
        raise SystemExit('INSTALLED_VISIBLE_DURATION_REMAINS')
    if 'w:val="result.userTime"' not in document or '<w:vanish' not in document:
        raise SystemExit('INSTALLED_HIDDEN_CONTRACT_MISSING')
PY
[ "$(systemctl is-active talent-assessment)" = active ]
[ "$(curl -fsS http://127.0.0.1:8092/health)" = '{"status":"ok"}' ]
[ "$(systemctl show talent-assessment -p MainPID --value)" = "$PID_BEFORE" ]
[ "$(systemctl show talent-assessment -p NRestarts --value)" = "$RESTARTS_BEFORE" ]
[ "$(curl -fsS -o /dev/null -w '%{http_code}' http://127.0.0.1:8090/)" = 200 ]
[ "$(curl -fsS -o /dev/null -w '%{http_code}' http://127.0.0.1:8090/prod-api/health)" = 200 ]
trap - ERR
rm -rf "$PAYLOAD"
printf '{"status":"passed","environment":"production","stage":"deploy_accept","oldSha256":"%s","newSha256":"%s","templateBytes":%s,"backup":"%s","mainPid":%s,"nRestarts":%s,"databaseAccess":false,"databaseWrites":false}\n' \
    "$OLD_SHA" "$NEW_SHA" "$(stat -c %s "$TEMPLATE")" "$BACKUP" "$PID_BEFORE" "$RESTARTS_BEFORE"
