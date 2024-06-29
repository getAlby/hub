# Alby Hub <3 phoenixd

Run your Alby Hub with phoenixd as a backend.

## Requirements

+ Linux distribution
+ Runs pretty much on any VPS or server

## Docker

To run Alby Hub with phoenixd use [docker-compose](https://docs.docker.com/compose/) using the [docker-compose.yml file](https://raw.githubusercontent.com/getAlby/nostr-wallet-connect-next/master/scripts/linux-x86_64/phoenixd/docker-compose.yml).

    $ wget https://raw.githubusercontent.com/getAlby/nostr-wallet-connect-next/master/scripts/linux-x86_64/phoenixd/docker-compose.yml
    $ docker-compose up # or docker-compose up --pull=always <- to make sure you get the latest images

It will run on localhost:8080 by default. You can configure the port by editing the docker-compose.yml file.

Note: for simplicity it uses a preconfigired phoenixd password (see docker-compose.yml) this is fine as long as the service is not publicly exposed (change this password if you like).

### Backup

Make sure to backup the `albyhub-phoenixd` which is used as volume for albyhub and phoenixd data files.

## Non Docker

### Installation (non-Docker)

    $ wget https://raw.githubusercontent.com/getAlby/nostr-wallet-connect-next/master/scripts/linux-x86_64/phoenixd/install.sh
    $ ./install.sh

The install script will prompt you for a installation folder and will install phoenixd and Alby Hub there.

Optionally it also creates a systemd services.

It will run on localhost:8080 by default.

### Running the services

Either use systemd:

    $ sudo systemctl [start|stop] phoenixd.service
    $ sudo systemctl [start|stop] albyhub.service

Or us the start scripts:

    $ [your install path]/phoenixd/start.sh
    $ [your install path]/albyhub/start.sh


### Backup

Make sure to backup your data directories:

+ `[your install path]/phoenixd/data`
+ `[your install path]/albyhub/data`
