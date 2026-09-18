#!/bin/bash
# DigitalOcean Marketplace image validator (local dry-run). The official img_check.sh
# is shipped via https://github.com/digitalocean/marketplace-partners and must pass
# before submission. This local version covers the Alby Hub-specific invariants.
set -euo pipefail

fail() { echo "FAIL: $*" >&2; exit 1; }

command -v docker >/dev/null || fail "docker not installed"
docker image inspect ghcr.io/getalby/hub:latest >/dev/null || fail "alby hub image not pre-pulled"
command -v caddy >/dev/null || fail "caddy not installed"

[[ -f /etc/systemd/system/albyhub.service ]] || fail "albyhub.service missing"
systemctl is-enabled albyhub.service >/dev/null || fail "albyhub.service not enabled"
systemctl is-enabled caddy >/dev/null || fail "caddy not enabled"

[[ -f /etc/caddy/Caddyfile ]] || fail "Caddyfile missing"
[[ -f /var/lib/cloud/scripts/per-instance/01-albyhub-firstboot.sh ]] || fail "firstboot script missing"
[[ -f /etc/update-motd.d/99-albyhub ]] || fail "motd missing"
[[ -d /opt/albyhub/data ]] || fail "/opt/albyhub/data not created"

echo "OK: image looks valid for marketplace submission"
