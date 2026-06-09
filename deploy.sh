#!/usr/bin/env bash
set -e

echo "=== Deploying Alby Hub ==="

# Build with Nix
nix build --impure

# Get Nix store path
STORE_PATH=$(readlink -f result)
echo "Built: $STORE_PATH"

# Create data directory
mkdir -p ~/.local/share/albyhub

# Generate systemd service
sed "s|STORE_PATH_PLACEHOLDER|$STORE_PATH|g" albyhub.service > albyhub-generated.service

echo ""
echo "=== Install ==="
echo "1. Systemd:"
mkdir -p ~/.config/systemd/user/
cp albyhub-generated.service ~/.config/systemd/user/albyhub.service
systemctl --user daemon-reload
systemctl --user enable albyhub
systemctl --user restart albyhub
echo "   ✅ Service restarted"
echo ""
echo "2. Status:"
systemctl --user status albyhub --no-pager
echo ""
echo "3. Access:"
echo "   http://localhost:8087"
echo ""
echo "View logs: journalctl --user -u albyhub -f"
