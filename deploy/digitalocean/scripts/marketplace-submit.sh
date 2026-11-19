#!/bin/bash

set -euo pipefail

: "${DIGITALOCEAN_API_TOKEN:?DIGITALOCEAN_API_TOKEN is required}"
: "${ALBYHUB_VERSION:?ALBYHUB_VERSION is required}"
[[ "$ALBYHUB_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "Expected a release tag such as v1.24.8" >&2; exit 1; }
IMAGE_ID=$(jq -er '.builds[-1].artifact_id | capture("^[^:]+:(?<id>[0-9]+)$").id | tonumber' manifest.json)
# Vendor Portal app ID, not the numeric Droplet image ID in createApp URLs.
DIGITALOCEAN_APP_ID=095e202e5db059c55e506b81
jq -e --arg version "$ALBYHUB_VERSION" '.hub == $version and (.docker | length > 0) and (.compose | length > 0) and (.ubuntu | length > 0)' software-versions.json > /dev/null

jq --arg version "$ALBYHUB_VERSION" --argjson image "$IMAGE_ID" '{
  reasonForUpdate: "new release",
  version: $version,
  imageId: $image,
  osVersion: ("Ubuntu " + .ubuntu),
  softwareIncluded: [
    {name: "Alby Hub", version: $version, releaseNotes: ("https://github.com/getAlby/hub/releases/tag/" + $version)},
    {name: "Docker CE", version: .docker},
    {name: "Docker Compose", version: .compose},
    {name: "Ubuntu", version: .ubuntu}
  ]
}' software-versions.json > update.json

cat update.json

echo ${ALBYHUB_VERSION}

echo https://api.digitalocean.com/api/v1/vendor-portal/apps/${DIGITALOCEAN_APP_ID}

if curl --silent --show-error --fail-with-body -X PATCH -H "Content-Type: application/json" -H "Authorization: Bearer ${DIGITALOCEAN_API_TOKEN}" \
  -d @update.json "https://api.digitalocean.com/api/v1/vendor-portal/apps/${DIGITALOCEAN_APP_ID}" \
  -o marketplace-response.json; then
  echo "Marketplace submission accepted; inspect marketplace-response.json and verify the Vendor Portal review state."
else
  echo "Marketplace submission failed; inspect marketplace-response.json." >&2
  exit 1
fi
