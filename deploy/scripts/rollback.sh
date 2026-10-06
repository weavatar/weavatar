#!/usr/bin/env bash
# rollback.sh API_DIR [SERVICE]: restores app.prev, cli.prev and VERSION.prev and reloads
# the systemd unit SERVICE (restarting it when it is not running) (defaults to $SERVICE_NAME). Database migrations are not reverted.
set -euo pipefail
dir="${1:?usage: rollback.sh API_DIR [SERVICE]}"
service="${2:-${SERVICE_NAME:-}}"
if [ -z "$service" ]; then
  echo "usage: rollback.sh API_DIR SERVICE (or set SERVICE_NAME)" >&2
  exit 2
fi
cd "$dir"
if [ ! -f app.prev ]; then
  echo "no previous binary to roll back to in $dir" >&2
  exit 1
fi
mv -f app.prev app
if [ -f cli.prev ]; then mv -f cli.prev cli; fi
if [ -f VERSION.prev ]; then mv -f VERSION.prev VERSION; fi
echo "rolled back to the previous binary: $(cat VERSION 2>/dev/null || echo unknown)"
sudo systemctl reload-or-restart "$service"
