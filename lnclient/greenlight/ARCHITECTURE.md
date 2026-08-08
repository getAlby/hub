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

## Known issues


Status of each issue: **blocked** (upstream dependency), **mitigated**
(workaround shipped), or **fixed** (resolved + verified).

## 1. signmessage freezes the production signer — BLOCKED (upstream #739)

`signmessage` sends a signing request the production VLS signer cannot
answer; hsmd's queue wedges and the whole node freezes until Blockstream
restarts it. Reproduced twice (hosted testnet node A; regtest harness,
2026-08-08). Filed as [Blockstream/greenlight#739](https://github.com/Blockstream/greenlight/issues/739)
— unacknowledged.

- Mitigation: the backend stubs `SignMessage` (returns a clear error,
  never forwards). The health watchdog (health.go) detects a freeze and the
  hub boots degraded instead of locking the user out. RUNBOOK.md has the
  recovery path.
- Unblocks when: Blockstream fixes the signer (or the issue is formally
  acknowledged + a workaround documented).

## 2. Hold invoices — BLOCKED (stack-level)

Greenlight hosted nodes run no plugins and gl-client exposes no hold hooks
(verified by grep, 2026-08-08). `MakeHoldInvoice`/`Settle`/`Cancel` return a
clear error. Only feature gap vs LDK/CLN/LND. Do not ship a fake hold path.

## 3. Incoming keysend TLV records — FIXED (correction)

Earlier audits noted "TLVs dropped". Correction: `streamIncoming` maps the
incoming keysend `ExtraTLVs` into the transaction `Metadata` (`tlv_records`)
and they surface to NIP-47 clients via the freeform metadata blob. Verified
in source (streamincoming.go:87-96) + models.go comment.

## 4. Outgoing keysend preimage/hash recording — FIXED

CLN-family keysend derives its own preimage server-side; the caller's
preimage cannot be honored (no gRPC field). The backend now reports the
actual preimage + payment hash from the KeySend response and the
transactions service records those — the stored hash is the one that
actually settled (verified live on CLN + GL, 2026-08-08; upstream PR #2521).

## 5. ResetRouter — MITIGATED

The hosted node owns the network graph; there is no reset RPC. The stub
returns a clear error; the backup path tolerates it (warns + continues,
api/backup.go:63). Only the explicit reset-router API endpoint surfaces it.

## 6. Node hosted elsewhere — INHERENT (design)

The node's availability depends on Blockstream infrastructure (the flip
side of the custody split). The watchdog + runbook cover detection and
recovery; a frozen node cannot be restarted by the user.

## 7. channel-state notifications — MITIGATED

The GL node server leaves all six `cln.Node Subscribe*` stream RPCs
unimplemented (they panic the gl-plugin process). `nwc_channel_ready/closed`
notifications are not emitted; incoming payments are covered by the
restart-safe pump instead.

## 8. Signer seed at rest — PLAINTEXT (posture note, mainnet audit 2026-08-08)

The hub supervisor spawns `glcli signer run` directly, so the seed
(`hsm_secret`) sits plaintext in the 0700 data dir while the hub is
unlocked — the same posture as the hub's own LDK seed. An at-rest
encryption wrapper exists (`lnclient/greenlight/signer/encrypt-seed.py` +
`greenlight-signer-run.sh`, AES-256-GCM, for systemd deployments) but is
not wired into the supervisor. Wiring it is a follow-up; for a single-user
device deployment the 0700/0600 file perms match the rest of the wallet.
Audit-verified clean: dataDir 0700, hsm_secret/signer.log/signer.pid 0600,
device PEMs 0600, seed write-once (a different mnemonic can never desync
the signer from the node), signer liveness tracks the live process.

See [ARCHITECTURE.md](../lnclient/greenlight/ARCHITECTURE.md) for the
full design rationale (why we integrate through `glcli`, how GL primitives
map to the Go layer, and what operational additions we provide).
