# Contributing

Thanks for wanting to help. This is a small fork with a specific goal: keep the Routstr integration a gentle, pattern-conforming addition to Alby Hub.

## Ground rules

- Follow the Alby Hub conventions (see [AGENTS.md](AGENTS.md) and [docs/development.md](docs/development.md)): shadcn/ui, flat layouts, strict TypeScript, the `request()` helper, read-modify-write for app metadata.
- One change at a time, small commits (`feat/`, `chore/`, `fix/`).
- Verify against the live system, not just the compiler: build, deploy, browser-verify, probe the daemon.
- Never commit daemon dist patches (see [docs/daemon-patches.md](docs/daemon-patches.md)) or secrets (`.env`, `.data`, configs).

## Setup

```sh
cd frontend && yarn install && yarn build:http   # build:http emits frontend/dist, which cmd/http/main.go embeds
cd .. && go build -o hub cmd/http/main.go
```

## Before opening a PR

- `go test ./...` passes
- `cd frontend && yarn tsc:compile` passes
- The change is documented (README, docs/, or CHANGELOG as appropriate)
- No model counts in copy, no em-dashes, "Lightning" not "Bitcoin Lightning"

## Where things live

- Wizard: `frontend/src/screens/internal-apps/Routstr.tsx`
- Connection page: `frontend/src/components/connections/routstr/`
- Daemon client + money functions: `frontend/src/hooks/useRoutstrd.ts`
- Supervision: `service/routstrd.go`
- Backup: `api/backup.go`
