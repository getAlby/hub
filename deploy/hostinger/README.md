# Deploy Alby Hub on Hostinger VPS

A Hostinger Application Catalog template for Alby Hub. Hostinger wraps a `docker-compose.yml` with a prompt-driven env-var UI and ships it under their VPS Docker templates (see precedents: [openclaw](https://www.hostinger.com/vps/docker/openclaw), [hermes-agent](https://www.hostinger.com/vps/docker/hermes-agent)).

## Artifacts in this directory

- **[docker-compose.yml](./docker-compose.yml)** — two services: Alby Hub (port 8080, lightning p2p port 9735) and an optional Caddy reverse proxy (ports 80/443) activated with the `tls` compose profile. Volume `albyhub-data` persists the SQLite db, seed, channel state, and LDK data.
- **[Caddyfile](./Caddyfile)** — reverse proxy config that terminates TLS for `${DOMAIN}` and proxies to `albyhub:8080`. Caddy auto-provisions a Let's Encrypt cert on startup when a real DNS name is set.

## Deploying without submitting to Hostinger (works today)

Any Hostinger VPS with the Ubuntu 24.04 + Docker template already provisioned can run this:

```bash
curl -O https://raw.githubusercontent.com/getAlby/hub/master/deploy/hostinger/docker-compose.yml
curl -O https://raw.githubusercontent.com/getAlby/hub/master/deploy/hostinger/Caddyfile

# No TLS (IP-only access on port 8080):
docker compose up -d

# With TLS (requires DNS pointed at the VPS):
DOMAIN=albyhub.example.com BASE_URL=https://albyhub.example.com docker compose --profile tls up -d
```

## Submitting to the Hostinger Application Catalog

Hostinger does not publish a public submission process. Reach out to Hostinger BD / partnerships with this package:

| Asset | Source |
|---|---|
| Title | `Alby Hub` |
| Tagline | `Your own bitcoin lightning node` |
| Category | `Finance` / `Self-hosted` |
| Logo (SVG) | [doc/logo.svg](../../doc/logo.svg) |
| Long description | Adapt the top of the main [README.md](../../README.md) |
| `docker-compose.yml` | [./docker-compose.yml](./docker-compose.yml) |
| Reverse proxy config | [./Caddyfile](./Caddyfile) |
| User-facing env-var schema | See table below |
| Minimum VPS plan | KVM 2 (2 GB RAM, 1 shared vCPU, 50 GB NVMe) |
| Post-deploy URL | `http://{public_ip}:8080` (or `https://{domain}` when TLS profile active) |

### Env vars to surface in Hostinger's template UI

| Key | Label | Default | Required |
|---|---|---|---|
| `DOMAIN` | Your domain (optional, for TLS) | *empty* | No |
| `BASE_URL` | Public URL of this Hub (e.g. `https://albyhub.example.com`) | `http://localhost:8080` | No |
| `LDK_ESPLORA_SERVER` | Blockchain API endpoint | `https://electrs.getalbypro.com` | No |

Every other env var from [config/models.go](../../config/models.go) keeps its default; do not surface them in the Hostinger UI or users will be overwhelmed.

## Caveats to surface on the listing page

- **Backup the `albyhub-data` volume regularly.** It holds the encrypted seed, channel state, and SQLite db. Loss = loss of funds.
- **Lightning p2p on port 9735 needs the VPS firewall open.** Hostinger's default firewall blocks it; link users to Hostinger's firewall docs.
- **Self-custody tradeoff.** A VPS-hosted lightning node is less sovereign than a home node-distribution (umbrel, start9). Mention this honestly in the listing so users choose eyes-open.

## Support surface

A new Alby Hub release that changes the container interface must also be smoke-tested against this compose file. Add to the Hub release checklist.
