#!/bin/bash
# Standard DigitalOcean Marketplace cleanup: strip logs, shell history, cloud-init state,
# and package-manager caches before snapshot. Mirrors the upstream sample in
# https://github.com/digitalocean/marketplace-partners/blob/master/scripts/90-cleanup.sh
set -euxo pipefail

apt-get -y autoremove
apt-get -y autoclean

rm -rf /tmp/* /var/tmp/*
rm -rf /var/log/*.gz /var/log/*.[0-9] /var/log/*-????????
: > /var/log/wtmp
: > /var/log/lastlog
: > /var/log/auth.log
: > /var/log/syslog

rm -rf /root/.ssh /root/.bash_history /root/.cache /root/.lesshst
rm -rf /home/*/.ssh /home/*/.bash_history /home/*/.cache /home/*/.lesshst

cloud-init clean --logs
rm -rf /var/lib/cloud/instances/*

truncate -s 0 /etc/machine-id
