#!/usr/bin/env bash
# Atomic hub deploy: build -> systemd restart -> unlock -> verify in ONE
# sequence. Prevents the deploy-kill trap (a killed hub left un-restarted
# takes the whole stack down: routstrd + cocod die with it).
# Usage: UNLOCK_PASSWORD=... ./deploy.sh   (password required, never defaulted)
set -euo pipefail
cd /root/hub
# Required: the Hub's unlock password (never defaulted in a public repo).
UNLOCK_PASSWORD="${UNLOCK_PASSWORD:?set UNLOCK_PASSWORD, e.g. UNLOCK_PASSWORD=... ./deploy.sh}"

# Pinned daemon versions — the supervision code and the dist patch are
# written against these; an unpinned `bun add -g` update can break the
# integration silently.
ROUTSTRD_VERSION="0.3.11"
COCOD_VERSION="0.0.24"

echo "== ensure pinned daemon versions =="
INSTALLED_ROUTSTRD=$(python3 -c "import json; print(json.load(open('/root/.bun/install/global/node_modules/routstrd/package.json')).get('version',''))" 2>/dev/null || echo "")
if [ "$INSTALLED_ROUTSTRD" != "$ROUTSTRD_VERSION" ]; then
  echo "installing routstrd@$ROUTSTRD_VERSION (had $INSTALLED_ROUTSTRD)"
  bun add -g routstrd@$ROUTSTRD_VERSION
fi
INSTALLED_COCOD=$(python3 -c "import json; print(json.load(open('/root/.bun/install/global/node_modules/@routstr/cocod/package.json')).get('version',''))" 2>/dev/null || echo "")
if [ "$INSTALLED_COCOD" != "$COCOD_VERSION" ]; then
  echo "installing cocod@$COCOD_VERSION (had $INSTALLED_COCOD)"
  bun add -g @routstr/cocod@$COCOD_VERSION
fi

echo "== apply routstrd dist patch (idempotent) =="
scripts/patch-routstrd-dist.sh

echo "== build frontend =="
(cd frontend && yarn build:http 2>&1 | tail -2)
echo "== build binary =="
go build -o hub cmd/http/main.go
echo "bundle: $(ls -t frontend/dist/assets/index-*.js | head -1)"

echo "== restart via systemd =="
systemctl restart alby-hub
sleep 6

echo "== unlock =="
# Build the JSON payload with python3 so quotes/backslashes in the password
# cannot produce malformed JSON.
TOKEN=$(printf '%s' "$UNLOCK_PASSWORD" \
  | python3 -c "import sys,json; print(json.dumps({'unlockPassword': sys.stdin.read()}))" \
  | curl -s -m 15 -X POST http://localhost:8080/api/start \
      -H "Content-Type: application/json" --data-binary @- \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))" 2>/dev/null)
[ -z "$TOKEN" ] && { echo "UNLOCK FAILED"; exit 1; }
echo "token: ${TOKEN:0:12}..."

echo "== wait for daemons (max 60s) =="
STATE=""
for i in $(seq 1 12); do
  STATE=$(curl -s -m 5 http://localhost:8080/api/routstrd/autorefill/status \
    -H "Authorization: Bearer $TOKEN" 2>/dev/null \
    | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['routstrdHealthy'], d['cocodHealthy'])" 2>/dev/null || echo "false false")
  [ "$STATE" = "True True" ] && break
  sleep 5
done
echo "daemons healthy: $STATE"
if [ "$STATE" != "True True" ]; then echo "DAEMONS NOT HEALTHY"; exit 1; fi

echo "== status =="
curl -s -m 10 http://localhost:8080/api/routstrd/autorefill/status \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
echo "DEPLOY OK"
