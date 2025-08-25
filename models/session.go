package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string
	UserID    int64
	ExpiresAt time.Time
	CreatedAt time.Time
}

const SessionTTL = 60 * time.Minute

// CreateSingleSession deletes any existing sessions for user, then creates a new one.
func CreateSingleSession(userID int64) (*Session, error) {
	if _, err := DB.Exec(`DELETE FROM sessions WHERE user_id=?`, userID); err != nil {
		return nil, err
	}
	sid := uuid.New().String()
	exp := time.Now().Add(SessionTTL).UTC()
	_, err := DB.Exec(`INSERT INTO sessions(id, user_id, expires_at, created_at) VALUES(?,?,?,datetime('now'))`,
		sid, userID, exp.Format(time.RFC3339))
	if err != nil { return nil, err }
	return &Session{ID: sid, UserID: userID, ExpiresAt: exp, CreatedAt: time.Now().UTC()}, nil
}

func SessionFromID(id string) (*Session, error) {
	row := DB.QueryRow(`SELECT id, user_id, expires_at, created_at FROM sessions WHERE id=?`, id)
	s := &Session{}
	var expStr, cStr string
	if err := row.Scan(&s.ID, &s.UserID, &expStr, &cStr); err != nil {
		return nil, ErrNotFound
	}
	var err error
	s.ExpiresAt, err = time.Parse(time.RFC3339, expStr)
	if err != nil { return nil, err }
	s.CreatedAt, _ = time.Parse(time.RFC3339, cStr)
	if time.Now().After(s.ExpiresAt) {
		_ = DeleteSession(id)
		return nil, errors.New("session expired")
	}
	return s, nil
}

func DeleteSession(id string) error {
	_, err := DB.Exec(`DELETE FROM sessions WHERE id=?`, id)
	return err
}
