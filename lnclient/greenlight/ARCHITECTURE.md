# Greenlight backend — architecture

## Why `glcli` (not the GL SDK directly)

Go cannot call Rust. Greenlight ships two integration surfaces:

1. **The `gl-client` Rust crate** — for Rust and Python consumers. Embeds the signer
   in-process. Exposes `Scheduler`, `Signer`, `Node`, `Device`, and backup primitives
   (`SignerBackupConfig`, `SignerBackupSnapshot::to_cln_backup()`).

2. **The `glcli` CLI** — the cross-language bridge for every other language. Every
   subcommand maps directly to a `gl-client` API call. The Getting Started docs
   show a CLI-equivalent flow: seed → Signer::new → Scheduler::register →
   Device::from_bytes. Our Go code calls `glcli` as a subprocess to execute the
   same sequence.

We integrate through `glcli`. This is not a workaround — it is the documented
integration path for non-Rust/non-Python consumers. Every other language binding
(Android via `gl-sdk-android`, iOS via `gl-sdk-swift`, Node via `gl-sdk-napi`)
wraps the same underlying `gl-client` — `glcli` is the same thing, just through
the CLI boundary.

## Architecture map: GL primitives → Go layer

```
Greenlight platform                     Alby Hub (Go)
──────────────────                     ─────────────
Scheduler::register                    glcli scheduler register
Scheduler::recover                     glcli scheduler recover
Scheduler::schedule (wake)             glcli scheduler schedule
      │                                       │
      ▼                                       ▼
Device::from_bytes(creds)      ───►    extract_creds.py (Go-embedded)
      │                                       │
      ▼                                       ▼
  ca.pem + client.pem                   Config.CredsPath
  + client-key.pem + rune                     │
      │                                       ▼
      ▼                               loadTLSCredentials → mTLS
Signer::new(seed, net, creds)  ───►    GreenlightSignerService
  (Rust in-process)                         │
      │                              glcli signer run
      ▼                                   --backup-path (signer state)
  signer event loop                   supervisor (15s respawn,
      │                               SIGTERM→KILL, health events)
      ▼
  [writes backup.json on first
   recoverable channel — absent
   on a fresh node until B1;
   the .bkp includes it when
   it exists]
      │
      ▼
Node (cln-grpc)               ───►    GreenlightService
  getinfo, pay, keysend,                   │
  waitanyinvoice, listfunds,        clngrpc.NodeClient (Go gRPC)
  etc.                                    │
                                    WaitAnyInvoice pump + streamIncoming
```

## What the Go layer adds (not in the GL SDK)

These are hosting concerns, not signing concerns — the GL SDK doesn't provide
them because a library can't self-monitor the way a process supervisor can.

| Addition | Why it exists |
|---|---|
| **Watchdog** (health.go) | Periodic Getinfo + pump-stall detection. A dead or wedged signer leaves the node unresponsive; the SDK's signer would just block. The watchdog detects the wedge, publishes events, and serves a cached verdict (21ms) so the API never hangs. |
| **Degraded boot** | If the node is unreachable at hub startup, the GL backend still initialises — the hub answers `isReady:false` instead of refusing to start. Without this, a frozen node locks the user out of their own hub. |
| **Health events** | `nwc_gl_node_health` + `nwc_gl_signer_health` structured events for an operations pipeline. The SDK emits logs; we emit events the hub's event bus can route to alerting. |
| **Signer supervision** | Respawn on crash, write-once seed, process-tracked liveness, SIGTERM→KILL escalation. The SDK runs the signer in-process (a thread); we supervise the CLI process (a separate process — the VLS ideal of process separation). |
| **Status surfacing** | Signer health merged into `GetNodeStatus.InternalNodeStatus` (so a dead signer isn't reported as a node outage). The SDK has no equivalent API. |

## Integration surface (LNClient compliance)

| Method | Status | GL primitive |
|---|---|---|
| GetInfo | ✅ | `cln.Getinfo` (gRPC) |
| GetBalances | ✅ | `cln.ListFunds` + `cln.ListPeerChannels` |
| SendPaymentSync | ✅ | `cln.Pay` |
| SendKeysend | ✅ | `cln.Keysend` (records real preimage from response) |
| MakeInvoice | ✅ | `cln.Offer` |
| WaitInvoice | ✅ | WaitAnyInvoice pump (30s timeout, restart-safe index) |
| NewAddress | ✅ | `cln.NewAddr` |
| Withdraw | ✅ | `cln.Withdraw` |
| OpenChannel | ⚠️ Deferred | `cln.FundChannel` + scheduler coordination |
| CloseChannel | ✅ | `cln.Close` (mutual; force works) |
| ListChannels | ✅ | `cln.ListPeerChannels` |
| ListPeers | ✅ | `cln.ListPeers` |
| GetNodeStatus | ✅ | Watchdog cached verdict + signer status |
| Backup / Restore | ✅ | .bkp includes SignerDataDir (seed + creds + backup.json) |
| ResetRouter | ⚠️ Stub | Documented; needs scheduler-level migration |
| HoldInvoice / SettleHoldInvoice | ❌ Blocked | Needs `cln.HtlcAccepted` hook — no equivalent in `cln-grpc` or gl-client. Documented in KNOWN-ISSUES.md |
| Subscribe events | ⚠️ Partial | `Subscribe*` streams panic the gl-plugin (CLN restricts them); incoming payments use `streamIncoming` instead |

## Credential lifecycle

```
Setup wizard (user enters 12-word mnemonic)
  │
  ▼
EnsureProvisioned (provision.go)
  ├── MnemonicToSeed32: bip39 to_seed("")[0:32] (byte-identical to gl-cli)
  ├── ensureSeedFile: write-once (never overwrites)
  ├── glcli scheduler register (Nobody certs → credentials.gfs)
  │     └── fallback: glcli scheduler recover (if node was previously registered)
  ├── glcli scheduler schedule (wake)
  └── extract_creds.py: credentials.gfs → {ca.pem, client.pem, client-key.pem, rune}
        │
        ▼
  Config.CredsPath = device-creds dir
  Config.NodeURI = parsed from extract output
        │
        ▼
GreenlightSignerService.Start(ctx, dataDir, network, glcliPath, eventPublisher)
  └── spawn: glcli -d dataDir -n net signer run --backup-path backup.json
        │
        ▼
GreenlightService (gRPC → node + watchdog + pump + streamIncoming)
```

## Failure-mode hierarchy

See RUNBOOK.md for the actionable procedures. The architecture enforces three
layers of detection:

1. **Signer** (local, supervised, 15s respawn) — publishes `nwc_gl_signer_health`
2. **Node** (hosted, watchdog-probed, 30s interval + pump-stall) — publishes `nwc_gl_node_health`
3. **Hub** (degraded boot) — never refuses to start if the node is unreachable; serves `isReady:false` in the cached verdict
