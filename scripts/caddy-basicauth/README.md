# Caddy Basic Auth

This is an example of how to run Alby Hub with a second layer of authentication using Caddy.

After a successful Basic Auth login Caddy issues a short-lived `gate` cookie, so the browser does not need to send the Basic Auth credentials on every request.

1. Generate the gate cookie secret. This creates `gate.env` next to the Caddyfile:

   ```bash
   ./generate-secret.sh
   ```

   Re-run it to rotate the secret (existing gate cookies become invalid and users have to log in again).

2. Generate a password hash with `caddy hash-password --plaintext YOUR_PASSWORD` and add it to `gate.env` along with your username:

   ```bash
   BASIC_AUTH_USERNAME=your-username
   BASIC_AUTH_PASSWORD_HASH=$2a$14$...
   ```

3. Copy `Caddyfile.example` to `Caddyfile` and replace `your-domain.com` with your domain. If Alby Hub is not running on the default port 8080 you will need to update the Caddyfile.

4. Run Alby Hub as normal.

5. Start caddy, passing `gate.env`:

   ```bash
   caddy run --config ./Caddyfile --envfile ./gate.env
   ```

6. Visit `https://your-domain.com` and log in with the Basic Auth credentials.

To test locally edit `sudo nano /etc/hosts` and add `127.0.0.1 your-domain.com`, and add `tls internal` to the site block in the Caddyfile.
