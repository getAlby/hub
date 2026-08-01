#!/usr/bin/env bash
# Ensure the routstrd dist patch is applied (idempotent).
#
# The patched daemon keeps the model catalog fresh: 30-min cache TTL +
# a background warm-refresh loop. A routstrd update wipes the patch
# (stock default is a 210-min TTL with no warm loop), which makes the
# UI/models stale. This script re-applies it or fails the deploy loudly
# instead of shipping an unpatched daemon.
#
# Usage: scripts/patch-routstrd-dist.sh
set -euo pipefail

BUNDLE="/root/.bun/install/global/node_modules/routstrd/dist/daemon/index.js"
MARKER="cacheTTL: 30 * 60 * 1000"

if [ ! -f "$BUNDLE" ]; then
  echo "ERROR: daemon bundle not found at $BUNDLE" >&2
  exit 1
fi

if grep -qF "$MARKER" "$BUNDLE"; then
  echo "routstrd dist patch already applied"
  exit 0
fi

echo "routstrd dist patch missing — applying..."
python3 - "$BUNDLE" <<'EOF'
import sys

path = sys.argv[1]
src = open(path, encoding="utf-8", errors="replace").read()

# 1. Cache TTL: 210 min stock default -> 30 min.
OLD_TTL = "cacheTTL: 210 * 60 * 1000"
NEW_TTL = "cacheTTL: 30 * 60 * 1000"
if OLD_TTL in src:
    src = src.replace(OLD_TTL, NEW_TTL, 1)
elif "cacheTTL" in src and NEW_TTL not in src:
    # The default value may differ across versions; fail loudly rather than
    # guessing and shipping a stale catalog.
    print("ERROR: stock cacheTTL default not found (looked for '" + OLD_TTL + "')", file=sys.stderr)
    sys.exit(1)

# 2. Warm loop: refresh the catalog in the background every 30 min.
ANCHOR = "const { ensureProvidersBootstrapped, getRoutstr21Models, getAllModels, getModelProviders, refreshProvidersAndModels } = createModelService(modelManager, store);"
WARM = ANCHOR + """
  // HOTFIX: background warm-refresh — keep the model catalog fresh so the
  // UI, /v1/models consumers and routing decisions never see data older
  // than ~30 min. One failed cycle must not crash the daemon.
  setInterval(() => {
    refreshProvidersAndModels().catch((error53) => {
      logger5.error(`[models] background refresh failed: ${toErrorMessage(error53)}`);
    });
  }, 30 * 60 * 1000).unref();"""
if ANCHOR in src:
    if "background warm-refresh" not in src:
        src = src.replace(ANCHOR, WARM, 1)
else:
    print("ERROR: model service anchor not found; cannot insert warm loop", file=sys.stderr)
    sys.exit(1)

open(path, "w", encoding="utf-8").write(src)
print("routstrd dist patch applied")
EOF

# Verify the patch took.
if grep -qF "$MARKER" "$BUNDLE" && grep -qF "background warm-refresh" "$BUNDLE"; then
  echo "routstrd dist patch verified"
else
  echo "ERROR: patch verification failed after apply" >&2
  exit 1
fi
