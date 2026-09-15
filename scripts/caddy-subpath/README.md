# Caddy Subpath

This is an example of how to run Alby Hub on a subpath using Caddy

To test locally edit `sudo nano /etc/hosts` and add `127.0.0.1 your-domain.com`

Use the following environment variables when building the frontend:

```bash
BASE_PATH="/example-path" yarn build:http
```

Then run Alby Hub as normal. (if default port is not 8080 you will need to update the Caddyfile)

Then start caddy: `sudo caddy run -c ./Caddyfile`

and visit `http://your-domain.com/example-path`

## Adding a password prompt

To put an extra password prompt in front of Alby Hub, see [`../caddy-gate`](../caddy-gate). Wrap its `handle @gated` / `handle` blocks in the `handle_path /example-path*` block above and keep `Path=/` on the gate cookie - the `__Host-` cookie prefix requires it.

Do not just wrap this example in `basic_auth`: Alby Hub authenticates its own API with an `Authorization: Bearer <JWT>` header, which the frontend sets explicitly, so the browser's Basic credentials never make it to Caddy and every API call is rejected with a 401.
