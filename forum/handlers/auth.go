package handlers

import (
	"net/http"

	"forum/models"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"Title": "Login", "Theme": h.theme(r)}
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		user, err := models.AuthenticateUser(h.DB, username, password)
		if err != nil {
			data["Error"] = "invalid credentials"
		} else {
			h.setUserCookie(w, user.ID)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}
	h.Tmpl.ExecuteTemplate(w, "login", data)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"Title": "Register", "Theme": h.theme(r)}
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirm := r.FormValue("confirm")
		if len(username) < 6 || password != confirm || len(password) < 8 {
			data["Error"] = "invalid input"
		} else {
			if err := models.CreateUser(h.DB, email, username, password); err != nil {
				data["Error"] = "user exists"
			} else {
				data["Success"] = "registered, please login"
			}
		}
	}
	h.Tmpl.ExecuteTemplate(w, "register", data)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearUserCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) ToggleTheme(w http.ResponseWriter, r *http.Request) {
	t := "dark"
	if h.theme(r) == "dark" {
		t = "light"
	}
	http.SetCookie(w, &http.Cookie{Name: "theme", Value: t, Path: "/", MaxAge: 86400 * 30})
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}
