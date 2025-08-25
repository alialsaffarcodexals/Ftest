package models

import "database/sql"

func AddLike(db *sql.DB, postID int, username string, value int) error {
	_, err := db.Exec(`INSERT INTO likes (post_id, username, value) VALUES (?, ?, ?)
        ON CONFLICT(post_id, username) DO UPDATE SET value=excluded.value`, postID, username, value)
	return err
}

func CountLikes(db *sql.DB, postID int) (int, int, error) {
	rows, err := db.Query("SELECT value, COUNT(*) FROM likes WHERE post_id = ? GROUP BY value", postID)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	likes, dislikes := 0, 0
	for rows.Next() {
		var val, count int
		rows.Scan(&val, &count)
		if val > 0 {
			likes = count
		} else {
			dislikes = count
		}
	}
	return likes, dislikes, nil
}
