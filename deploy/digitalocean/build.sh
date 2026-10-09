#!/bin/bash

set -eu

if [ -z "${DIGITALOCEAN_API_TOKEN:-}" ]; then
  echo "DIGITALOCEAN_API_TOKEN is required" >&2
  exit 1
fi

if [ -z "${ALBYHUB_VERSION:-}" ]; then
  echo "ALBYHUB_VERSION is required (example: v1.22.2)" >&2
  exit 1
fi

export DIGITALOCEAN_API_TOKEN
export ALBYHUB_VERSION
export ALBYHUB_DASH_VERSION=$(printf '%s' "$ALBYHUB_VERSION" | sed 's/\./-/g')

packer build template.json
