package models

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64
	Email        string
	Username     string
	PasswordHash string
	CreatedAt    string
}

// CreateUser inserts a new user (email + username unique). Password will be hashed.
func CreateUser(email, username, password string) (*User, error) {
	if email == "" || username == "" || password == "" {
		return nil, errors.New("missing fields")
	}
	// check duplicates
	var exists int
	row := DB.QueryRow(`SELECT 1 FROM users WHERE email=? OR username=? LIMIT 1`, email, username)
	if err := row.Scan(&exists); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if exists == 1 {
		return nil, errors.New("email or username taken")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil { return nil, err }

	res, err := DB.Exec(`INSERT INTO users(email, username, password_hash, created_at) VALUES(?,?,?,datetime('now'))`,
		email, username, string(hash))
	if err != nil { return nil, err }
	id, _ := res.LastInsertId()
	return &User{ID: id, Email: email, Username: username, PasswordHash: string(hash)}, nil
}

func GetUserByID(id int64) (*User, error) {
	row := DB.QueryRow(`SELECT id, email, username, password_hash, created_at FROM users WHERE id=?`, id)
	u := &User{}
	if err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows { return nil, ErrNotFound }
		return nil, err
	}
	return u, nil
}

func FindUserByLogin(login string) (*User, error) {
	row := DB.QueryRow(`SELECT id, email, username, password_hash, created_at
		FROM users WHERE email=? OR username=?`, login, login)
	u := &User{}
	if err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows { return nil, ErrNotFound }
		return nil, err
	}
	return u, nil
}

func Authenticate(login, password string) (*User, error) {
	u, err := FindUserByLogin(login)
	if err != nil { return nil, err }
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}
