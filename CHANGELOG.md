# Changelog

All changes since forking [getAlby/hub](https://github.com/getAlby/hub).

## Routstr integration

- **Wizard**: 5-step setup (Configure, Top Up, Create Key, Fund Key, Done). Uses the standard app-creation flow (`createApp`, `ISOLATED_SCOPES`, isolated app wallet).
- **Connection page**: API balance with refresh, models browse, Top Up, Refund, Delete Key, auto top-up, base URLs. Rendered on the standard app detail page.
- **API keys**: created by the daemon (one key, all providers). Key + client id stored in app metadata. Ghost-key self-heal against the daemon `/clients` list.
- **Delete**: 404-tolerant (already-removed = success).
- **Models browse**: full catalog, searchable, with per-model sats pricing.
- **Copy**: "API key balance" naming, selection-aware top-up balance, Back → Send → Next footer.

## Supervision

- `RoutstrdService` manages routstrd + cocod: start, stop, restart on health failure (15s tick).
- Auto top-up loop reads `routstr.autoRefill` app metadata (threshold, amount, cooldown).
- **Auto top-up control (2026-08-01):** `GET /api/routstrd/autorefill/status` + `POST /start|/stop` (Hub-native, not daemon-proxied). The UI card uses explicit Start/Stop buttons with a live status line (pool balance, last refill, last error, 30s poll). `enabled` is owned server-side by the Start/Stop API; input blur-saves never resurrect a stale value. A mutex serializes the 15s ticker against the Start-triggered immediate check (was double-refilling). `findRoutstrApp` honors: enabled config > any configured app > first Routstr app, so status stays on the right app even while stopped.
- **Ops hardening (2026-08-01):** `alby-hub.service` systemd unit (Restart=always, `PATH` includes `/root/.bun/bin` for the bun-shebang daemons) + auto-unlock, so a crash or reboot self-heals in ~15s with no manual step. `deploy.sh` runs build → restart → unlock → verify atomically.

## Money rail

- Hub-direct funding: daemon mint invoice, paid by the Hub from the Routstr app wallet (`fromAppId`).
- Hub-direct refunds: app-scoped invoice, melted from the daemon's Cashu wallet, sats land directly in the Routstr wallet.
- Never touches the main wallet.

## Backup

- Archive now includes `coco.db`, `routstr.db`, their configs, and Bark state.
- WAL checkpoint (TRUNCATE) before archiving.
- `../` restore entries resolve against `$HOME`.
- Restore requires a process restart before unlock.

## Frontend

- New internal-app wizard (`Routstr.tsx`), Routstr connection components, `useRoutstrd` hook, suggested-app store entry.
- **Live app page (2026-08-01):** the app detail page polls (3s app data, 10s transactions) so new payments appear without a manual refresh. The Total Spent/Received tally now mirrors the balance computation (counts pending outgoing, settled incoming only), so `Balance = Received − Spent` always holds. Auto top-up card warns when the refill amount is under ~100 sats (fees dominate) or the Routstr wallet can't cover a refill.

## Security

- Daemon `:8008` exposure documented; firewall required before funding.
- Key model documented: keys are native to their daemon; the wallet balance is the spend gate.
