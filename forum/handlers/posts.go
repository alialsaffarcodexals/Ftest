package handlers

import (
    "net/http"

    "github.com/gorilla/mux"

    "forum/models"
)

// Index shows list of posts.
func Index(w http.ResponseWriter, r *http.Request) {
    posts, _ := models.GetPosts()
    data := map[string]interface{}{"Posts": posts, "User": currentUser(r)}
    render(w, "index.html", data)
}

// NewPost handles post creation.
func NewPost(w http.ResponseWriter, r *http.Request) {
    user := currentUser(r)
    if user == nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
    switch r.Method {
    case http.MethodGet:
        cats, _ := models.GetCategories()
        render(w, "create_post.html", map[string]interface{}{"Categories": cats, "User": user})
    case http.MethodPost:
        title := r.FormValue("title")
        content := r.FormValue("content")
        catID := parseID(r.FormValue("category"))
        models.CreatePost(user.ID, catID, title, content)
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

// ViewPost shows a post with comments.
func ViewPost(w http.ResponseWriter, r *http.Request) {
    id := parseID(mux.Vars(r)["id"])
    post, err := models.GetPost(id)
    if err != nil {
        http.NotFound(w, r)
        return
    }
    comments, _ := models.GetComments(id)
    data := map[string]interface{}{"Post": post, "Comments": comments, "User": currentUser(r), "Likes": post.Likes, "Dislikes": post.Dislikes}
    render(w, "post.html", data)
}

// AddComment adds a comment to a post.
func AddComment(w http.ResponseWriter, r *http.Request) {
    user := currentUser(r)
    if user == nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
    id := parseID(mux.Vars(r)["id"])
    if r.Method == http.MethodPost {
        content := r.FormValue("content")
        models.CreateComment(id, user.ID, content)
    }
    http.Redirect(w, r, "/post/"+mux.Vars(r)["id"], http.StatusSeeOther)
}

// LikePostHandler handles likes for posts.
func LikePostHandler(w http.ResponseWriter, r *http.Request) {
    user := currentUser(r)
    if user == nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
    id := parseID(mux.Vars(r)["id"])
    val := parseID(r.FormValue("value"))
    models.LikePost(user.ID, id, val)
    http.Redirect(w, r, "/post/"+mux.Vars(r)["id"], http.StatusSeeOther)
}

// LikeCommentHandler handles likes for comments.
func LikeCommentHandler(w http.ResponseWriter, r *http.Request) {
    user := currentUser(r)
    if user == nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
    id := parseID(mux.Vars(r)["id"])
    val := parseID(r.FormValue("value"))
    models.LikeComment(user.ID, id, val)
    http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}
