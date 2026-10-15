#!/bin/sh
# Generates a new gate cookie secret and writes it to gate.env.
# Other variables in gate.env (e.g. BASIC_AUTH_USERNAME, BASIC_AUTH_PASSWORD_HASH) are kept.
# Restart Caddy afterwards to pick up the new secret.
set -eu
ENV_FILE=./gate.env

NEW=$(openssl rand -base64 32 | tr -d '=+/' | cut -c1-43)

umask 077
touch "$ENV_FILE"
{ grep -v '^GATE_SECRET=' "$ENV_FILE" || true; echo "GATE_SECRET=$NEW"; } > "$ENV_FILE.tmp"
mv "$ENV_FILE.tmp" "$ENV_FILE"
