# Unravel — development context

Load this file when starting a Claude session in this repo to continue development with full context. It captures *what Unravel is*, *how it's built*, *current state*, and *where to go next* — distilled from the design docs and the work done so far.

---

## 1. What Unravel is

A terminal UI (TUI) that helps developers **read, understand, and own AI-generated code**. The thesis: AI produces code faster than humans can move it into long-term memory, eroding agency and ownership. Unravel is a deliberate *reading* discipline that refuses to short-circuit comprehension.

It is **not** a diff viewer, not an AI summarizer, and not a linter. It is the human-anchored *review/verify* surface you reach for after an agent drops code on your branch.

**Design philosophy (do not violate without discussion):**
- **Protect germane cognitive load.** No AI summaries, no auto-marking "trivial" hunks, no gamification. The work must stay hard enough that something sticks.
- **Marking a hunk understood requires a written "why" note** in the developer's own words. The gesture must carry weight.
- **The git working tree is the only source of truth.** Unravel reflects it; it never manages the user's git workflow.

Full theory: `documentation/articles/REVIEW_CODE.md` (cognitive science / Hermans).
Industry context 2026: `documentation/articles/REVIEW_CODE_IN_AI.md`.

---

## 2. Tech stack

- **Language:** Go (module declares `go 1.25.0`; `.tool-versions` pins toolchain `golang 1.23.5`).
- **TUI:** Bubble Tea (Elm architecture) `v1.3.10`, Lip Gloss `v1.1.0`, Bubbles `v1.0.0` (textinput).
- **Git:** go-git `v5.19.0` (in-process, no shelling out).
- **Diffing:** `sergi/go-diff` (diffmatchpatch) — line-level diff, custom hunk computation.
- **Storage:** SQLite via `modernc.org/sqlite` `v1.50.1` (pure-Go, no cgo). DB path from `internal/config`.
- **Distribution:** single static binary (brew/go install/tarball are future, not now).

---

## 3. Architecture (Elm-style, strict)

Single `Model`, pure `Update(msg) (Model, Cmd)` transitions, all side-effects in `tea.Cmd`s.

```
cmd/unravel/main.go        Entry: parse target dir, open repo+store, run program.
internal/app/
  model.go                 Model struct, Screen enum, New(), Init(), helpers.
  update.go                Update() + per-screen key handlers.
  view.go                  View() + per-screen renderers.
  messages.go              tea.Msg types (DiffLoaded, SessionStarted, HunkMarked, ...).
  commands.go              tea.Cmd factories (loadDiff, startSession, markHunk, ...).
  model_test.go            Screen-transition + title-edit tests.
internal/git/
  git.go                   Repo open, ChangedFiles(), status classification.
  diff.go                  computeHunks() — line ops → hunks with context.
internal/session/
  store.go                 SQLite store: sessions, hunk_marks, line_notes tables.
  types.go                 Session / HunkMark / LineNote structs.
internal/config/paths.go   XDG-aware DB path.
internal/view/styles.go    Lip Gloss styles + color palette.
```

**Data model (SQLite):** `sessions(id, project_dir, head_sha, hypothesis, started_at, closed_at)`, `hunk_marks(session_id, file, hunk_id, why_note, marked_at)`, `line_notes(session_id, file, line, note, created_at)`. The `hypothesis` column doubles as the editable session **title**.

---

## 4. Current behavior (after Sprint 2)

### Screens (`Screen` enum in `model.go`)
`ScreenFileList`, `ScreenDiff`, `ScreenTitleEdit`, `ScreenSummary`, `ScreenConfirmClose`, `ScreenFatal`.
> Note: the old `ScreenHypothesis` and `ScreenNotePrompt` were removed in Sprint 2 — the hypothesis prompt is gone (auto-titled) and notes are now inline, not a separate screen.

### Flow
1. `Init()` auto-starts a session titled **"Check code"** (`defaultTitle`) — no upfront prompt. Loading state shows until the session + diff load.
2. **File list** (left sidebar, lazygit-inspired): files grouped by status (Added/Modified/Deleted/Renamed), filtered to `.ex`, `.exs`, `.py`. Markers: `✓` all hunks marked, `◐` partial.
3. **Diff view:** hunks are **collapsed by default** (header-only list for easy navigation). Default rendering shows **only added (green) lines**; toggle to full diff (with removed/context) with `d`.
4. **Inline notes:** `m` opens a textinput *under the current hunk header*, pre-filled with the existing note so it can be edited (not replaced). Saved notes render inline next to the hunk header. `u` unmarks.
5. **Summary** (`s`): lists every file with its hunks and the why-notes; shows progress. Close only allowed when all hunks marked.
6. **Confirm-close** (`y` from summary): warns that notes live only inside the session and cannot be recovered; `y` confirms, `n`/Esc backs out.

### Keybindings
| Screen | Keys |
|---|---|
| File list | `j/k` move · `enter`/`l` open · `t` title · `s` summary · `q` quit |
| Diff | `j/k` hunk · `space`/`enter` expand · `d` full/added · `m` note · `u` unmark · `f` focus · `s` summary · `t` title · `esc`/`h` back · `q` back |
| Inline note | `enter` save · `esc` cancel |
| Title edit | `enter` save · `esc` cancel |
| Summary | `y`/`enter` → confirm close (if all marked) · `n`/`esc`/`q` back |
| Confirm close | `y`/`enter` close · `n`/`esc`/`q` back |

---

## 5. Build / run / install

```bash
go run ./cmd/unravel               # run against current dir (dev)
go run ./cmd/unravel /path/to/repo # run against another git repo
go build -o ~/bin/unravel ./cmd/unravel   # install/upgrade (~/bin is on PATH)
go test ./...                      # all tests
go vet ./...
```

Install gotcha: `go install` lands the binary in asdf's Go dir (`$GOBIN`), which is **not** on PATH and triggers a stale asdf shim ("No version is set for command unravel"). Always use `go build -o ~/bin/unravel`. Full notes in `readme.md` → "Local development".

---

## 6. Sprint 2 — done in this work stream

All six items from `documentation/sprints/sprint2.md` are implemented, tested, and installed:
1. Default title "Check code", editable any time via `t` (`Store.UpdateHypothesis`).
2. Added-only diff by default; `d` toggles full diff.
3. Collapsable hunks + `DiffCursor` reset on file change (fixes "navigation broken in some files").
4. Inline note next to hunk header; pre-filled for editing; `u` to unmark (split from `m`).
5. Summary screen rebuilt to show files + hunks + descriptions; reachable anytime via `s`.
6. Confirm-close screen with "notes are not recoverable" warning before session close.

---

## 7. Deferred / open ideas (from readme + theory docs)

Not yet built — candidates for future sprints:
- **Recent projects list** (`~/.local/share/unravel/recents.json`, `r` key) — described in Sprint 1, not implemented.
- **Per-line notes UI** — schema exists (`line_notes`, `SaveLineNote`), no UI yet.
- **Live editor loop** — fsnotify re-render on working-tree change; modified hunks reset to unreviewed.
- **Resume a paused session** — restore marks/notes/last file on reopen (store supports it; not wired).
- **Syntax highlighting** — chroma (v1.5 target).
- **Roles-of-variables tagging** (Sajaniemi) and **modification-test prompt** before marking.
- **"Last reviewed at" per file** to surface ownership decay over time.
- Empty/error full-screen states beyond the current fatal screen.

---

## 8. Conventions & guardrails for this repo

- Keep the Elm separation strict: state in `Model`, transitions pure, side-effects in `Cmd`s. Define new message types in `messages.go`, new commands in `commands.go`.
- Don't add AI-summary / auto-grade / gamification features — they contradict the product thesis.
- Filter remains `.ex`, `.exs`, `.py` (Elixir + Python) for now; other languages explicitly hidden.
- Update `model_test.go` whenever screen flow or keybindings change.
- Sprint docs live in `documentation/sprints/`; long-form articles in `documentation/articles/`.
