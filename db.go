package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(filepath string) error {
	var err error

	DB, err = sql.Open("sqlite", filepath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Print("database initialized succsessfuly.")
	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		feed_url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, feed_url)
	);

	CREATE TABLE IF NOT EXISTS last_sent (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		feed_url TEXT NOT NULL,
		item_link TEXT NOT NULL,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, feed_url)
	);

	CREATE TABLE IF NOT EXISTS pending_feed_selections (
		user_id INTEGER PRIMARY KEY,
		feeds_json TEXT NOT NULL,
		created_unix INTEGER NOT NULL DEFAULT (strftime('%s','now'))
	);

	CREATE TABLE IF NOT EXISTS pending_remove_selections (
		user_id INTEGER PRIMARY KEY,
		feeds_json TEXT NOT NULL,
		created_unix INTEGER NOT NULL DEFAULT (strftime('%s','now'))
	);

	CREATE TABLE IF NOT EXISTS pending_user_actions (
		user_id INTEGER PRIMARY KEY,
		action TEXT NOT NULL,
		created_unix INTEGER NOT NULL DEFAULT (strftime('%s','now'))
	);

	CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
	CREATE INDEX IF NOT EXISTS idx_last_sent_user_id ON last_sent(user_id);
	`

	_, err := DB.Exec(schema)
	return err
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}

	return nil
}
