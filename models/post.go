
package models

import (
	"database/sql"
	"errors"
	"strings"
)

type Post struct {
	ID           int64
	Title        string
	Content      string
	AuthorID     int64
	Author       string
	CreatedAt    string
	LikeCount    int
	DislikeCount int
	Categories   []string
}

func CreatePost(authorID int64, title, content string, categories []string) (int64, error) {
	if title == "" || content == "" {
		return 0, errors.New("missing fields")
	}
	tx, err := DB.Begin()
	if err != nil { return 0, err }
	defer func(){ if err != nil { _ = tx.Rollback() } }()

	res, err := tx.Exec(`INSERT INTO posts(title, content, author_id, created_at) VALUES(?,?,?,datetime('now'))`,
		title, content, authorID)
	if err != nil { return 0, err }
	pid, _ := res.LastInsertId()

	for _, c := range categories {
		c = strings.TrimSpace(c)
		if c == "" { continue }
		var cid int64
		err = tx.QueryRow(`SELECT id FROM categories WHERE name=?`, c).Scan(&cid)
		if err == sql.ErrNoRows {
			res2, err2 := tx.Exec(`INSERT INTO categories(name) VALUES(?)`, c)
			if err2 != nil { return 0, err2 }
			cid, _ = res2.LastInsertId()
		} else if err != nil {
			return 0, err
		}
		if _, err = tx.Exec(`INSERT OR IGNORE INTO post_categories(post_id, category_id) VALUES(?,?)`, pid, cid); err != nil {
			return 0, err
		}
	}
	if err = tx.Commit(); err != nil { return 0, err }
	return pid, nil
}

type PostFilter struct {
	Categories     []string
	MineUserID     int64
	LikedByUserID  int64
}

// ANY-of selected categories
func ListPosts(f PostFilter) ([]Post, error) {
	q := `
SELECT p.id, p.title, p.content, p.author_id, u.username, p.created_at,
       COALESCE(SUM(CASE WHEN l.value=1 THEN 1 ELSE 0 END),0) as likes,
       COALESCE(SUM(CASE WHEN l.value=-1 THEN 1 ELSE 0 END),0) as dislikes,
       GROUP_CONCAT(c.name)
FROM posts p
JOIN users u ON u.id=p.author_id
LEFT JOIN likes l ON l.target_type='post' AND l.target_id=p.id
LEFT JOIN post_categories pc ON pc.post_id = p.id
LEFT JOIN categories c ON c.id = pc.category_id
`
	args := []interface{}{}
	if len(f.Categories) > 0 {
		q += " WHERE p.id IN (SELECT pc.post_id FROM post_categories pc JOIN categories c ON c.id=pc.category_id WHERE c.name IN ("
		for i := range f.Categories {
			if i > 0 { q += "," }
			q += "?"
			args = append(args, f.Categories[i])
		}
		q += ") GROUP BY pc.post_id) "
	} else {
		q += " WHERE 1=1 "
	}
	if f.MineUserID > 0 {
		q += " AND p.author_id=? "
		args = append(args, f.MineUserID)
	}
	if f.LikedByUserID > 0 {
		q += " AND p.id IN (SELECT target_id FROM likes WHERE target_type='post' AND user_id=? AND value=1) "
		args = append(args, f.LikedByUserID)
	}

	q += " GROUP BY p.id ORDER BY p.id DESC LIMIT 200"

	rows, err := DB.Query(q, args...)
	if err != nil { return nil, err }
	defer rows.Close()

	var out []Post
	for rows.Next() {
		var p Post
		var cats sql.NullString
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.Author, &p.CreatedAt, &p.LikeCount, &p.DislikeCount, &cats); err != nil {
			return nil, err
		}
		if cats.Valid {
			p.Categories = strings.Split(cats.String, ",")
		}
		out = append(out, p)
	}
	return out, nil
}

func GetPost(id int64) (*Post, error) {
	row := DB.QueryRow(`
SELECT p.id, p.title, p.content, p.author_id, u.username, p.created_at,
       COALESCE(SUM(CASE WHEN l.value=1 THEN 1 ELSE 0 END),0),
       COALESCE(SUM(CASE WHEN l.value=-1 THEN 1 ELSE 0 END),0)
FROM posts p JOIN users u ON u.id=p.author_id
LEFT JOIN likes l ON l.target_type='post' AND l.target_id=p.id
WHERE p.id=?`, id)
	var p Post
	if err := row.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.Author, &p.CreatedAt, &p.LikeCount, &p.DislikeCount); err != nil {
		if err == sql.ErrNoRows { return nil, ErrNotFound }
		return nil, err
	}
	crows, err := DB.Query(`SELECT c.name FROM categories c JOIN post_categories pc ON pc.category_id=c.id WHERE pc.post_id=?`, p.ID)
	if err == nil {
		for crows.Next() { var name string; _ = crows.Scan(&name); p.Categories = append(p.Categories, name) }
		crows.Close()
	}
	return &p, nil
}

// ListAllCategories returns all category names
func ListAllCategories() ([]string, error) {
	rows, err := DB.Query(`SELECT name FROM categories ORDER BY name ASC`)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}
