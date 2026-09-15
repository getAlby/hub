#!/bin/sh
#
# Rotate the secret behind the Alby Hub Caddy password gate.
#
# The current secret becomes GATE_SECRET_OLD and a fresh one becomes
# GATE_SECRET. The Caddyfile accepts both, so active sessions keep working and
# are upgraded to the new secret on their next request; sessions that were idle
# since the previous rotation have to log in again.
#
# Run it after someone leaves, after a device is lost, or on a schedule:
#
#   sudo ./rotate-gate-secret.sh
#
# To rotate monthly, drop it in /etc/cron.monthly/ (without the .sh suffix) or
# add a systemd timer.
#
# Set GATE_ENV_FILE to use a different path than /etc/caddy/gate.env.

set -eu

ENV_FILE="${GATE_ENV_FILE:-/etc/caddy/gate.env}"

if [ "$(id -u)" -ne 0 ]; then
  echo "❌ This script must be run as root (try: sudo $0)" >&2
  exit 1
fi

generate_secret() {
  if command -v openssl > /dev/null 2>&1; then
    openssl rand -base64 48 | tr -d '=+/\n' | cut -c1-43
  elif [ -r /dev/urandom ]; then
    LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 43
  else
    echo "❌ Need either openssl or /dev/urandom to generate a secret" >&2
    exit 1
  fi
}

CURRENT=""
if [ -f "$ENV_FILE" ]; then
  # matches GATE_SECRET= but not GATE_SECRET_OLD=
  CURRENT=$(sed -n 's/^GATE_SECRET=//p' "$ENV_FILE" | head -n 1)
fi

# On the very first run there is no previous secret to keep alive. Use a fresh
# random value rather than an empty one: an empty GATE_SECRET_OLD combined with
# a missing cookie would otherwise be a guessable value.
if [ -z "$CURRENT" ]; then
  CURRENT=$(generate_secret)
fi

NEW=$(generate_secret)

if [ -z "$NEW" ]; then
  echo "❌ Failed to generate a secret" >&2
  exit 1
fi

mkdir -p "$(dirname "$ENV_FILE")"

umask 077
TMP_FILE="$ENV_FILE.tmp.$$"
# shellcheck disable=SC2064
trap "rm -f '$TMP_FILE'" EXIT INT TERM

cat > "$TMP_FILE" << EOF
# Managed by rotate-gate-secret.sh - do not edit by hand.
# Last rotated: $(date -u '+%Y-%m-%dT%H:%M:%SZ')
GATE_SECRET=$NEW
GATE_SECRET_OLD=$CURRENT
EOF

chmod 600 "$TMP_FILE"
mv "$TMP_FILE" "$ENV_FILE"

echo "✅ Wrote a new gate secret to $ENV_FILE"

# systemd only reads EnvironmentFile when the service starts, so `systemctl
# reload caddy` would keep serving the old secret.
if command -v systemctl > /dev/null 2>&1 && systemctl is-active --quiet caddy; then
  systemctl restart caddy
  echo "✅ Restarted Caddy"
else
  echo "ℹ️  Caddy does not seem to be running under systemd."
  echo "   Restart it manually so it picks up the new secret."
fi
