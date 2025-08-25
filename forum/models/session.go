package models

import (
    "crypto/rand"
    "encoding/hex"
    "sync"
    "time"
)

type Session struct {
    Token   string
    UserID  int
    Expires time.Time
}

var (
    sessions     = make(map[string]Session)
    userSessions = make(map[int]string)
    mu           sync.RWMutex
)

func newToken() (string, error) {
    b := make([]byte, 16)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}

// CreateSession creates a new session for a user ensuring single session per user.
func CreateSession(userID int) (string, error) {
    token, err := newToken()
    if err != nil {
        return "", err
    }
    mu.Lock()
    defer mu.Unlock()
    if old, ok := userSessions[userID]; ok {
        delete(sessions, old)
    }
    s := Session{Token: token, UserID: userID, Expires: time.Now().Add(24 * time.Hour)}
    sessions[token] = s
    userSessions[userID] = token
    return token, nil
}

// GetSession retrieves a session by token.
func GetSession(token string) (Session, bool) {
    mu.RLock()
    defer mu.RUnlock()
    s, ok := sessions[token]
    if !ok || time.Now().After(s.Expires) {
        return Session{}, false
    }
    return s, true
}

// DeleteSession removes a session.
func DeleteSession(token string) {
    mu.Lock()
    defer mu.Unlock()
    if s, ok := sessions[token]; ok {
        delete(userSessions, s.UserID)
        delete(sessions, token)
    }
}
