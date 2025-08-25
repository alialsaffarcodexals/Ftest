package handlers

import (
    "net/http"
    "time"

    "forum/models"
)

// Register handles user registration.
func Register(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        render(w, "register.html", nil)
    case http.MethodPost:
        email := r.FormValue("email")
        username := r.FormValue("username")
        password := r.FormValue("password")
        if err := models.CreateUser(email, username, password); err != nil {
            render(w, "register.html", map[string]interface{}{"Message": err.Error(), "Error": true})
            return
        }
        http.Redirect(w, r, "/login", http.StatusSeeOther)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}

// Login handles user login.
func Login(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        render(w, "login.html", nil)
    case http.MethodPost:
        username := r.FormValue("username")
        password := r.FormValue("password")
        user, err := models.Authenticate(username, password)
        if err != nil {
            render(w, "login.html", map[string]interface{}{"Message": "invalid credentials", "Error": true})
            return
        }
        token, err := models.CreateSession(user.ID)
        if err != nil {
            render(w, "login.html", map[string]interface{}{"Message": "session error", "Error": true})
            return
        }
        http.SetCookie(w, &http.Cookie{Name: "session", Value: token, Path: "/", Expires: time.Now().Add(24 * time.Hour), HttpOnly: true})
        http.Redirect(w, r, "/", http.StatusSeeOther)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}

// Logout removes user session.
func Logout(w http.ResponseWriter, r *http.Request) {
    c, err := r.Cookie("session")
    if err == nil {
        models.DeleteSession(c.Value)
        c.Value = ""
        c.Path = "/"
        c.Expires = time.Unix(0, 0)
        http.SetCookie(w, c)
    }
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
