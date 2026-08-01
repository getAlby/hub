# Deploy

## Topology

Three processes on one machine:

| Process | Port | Role |
|---|---|---|
| Hub (Go) | 8080 | wallet, controller, UI |
| routstrd (Bun) | 8008 | router, key registry, Cashu wallet client |
| cocod (Cashu daemon) | socket | mint operations |

The Hub supervises routstrd and cocod: it starts them, health-checks them, restarts them, and runs the auto top-up loop. **Port 8081 is reserved for a buzz relay and must not be used.**

## Prerequisites

- Go 1.25, Node 20+ with yarn, Bun
- A reachable Cashu mint (this deployment uses `https://mint.cubabitcoin.org`, Nutshell 0.20)
- A Lightning backend for the Hub (this deployment uses Bark; the Hub also supports LDK, LND, Phoenixd, and Cashu)

## Install

```sh
# 1. Build the Hub
cd frontend && yarn install && yarn build:http
cd .. && go build -o hub cmd/http/main.go

# 2. Install the daemons
bun add -g routstrd        # the router daemon (Bun global package)
# install cocod (Cashu daemon) per its own instructions, data under ~/.cocod
# IMPORTANT: pin daemon versions and re-apply the documented dist patches
# (see daemon-patches.md) so the artifact is reproducible.

# 3. Configure routstrd (~/.routstrd/config.json)
#    keys: port, provider, cocodPath, mode ("apikeys"), nsec, nwc
#    mode "apikeys" = client keys minted by this daemon

# 4. Run the Hub
./hub
```

On first run, complete setup and unlock. The Hub starts the daemons itself; there is no separate daemon management.

## Firewall

- **8080 open** to whatever needs the Hub UI and the external OpenAI endpoint (`/routstr/v1`)
- **8008 closed** at all costs. The daemon binds `*:8008` and its admin endpoints are unauthenticated. Anyone who can reach 8008 can list, create, and delete clients, and spend a funded wallet (the completions path has no key check; the spend gate is the Cashu balance). Block 8008 in the firewall **before** the wallet holds sats, and verify the block: `curl -m 3 http://<public-ip>:8008/health` must time out from outside the host, and the Hub's public proxy must NOT forward admin paths to the daemon (`curl -s -X POST http://<public-ip>:8080/routstr/stop` must return the SPA index page, never a daemon response — the proxy only exposes `/routstr/v1/*`).

## Operations

- **Unlock after restart.** The Hub boots locked: every request 401s until the unlock password is supplied (`POST /api/start`). Start, unlock, then let the daemons come up (see [troubleshooting](troubleshooting.md)).
- **Backups.** See [backup-restore.md](backup-restore.md).
- **Mint dependency.** Funding and auto top-up need the mint reachable. A mint outage blocks Top Up, wizard funding, and auto-refill (diagnosis in [troubleshooting.md](troubleshooting.md)).
