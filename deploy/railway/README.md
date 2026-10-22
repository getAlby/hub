# Deploy Alby Hub on Railway

A "Deploy on Railway" one-click template for Alby Hub. Railway templates are created in Railway's dashboard and shared as a `https://railway.com/template/<id>` URL — there is no repo-committed template file. This document defines the exact config to enter in the dashboard so the template is reproducible.

## Template configuration

Create a new service in a Railway template with the following settings.

### Source
- **Type:** Docker image
- **Image:** `ghcr.io/getalby/hub:latest`

### Volumes
- **Mount path:** `/data`
- **Size:** 1 GB (users on larger wallets should resize after deploy)

### Networking
- **Public HTTP port:** `8080` (Railway will assign a `*.up.railway.app` domain and terminate TLS)
- **Lightning p2p port 9735:** *not exposed.* Railway's HTTP proxy does not forward arbitrary TCP, so inbound channel opens are not reachable by default. Users who need inbound p2p should deploy on a VPS provider (DigitalOcean, Hostinger) instead. Document this clearly on the template's Railway listing.

### Environment variables
| Key | Value | Notes |
|---|---|---|
| `WORK_DIR` | `/data/albyhub` | Matches `render.yaml`. Data persists on the `/data` volume. |
| `BASE_URL` | `${{RAILWAY_PUBLIC_DOMAIN}}` | Prefix with `https://` if Railway's variable doesn't include it. Needed for OAuth redirects and copy-to-clipboard URLs. |
| `LDK_ESPLORA_SERVER` | `https://electrs.getalbypro.com` | Same default as `render.yaml`. Override for self-hosted esplora. |
| `PORT` | *(leave unset)* | Defaults to 8080, matches Railway's exposed port. |
| `AUTO_UNLOCK_PASSWORD` | *(leave unset)* | Users can opt in after first-run for auto-start on reboot. Trades custody for convenience — do not pre-fill. |

All other env vars from [config/models.go](../../config/models.go) use their defaults. Surface only `LDK_ESPLORA_SERVER` and `AUTO_UNLOCK_PASSWORD` as user-editable in the Railway template UI; keep `WORK_DIR` and `BASE_URL` locked.

## README button (paste once template is live)

Once the template URL is known, add this to the main [README.md](../../README.md) next to the existing Render button:

```markdown
[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/template/<TEMPLATE_ID>?utm_campaign=alby-hub)
```

Replace `utm_campaign=alby-hub` consistently with the other marketplace buttons so conversion can be attributed.

## Economics

Railway's [template kickback program](https://blog.railway.com/p/template-kickback-program-cash) returns 25% of the first 12 months of infrastructure spend from each deploy to the template owner. The receiving account should be owned by Alby (not a personal account) so the revenue flows through the business.

## Support surface

Each template deploy is a live Alby Hub instance that expects `ghcr.io/getalby/hub:latest` to keep working. Before a breaking image change lands, verify that a fresh Railway deploy still completes onboarding. Add to the Hub release checklist.