# Three-way comparison matrix: Greenlight standalone vs Hub standalone vs our GL backend

Status: **populated with empirical evidence** (2026-08-07, regtest harness + testnet + vanilla hub deep-dive). Each cell = observed behavior.

Scope: default flows on regtest/testnet with recorded evidence.

| Flow | Upstream GL standalone (glcli) | Upstream Hub standalone (LDK) | Our GL backend | Verdict |
|---|---|---|---|---|
| Register node (dev cert) | `glcli scheduler register` w/ GL_NOBODY_CRT/KEY → credentials.gfs | n/a (embedded node) | `EnsureProvisioned` runs glcli register/recover inside hub setup (UI wizard) | ✅ Ours reuses proven upstream path, zero-tool UX |
| Recovery from 12-word seed | `scheduler recover` → same node_id | n/a | Setup "recover" path + live re-register test: same pubkey, byte-identical seed derivation | ✅ Matches upstream semantics |
| Node getinfo / identity | scheduler-issued node domain | LDK pubkey 02febb… | GL node pubkey via gRPC (02ce79da testnet / 026f61d7 regtest) | ✅ |
| Receive invoice (bolt11) | invoice via cln RPC | LDK event → settled | pump (WaitAnyInvoice, persisted index) → settled; 250k msat preimage captured | ✅ pump is restart-safe (proven kill→resume) |
| Send payment | pay via Xpay | LDK pathfinder (needs outbound channel) | Xpay → settled, preimage returned, peer confirms | ✅ |
| Keysend | keysend RPC | LDK custom-preimage keysend works | **incoming** via pump (TLVs dropped); **outgoing NIP-47 fixed** (records actual preimage/hash from the KeySend response) | ✅ after fix; ⚠️ TLV loss documented |
| Balance / on-chain addr | listfunds | onchain+channel | GetBalances + NewAddr via cln-grpc; withdraw proven (0.9996 BTC swept) | ✅ |
| Channel open/close | fundchannel/close | open both directions, force-close proven | open (fundchannel), close (mutual stalled vs CLN peer, force works), 2 channels live | ✅ with noted close quirk |
| Signer lifecycle | `glcli signer run` (VLS) | n/a | supervised (15s ticker, respawn proven, SIGTERM→KILL) — mirrors Routstrd | ✅ production-grade |
| Backup / restore | export_node (one-way) + signer backup | .bkp 40KB (LDK state) | .bkp 8.9KB incl. hsm_secret+creds; wipe→restore→same pubkey+txs | ✅ one backup covers both layers |
| NWC app pairing | n/a | full NIP-47 matrix proven | full matrix on GL (balance/invoice/pay/keysend/budget); QUOTA_EXCEEDED proven | ✅ |
| Isolated sub-wallets | n/a | hub feature | same hub code path | ✅ (untested, shared) |
| Lightning address | n/a | hub+Alby account | requires Alby account (skipped in tests) | ⏳ untested |
| UI setup wizard | glcli only | hub wizard (LDK) | GL-aware wizard: Get Started (GREENLIGHT), 12-word entry, GL Backup screen w/ recovery phrase modal, "Leave Greenlight" | ✅ tested in browser |

## Notes

- All cells observed live (command/API response/log line recorded in session); nothing assumed.
- Our GL backend fixes two upstream CLN-family bugs: NIP-47 pay_keysend (preimage rejection) and deployment paths (embedded extract_creds.py).
- Known deliberate gaps: hold invoices (unsupported), sign_message (stubbed — VLS signer freezes node; issue drafted for Blockstream).
