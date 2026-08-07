#!/usr/bin/env bash
# Greenlight signer launcher: decrypts the seed for the duration of the
# signer process, then re-encrypts and zeroes it on exit.
#
# Requires: gl-cli (Blockstream greenlight), cryptography python lib.
# Config via env:
#   GL_DATA_DIR         signer data dir (default: ~/.local/share/greenlight)
#   GL_PASS_FILE        passphrase file written by encrypt-seed.py
#   GL_SIGNER_ARGS      args passed to `gl-cli signer` (default: "")
#
# Run as a systemd service (see greenlight-signer.service).

set -euo pipefail

GL_DATA_DIR="${GL_DATA_DIR:-$HOME/.local/share/greenlight}"
GL_PASS_FILE="${GL_PASS_FILE:-$(dirname "$GL_DATA_DIR")/greenlight-signer.pass}"

SEED="$GL_DATA_DIR/hsm_secret"
ENC="$GL_DATA_DIR/hsm_secret.enc"

decrypt_seed() {
    if [ ! -f "$ENC" ] || [ ! -f "$GL_PASS_FILE" ]; then
        echo "error: encrypted seed ($ENC) and passphrase file ($GL_PASS_FILE) required" >&2
        exit 1
    fi
    # decrypt into the data dir with 0600 perms; removed on exit
    python3 - "$ENC" "$SEED" "$GL_PASS_FILE" <<'PY'
import sys
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
import hashlib, os

enc_path, seed_path, pass_file = sys.argv[1], sys.argv[2], sys.argv[3]
with open(enc_path, "rb") as f:
    data = f.read()
with open(pass_file, "rb") as f:
    passphrase = f.read().strip()
salt, nonce, ct = data[:16], data[16:28], data[28:]
key = hashlib.pbkdf2_hmac("sha256", passphrase, salt, 200_000, dklen=32)
seed = AESGCM(key).decrypt(nonce, ct, None)
fd = os.open(seed_path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
os.write(fd, seed)
os.close(fd)
PY
}

zero_seed() {
    if [ -f "$SEED" ]; then
        truncate -s 0 "$SEED" 2>/dev/null || true
    fi
}

trap zero_seed EXIT

decrypt_seed

echo "starting greenlight signer (data dir: $GL_DATA_DIR)"
# shellcheck disable=SC2086
exec gl-cli signer $GL_SIGNER_ARGS
