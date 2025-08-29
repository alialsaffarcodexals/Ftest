package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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

func GuestHandler(w http.ResponseWriter, r *http.Request) {
	// ✅ If it's a GET request → create a guest session
	if r.Method == http.MethodGet {
		user, err := db.Guest()
		if err != nil {
			RenderError(w, "We couldn’t create a guest session. Please try again.", http.StatusInternalServerError)
			return
		}

		// ✅ Check if a user session already exists
		if cookie, err := r.Cookie("user"); err == nil && cookie.Value != "" {
			db.DeleteUser(cookie.Value)
		}

		// ✅ Set cookie manually
		SetUserCookie(w, user.UUID)

		// ✅ Redirect to /home
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// ❌ Method not allowed
	RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// LogoutHandler handles POST /logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	uuid, err := GetUserFromCookie(r)
	if err != nil || uuid == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Mark user as logged out
	_, err = db.Conn.Exec("UPDATE users SET loggedin = 0 WHERE uuid = ?", uuid)
	if err != nil {
		http.Error(w, "Failed to log out", http.StatusInternalServerError)
		return
	}

	// Clear cookie using helper
	ClearUserCookie(w)

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Authenticate
		user, err := db.Login(w, r, username, email, password)
		if err != nil {
			RenderError(w, "Invalid username, email, or password", http.StatusBadRequest)

			return
		}

		// Store cookie
		SetUserCookie(w, user.UUID)

		// Redirect (doesn't show POST response to user)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodGet {
		uuid, err := GetUserFromCookie(r)
		if err == nil && uuid != "" {
			// User already logged in → redirect to home
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}

		// Otherwise show login form
		InitTemplate(w, "templates/login.html", nil)
		return
	}

	RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Get UUID from cookie
	uuid, err := GetUserFromCookie(r)
	if err != nil || uuid == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check session validity
	if err := db.CheckSession(w, uuid); err != nil {
		RenderError(w, "Session expired. Please log in again.", http.StatusUnauthorized)
		return
	}

	// Refresh session
	_ = db.RefreshSession(uuid)

	// ✅ Fetch posts from DB
	rows, err := db.Conn.Query(`
        SELECT posts.id, posts.title, posts.content, users.username
        FROM posts
        JOIN users ON posts.author_uuid = users.uuid
        ORDER BY posts.id DESC
    `)
	if err != nil {
		RenderError(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []map[string]string
	for rows.Next() {
		var id int
		var title, content, author string
		if err := rows.Scan(&id, &title, &content, &author); err != nil {
			continue
		}
		// Count comments
		var commentCount int
		db.Conn.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id = ?", id).Scan(&commentCount)

		// Count likes
		var likeCount int
		db.Conn.QueryRow("SELECT COUNT(*) FROM interactions WHERE post_id = ? AND liked = 1", id).Scan(&likeCount)

		posts = append(posts, map[string]string{
			"ID":           fmt.Sprint(id),
			"Title":        title,
			"Content":      content,
			"Author":       author,
			"CommentCount": fmt.Sprint(commentCount),
			"LikeCount":    fmt.Sprint(likeCount),
		})
	}
	var notRegistered bool
	err = db.Conn.QueryRow("SELECT notregistered FROM users WHERE uuid = ?", uuid).Scan(&notRegistered)
	if err != nil {
		RenderError(w, "Failed to check user", http.StatusInternalServerError)
		return
	}

	catRows, err := db.Conn.Query("SELECT name FROM categories ORDER BY name ASC")
	if err != nil {
		RenderError(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}
	defer catRows.Close()

	var categories []string

	for catRows.Next() {
		var name string
		if err := catRows.Scan(&name); err == nil {
			categories = append(categories, name)
		}
	}

	// Render template with posts
	data := map[string]interface{}{
		"UUID":          uuid,
		"Posts":         posts,
		"NotRegistered": notRegistered,
		"Categories":    categories,
	}
	InitTemplate(w, "templates/home.html", data)
}

func (db *DataBase) Guest() (*User, error) {
	uuid, err := GenerateUserID()
	if err != nil {
		return nil, err
	}

	user := User{
		UUID:          uuid,
		NotRegistered: true,
		Username:      "guest_" + uuid[:8],
		Email:         "",
		Password:      "",
		Lastseen:      time.Now(),
	}

	if err := db.SafeWriter("users", user); err != nil {
		return nil, err
	}

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

			RenderError(w, "All fields are required", http.StatusBadRequest)
			return
		}

		if password != confirmPassword {

			RenderError(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		// Register user
		user, err := db.Register(w, username, email, password)
		if err != nil {

			RenderError(w, "Username or email already exists", http.StatusBadRequest)
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
		Lastseen:      time.Now(),
	}

	// Insert safely using SafeWriter
	if err := db.SafeWriter("users", user); err != nil {
		log.Println("Failed to insert user:", err)
		RenderError(w, "Internal server error", http.StatusInternalServerError)
		return nil, err
	}

	return &user, nil
}

// CreatePostHandler handles creating new posts
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	uuid, err := GetUserFromCookie(r)
	if err != nil || uuid == "" {
		// Not logged in → redirect to login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Ensure only registered users can post
	var notRegistered bool
	err = db.Conn.QueryRow("SELECT notregistered FROM users WHERE uuid = ?", uuid).Scan(&notRegistered)
	if err != nil {
		RenderError(w, "Failed to check user type", http.StatusInternalServerError)
		return
	}
	if notRegistered {
		RenderError(w, "Guests cannot create posts", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet {
		// Show form template
		InitTemplate(w, "templates/create_post.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		rawCats := r.FormValue("categories") // multiple values
		if title == "" || content == "" {
			RenderError(w, "Title and content cannot be empty", http.StatusBadRequest)
			return
		}

		res, err := db.Conn.Exec("INSERT INTO posts (title, content, author_uuid) VALUES (?, ?, ?)", title, content, uuid)
		if err != nil {
			RenderError(w, "Failed to create post: "+err.Error(), http.StatusInternalServerError)
			return
		}
		postID64, _ := res.LastInsertId()
		postID := int(postID64)

		categories := strings.Split(rawCats, ",")

		for i := range categories {
			categories[i] = strings.TrimSpace(categories[i])
		}
		// Insert categories
		for _, cat := range categories {
			// Ensure category exists
			var catID int
			err := db.Conn.QueryRow("SELECT id FROM categories WHERE name = ?", cat).Scan(&catID)
			if err == sql.ErrNoRows {
				res, err := db.Conn.Exec("INSERT INTO categories (name) VALUES (?)", cat)
				if err == nil {
					catID64, _ := res.LastInsertId()
					catID = int(catID64)
				}
			}
			if catID > 0 {
				db.Conn.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catID)
			}
		}

		// Redirect back to home after success
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// PostHandler handles viewing a single post and adding comments
func PostHandler(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path
	postIDStr := strings.TrimPrefix(r.URL.Path, "/post/")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		RenderError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Fetch post from DB
	var title, content, author string
	err = db.Conn.QueryRow(`
        SELECT posts.title, posts.content, users.username
        FROM posts
        JOIN users ON posts.author_uuid = users.uuid
        WHERE posts.id = ?
    `, postID).Scan(&title, &content, &author)
	if err != nil {
		RenderError(w, "Post not found", http.StatusNotFound)
		return
	}

	// If POST → add comment
	if r.Method == http.MethodPost {
		uuid, err := GetUserFromCookie(r)
		if err != nil || uuid == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var notRegistered bool
		err = db.Conn.QueryRow("SELECT notregistered FROM users WHERE uuid = ?", uuid).Scan(&notRegistered)
		if err != nil {
			RenderError(w, "Failed to check user type", http.StatusInternalServerError)
			return
		}
		if notRegistered {
			RenderError(w, "Guests cannot comment", http.StatusForbidden)
			return
		}

		content := r.FormValue("comment")
		if content == "" {
			RenderError(w, "Comment cannot be empty", http.StatusBadRequest)
			return
		}

		_, err = db.Conn.Exec("INSERT INTO comments (content, comment_author_uuid, post_id) VALUES (?, ?, ?)", content, uuid, postID)
		if err != nil {
			RenderError(w, "Failed to add comment", http.StatusInternalServerError)
			return
		}

		// Redirect to same post page
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
		return
	}

	// Fetch comments for this post
	rows, err := db.Conn.Query(`
        SELECT comments.content, users.username
        FROM comments
        JOIN users ON comments.comment_author_uuid = users.uuid
        WHERE comments.post_id = ?
        ORDER BY comments.id DESC
    `, postID)
	if err != nil {
		RenderError(w, "Failed to load comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []map[string]string
	for rows.Next() {
		var cContent, cAuthor string
		if err := rows.Scan(&cContent, &cAuthor); err == nil {
			comments = append(comments, map[string]string{
				"Author":  cAuthor,
				"Content": cContent,
			})
		}
	}
	// Count likes & dislikes
	var likeCount, dislikeCount int
	db.Conn.QueryRow("SELECT COUNT(*) FROM interactions WHERE post_id = ? AND liked = 1", postID).Scan(&likeCount)
	db.Conn.QueryRow("SELECT COUNT(*) FROM interactions WHERE post_id = ? AND disliked = 1", postID).Scan(&dislikeCount)

	// Render template
	data := map[string]interface{}{
		"Title":    title,
		"Content":  content,
		"Author":   author,
		"Comments": comments,
		"PostID":   postID,
		"Likes":    likeCount,
		"Dislikes": dislikeCount,
	}
	InitTemplate(w, "templates/post.html", data)
}

// LikeHandler handles liking a post
func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uuid, err := GetUserFromCookie(r)
	if err != nil || uuid == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postID := r.FormValue("post_id")
	if postID == "" {
		RenderError(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	// Ensure user is registered
	var notRegistered bool
	err = db.Conn.QueryRow("SELECT notregistered FROM users WHERE uuid = ?", uuid).Scan(&notRegistered)
	if err != nil || notRegistered {
		RenderError(w, "Guests cannot like posts", http.StatusForbidden)
		return
	}

	// Remove any existing interaction by this user on this post
	_, _ = db.Conn.Exec("DELETE FROM interactions WHERE user_uuid = ? AND post_id = ?", uuid, postID)

	// Insert like
	_, err = db.Conn.Exec("INSERT INTO interactions (user_uuid, post_id, liked, disliked) VALUES (?, ?, 1, 0)", uuid, postID)
	if err != nil {
		RenderError(w, "Failed to like post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
}

// DislikeHandler handles disliking a post
func DislikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uuid, err := GetUserFromCookie(r)
	if err != nil || uuid == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postID := r.FormValue("post_id")
	if postID == "" {
		RenderError(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	// Ensure user is registered
	var notRegistered bool
	err = db.Conn.QueryRow("SELECT notregistered FROM users WHERE uuid = ?", uuid).Scan(&notRegistered)
	if err != nil || notRegistered {
		RenderError(w, "Guests cannot dislike posts", http.StatusForbidden)
		return
	}

	// Remove any existing interaction by this user on this post
	_, _ = db.Conn.Exec("DELETE FROM interactions WHERE user_uuid = ? AND post_id = ?", uuid, postID)

	// Insert dislike
	_, err = db.Conn.Exec("INSERT INTO interactions (user_uuid, post_id, liked, disliked) VALUES (?, ?, 0, 1)", uuid, postID)
	if err != nil {
		RenderError(w, "Failed to dislike post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
}

func FilterHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	filterType := r.URL.Query().Get("filter")
	uuid, _ := GetUserFromCookie(r)

	var rows *sql.Rows
	var err error
	var label string

	switch {
	case category != "":
		label = "Category: " + category
		rows, err = db.Conn.Query(`
            SELECT posts.id, posts.title, posts.content, users.username
            FROM posts
            JOIN users ON posts.author_uuid = users.uuid
            JOIN post_categories ON posts.id = post_categories.post_id
            JOIN categories ON post_categories.category_id = categories.id
            WHERE categories.name = ?
            ORDER BY posts.id DESC
        `, category)

	case filterType == "myposts" && uuid != "":
		label = "My Posts"
		rows, err = db.Conn.Query(`
            SELECT posts.id, posts.title, posts.content, users.username
            FROM posts
            JOIN users ON posts.author_uuid = users.uuid
            WHERE users.uuid = ?
            ORDER BY posts.id DESC
        `, uuid)

	case filterType == "mylikes" && uuid != "":
		label = "My Liked Posts"
		rows, err = db.Conn.Query(`
            SELECT posts.id, posts.title, posts.content, users.username
            FROM posts
            JOIN users ON posts.author_uuid = users.uuid
            JOIN interactions ON posts.id = interactions.post_id
            WHERE interactions.user_uuid = ? AND interactions.liked = 1
            ORDER BY posts.id DESC
        `, uuid)

	default:
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	if err != nil {
		RenderError(w, "Failed to filter posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []map[string]string
	for rows.Next() {
		var id int
		var title, content, author string
		rows.Scan(&id, &title, &content, &author)
		posts = append(posts, map[string]string{
			"ID":      fmt.Sprint(id),
			"Title":   title,
			"Content": content,
			"Author":  author,
		})
	}

	data := map[string]interface{}{
		"FilterLabel": label,
		"Posts":       posts,
	}
	InitTemplate(w, "templates/filter.html", data)
}
