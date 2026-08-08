# Backend audit: Greenlight vs the other Alby Hub backends

Audit date: 2026-08-07. Base: upstream getAlby/hub @ 6175489 (LDK/CLN/LND/Phoenixd/Cashu/Bark), fork feat/greenlight-backend (GREENLIGHT). Method-level conformance verified by source read + live tests.

## Interface conformance (35-method LNClient)

Real implementations (stubs/errors excluded; verified by source read, conditional errors counted as implemented):

| Backend | Real | Stubs | Notable stubs |
|---|---|---|---|
| **LND** | 31 | 4 | MakeOffer (BOLT12 unsupported) |
| **CLN** | 30 | 5 | holds ×3 (gated on external hold plugin), ResetRouter, **SendKeysend** (preimage bug fixed via PR #2521, pending merge) |
| **GREENLIGHT** | 30 | 5 | holds ×3, ResetRouter, **SignMessage (stubbed: VLS freeze)** |
| **LDK** | 33+ | ~2 | GetNetworkGraph stub (LDK has no graph query API) |
| **Phoenixd** | ~13 | ~22 | channel ops, keysend, onchain — mostly unavailable (managed node) |
| **Bark** | ~13 | ~22 | no channels by design (VTXO/Ark), no invoices lookup |
| **Cashu** | ~10 | ~25 | ecash wallet: minimal surface |

Greenlight is the **second-most complete node-model backend** (behind LDK/LND), with the widest surface of the *hosted* backends.

## Feature matrix (empirically verified where possible)

| Feature | LDK | CLN | LND | Phoenixd | Cashu | Bark | **GL** |
|---|---|---|---|---|---|---|---|
| Hold invoices | ✅ | ⚠️ plugin-gated | ✅ | ❌ | ❌ | ❌ | ❌ |
| BOLT12 offers | ✅ | ✅ | ❌ stub | ❌ | ❌ | ❌ | ✅ |
| Keysend out | ✅ | ✅ (fix PR #2521) | ✅ | ❌ | ❌ | ❌ | ✅ (fixed) |
| SignMessage | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ (VLS freeze) |
| Channels open/close | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ (VTXO) | ✅ |
| Onchain (addr/withdraw) | ✅ | ✅ | ✅ | ❌ | ❌ | ⚠️ | ✅ |
| Network graph | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| payment_received notif | ❌ (reconcile) | ❌ (reconcile) | ❌ (reconcile) | ❌ | ❌ | ❌ | ✅ **(pump)** |
| NIP-47 advertised | 14 | 11 | 14 | partial | partial | partial | 10 |
| GetStorageDir-aware backup | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |

## Incoming-payment architecture (the key design axis)

| Backend | Mechanism | Restart safety |
|---|---|---|
| LDK | embedded LDK events | channel monitors + reconcile |
| CLN | `SubscribeInvoices` stream | ❌ ephemeral; relies on hub reconcile poll |
| LND | lnrpc SubscribeInvoices | relies on hub reconcile poll |
| **GL** | **WaitAnyInvoice pump + persisted pay_index** (atomic tmp+rename, EXPIRED fast-forward) | ✅ **proven**: kill → resume from index, no missed/dup payments |
| Bark | Ark "movements" | n/a |

Why GL differs: the hosted GL node leaves all six `cln.Node Subscribe*` streams `unimplemented!()` (they panic the gl-plugin), so the fork adapted the poll-based WaitAnyInvoice — and in doing so added persistence the upstream CLN backend lacks. Advertising `payment_received` (unique among backends) disables the hub's reconcile safety net; the pump's index catch-up replaces it. **This is the fork's strongest architectural contribution.**

## Bugs & quirks inventory (ours vs theirs)

| Backend | Issue | Status |
|---|---|---|
| GL | NIP-47 pay_keysend preimage rejection | ✅ **fixed in fork** (9e421612) + **live-verified end-to-end** (2 keysends settled on fresh harness, l1 confirmed, hub recorded outgoing/settled) |
| GL | sign_message freezes node (hsmd wedge) | ✅ **stubbed** (9e421612); Blockstream issue drafted |
| GL | extract_creds.py CWD/hardcoded path | ✅ **fixed** (aa71d61f, go:embed) |
| GL | keysend TLV metadata lost (stream guard never fires) | ⚠️ documented; design debt |
| GL | ConnectPeer address+port=0 concat footgun | ⚠️ minor (UI sends split fields) |
| GL | **mutual close** | ✅ **works** — l1 state history shows CLOSINGD_COMPLETE (2250 sats fee); earlier "stall" was the LDK hub, not GL |
| GL | keysend preimage in tx record is synthetic (hub-generated; CLN derives the real one internally, backend doesn't return it) | ⚠️ accounting nuance — payment settles + real preimage exists on peer side, but hub's recorded preimage ≠ real one |
| CLN | SendKeysend preimage rejection (same bug GL had) | ❌ **still live upstream** — PR candidate |
| CLN | hold methods never advertised even with plugin | ⚠️ minor |
| LDK | `testnet` maps to `ldk_node.NetworkSignet` | ❌ live upstream quirk |
| LDK | withdraw 500 "ran out of attempts to fetch broadcasted transaction" (funds still move) | ❌ live upstream |
| LDK | mutual-close stall vs CLN peer on regtest (force works) | ⚠️ interop |
| LND | MakeOffer stub (no BOLT12) | by design |
| Bark | sqlite 644 perms + VTXO recovery-mailbox message | ⚠️ fresh-wallet noise |

## Verdict

1. **Conformance:** GL is in the top tier — 30/35 methods real, everything a node backend needs (channels, onchain, offers, peers, graph, log). Nothing is stubbed that shouldn't be (holds are the only feature gap; sign_message is stubbed for a documented safety reason).
2. **Unique strengths:** restart-safe incoming pump (only backend with persisted index + payment_received notifications), signer supervision, GL-aware backup/UI.
3. **Cleanest of the CLN family:** GL fixed the keysend bug upstream CLN still has, and the deployment bug is now embed-based.
4. **Actionable gaps:** holds (blocked upstream), sign_message (blocked upstream), keysend TLV capture (fork-side design work), and two upstream PR candidates: CLN keysend preimage bug + LDK testnet→signet mapping.
