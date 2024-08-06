# Alby Hub on a Linux server

## Requirements

- Linux distribution
- Runs pretty much on any VPS/server with 512MB RAM or more (+some swap space ideally)

### Installation (non-Docker)

We have prepared an installation script that installs Alby Hub for you.
We recommend inspecting the install script and if needed adjusting it or taking inspiration from it for your setup.

As a general good practice we recommend creating a new system user.

    $ adduser albyhub

Run the installation script on your server:

    $ wget https://raw.githubusercontent.com/getAlby/hub/master/scripts/linux-x86_64/install.sh
    $ ./install.sh

Or directly through SSH:

    $ ssh albyhub@[YOUR IP] '/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/getAlby/hub/master/scripts/linux-x86_64/install.sh)"'


The install script will prompt you for an installation folder and will install Alby Hub.

Optionally it can also create a systemd service for you.

Alby Hub will run on localhost:8080 by default (configurable using the `PORT` environment variable)
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
