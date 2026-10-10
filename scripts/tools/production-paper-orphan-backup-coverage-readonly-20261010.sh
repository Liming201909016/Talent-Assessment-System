#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT=/opt/talent-assessment
EXPECTED_HOST=iZ0yosjdcen2p4Z
BACKUP=$ROOT/backups/release_00401_20260727_153001/element.sql.gz
CLIENT=/run/talent-assessment-paper-orphan-backup-readonly.cnf
cleanup_status=not_started
failed_stage=bootstrap

cleanup() {
  if rm -f "$CLIENT"; then cleanup_status=completed; else cleanup_status=failed; fi
}
trap cleanup EXIT HUP INT TERM

fail() {
  local message=${1:-unknown_error}
  cleanup
  trap - EXIT HUP INT TERM
  python3 - "$failed_stage" "$message" "$cleanup_status" <<'PY'
import json, sys
print(json.dumps({"status":"failed","failedStage":sys.argv[1],"error":sys.argv[2],"createdResources":["ephemeral_mysql_client_config"],"cleanup":sys.argv[3],"databaseWrites":0},ensure_ascii=True,separators=(",",":")))
PY
  exit 1
}

[ "$(hostname)" = "$EXPECTED_HOST" ] || fail wrong_host
[ "$(id -u)" = 0 ] || fail root_required
[ -r "$BACKUP" ] || fail backup_unreadable

failed_stage=config
python3 - "$CLIENT" <<'PY' || fail db_config_unavailable
from pathlib import Path
import re, sys
client = Path(sys.argv[1])
paths = [Path('/opt/talent-assessment/configs/application.yml'), Path('/opt/talent-assessment/configs/application-production.yml')]
text = '\n'.join(p.read_text(encoding='utf-8-sig').replace('\r', '') for p in paths if p.exists())
dsn = None
for line in text.splitlines():
  if line.strip().startswith('dsn:'):
    dsn = line.split(':',1)[1].strip().strip('"\'')
match = re.fullmatch(r'([^:]+):([^@]*)@tcp\(([^:)]+):(\d+)\)/([^?]+)(?:\?.*)?', dsn or '')
if not match or match.group(5) != 'element': raise SystemExit('DB_DSN_UNAVAILABLE')
user,password,host,port,_ = match.groups()
client.write_text(f'[client]\nuser={user}\npassword={password}\nhost={host}\nport={port}\n',encoding='utf-8')
client.chmod(0o600)
PY

failed_stage=backup_scan
result=$(python3 - "$CLIENT" "$BACKUP" <<'PY'
import gzip, hashlib, json, subprocess, sys
client, backup = sys.argv[1:]
base=['mysql','--defaults-extra-file='+client,'--batch','--skip-column-names','--raw','element']
query="SET SESSION TRANSACTION READ ONLY; SELECT DISTINCT p.exam_id FROM el_paper p WHERE NOT EXISTS (SELECT 1 FROM el_exam e WHERE e.id=p.exam_id) ORDER BY p.exam_id"
completed=subprocess.run(base+['-e',query],check=True,capture_output=True,text=True)
ids=[line for line in completed.stdout.splitlines() if line]
insert_chunks=[]
with gzip.open(backup,'rt',encoding='utf-8',errors='replace') as handle:
    for line in handle:
        if line.startswith('INSERT INTO `el_exam` '):
            insert_chunks.append(line)
exam_insert=''.join(insert_chunks)
coverage=[]
for value in ids:
    digest=hashlib.sha256(value.encode()).hexdigest()
    coverage.append({'examIdSha256':digest,'parentFoundInBackup':("'"+value+"'") in exam_insert})
print(json.dumps({'status':'passed','failedStage':None,'error':None,'readOnly':True,'databaseWrites':0,'backupRelativePath':'release_00401_20260727_153001/element.sql.gz','backupBytes':__import__('os').path.getsize(backup),'missingExamIds':len(ids),'parentsFound':sum(1 for x in coverage if x['parentFoundInBackup']),'coverage':coverage,'createdResources':['ephemeral_mysql_client_config'],'cleanup':'pending'},ensure_ascii=True,separators=(',',':')))
PY
) || fail backup_scan_failed

failed_stage=cleanup
cleanup
trap - EXIT HUP INT TERM
[ "$cleanup_status" = completed ] || fail cleanup_failed
python3 - "$result" <<'PY'
import json,sys
payload=json.loads(sys.argv[1]); payload['cleanup']='completed'
print(json.dumps(payload,ensure_ascii=True,separators=(',',':')))
PY
