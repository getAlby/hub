# Deploy Alby Hub on DigitalOcean (Marketplace 1-Click App)

A submission-ready scaffold for publishing Alby Hub as a DigitalOcean Marketplace 1-Click Droplet App. Follows the layout documented at [digitalocean/marketplace-partners](https://github.com/digitalocean/marketplace-partners).

## What this scaffold produces

A Droplet image (Ubuntu 24.04) with:
- docker-ce pre-installed
- `ghcr.io/getalby/hub:latest` pre-pulled
- caddy 2 for TLS termination (Let's Encrypt)
- A `systemd` unit that runs Alby Hub at boot with a persistent `/opt/albyhub/data` volume
- A first-boot cloud-init script that prompts for `DOMAIN` via droplet user-data or auto-configures IP-only access
- A first-login MOTD with the Hub URL and onboarding steps

## Directory layout

```
deploy/digitalocean/
├── README.md                               # this file
├── marketplace-image.json                  # Packer template (input to `packer build`)
├── scripts/
│   ├── 01-install-albyhub.sh               # installs docker, caddy, pulls image, drops systemd unit
│   ├── 90-cleanup.sh                       # DigitalOcean-required log/history cleanup
│   └── 99-img_check.sh                     # DigitalOcean-required image validation
└── files/
    ├── etc/
    │   ├── systemd/system/albyhub.service  # systemd unit for the container
    │   ├── caddy/Caddyfile                 # caddy reverse-proxy config
    │   └── update-motd.d/99-albyhub        # first-login message
    └── var/lib/cloud/scripts/per-instance/
        └── 01-albyhub-firstboot.sh         # first-boot config (runs once on each new droplet)
```

## Building the image

Prerequisites: `packer` ≥ 1.9, a DigitalOcean API token with `write` scope.

```bash
export DIGITALOCEAN_API_TOKEN=dop_v1_...
packer init deploy/digitalocean/marketplace-image.json
packer build deploy/digitalocean/marketplace-image.json
```

Packer creates a snapshot in your DO account. To test: launch a droplet from that snapshot, wait ~60 seconds, SSH in, read the MOTD, open the URL.

## Submitting to the Marketplace

1. Apply at [digitalocean.com/landing/partner-application](https://www.digitalocean.com/landing/partner-application) (Partner Pod program).
2. Fork [digitalocean/marketplace-partners](https://github.com/digitalocean/marketplace-partners) and open a PR that references the snapshot ID.
3. Provide marketing assets:
   - Logo: [doc/logo.svg](../../doc/logo.svg)
   - Category: `Blockchain` (they have a dedicated bitcoin subcategory)
   - Short description: "Your own bitcoin lightning node — connectable, feature-rich, self-sovereign."
   - Long description: adapt the top of the main [README.md](../../README.md).
   - Recommended droplet size: **Basic Regular 2 GB RAM / 1 vCPU / 50 GB SSD** ($12/mo at the time of writing). Call out that 1 GB is not enough for the embedded LDK node.
4. Iterate on DO's validation feedback (their reviewers will run `99-img_check.sh` and manual checks).

## Droplet user-data options

Users can pre-configure the deploy by passing cloud-init user-data when creating the droplet:

```yaml
#cloud-config
write_files:
  - path: /etc/albyhub/hub.env
    content: |
      DOMAIN=albyhub.example.com
      BASE_URL=https://albyhub.example.com
```

If `DOMAIN` is set, Caddy requests a Let's Encrypt cert automatically on first boot. If not, the droplet serves Hub on `http://<public-ip>:8080` — the MOTD walks users through setting a domain afterwards.

## Affiliate attribution

Add `?refcode=albyhub` (or whatever Alby's DO affiliate code is at submission time) to every "Deploy on DigitalOcean" link in marketing materials. Per DO's affiliate terms: 10%/mo for 12 months on referred spend + $25 per paying referral after first $25 spend.

## Caveats to surface on the listing page

- **Backup `/opt/albyhub/data` regularly.** Holds the encrypted seed, channel state, and SQLite db. Loss = loss of funds.
- **Lightning p2p port 9735 is opened in the droplet firewall by the provisioner.** If the user enables DigitalOcean's Cloud Firewall on top, they must allow 9735 there too.
- **Self-custody tradeoff.** Same wording as the Hostinger listing — a VPS-hosted lightning node is less sovereign than a home node-distribution.

## Support surface

A new Alby Hub release that changes the container interface must trigger a new Packer build + snapshot + marketplace update. Add to the Hub release checklist. DO supports "update image in place" via a new snapshot push — existing users don't auto-upgrade, so release notes must tell them to `docker compose pull && docker compose up -d` (or use the MOTD update hint).
