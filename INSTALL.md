# Alby Hub Nix Installation

## Quick Start

```bash
./deploy.sh
```

This will:
- Build Alby Hub with Nix
- Install as user systemd service
- Run on port 8087
- Store data in `~/.local/share/albyhub`

## Access

Open browser to: `http://localhost:8087`

## Manage Service

```bash
# Check status
systemctl --user status albyhub

# View logs
journalctl --user -u albyhub -f

# Stop service
systemctl --user stop albyhub

# Restart service
systemctl --user restart albyhub
```

## Files

- `flake.nix` - Nix flake configuration
- `albyhub.service` - Systemd service template
- `deploy.sh` - Build and deploy script
