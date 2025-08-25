package handlers

import (
	"net/http"
	"strconv"

	"forum/models"
)

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/post/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.Error(w, r, 404)
		return
	}
	post, err := models.GetPost(h.DB, id)
	if err != nil {
		h.Error(w, r, 404)
		return
	}
	comments, _ := models.GetComments(h.DB, id)
	likes, dislikes, _ := models.CountLikes(h.DB, id)
	data := map[string]interface{}{
		"Title":    post.Title,
		"Theme":    h.theme(r),
		"Post":     post,
		"Comments": comments,
		"Likes":    likes,
		"Dislikes": dislikes,
	}
	if u, err := h.currentUser(r); err == nil {
		data["User"] = u
	}
	h.Tmpl.ExecuteTemplate(w, "post", data)
}

func (h *Handler) Comment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	u, err := h.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	postID, _ := strconv.Atoi(r.FormValue("post_id"))
	content := r.FormValue("content")
	models.CreateComment(h.DB, postID, u.Username, content)
	http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
}

func (h *Handler) Like(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	u, err := h.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	postID, _ := strconv.Atoi(r.FormValue("post_id"))
	val, _ := strconv.Atoi(r.FormValue("value"))
	models.AddLike(h.DB, postID, u.Username, val)
	http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
}
