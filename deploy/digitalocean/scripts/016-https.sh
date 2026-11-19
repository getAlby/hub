#!/bin/bash
set -euo pipefail

# Ubuntu's distro Certbot may predate IP/webroot support. Use the documented
# isolated pip installation, with a minimum of 5.4 and no unreviewed major jump.
apt-get -y install nginx python3-venv
python3 -m venv /opt/certbot
/opt/certbot/bin/pip install --upgrade pip
/opt/certbot/bin/pip install 'certbot>=5.4,<6'
/opt/certbot/bin/certbot --help all | grep -- --ip-address
/opt/certbot/bin/certbot --help all | grep -- --required-profile

rm -f /etc/nginx/sites-enabled/default
install -d -m 755 /var/www/acme/.well-known/acme-challenge
nginx -t
systemctl enable nginx
# Do not enable the timer, run Certbot, generate secrets, or start Hub in the
# builder. Customer cloud-init enables the timer on first boot.
systemctl daemon-reload
jq --arg nginx "$(nginx -v 2>&1 | cut -d/ -f2)"    --arg certbot "$(/opt/certbot/bin/certbot --version | cut -d' ' -f2)"    '. + {nginx: $nginx, certbot: $certbot}' /opt/albyhub/software-versions.json    > /opt/albyhub/software-versions.json.new
mv /opt/albyhub/software-versions.json.new /opt/albyhub/software-versions.json
