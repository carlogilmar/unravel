package app

import (
	"path/filepath"
	"testing"
	"time"

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
	if m.Screen != ScreenHypothesis {
		t.Fatalf("initial screen: %v", m.Screen)
	}

	// Empty hypothesis: pressing Enter should not advance.
	upd, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := upd.(Model)
	if m2.Screen != ScreenHypothesis {
		t.Fatal("empty hypothesis should not advance")
	}

	// Type a hypothesis and submit.
	for _, r := range "I think this adds multiply" {
		upd, _ = upd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	upd, cmd := upd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 = upd.(Model)
	if cmd == nil {
		t.Fatal("expected startSession cmd")
	}

	// Run the cmd; it returns SessionStartedMsg.
	msg := cmd()
	upd, _ = m2.Update(msg)
	m3 := upd.(Model)
	if m3.Sess == nil || m3.Sess.Hypothesis == "" {
		t.Fatalf("session not started: %+v", m3.Sess)
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
	upd, _ = m3.Update(DiffLoadedMsg{Files: files})
	m4 := upd.(Model)
	if m4.Screen != ScreenFileList {
		t.Fatalf("expected file list screen, got %v", m4.Screen)
	}

	// Open the file.
	upd, _ = m4.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m5 := upd.(Model)
	if m5.Screen != ScreenDiff {
		t.Fatalf("expected diff screen, got %v", m5.Screen)
	}

	// Press 'm' to open note prompt.
	upd, _ = m5.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m6 := upd.(Model)
	if m6.Screen != ScreenNotePrompt {
		t.Fatalf("expected note prompt, got %v", m6.Screen)
	}

	// Empty note: enter should not advance.
	upd, _ = m6.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m7 := upd.(Model)
	if m7.Screen != ScreenNotePrompt {
		t.Fatal("empty note should not advance")
	}

	// Type a note and submit.
	for _, r := range "I traced the call sites" {
		upd, _ = upd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	upd, cmd = upd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected markHunk cmd")
	}
	msg = cmd()
	upd, _ = upd.Update(msg)
	m8 := upd.(Model)
	if !m8.Marked[m8.markedKey("foo.py", "h1")] {
		t.Fatal("hunk not marked after note submission")
	}
	if !m8.allMarked() {
		t.Fatal("all hunks should be marked now")
	}

	// q from file list should go to summary now (all marked).
	upd, _ = m8.Update(tea.KeyMsg{Type: tea.KeyEsc}) // back to file list
	m9 := upd.(Model)
	if m9.Screen != ScreenFileList {
		t.Fatalf("expected file list, got %v", m9.Screen)
	}
	upd, _ = m9.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m10 := upd.(Model)
	if m10.Screen != ScreenSummary {
		t.Fatalf("expected summary, got %v", m10.Screen)
	}

	// Confirm close: y triggers closeSession cmd → SessionClosedMsg.
	upd, cmd = m10.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected closeSession cmd")
	}
	msg = cmd()
	upd, cmd = upd.Update(msg)
	if cmd == nil {
		t.Fatal("expected tea.Quit cmd after session close")
	}
	// tea.Quit returns a quitMsg; we don't need to verify its concrete type.
	_ = time.Now()
}
