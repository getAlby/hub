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
