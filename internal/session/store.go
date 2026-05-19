package session

import (
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_dir TEXT NOT NULL,
			head_sha   TEXT NOT NULL,
			hypothesis TEXT NOT NULL,
			started_at INTEGER NOT NULL,
			closed_at  INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS hunk_marks (
			session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			file       TEXT NOT NULL,
			hunk_id    TEXT NOT NULL,
			why_note   TEXT NOT NULL,
			marked_at  INTEGER NOT NULL,
			PRIMARY KEY (session_id, file, hunk_id)
		)`,
		`CREATE TABLE IF NOT EXISTS line_notes (
			session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			file       TEXT NOT NULL,
			line       INTEGER NOT NULL,
			note       TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			PRIMARY KEY (session_id, file, line)
		)`,
	}
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) StartSession(projectDir, headSHA, hypothesis string) (*Session, error) {
	now := time.Now()
	res, err := s.db.Exec(
		`INSERT INTO sessions (project_dir, head_sha, hypothesis, started_at) VALUES (?, ?, ?, ?)`,
		projectDir, headSHA, hypothesis, now.Unix(),
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &Session{ID: id, ProjectDir: projectDir, HeadSHA: headSHA, Hypothesis: hypothesis, StartedAt: now}, nil
}

func (s *Store) CloseSession(id int64) error {
	now := time.Now()
	res, err := s.db.Exec(`UPDATE sessions SET closed_at = ? WHERE id = ? AND closed_at IS NULL`, now.Unix(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("session already closed or not found")
	}
	return nil
}

func (s *Store) UpdateHypothesis(sessionID int64, hypothesis string) error {
	_, err := s.db.Exec(`UPDATE sessions SET hypothesis = ? WHERE id = ?`, hypothesis, sessionID)
	return err
}

func (s *Store) MarkHunk(sessionID int64, file, hunkID, whyNote string) error {
	if whyNote == "" {
		return errors.New("why note required")
	}
	_, err := s.db.Exec(
		`INSERT INTO hunk_marks (session_id, file, hunk_id, why_note, marked_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(session_id, file, hunk_id) DO UPDATE SET why_note = excluded.why_note, marked_at = excluded.marked_at`,
		sessionID, file, hunkID, whyNote, time.Now().Unix(),
	)
	return err
}

func (s *Store) UnmarkHunk(sessionID int64, file, hunkID string) error {
	_, err := s.db.Exec(
		`DELETE FROM hunk_marks WHERE session_id = ? AND file = ? AND hunk_id = ?`,
		sessionID, file, hunkID,
	)
	return err
}

func (s *Store) IsMarked(sessionID int64, file, hunkID string) (bool, string, error) {
	var note string
	err := s.db.QueryRow(
		`SELECT why_note FROM hunk_marks WHERE session_id = ? AND file = ? AND hunk_id = ?`,
		sessionID, file, hunkID,
	).Scan(&note)
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, note, nil
}

func (s *Store) MarkedHunks(sessionID int64) (map[string]bool, error) {
	rows, err := s.db.Query(`SELECT file, hunk_id FROM hunk_marks WHERE session_id = ?`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var file, hunkID string
		if err := rows.Scan(&file, &hunkID); err != nil {
			return nil, err
		}
		out[file+"\x00"+hunkID] = true
	}
	return out, rows.Err()
}

func (s *Store) SaveLineNote(sessionID int64, file string, line int, note string) error {
	_, err := s.db.Exec(
		`INSERT INTO line_notes (session_id, file, line, note, created_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(session_id, file, line) DO UPDATE SET note = excluded.note, created_at = excluded.created_at`,
		sessionID, file, line, note, time.Now().Unix(),
	)
	return err
}
