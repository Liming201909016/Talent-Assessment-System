#!/usr/bin/env bash
set -euo pipefail

echo "backend_sha256=$(sha256sum /opt/talent-assessment/server | cut -d' ' -f1)"
echo "frontend_index_sha256=$(sha256sum /opt/talent-assessment/dist/index.html | cut -d' ' -f1)"
echo "service=$(systemctl is-active talent-assessment)"
echo "nginx=$(systemctl is-active nginx)"
echo "health=$(curl -fsS http://127.0.0.1:8092/health)"
echo "server_backup=$(ls -1t /opt/talent-assessment/server.bak.* | head -1)"
echo "dist_backup=$(ls -1dt /opt/talent-assessment/dist.bak.* | head -1)"
