#!/bin/bash
# Provisioner: install docker + caddy, pre-pull Alby Hub image, register systemd unit.
# Runs during Packer build against a fresh Ubuntu 24.04 droplet.
set -euxo pipefail

: "${ALBYHUB_IMAGE:=ghcr.io/getalby/hub:latest}"

apt-get update
apt-get install -y ca-certificates curl gnupg ufw

install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" > /etc/apt/sources.list.d/docker.list

curl -fsSL https://dl.cloudsmith.io/public/caddy/stable/gpg.key | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -fsSL https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt | tee /etc/apt/sources.list.d/caddy-stable.list

apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin caddy

docker pull "$ALBYHUB_IMAGE"

mkdir -p /opt/albyhub/data /etc/albyhub
touch /etc/albyhub/hub.env

ufw --force enable
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 9735/tcp

systemctl enable albyhub.service
systemctl enable caddy
