
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"forum/models"
)

const sessionCookie = "sid"

func currentUserID(r *http.Request) int64 {
	if v := r.Header.Get("X-User-ID"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil { return id }
	}
	return 0
}

func requireAuth(w http.ResponseWriter, r *http.Request) (int64, bool) {
	uid := currentUserID(r)
	if uid <= 0 {
		RenderHTTPError(w, http.StatusUnauthorized, "Please login to continue.")
		return 0, false
	}
	return uid, true
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		render(w, r, "login.html", map[string]interface{}{"Title": "Login"})
		return
	}
	if r.Method != http.MethodPost {
		RenderHTTPError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	login := r.FormValue("login")
	password := r.FormValue("password")
	u, err := models.Authenticate(login, password)
	if err != nil {
		SetFlash(w, "error", "Invalid credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	s, err := models.CreateSingleSession(u.ID)
	if err != nil {
		RenderHTTPError(w, http.StatusInternalServerError, "Failed to start session")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    s.ID,
		Path:     "/",
		Expires:  time.Now().Add(models.SessionTTL),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	SetFlash(w, "success", "Welcome back, "+u.Username+"!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		render(w, r, "register.html", map[string]interface{}{"Title": "Register"})
		return
	}
	if r.Method != http.MethodPost {
		RenderHTTPError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")
	confirm := r.FormValue("confirm_password")
	if password != confirm {
		SetFlash(w, "error", "Passwords do not match")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	u, err := models.CreateUser(email, username, password)
	if err != nil {
		SetFlash(w, "error", err.Error())
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	s, err := models.CreateSingleSession(u.ID)
	if err != nil {
		RenderHTTPError(w, http.StatusInternalServerError, "Failed to start session")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    s.ID,
		Path:     "/",
		Expires:  time.Now().Add(models.SessionTTL),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	SetFlash(w, "success", "Account created! You are now logged in.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		_ = models.DeleteSession(c.Value)
		http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Expires: time.Unix(0,0), Path: "/", HttpOnly: true})
	}
	SetFlash(w, "success", "You have been logged out.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Guest mode: clear any session and continue as guest
func Guest(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		_ = models.DeleteSession(c.Value)
		http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Expires: time.Unix(0,0), Path: "/", HttpOnly: true})
	}
	SetFlash(w, "success", "Continuing as Guest. Sign in to post or react.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
