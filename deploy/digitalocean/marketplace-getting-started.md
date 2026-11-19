# Proposed Marketplace Getting Started instructions

Review copy only: do not publish until a newly built HTTPS image passes live
issuance, renewal, reboot, and onboarding tests. The image-update API does not
update this text. This does not describe the older tunnel-only snapshot 255374536.

## Connect to your Alby Hub

1. Create your Droplet and allow a few minutes for HTTPS provisioning. If you use
   a DigitalOcean Cloud Firewall, allow inbound TCP 80 (certificate validation),
   443 (HTTPS), and SSH as appropriate. Do not expose port 8080.
2. Open `https://<your-droplet-public-ip>` in your browser. No domain or SSH tunnel
   is needed. Never use public HTTP or bypass a certificate warning.
3. Connect with `ssh root@<your-droplet-public-ip>` using your Droplet credentials,
   or authenticate through the DigitalOcean Droplet console. As root, run:

   ```bash
   cat /opt/albyhub/.env
   ```

4. Copy only the value after `SETUP_CODE=` into Hub's setup-code screen, then
   follow Hub's existing setup or restore flow. Keep the code private; do not
   paste it into URLs, support tickets, or screenshots.

Each Droplet generates its own code before Hub starts. It protects initial setup
and restore until setup is complete; it is not your unlock password or recovery
phrase. Keep your recovery information safe separately.

If HTTPS is not ready, the Droplet retries hourly. Through authenticated SSH or
the Droplet console, inspect `journalctl -u albyhub-https.service -n 50 --no-pager`
and retry with `systemctl start albyhub-https.service`. Do not work around a
failure by exposing Hub over HTTP. Port 80 is for ACME challenges only.

Certificates renew automatically. Monitor HTTPS/service health and maintain the
Certbot client as described in the deployment README. Your Hub configuration and
data live in `/opt/albyhub` and `/opt/albyhub/data`.
