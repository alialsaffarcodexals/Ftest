package handlers

import (
	"net/http"
	"strconv"

	"forum/models"
)

func (h *Handler) currentUser(r *http.Request) (*models.User, error) {
	c, err := r.Cookie("user")
	if err != nil {
		return nil, err
	}
	id, err := strconv.Atoi(c.Value)
	if err != nil {
		return nil, err
	}
	u, err := models.GetUserByID(h.DB, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (h *Handler) theme(r *http.Request) string {
	c, err := r.Cookie("theme")
	if err != nil {
		return "light"
	}
	return c.Value
}

func (h *Handler) setUserCookie(w http.ResponseWriter, id int) {
	http.SetCookie(w, &http.Cookie{Name: "user", Value: strconv.Itoa(id), Path: "/"})
}

func (h *Handler) clearUserCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "user", Value: "", Path: "/", MaxAge: -1})
}
