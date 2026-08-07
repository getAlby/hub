# Upstream sync & push workflow (verified 2026-08-07)

How to push this fork and merge upstream getAlby/hub into it. Every command
below was executed and verified on 2026-08-07 (fork at `b330fc3e`, merged 37
upstream commits, pushed to `welliv/alby-hub-greenlight`).

## Prerequisites

- `gh` authenticated as welliv (or `GITHUB_TOKEN` exported):

```bash
export GH_TOKEN=$(uv run python3 "$HOME/.hermes/skills/github/github-auth/scripts/git-credential-token.py")
gh auth status   # → Logged in to github.com account welliv
```

- Local git identity pinned to the GitHub noreply address. **Required**:
  welliv's account has email-protection enabled, so any push containing
  commits authored with a non-associated email (e.g. the private
  `jackal-botch-icon@duck.com`) is declined with
  `push declined due to email privacy restrictions`.

```bash
git config user.name  "welliv"
git config user.email "174869511+welliv@users.noreply.github.com"
```

If a push is declined anyway (older commits with the private email), rewrite
them before pushing:

```bash
export FILTER_BRANCH_SQUELCH_WARNING=1
git filter-branch -f --env-filter '
if [ "$GIT_AUTHOR_EMAIL" = "jackal-botch-icon@duck.com" ]; then
    export GIT_AUTHOR_NAME="welliv"
    export GIT_AUTHOR_EMAIL="174869511+welliv@users.noreply.github.com"
    export GIT_COMMITTER_NAME="welliv"
    export GIT_COMMITTER_EMAIL="174869511+welliv@users.noreply.github.com"
fi' -- <base-sha>..HEAD
# NOTE: rewrites every commit SHA in the range (trees unchanged); force-push after.
```

## Part A — Push the fork

```bash
git push origin feat/greenlight-backend
# if the local branch was rewritten: git push --force-with-lease origin feat/greenlight-backend
```

Verify the remote head matches local:

```bash
git rev-parse HEAD | cut -c1-8                                  # local tip
gh api repos/welliv/alby-hub-greenlight/branches/feat/greenlight-backend \
  --jq '.commit.sha' | cut -c1-8                                # remote tip
```

## Part B — Merge upstream

```bash
git remote add upstream https://github.com/getAlby/hub.git   # once
git fetch upstream master

# Size the gap (should match the merge-base of the fork's own base):
git rev-list --count HEAD..upstream/master                    # commits behind
git merge-base HEAD upstream/master | cut -c1-12              # fork's base SHA

# Dry-run to see the conflicts before committing to anything:
git merge --no-commit --no-ff upstream/master
git diff --name-only --diff-filter=U                           # conflicted files
git merge --abort                                              # clean up dry-run
```

### Resolving conflicts (2026-08-07: only 2, both frontend)

1. **`frontend/src/lib/backendType.ts` (modify/delete)** — upstream moved it to
   `backendType.tsx` with a typed `backendTypeConfigs` record. Resolution:
   `git rm` the old file and port the fork's `GREENLIGHT` entry into the new
   `.tsx` (same shape: `hasMnemonic`, `hasChannelManagement`, `hasNodeBackup`).
2. **`frontend/src/screens/setup/SetupNode.tsx` (content)** — fork kept a local
   display-config copy; upstream now renders from `backendTypeConfigs`. Take
   upstream's version; the ported GREENLIGHT entry flows through automatically.

After resolving: `grep -rn "<<<<<<<" frontend/src/ --include="*.tsx" --include="*.ts"`
must return nothing.

## Part C — Verify the merge (all four gates)

```bash
export PATH=$PATH:/usr/local/go/bin
go build ./...                                                  # 1. Go compiles
go test ./lnclient/greenlight/ -count=1 -timeout 180s           # 2. GL tests pass
cd frontend && yarn install --network-timeout 300000            # new upstream deps (e.g. qr-code-styling)
yarn build:http                                                 # 3. frontend compiles (tsc + vite)
cd .. && go build -o /tmp/hub-merged cmd/http/main.go           # 4. live smoke (below)
```

Live smoke test against the regtest harness (preconfigured creds, fresh
workdir so env wins over the persisted config DB):

```bash
# fresh dir with .env: WORK_DIR, PORT, UNLOCK_PASSWORD, NETWORK=regtest,
# LN_BACKEND_TYPE=GREENLIGHT, GREENLIGHT_CREDS_PATH=/root/gl-harness/creds,
# GREENLIGHT_NODE_URI=localhost:<harness grpc port>
# setup (any valid BIP-39 mnemonic; preconfigured path ignores it) → start → check:
curl -s http://127.0.0.1:PORT/api/info    # running: True, backend: GREENLIGHT
curl -s -X POST .../api/invoices -d '{"amountMsat":5000,...}'   # invoice mints
```

## Part D — Commit & push the merge

```bash
git add -A
git commit -m "merge: upstream getAlby/hub master (N commits, <base>..<head>)

Conflict resolution (K files): ..."

git push origin feat/greenlight-backend
```

## Evidence record (2026-08-07 run)

| Step | Result |
|---|---|
| Commits behind upstream master | 37 (merge-base `bf9c346a9899`) |
| Conflicts | 2 (both frontend, see Part B) |
| `go build ./...` | OK |
| `go test ./lnclient/greenlight/` | ok (~1.5s, all pass) |
| `yarn build:http` | OK (dist 4.8MB; needed `yarn install` for `qr-code-styling`) |
| Live smoke (regtest harness) | running: True, GREENLIGHT, invoice minted, UI 200 |
| Post-merge drift vs upstream | **0** (relationship clean — see pitfall below) |
| Merge commit | `580b91e1` |
| Branch tip / pushed | `2d7529d1` → `+ b330fc3e...2d7529d1 (forced update)` ✓ |

## Pitfalls (learned the hard way)

- **Email privacy restriction**: welliv's account blocks pushes exposing the
  private email. Always commit with the noreply address (Part Prereqs).
- **⚠️ NEVER use `git filter-branch` to rewrite the email of a merged branch**:
  it rewrites *every* commit SHA in the range — including the upstream commits
  — so the fork's history no longer matches upstream's SHAs and every future
  `git merge upstream/master` re-conflicts on the same files
  (`git rev-list --count HEAD..upstream/master` stays > 0 forever).
  **Correct approach** (what this repo's history uses):
  1. `git branch keep <rewritten-tip>` (save your commits)
  2. `git reset --hard <fork-base>` (the last commit before the merge)
  3. `git merge upstream/master` (brings **original** upstream SHAs) → resolve
     conflicts
  4. `git cherry-pick <fork-base>..keep^` (re-applies your local commits;
     cherry-pick uses the current committer identity = noreply)
  5. verify `git rev-list --count HEAD..upstream/master` → **0**
  6. `git push --force-with-lease`
- **Persisted GL config**: `GREENLIGHT_CREDS_PATH`/`GREENLIGHT_NODE_URI` are
  stored in the encrypted config DB at provisioning; env changes alone don't
  take effect on an existing workdir. Use a fresh workdir to re-point a hub.
- **Frontend deps**: upstream adds deps (`qr-code-styling`); `yarn install`
  before `yarn build:http` or tsc fails with TS2307.
- **Shallow clones can't serve as merge sources** — fetch upstream from GitHub
  (https), not from a `--depth 1` local clone (no common ancestor).
