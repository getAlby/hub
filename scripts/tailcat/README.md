# Tailcat

This is an example of how to access Alby Hub running on a remote machine (for example a VPS) without exposing it publicly, using [tailcat](https://github.com/tailscale/tailcat).

Tailcat forwards a port from the remote machine to your laptop through an encrypted WireGuard tunnel. No account is needed, and no ports have to be opened on the remote machine, so Alby Hub's port can stay blocked by your firewall.

The steps below assume Alby Hub is listening on port 8029 (the default when run with the systemd service). If you use a different port (for example 8080) replace `8029` accordingly.

## Laptop: generate a key

1. [Install tailcat](https://github.com/tailscale/tailcat/blob/main/INSTALL.md).

2. Generate a client key. This prints the public key of your laptop (`nodekey:...`), which is needed on the remote machine:

   ```bash
   tailcat genkey --client --key=mylaptopkey
   ```

## Remote machine / VPS: serve Alby Hub through Tailcat

1. [Install tailcat](https://github.com/tailscale/tailcat/blob/main/INSTALL.md).

2. Generate a server key. It is saved to disk, so the tailcat address stays the same across restarts:

   ```bash
   tailcat genkey --key=myremotekey --fixed-region
   ```

3. Run Alby Hub as normal.

4. Serve the Alby Hub port, only allowing your laptop to connect. Replace `nodekey:...` with the public key printed on your laptop:

   ```bash
   tailcat serve --key=myremotekey --allow=nodekey:... 8029
   ```

   This prints the tailcat address of the remote machine (`tcXXXXXXXXX`).

## Laptop: connect

1. Forward the remote port to your laptop, replacing `tcXXXXXXXXX` with the tailcat address printed on the remote machine:

   ```bash
   tailcat forward --key=mylaptopkey tcXXXXXXXXX 8029:8029
   ```

2. Visit `http://localhost:8029`.

## Notes

- Always use `--allow`. Without it, anyone who knows the tailcat address can connect. To allow multiple devices, pass a comma-separated list of public keys.
- `tailcat serve` needs to keep running on the remote machine. To keep it running after you log out or reboot, run it as a systemd service, similar to `albyhub.service`.
- The key names are up to you. If you name them `default` (remote machine) and `client-default` (laptop), tailcat picks them up automatically and the `--key` flags can be omitted from `serve` and `forward`.
