package git

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Status string

const (
	StatusAdded    Status = "Added"
	StatusModified Status = "Modified"
	StatusDeleted  Status = "Deleted"
	StatusRenamed  Status = "Renamed"
)

type FileChange struct {
	Path     string
	OldPath  string
	Status   Status
	Hunks    []Hunk
	Language string
}

type Hunk struct {
	ID       string
	Header   string
	OldStart int
	NewStart int
	Lines    []Line
}

type LineKind int

const (
	LineContext LineKind = iota
	LineAdd
	LineDel
)

type Line struct {
	Kind    LineKind
	OldNum  int
	NewNum  int
	Content string
}

type Repo struct {
	repo *git.Repository
	root string
}

func Open(path string) (*Repo, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	r, err := git.PlainOpenWithOptions(abs, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}
	wt, err := r.Worktree()
	if err != nil {
		return nil, err
	}
	return &Repo{repo: r, root: wt.Filesystem.Root()}, nil
}

func (r *Repo) Root() string { return r.root }

func (r *Repo) HeadSHA() (string, error) {
	ref, err := r.repo.Head()
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return "", nil
		}
		return "", err
	}
	return ref.Hash().String(), nil
}

func (r *Repo) ChangedFiles(allowedExts []string) ([]FileChange, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return nil, err
	}
	st, err := wt.Status()
	if err != nil {
		return nil, err
	}

	headTree, err := r.headTree()
	if err != nil {
		return nil, err
	}

	allowed := make(map[string]bool, len(allowedExts))
	for _, e := range allowedExts {
		allowed[e] = true
	}

	var out []FileChange
	for path, s := range st {
		ext := strings.ToLower(filepath.Ext(path))
		if !allowed[ext] {
			continue
		}
		status := classify(s.Staging, s.Worktree)
		if status == "" {
			continue
		}
		fc := FileChange{
			Path:     path,
			Status:   status,
			Language: langFromExt(ext),
		}
		hunks, err := r.diffFile(headTree, wt, path, status)
		if err != nil {
			return nil, fmt.Errorf("diff %s: %w", path, err)
		}
		fc.Hunks = hunks
		out = append(out, fc)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Status != out[j].Status {
			return statusOrder(out[i].Status) < statusOrder(out[j].Status)
		}
		return out[i].Path < out[j].Path
	})
	return out, nil
}

func (r *Repo) headTree() (*object.Tree, error) {
	ref, err := r.repo.Head()
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return nil, nil
		}
		return nil, err
	}
	commit, err := r.repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}
	return commit.Tree()
}

func (r *Repo) diffFile(headTree *object.Tree, wt *git.Worktree, path string, status Status) ([]Hunk, error) {
	var oldContent, newContent string

	if status != StatusAdded && headTree != nil {
		f, err := headTree.File(path)
		if err == nil {
			oldContent, _ = f.Contents()
		}
	}
	if status != StatusDeleted {
		bs, err := readWorktree(wt, path)
		if err == nil {
			newContent = string(bs)
		}
	}

	return computeHunks(oldContent, newContent), nil
}

func readWorktree(wt *git.Worktree, path string) ([]byte, error) {
	f, err := wt.Filesystem.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var b strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			b.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	return []byte(b.String()), nil
}

// ensure diff package is used for type references in case we extend later.
var _ diff.Operation = diff.Add

func classify(staging, worktree git.StatusCode) Status {
	switch {
	case worktree == git.Renamed || staging == git.Renamed:
		return StatusRenamed
	case worktree == git.Deleted || staging == git.Deleted:
		return StatusDeleted
	case (staging == git.Added && worktree == git.Unmodified) || worktree == git.Added || worktree == git.Untracked:
		return StatusAdded
	case worktree == git.Modified || staging == git.Modified:
		return StatusModified
	}
	if worktree != git.Unmodified || staging != git.Unmodified {
		return StatusModified
	}
	return ""
}

func statusOrder(s Status) int {
	switch s {
	case StatusAdded:
		return 0
	case StatusModified:
		return 1
	case StatusRenamed:
		return 2
	case StatusDeleted:
		return 3
	}
	return 9
}

func langFromExt(ext string) string {
	switch ext {
	case ".ex", ".exs":
		return "elixir"
	case ".py":
		return "python"
	}
	return ""
}

func hunkID(header string, idx int) string {
	h := sha1.Sum([]byte(fmt.Sprintf("%d|%s", idx, header)))
	return hex.EncodeToString(h[:])[:12]
}
