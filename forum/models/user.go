package models

import (
    "errors"
    "regexp"

    "golang.org/x/crypto/bcrypt"
)

// User represents a registered user.
type User struct {
    ID       int
    Email    string
    Username string
    Password string
}

// validPassword checks for a simple strong password.
func validPassword(p string) bool {
    if len(p) < 8 {
        return false
    }
    hasNum, _ := regexp.MatchString(`[0-9]`, p)
    hasLetter, _ := regexp.MatchString(`[A-Za-z]`, p)
    return hasNum && hasLetter
}

// CreateUser registers a new user with bcrypt password.
func CreateUser(email, username, password string) error {
    if len(username) < 6 || !validPassword(password) {
        return errors.New("invalid credentials")
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    _, err = DB.Exec(`INSERT INTO users(email, username, password) VALUES(?,?,?)`, email, username, string(hash))
    return err
}

// GetUserByUsername returns a user by username.
func GetUserByUsername(username string) (*User, error) {
    u := &User{}
    err := DB.QueryRow(`SELECT id, email, username, password FROM users WHERE username = ?`, username).Scan(&u.ID, &u.Email, &u.Username, &u.Password)
    if err != nil {
        return nil, err
    }
    return u, nil
}

// GetUserByID returns user by id.
func GetUserByID(id int) (*User, error) {
    u := &User{}
    err := DB.QueryRow(`SELECT id, email, username, password FROM users WHERE id = ?`, id).Scan(&u.ID, &u.Email, &u.Username, &u.Password)
    if err != nil {
        return nil, err
    }
    return u, nil
}

// Authenticate checks user credentials.
func Authenticate(username, password string) (*User, error) {
    u, err := GetUserByUsername(username)
    if err != nil {
        return nil, err
    }
    if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
        return nil, errors.New("invalid password")
    }
    return u, nil
}
