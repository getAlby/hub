# Caddy Subpath

This is an example of how to run Alby Hub on a subpath using Caddy

To test locally edit `sudo nano /etc/hosts` and add `127.0.0.1 your-domain.com`

Use the following environment variables when building the frontend:

```bash
BASE_PATH="/example-path" yarn build:http
```

Then run Alby Hub as normal. (if default port is not 8080 you will need to update the Caddyfile)

Then start caddy: `sudo caddy run -c ./Caddyfile`

and visit `http://your-domain.com/example-path
