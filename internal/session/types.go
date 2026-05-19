package session

import "time"

type Session struct {
	ID         int64
	ProjectDir string
	HeadSHA    string
	Hypothesis string
	StartedAt  time.Time
	ClosedAt   *time.Time
}

type HunkMark struct {
	SessionID int64
	File      string
	HunkID    string
	WhyNote   string
	MarkedAt  time.Time
}

type LineNote struct {
	SessionID int64
	File      string
	Line      int
	Note      string
	CreatedAt time.Time
}
