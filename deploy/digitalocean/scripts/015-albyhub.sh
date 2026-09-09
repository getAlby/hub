#!/bin/sh

set -eu

mkdir -p /opt/albyhub/data

sed -i.bak "s|__ALBYHUB_VERSION__|${application_version}|g" /opt/albyhub/docker-compose.yml
rm -f /opt/albyhub/docker-compose.yml.bak

docker pull ghcr.io/getalby/hub:${application_version}
