package models

import "time"

// Post represents a forum post.
type Post struct {
    ID         int
    UserID     int
    CategoryID int
    Username   string
    Category   string
    Title      string
    Content    string
    CreatedAt  time.Time
    Likes      int
    Dislikes   int
}

// CreatePost inserts a new post.
func CreatePost(userID, categoryID int, title, content string) error {
    _, err := DB.Exec(`INSERT INTO posts(user_id, category_id, title, content) VALUES(?,?,?,?)`, userID, categoryID, title, content)
    return err
}

// GetPosts returns all posts with like counters.
func GetPosts() ([]Post, error) {
    rows, err := DB.Query(`SELECT p.id, p.user_id, p.category_id, u.username, c.name, p.title, p.content, p.created_at,
        COALESCE(SUM(CASE WHEN l.value=1 THEN 1 END),0) as likes,
        COALESCE(SUM(CASE WHEN l.value=-1 THEN 1 END),0) as dislikes
        FROM posts p
        JOIN users u ON p.user_id=u.id
        LEFT JOIN categories c ON p.category_id=c.id
        LEFT JOIN likes l ON l.post_id=p.id
        GROUP BY p.id
        ORDER BY p.created_at DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var posts []Post
    for rows.Next() {
        var p Post
        if err := rows.Scan(&p.ID, &p.UserID, &p.CategoryID, &p.Username, &p.Category, &p.Title, &p.Content, &p.CreatedAt, &p.Likes, &p.Dislikes); err != nil {
            return nil, err
        }
        posts = append(posts, p)
    }
    return posts, nil
}

// GetPost returns single post and its data.
func GetPost(id int) (*Post, error) {
    p := &Post{}
    err := DB.QueryRow(`SELECT p.id, p.user_id, p.category_id, u.username, c.name, p.title, p.content, p.created_at
        FROM posts p
        JOIN users u ON p.user_id=u.id
        LEFT JOIN categories c ON p.category_id=c.id
        WHERE p.id=?`, id).Scan(&p.ID, &p.UserID, &p.CategoryID, &p.Username, &p.Category, &p.Title, &p.Content, &p.CreatedAt)
    if err != nil {
        return nil, err
    }
    likes, dislikes := GetPostLikes(id)
    p.Likes = likes
    p.Dislikes = dislikes
    return p, nil
}
