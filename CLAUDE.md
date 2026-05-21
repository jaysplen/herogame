# CLAUDE.md — agent operating rules for `herogame`

This file is consumed automatically by Claude Code and Cowork. Read it before
touching files. It encodes lessons from earlier sessions so we don't relearn
them the hard way.

## 1. File-write reliability (Windows mount)

This repo is normally checked out on Windows (`C:\projects\herogame`) and
mounted into a Linux sandbox via a network filesystem. That mount has two
known failure modes:

### 1a. The `Write` and `Edit` tools silently truncate large files

When writing through Anthropic's `Write` / `Edit` tools to a file on the
Windows mount, content past roughly the first ~10–60 KB can be dropped
mid-line. The tool reports success, but the file on disk is incomplete.
We've seen this with `frontend/src/index.css` (~25 KB) and several `.tsx`
files in the 2–6 KB range.

**Rule — use bash heredoc for any file write over ~50 lines or 2 KB:**

```bash
cat > /absolute/path/to/file.ext << 'EOF'
...full content...
EOF
wc -l /absolute/path/to/file.ext   # verify
tail -3 /absolute/path/to/file.ext # confirm tail landed
```

Use a quoted heredoc delimiter (`'EOF'`, not `EOF`) to disable shell
interpolation — your CSS / JSX / regexes will contain `$`, backticks, and
backslashes that would otherwise be eaten.

For small files (< 50 lines, plain text), the `Write` tool is fine. For
single-line tweaks, `Edit` is fine.

**After any write, verify:** `wc -l`, `tail -3`, then run the build
(`npx tsc -b`, `npx vite build`, `go build ./...`) before declaring done.

### 1b. File-mode noise pollutes `git status`

The mount surfaces files as mode 755 even though they were 644 on Windows.
`git status` then shows hundreds of files as modified with `old mode 100644
/ new mode 100755` and no content diff.

**Rule — disable filemode tracking on first contact:**

```bash
git -C /sessions/<session>/mnt/herogame config core.fileMode false
```

After that, `git status` shows only real changes.

### 1c. Stuck `.git/index.lock`

If a git operation in the mounted repo is interrupted, `.git/index.lock`
gets left behind and **cannot be removed from inside the sandbox** (Windows
holds the inode). Symptom: `fatal: Unable to create '.git/index.lock':
File exists.`

**Workaround — work in a clean clone outside the mount:**

```bash
rm -rf /tmp/hg
git clone https://github.com/jaysplen/herogame.git /tmp/hg
# do your work in /tmp/hg, commit there, push from there
# then `cp` the changed files back into the mount for the user
```

The user must remove the stale lock manually on Windows
(`Remove-Item .\.git\index.lock -Force` in PowerShell).

## 2. Authentication for `git push`

The sandbox has no GitHub credentials by default. To push:

- Ask the user for a fine-scoped Personal Access Token (`repo` scope, short
  expiry). Inject it as `https://x-access-token:${TOKEN}@github.com/...` in
  the remote URL, push, then **immediately reset the remote URL** to the
  plain HTTPS form and `unset TOKEN`. Never write the token to disk if
  avoidable.
- Alternative: use `gh auth login` with the device-code flow so the token
  never crosses chat.
- Remind the user to revoke the token after the push lands.

## 3. Repository conventions (from AGENTS.md)

- Server owns: travel time, arrivals, combat, economy ticks, hero position,
  resources. Client interpolates movement and computes display-only gold
  estimates between `castle.tick`.
- Read `docs/architecture.md`, `docs/game_rules.md`, and `docs/agent_tasks.md`
  before implementing anything new.
- Use `make server` (sets `DATABASE_URL` correctly), not ad-hoc binaries.
- Minimal diffs; match existing patterns. No commits unless the user asks.
- Feature work goes on a branch per `docs/BRANCH_WORKFLOW.md`, not on
  `master`.

## 4. Pre-flight checklist

Before the first file write of a session, run once:

```bash
REPO=/sessions/<your-session-id>/mnt/herogame
git -C "$REPO" config core.fileMode false
git -C "$REPO" status --short                 # should show only your real changes
ls "$REPO/.git/index.lock" 2>/dev/null && echo "STALE LOCK — clone to /tmp instead"
```

Before claiming any task done:

```bash
cd /sessions/<session>/mnt/herogame/frontend && npx tsc -b && npx vite build
cd /sessions/<session>/mnt/herogame/backend  && go test ./... -count=1
```
