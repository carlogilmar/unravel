package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/carlogilmar/unravel/internal/git"
)

func (m Model) loadDiff() tea.Cmd {
	repo := m.Repo
	return func() tea.Msg {
		files, err := repo.ChangedFiles([]string{".ex", ".exs", ".py"})
		return DiffLoadedMsg{Files: files, Err: err}
	}
}

func (m Model) startSession(hypothesis string) tea.Cmd {
	store := m.Store
	dir := m.ProjectDir
	sha := m.HeadSHA
	return func() tea.Msg {
		s, err := store.StartSession(dir, sha, hypothesis)
		return SessionStartedMsg{Sess: s, Err: err}
	}
}

func (m Model) markHunk(file, hunkID, why string) tea.Cmd {
	store := m.Store
	sessID := m.Sess.ID
	return func() tea.Msg {
		err := store.MarkHunk(sessID, file, hunkID, why)
		return HunkMarkedMsg{File: file, HunkID: hunkID, Why: why, Err: err}
	}
}

func (m Model) unmarkHunk(file, hunkID string) tea.Cmd {
	store := m.Store
	sessID := m.Sess.ID
	return func() tea.Msg {
		err := store.UnmarkHunk(sessID, file, hunkID)
		return HunkUnmarkedMsg{File: file, HunkID: hunkID, Err: err}
	}
}

func (m Model) updateTitle(title string) tea.Cmd {
	store := m.Store
	sessID := m.Sess.ID
	return func() tea.Msg {
		return TitleUpdatedMsg{Title: title, Err: store.UpdateHypothesis(sessID, title)}
	}
}

func (m Model) closeSession() tea.Cmd {
	store := m.Store
	sessID := m.Sess.ID
	return func() tea.Msg {
		return SessionClosedMsg{Err: store.CloseSession(sessID)}
	}
}

// keep imports honest
var _ = git.StatusAdded
