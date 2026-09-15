# Caddy password gate for Alby Hub

An extra password prompt on the web server level, in front of Alby Hub.

> [!WARNING]
> This is defence in depth, not a green light to put Alby Hub on the public
> internet. Alby Hub is meant to run on a secure, private network. If you need
> remote access, a VPN such as WireGuard or Tailscale is still the safer
> option. The gate is useful as an extra layer *on top of* a VPN, a firewall or
> an IP allowlist, and as a way to keep casual scanners away from the login
> screen.

## Why not just `basic_auth`?

Alby Hub authenticates its own API with a JWT that the frontend sends as
`Authorization: Bearer <token>` on every request. That header is set explicitly
by the frontend, so it replaces the `Authorization: Basic ...` header the
browser attaches after an HTTP Basic login.

The result: with a plain `basic_auth` in front of Alby Hub, page loads work but
every API call arrives at Caddy without Basic credentials, Caddy answers `401`,
and the web UI is broken. Passing the Basic header through instead does not
help either, because then Alby Hub never receives its own Bearer token.

Only one `Authorization` header exists per request, and both layers want it.

## How the gate works

Basic Auth is used only for the *first* request. On success Caddy hands out a
random, `HttpOnly` cookie, and every later request is judged on that cookie
instead - which leaves the `Authorization` header free for Alby Hub.

```
browser --(no cookie)-------------------------> Caddy --(401 Basic)--> browser
browser --(username + password)---------------> Caddy --(Set-Cookie)-> Alby Hub
browser --(cookie + Authorization: Bearer ...)-> Caddy ---------------> Alby Hub
```

Details worth knowing:

- **The cookie secret lives in the environment**, not in the Caddyfile, so it
  can be rotated without touching the config. Caddy reads it as
  `{env.GATE_SECRET}`.
- **Two secrets are accepted**, the current one and the previous one, so
  rotating does not sign everybody out at once.
- **The cookie is refreshed on every request** (sliding expiry, 7 days by
  default). Active users are never prompted again, and a cookie still holding
  the previous secret is silently upgraded to the current one.
- **Neither the Basic credentials nor the gate cookie reach Alby Hub.** Caddy
  strips them. Alby Hub's own `Authorization: Bearer` token is left untouched.
- **HTTPS is required.** The cookie is `Secure` and uses the `__Host-` prefix,
  so browsers refuse to store it over plain HTTP, and a sibling subdomain
  cannot overwrite it. Caddy gets a certificate automatically once you point a
  domain at the server.

## Files in this folder

| File | Goes to | Purpose |
|------|---------|---------|
| `Caddyfile.example` | `/etc/caddy/Caddyfile` | The reverse proxy + gate config |
| `gate.env.example` | `/etc/caddy/gate.env` | The gate secrets (chmod 600) |
| `caddy-gate.conf` | `/etc/systemd/system/caddy.service.d/` | Loads `gate.env` into Caddy's environment |
| `rotate-gate-secret.sh` | anywhere, e.g. `/usr/local/bin/` | Rotates the secret and restarts Caddy |

## Setup

The Alby Hub [install script](../linux-x86_64/README.md) can do all of this for
you - it asks whether you want the gate. These are the manual steps.

Requires **Caddy v2.8 or newer** (`basic_auth` was called `basicauth` before)
and a domain name pointing at your server.

### 1. Install Caddy

See the [Caddy install docs](https://caddyserver.com/docs/install). On Debian
or Ubuntu the official apt repository is the usual route.

### 2. Generate the gate secrets

```bash
sudo mkdir -p /etc/systemd/system/caddy.service.d
sudo cp caddy-gate.conf /etc/systemd/system/caddy.service.d/
sudo systemctl daemon-reload

sudo cp rotate-gate-secret.sh /usr/local/bin/albyhub-rotate-gate-secret.sh
sudo chmod +x /usr/local/bin/albyhub-rotate-gate-secret.sh
sudo /usr/local/bin/albyhub-rotate-gate-secret.sh
```

This writes `/etc/caddy/gate.env` with `GATE_SECRET` and `GATE_SECRET_OLD`,
readable by root only.

### 3. Choose a password

```bash
caddy hash-password
```

Enter a long, random password - Caddy has no rate limiting, so the password is
the only thing standing between a scanner and your hub. Copy the printed
`$2a$14$...` hash.

### 4. Write the Caddyfile

Copy `Caddyfile.example` to `/etc/caddy/Caddyfile`, then replace:

- `your-domain.com` with your domain
- the `bob $2a$14$...` line with your username and the hash from step 3
- `127.0.0.1:8029` with the port Alby Hub listens on (`8029` when it runs as a
  systemd service created by the install script, `8080` when started with
  `start.sh`)

### 5. Validate and start

```bash
sudo caddy validate --config /etc/caddy/Caddyfile --envfile /etc/caddy/gate.env
sudo systemctl restart caddy
```

### 6. Close the direct port

The gate is pointless if Alby Hub is still reachable on port 8029 directly.
Restrict it with your firewall, for example:

```bash
sudo ufw allow 80,443/tcp
sudo ufw deny 8029/tcp
```

## Rotating the secret

```bash
sudo /usr/local/bin/albyhub-rotate-gate-secret.sh
```

The current secret becomes `GATE_SECRET_OLD`, a fresh one becomes
`GATE_SECRET`, and Caddy is restarted. Anyone who used Alby Hub since the
previous rotation stays signed in; anyone who did not gets the password prompt
again.

Rotate after someone leaves, after a device is lost or stolen, or on a
schedule. For monthly rotation, link it into cron:

```bash
sudo ln -s /usr/local/bin/albyhub-rotate-gate-secret.sh /etc/cron.monthly/albyhub-rotate-gate-secret
```

(`run-parts` ignores file names containing a dot, hence the missing `.sh`.)

Rotating does **not** change the Basic Auth password - edit the hash in the
Caddyfile for that.

## Options

**Shorter or longer sessions.** Change `Max-Age=604800` (7 days) in both
`Set-Cookie` lines. Because the cookie is refreshed on every request, this is
an *idle* timeout, not an absolute one.

**Bind the cookie to your network.** Add a `remote_ip` matcher so a stolen
cookie is useless from anywhere else. Matchers inside a named matcher block are
AND-ed:

```caddy
@gated {
	expression `{http.request.cookie.__Host-albyhub_gate} != "" && ({http.request.cookie.__Host-albyhub_gate} == {env.GATE_SECRET} || {http.request.cookie.__Host-albyhub_gate} == {env.GATE_SECRET_OLD})`
	remote_ip 192.168.0.0/16
}
```

**Running on a subpath.** See [`../caddy-subpath`](../caddy-subpath). Keep
`Path=/` on the cookie - the `__Host-` prefix requires it.

**Custom Alby OAuth client.** If you replaced the default
`ALBY_OAUTH_CLIENT_ID`, set `BASE_URL=https://your-domain.com` for Alby Hub so
the OAuth redirect points at your domain. The default client id uses a
copy-and-paste auth code flow and needs no configuration.

## Troubleshooting

**The browser asks for the password on every page load.** The cookie is not
being stored. It is `Secure` + `__Host-`, so the site must be served over
HTTPS with a valid certificate, and `Path=/` must be present on the
`Set-Cookie` line.

**The UI loads but every action fails with a 401.** The gate cookie expired
while the tab was open. Reload the page - the reload is a normal navigation,
so the browser can answer the Basic Auth challenge and pick up a fresh cookie.
If it happens regularly, raise `Max-Age`.

**Everything gets through without a password.** `gate.env` is probably not
loaded. Check that the drop-in is in place and that you restarted (not
reloaded) Caddy:

```bash
systemctl show caddy -p EnvironmentFiles
sudo systemctl restart caddy
```

The `!= ""` guard in the `@gated` expression is what keeps a missing
`gate.env` from turning into an open door, so do not remove it.

**`caddy validate` complains about `basic_auth`.** You are on Caddy older than
v2.8, where the directive was called `basicauth`. Upgrade Caddy.
