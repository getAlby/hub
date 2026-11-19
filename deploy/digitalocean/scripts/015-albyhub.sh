#!/bin/sh

set -eu

mkdir -p /opt/albyhub/data

sed -i.bak "s|__ALBYHUB_VERSION__|${application_version}|g" /opt/albyhub/docker-compose.yml
rm -f /opt/albyhub/docker-compose.yml.bak

docker pull ghcr.io/getalby/hub:${application_version}

jq -n --arg hub "$application_version" \
  --arg docker "$(docker version --format '{{.Server.Version}}')" \
  --arg compose "$(docker compose version --short)" \
  --arg ubuntu "$(. /etc/os-release && printf '%s' "$VERSION_ID")" \
  '{hub: $hub, docker: $docker, compose: $compose, ubuntu: $ubuntu}' \
  > /opt/albyhub/software-versions.json
