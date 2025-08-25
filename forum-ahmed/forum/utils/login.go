package utils

import (
	"errors"
	"log"
)

// Login checks if a user exists and optionally registers them.
// Returns the fully populated User struct.
func (db *DataBase) Login(username, email, password string) (User, error) {
	var user User

	// Query the user by username or email
	row := db.Conn.QueryRow(
		"SELECT uuid, username, email, password, notregistered FROM users WHERE username = ? OR email = ?",
		username, email,
	)

	hash, err := HashPassword(password)
	if err != nil {
		log.Println("Failed to hash password:", err)
		return User{}, err
	}
	password = hash

	// Scan the result into the User struct
	errScan := row.Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.NotRegistered)
	if errScan != nil {
		log.Println("Error scanning user:", errScan)
		return User{}, errScan
	}

	// User exists, check password
	if user.Password != password {
		return User{}, errors.New("invalid password")
	}

	// Login successful
	return user, nil
}
