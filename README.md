# Alby Hub + Routstr

**Self-hosted AI gateway on Lightning.** This is a fork of [Alby Hub](https://github.com/getAlby/hub) in which the Hub supervises **Routstr**: an OpenAI-compatible API that gives you one key, every model, and pay-per-request pricing in sats.

Point any OpenAI-compatible client (your own app, an agent, a CLI tool) at your Hub, and Routstr routes each request to the cheapest provider in a federation of community-run endpoints that has that model. No subscriptions, no signups, no credit card. Your sats, your gateway.


## Why Routstr fits Alby Hub

- **Same ethos.** Self-custody, Lightning-native, no accounts.
- **The Hub is the wallet and controller.** Sats move through Hub primitives only: main wallet to an isolated Routstr app wallet, then to the daemon's Cashu wallet via a mint. The Hub supervises the daemons, runs auto top-up, and backs everything up.
- **One key, every model.** The key is minted by your daemon, works with all models, and the model is chosen per request in your app.
- **A federation, not a vendor.** Routstr discovers providers over Nostr, validates them by pubkey, and routes to the cheapest. Any provider can join.

## What this fork adds over upstream Alby Hub

| Area | Change |
|---|---|
| AI gateway | Routstr connection page, setup wizard, and API keys UI |
| Supervision | Hub starts/stops/restarts the routstrd daemon and cocod Cashu wallet |
| Money rail | Hub-direct funding and refunds via the Routstr app wallet (never the main wallet) |
| Auto top-up | Hub loop refills the Cashu wallet from the Routstr wallet when the balance drops below a threshold |
| Backup | Backups now include the daemon databases and configs (coco.db, routstr.db, bark.sqlite), with verified restore |
| Security | API keys are native to the daemon that minted them; the wallet mint balance is the spend gate |

## Quick start

**Prerequisites:** a machine with Go, Node (yarn), and Bun; a reachable Cashu mint; and a Lightning backend (the Hub supports Bark).

```sh
# 1. Build the frontend and the Hub
cd frontend && yarn install && yarn build:http
cd .. && go build -o hub cmd/http/main.go

# 2. Run it
./hub
# open http://localhost:8080, complete setup, unlock

# 3. Install the daemons (routstrd + cocod) and point the Hub at them

# 4. In the Hub: AI & Agents → Routstr → Connect → follow the wizard
```

The wizard creates an isolated app wallet, mints your API key, and funds it with Cashu ecash. You end up with:

- **Same device:** `http://localhost:8008/v1`
- **External device:** `http://<hub-address>:8080/routstr/v1`

plus your `sk-` key. That is the whole integration surface for any OpenAI-compatible client.

See [docs/deploy.md](docs/deploy.md) for the full install, and [docs/user-flow.md](docs/user-flow.md) for the annotated walkthrough of every screen.

## Documentation

- [User flow](docs/user-flow.md): annotated screenshots of the wizard, connection page, and dialogs, with what happens behind the scenes
- [Architecture](docs/architecture.md): processes, the Nostr provider federation, the money pipeline, and why each decision was made
- [Deploy](docs/deploy.md): install, ports, firewall, mint requirements
- [Backup & restore](docs/backup-restore.md): what is archived and how to restore
- [Daemon patches](docs/daemon-patches.md): the upstream patches carried on routstrd and how to reapply them
- [Troubleshooting](docs/troubleshooting.md): operational knowledge, the hard-won way
- [Development](docs/development.md): build, test, verify
- [Security](docs/security.md): the key model and the deployment hardening rules
- [Reference](docs/reference.md): daemon and Hub API surface, config keys, metadata

## Upstream

This repo tracks [getAlby/hub](https://github.com/getAlby/hub). See [docs/upstream.md](docs/upstream.md) for how merges and the daemon patch policy work.

## License

Apache-2.0, same as upstream Alby Hub.
