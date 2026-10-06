#!/usr/bin/env bash
# probe.sh BASE_URL [ATTEMPTS]: polls /healthz and /readyz every two seconds until both
# return 200; exits 1 after ATTEMPTS (default 30) failures.
set -euo pipefail
base="${1:?usage: probe.sh BASE_URL [ATTEMPTS]}"
attempts="${2:-30}"
for _ in $(seq 1 "$attempts"); do
  sleep 2
  if curl -fsS --max-time 5 "$base/healthz" >/dev/null && curl -fsS --max-time 5 "$base/readyz" >/dev/null; then
    echo "healthy: $base"
    exit 0
  fi
done
echo "not ready after $attempts attempts: $base" >&2
curl -sS --max-time 5 "$base/readyz" >&2 || true
echo >&2
exit 1
