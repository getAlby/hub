# Security

## The key model

API keys are **native to the daemon that minted them**. A key is a bearer token whose identity is a client record in the daemon's own registry (`routstr.db`). A key from another instance is not usable here, and a key minted here only works against this daemon.

**The key is not a security boundary.** Verified behavior: no key, a garbage key, and a foreign key all get the same result on the completions path (`402`), because the spend gate is the **Cashu wallet balance**, not the key. Client keys exist for usage accounting and local registry identity. Per-key budget gating was audited and rejected as security theater; the wallet balance is the real gate.

## Deployment hardening (required)

The daemon binds `*:8008` and exposes **unauthenticated admin endpoints** (`/clients`, `/clients/add`, `/clients/delete`, `/keys/balance`), and the completions path has no key check (spend gate only). This means:

- Anyone who can reach 8008 can list, create, and delete clients.
- Anyone who can reach 8008 can spend a funded wallet.

**The firewall must block 8008 before the wallet holds sats.** This is a hard requirement of the deployment, not a recommendation. Upstream, the daemon should bind 127.0.0.1 and/or require admin auth.

## Secrets

- The unlock password unlocks the Hub; the JWT secret is persisted encrypted and derived from it.
- The cocod config holds the Cashu mnemonic in plaintext. It is protected by filesystem permissions and by the encrypted backup. Cashu proofs are re-derivable from the mnemonic.
- The daemon config holds an nsec (Nostr identity for provider discovery and NWC).
- Never commit `.env`, `.data`, or any config with real secrets.

## Backup crypto

`albyhub.bkp` is an encrypted ZIP: PBKDF2 (4096 iterations, 8-byte salt) derives the key, AES-256-OFB (16-byte IV) encrypts. See [backup-restore.md](backup-restore.md).

> **Known weakness (inherited from upstream).** 4096 PBKDF2 iterations is weak for a password-derived key, and the archive has no authentication (an attacker who knows the password can tamper undetected; OFB has no integrity check). This fork adds the daemon databases (Cashu mnemonic, keys) to the archive, raising the value of the target. Mitigations: use a long random backup password, keep backups offline. A stronger KDF (scrypt/Argon2) + authenticated encryption (AES-GCM) is the right upstream fix.

## Reporting

Open an issue in this repository. For sensitive findings, contact the maintainers directly.
