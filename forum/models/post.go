package models

import "database/sql"

type Post struct {
	ID      int
	Title   string
	Content string
	Author  string
	Created string
}

func CreatePost(db *sql.DB, title, content, author string) error {
	_, err := db.Exec("INSERT INTO posts (title, content, author) VALUES (?, ?, ?)", title, content, author)
	return err
}

func GetRecentPosts(db *sql.DB) ([]Post, error) {
	rows, err := db.Query("SELECT id, title, content, author, created FROM posts ORDER BY created DESC LIMIT 10")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.Created)
		posts = append(posts, p)
	}
	return posts, nil
}

func GetPost(db *sql.DB, id int) (Post, error) {
	var p Post
	err := db.QueryRow("SELECT id, title, content, author, created FROM posts WHERE id = ?", id).Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.Created)
	return p, err
}
