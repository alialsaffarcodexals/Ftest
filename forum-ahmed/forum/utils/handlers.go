package utils

import (
	"database/sql"
	"errors"
	"html/template"
	"log"
	"net/http"
)

var tpl *template.Template

// InitTemplate parses and executes a template
func InitTemplate(w http.ResponseWriter, file string, data interface{}) {
	var err error
	tpl, err = template.ParseFiles(file)
	if err != nil {
		http.Error(w, "Template parsing error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tpl.Execute(w, data); err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// DefaultHandler redirects "/" to "/login"
func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// GuestHandler handles guest login (not registered user)
func GuestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		user, err := db.Guest(w, r)
		if err != nil {
			http.Error(w, "Failed to create guest: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Store user UUID in a cookie
		SetUserCookie(w, user.UUID)

		// Redirect to home
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// If not POST → just show login.html
	InitTemplate(w, "templates/home.html", nil)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Authenticate
		user, err := db.Login(username, email, password)
		if err != nil {
			http.Error(w, "Login failed: "+err.Error(), http.StatusBadRequest)
			RenderError(w, "login failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Store cookie
		SetUserCookie(w, user.UUID)

		// Redirect (doesn't show POST response to user)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodGet {
		// Show login form
		InitTemplate(w, "templates/login.html", nil)
		return
	}

	RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HomeHandler requires a valid cookie
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Get UUID from cookie
	uuid, err := GetUserFromCookie(r)
	if err != nil || uuid == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	InitTemplate(w, "templates/home.html", map[string]string{"UUID": uuid})
}

func (db *DataBase) Guest(w http.ResponseWriter, r *http.Request) (*User, error) {
	uuid, err := GenerateUserID()
	if err != nil {
		return nil, err
	}
	// User does not exist, create new user
	user := User{
		UUID:          uuid,
		NotRegistered: true,
		Username:      "guest" + uuid,
		Email:         "",
		Password:      "",
	}

	// Insert safely using SafeWriter
	if err := db.SafeWriter("users", user); err != nil {
		return nil, err
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "user",
		Value: user.UUID,
		Path:  "/",
	})

	http.Redirect(w, r, "/home", http.StatusSeeOther)
	return &user, nil
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		// Validate form fields
		if username == "" || email == "" || password == "" || confirmPassword == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			RenderError(w, "All fields are required", http.StatusBadRequest)
			return
		}

		if password != confirmPassword {
			http.Error(w, "Passwords do not match", http.StatusBadRequest)
			RenderError(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		// Register user
		user, err := db.Register(w, username, email, password)
		if err != nil {
			http.Error(w, "Registration failed: "+err.Error(), http.StatusBadRequest)
			RenderError(w, "Registration failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Store cookie
		SetUserCookie(w, user.UUID)

		// Redirect to home
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// Show registration form
	InitTemplate(w, "templates/register.html", nil)
}

func (db *DataBase) Register(w http.ResponseWriter, username, email, password string) (*User, error) {
	uuid, err := GenerateUserID()
	if err != nil {
		return nil, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		log.Println("Failed to hash password:", err)
		RenderError(w, "Internal server error", http.StatusInternalServerError)
		return nil, err
	}
	password = hash

	// Check if user already exists
	var existing User
	err = db.Conn.QueryRow("SELECT uuid FROM users WHERE username = ? OR email = ?", username, email).Scan(&existing.UUID)
	if err != sql.ErrNoRows {
		if err != nil {
			log.Println("Database error:", err)
			RenderError(w, "Internal server error", http.StatusInternalServerError)
			return nil, err
		}
		return nil, errors.New("user with this username or email already exists")
	}

	// Create new user
	user := User{
		UUID:          uuid,
		NotRegistered: false,
		Username:      username,
		Email:         email,
		Password:      password,
	}

	// Insert safely using SafeWriter
	if err := db.SafeWriter("users", user); err != nil {
		log.Println("Failed to insert user:", err)
		RenderError(w, "Internal server error", http.StatusInternalServerError)
		return nil, err
	}

	return &user, nil
}
