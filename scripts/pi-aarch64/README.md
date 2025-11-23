### Installation on a Raspberry Pi 4/5 (aarch64)

This install scripts will help you installing Alby Hub on a Raspberry Pi with Raspberry Pi OS (previously called Raspbian).
You should have some basic Linux understanding to install and operate it.

Have a look at our [installation guide](https://github.com/getAlby/hub/tree/master/scripts/pi-arm) for more details and inspiration.

SSH into your Pi and run:

```shell
/bin/bash -c "$(curl -fsSL https://getalby.com/install/hub/pi-aarch64-install.sh)"
```

### Updating a running instance

SSH into your Pi and cd into `/opt/albyhub`

Run `./update.sh`
