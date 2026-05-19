package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width, m.Height = msg.Width, msg.Height
		return m, nil

	case SessionStartedMsg:
		if msg.Err != nil {
			m.FatalErr = msg.Err
			m.Screen = ScreenFatal
			return m, nil
		}
		m.Sess = msg.Sess
		return m, m.loadDiff()

	case DiffLoadedMsg:
		if msg.Err != nil {
			m.FatalErr = msg.Err
			m.Screen = ScreenFatal
			return m, nil
		}
		m.Files = msg.Files
		m.FileCursor = 0
		m.Screen = ScreenFileList
		return m, nil

	case HunkMarkedMsg:
		if msg.Err == nil {
			m.Marked[m.markedKey(msg.File, msg.HunkID)] = true
		}
		m.Screen = ScreenDiff
		m.NoteInput.SetValue("")
		m.NoteInput.Blur()
		return m, nil

	case HunkUnmarkedMsg:
		if msg.Err == nil {
			delete(m.Marked, m.markedKey(msg.File, msg.HunkID))
		}
		return m, nil

	case SessionClosedMsg:
		if msg.Err != nil {
			m.FatalErr = msg.Err
			m.Screen = ScreenFatal
			return m, nil
		}
		return m, tea.Quit

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	switch m.Screen {
	case ScreenHypothesis:
		return m.keyHypothesis(msg)
	case ScreenFileList:
		return m.keyFileList(msg)
	case ScreenDiff:
		return m.keyDiff(msg)
	case ScreenNotePrompt:
		return m.keyNote(msg)
	case ScreenSummary:
		return m.keySummary(msg)
	case ScreenFatal:
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) keyHypothesis(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		text := strings.TrimSpace(m.HypothesisInput.Value())
		if text == "" {
			return m, nil
		}
		return m, m.startSession(text)
	}
	var cmd tea.Cmd
	m.HypothesisInput, cmd = m.HypothesisInput.Update(msg)
	return m, cmd
}

func (m Model) keyFileList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		if m.allMarked() {
			m.Screen = ScreenSummary
			return m, nil
		}
		return m, tea.Quit
	case ":":
		if m.allMarked() {
			m.Screen = ScreenSummary
		}
		return m, nil
	case "j", "down":
		if m.FileCursor < len(m.Files)-1 {
			m.FileCursor++
		}
	case "k", "up":
		if m.FileCursor > 0 {
			m.FileCursor--
		}
	case "enter", "l", "right":
		if len(m.Files) > 0 {
			m.DiffCursor = 0
			m.Screen = ScreenDiff
		}
	}
	return m, nil
}

func (m Model) keyDiff(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "h", "left":
		m.Screen = ScreenFileList
		m.Focus = false
	case "j", "down":
		f := m.currentFile()
		if f != nil && m.DiffCursor < len(f.Hunks)-1 {
			m.DiffCursor++
		}
	case "k", "up":
		if m.DiffCursor > 0 {
			m.DiffCursor--
		}
	case "f":
		m.Focus = !m.Focus
	case "m":
		f := m.currentFile()
		if f == nil || m.DiffCursor >= len(f.Hunks) {
			return m, nil
		}
		h := f.Hunks[m.DiffCursor]
		if m.Marked[m.markedKey(f.Path, h.ID)] {
			return m, m.unmarkHunk(f.Path, h.ID)
		}
		m.NoteTarget = noteTarget{File: f.Path, HunkID: h.ID}
		m.NoteInput.SetValue("")
		m.NoteInput.Focus()
		m.Screen = ScreenNotePrompt
	case "q":
		if m.allMarked() {
			m.Screen = ScreenSummary
			return m, nil
		}
		m.Screen = ScreenFileList
	}
	return m, nil
}

func (m Model) keyNote(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		text := strings.TrimSpace(m.NoteInput.Value())
		if text == "" {
			return m, nil
		}
		return m, m.markHunk(m.NoteTarget.File, m.NoteTarget.HunkID, text)
	case tea.KeyEsc:
		m.Screen = ScreenDiff
		m.NoteInput.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.NoteInput, cmd = m.NoteInput.Update(msg)
	return m, cmd
}

func (m Model) keySummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		return m, m.closeSession()
	case "n", "esc", "q":
		m.Screen = ScreenFileList
		return m, nil
	}
	return m, nil
}
