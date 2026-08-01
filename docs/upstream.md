# Upstream

This repository is a fork of [getAlby/hub](https://github.com/getAlby/hub). It tracks upstream and carries the Routstr integration as the delta.

## Remote layout

```sh
git remote add upstream https://github.com/getAlby/hub
# origin = this fork, upstream = getAlby/hub
```

## Merging upstream

1. `git fetch upstream && git merge upstream/main`
2. Resolve conflicts. The fork touches a bounded set of files (see the delta table in the README); conflicts concentrate in `http/http_service.go`, `service/*`, `api/*`, and `frontend/src/routes.tsx`.
3. Reapply the daemon dist patches (they live outside the tree; see [daemon-patches.md](daemon-patches.md)).
4. Rebuild and verify (see [development.md](development.md)).

## Patch policy

- **Daemon dist patches** are never committed to this tree. They are parked as upstream PR drafts against the routstrd repository and reapplied after updates. The goal is to land them upstream.
- **Backup fixes** (`api/backup.go`) are generic and are candidates to upstream to getAlby/hub.
- Everything else is the Routstr delta and belongs to this fork.

## Commit style

Logical chunks, type-prefixed (`feat/`, `chore/`, `fix/`). The delta breaks down as: RoutstrdService (supervision + auto top-up control), Hub-direct money rail + auto top-up API, backup fixes, connection page + wizard, docs, chore.
