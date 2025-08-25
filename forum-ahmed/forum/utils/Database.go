package utils

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var db *DataBase

// DBInitialize connects to SQLite
func DBInitialize(dataSourceName string) (*DataBase, error) {
	conn, err := sql.Open("sqlite3", "./"+dataSourceName+".db")
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	db = &DataBase{Conn: conn}
	// ✅ Ensure the users table exists
	if err := db.ExecuteSQLFile("sql/tables.sql"); err != nil {
		fmt.Println("Error initializing tables:", err)
	}
	return db, nil
}

// ExecuteSQLFile reads an SQL file and executes all statements in it.
func (db *DataBase) ExecuteSQLFile(filepath string) error {
	db.Write.Lock()
	defer db.Write.Unlock()

	// Read the SQL file using os.ReadFile (not ioutil)
	sqlBytes, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	// Convert to string
	sqlContent := string(sqlBytes)

	// Split statements by semicolon
	statements := strings.Split(sqlContent, ";")

	// Execute each statement
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		_, err := db.Conn.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w \nstmt: %s", err, stmt)
		}
	}

	return nil
}

// flattenStruct recursively flattens a struct into column names and values
// It skips unexported fields and ID if it's zero (to allow AUTOINCREMENT)
func flattenStruct(data interface{}) ([]string, []interface{}) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Struct {
		panic("flattenStruct expects a struct")
	}

	var columns []string
	var values []interface{}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Skip zero ID so SQLite can autoincrement
		if strings.ToLower(field.Name) == "id" && fieldValue.Int() == 0 {
			continue
		}

		// Flatten nested structs (except time.Time)
		if fieldValue.Kind() == reflect.Struct && fieldValue.Type().Name() != "Time" {
			subCols, subVals := flattenStruct(fieldValue.Interface())
			for j, c := range subCols {
				columns = append(columns, strings.ToLower(field.Name)+"_"+c)
				values = append(values, subVals[j])
			}
		} else {
			columns = append(columns, strings.ToLower(field.Name))
			values = append(values, fieldValue.Interface())
		}
	}

	return columns, values
}

// SafeWriter inserts a single struct into the table safely
func (db *DataBase) SafeWriter(table string, data interface{}) error {
	db.Write.Lock()
	defer db.Write.Unlock()

	columns, values := flattenStruct(data)

	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := db.Conn.Exec(query, values...)
	return err
}

