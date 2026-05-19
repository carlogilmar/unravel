package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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
		if m.FileCursor >= len(m.Files) {
			m.FileCursor = 0
		}
		m.DiffCursor = 0
		if m.Screen != ScreenDiff && m.Screen != ScreenSummary && m.Screen != ScreenTitleEdit {
			m.Screen = ScreenFileList
		}
		return m, nil

	case HunkMarkedMsg:
		if msg.Err == nil {
			key := m.markedKey(msg.File, msg.HunkID)
			m.Marked[key] = true
			m.Notes[key] = msg.Why
		}
		m.InlineNote = false
		m.NoteInput.SetValue("")
		m.NoteInput.Blur()
		return m, nil

	case HunkUnmarkedMsg:
		if msg.Err == nil {
			key := m.markedKey(msg.File, msg.HunkID)
			delete(m.Marked, key)
			delete(m.Notes, key)
		}
		return m, nil

	case TitleUpdatedMsg:
		if msg.Err == nil && m.Sess != nil {
			m.Sess.Hypothesis = msg.Title
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
	case ScreenFileList:
		return m.keyFileList(msg)
	case ScreenDiff:
		return m.keyDiff(msg)
	case ScreenTitleEdit:
		return m.keyTitleEdit(msg)
	case ScreenSummary:
		return m.keySummary(msg)
	case ScreenConfirmClose:
		return m.keyConfirmClose(msg)
	case ScreenFatal:
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) keyFileList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "s":
		m.Prev = ScreenFileList
		m.Screen = ScreenSummary
		return m, nil
	case "t":
		return m.openTitleEdit(ScreenFileList), textinput.Blink
	case "j", "down":
		if m.FileCursor < len(m.Files)-1 {
			m.FileCursor++
			m.DiffCursor = 0
		}
	case "k", "up":
		if m.FileCursor > 0 {
			m.FileCursor--
			m.DiffCursor = 0
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
	if m.InlineNote {
		return m.keyInlineNote(msg)
	}

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
	case " ", "enter":
		f := m.currentFile()
		if f != nil && m.DiffCursor < len(f.Hunks) {
			key := m.markedKey(f.Path, f.Hunks[m.DiffCursor].ID)
			m.Expanded[key] = !m.Expanded[key]
		}
	case "d":
		m.ShowFullDiff = !m.ShowFullDiff
	case "f":
		m.Focus = !m.Focus
	case "s":
		m.Prev = ScreenDiff
		m.Screen = ScreenSummary
	case "t":
		return m.openTitleEdit(ScreenDiff), textinput.Blink
	case "m":
		f := m.currentFile()
		if f == nil || m.DiffCursor >= len(f.Hunks) {
			return m, nil
		}
		h := f.Hunks[m.DiffCursor]
		key := m.markedKey(f.Path, h.ID)
		m.NoteTarget = noteTarget{File: f.Path, HunkID: h.ID}
		m.NoteInput.SetValue(m.Notes[key])
		m.NoteInput.CursorEnd()
		m.NoteInput.Focus()
		m.InlineNote = true
		m.Expanded[key] = true
		return m, textinput.Blink
	case "u":
		f := m.currentFile()
		if f == nil || m.DiffCursor >= len(f.Hunks) {
			return m, nil
		}
		h := f.Hunks[m.DiffCursor]
		if m.Marked[m.markedKey(f.Path, h.ID)] {
			return m, m.unmarkHunk(f.Path, h.ID)
		}
		return m, nil
	case "q":
		m.Screen = ScreenFileList
	}
	return m, nil
}

func (m Model) keyInlineNote(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		text := strings.TrimSpace(m.NoteInput.Value())
		if text == "" {
			return m, nil
		}
		return m, m.markHunk(m.NoteTarget.File, m.NoteTarget.HunkID, text)
	case tea.KeyEsc:
		m.InlineNote = false
		m.NoteInput.SetValue("")
		m.NoteInput.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.NoteInput, cmd = m.NoteInput.Update(msg)
	return m, cmd
}

func (m Model) openTitleEdit(prev Screen) Model {
	m.Prev = prev
	current := defaultTitle
	if m.Sess != nil && m.Sess.Hypothesis != "" {
		current = m.Sess.Hypothesis
	}
	m.TitleInput.SetValue(current)
	m.TitleInput.CursorEnd()
	m.TitleInput.Focus()
	m.Screen = ScreenTitleEdit
	return m
}

func (m Model) keyTitleEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		text := strings.TrimSpace(m.TitleInput.Value())
		if text == "" {
			return m, nil
		}
		m.TitleInput.Blur()
		m.Screen = m.Prev
		if m.Sess != nil {
			m.Sess.Hypothesis = text
			return m, m.updateTitle(text)
		}
		return m, nil
	case tea.KeyEsc:
		m.TitleInput.Blur()
		m.Screen = m.Prev
		return m, nil
	}
	var cmd tea.Cmd
	m.TitleInput, cmd = m.TitleInput.Update(msg)
	return m, cmd
}

func (m Model) keySummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		if m.allMarked() {
			m.Screen = ScreenConfirmClose
		}
		return m, nil
	case "n", "esc", "q":
		m.Screen = m.Prev
		if m.Screen == ScreenSummary {
			m.Screen = ScreenFileList
		}
		return m, nil
	}
	return m, nil
}

func (m Model) keyConfirmClose(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		return m, m.closeSession()
	case "n", "esc", "q":
		m.Screen = ScreenSummary
		return m, nil
	}
	return m, nil
}
