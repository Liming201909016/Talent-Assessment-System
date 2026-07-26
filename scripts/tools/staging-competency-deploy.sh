#!/usr/bin/env bash
set -euo pipefail

expected_backend=${1:?expected backend sha256 required}
expected_index=${2:?expected frontend index sha256 required}
app_dir=/opt/talent-assessment
stamp=$(date +%Y%m%d_%H%M%S)

actual_backend=$(sha256sum /tmp/server-linux | cut -d' ' -f1)
actual_index=$(sha256sum /tmp/dist-new/index.html | cut -d' ' -f1)
[ "$actual_backend" = "$expected_backend" ] || { echo "backend checksum mismatch"; exit 1; }
[ "$actual_index" = "$expected_index" ] || { echo "frontend checksum mismatch"; exit 1; }

echo "UPLOAD_CHECK_OK"

sudo -n cp -a "$app_dir/server" "$app_dir/server.bak.$stamp"
sudo -n cp -a "$app_dir/dist" "$app_dir/dist.bak.$stamp"
echo "APP_BACKUP_OK stamp=$stamp"

sudo -n systemctl stop talent-assessment
sudo -n install -o root -g root -m 0755 /tmp/server-linux "$app_dir/server"
sudo -n rm -rf "$app_dir/dist.new"
sudo -n cp -a /tmp/dist-new "$app_dir/dist.new"
sudo -n chmod -R 755 "$app_dir/dist.new"
sudo -n chown -R root:root "$app_dir/dist.new"
sudo -n rm -rf "$app_dir/dist"
sudo -n mv "$app_dir/dist.new" "$app_dir/dist"
sudo -n systemctl start talent-assessment
sudo -n systemctl is-active --quiet talent-assessment
sudo -n nginx -t
sudo -n systemctl reload nginx

for attempt in $(seq 1 20); do
  if curl -fsS http://127.0.0.1:8092/health >/tmp/competency-health.json; then
    break
  fi
  if [ "$attempt" -eq 20 ]; then
    echo "backend health check failed"
    sudo -n systemctl status talent-assessment --no-pager | tail -30
    exit 1
  fi
done

deployed_backend=$(sha256sum "$app_dir/server" | cut -d' ' -f1)
deployed_index=$(sha256sum "$app_dir/dist/index.html" | cut -d' ' -f1)
[ "$deployed_backend" = "$expected_backend" ]
[ "$deployed_index" = "$expected_index" ]

echo "DEPLOY_OK"
echo "backend_sha256=$deployed_backend"
echo "frontend_index_sha256=$deployed_index"
echo "service=$(systemctl is-active talent-assessment)"
echo "nginx=$(systemctl is-active nginx)"
echo "health=$(cat /tmp/competency-health.json)"
