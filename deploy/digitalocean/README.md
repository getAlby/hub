# Deploy Alby Hub on DigitalOcean

This folder builds a DigitalOcean snapshot for Alby Hub using Packer.

## Prerequisites

Install:

- `packer`
- the DigitalOcean Packer plugin

Example install on macOS with Homebrew:

```bash
brew tap hashicorp/tap
brew install hashicorp/tap/packer
packer plugins install github.com/digitalocean/digitalocean
```

You also need:

- a DigitalOcean API token with write access

## Required Environment Variables

Before running the build, set:

- `DIGITALOCEAN_API_TOKEN`
- `ALBYHUB_VERSION`

Example:

```bash
export DIGITALOCEAN_API_TOKEN=dop_v1_...
export ALBYHUB_VERSION=v1.24.8
```

## Build The Snapshot

From this directory, run:

```bash
./build.sh
```

### Note

The build does not produce a local file you upload manually. Instead, it creates a DigitalOcean snapshot in your account.

At the end of a successful build, Packer should print something like:

```text
A snapshot was created: 'albyhub-v1-22-2-snapshot-1778744165'
```

That snapshot is the artifact you use.

## How To Deploy

1. Open the DigitalOcean dashboard.
2. Go to `Create` -> `Droplet`.
3. Find the new snapshot under "Choose an image".
4. Create a droplet from it.
5. Wait a few minutes for first-boot HTTPS provisioning.
6. Open `https://<droplet-public-ip>` directly. No domain or browser SSH tunnel is needed.
7. Authenticate with SSH (`ssh root@<droplet-public-ip>`) or the DigitalOcean
   Droplet console, then run `cat /opt/albyhub/.env` as root. Enter only the value
   after `SETUP_CODE=` in Hub's existing setup-code screen.

The code is generated on the customer's instance before Hub starts, stored with
mode 0600, and preserved across retries/reboots. No setup code, ACME account, or
TLS private key is baked into the snapshot. Use a Hub release with `SETUP_CODE`
support (v1.24.8 is verified); the template requires it but changes no Hub code.
The code protects initial setup and restore until setup is complete. It is not
the unlock password or recovery phrase. Welcome instructions live here, in the
SSH banner, and in the proposed [Marketplace instructions](marketplace-getting-started.md),
not in Hub's UI. Do not publish the code or put it in a URL.

### HTTPS and renewal

Nginx terminates TLS on port 443 and proxies to Hub on `127.0.0.1:8080`.
Port 80 serves **only** `/.well-known/acme-challenge/`; all other HTTP paths
return 404, not the app or a redirect. Allow inbound TCP 80/443 in any DigitalOcean
Cloud Firewall as well as UFW. Keep port 8080 closed; do not alter the loopback binding.
Outbound access to metadata, Let's Encrypt, and the Internet is needed.

The instance obtains a Let's Encrypt IP certificate using Certbot's webroot
plugin and the required `shortlived` profile. These certificates last six days;
Certbot 5.4+ is required for IP/webroot support. The build installs `certbot>=5.4,<6`
in `/opt/certbot`, rather than assuming Ubuntu's packaged version supports IPs.
This is Certbot's documented, **best-effort-supported pip/venv** installation
(snap is their recommended default); this image supplies its own systemd timer.
No domain, DNS update, or nginx Certbot plugin is used. First issuance accepts
Let's Encrypt's subscriber agreement and registers without an email address.

`albyhub-https.timer` runs after boot and hourly (with up to five minutes of jitter).
It retries initial issuance, calls `certbot renew` for the instance's certificate,
and validates/reloads Nginx on success, including recovery from failed reloads.
It preserves a still-valid certificate during a transient renewal failure and
removes an expired/wrong-IP TLS configuration on the next successful metadata
lookup. Browsers reject an expired certificate even before that check.
There is **no public plaintext Hub fallback**, no self-signed fallback, and no
instruction to bypass TLS warnings. If issuance fails, HTTPS stays unavailable.
Persistent failures require operator attention well before the six-day expiry:

```bash
journalctl -u albyhub-https.service -n 50 --no-pager
systemctl list-timers albyhub-https.timer
systemctl start albyhub-https.service
```

Monitor the service's failures and certificate expiry externally. Maintain the
pip-installed client monthly (certificate renewal is automatic, client upgrades
are not), checking upstream release notes before moving to a new major:

```bash
systemctl stop albyhub-https.timer albyhub-https.service
/opt/certbot/bin/pip install --upgrade pip
/opt/certbot/bin/pip install --upgrade 'certbot>=5.4,<6'
systemctl start albyhub-https.timer albyhub-https.service
/opt/certbot/bin/certbot renew --dry-run
```

Primary references: [Let's Encrypt IP certificates and Certbot requirements](https://letsencrypt.org/2026/03/11/shorter-certs-certbot/),
[Certbot installation support](https://eff-certbot.readthedocs.io/en/stable/install.html),
[Certbot renewal/profile options](https://eff-certbot.readthedocs.io/en/stable/man/certbot.html).

**Review boundary:** this HTTPS revision has local validation only. Before any
publication, separately authorize a new build and test real issuance, renewal
(`renew --dry-run` plus successful certificate reload), reboot persistence,
setup-code rejection/acceptance, and network/ACME failure recovery on a Droplet.
The previously tested snapshot `255374536` is still tunnel-only; these changes do
not modify it and must not be advertised as already present there.

## Updating

The DigitalOcean Marketplace update is currently run manually instead of using the GitHub release workflow.

This keeps the release workflow simpler. Otherwise, we would need to wait for the GHCR Docker image for the release tag to become available before building the snapshot, because the Packer build pulls:

```text
ghcr.io/getalby/hub:<version>
```

For each new release:

1. Wait until the Docker image for the release tag is available in GHCR.
2. Export the required environment variables:

```bash
export DIGITALOCEAN_API_TOKEN=dop_v1_...
export ALBYHUB_VERSION=v1.24.8
```

3. Build the DigitalOcean snapshot:

```bash
cd deploy/digitalocean
./build.sh
```

4. Confirm that `manifest.json` contains the new snapshot artifact and
   `software-versions.json` records the bundled versions. Boot-test a new Droplet
   from the snapshot before submitting it, then remove that temporary Droplet.
5. Before submitting, update the Marketplace listing's Getting Started instructions
   in the Vendor Portal to match the HTTPS and setup-code flow above (see the proposed
   `marketplace-getting-started.md`). The
   image-update API does not update those instructions. Confirm that the Vendor
   Portal app ID (not the numeric image ID) matches Alby Hub.
6. Submit the Marketplace update:

```bash
./scripts/marketplace-submit.sh
```

The script reads the snapshot image ID from `manifest.json` and sends the update request to the DigitalOcean Vendor Portal.
An accepted submission is not proof of Marketplace publication; verify the
submitted image and review state in the Vendor Portal.

Once this manual flow has been verified during a release, we can move it into `.github/workflows/release.yaml` by adding a job that waits for the GHCR image, builds the snapshot and runs the Marketplace submission script.
