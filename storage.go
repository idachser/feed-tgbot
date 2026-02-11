package main

import (
	"database/sql"
	"fmt"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) AddFeed(userID int64, feedURL string) error {
	query := `INSERT OR IGNORE INTO subscriptions (user_id, feed_url) VALUES (?, ?)`

	_, err := s.db.Exec(query, userID, feedURL)
	if err != nil {
		return fmt.Errorf("failed to add feed: %w", err)
	}

	return nil
}

func (s *Storage) RemoveFeed(userID int64, feedURL string) (bool, error) {
	query := `DELETE FROM subscriptions WHERE user_id = ? AND feed_url = ?`

	result, err := s.db.Exec(query, userID, feedURL)
	if err != nil {
		return false, fmt.Errorf("failed to remove feed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

func (s *Storage) GetFeeds(userID int64) ([]string, error) {
	query := `SELECT feed_url FROM subscriptions WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feeds: %w", err)
	}
	defer rows.Close()

	var feeds []string
	for rows.Next() {
		var feedURL string
		if err := rows.Scan(&feedURL); err != nil {
			return nil, fmt.Errorf("failed to scan feed: %w", err)
		}
		feeds = append(feeds, feedURL)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating feeds: %w", err)
	}

	return feeds, nil
}

func (s *Storage) GetAllUsers() ([]int64, error) {
	query := `SELECT DISTINCT user_id FROM subscriptions`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, userID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (s *Storage) SetLastSent(userID int64, feedURL string, itemLink string) error {
	query := `
	INSERT INTO last_sent (user_id, feed_url, item_link) 
	VALUES (?, ?, ?)
	ON CONFLICT(user_id, feed_url) 
	DO UPDATE SET item_link = ?, sent_at = CURRENT_TIMESTAMP
	`

	_, err := s.db.Exec(query, userID, feedURL, itemLink, itemLink)
	if err != nil {
		return fmt.Errorf("failed to set last sent: %w", err)
	}

	return nil
}

func (s *Storage) GetLastSent(userID int64, feedURL string) (string, error) {
	query := `SELECT item_link FROM last_sent WHERE user_id = ? AND feed_url = ?`

	var itemLink string
	err := s.db.QueryRow(query, userID, feedURL).Scan(&itemLink)

	if err == sql.ErrNoRows {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("failed to get last sent: %w", err)
	}

	return itemLink, nil
}
