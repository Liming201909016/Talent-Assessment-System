#!/usr/bin/env bash
set -euo pipefail
umask 077
ROOT=/opt/talent-assessment
SERVICE=talent-assessment
BACKUP=${1:?backup directory required}
EXPECTED_BACKUP=/opt/talent-assessment/backups/phase1_staging_sync_20261009211916
CLIENT=/run/phase1-staging-sync-rollback.cnf
cleanup(){ rm -f "$CLIENT"; }
trap cleanup EXIT HUP INT TERM
[ "$(hostname)" = iZ0yosjdcen2p4Z ]
[ "$(id -u)" -eq 0 ]
[ "$BACKUP" = "$EXPECTED_BACKUP" ]
[ -f "$BACKUP/element.sql.gz" ]
[ -f "$BACKUP/server" ]
[ -f "$BACKUP/dist.tar.gz" ]
[ -d "$BACKUP/export-templates" ]
gzip -t "$BACKUP/element.sql.gz"
python3 - "$CLIENT" <<'PY'
from pathlib import Path
import re,sys
out=Path(sys.argv[1]); text='\n'.join(p.read_text(encoding='utf-8-sig').replace('\r','') for p in (Path('/opt/talent-assessment/configs/application.yml'),Path('/opt/talent-assessment/configs/application-production.yml')) if p.exists()); dsn=''
for line in text.splitlines():
    if line.strip().startswith('dsn:'): dsn=line.split(':',1)[1].strip().strip('"\'')
m=re.fullmatch(r'([^:]+):([^@]*)@tcp\(([^:)]+):(\d+)\)/([^?]+)(?:\?.*)?',dsn)
if not m or m.group(5)!='element': raise SystemExit('DB_DSN_INVALID')
u,p,h,port,_=m.groups(); esc=lambda v:v.replace('\\','\\\\').replace('"','\\"')
out.write_text('[client]\nuser="%s"\npassword="%s"\nhost="%s"\nport=%s\n'%(esc(u),esc(p),esc(h),port),encoding='utf-8'); out.chmod(0o600)
PY
systemctl stop "$SERVICE"
cp -a "$BACKUP/server" "$ROOT/server"
rm -rf "$ROOT/dist"; mkdir -p "$ROOT/dist"; tar -xzf "$BACKUP/dist.tar.gz" -C "$ROOT/dist"
rm -rf "$ROOT/configs/export-templates"; cp -a "$BACKUP/export-templates" "$ROOT/configs/export-templates"
gunzip -c "$BACKUP/element.sql.gz" | mysql --defaults-extra-file="$CLIENT" element
chown -R root:root "$ROOT/server" "$ROOT/dist" "$ROOT/configs/export-templates"
chmod 755 "$ROOT/server"; find "$ROOT/dist" -type d -exec chmod 755 {} +; find "$ROOT/dist" -type f -exec chmod 644 {} +
systemctl start "$SERVICE"
for _ in $(seq 1 30); do [ "$(systemctl is-active "$SERVICE")" = active ] && curl -fsS http://127.0.0.1:8092/health >/dev/null && break; done
[ "$(sha256sum "$ROOT/server" | cut -d' ' -f1)" = f850575b1dfa6eac3f5b4533148baf715eabc8afa13ccd32c7655d0507a7d600 ]
[ "$(sha256sum "$ROOT/dist/index.html" | cut -d' ' -f1)" = abf93dd1fcd6ca6d94a1da393cc492594f6c6d00117152cd987b9c490bbfcbc4 ]
[ "$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8090/prod-api/health)" = 200 ]
printf 'PRODUCTION_ROLLBACK=PASS\nPID=%s\n' "$(systemctl show "$SERVICE" -p MainPID --value)"
