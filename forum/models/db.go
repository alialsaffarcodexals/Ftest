package models

import (
    "database/sql"
    "io/ioutil"

    _ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB opens the SQLite database and executes the schema.
func InitDB(path string) error {
    var err error
    DB, err = sql.Open("sqlite3", path)
    if err != nil {
        return err
    }
    schema, err := ioutil.ReadFile("database/schema.sql")
    if err != nil {
        return err
    }
    if _, err = DB.Exec(string(schema)); err != nil {
        return err
    }
    return nil
}

// Close closes the database connection.
func Close() {
    if DB != nil {
        DB.Close()
    }
}
