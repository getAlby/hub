### Installation on a Raspberry Pi Zero (arm)

This install scripts will help you installing Alby Hub on a Raspberry Pi with Raspberry Pi OS (previously called Raspbian).
You should have some basic Linux understanding to install and operate it.

SSH into your Pi and run:

```shell
/bin/bash -c "$(curl -fsSL https://getalby.com/install/hub/pi-zero-install.sh)"
```

### Updating a running instance

SSH into your Pi and cd into `/opt/albyhub`

Run `./update.sh`
