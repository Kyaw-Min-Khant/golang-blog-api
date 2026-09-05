package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := migrateDb(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrateDb(db *sql.DB) error {
	schema := `CREATE TABLE IF NOT EXISTS users (
	 id   INTEGER PRIMARY KEY AUTOINCREMENT,
	 username TEXT NOT NULL UNIQUE,
	 password TEXT NOT NULL,
	 created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS blogs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL REFERENCES users(id),
	title   TEXT NOT NULL,
	content TEXT NOT NULL,
	created_at DATETIME  DEFAULT CURRENT_TIMESTAMP,
	updated_at  DATETIME  DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(schema)
	return err
}
