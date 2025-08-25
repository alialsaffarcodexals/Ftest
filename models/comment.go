package models

import (
	"context"
	"log"
)

type Comment struct {
	ID           int64
	PostID       int64
	AuthorID     int64
	Author       string
	Content      string
	CreatedAt    string
	LikeCount    int
	DislikeCount int
}

func CreateComment(postID, authorID int64, content string) (id int64, err error) {
	log.Printf("models.CreateComment start")
	defer func() { log.Printf("models.CreateComment end: id=%d err=%v", id, err) }()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultQueryTimeout)
	defer cancel()
	res, err := DB.ExecContext(ctx, `INSERT INTO comments(post_id, author_id, content, created_at) VALUES(?,?,?,datetime('now'))`, postID, authorID, content)
	if err != nil {
		return 0, err
	}
	id, err = res.LastInsertId()
	return id, err
}

func ListComments(postID int64) (out []Comment, err error) {
	log.Printf("models.ListComments start postID=%d", postID)
	defer func() { log.Printf("models.ListComments end: n=%d err=%v", len(out), err) }()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultQueryTimeout)
	defer cancel()
	rows, err := DB.QueryContext(ctx, `
        SELECT c.id, c.post_id, c.author_id, u.username, c.content, c.created_at,
                COALESCE(SUM(CASE WHEN l.value=1 THEN 1 ELSE 0 END),0),
                COALESCE(SUM(CASE WHEN l.value=-1 THEN 1 ELSE 0 END),0)
        FROM comments c JOIN users u ON u.id=c.author_id
        LEFT JOIN likes l ON l.target_type='comment' AND l.target_id=c.id
        WHERE c.post_id=?
        GROUP BY c.id ORDER BY c.id ASC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Author, &c.Content, &c.CreatedAt, &c.LikeCount, &c.DislikeCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}
