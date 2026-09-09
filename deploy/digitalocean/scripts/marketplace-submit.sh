#!/bin/bash

set -euo pipefail

IMAGE_ID=$(jq '.builds[-1].artifact_id | split(":")[1] | tonumber' manifest.json)
DIGITALOCEAN_APP_ID=236815239

cat << EOF > update.json
{
  "reasonForUpdate": "new release",
  "version": "${ALBYHUB_VERSION}",
  "imageId": ${IMAGE_ID},
  "softwareIncluded": [
    { "name": "Alby Hub", "version": "${ALBYHUB_VERSION}", "releaseNotes": "https://github.com/getAlby/hub/releases/tag/v${ALBYHUB_VERSION}" },
    { "name": "Docker CE", "version": "29.4.3" },
    { "name": "Docker Compose", "version": "5.1.3" },
    { "name": "Ubuntu", "version": "24.04" }
  ]
}
EOF

cat update.json

echo ${ALBYHUB_VERSION}

echo https://api.digitalocean.com/api/v1/vendor-portal/apps/${DIGITALOCEAN_APP_ID}

curl --fail-with-body -X PATCH -H "Content-Type: application/json" -H "Authorization: Bearer ${DIGITALOCEAN_API_TOKEN}" \
  -d @update.json  https://api.digitalocean.com/api/v1/vendor-portal/apps/${DIGITALOCEAN_APP_ID}

echo "Digital Ocean Market Place update complete"
rm update.json
