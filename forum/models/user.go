package models

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Email    string
	Username string
	Password string
}

func CreateUser(db *sql.DB, email, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, hash)
	return err
}

func AuthenticateUser(db *sql.DB, username, password string) (User, error) {
	var u User
	err := db.QueryRow("SELECT id, email, password FROM users WHERE username = ?", username).Scan(&u.ID, &u.Email, &u.Password)
	if err != nil {
		return u, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return u, err
	}
	u.Username = username
	return u, nil
}

func GetUserByID(db *sql.DB, id int) (User, error) {
	var u User
	err := db.QueryRow("SELECT id, email, username FROM users WHERE id = ?", id).Scan(&u.ID, &u.Email, &u.Username)
	return u, err
}
