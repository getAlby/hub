# Troubleshooting

Operational knowledge, all verified live on a running instance.

## Mint outage

**Symptom:** Top Up, wizard funding, and auto top-up fail; `POST /wallet/receive/bolt11` returns `400` with `{"detail":"could not fetch bolt11 payment request from backend","code":20000}` (Nutshell/0.20).

**Impact:** the daemon still serves requests against its existing Cashu balance; only funding paths are blocked.

**Check:** `curl -s https://<mint>/v1/info` (or the bolt11 endpoint) from the host.

**Resolution:** external. The mint's Lightning backend must recover. Nothing in the stack caches mint state that would let funding continue.

## The locked-window 401

**Symptom:** after a Hub restart, every API call returns `401 missing or malformed jwt`, even with a valid token.

**Cause:** the Hub boots locked. Until the unlock password is supplied (`POST /api/start`), the JWT secret is empty and every request 401s, including the user's.

**Fix:** supply the unlock password right after the process starts. This is a deploy-time window, not an app bug. Do not deploy while a user is mid-flow without warning.

## Ghost keys and the create/delete saga

The "API key doesn't stick" bug had five root causes, each found by watching the live hub log during browser clicks. If key behavior ever looks wrong again, check these in order:

1. **Metadata PATCH replaces the whole object (no merge).** Every metadata write must be read-modify-write against the server (`GET /api/v2/apps/:id`, merge, then PATCH). A stale-closure write wipes other keys (including `app_store_app_id`, which hides the whole Routstr section).
2. **`request()` does not prepend `/api`.** Paths must be fully qualified (`/api/v2/apps/74`). `/v2/apps/74` silently returns the SPA HTML, `.json()` fails, and code paths die silently.
3. **Self-heal validates against the daemon `/clients` list.** `/keys/balance` never contains client ids (it returns wallet + provider-token entries). Validating against it wipes every key as a "ghost".
4. **React hooks discipline.** No conditional `return null` between hook calls. Use a hook-free wrapper + inner component (React error #300).
5. **Delete of a ghost client returns 404.** Treated as "already removed" = success, metadata cleared.

**Recovery when metadata is lost:** the full key strings live in `routstr.db` (`sdk_storage`, key `client_ids`) and can be restored into metadata.

## Delete Key fails with 404

The client was already removed (ghost). The dialog treats it as success: "already removed", metadata cleared. If the delete returns 401 instead, you are in the locked window (above).

## cocod stuck: daemon alive but socket missing

**Symptom:** routstrd is running but `cocod.sock` is missing; startup hangs on the mint rate limiter (stuck pending operations).

**Fix:**

```sh
pkill -f cocod          # kill ALL cocod daemons
rm -f cocod.sock cocod.pid
# clear stuck pending mint operations from coco.db:
# DELETE FROM coco_cashu_mint_operations WHERE state='pending';
start one cocod daemon
```

## Models available shows nothing (or 502 on refresh)

**Symptom:** the wizard's Models dialog fails to load; hub log shows `502 proxy error: context canceled` on `/api/routstrd/models/all?refresh=true`.

**Cause:** the daemon's full federation refresh (`refresh=true`) has no per-provider timeouts — dead or unreachable provider nodes hang it for 90-120s+, and the frontend aborts at 15s. Verified on a clean single daemon; upstream issue.

**Fix (deployed):** the wizard opens the dialog with the cached catalog (fast); the daemon's 30-minute warm loop keeps it fresh. Never trigger `refresh=true` from a blocking UI path.

**Related:** a Hub restart can leave two routstrd daemons (the old one keeps port 8008, the new one can't bind, and the pid file points at the new one). Symptoms: refresh hangs (DB contention), health still 200. Fix: kill all routstrd daemons, remove `routstrd.pid`, let supervision start one.

## Daemon unreachable

`/api/routstrd/*` fails. Check the daemon process (`pgrep -f routstrd`), the port (`ss -ltnp | grep 8008`), and the Hub log for supervision restarts. The Hub restarts the daemon on health-check failure.

## Melt-quote circular fee

Melt quotes can quote a fee on the fee (circular). The `needsSwap=false` patch forces direct melts. If melts misbehave, verify the patch is still applied (see [daemon-patches.md](daemon-patches.md)).

## The `usagePi` typo

The daemon's usage endpoint is exposed as `/usagePi` in one place and `/usage/summary` in others. Check both if a usage call 404s.

## Stale bundle

The frontend pins the bundle in the service worker (`sw.js`). After a deploy, the browser can keep serving the old bundle. Hard refresh (Ctrl+Shift+R) after every deploy.
