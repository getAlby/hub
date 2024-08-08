### Installation on a Raspberry Pi Zero

Have a look at our [installation guide](https://guides.getalby.com/user-guide/v/alby-account-and-browser-extension/alby-hub/alby-hub-flavors/raspberry-pi-zero) for more details.

```shell
 $ ssh albyhub@albyhub.local '/bin/bash -c "$(curl -fsSL https://getalby.com/install/hub/pi-zero-install.sh)"'
 ```

or on the Pi directly:
```shell
/bin/bash -c "$(curl -fsSL https://getalby.com/install/hub/pi-zero-install.sh)"
```

### Updating a running instance

```shell
 $ ssh albyhub@albyhub.local '/bin/bash -c "$(curl -fsSL https://getalby.com/install/hub/pi-zero-install.sh)"'
 ```

or on the Pi directly:
```shell
/bin/bash -c "$(curl -fsSL https://getalby.com/install/hub/pi-zero-update.sh)"
```

And see install.sh and update.sh for details.
