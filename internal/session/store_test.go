package session

import (
	"path/filepath"
	"testing"
)

func TestStoreLifecycle(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	s, err := store.StartSession("/project", "abc123", "this change adds a multiplier")
	if err != nil {
		t.Fatal(err)
	}
	if s.ID == 0 || s.Hypothesis == "" {
		t.Fatalf("unexpected session: %+v", s)
	}

	if err := store.MarkHunk(s.ID, "foo.py", "h1", ""); err == nil {
		t.Fatal("expected error marking with empty note")
	}
	if err := store.MarkHunk(s.ID, "foo.py", "h1", "I traced add() and it's safe"); err != nil {
		t.Fatal(err)
	}

	marked, note, err := store.IsMarked(s.ID, "foo.py", "h1")
	if err != nil || !marked || note == "" {
		t.Fatalf("expected marked with note, got %v %q %v", marked, note, err)
	}

	all, err := store.MarkedHunks(s.ID)
	if err != nil || len(all) != 1 {
		t.Fatalf("MarkedHunks: %v %v", all, err)
	}

	if err := store.UnmarkHunk(s.ID, "foo.py", "h1"); err != nil {
		t.Fatal(err)
	}
	marked, _, _ = store.IsMarked(s.ID, "foo.py", "h1")
	if marked {
		t.Fatal("expected unmarked")
	}

	if err := store.CloseSession(s.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.CloseSession(s.ID); err == nil {
		t.Fatal("expected error closing already-closed session")
	}
}
