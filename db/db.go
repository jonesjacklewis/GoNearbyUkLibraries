package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDb(filepath string) error {
	database, err := sql.Open("sqlite3", filepath)

	if err != nil {
		panic(err)
	}

	if database == nil {
		panic("Database is nil")
	}

	DB = database

	return DB.Ping()
}

func CreateTables() {
	sqlStmt := `
	CREATE TABLE
	IF NOT EXISTS
	libraries (
		name TEXT,
		latitude REAL,
		longitude REAL,
		date_added DATE
	);
	`

	_, err := DB.Exec(sqlStmt)

	if err != nil {
		panic(err)
	}

	sqlStmt = `
	CREATE TABLE
	IF NOT EXISTS
	api_keys (
		email TEXT NOT NULL UNIQUE,
		api_key TEXT,
		datetime_added DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = DB.Exec(sqlStmt)

	if err != nil {
		panic(err)
	}
}

func TidyUp(maxDaysLibrary int, maxDaysApiKey int) {
	// delete libraries that are older than maxDaysLibrary
	sqlStmt := `
	DELETE FROM libraries WHERE date_added < date('now', '-` + string(rune(maxDaysLibrary)) + ` days');
	`

	_, err := DB.Exec(sqlStmt)

	if err != nil {
		panic(err)
	}

	// delete api keys that are older than maxDaysApiKey
	sqlStmt = `
	DELETE FROM api_keys WHERE datetime_added < datetime('now', '-` + string(rune(maxDaysApiKey)) + ` days');
	`
}

func InsertApiKey(email string, apiKey string) error {
	sqlStmt := `
	INSERT OR IGNORE INTO api_keys (email, api_key) VALUES (?, ?);
	`

	_, err := DB.Exec(sqlStmt, email, apiKey)

	return err
}

func GetApiKey(email string) (string, error) {
	sqlStmt := `
	SELECT api_key FROM api_keys WHERE email = ?;
	`

	var apiKey string

	err := DB.QueryRow(sqlStmt, email).Scan(&apiKey)

	return apiKey, err
}
