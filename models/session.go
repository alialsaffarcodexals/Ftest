package models

import (
	"context"
	"errors"
	"log"
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
func CreateSingleSession(userID int64) (s *Session, err error) {
	log.Printf("models.CreateSingleSession start userID=%d", userID)
	defer func() { log.Printf("models.CreateSingleSession end err=%v", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultQueryTimeout)
	defer cancel()
	if _, err = DB.ExecContext(ctx, `DELETE FROM sessions WHERE user_id=?`, userID); err != nil {
		return nil, err
	}
	sid := uuid.New().String()
	exp := time.Now().Add(SessionTTL).UTC()
	_, err = DB.ExecContext(ctx, `INSERT INTO sessions(id, user_id, expires_at, created_at) VALUES(?,?,?,datetime('now'))`, sid, userID, exp.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	return &Session{ID: sid, UserID: userID, ExpiresAt: exp, CreatedAt: time.Now().UTC()}, nil
}

func SessionFromID(id string) (s *Session, err error) {
	log.Printf("models.SessionFromID start id=%s", id)
	defer func() { log.Printf("models.SessionFromID end err=%v", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultQueryTimeout)
	defer cancel()
	row := DB.QueryRowContext(ctx, `SELECT id, user_id, expires_at, created_at FROM sessions WHERE id=?`, id)
	s = &Session{}
	var expStr, cStr string
	if err = row.Scan(&s.ID, &s.UserID, &expStr, &cStr); err != nil {
		return nil, ErrNotFound
	}
	if s.ExpiresAt, err = time.Parse(time.RFC3339, expStr); err != nil {
		return nil, err
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, cStr)
	if time.Now().After(s.ExpiresAt) {
		_ = DeleteSession(id)
		return nil, errors.New("session expired")
	}
	return s, nil
}

func DeleteSession(id string) (err error) {
	log.Printf("models.DeleteSession start id=%s", id)
	defer func() { log.Printf("models.DeleteSession end err=%v", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultQueryTimeout)
	defer cancel()
	_, err = DB.ExecContext(ctx, `DELETE FROM sessions WHERE id=?`, id)
	return err
}
