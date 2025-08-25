package models

import "time"

// Comment represents a post comment.
type Comment struct {
    ID        int
    PostID    int
    UserID    int
    Username  string
    Content   string
    CreatedAt time.Time
    Likes     int
    Dislikes  int
}

// CreateComment inserts a new comment.
func CreateComment(postID, userID int, content string) error {
    _, err := DB.Exec(`INSERT INTO comments(post_id, user_id, content) VALUES(?,?,?)`, postID, userID, content)
    return err
}

// GetComments returns comments for a post.
func GetComments(postID int) ([]Comment, error) {
    rows, err := DB.Query(`SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at,
        COALESCE(SUM(CASE WHEN l.value=1 THEN 1 END),0) as likes,
        COALESCE(SUM(CASE WHEN l.value=-1 THEN 1 END),0) as dislikes
        FROM comments c
        JOIN users u ON c.user_id=u.id
        LEFT JOIN likes l ON l.comment_id=c.id
        WHERE c.post_id=?
        GROUP BY c.id
        ORDER BY c.created_at ASC`, postID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var cs []Comment
    for rows.Next() {
        var c Comment
        if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &c.CreatedAt, &c.Likes, &c.Dislikes); err != nil {
            return nil, err
        }
        cs = append(cs, c)
    }
    return cs, nil
}
