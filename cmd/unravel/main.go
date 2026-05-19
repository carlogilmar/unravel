package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/carlogilmar/unravel/internal/app"
	"github.com/carlogilmar/unravel/internal/config"
	"github.com/carlogilmar/unravel/internal/git"
	"github.com/carlogilmar/unravel/internal/session"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "unravel:", err)
		os.Exit(1)
	}
}

func run() error {
	target := "."
	if len(os.Args) > 1 {
		target = os.Args[1]
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return err
	}

	repo, err := git.Open(abs)
	if err != nil {
		return err
	}
	headSHA, err := repo.HeadSHA()
	if err != nil {
		return err
	}

	dbPath, err := config.DBPath()
	if err != nil {
		return err
	}
	store, err := session.Open(dbPath)
	if err != nil {
		return err
	}
	defer store.Close()

	m := app.New(repo.Root(), repo, store, headSHA)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
