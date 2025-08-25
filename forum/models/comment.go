package models

import "database/sql"

type Comment struct {
	ID      int
	PostID  int
	Author  string
	Content string
	Created string
}

func GetComments(db *sql.DB, postID int) ([]Comment, error) {
	rows, err := db.Query("SELECT id, post_id, author, content, created FROM comments WHERE post_id = ? ORDER BY created ASC", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []Comment
	for rows.Next() {
		var c Comment
		rows.Scan(&c.ID, &c.PostID, &c.Author, &c.Content, &c.Created)
		comments = append(comments, c)
	}
	return comments, nil
}

func CreateComment(db *sql.DB, postID int, author, content string) error {
	_, err := db.Exec("INSERT INTO comments (post_id, author, content) VALUES (?, ?, ?)", postID, author, content)
	return err
}
