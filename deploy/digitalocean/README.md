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
export ALBYHUB_VERSION=v1.22.2
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
5. Wait for the droplet to boot.
6. You should now see Alby Hub running at:

```text
http://<droplet-ip>
```

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
export ALBYHUB_VERSION=v1.23.0
```

3. Build the DigitalOcean snapshot:

```bash
cd deploy/digitalocean
./build.sh
```

4. Confirm that `manifest.json` was created and contains the new snapshot artifact.
5. Submit the Marketplace update:

```bash
./scripts/marketplace-submit.sh
```

The script reads the snapshot image ID from `manifest.json` and sends the update request to the DigitalOcean Vendor Portal.

Once this manual flow has been verified during a release, we can move it into `.github/workflows/release.yaml` by adding a job that waits for the GHCR image, builds the snapshot and runs the Marketplace submission script.
