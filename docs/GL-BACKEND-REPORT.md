# Greenlight backend for Alby Hub: what and why

Status snapshot: 2026-08-07, branch `feat/greenlight-backend`, commits `535e66e` + `19ec7cc` on upstream base `bf9c346`.

## Goal

Make Greenlight a first-class LN backend inside Alby Hub. Cloud node (Blockstream CLN) with self-custody keys on the device. Custody split:

- Hub owns: control plane, NWC, encrypted 12-word mnemonic, supervised signer process, payment pump
- Blockstream owns: lightningd and the channel DB

Why this shape: LDK is fully local and heavy, Bark/Ark has no channels. Greenlight gives an always-on node with real channels and user-held keys. Hub supervises the signer the same way Routstrd supervises its daemons.

## Research findings (source-verified)

- Node seed = 12-word BIP-39 mnemonic, derived as `to_seed("")[0:32]` (gl-cli `util.rs`). One-way derivation, 24-word phrases rejected
- Recovery: seed alone restores node identity, device creds, on-chain funds. Channels additionally need the signer's periodic backup file
- Signer is a local process; without it money RPCs hang. Supervision pattern mirrors `RoutstrdService` (15s health ticker, respawn)
- GL networks: bitcoin, testnet, regtest. Signet unsupported (live TLS 223)

## Build

Backend (`lnclient/greenlight/`):
- `greenlight.go`: full 36-method LNClient over GL gRPC (invoices, pays, keysend, balance, channels, waitanyinvoice, streams)
- `provision.go`: mnemonic to seed, register/recover via Nobody dev cert, schedule, extract mTLS PEMs + node URI
- `service/gl_signer.go`: supervised `glcli signer run` (health loop, respawn, clean kill)
- `service/start.go`: product startup path, network mapping fix (testnet passes through)
- Wired at the 4 boot-contract points: config/models.go, config/config.go, service/start.go, constructor

Frontend:
- `GreenlightForm.tsx` (12-word setup), `SetupRecover.tsx`, GL-aware `Backup.tsx`, `MigrateNode.tsx`
- Registered with `hasMnemonic: true`, `hasChannelManagement: true`, `hasNodeBackup: false`

Delta: 40 files, +5243/-131.

## Testing ladder

1. Mock suite (23 tests, in-process bufconn gRPC)
2. Local gl-testing harness (regtest): real lightningd + gl-plugin + VLS signer proxy, v25.05gl1
3. Two-node in-harness E2E: channel opened, 2x50k received + 1x10k sent settled
4. TestLiveE2E gated behind GREENLIGHT_LIVE=1

## Live testnet results

- Registered real Blockstream testnet nodes with dev Nobody certs, mTLS gRPC connected
- Fixed real bug: start.go forced non-mainnet to signet, signer would be read-only on testnet
- Redeployed from committed source, deployed binary == ./hub byte-for-byte
- Minted lntb invoices live (proves signer attached)
- Deterministic restore: two scheduler recover runs from one phrase reproduce the same node pubkey

## Verified gaps (honest)

- Nothing pushed upstream; no PR (this private repo is the housing)
- Live send on testnet blocked: fresh node has zero on-chain, all public testnet faucets down
- Mainnet channel + fund path not exercised

## Three-way comparison (next phase)

See `COMPARISON-MATRIX.md`. Deploy upstream Greenlight standalone, upstream Hub standalone, run default flows on testnet, compare against this backend.

## Excluded from repo (operator secrets stay local)

- `/root/gl-tools/certs` (GL_NOBODY_CRT/KEY)
- `.data`, `.env`, compiled binaries
