package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/carlogilmar/unravel/internal/git"
	"github.com/carlogilmar/unravel/internal/view"
)

func (m Model) View() string {
	switch m.Screen {
	case ScreenHypothesis:
		return m.viewHypothesis()
	case ScreenFileList:
		return m.viewFileList()
	case ScreenDiff:
		return m.viewDiff()
	case ScreenNotePrompt:
		return m.viewNotePrompt()
	case ScreenSummary:
		return m.viewSummary()
	case ScreenFatal:
		return m.viewFatal()
	}
	return ""
}

func (m Model) viewHypothesis() string {
	title := view.Header.Render("Unravel — start a review session")
	prompt := "Before reading any code, write one sentence:\n"
	hint := view.Help.Render("Enter to start  ·  Ctrl+C to quit")
	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		view.Hypothesis.Render(prompt),
		m.HypothesisInput.View(),
		"",
		hint,
	)
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
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

func (m Model) viewNotePrompt() string {
	hdr := m.headerBar()
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(view.ColorAccent).
		Padding(1, 2).
		Width(80)
	prompt := "Modification test — answer in one sentence:\n"
	help := view.Help.Render("Enter to save  ·  Esc to cancel")
	body := box.Render(lipgloss.JoinVertical(lipgloss.Left,
		view.Hypothesis.Render(prompt),
		m.NoteInput.View(),
		"",
		help,
	))
	return joinFull(m.Height, hdr, lipgloss.NewStyle().Padding(2, 2).Render(body), m.statusBar())
}

func (m Model) viewSummary() string {
	hdr := m.headerBar()
	lines := []string{
		view.Header.Render("Close session?"),
		"",
		fmt.Sprintf("Hypothesis:  %s", m.Sess.Hypothesis),
		fmt.Sprintf("Files:       %d", len(m.Files)),
		fmt.Sprintf("Hunks read:  %d / %d", m.markedCount(), m.totalHunks()),
		"",
		"What happens after this is outside Unravel's scope.",
		"Commit, push, PR, walk away — your call.",
		"",
		view.Help.Render("y / Enter — close   ·   n / Esc — back to review"),
	}
	body := lipgloss.NewStyle().Padding(2, 4).Render(strings.Join(lines, "\n"))
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
		hyp = view.Hypothesis.Render("hypothesis: " + m.Sess.Hypothesis)
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
	}
	help := m.helpForScreen()
	left := fmt.Sprintf(" [%s]  %s ", mode, progress)
	return view.StatusBar.Width(m.Width).Render(left + "  " + help)
}

func (m Model) helpForScreen() string {
	switch m.Screen {
	case ScreenFileList:
		if m.allMarked() {
			return "j/k move · enter open · q close session"
		}
		return "j/k move · enter open · q quit"
	case ScreenDiff:
		return "j/k hunk · f focus · m mark · esc back"
	case ScreenNotePrompt:
		return "enter save · esc cancel"
	case ScreenSummary:
		return "y close · n back"
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
		marked := m.Marked[m.markedKey(f.Path, h.ID)]
		tag := view.UnmarkedTag.Render("◯")
		if marked {
			tag = view.MarkedTag.Render("✓")
		}
		header := fmt.Sprintf("%s  Hunk %d/%d  %s", tag, i+1, len(f.Hunks), h.Header)
		if i == m.DiffCursor {
			header = view.DiffCurrent.Render(view.DiffHunkHdr.Render(header))
		} else {
			header = view.DiffHunkHdr.Render(header)
		}
		b.WriteString(header + "\n")

		if m.Focus && i != m.DiffCursor {
			b.WriteString(view.DiffDimmed.Render("  ··· dimmed ···") + "\n\n")
			continue
		}
		b.WriteString(m.renderHunk(h, m.Focus && i == m.DiffCursor, width))
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) renderHunk(h git.Hunk, focus bool, width int) string {
	var b strings.Builder
	for _, l := range h.Lines {
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
		num := ""
		if l.NewNum > 0 {
			num = fmt.Sprintf("%4d ", l.NewNum)
		} else {
			num = "     "
		}
		text := style.Render(prefix + l.Content)
		line := view.DiffDimmed.Render(num) + text
		if width > 0 {
			line = lipgloss.NewStyle().MaxWidth(width).Render(line)
		}
		b.WriteString(line + "\n")
	}
	if focus {
		// reserved: future per-line cursor inside focus mode
		_ = focus
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
