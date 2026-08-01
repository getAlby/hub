# Backup & Restore

The Hub's backup is an encrypted ZIP (`albyhub.bkp`) covering the whole system: the Hub database, the Lightning backend state, the daemon databases, and their configs.

## What is archived

| Entry | Where it lives |
|---|---|
| `nwc.db` | Hub database |
| `bark/bark.sqlite` | Lightning backend state (this deployment) |
| `coco.db` + `config.json` | Cashu wallet (proofs, balances, mint state) |
| `routstr.db` + `config.json` | daemon (clients, usage, provider registry) |

The stale SQLite `LOCK` file is excluded. Paths that escape the storage dir (`../.cocod/...`, `../.routstrd/...`) are restored under the user's home directory.

## How it works (the fixes, from the source)

1. **WAL checkpoint before archiving.** `coco.db` and `routstr.db` are checkpointed (`PRAGMA wal_checkpoint(TRUNCATE)`) before archive. Without this, recent mints/melts and client state can sit only in the WAL (measured: coco.db-wal can hold 4+ MB, routstr.db-wal 4+ MB) and be silently lost.
2. **Bark included via an absolute glob.** The Bark state was never backed up before (its storage dir is outside the default glob). Now it is.
3. **ResetRouter is warn-only.** The restore flow does not destructive-reset the Lightning router on a fresh instance.
4. **Restore resolves `../` entries against `$HOME`**, so daemon state lands where the daemons expect it.
5. **No continuous backups for Cashu state.** LN channel state is counterparty-punishable if stale, but Cashu proofs are re-derivable from the mnemonic in the cocod config, which is backed up. So a snapshot backup is the right model.

## Restore procedure

1. Install a fresh instance, do not complete setup (restore is rejected on setup-completed instances).
2. Restore the backup.
3. **Restart the Hub process**, then unlock (`/api/start`). This order matters: `/api/start` alone on a fresh process half-starts with `sql: database is closed`, because the restore replaced the database underneath the running process.
4. Verify (below).

## Verification checklist

- Hub login works and apps are present
- Daemon is healthy: `GET :8008/clients`, `GET :8008/keys/balance` return data
- Client keys are recoverable from `routstr.db` (`sdk_storage`, key `client_ids`) if metadata was ever lost
- Cashu balance matches the backup point
- A test payment round-trips (wallet to app wallet, and a mint invoice)

## Crypto format

PBKDF2 (4096 iterations, 8-byte salt) derives the key; AES-256-OFB (16-byte IV) encrypts the ZIP.
