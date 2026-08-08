# Greenlight backend — operations runbook

Applies to the Greenlight backend (`lnclient/greenlight`). The node runs on
Blockstream infrastructure; your device holds the VLS signer keys. The hub
process supervises the signer and connects to the node over gRPC.

## Failure modes

### 1. Node frozen (hsmd queue wedge) — the critical one

**Cause**: a signing request the VLS signer cannot answer (historically
`signmessage`, see KNOWN-ISSUES.md and Blockstream/greenlight#739) wedges the
node's hsmd queue. The node stops processing: payments stall, invoices don't
settle, every RPC hangs.

**Signs**:
- Hub log: `greenlight node is unhealthy: RPC or invoice path stalled (possible frozen node...)`
  (after 3 consecutive failed health probes, ~90s)
- `/api/node/status` → `isReady: false`, `internalNodeStatus.healthy: false`
  (the API keeps answering — the watchdog serves a cached verdict)
- The node's lightning-rpc stops responding (all calls hang)

**Recovery**:
1. **Hosted (production)**: the node must be restarted by Blockstream —
   contact support / open an issue, reference #739. The hub stays usable
   (degraded) meanwhile: balances (cached), backups, apps all work; new
   payments will fail until the node returns.
2. **Self-hosted harness/dev**: restart the node process with its original
   invocation (the supervisor does this for gl-testing; for a manual node,
   restart `lightningd` with the same args + env, including `GL_NODE_INIT`).
3. After the node returns, the watchdog publishes the healthy event within
   30s and payments resume. No hub restart needed.

**Prevention**: the backend never calls `signmessage` (stubbed — it would
re-trigger the freeze). Do not add it back without Blockstream fixing #739.

### 2. Signer process died (device side)

The hub's supervisor respawns the signer (15s ticker; SIGTERM→KILL escalation).
A payment in flight during the window fails cleanly; retry it. The signer is
the only holder of the signing keys — never delete its data dir
(`GREENLIGHT_SIGNER_DATA_DIR`) without a recovery phrase backup.

### 3. Hub crash / restart

The incoming-payment pump persists its `pay_index` (`pump_state.json`) — on
restart it replays anything that settled while the hub was down. No payments
are missed (this is the restart-safe design; other backends rely on
reconciliation polling).

### 4. Node unreachable at hub boot

The hub starts **degraded** (never locks the user out): the backend boots,
the watchdog probes, `/api/node/status` reports `healthy: false`. Fix the
node connectivity; the watchdog clears itself.

## Backup / restore

1. Settings → Backup: disable **Auto Unlock** first (the backup endpoint
   refuses while it's on), then download the hub `.bkp`.
2. The `.bkp` contains node creds + signer seed — one backup covers both
   layers (verified: wipe → restore → same pubkey, same transactions).
3. The recovery phrase (Settings → Backup → recovery phrase) is the signer
   seed: `recover` re-derives the node identity from it.
4. Restore: fresh hub instance → Settings → Backup → restore `.bkp`.

## Leaving Greenlight (one-way)

The exported node backup is a one-way exit: you can export the node's data
for a self-hosted CLN recovery path, but you cannot import an external node
into Greenlight. Export before deleting the hub.

## Monitoring

- The watchdog publishes `nwc_gl_node_health` events (`healthy: true/false`)
  — wire these to your alerting (they are the node-health signal).
- The signer service publishes `nwc_gl_signer_health` events on state transitions
  (`healthy: true/false` + `error` on failure) — wire for the signing-key health signal.
- `/api/node/status` carries both node health and signer state in
  `internalNodeStatus` (`healthy`, `last_error` for the node, `signer.running` /
  `signer.last_error` for the signer). A dead signer is surfaced there rather
  than being reported as a node outage.
- Hub ERROR logs with `greenlight node is unhealthy` or `greenlight signer unhealthy`
  are the actionable alerts.
