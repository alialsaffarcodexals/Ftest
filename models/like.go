package models

import "database/sql"

// likes table is used for posts & comments via target_type + target_id
// value = 1 (like), -1 (dislike)
func SetReaction(userID int64, targetType string, targetID int64, value int) error {
	if targetType != "post" && targetType != "comment" { return ErrNotFound }
	if value != 1 && value != -1 { return ErrNotFound }
	// If same reaction exists -> remove (toggle). If opposite exists -> update.
	var existing int
	var curVal sql.NullInt64
	row := DB.QueryRow(`SELECT value FROM likes WHERE user_id=? AND target_type=? AND target_id=?`,
		userID, targetType, targetID)
	if err := row.Scan(&curVal); err != nil && err != sql.ErrNoRows {
		return err
	}
	if curVal.Valid {
		existing = int(curVal.Int64)
	}
	if existing == value {
		_, err := DB.Exec(`DELETE FROM likes WHERE user_id=? AND target_type=? AND target_id=?`,
			userID, targetType, targetID)
		return err
	}
	if existing == -value {
		_, err := DB.Exec(`UPDATE likes SET value=? WHERE user_id=? AND target_type=? AND target_id=?`,
			value, userID, targetType, targetID)
		return err
	}
	_, err := DB.Exec(`INSERT INTO likes(user_id, target_type, target_id, value, created_at) VALUES(?,?,?,?,datetime('now'))`,
		userID, targetType, targetID, value)
	return err
}
