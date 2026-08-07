# Mainnet readiness test plan — GL backend in Alby Hub

Gate: do each step in order. Do not start the next until the current one passes and evidence is recorded. Use small real amounts. This is the "rigor before real money" checklist.

## Phase 0 — Funded node on mainnet (the common pillar)
- Register a real Blockstream **mainnet** node (glcli `-n mainnet` register) with dev/staging certs; the network-map fix already covers `mainnet→bitcoin`.
- Point the Hub DB at it (`.env NETWORK=mainnet`, GreenlightNodeURI, CredsPath, SignerDataDir), restart, unlock.
- **Pass =** `/api/info` shows GREENLIGHT / mainnet / running; signer logs `-n mainnet`; `POST /api/invoice` returns a real `lnbc` (mainnet) BOLT11, pending.
- Gate: this phase must be stable across 3 restarts before anything else.

## Phase 1 — Funded real channel + small payment (SMALL: 10–50k sats total)
- Fund the node's onchain address from a working mainnet faucet/self-send.
- Open a 1M–5M sat channel to a healthy mainnet peer with routing.
- Issue an invoice on the peer; pay it through the Hub API (`POST /api/payments/:invoice`).
- **Pass =** tx in DB `outgoing / settled / <amount>` with a released **preimage**; balances reflect the route fee. This is the exact scenario I could NOT finish on testnet (no faucet) and must be proven on mainnet with real sats.

## Phase 2 — Backup / restore round-trip on a FUNDED mainnet node
- Trigger `/api/backup` → encrypted `.bkp` (note: it StopApp + db.Stop by design; API is down until `systemctl restart`).
- Wipe `.data`, restore via `/api/restore` (multipart: unlockPassword + backup file, NO jwt, setupCompleted=false) → 204.
- **Verify =** same node pubkey, same mnemonic, same onchain/channel balance, same invoice history. This is the data-loss insurance; it must be proven with money actually in the node.

## Phase 3 — Recovery under fault
- Kill the supervised signer mid-operation; confirm the supervisor restarts it without corrupting state.
- Confirm a crashed-while-paying tx does not coinbine/leak; the failure path returns a clean error.
- Confirm `ResetRouter` stub is never invoked in normal operation (it's a known intended stub from my audit; make sure no path calls it during regular use).

## Phase 4 — Security / release polish
- **Hardening already committed:** `.gitignore` now ignores `hub`, `.data.*`, `.env.bak*`; the network-mapping bug is fixed and committed (`535e66e`).
- Review the service is not exposed: confirm the HTTP/API port is not internet-reachable without auth (reference: routstrd :8008 bind lesson — never run the fund wallet reachable unbindfirewalled).
- Secret/seed material: no hsm_secret / GL creds committed (verify `git ls-files | grep -iE 'hsm|gfs|pem'` is empty).
- Run the full test suite green on the committed ref.

## Phase 5 — Regression on the release branch
- `MOVED ALL WORK: still ~20 modified + 5 untracked GL files. Before mainnet, decide: commit the remaining delta to the fork and open the PR against upstream getAlby/hub, OR keep on `feat/greenlight-backend`. At minimum the working tree should be committed so there's a reproducible build and rollback point.

## Explicit non-goals (do not slip in)
- No auto top-up / auto route logic changes here (separate feature).
- No signing/verify of the release binary yet — out of scope unless asked.

## Definition of done
Only when Phases 0–4 all pass with recorded evidence and the tree is committed/PR-d is the GL backend acceptable to run on mainnet. Until then: **dev/testnet/regtest only, small amounts, nothing internet-reachable without auth.**