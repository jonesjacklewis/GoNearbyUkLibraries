package db

import (
	"database/sql"
	"example/hello/models"
	"time"

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
		date_added DATE DEFAULT CURRENT_TIMESTAMP
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

	_, err = DB.Exec(sqlStmt)

	if err != nil {
		panic(err)
	}
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

func ValidateApiKey(apiKey string) (bool, error) {
	sqlStmt := `
	SELECT EXISTS(SELECT 1 FROM api_keys WHERE api_key = ? LIMIT 1);
	`

	var exists bool

	err := DB.QueryRow(sqlStmt, apiKey).Scan(&exists)

	return exists, err
}

func GetOldestDateAddedLibraries() time.Time {
	sqlStmt := `
    SELECT date_added FROM libraries ORDER BY date_added ASC LIMIT 1;
    `

	var dateAddedString string
	err := DB.QueryRow(sqlStmt).Scan(&dateAddedString)

	// Check if the query returned any rows
	if err == sql.ErrNoRows {
		// Return today's date minus 12 months
		return time.Now().AddDate(-1, 0, 0)
	} else if err != nil {
		// If there's a different error, panic
		panic(err)
	}

	// Corrected layout to match the ISO 8601 format
	layout := "2006-01-02T15:04:05Z"
	dateAdded, err := time.Parse(layout, dateAddedString)
	if err != nil {
		panic(err)
	}

	return dateAdded
}

func InsertLibrary(library models.Library) error {
	sqlStmt := `
	INSERT OR IGNORE INTO libraries (name, latitude, longitude) VALUES (?, ?, ?);
	`

	_, err := DB.Exec(sqlStmt, library.Name, library.Point.Latitude, library.Point.Longitude)

	return err
}

func GetLibraries() ([]models.Library, error) {
	sqlStmt := `
	SELECT name, latitude, longitude FROM libraries;
	`

	rows, err := DB.Query(sqlStmt)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var libraries []models.Library

	for rows.Next() {
		var library models.Library

		err = rows.Scan(&library.Name, &library.Point.Latitude, &library.Point.Longitude)

		if err != nil {
			return nil, err
		}

		libraries = append(libraries, library)
	}

	return libraries, nil
}
