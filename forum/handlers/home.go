package handlers

import (
	"net/http"

	"forum/models"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	posts, _ := models.GetRecentPosts(h.DB)
	data := map[string]interface{}{
		"Title": "Home",
		"Theme": h.theme(r),
		"Posts": posts,
	}
	if u, err := h.currentUser(r); err == nil {
		data["User"] = u
	}
	h.Tmpl.ExecuteTemplate(w, "home", data)
}
