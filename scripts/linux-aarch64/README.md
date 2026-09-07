# Alby Hub on a Linux server

> [!WARNING]
> Alby Hub is intended to run on a secure, private network (for example your home network or behind a VPN). Do not expose it directly to the public internet. By default the HTTP server listens on **all network interfaces**, so restrict access with a firewall or bind it to a private address.

## Requirements

- Linux distribution
- Runs pretty much on any VPS/server with 512MB RAM or more (1GB recommended / plus some swap space ideally)
- lightning port 9735 must be available

### Installation (non-Docker)

We have prepared an installation script that installs Alby Hub for you.
We recommend inspecting the install script and if needed adjusting it or taking inspiration from it for your setup.

If you do a fresh server setup make sure to do the basic setup like for example creating a new user and configuring the firewall. Here is a [simple tutorial for this](https://www.digitalocean.com/community/tutorials/initial-server-setup-with-ubuntu).

Run the installation script on your server:

    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/getAlby/hub/master/scripts/linux-aarch64/install.sh)"

The install script will prompt you for an installation folder and will install Alby Hub.
Optionally it can also create a systemd service for you.

You can also do these quite simple steps manually, have a look in the install script for details.

Alby Hub listens on port 8080 (standalone) or port 8029 (when run with a systemd service) on **all network interfaces**, so it is reachable from any machine that can reach the server, for example at `http://<server-ip>:8029`. Restrict access with a firewall so that only your local network or VPN can reach it. The port is configurable using the `PORT` environment variable or by editing `Environment="PORT=8029"` in the albyhub.service systemd config file - See "Editing The Service" below)

Alby Hub is not designed to be exposed on the public internet. If you need remote access, prefer a VPN such as WireGuard or Tailscale. If you nevertheless run it on a public domain, you do so at your own risk: put it behind an HTTPS reverse proxy such as [Caddy](https://caddyserver.com/) and restrict who can reach it (for example by IP allowlist or client certificates).

### Running the services

Either use systemd:

    $ sudo systemctl [start|stop] albyhub.service

Or use the start scripts:

    $ [your install path]/start.sh

### Viewing Logs (systemd)

    $ sudo journalctl -u albyhub

### Editing The Service (systemd)

    $ sudo nano /etc/systemd/system/albyhub.service
    $ sudo systemctl daemon-reload
    $ sudo systemctl restart albyhub.service

### Backup !

Make sure to backup your data directories:

- `[your install path]/data`

### Update

The install script will add an update.sh script to update Alby Hub. It will download the latest version for you.

After the update you will have to unlock Alby Hub again.

### Using Docker

Alby Hub comes as docker image: [ghcr.io/getalby/hub:latest](https://github.com/getAlby/hub/pkgs/container/hub)

    $ docker run -v .albyhub-data:/data -e WORK_DIR='/data' -p 127.0.0.1:8080:8080 ghcr.io/getalby/hub:latest

This publishes the port on `127.0.0.1` only. To reach Alby Hub from other devices on your local network use `-p 8080:8080` and make sure your firewall blocks the port from the internet.

We also provide a simple docker-compose file:

    $ wget https://raw.githubusercontent.com/getAlby/hub/master/docker-compose.yml # <- make sure to update platform
    $ mkdir ./albyhub-data
    $ docker-compose up # or docker-compose up --pull=always <- to make sure you get the latest images

Make sure to mount and backup the data working directory.
