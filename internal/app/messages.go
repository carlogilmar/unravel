package app

import (
	"github.com/carlogilmar/unravel/internal/git"
	"github.com/carlogilmar/unravel/internal/session"
)

type DiffLoadedMsg struct {
	Files []git.FileChange
	Err   error
}

type SessionStartedMsg struct {
	Sess *session.Session
	Err  error
}

type HunkMarkedMsg struct {
	File   string
	HunkID string
	Why    string
	Err    error
}

type HunkUnmarkedMsg struct {
	File   string
	HunkID string
	Err    error
}

type SessionClosedMsg struct {
	Err error
}

type TitleUpdatedMsg struct {
	Title string
	Err   error
}

type FatalErrMsg struct{ Err error }
