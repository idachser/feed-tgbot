package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Storage struct {
	db *sql.DB
}

type UserUpdateSettings struct {
	UserID          int64
	Enabled         bool
	IntervalMinutes int
	NextCheckUnix   int64
}

const defaultUserUpdateIntervalMinutes = 30

var allowedUserUpdateIntervals = map[int]struct{}{
	30: {},
	60: {},
	360: {},
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func isValidUserUpdateInterval(intervalMinutes int) bool {
	_, ok := allowedUserUpdateIntervals[intervalMinutes]
	return ok
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

func (s *Storage) SetPendingFeeds(userID int64, feeds []DiscoveredFeed) error {
	feedsJSON, err := json.Marshal(feeds)
	if err != nil {
		return fmt.Errorf("failed to marshal pending feeds: %w", err)
	}

	query := `
	INSERT INTO pending_feed_selections (user_id, feeds_json, created_unix)
	VALUES (?, ?, strftime('%s','now'))
	ON CONFLICT(user_id)
	DO UPDATE SET feeds_json = excluded.feeds_json, created_unix = strftime('%s','now')
	`

	_, err = s.db.Exec(query, userID, string(feedsJSON))
	if err != nil {
		return fmt.Errorf("failed to set pending feeds: %w", err)
	}

	return nil
}

func (s *Storage) GetPendingFeeds(userID int64, maxAge time.Duration) ([]DiscoveredFeed, bool, error) {
	query := `SELECT feeds_json, created_unix FROM pending_feed_selections WHERE user_id = ?`

	var feedsJSON string
	var createdUnix int64
	err := s.db.QueryRow(query, userID).Scan(&feedsJSON, &createdUnix)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get pending feeds: %w", err)
	}

	if maxAge > 0 && time.Now().Unix()-createdUnix > int64(maxAge.Seconds()) {
		if delErr := s.DeletePendingFeeds(userID); delErr != nil {
			return nil, false, fmt.Errorf("failed to cleanup expired pending feeds: %w", delErr)
		}
		return nil, false, nil
	}

	var feeds []DiscoveredFeed
	if err := json.Unmarshal([]byte(feedsJSON), &feeds); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal pending feeds: %w", err)
	}

	return feeds, true, nil
}

func (s *Storage) DeletePendingFeeds(userID int64) error {
	query := `DELETE FROM pending_feed_selections WHERE user_id = ?`

	if _, err := s.db.Exec(query, userID); err != nil {
		return fmt.Errorf("failed to delete pending feeds: %w", err)
	}

	return nil
}

func (s *Storage) SetPendingRemoveFeeds(userID int64, feeds []string) error {
	feedsJSON, err := json.Marshal(feeds)
	if err != nil {
		return fmt.Errorf("failed to marshal pending remove feeds: %w", err)
	}

	query := `
	INSERT INTO pending_remove_selections (user_id, feeds_json, created_unix)
	VALUES (?, ?, strftime('%s','now'))
	ON CONFLICT(user_id)
	DO UPDATE SET feeds_json = excluded.feeds_json, created_unix = strftime('%s','now')
	`

	_, err = s.db.Exec(query, userID, string(feedsJSON))
	if err != nil {
		return fmt.Errorf("failed to set pending remove feeds: %w", err)
	}

	return nil
}

func (s *Storage) GetPendingRemoveFeeds(userID int64, maxAge time.Duration) ([]string, bool, error) {
	query := `SELECT feeds_json, created_unix FROM pending_remove_selections WHERE user_id = ?`

	var feedsJSON string
	var createdUnix int64
	err := s.db.QueryRow(query, userID).Scan(&feedsJSON, &createdUnix)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get pending remove feeds: %w", err)
	}

	if maxAge > 0 && time.Now().Unix()-createdUnix > int64(maxAge.Seconds()) {
		if delErr := s.DeletePendingRemoveFeeds(userID); delErr != nil {
			return nil, false, fmt.Errorf("failed to cleanup expired pending remove feeds: %w", delErr)
		}
		return nil, false, nil
	}

	var feeds []string
	if err := json.Unmarshal([]byte(feedsJSON), &feeds); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal pending remove feeds: %w", err)
	}

	return feeds, true, nil
}

func (s *Storage) DeletePendingRemoveFeeds(userID int64) error {
	query := `DELETE FROM pending_remove_selections WHERE user_id = ?`

	if _, err := s.db.Exec(query, userID); err != nil {
		return fmt.Errorf("failed to delete pending remove feeds: %w", err)
	}

	return nil
}

func (s *Storage) SetPendingNewsFeeds(userID int64, feeds []string) error {
	feedsJSON, err := json.Marshal(feeds)
	if err != nil {
		return fmt.Errorf("failed to marshal pending news feeds: %w", err)
	}

	query := `
	INSERT INTO pending_news_selections (user_id, feeds_json, created_unix)
	VALUES (?, ?, strftime('%s','now'))
	ON CONFLICT(user_id)
	DO UPDATE SET feeds_json = excluded.feeds_json, created_unix = strftime('%s','now')
	`

	_, err = s.db.Exec(query, userID, string(feedsJSON))
	if err != nil {
		return fmt.Errorf("failed to set pending news feeds: %w", err)
	}

	return nil
}

func (s *Storage) GetPendingNewsFeeds(userID int64, maxAge time.Duration) ([]string, bool, error) {
	query := `SELECT feeds_json, created_unix FROM pending_news_selections WHERE user_id = ?`

	var feedsJSON string
	var createdUnix int64
	err := s.db.QueryRow(query, userID).Scan(&feedsJSON, &createdUnix)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get pending news feeds: %w", err)
	}

	if maxAge > 0 && time.Now().Unix()-createdUnix > int64(maxAge.Seconds()) {
		if delErr := s.DeletePendingNewsFeeds(userID); delErr != nil {
			return nil, false, fmt.Errorf("failed to cleanup expired pending news feeds: %w", delErr)
		}
		return nil, false, nil
	}

	var feeds []string
	if err := json.Unmarshal([]byte(feedsJSON), &feeds); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal pending news feeds: %w", err)
	}

	return feeds, true, nil
}

func (s *Storage) DeletePendingNewsFeeds(userID int64) error {
	query := `DELETE FROM pending_news_selections WHERE user_id = ?`

	if _, err := s.db.Exec(query, userID); err != nil {
		return fmt.Errorf("failed to delete pending news feeds: %w", err)
	}

	return nil
}

func (s *Storage) SetPendingAction(userID int64, action string) error {
	query := `
	INSERT INTO pending_user_actions (user_id, action, created_unix)
	VALUES (?, ?, strftime('%s','now'))
	ON CONFLICT(user_id)
	DO UPDATE SET action = excluded.action, created_unix = strftime('%s','now')
	`

	_, err := s.db.Exec(query, userID, action)
	if err != nil {
		return fmt.Errorf("failed to set pending action: %w", err)
	}

	return nil
}

func (s *Storage) GetPendingAction(userID int64, maxAge time.Duration) (string, bool, error) {
	query := `SELECT action, created_unix FROM pending_user_actions WHERE user_id = ?`

	var action string
	var createdUnix int64
	err := s.db.QueryRow(query, userID).Scan(&action, &createdUnix)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to get pending action: %w", err)
	}

	if maxAge > 0 && time.Now().Unix()-createdUnix > int64(maxAge.Seconds()) {
		if delErr := s.DeletePendingAction(userID); delErr != nil {
			return "", false, fmt.Errorf("failed to cleanup expired pending action: %w", delErr)
		}
		return "", false, nil
	}

	return action, true, nil
}

func (s *Storage) DeletePendingAction(userID int64) error {
	query := `DELETE FROM pending_user_actions WHERE user_id = ?`

	if _, err := s.db.Exec(query, userID); err != nil {
		return fmt.Errorf("failed to delete pending action: %w", err)
	}

	return nil
}

func (s *Storage) EnsureUserUpdateSettings(userID int64) error {
	query := `
	INSERT OR IGNORE INTO user_update_settings (user_id, enabled, interval_minutes, next_check_unix)
	VALUES (?, 1, ?, strftime('%s','now'))
	`

	if _, err := s.db.Exec(query, userID, defaultUserUpdateIntervalMinutes); err != nil {
		return fmt.Errorf("failed to ensure user update settings: %w", err)
	}

	return nil
}

func (s *Storage) GetUserUpdateSettings(userID int64) (UserUpdateSettings, error) {
	if err := s.EnsureUserUpdateSettings(userID); err != nil {
		return UserUpdateSettings{}, err
	}

	query := `
	SELECT user_id, enabled, interval_minutes, next_check_unix
	FROM user_update_settings
	WHERE user_id = ?
	`

	var settings UserUpdateSettings
	var enabledInt int
	err := s.db.QueryRow(query, userID).Scan(
		&settings.UserID,
		&enabledInt,
		&settings.IntervalMinutes,
		&settings.NextCheckUnix,
	)
	if err != nil {
		return UserUpdateSettings{}, fmt.Errorf("failed to get user update settings: %w", err)
	}

	settings.Enabled = enabledInt == 1
	return settings, nil
}

func (s *Storage) SetUserUpdatesEnabled(userID int64, enabled bool) error {
	if err := s.EnsureUserUpdateSettings(userID); err != nil {
		return err
	}

	enabledInt := 0
	query := `
	UPDATE user_update_settings
	SET enabled = ?, updated_at = CURRENT_TIMESTAMP
	WHERE user_id = ?
	`
	args := []any{enabledInt, userID}

	if enabled {
		enabledInt = 1
		query = `
		UPDATE user_update_settings
		SET enabled = ?, next_check_unix = strftime('%s','now'), updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
		`
		args = []any{enabledInt, userID}
	}

	if _, err := s.db.Exec(query, args...); err != nil {
		return fmt.Errorf("failed to set user updates enabled: %w", err)
	}

	return nil
}

func (s *Storage) SetUserUpdateInterval(userID int64, intervalMinutes int) error {
	if !isValidUserUpdateInterval(intervalMinutes) {
		return fmt.Errorf("invalid interval: %d", intervalMinutes)
	}

	if err := s.EnsureUserUpdateSettings(userID); err != nil {
		return err
	}

	query := `
	UPDATE user_update_settings
	SET interval_minutes = ?, next_check_unix = strftime('%s','now') + (? * 60), updated_at = CURRENT_TIMESTAMP
	WHERE user_id = ?
	`

	if _, err := s.db.Exec(query, intervalMinutes, intervalMinutes, userID); err != nil {
		return fmt.Errorf("failed to set user update interval: %w", err)
	}

	return nil
}

func (s *Storage) SetNextCheckUnix(userID int64, nextCheckUnix int64) error {
	if err := s.EnsureUserUpdateSettings(userID); err != nil {
		return err
	}

	query := `
	UPDATE user_update_settings
	SET next_check_unix = ?, updated_at = CURRENT_TIMESTAMP
	WHERE user_id = ?
	`

	if _, err := s.db.Exec(query, nextCheckUnix, userID); err != nil {
		return fmt.Errorf("failed to set next check unix: %w", err)
	}

	return nil
}

func (s *Storage) ListDueUsers(nowUnix int64) ([]int64, error) {
	query := `
	SELECT user_id
	FROM user_update_settings
	WHERE enabled = 1 AND next_check_unix <= ?
	ORDER BY next_check_unix ASC
	`

	rows, err := s.db.Query(query, nowUnix)
	if err != nil {
		return nil, fmt.Errorf("failed to list due users: %w", err)
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan due user: %w", err)
		}
		users = append(users, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating due users: %w", err)
	}

	return users, nil
}
