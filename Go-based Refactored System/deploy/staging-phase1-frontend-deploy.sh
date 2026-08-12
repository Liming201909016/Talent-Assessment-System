#!/usr/bin/env bash
set -euo pipefail

EXPECTED_ARCHIVE=${1:?expected archive sha256 required}
EXPECTED_INDEX=${2:?expected index sha256 required}
ARCHIVE=/tmp/phase1-report-fixed-text-dist.tar.gz
EXTRACT=/tmp/phase1-report-fixed-text-dist
APP=/opt/talent-assessment
STAMP=$(date +%Y%m%d_%H%M%S)

[ "$(sha256sum "$ARCHIVE" | cut -d' ' -f1)" = "$EXPECTED_ARCHIVE" ]
rm -rf "$EXTRACT"
mkdir -p "$EXTRACT"
tar -xzf "$ARCHIVE" -C "$EXTRACT"
[ "$(sha256sum "$EXTRACT/index.html" | cut -d' ' -f1)" = "$EXPECTED_INDEX" ]

sudo cp -a "$APP/dist" "$APP/dist.bak.phase1_fixed_text.$STAMP"
sudo rm -rf "$APP/dist.new"
sudo cp -a "$EXTRACT" "$APP/dist.new"
sudo chmod -R 755 "$APP/dist.new"
sudo chown -R root:root "$APP/dist.new"
sudo rm -rf "$APP/dist"
sudo mv "$APP/dist.new" "$APP/dist"
sudo nginx -t
sudo systemctl reload nginx

[ "$(sha256sum "$APP/dist/index.html" | cut -d' ' -f1)" = "$EXPECTED_INDEX" ]
echo FRONTEND_DEPLOY_PASS
echo "backup=$APP/dist.bak.phase1_fixed_text.$STAMP"
echo "archive_sha256=$EXPECTED_ARCHIVE"
echo "index_sha256=$EXPECTED_INDEX"
