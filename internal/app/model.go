package app

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/carlogilmar/unravel/internal/git"
	"github.com/carlogilmar/unravel/internal/session"
)

type Screen int

const (
	ScreenHypothesis Screen = iota
	ScreenFileList
	ScreenDiff
	ScreenNotePrompt
	ScreenSummary
	ScreenFatal
)

type Model struct {
	ProjectDir string
	Repo       *git.Repo
	Store      *session.Store
	HeadSHA    string

	Screen   Screen
	Prev     Screen
	FatalErr error

	HypothesisInput textinput.Model
	NoteInput       textinput.Model

	Files       []git.FileChange
	FileCursor  int
	Marked      map[string]bool

	DiffCursor int
	Focus      bool

	NoteTarget noteTarget

	Sess   *session.Session
	Width  int
	Height int
}

type noteTarget struct {
	File   string
	HunkID string
}

func New(projectDir string, repo *git.Repo, store *session.Store, headSHA string) Model {
	hi := textinput.New()
	hi.Placeholder = "What do I think this change does?"
	hi.CharLimit = 240
	hi.Width = 80
	hi.Focus()

	ni := textinput.New()
	ni.Placeholder = "Could I change this safely? Why is it the way it is?"
	ni.CharLimit = 240
	ni.Width = 80

	return Model{
		ProjectDir:      projectDir,
		Repo:            repo,
		Store:           store,
		HeadSHA:         headSHA,
		Screen:          ScreenHypothesis,
		HypothesisInput: hi,
		NoteInput:       ni,
		Marked:          map[string]bool{},
	}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) markedKey(file, hunkID string) string { return file + "\x00" + hunkID }

func (m Model) currentFile() *git.FileChange {
	if m.FileCursor < 0 || m.FileCursor >= len(m.Files) {
		return nil
	}
	return &m.Files[m.FileCursor]
}

func (m Model) totalHunks() int {
	n := 0
	for _, f := range m.Files {
		n += len(f.Hunks)
	}
	return n
}

func (m Model) markedCount() int { return len(m.Marked) }

func (m Model) allMarked() bool {
	return m.totalHunks() > 0 && m.markedCount() >= m.totalHunks()
}
