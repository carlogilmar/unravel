package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/carlogilmar/unravel/internal/git"
	"github.com/carlogilmar/unravel/internal/view"
)

func (m Model) View() string {
	if m.Screen == ScreenFatal {
		return m.viewFatal()
	}
	if m.Sess == nil {
		return m.viewLoading()
	}
	switch m.Screen {
	case ScreenFileList:
		return m.viewFileList()
	case ScreenDiff:
		return m.viewDiff()
	case ScreenTitleEdit:
		return m.viewTitleEdit()
	case ScreenSummary:
		return m.viewSummary()
	case ScreenConfirmClose:
		return m.viewConfirmClose()
	}
	return ""
}

func (m Model) viewConfirmClose() string {
	hdr := m.headerBar()
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(view.ColorWarn).
		Padding(1, 2).
		Width(70)
	lines := []string{
		view.Header.Render("Close this session?"),
		"",
		fmt.Sprintf("Title:      %s", m.Sess.Hypothesis),
		fmt.Sprintf("Files:      %d", len(m.Files)),
		fmt.Sprintf("Hunks:      %d / %d understood", m.markedCount(), m.totalHunks()),
		"",
		"Heads up — every note you wrote lives only inside this session.",
		"Once you close it, the current notes cannot be recovered from",
		"within Unravel. There is no undo.",
		"",
		view.Help.Render("y / Enter — yes, close   ·   n / Esc — no, go back"),
	}
	body := box.Render(strings.Join(lines, "\n"))
	return joinFull(m.Height, hdr, lipgloss.NewStyle().Padding(2, 2).Render(body), m.statusBar())
}

func (m Model) viewLoading() string {
	body := lipgloss.NewStyle().Padding(2, 2).Render(view.Hypothesis.Render("Loading session…"))
	return body
}

func (m Model) viewFileList() string {
	hdr := m.headerBar()
	body := m.renderFileList(m.Width)
	footer := m.statusBar()
	return joinFull(m.Height, hdr, body, footer)
}

func (m Model) viewDiff() string {
	hdr := m.headerBar()
	left := m.renderFileList(28)
	right := m.renderDiff(m.Width - 30)
	body := lipgloss.JoinHorizontal(lipgloss.Top, view.FilePane.Render(left), right)
	footer := m.statusBar()
	return joinFull(m.Height, hdr, body, footer)
}

func (m Model) viewTitleEdit() string {
	hdr := m.headerBar()
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(view.ColorAccent).
		Padding(1, 2).
		Width(70)
	prompt := "Update session title:"
	help := view.Help.Render("Enter to save  ·  Esc to cancel")
	body := box.Render(lipgloss.JoinVertical(lipgloss.Left,
		view.Hypothesis.Render(prompt),
		m.TitleInput.View(),
		"",
		help,
	))
	return joinFull(m.Height, hdr, lipgloss.NewStyle().Padding(2, 2).Render(body), m.statusBar())
}

func (m Model) viewSummary() string {
	hdr := m.headerBar()
	var lines []string
	lines = append(lines, view.Header.Render("Session summary"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Title:      %s", m.Sess.Hypothesis))
	lines = append(lines, fmt.Sprintf("Files:      %d", len(m.Files)))
	lines = append(lines, fmt.Sprintf("Progress:   %d / %d hunks understood", m.markedCount(), m.totalHunks()))
	lines = append(lines, "")

	if len(m.Files) == 0 {
		lines = append(lines, view.Help.Render("No files in the working tree."))
	}

	for _, f := range m.Files {
		status := statusStyle(f.Status).Render(string(f.Status))
		lines = append(lines, fmt.Sprintf("%s  %s", status, view.Header.Render(f.Path)))
		for i, h := range f.Hunks {
			key := m.markedKey(f.Path, h.ID)
			marker := view.UnmarkedTag.Render("◯")
			if m.Marked[key] {
				marker = view.MarkedTag.Render("✓")
			}
			hunkLine := fmt.Sprintf("    %s  Hunk %d/%d  %s", marker, i+1, len(f.Hunks),
				view.DiffHunkHdr.Render(h.Header))
			lines = append(lines, hunkLine)
			if note := m.Notes[key]; note != "" {
				lines = append(lines, view.Hypothesis.Render("        — "+note))
			}
		}
		lines = append(lines, "")
	}

	if m.allMarked() {
		lines = append(lines, view.Help.Render("y / Enter — close session   ·   n / Esc — back"))
	} else {
		remaining := m.totalHunks() - m.markedCount()
		lines = append(lines, view.Help.Render(fmt.Sprintf("%d hunk(s) still need a why-note before you can close   ·   n / Esc — back", remaining)))
	}

	body := lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(lines, "\n"))
	return joinFull(m.Height, hdr, body, m.statusBar())
}

func (m Model) viewFatal() string {
	box := view.ErrorBox.Render(fmt.Sprintf("Fatal: %v", m.FatalErr))
	hint := view.Help.Render("Any key to quit")
	return lipgloss.NewStyle().Padding(2, 2).Render(lipgloss.JoinVertical(lipgloss.Left, box, "", hint))
}

func (m Model) headerBar() string {
	title := "Unravel"
	hyp := ""
	if m.Sess != nil {
		hyp = view.Hypothesis.Render("title: " + m.Sess.Hypothesis)
	}
	return view.Header.Render(title) + " " + hyp
}

func (m Model) statusBar() string {
	total := m.totalHunks()
	marked := m.markedCount()
	progress := fmt.Sprintf("%d/%d hunks understood", marked, total)
	mode := "browse"
	if m.Screen == ScreenDiff {
		mode = "read"
		if m.Focus {
			mode = "focus"
		}
		if m.InlineNote {
			mode = "note"
		}
	}
	if m.ShowFullDiff && m.Screen == ScreenDiff {
		mode += "·full"
	}
	help := m.helpForScreen()
	left := fmt.Sprintf(" [%s]  %s ", mode, progress)
	return view.StatusBar.Width(m.Width).Render(left + "  " + help)
}

func (m Model) helpForScreen() string {
	switch m.Screen {
	case ScreenFileList:
		return "j/k move · enter open · t title · s summary · q quit"
	case ScreenDiff:
		if m.InlineNote {
			return "enter save · esc cancel"
		}
		return "j/k hunk · space expand · d full/added · m note · u unmark · f focus · s summary · t title · esc back"
	case ScreenTitleEdit:
		return "enter save · esc cancel"
	case ScreenSummary:
		if m.allMarked() {
			return "y close · n back"
		}
		return "n back"
	case ScreenConfirmClose:
		return "y confirm close · n back"
	}
	return ""
}

func (m Model) renderFileList(width int) string {
	if len(m.Files) == 0 {
		return view.Hypothesis.Render("No matching changes in the working tree.\n\nNothing to review.")
	}
	var b strings.Builder
	curStatus := git.Status("")
	for i, f := range m.Files {
		if f.Status != curStatus {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(statusStyle(f.Status).Render(string(f.Status)) + "\n")
			curStatus = f.Status
		}
		marked := m.fileMarkedCount(f) == len(f.Hunks) && len(f.Hunks) > 0
		var tag string
		if marked {
			tag = view.MarkedTag.Render("✓ ")
		} else if m.fileMarkedCount(f) > 0 {
			tag = view.UnmarkedTag.Render("◐ ")
		} else {
			tag = "  "
		}
		line := fmt.Sprintf("%s%s  (%d hunks)", tag, f.Path, len(f.Hunks))
		if i == m.FileCursor {
			line = view.FileSelected.Render("> " + line)
		} else {
			line = "  " + line
		}
		if width > 0 {
			line = lipgloss.NewStyle().MaxWidth(width).Render(line)
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (m Model) fileMarkedCount(f git.FileChange) int {
	n := 0
	for _, h := range f.Hunks {
		if m.Marked[m.markedKey(f.Path, h.ID)] {
			n++
		}
	}
	return n
}

func (m Model) renderDiff(width int) string {
	f := m.currentFile()
	if f == nil {
		return ""
	}
	if len(f.Hunks) == 0 {
		return view.Help.Render("(no diff hunks for this file)")
	}
	if m.DiffCursor >= len(f.Hunks) {
		m.DiffCursor = len(f.Hunks) - 1
	}

	var b strings.Builder
	b.WriteString(view.Header.Render(f.Path) + "\n\n")

	for i, h := range f.Hunks {
		key := m.markedKey(f.Path, h.ID)
		marked := m.Marked[key]
		expanded := m.Expanded[key]

		tag := view.UnmarkedTag.Render("◯")
		if marked {
			tag = view.MarkedTag.Render("✓")
		}
		header := fmt.Sprintf("%s  Hunk %d/%d  %s", tag, i+1, len(f.Hunks), h.Header)
		if note := m.Notes[key]; note != "" {
			header += "  " + view.Hypothesis.Render("— "+note)
		}
		if i == m.DiffCursor {
			header = view.DiffCurrent.Render(view.DiffHunkHdr.Render(header))
		} else {
			header = view.DiffHunkHdr.Render(header)
		}
		b.WriteString(header + "\n")

		if m.InlineNote && i == m.DiffCursor {
			b.WriteString("    " + view.Hypothesis.Render("why → ") + m.NoteInput.View() + "\n")
		}

		focusHide := m.Focus && i != m.DiffCursor
		if !expanded || focusHide {
			if focusHide {
				b.WriteString("  " + view.DiffDimmed.Render("··· dimmed ···") + "\n")
			}
			b.WriteString("\n")
			continue
		}
		b.WriteString(m.renderHunk(h, width, m.ShowFullDiff))
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) renderHunk(h git.Hunk, width int, showFull bool) string {
	var b strings.Builder
	for _, l := range h.Lines {
		if !showFull && l.Kind != git.LineAdd {
			continue
		}
		var prefix string
		var style lipgloss.Style
		switch l.Kind {
		case git.LineAdd:
			prefix = "+ "
			style = view.DiffAdd
		case git.LineDel:
			prefix = "- "
			style = view.DiffDel
		default:
			prefix = "  "
			style = view.DiffContext
		}
		num := "     "
		if l.NewNum > 0 {
			num = fmt.Sprintf("%4d ", l.NewNum)
		}
		text := style.Render(prefix + l.Content)
		line := view.DiffDimmed.Render(num) + text
		if width > 0 {
			line = lipgloss.NewStyle().MaxWidth(width).Render(line)
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func statusStyle(s git.Status) lipgloss.Style {
	switch s {
	case git.StatusAdded:
		return view.StatusAdded
	case git.StatusModified:
		return view.StatusModified
	case git.StatusDeleted:
		return view.StatusDeleted
	case git.StatusRenamed:
		return view.StatusRenamed
	}
	return lipgloss.NewStyle()
}

func joinFull(height int, hdr, body, footer string) string {
	parts := []string{hdr, body}
	if height > 0 {
		used := lipgloss.Height(hdr) + lipgloss.Height(body) + lipgloss.Height(footer)
		if pad := height - used; pad > 0 {
			parts = append(parts, strings.Repeat("\n", pad-1))
		}
	}
	parts = append(parts, footer)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
