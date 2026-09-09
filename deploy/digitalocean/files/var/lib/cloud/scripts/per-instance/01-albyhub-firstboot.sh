#!/bin/bash
# Runs once on the first boot of a droplet launched from the Alby Hub marketplace snapshot.
# Reads /etc/albyhub/hub.env (optionally pre-populated via cloud-init user-data) and
# rewrites the Caddy config to match: TLS + domain if DOMAIN is set, plain HTTP otherwise.
set -euo pipefail

ENV_FILE=/etc/albyhub/hub.env
CADDYFILE=/etc/caddy/Caddyfile

touch "$ENV_FILE"

DOMAIN=$(grep -s '^DOMAIN=' "$ENV_FILE" | cut -d= -f2- || true)

if [ -n "$DOMAIN" ]; then
  cat > "$CADDYFILE" <<EOF
$DOMAIN {
	reverse_proxy 127.0.0.1:8080
}
EOF
  grep -q '^BASE_URL=' "$ENV_FILE" || echo "BASE_URL=https://$DOMAIN" >> "$ENV_FILE"
else
  # No domain — expose Hub on :80 over plain HTTP for IP-based access.
  cat > "$CADDYFILE" <<EOF
:80 {
	reverse_proxy 127.0.0.1:8080
}
EOF
fi

systemctl restart caddy
systemctl restart albyhub
