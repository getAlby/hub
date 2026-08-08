# Greenlight backend — known issues (honest gaps)

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
