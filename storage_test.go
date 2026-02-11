package main

import (
	"database/sql"
	"testing"
)

func setupTestDB(t *testing.T) *sql.DB {
	tmpFile := t.TempDir() + "/test.db"

	db, err := sql.Open("sqlite", tmpFile)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
	CREATE TABLE subscriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		feed_url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, feed_url)
	);

	CREATE TABLE last_sent (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		feed_url TEXT NOT NULL,
		item_link TEXT NOT NULL,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, feed_url)
	);

	CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
	CREATE INDEX idx_last_sent_user_id ON last_sent(user_id);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestStorage_AddFeed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)

	userID := int64(123)
	feedURL := "https://example.com/feed.rss"

	err := storage.AddFeed(userID, feedURL)
	if err != nil {
		t.Errorf("AddFeed failed: %v", err)
	}

	err = storage.AddFeed(userID, feedURL)
	if err != nil {
		t.Errorf("AddFeed duplicate failed: %v", err)
	}

	feeds, err := storage.GetFeeds(userID)
	if err != nil {
		t.Errorf("GetFeeds failed: %v", err)
	}

	if len(feeds) != 1 {
		t.Errorf("expected 1 feed, got %d", len(feeds))
	}

	if feeds[0] != feedURL {
		t.Errorf("expected feed URL %q, got %q", feedURL, feeds[0])
	}
}

func TestStorage_RemoveFeed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)

	userID := int64(123)
	feedURL := "https://example.com/feed.rss"

	storage.AddFeed(userID, feedURL)

	removed, err := storage.RemoveFeed(userID, feedURL)
	if err != nil {
		t.Errorf("RemoveFeed failed: %v", err)
	}

	if !removed {
		t.Error("expected removed=true, got false")
	}

	removed, err = storage.RemoveFeed(userID, feedURL)
	if err != nil {
		t.Errorf("RemoveFeed failed: %v", err)
	}

	if removed {
		t.Error("expected removed=false, got true")
	}
}

func TestStorage_GetFeeds(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)

	userID := int64(123)
	feed1 := "https://example.com/feed1.rss"
	feed2 := "https://example.com/feed2.rss"

	storage.AddFeed(userID, feed1)
	storage.AddFeed(userID, feed2)

	feeds, err := storage.GetFeeds(userID)
	if err != nil {
		t.Errorf("GetFeeds failed: %v", err)
	}

	if len(feeds) != 2 {
		t.Errorf("expected 2 feeds, got %d", len(feeds))
	}

	found1, found2 := false, false
	for _, feed := range feeds {
		if feed == feed1 {
			found1 = true
		}
		if feed == feed2 {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("not all feeds found")
	}
}

func TestStorage_GetAllUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)

	user1 := int64(123)
	user2 := int64(456)
	feedURL := "https://example.com/feed.rss"

	storage.AddFeed(user1, feedURL)
	storage.AddFeed(user2, feedURL)

	users, err := storage.GetAllUsers()
	if err != nil {
		t.Errorf("GetAllUsers failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}

	foundUser1, foundUser2 := false, false
	for _, user := range users {
		if user == user1 {
			foundUser1 = true
		}
		if user == user2 {
			foundUser2 = true
		}
	}

	if !foundUser1 || !foundUser2 {
		t.Error("not all users found")
	}
}

func TestStorage_LastSent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)

	userID := int64(123)
	feedURL := "https://example.com/feed.rss"
	itemLink1 := "https://example.com/article1"
	itemLink2 := "https://example.com/article2"

	lastSent, err := storage.GetLastSent(userID, feedURL)
	if err != nil {
		t.Errorf("GetLastSent failed: %v", err)
	}

	if lastSent != "" {
		t.Errorf("expected empty string, got %q", lastSent)
	}

	err = storage.SetLastSent(userID, feedURL, itemLink1)
	if err != nil {
		t.Errorf("SetLastSent failed: %v", err)
	}

	lastSent, err = storage.GetLastSent(userID, feedURL)
	if err != nil {
		t.Errorf("GetLastSent failed: %v", err)
	}

	if lastSent != itemLink1 {
		t.Errorf("expected %q, got %q", itemLink1, lastSent)
	}

	err = storage.SetLastSent(userID, feedURL, itemLink2)
	if err != nil {
		t.Errorf("SetLastSent update failed: %v", err)
	}

	lastSent, err = storage.GetLastSent(userID, feedURL)
	if err != nil {
		t.Errorf("GetLastSent failed: %v", err)
	}

	if lastSent != itemLink2 {
		t.Errorf("expected %q, got %q", itemLink2, lastSent)
	}
}
