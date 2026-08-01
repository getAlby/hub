# Daemon Patches

routstrd ships as a compiled Bun bundle. Four behavioral fixes are carried as patches on the installed dist and must be reapplied after every routstrd update.

## The patches

| Patch | Effect |
|---|---|
| Model cache TTL 30 minutes | The aggregated model catalog refreshes every 30 min instead of the default, so newly announced providers and models appear without stale gaps |
| Background warm-refresh loop | A 30-minute timer re-fetches catalogs in the background, so the first request after a refresh never pays the fetch latency |
| Melt `needsSwap=false` | Forces direct melts instead of letting the mint quote a swap path, avoiding the circular-fee pitfall |
| Model-cap removal | Removes the 21-model ceiling per provider so the full catalog is routable |

## How to reapply

The daemon is installed as a global Bun package (`~/.bun/install/global/node_modules/routstrd/dist/daemon/index.js`). After a routstrd update (`bun update -g routstrd`, or however it was installed):

1. Verify the patches are gone: the defaults are back (e.g., the TTL constant is no longer `30 * 60 * 1000`).
2. Reapply each patch to the dist file (they are small, targeted edits; the exact diffs are tracked alongside the upstream PR drafts).
3. Restart the daemon (the Hub restarts it, or `kill` + the Hub's supervision brings it back).
4. Verify: `GET :8008/models` returns the full catalog and the TTL/warm-loop constants are present in the bundle.

## Status

The patches are parked as upstream PR drafts against the routstrd repository, with the same content as the applied diffs. The goal is to land them upstream and delete this page.
