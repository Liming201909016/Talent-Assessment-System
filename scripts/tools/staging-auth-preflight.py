#!/usr/bin/env python3
import os
import re
import subprocess
from pathlib import Path

app_dir = Path('/opt/talent-assessment')
env_line = subprocess.run(
    ['systemctl', 'show', 'talent-assessment', '--property=Environment', '--value'],
    check=True,
    capture_output=True,
    text=True,
).stdout.strip()
match = re.search(r'(?:^|\s)APP_ENV=([^\s]+)', env_line)
app_env = match.group(1) if match else 'local'

files = [app_dir / 'configs' / 'application.yml', app_dir / 'configs' / f'application-{app_env}.yml']
merged = '\n'.join(path.read_text(encoding='utf-8') for path in files if path.exists())

def scalar(section, key, default=''):
    section_match = re.search(rf'(?ms)^{re.escape(section)}:\s*\n(.*?)(?=^[A-Za-z][\w-]*:\s*$|\Z)', merged)
    if not section_match:
        return default
    key_match = re.search(rf'(?m)^\s+{re.escape(key)}:\s*(.+?)\s*$', section_match.group(1))
    if not key_match:
        return default
    return key_match.group(1).strip().strip('"\'')

secret = scalar('jwt', 'secret')
print(f'APP_ENV={app_env}')
print(f'CONFIG_FILES={",".join(path.name for path in files if path.exists())}')
print(f'REDIS_ADDR={scalar("redis", "addr")}')
print(f'REDIS_DB={scalar("redis", "db", "0")}')
print(f'JWT_SECRET_PRESENT={"yes" if secret else "no"}')
print(f'JWT_SECRET_LENGTH={len(secret)}')
