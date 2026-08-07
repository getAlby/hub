#!/usr/bin/env python3
"""
Encrypt the Greenlight signer seed (hsm_secret) at rest.

Greenlight's signer stores the seed as PLAINTEXT `hsm_secret` in its data
dir (~/.local/share/greenlight). This script replaces the plaintext file
with an AES-256-GCM encrypted version (hsm_secret.enc) and stores the
randomly generated passphrase in a separate 0600 file.

Usage:
    encrypt-seed.py <data_dir> [--passphrase-file <path>]

    <data_dir>          signer data dir containing hsm_secret
    --passphrase-file   where to write the passphrase (default:
                        <data_dir>/../greenlight-signer.pass)

The wrapper greenlight-signer-run.sh decrypts on start and re-encrypts on
stop, so the plaintext seed only exists while the signer process runs.

If hsm_secret.enc already exists and hsm_secret is missing, the seed is
already encrypted; run with --verify to check the passphrase decrypts it.
"""
import argparse
import hashlib
import os
import secrets
import sys


def decrypt(data: bytes, passphrase: bytes) -> bytes:
    import hashlib as _h

    salt = data[:16]
    nonce = data[16:28]
    key = _h.pbkdf2_hmac("sha256", passphrase, salt, 200_000, dklen=32)
    from cryptography.hazmat.primitives.ciphers.aead import AESGCM

    aesgcm = AESGCM(key)
    return aesgcm.decrypt(nonce, data[28:], None)


def encrypt(plaintext: bytes, passphrase: bytes) -> bytes:
    import hashlib as _h

    salt = secrets.token_bytes(16)
    nonce = secrets.token_bytes(12)
    key = _h.pbkdf2_hmac("sha256", passphrase, salt, 200_000, dklen=32)
    from cryptography.hazmat.primitives.ciphers.aead import AESGCM

    aesgcm = AESGCM(key)
    return salt + nonce + aesgcm.encrypt(nonce, plaintext, None)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("data_dir", help="signer data dir containing hsm_secret")
    parser.add_argument("--passphrase-file", default=None)
    parser.add_argument("--verify", action="store_true", help="only verify the encrypted seed decrypts")
    parser.add_argument("--force", action="store_true", help="re-encrypt even if hsm_secret.enc exists")
    args = parser.parse_args()

    seed_path = os.path.join(args.data_dir, "hsm_secret")
    enc_path = os.path.join(args.data_dir, "hsm_secret.enc")
    pass_file = args.passphrase_file or os.path.join(
        os.path.dirname(os.path.abspath(args.data_dir)), "greenlight-signer.pass"
    )

    if args.verify:
        with open(enc_path, "rb") as f:
            data = f.read()
        with open(pass_file, "rb") as f:
            passphrase = f.read().strip()
        seed = decrypt(data, passphrase)
        print(f"OK: {enc_path} decrypts ({len(seed)} bytes), matches hsm_secret: {seed == open(seed_path, 'rb').read() if os.path.exists(seed_path) else 'seed file absent'}")
        return 0

    if not os.path.exists(seed_path):
        print(f"error: {seed_path} not found", file=sys.stderr)
        return 1

    if os.path.exists(enc_path) and not args.force:
        print(f"error: {enc_path} already exists (re-run with --force to replace)", file=sys.stderr)
        return 1

    with open(seed_path, "rb") as f:
        seed = f.read()

    if os.path.exists(pass_file):
        with open(pass_file, "rb") as f:
            passphrase = f.read().strip()
    else:
        passphrase = secrets.token_urlsafe(32).encode()

    enc = encrypt(seed, passphrase)

    os.makedirs(args.data_dir, exist_ok=True)
    with open(enc_path, "wb") as f:
        f.write(enc)
    os.chmod(enc_path, 0o600)

    with open(pass_file, "wb") as f:
        f.write(passphrase + b"\n")
    os.chmod(pass_file, 0o600)

    # replace the plaintext seed with a zeroed stub; the real seed only
    # exists while the signer is running (see greenlight-signer-run.sh)
    with open(seed_path, "wb") as f:
        f.write(b"\x00" * len(seed))
    os.chmod(seed_path, 0o600)

    print(f"encrypted: {enc_path}")
    print(f"passphrase: {pass_file} (0600)")
    print("note: hsm_secret is now zeroed; greenlight-signer-run.sh decrypts on start")
    return 0


if __name__ == "__main__":
    sys.exit(main())
