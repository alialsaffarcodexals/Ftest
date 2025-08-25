package models

// Like or dislike actions.

// LikePost registers a like or dislike for a post. value should be 1 or -1.
func LikePost(userID, postID, value int) error {
    _, err := DB.Exec(`INSERT INTO likes(user_id, post_id, value) VALUES(?,?,?)
        ON CONFLICT(user_id, post_id, comment_id) DO UPDATE SET value=excluded.value`, userID, postID, value)
    return err
}

// LikeComment registers a like or dislike for a comment.
func LikeComment(userID, commentID, value int) error {
    _, err := DB.Exec(`INSERT INTO likes(user_id, comment_id, value) VALUES(?,?,?)
        ON CONFLICT(user_id, post_id, comment_id) DO UPDATE SET value=excluded.value`, userID, commentID, value)
    return err
}

// GetPostLikes returns like and dislike counts for a post.
func GetPostLikes(postID int) (int, int) {
    var likes, dislikes int
    DB.QueryRow(`SELECT COALESCE(SUM(CASE WHEN value=1 THEN 1 END),0) as likes,
        COALESCE(SUM(CASE WHEN value=-1 THEN 1 END),0) as dislikes FROM likes WHERE post_id=?`, postID).Scan(&likes, &dislikes)
    return likes, dislikes
}

// GetCommentLikes returns like and dislike counts for a comment.
func GetCommentLikes(commentID int) (int, int) {
    var likes, dislikes int
    DB.QueryRow(`SELECT COALESCE(SUM(CASE WHEN value=1 THEN 1 END),0) as likes,
        COALESCE(SUM(CASE WHEN value=-1 THEN 1 END),0) as dislikes FROM likes WHERE comment_id=?`, commentID).Scan(&likes, &dislikes)
    return likes, dislikes
}
