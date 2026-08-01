# Reference

## Daemon API (routstrd)

Base: `http://localhost:8008` (or through the Hub proxy).

| Endpoint | Purpose |
|---|---|
| `GET /v1/models` | aggregated model catalog (30-min cache) |
| `POST /v1/chat/completions` | OpenAI-compatible completions |
| `GET /clients` | client registry (the source of truth for key identity) |
| `POST /clients/add` | mint a new client key |
| `POST /clients/delete` | remove a client |
| `GET /keys/balance` | wallet balance + prepaid provider tokens |
| `GET /usage/summary` | usage accounting by client |
| `GET /status`, `GET /health` | health |
| `POST /wallet/receive/bolt11` | mint a Lightning invoice to buy ecash |
| `POST /wallet/send/bolt11` | melt ecash to pay a Lightning invoice |
| `GET /usagePi` | (typo) alternate usage endpoint |

Provider federation: `provider_index` and `disabled_providers` live in the daemon store; providers announce over Nostr and are validated by pubkey.

## Hub API additions

| Endpoint | Purpose |
|---|---|
| `/api/routstrd/*` | JWT-authenticated proxy to the daemon for management |
| `/routstr/v1/*` | public OpenAI-compatible proxy for external clients |
| `/api/transfers` | wallet-to-wallet transfer (with `fromAppId`) |
| `/api/payments/{invoice}` | pay an invoice from a specific app wallet (`fromAppId`) |
| `/api/invoices` | create an invoice, optionally app-scoped (`appId`) |

## Money flow mapping

| Flow | Steps |
|---|---|
| Top Up / Fund Key | daemon `POST /wallet/receive/bolt11` (mint invoice) → Hub `POST /api/payments/{invoice}` with `fromAppId` → mint issues ecash |
| Refund | Hub `POST /api/invoices` with `appId` (app-scoped) → daemon `POST /wallet/send/bolt11` melt → sats credit directly to the Routstr app wallet |
| Auto top-up | Hub supervision loop (15s tick) reads `routstr.autoRefill` metadata; below threshold, funds the Cashu wallet from the Routstr wallet; 5-min cooldown |

## Config keys

Daemon (`~/.routstrd/config.json`): `port` (8008), `provider`, `cocodPath`, `mode` (`apikeys` | `xcashu`), `nsec`, `nwc`. There is no bind-address option.

Hub: unlock password, JWT secret (persisted encrypted), Lightning backend (Bark on this deployment), mint URL.

## App metadata schema

```json
{
  "app_store_app_id": "routstr",
  "routstr": {
    "apiKey": "sk-...",
    "clientId": "routstr-app-<id>-<ts>",
    "balance": 0,
    "autoRefill": {
      "enabled": true,
      "threshold": 500,
      "amount": 1000,
      "cooldownMs": 300000
    }
  }
}
```

`app_store_app_id` lives only in metadata (no DB column). PATCH replaces the whole metadata object; always read-modify-write.
