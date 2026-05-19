package view

import "github.com/charmbracelet/lipgloss"

var (
	ColorAdd     = lipgloss.Color("#2ea043")
	ColorDel     = lipgloss.Color("#f85149")
	ColorContext = lipgloss.Color("#8b949e")
	ColorDim     = lipgloss.Color("#3a3f4b")
	ColorAccent  = lipgloss.Color("#58a6ff")
	ColorOK      = lipgloss.Color("#3fb950")
	ColorWarn    = lipgloss.Color("#d29922")

	Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#f0f6fc")).
		Padding(0, 1)

	Hypothesis = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Italic(true).
			Padding(0, 1)

	StatusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8b949e")).
			Background(lipgloss.Color("#161b22")).
			Padding(0, 1)

	FilePane = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color("#30363d")).
			Padding(0, 1)

	FileSelected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f0f6fc")).
			Background(lipgloss.Color("#1f6feb")).
			Bold(true)

	StatusAdded    = lipgloss.NewStyle().Foreground(ColorAdd).Bold(true)
	StatusModified = lipgloss.NewStyle().Foreground(ColorWarn).Bold(true)
	StatusDeleted  = lipgloss.NewStyle().Foreground(ColorDel).Bold(true)
	StatusRenamed  = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)

	DiffAdd     = lipgloss.NewStyle().Foreground(ColorAdd)
	DiffDel     = lipgloss.NewStyle().Foreground(ColorDel)
	DiffContext = lipgloss.NewStyle().Foreground(ColorContext)
	DiffDimmed  = lipgloss.NewStyle().Foreground(ColorDim)
	DiffHunkHdr = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	DiffCurrent = lipgloss.NewStyle().Background(lipgloss.Color("#1f2937"))

	MarkedTag   = lipgloss.NewStyle().Foreground(ColorOK).Bold(true)
	UnmarkedTag = lipgloss.NewStyle().Foreground(ColorWarn)

	Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6e7681")).
		Padding(0, 1)

	ErrorBox = lipgloss.NewStyle().
			Foreground(ColorDel).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDel).
			Padding(1, 2)
)
