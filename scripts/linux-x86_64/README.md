# Alby Hub on a Linux server

## Requirements

- Linux distribution
- Runs pretty much on any VPS/server with 512MB RAM or more (1GB recommended / plus some swap space ideally)
- lightning port 9735 must be available

### Installation (non-Docker)

We have prepared an installation script that installs Alby Hub for you.
We recommend inspecting the install script and if needed adjusting it or taking inspiration from it for your setup.

If you do a fresh server setup make sure to do the basic setup like for example creating a new user and configuring the firewall. Here is a [simple tutorial for this](https://www.digitalocean.com/community/tutorials/initial-server-setup-with-ubuntu).


Run the installation script on your server:

    $ /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/getAlby/hub/master/scripts/linux-x86_64/install.sh)"

The install script will prompt you for an installation folder and will install Alby Hub.
Optionally it can also create a systemd service for you.

You can also do these quite simple steps manually, have a look in the install script for details.

Alby Hub will run on localhost:8080 by default (configurable using the `PORT` environment variable or by adding `Environment="PORT=8081"` to the `[Service]` section of the albyhub.service file)

To run on a public domain we recommend the use of a reverse proxy using [Caddy](https://caddyserver.com/)

### Running the services

Either use systemd:

    $ sudo systemctl [start|stop] albyhub.service

Or use the start scripts:

    $ [your install path]/start.sh

### Backup !

Make sure to backup your data directories:

- `[your install path]/data`

### Update

The install script will add an update.sh script to update Alby Hub. It will download the latest version for you.

After the update you will have to unlock Alby Hub again.

### Using Docker

Alby Hub comes as docker image: [ghcr.io/getalby/hub:latest](https://github.com/getAlby/hub/pkgs/container/hub)

    $ docker run -v .albyhub-data:/data -e WORK_DIR='/data' -p 8080:8080 ghcr.io/getalby/hub:latest`

We also provide a simple docker-compose file:

    $ wget https://raw.githubusercontent.com/getAlby/hub/master/docker-compose.yml # <- edit for your needs, but defaults should work well
    $ mkdir ./albyhub-data
    $ docker-compose up # or docker-compose up --pull=always <- to make sure you get the latest images

Make sure to mount and backup the data working directory.
