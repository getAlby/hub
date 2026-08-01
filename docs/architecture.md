# Routstr Architecture

This document explains the Routstr daemon in detail, why it settles with Cashu ecash instead of paying providers over Lightning, and how the whole implementation is a gentle, pattern-conforming addition to Alby Hub.

**TL;DR**

- The daemon is an OpenAI-compatible router with its own key registry and Cashu wallet. Providers are a federation discovered over Nostr, and every request routes to the cheapest one.
- It pays providers with ecash because that is what they accept, and it is cheap. Lightning is only the entry and exit (buying ecash).
- It fits the Hub without touching it: no Hub schema changes, standard app creation and app-wallet money rails, a normal service manager, an extended backup, and the existing UI patterns.

## The daemon

### Role

`routstrd` is a small service (Bun) that:

- exposes an **OpenAI-compatible API** (`/v1/chat/completions`, `/v1/models`)
- holds the **client key registry** (keys are native to the daemon that minted them)
- owns a **Cashu wallet** (via `cocod`, the Cashu daemon) that pays for requests
- routes each request to the cheapest provider in a federation

### The provider federation

Routstr does not depend on one AI vendor. Providers are community-run Routstr endpoints that announce themselves over **Nostr**: the daemon subscribes to provider-announcement events on relays, validates each endpoint against the announcing pubkey, and keeps a persistent registry (`provider_index`) plus a blocklist (`disabled_providers`) for dead or malicious nodes. Model catalogs are fetched from every live provider and aggregated into one searchable catalog, cached with a 30-minute TTL and a background warm-refresh loop.

### The request lifecycle

```
Your app → base URL + key → daemon
  1. client lookup (usage accounting)
  2. wallet balance gate (Cashu mint balance)
  3. pick the cheapest provider that has the model
  4. call the provider, authenticated with an ecash token
  5. stream the response back
```

Providers that fail are skipped for 5 minutes; the next cheapest takes over.

### The payment model (verified from source)

Providers are paid with **prepaid ecash tokens, not invoices**:

- The daemon holds a prepaid token per provider (visible as the `apikey:<provider>` entries in the key balance).
- Each request is authenticated to the provider with that token (`X-Cashu` header in token mode, `Authorization: Bearer` with the token in key mode).
- The provider spends proofs per request. When a token runs out (`401 proofs already spent`), the daemon tops it up from its own Cashu wallet (`cashuSpender.spend`) and fails over.
- Unspent token balance can be claimed back through the provider's `/v1/wallet/refund` endpoint, which returns an ecash token the daemon receives back into its wallet.

There is a second mode (`xcashu`) where a caller without a client key can pay by attaching an `X-Cashu` token directly to each request. On this Hub the daemon runs in `apikeys` mode: clients authenticate with keys, and the spend comes from the daemon's wallet.

## Why Cashu

Routstr uses ecash because that is what the provider network accepts: **providers receive payments only via Cashu**. The daemon carries a cocod Cashu wallet for exactly this reason, and Cashu payments are cheap (mint fees, no channel or routing costs), which is why the providers prefer them.

The rest follows from the mechanics, verified in the daemon:

**1. No bilateral Lightning relationships.** An ecash token carries its own mint, and that mint is the clearing house for its tokens. The token is both the credential and the payment, so there is nothing to set up between the requester and any individual provider.

**2. No per-request invoice round-trip.** A Bolt 11 flow would add an invoice request and payment wait to every completion. Ecash spending is a local operation on proofs the daemon already holds; the provider verifies them at the mint.

**3. Refundability.** An unspent token is just unspent proofs: the provider returns it and the daemon receives it back into its wallet. A channel has to be closed and settled; a token is simply handed back.

**4. Lightning is still the entry and exit.** Buying ecash happens at the mint via a Bolt 11 invoice, paid by the Hub's Lightning node from the Routstr app wallet. The Hub side never needs ecash; only the daemon does.

## How this fits Alby Hub

The implementation is designed as a minimal, pattern-conforming addition. Every capability it needs already exists in the Hub; the fork reuses it rather than re-implementing it.

### Backend-agnostic on the Hub side

Routstr works with any Hub Lightning backend. Funding buys ecash through a standard Bolt 11 mint invoice paid by the Hub's Lightning node from the Routstr app wallet (`/api/payments/{invoice}` with `fromAppId`). Refunds mint ecash back to the Routstr app's NWC wallet via an app-scoped invoice. The ecash layer is entirely the daemon's domain: the Hub side is a normal app with a normal wallet and the standard payment APIs.

### The Hub database schema is untouched

There are no new tables, columns, or migrations. Routstr state lives in two places the Hub already has:

- **App metadata** (key, client id, balance, auto top-up config)
- **The daemon's own databases** (`routstr.db`, `coco.db`), which the Hub treats like any external service's state

> **Lifecycle hazard (hit 2026-08-01):** the auto top-up config and the isolated wallet are *owned by the app*. Deleting a Routstr app (Apps Cleanup or the app-page Delete) removes the config with it and strands the wallet's sats — `DeleteApp` deletes only the app row; it does not refund the wallet, and backup zips are client-downloaded (none exist on disk). After any cleanup/reinstall, re-enable auto top-up with Start — it uses the sane defaults (threshold 500, amount 1000) when no config exists.

### Wallet setup is the standard app flow

The wizard creates the Routstr connection through the same path as every other Hub app: `createApp` with NWC permission scopes (`ISOLATED_SCOPES`), an optional budget and renewal, and `isolated: true`. The result is an ordinary Hub app with its own app wallet, exactly like any NWC app. The money rails are the Hub's own (endpoints in [reference.md](reference.md)):

- Funding the app wallet: a Hub transfer (with `fromAppId` where applicable)
- Buying ecash: the Hub pays the mint's invoice from the app wallet
- Refunding: an **app-scoped invoice** that the daemon melts ecash to pay, so sats land directly back in the Routstr app wallet

No custom payment code touches the main wallet. Everything is the standard app-wallet model.

### Background management follows the service pattern

`RoutstrdService` is a normal Hub service manager: constructed in `service.go`, started and stopped with the Hub lifecycle, and exposed through a getter like the other services. It owns the daemons, health-checks them, restarts them, and runs the auto top-up loop (15-second tick reading the `autoRefill` config from app metadata; refills the Cashu wallet from the Routstr app wallet below the threshold, 5-minute cooldown). Nothing about supervision is special-cased in the Hub core: it is one more service in the manager.

### Backup is an extension of the existing mechanism

The Hub already had an encrypted backup archive (`nwc.db`, the Lightning backend state, configs). The fork extends the *same* archive with the daemon databases (`coco.db`, `routstr.db`) and their configs, checkpoints their WAL files before archiving so nothing recent is lost, and restores them under the user's home directory. Same crypto, same restore flow, same recovery procedure: after a restore the Hub process is restarted, then unlocked. See [backup-restore.md](backup-restore.md).

### The UI rides the Hub's own patterns

- Routstr is a **suggested app** in the Connections app store (logo, description, link), like every other app
- The **connection page** is the standard app detail page with a Routstr section
- The **wizard** is an internal-app wizard (`Routstr.tsx` alongside ZapPlanner and the other internal apps), using the same step layout, `Permissions` component, and isolated-app configuration
- New HTTP surface is two routes: `/api/routstrd/*` (JWT-authenticated proxy to the daemon for management) and `/routstr/v1` (the OpenAI-compatible proxy for external clients)

### What the fork does *not* do

- No changes to the Hub database schema
- No changes to the main wallet or the Lightning backend
- No new funding paths outside the app-wallet model
- No changes to NWC protocol handling (NWC is used for pairing and status, as with any app)

The daemon is a self-contained service that happens to be a Hub app. That is the point: from the Hub's perspective, Routstr is a normal app with a normal wallet; the Cashu layer is entirely inside the daemon's own domain.
