# Alby Hub <3 phoenixd

Run your Alby Hub with phoenixd as a backend.

> [!WARNING]
> Alby Hub is intended to run on a secure, private network (for example your home network or behind a VPN). Do not expose it directly to the public internet. By default the HTTP server listens on **all network interfaces**, so restrict access with a firewall or bind it to a private address.

## Requirements

- Linux distribution
- Runs pretty much on any VPS or server (512MB+ memory recommended)

## Docker

To run Alby Hub with phoenixd use [docker-compose](https://docs.docker.com/compose/) using the [docker-compose.yml file](https://raw.githubusercontent.com/getAlby/hub/master/scripts/linux-x86_64/phoenixd/docker-compose.yml).

    $ wget https://raw.githubusercontent.com/getAlby/hub/master/scripts/linux-x86_64/phoenixd/docker-compose.yml # <- edit for your needs, but defaults should work well
    $ mkdir -p ./albyhub-phoenixd/phoenixd && mkdir -p ./albyhub-phoenixd/albyhub # <- create the data directories for phoenixd and albyhub. make sure to have backups of this
    $ docker-compose up # or docker-compose up --pull=always <- to make sure you get the latest images

Alby Hub is reachable at http://localhost:8080 on the Docker host only. To reach it from other devices on your local network, change the `ports` entry in docker-compose.yml to `"8080:8080"` and make sure your firewall blocks the port from the internet.

Note: for simplicity it uses a preconfigured phoenixd password (see docker-compose.yml). This is only acceptable because the services are not publicly exposed; change this password if you like.

### Backup

Make sure to backup the `albyhub-phoenixd` which is used as volume for albyhub and phoenixd data files.

## Non Docker

### Installation (non-Docker)

    $ wget https://raw.githubusercontent.com/getAlby/hub/master/scripts/linux-x86_64/phoenixd/install.sh
    $ chmod +x install.sh
    $ ./install.sh

The install script will prompt you for a installation folder and will install phoenixd and Alby Hub there.

Optionally it also creates a systemd services.

Alby Hub listens on port 8080 on **all network interfaces**, so it is reachable from any machine that can reach the server. Restrict access with a firewall so that only your local network or VPN can reach it.

### Running the services

Either use systemd:

    $ sudo systemctl [start|stop] phoenixd.service
    $ sudo systemctl [start|stop] albyhub.service

Or us the start scripts:

    $ [your install path]/phoenixd/start.sh
    $ [your install path]/albyhub/start.sh

### Backup

Make sure to backup your data directories:

- `[your install path]/phoenixd/data`
- `[your install path]/albyhub/data`
