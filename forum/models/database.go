package models

import (
	"database/sql"
	"io/ioutil"
)

func InitDB(dbPath, schemaPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(string(schema)); err != nil {
		return nil, err
	}
	return db, nil
}
