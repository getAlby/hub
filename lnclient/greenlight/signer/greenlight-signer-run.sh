#!/usr/bin/env bash
#
# Greenlight signer runner — decrypts the encrypted seed on startup,
# launches the signer as a child process (not exec) so the zero_seed
# EXIT trap always runs, and forwards TERM/INT for graceful shutdown.
#
# Required env:  GL_DATA_DIR  (contains hsm_secret.gpg and seed.key)
#                 GREENLIGHT_GLCLI_PATH  (path to glcli)
set -euo pipefail

SEED="$GL_DATA_DIR/hsm_secret"
GL_SIGNER_ARGS="${GL_SIGNER_ARGS:-signer run}"

# Force CD to the signer data dir.
cd "$GL_DATA_DIR"

zero_seed() {
    if [ -f "$SEED" ]; then
        truncate -s 0 "$SEED" 2>/dev/null || true
    fi
}

forward_signal() {
    if [ -n "${SIGNER_PID:-}" ]; then
        kill -TERM "$SIGNER_PID" 2>/dev/null || true
    fi
}

decrypt_seed() {
    if [ ! -f "${GREENLIGHT_GLCLI_PATH:-/usr/local/bin/glcli}" ]; then
        echo "glcli not found at ${GREENLIGHT_GLCLI_PATH:-/usr/local/bin/glcli}" >&2
        exit 1
    fi

    "$GREENLIGHT_GLCLI_PATH" -d "$GL_DATA_DIR" signer decrypt-seed >/dev/null 2>&1
}

trap zero_seed EXIT
trap forward_signal TERM INT

decrypt_seed

echo "starting greenlight signer (data dir: $GL_DATA_DIR)"
# shellcheck disable=SC2086
gl-cli signer ${GL_SIGNER_ARGS:-} &
SIGNER_PID=$!
wait "$SIGNER_PID"
