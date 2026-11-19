#!/bin/bash
set -euo pipefail
umask 077

# One service/timer owns provisioning and renewal. Also serialize manual runs.
exec 9>/run/albyhub-https.lock
flock -n 9 || exit 0
cd /opt/albyhub
mkdir -p data
if [ ! -e .env ]; then
  code=$(openssl rand -hex 32)
  printf 'SETUP_CODE=%s\n' "$code" > .env.new
  mv .env.new .env
fi
chmod 600 .env
# Refuse damaged/empty state instead of silently changing an existing code.
grep -Eq '^SETUP_CODE=[0-9a-f]{64}$' .env
docker compose up -d --pull never

ip=$(curl --noproxy '*' --fail --silent --show-error --connect-timeout 3 --max-time 10 \
  http://169.254.169.254/metadata/v1/interfaces/public/0/ipv4/address)
python3 -c 'import ipaddress, sys; a = ipaddress.IPv4Address(sys.argv[1]); assert a.is_global' "$ip"
cert_name="albyhub-$ip"
cert_dir="/etc/letsencrypt/live/$cert_name"
tls_config=/etc/nginx/conf.d/albyhub-https.conf

certificate_valid() {
  openssl x509 -in "$cert_dir/fullchain.pem" -noout -checkend 0 >/dev/null 2>&1 &&
    openssl x509 -in "$cert_dir/fullchain.pem" -noout -checkip "$ip" >/dev/null 2>&1
}

# Remove a stale/expired listener before retrying issuance. Port 80 remains
# challenge-only, never a fallback proxy for Hub (including on first boot).
if ! certificate_valid; then
  rm -f "$tls_config"
  nginx -t
  systemctl reload-or-restart nginx
fi
systemctl start nginx

if [ -f "/etc/letsencrypt/renewal/$cert_name.conf" ]; then
  /opt/certbot/bin/certbot renew --non-interactive --cert-name "$cert_name"
else
  /opt/certbot/bin/certbot certonly --non-interactive --agree-tos \
    --register-unsafely-without-email --cert-name "$cert_name" \
    --server https://acme-v02.api.letsencrypt.org/directory \
    --required-profile shortlived --webroot --webroot-path /var/www/acme \
    --ip-address "$ip"
fi
certificate_valid

# Render outside conf.d so nginx never sees a partially written configuration.
sed "s|__CERT_NAME__|$cert_name|g" /opt/albyhub/nginx-https.conf.template > /etc/nginx/albyhub-https.conf.new
chmod 644 /etc/nginx/albyhub-https.conf.new
mv /etc/nginx/albyhub-https.conf.new "$tls_config"
nginx -t
# Reload on every successful reconciliation: a prior reload failure is retried
# even if Certbot reports that the certificate is not due for renewal yet.
systemctl reload-or-restart nginx
