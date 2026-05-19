package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestChangedFilesAndHunks(t *testing.T) {
	dir := t.TempDir()

	mustGit(t, dir, "init", "-q")
	mustGit(t, dir, "config", "user.email", "test@example.com")
	mustGit(t, dir, "config", "user.name", "test")

	writeFile(t, dir, "sample.py", "def add(a, b):\n    return a + b\n\ndef sub(a, b):\n    return a - b\n")
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-qm", "initial")

	writeFile(t, dir, "sample.py", "def add(a, b):\n    return a + b + 0\n\ndef sub(a, b):\n    return a - b\n\ndef mul(a, b):\n    return a * b\n")
	writeFile(t, dir, "new.py", "x = 42\n")
	writeFile(t, dir, "ignored.txt", "not a target language\n")

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	files, err := repo.ChangedFiles([]string{".py", ".ex", ".exs"})
	if err != nil {
		t.Fatalf("changed: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files (sample.py modified, new.py added), got %d: %+v", len(files), files)
	}

	var newPy, samplePy *FileChange
	for i := range files {
		switch files[i].Path {
		case "new.py":
			newPy = &files[i]
		case "sample.py":
			samplePy = &files[i]
		}
	}
	if newPy == nil || newPy.Status != StatusAdded {
		t.Fatalf("expected new.py added, got %+v", newPy)
	}
	if samplePy == nil || samplePy.Status != StatusModified {
		t.Fatalf("expected sample.py modified, got %+v", samplePy)
	}

	if len(samplePy.Hunks) == 0 {
		t.Fatal("expected hunks for modified file")
	}
	sawAdd := false
	for _, h := range samplePy.Hunks {
		for _, l := range h.Lines {
			if l.Kind == LineAdd {
				sawAdd = true
			}
		}
	}
	if !sawAdd {
		t.Fatal("expected at least one added line in sample.py diff")
	}

	for _, f := range files {
		if filepath.Ext(f.Path) == ".txt" {
			t.Fatalf("ignored.txt should have been filtered out")
		}
	}
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, string(out))
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
