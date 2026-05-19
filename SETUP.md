# Setup

Operational guide for Unravel: install, build, run, use, and extend.

For *what* the app is and *why* it works the way it does, see [readme.md](./readme.md) and [REVIEW_CODE.md](./REVIEW_CODE.md). This file is about the mechanics.

---

## 1. Prerequisites

- **asdf** (any version). Installed via `brew install asdf` on macOS.
- **git** on `PATH` (used by the test suite; the runtime uses `go-git` in-process).

The Go toolchain version is pinned in [`.tool-versions`](./.tool-versions). You do not install Go manually.

---

## 2. First-time setup

From the project root:

```sh
asdf install           # reads .tool-versions, installs the pinned Go
go mod download        # fetches all module deps into the local module cache
```

`asdf install` will skip versions already present. If you see *"No version is set for command go"*, you are not inside the project directory or `.tool-versions` is missing.

---

## 3. Build

A single static binary, no cgo (modernc.org/sqlite is pure-Go):

```sh
go build -o ./bin/unravel ./cmd/unravel
```

The resulting `./bin/unravel` is the whole app — around 15 MB. It can be copied or symlinked anywhere on `PATH`.

For an installed binary on your `PATH`:

```sh
go install ./cmd/unravel
```

This drops `unravel` into `$(go env GOBIN)` (defaults to `~/go/bin`). Make sure that directory is on your shell `PATH`.

---

## 4. Run

Unravel reviews the *unstaged working-tree diff* of a git repository.

```sh
./bin/unravel                  # review $PWD
./bin/unravel /path/to/project # review a specific directory
```

The directory must be a git repository. If it isn't, Unravel exits with an error before starting the TUI.

If there are no `.ex`, `.exs`, or `.py` changes in the working tree, the file-list screen shows an empty-state message. (v1 filters all other extensions.)

---

## 5. Use

The flow follows [readme.md → Use Case Scenarios](./readme.md). Concretely, screen-by-screen:

### 5.1 Hypothesis screen (session start)

The first thing you see. Type one sentence answering *"What do I think this change does?"* and press `Enter`. The session is created in SQLite at this moment; the rest of the UI is gated behind a non-empty hypothesis.

| Key      | Action                  |
|----------|-------------------------|
| `Enter`  | Submit hypothesis       |
| `Ctrl+C` | Quit without a session  |

### 5.2 File list (gist pass)

Left-pane list of changed files grouped by Added / Modified / Deleted / Renamed. Skim names and shapes — confirm or revise your hypothesis. No bodies yet.

| Key       | Action                                                |
|-----------|-------------------------------------------------------|
| `j` / `↓` | Next file                                             |
| `k` / `↑` | Previous file                                         |
| `Enter` / `l` / `→` | Open file in diff viewer                    |
| `q`       | Quit (if not all hunks marked) or open close summary  |
| `Ctrl+C`  | Quit immediately                                      |

Each file shows a small status indicator:
- `  ` — no hunks marked
- `◐` — some hunks marked
- `✓` — all hunks marked

### 5.3 Diff viewer (detail pass)

Right-pane unified diff with added / context lines styled separately. The current hunk is highlighted.

| Key            | Action                                                     |
|----------------|------------------------------------------------------------|
| `j` / `↓`      | Next hunk                                                  |
| `k` / `↑`      | Previous hunk                                              |
| `f`            | Toggle focus mode — dims all hunks except the current one  |
| `m`            | Open note prompt to mark current hunk (or unmark if marked)|
| `Esc` / `h` / `←` | Back to file list                                       |
| `q`            | Back to file list, or close summary if all hunks marked    |

### 5.4 Note prompt (the modification test)

Pressing `m` on an unmarked hunk opens this. The prompt is the question from [REVIEW_CODE.md](./REVIEW_CODE.md): *"Could I change this safely? Why is it the way it is?"*

The note is **required** — Enter on an empty input is ignored. This is intentional. The note is the externalized beacon: a piece of long-term memory written down because the AI did not write it for you.

| Key      | Action                                          |
|----------|-------------------------------------------------|
| `Enter`  | Save note and mark the hunk understood          |
| `Esc`    | Cancel without marking                          |

To unmark a hunk you already marked: navigate to it and press `m` again. The mark is removed; the note stays in the database for the session's history.

### 5.5 Close summary

Triggered by `q` from the file list when every hunk is marked. Shows hypothesis, file count, and hunk progress, then asks for confirmation.

| Key             | Action                                |
|-----------------|---------------------------------------|
| `y` / `Enter`   | Close the session and quit            |
| `n` / `Esc` / `q` | Back to the file list               |

What happens after closing is **outside Unravel's scope**: commit, push, open a PR, or walk away — your call.

### 5.6 Session storage

Sessions live in SQLite at:

```
$XDG_DATA_HOME/unravel/unravel.db
# or, if XDG is unset:
~/.local/share/unravel/unravel.db
```

Each session row records: project directory, `HEAD` SHA at session start, hypothesis, start timestamp, close timestamp. Hunk marks and line notes are foreign-keyed to the session. Sessions are never deleted automatically.

To inspect history:

```sh
sqlite3 ~/.local/share/unravel/unravel.db '.tables'
sqlite3 ~/.local/share/unravel/unravel.db 'SELECT id, hypothesis, started_at FROM sessions ORDER BY id DESC LIMIT 5'
```

---

## 6. Development loop

The standard cycle when adding a feature:

```sh
go build ./...                 # compile everything; nothing produced on disk
go vet ./...                   # static analysis
go test ./...                  # run all tests (fast except the git suite, ~25s)
go build -o ./bin/unravel ./cmd/unravel   # produce the binary
./bin/unravel /path/to/sample-dirty-repo  # smoke-test manually
```

For a tighter loop on a specific package:

```sh
go test ./internal/app/... -run TestScreenTransitions -v
```

### 6.1 Adding a new key binding

1. Add the binding to the right `keyXxx` function in `internal/app/update.go`.
2. If it dispatches a side-effecting operation, add a message type in `internal/app/messages.go` and a `tea.Cmd` constructor in `internal/app/commands.go`.
3. Add the new key to the help text in `helpForScreen` in `internal/app/view.go`.
4. Extend `TestScreenTransitions` in `internal/app/model_test.go` to exercise the new path.

### 6.2 Adding a new screen

1. Add a `ScreenXxx` const in `internal/app/model.go`.
2. Add a `keyXxx` handler in `update.go` and dispatch from `handleKey`.
3. Add a `viewXxx` renderer in `view.go` and dispatch from `View`.
4. Decide the transitions into and out of it — be explicit about which keys move where.
5. Update the help line in `helpForScreen`.

### 6.3 Adding a new persisted field or table

1. Add a `CREATE TABLE` or `ALTER TABLE` statement to the `stmts` slice in `(*Store).migrate` in `internal/session/store.go`. Migrations are append-only and idempotent (`IF NOT EXISTS` / safe `ALTER`).
2. Add or update the corresponding Go type in `internal/session/types.go`.
3. Add accessor methods on `*Store` and exercise them in `store_test.go`.

### 6.4 Filters and languages

The active extension allowlist is hard-coded in `cmd/unravel/main.go`:

```go
files, err := repo.ChangedFiles([]string{".py", ".ex", ".exs"})
```

To add a language for testing, extend this list and add a case to `langFromExt` in `internal/git/git.go`. This is intentionally not config-driven in v1.

### 6.5 Code style

- No comments unless the *why* is non-obvious. Don't explain *what* — the names should.
- No backwards-compatibility shims. If a signature changes, change all the callers.
- Prefer one new well-named function over reusing a poorly-named one.
- Tests live next to the code (`*_test.go`), not in a separate tree.

---

## 7. Updating after a feature change

Run before committing:

```sh
go vet ./...
go test ./...
go build -o ./bin/unravel ./cmd/unravel
```

If you touched dependencies:

```sh
go mod tidy        # prunes unused deps, adds missing ones, refreshes go.sum
```

If you bumped the Go floor (e.g. a dep now requires a newer toolchain):

1. Update `.tool-versions` to the new version.
2. Run `asdf install` to fetch it.
3. The `go` directive in `go.mod` is managed by the `go` tool — leave it alone unless intentionally lowering the floor.

---

## 8. Project layout (reference)

```
.
├── .tool-versions              # asdf-managed Go version
├── cmd/unravel/main.go         # binary entry: arg parse, repo open, program launch
├── go.mod / go.sum             # module + checksums
├── internal/
│   ├── app/                    # Bubble Tea Model, Update, View, Cmds, Msgs
│   │   ├── commands.go         # tea.Cmd constructors (side effects)
│   │   ├── messages.go         # message types delivered via Update
│   │   ├── model.go            # Model struct + helpers
│   │   ├── update.go           # Update + per-screen key handlers
│   │   └── view.go             # per-screen renderers + status bar
│   ├── config/paths.go         # XDG_DATA_HOME resolution, db path
│   ├── git/                    # go-git wrapper + line diff → hunks
│   │   ├── git.go              # Repo, ChangedFiles, status classification
│   │   └── diff.go             # diffmatchpatch line-level diff → Hunks
│   ├── session/                # SQLite persistence
│   │   ├── store.go            # Open, migrate, Mark/Unmark/Close/etc.
│   │   └── types.go            # Session, HunkMark, LineNote
│   └── view/styles.go          # Lip Gloss palette + named styles
└── bin/unravel                 # produced by `go build`, .gitignored
```

---

## 9. Troubleshooting

**`No version is set for command go`** — `.tool-versions` is not visible from your shell's working directory. Run from the project root, or `asdf install` first.

**`missing go.sum entry for module providing package X`** — A new transitive dep showed up. Run `go mod tidy` and rebuild.

**The first `go test ./internal/git/...` takes 25 seconds** — Expected. `go-git` is heavy on first compile. Subsequent runs are sub-second when the cache is warm.

**A dep declared `go 1.25` but `.tool-versions` says 1.23.5** — That's fine. Go's auto-toolchain transparently fetches and uses the higher version. If you want them aligned, update `.tool-versions` and run `asdf install`.

**`unravel: not a git repository`** — The target directory has no `.git`. Either you pointed at the wrong path, or you need to `git init` it.

**TUI renders garbage** — Your terminal likely doesn't support the truecolor / ANSI styles Lip Gloss emits. Try `TERM=xterm-256color ./bin/unravel`.
