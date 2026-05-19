package app

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/carlogilmar/unravel/internal/git"
	"github.com/carlogilmar/unravel/internal/session"
)

func TestScreenTransitions(t *testing.T) {
	dir := t.TempDir()
	store, err := session.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	m := New("/some/project", nil, store, "deadbeef")
	if m.Screen != ScreenFileList {
		t.Fatalf("initial screen should be file list, got %v", m.Screen)
	}

	// Init() should kick off a startSession cmd with the default title.
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a cmd")
	}

	// Drive startSession by running the session-start command directly.
	startMsg := m.startSession(defaultTitle)()
	upd, _ := m.Update(startMsg)
	m2 := upd.(Model)
	if m2.Sess == nil || m2.Sess.Hypothesis != defaultTitle {
		t.Fatalf("session not started with default title: %+v", m2.Sess)
	}

	// Feed a synthetic DiffLoadedMsg with one file with one hunk.
	files := []git.FileChange{{
		Path:   "foo.py",
		Status: git.StatusModified,
		Hunks: []git.Hunk{{
			ID:     "h1",
			Header: "@@ -1 +1 @@",
			Lines:  []git.Line{{Kind: git.LineAdd, NewNum: 1, Content: "x = 1"}},
		}},
	}}
	upd, _ = m2.Update(DiffLoadedMsg{Files: files})
	m3 := upd.(Model)
	if m3.Screen != ScreenFileList {
		t.Fatalf("expected file list screen, got %v", m3.Screen)
	}

	// Open the file.
	upd, _ = m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := upd.(Model)
	if m4.Screen != ScreenDiff {
		t.Fatalf("expected diff screen, got %v", m4.Screen)
	}

	// Pressing space should expand the current hunk.
	upd, _ = m4.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	m4 = upd.(Model)
	if !m4.Expanded[m4.markedKey("foo.py", "h1")] {
		t.Fatal("space should expand the current hunk")
	}

	// Press 'd' to toggle full diff.
	upd, _ = m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m4 = upd.(Model)
	if !m4.ShowFullDiff {
		t.Fatal("'d' should toggle full diff on")
	}

	// Press 'm' — now an inline note opens, no screen change.
	upd, _ = m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m5 := upd.(Model)
	if m5.Screen != ScreenDiff || !m5.InlineNote {
		t.Fatalf("expected inline note on diff screen, got screen=%v inline=%v", m5.Screen, m5.InlineNote)
	}

	// Empty inline note: enter should not advance.
	upd, _ = m5.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m6 := upd.(Model)
	if !m6.InlineNote {
		t.Fatal("empty note should keep inline note open")
	}

	// Type a note and submit.
	for _, r := range "I traced the call sites" {
		upd, _ = upd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	upd, cmd = upd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected markHunk cmd")
	}
	msg := cmd()
	upd, _ = upd.Update(msg)
	m7 := upd.(Model)
	if !m7.Marked[m7.markedKey("foo.py", "h1")] {
		t.Fatal("hunk not marked after note submission")
	}
	if m7.Notes[m7.markedKey("foo.py", "h1")] == "" {
		t.Fatal("note text should be persisted in memory")
	}
	if !m7.allMarked() {
		t.Fatal("all hunks should be marked now")
	}

	// Pressing 'm' again on a marked hunk should pre-fill the existing note
	// so it can be edited (not replaced from blank).
	upd, _ = m7.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	mEdit := upd.(Model)
	if !mEdit.InlineNote {
		t.Fatal("'m' on a marked hunk should re-open the inline editor")
	}
	if mEdit.NoteInput.Value() != "I traced the call sites" {
		t.Fatalf("inline editor should be pre-filled with existing note, got %q", mEdit.NoteInput.Value())
	}
	// Cancel the edit; mark and stored note should be untouched.
	upd, _ = mEdit.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mCancel := upd.(Model)
	if !mCancel.Marked[mCancel.markedKey("foo.py", "h1")] {
		t.Fatal("cancelling an edit must not unmark the hunk")
	}
	if mCancel.Notes[mCancel.markedKey("foo.py", "h1")] != "I traced the call sites" {
		t.Fatal("cancelling an edit must not change the stored note")
	}
	m7 = mCancel

	// Open summary explicitly via 's'.
	upd, _ = m7.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m8 := upd.(Model)
	if m8.Screen != ScreenSummary {
		t.Fatalf("expected summary, got %v", m8.Screen)
	}

	// From summary, 'y' should move to the confirm-close screen,
	// NOT close immediately.
	upd, _ = m8.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	mConfirm := upd.(Model)
	if mConfirm.Screen != ScreenConfirmClose {
		t.Fatalf("expected confirm-close screen, got %v", mConfirm.Screen)
	}

	// 'n' on confirm screen should return to the summary without closing.
	upd, _ = mConfirm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	mBack := upd.(Model)
	if mBack.Screen != ScreenSummary {
		t.Fatalf("expected to return to summary after declining close, got %v", mBack.Screen)
	}

	// Open confirm again and accept: y triggers closeSession.
	upd, _ = mBack.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	upd, cmd = upd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected closeSession cmd after confirmation")
	}
	msg = cmd()
	upd, cmd = upd.Update(msg)
	if cmd == nil {
		t.Fatal("expected tea.Quit cmd after session close")
	}
}

func TestTitleEdit(t *testing.T) {
	dir := t.TempDir()
	store, err := session.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	m := New("/some/project", nil, store, "deadbeef")
	startMsg := m.startSession(defaultTitle)()
	upd, _ := m.Update(startMsg)
	m = upd.(Model)

	// 't' on file list opens the title editor.
	upd, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = upd.(Model)
	if m.Screen != ScreenTitleEdit {
		t.Fatalf("expected title edit screen, got %v", m.Screen)
	}
	if m.TitleInput.Value() != defaultTitle {
		t.Fatalf("title input should be pre-filled, got %q", m.TitleInput.Value())
	}

	// Replace the title text.
	m.TitleInput.SetValue("Review auth refactor")
	upd, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = upd.(Model)
	if m.Screen != ScreenFileList {
		t.Fatalf("expected to return to file list after save, got %v", m.Screen)
	}
	if cmd == nil {
		t.Fatal("expected updateTitle cmd")
	}
	if m.Sess.Hypothesis != "Review auth refactor" {
		t.Fatalf("session title not updated in memory: %q", m.Sess.Hypothesis)
	}

	// Run the cmd to verify it persists without error.
	tm := cmd()
	if upd, _ := m.Update(tm); upd.(Model).Sess.Hypothesis != "Review auth refactor" {
		t.Fatal("title should remain after persistence")
	}
}
