
package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"forum/models"
)

func Home(w http.ResponseWriter, r *http.Request) {
	log.Println("Executing Home handler")
	uid := currentUserID(r)
	// selected categories (multi)
	selectedCats := r.URL.Query()["category"]
	selectedSet := map[string]bool{}
	for _, c := range selectedCats { if strings.TrimSpace(c) != "" { selectedSet[strings.TrimSpace(c)] = true } }

	f := models.PostFilter{Categories: selectedCats}
	if uid > 0 { // only allow mine/liked for logged-in users
		if r.URL.Query().Get("mine") == "1" { f.MineUserID = uid }
		if r.URL.Query().Get("liked") == "1" { f.LikedByUserID = uid }
	}

	posts, err := models.ListPosts(f)
	if err != nil {
		log.Printf("Error listing posts: %v", err)
		RenderHTTPError(w, http.StatusInternalServerError, "Failed to load posts")
		return
	}
	allCats, _ := models.ListAllCategories()

	render(w, r, "home.html", map[string]interface{}{
		"Title":           "Home",
		"Posts":           posts,
		"AllCategories":   allCats,
		"SelectedSet":     selectedSet,
		"CanUseOwnerFilt": uid > 0,
		"Mine":            func() int { if f.MineUserID>0 {return 1}; return 0 }(),
		"Liked":           func() int { if f.LikedByUserID>0 {return 1}; return 0 }(),
	})
}

func ViewPost(w http.ResponseWriter, r *http.Request) {
	// path: /post/{id}
	rest := strings.TrimPrefix(r.URL.Path, "/post/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		RenderHTTPError(w, http.StatusNotFound, "Post not found")
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		RenderHTTPError(w, http.StatusBadRequest, "Invalid post id")
		return
	}

	p, err := models.GetPost(id)
	if err != nil { RenderHTTPError(w, http.StatusNotFound, "Post not found"); return }
	comments, _ := models.ListComments(id)

	render(w, r, "post.html", map[string]interface{}{
		"Title":    p.Title,
		"Post":     p,
		"Comments": comments,
	})
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	uid, ok := requireAuth(w, r); if !ok { return }
	if r.Method == http.MethodGet {
		allCats, _ := models.ListAllCategories()
		render(w, r, "post.html", map[string]interface{}{
			"Title":         "Create post",
			"AllCategories": allCats,
		})
		return
	}
	if r.Method != http.MethodPost {
		RenderHTTPError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	cats := r.Form["categories"] // from checkboxes
	newCat := strings.TrimSpace(r.FormValue("new_category"))
	if newCat != "" {
		cats = append(cats, newCat)
	}
	// normalize
	out := make([]string, 0, len(cats))
	seen := map[string]bool{}
	for _, c := range cats {
		c = strings.TrimSpace(c)
		if c == "" || seen[c] { continue }
		seen[c] = true
		out = append(out, c)
	}

	pid, err := models.CreatePost(uid, title, content, out)
	if err != nil {
		SetFlash(w, "error", err.Error())
		http.Redirect(w, r, "/posts/create", http.StatusSeeOther)
		return
	}
	SetFlash(w, "success", "Post published!")
	http.Redirect(w, r, "/post/"+strconv.FormatInt(pid, 10), http.StatusSeeOther)
}

func CreateComment(w http.ResponseWriter, r *http.Request) {
	uid, ok := requireAuth(w, r); if !ok { return }
	if r.Method != http.MethodPost {
		RenderHTTPError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	pid, err := strconv.ParseInt(r.FormValue("post_id"), 10, 64)
	if err != nil {
		RenderHTTPError(w, http.StatusBadRequest, "Invalid post id for comment")
		return
	}
	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		SetFlash(w, "error", "Empty comment")
		http.Redirect(w, r, "/post/"+strconv.FormatInt(pid,10), http.StatusSeeOther)
		return
	}
	if _, err := models.CreateComment(pid, uid, content); err != nil {
		SetFlash(w, "error", "Failed to add comment")
		http.Redirect(w, r, "/post/"+strconv.FormatInt(pid,10), http.StatusSeeOther)
		return
	}
	SetFlash(w, "success", "Comment added")
	http.Redirect(w, r, "/post/"+strconv.FormatInt(pid,10), http.StatusSeeOther)
}
