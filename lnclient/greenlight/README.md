# Greenlight backend for Alby Hub

A Core Lightning node on Blockstream's infrastructure. A VLS signer on your
Alby Hub. The two communicate over mTLS — Blockstream runs the node, your hub
holds the keys.

## Why Greenlight for Alby Hub?

Greenlight is the only backend where the node lives in the cloud but the keys
stay local. Every other backend bundles both together:

| | Node | Keys |
|---|---|---|
| LND | Self-hosted | Self-hosted |
| LDK | Embedded | Embedded |
| CLN | Self-hosted | Self-hosted |
| Bark | Cloud | Cloud (custodial) |
| **Greenlight** | **Cloud** | **Local** |

You get CLN's full feature set (bolt12 offers, keysend, plugins) without disk
management, uptime worry, or channel operations. Channels are provisioned
automatically by Blockstream's LSP. If the platform disappears tomorrow, the
same 32-byte seed that runs your signer boots a standalone CLN node — no
vendor lock-in.

## Setup (two paths)

**Product path** — the hub wizard registers a new node from a 12-word mnemonic,
extracts credentials, and starts the signer under supervision.

**Preconfigured path** — point the hub at existing credentials and a node URI:

```sh
LN_BACKEND_TYPE=GREENLIGHT
GREENLIGHT_CREDS_PATH=/path/to/greenlight-creds
GREENLIGHT_NODE_URI=gl1<node_id>.nodes.gl.blckstrm.com:443
```

## Status

30 of 35 LNClient methods implemented via standard CLN gRPC. 4 are honest
`ErrUnsupported` (hold invoices, signmessage — architecturally blocked). 1
stub. Tested live on Blockstream testnet: invoices created, addresses
generated, balances reported, backup/restore proven.

## What we added on top

Greenlight provides the signer and node. We added what a hub needs to run
it as a service:

- **Signer supervisor** — 15s respawn, SIGTERM→KILL, Signal(0) liveness,
  health surfaced in `GetNodeStatus`
- **Health watchdog** — periodic Getinfo, pump-stall detection, degraded boot
  (a frozen node doesn't lock you out of your hub)
- **WaitAnyInvoice pump** — persisted pay index, busy-loop protection
- **Provisioning** — mnemonic→seed→register→extract creds→launch signer,
  fully automated

## Backups and exit strategy

The signer writes `backup.json` when channels exist. The hub `.bkp` includes
it. If Blockstream's cloud goes away: `hsm_secret` → CLN node identity,
`backup.json` → `glcli signer convert-backup --format cln` → SCB →
`lightning-cli recoverchannel`. Everything needed to self-host is already on
your disk.

[ARCHITECTURE.md](ARCHITECTURE.md) covers the integration design.
[docs/RUNBOOK.md](docs/RUNBOOK.md) covers operations and troubleshooting.
