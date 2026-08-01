# Development

## Build

```sh
cd frontend && yarn install && yarn build:http
cd .. && go build -o hub cmd/http/main.go
```

## Test and lint

```sh
go test ./...                       # Go tests
cd frontend && yarn tsc:compile     # TypeScript
cd frontend && yarn lint            # ESLint + Prettier (husky pre-commit)
```

There are no frontend unit tests; frontend quality is enforced via linting. Backend mocks are generated with mockery (`.mockery.yaml`).

## UI conventions (Alby Hub)

- **shadcn/ui components only.** Never create custom components when a shadcn equivalent exists; never modify core shadcn source (compose or wrap instead).
- **Flat layouts.** No cards nested in cards, no redundant borders. Match sibling spacing.
- **Copy from the user's perspective.** The Hub IS the wallet. Use the product vocabulary (sats, connections, apps). Concise. No em-dashes. "Lightning", not "Bitcoin Lightning". Never quote model counts (the catalog and the routable set drift; verify live via the daemon).
- **Strict TypeScript.** No `any`. `*Icon` suffix for lucide imports. SWR for server state, Zustand for client state.
- **`request()` helper** for all HTTP; remember it does **not** prepend `/api` (see troubleshooting).
- **Metadata writes are read-modify-write** against `/api/v2/apps/:id` (PATCH replaces the whole object).
- New screens go in `frontend/src/routes.tsx`; new endpoints in `api/api.go` with routes in `http/http_service.go`.

## The empirical verification culture

This project verifies changes against the live system, not just the compiler:

1. Build (tsc + go build must pass)
2. Deploy (restart hub, unlock immediately)
3. Verify in a fresh browser session (hard refresh; watch the hub log during clicks)
4. Probe the daemon (`/clients`, `/keys/balance`, `/models`) to confirm server-side state
5. Clean up test artifacts (throwaway apps, daemon clients)

For UI work: study the reference first, one change at a time, keep scope to what was asked.

## Branch and commit conventions

Type prefix + short dash-separated summary: `feat/`, `chore/`, `fix/`. Commit in logical chunks; keep the daemon dist patches out of the tree (see [daemon-patches.md](daemon-patches.md)).

## Running as a service (systemd)

The Hub can run as a supervised service so a crash or reboot self-heals with no manual step (verified live: `systemctl kill -s KILL alby-hub` recovers in ~15s):

```ini
# /etc/systemd/system/alby-hub.service
[Unit]
Description=Alby Hub (Routstr fork)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/hub
Environment=PATH=/root/.bun/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ExecStart=/root/hub/hub
Restart=always
RestartSec=5
TimeoutStopSec=30

[Install]
WantedBy=multi-user.target
```

Two gotchas, both hit in production:

- **PATH must include the bun bin dir.** cocod and routstrd are bun scripts with `#!/usr/bin/env bun`; systemd's default PATH lacks `/root/.bun/bin`, so the daemons die instantly on spawn (silently — the Hub discards their stderr). `Environment=PATH=...` fixes it.
- **Enable auto-unlock** (`PATCH /api/auto-unlock {"unlockPassword": "..."}`) so the Hub opens its wallet on restart without a human. Tradeoff: anyone with config access can start it — do this only with a strong unlock password.

`deploy.sh` at the repo root runs the atomic deploy sequence: build frontend, build binary, `systemctl restart alby-hub`, unlock, wait for daemons, verify. (`UNLOCK_PASSWORD` env var overrides the default.)
