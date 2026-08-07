# Greenlight backend for Alby Hub

A dedicated `lnclient` backend that runs an Alby Hub against a
[Blockstream Greenlight](https://blockstream.com/greenlight/) node: a real
Core Lightning node hosted by Blockstream, with the signing keys held in a
separate VLS (Validating Lightning Signer) process.

## What you get

- Full NIP-47 surface: payments (pay_invoice, pay_keysend), invoices,
  lookup, balances (channel + onchain), channels (open/close/update),
  onchain (withdraw, new address), offers (BOLT12), peers, network graph,
  sign_message.
- Incoming payments: `WaitAnyInvoice` pump with a persisted pay index
  (restart-safe catch-up, no missed payments) + `StreamIncoming` for
  incoming keysend with TLV records (NIP-47 app identification works).
- Custody: the hub never holds the seed. All signing happens in the
  signer process (VLS), which validates channel state and fees before
  signing — a compromised hub cannot move funds.

## Not supported (by design)

- Hold invoices (GL/cln node does not expose them)
- `ResetRouter` (LDK-only concept)
- Channel state notifications (`nwc_channel_ready` / `nwc_channel_closed`):
  all six `cln.Node` `Subscribe*` stream RPCs are `unimplemented!()` stubs
  on the GL node server — calling one **panics the gl-plugin process and
  kills the node** (verified live). The backend deliberately never calls
  them. `StreamNodeEvents` currently emits only `InvoicePaid`.

## Setup

1. **Get a node** — with `gl-cli` (Blockstream's CLI):

   ```sh
   cargo install gl-cli
   gl-cli register --network bitcoin   # or recover an existing seed
   # creates ~/.local/share/greenlight/{credentials.gfs, hsm_secret}
   ```

2. **Extract credentials** — the hub needs the device credentials as PEMs:

   ```sh
   python3 lnclient/greenlight/extract_creds.py \
     ~/.local/share/greenlight/credentials.gfs ./greenlight-creds
   # writes ca.pem, client.pem, client-key.pem, rune
   # prints the node URI, e.g. gl1<node_id>.gl.blckstrm.com:443
   ```

3. **Run the signer** — separate process, encrypted seed:

   ```sh
   python3 lnclient/greenlight/signer/encrypt-seed.py ~/.local/share/greenlight
   # installs lnclient/greenlight/signer/greenlight-signer.service (or run
   # greenlight-signer-run.sh by hand)
   ```

4. **Configure the hub**:

   ```sh
   LN_BACKEND_TYPE=GREENLIGHT
   GREENLIGHT_CREDS_PATH=/path/to/greenlight-creds
   GREENLIGHT_NODE_URI=gl1<node_id>.gl.blckstrm.com:443
   # optional:
   GREENLIGHT_SERVER_NAME=gl1<node_id>.gl.blckstrm.com   # TLS SNI, derived from URI if unset
   GREENLIGHT_SIGNER_DATA_DIR=/home/greenlight/.local/share/greenlight  # include seed in hub backups
   ```

## Architecture notes

- **Wire**: the node speaks the cln-grpc protocol (package `cln`, service
  `cln.Node`) over one mTLS port. The backend reuses the hub's vendored
  cln-grpc bindings (`lnclient/cln/clngrpc`), verified wire-compatible.
  `StreamIncoming` lives on the `greenlight.Node` service (not in the
  vendored bindings) and is consumed with a raw codec + protowire decoder
  in `streamincoming.go`.
- **Reconcile**: the hub's invoice reconcile stays on until this backend
  advertises `payment_received`; the pump's persisted index is what makes
  that safe (see `waitanyinvoice.go`).
- **Sends**: synchronous `Xpay` (preimage + fees returned), same as the
  CLN backend.
- **Duplicates**: invoice payments are published only by the pump;
  `StreamIncoming` publishes keysends only (its invoice events are
  ignored) — publishing both would double-create transactions.

## Testing

All functionality is tested against an in-process mock node (the vendored
`RegisterNodeServer` stub + a hand-registered `greenlight.Node` stream), so
the full suite runs without a live node or credentials:

```sh
go test ./lnclient/greenlight/ -v
```

### Live E2E (real node)

`live_e2e_test.go` runs against a real GL stack and is skipped unless
`GREENLIGHT_LIVE=1` plus the standard env vars are set. It has been
executed end-to-end against a fully local Greenlight stack (real
`lightningd` + real `gl-plugin` gRPC + real VLS signerproxy, the same
binaries Blockstream runs), proving:

- **production lightningd version is `v25.05gl1`** (readable only via a raw
  `Getinfo` — the hub's `NodeInfo` doesn't expose it);
- **`WaitAnyInvoice` timeout semantics**: with no pending payments it
  returns RPC error 904 `Timed out while waiting for invoice to be paid`
  after the requested timeout — the pump treats this as "nothing new" and
  keeps polling (no error spam, no missed payments);
- **`StreamIncoming`** establishes and stays open (no immediate close),
  confirming the raw-codec zero-codegen consumer works against the real
  `greenlight.Node` service;
- real invoices created/queried, real balances/peers/graph listed.
  Keysend and receive subtests are conditional on channel balance.

To reproduce on a box without a Blockstream node, run the gl-testing local
stack (requires `bitcoind`, `cfssl`, `cfssljson`, `cln-version-manager`,
and `cargo build -p gl-plugin -p gl-signerproxy` in a clone of
`Blockstream/greenlight`), then a harness test that registers a client,
starts the node, and exports `credentials.gfs` + `ca.pem` (see
`libs/gl-testing/tests/test_live_harness.py`). Note the gl-testing
in-process python signer hangs on `signmessage` (freezing the node), so
`SignMessage` is skipped in live mode; the production signer is the Rust
VLS binary, a different implementation.
