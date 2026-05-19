package app

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/carlogilmar/unravel/internal/git"
	"github.com/carlogilmar/unravel/internal/session"
)

type Screen int

const (
	ScreenFileList Screen = iota
	ScreenDiff
	ScreenTitleEdit
	ScreenSummary
	ScreenConfirmClose
	ScreenFatal
)

const defaultTitle = "Check code"

type Model struct {
	ProjectDir string
	Repo       *git.Repo
	Store      *session.Store
	HeadSHA    string

	Screen   Screen
	Prev     Screen
	FatalErr error

	TitleInput textinput.Model
	NoteInput  textinput.Model

	Files      []git.FileChange
	FileCursor int
	Marked     map[string]bool
	Notes      map[string]string
	Expanded   map[string]bool

	DiffCursor   int
	Focus        bool
	ShowFullDiff bool
	InlineNote   bool
	NoteTarget   noteTarget

	Sess   *session.Session
	Width  int
	Height int
}

type noteTarget struct {
	File   string
	HunkID string
}

func New(projectDir string, repo *git.Repo, store *session.Store, headSHA string) Model {
	ti := textinput.New()
	ti.Placeholder = "Session title"
	ti.CharLimit = 120
	ti.Width = 60

	ni := textinput.New()
	ni.Placeholder = "Why is this the way it is? Could I change it safely?"
	ni.CharLimit = 240
	ni.Width = 70

	return Model{
		ProjectDir: projectDir,
		Repo:       repo,
		Store:      store,
		HeadSHA:    headSHA,
		Screen:     ScreenFileList,
		TitleInput: ti,
		NoteInput:  ni,
		Marked:     map[string]bool{},
		Notes:      map[string]string{},
		Expanded:   map[string]bool{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.startSession(defaultTitle))
}

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
