package main

import (
	"database/sql"
	"testing"
	"time"
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

	CREATE TABLE pending_feed_selections (
		user_id INTEGER PRIMARY KEY,
		feeds_json TEXT NOT NULL,
		created_unix INTEGER NOT NULL DEFAULT (strftime('%s','now'))
	);

	CREATE TABLE pending_remove_selections (
		user_id INTEGER PRIMARY KEY,
		feeds_json TEXT NOT NULL,
		created_unix INTEGER NOT NULL DEFAULT (strftime('%s','now'))
	);

	CREATE TABLE pending_news_selections (
		user_id INTEGER PRIMARY KEY,
		feeds_json TEXT NOT NULL,
		created_unix INTEGER NOT NULL DEFAULT (strftime('%s','now'))
	);

	CREATE TABLE pending_user_actions (
		user_id INTEGER PRIMARY KEY,
		action TEXT NOT NULL,
		created_unix INTEGER NOT NULL DEFAULT (strftime('%s','now'))
	);

	CREATE TABLE user_update_settings (
		user_id INTEGER PRIMARY KEY,
		enabled INTEGER NOT NULL DEFAULT 1,
		interval_minutes INTEGER NOT NULL DEFAULT 30,
		next_check_unix INTEGER NOT NULL DEFAULT (strftime('%s','now')),
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
	CREATE INDEX idx_last_sent_user_id ON last_sent(user_id);
	CREATE INDEX idx_user_update_settings_due ON user_update_settings(enabled, next_check_unix);
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

func TestStorage_PendingFeeds(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(123)

	feeds := []DiscoveredFeed{
		{Title: "RSS Feed", URL: "https://example.com/feed.rss", Type: "RSS"},
		{Title: "Atom Feed", URL: "https://example.com/atom.xml", Type: "Atom"},
	}

	if err := storage.SetPendingFeeds(userID, feeds); err != nil {
		t.Fatalf("SetPendingFeeds failed: %v", err)
	}

	gotFeeds, ok, err := storage.GetPendingFeeds(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPendingFeeds failed: %v", err)
	}
	if !ok {
		t.Fatal("expected pending feeds to exist")
	}
	if len(gotFeeds) != len(feeds) {
		t.Fatalf("expected %d feeds, got %d", len(feeds), len(gotFeeds))
	}

	if err := storage.DeletePendingFeeds(userID); err != nil {
		t.Fatalf("DeletePendingFeeds failed: %v", err)
	}

	_, ok, err = storage.GetPendingFeeds(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPendingFeeds after delete failed: %v", err)
	}
	if ok {
		t.Fatal("expected no pending feeds after delete")
	}
}

func TestStorage_PendingFeeds_Expires(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(123)

	feeds := []DiscoveredFeed{
		{Title: "RSS Feed", URL: "https://example.com/feed.rss", Type: "RSS"},
	}

	if err := storage.SetPendingFeeds(userID, feeds); err != nil {
		t.Fatalf("SetPendingFeeds failed: %v", err)
	}

	_, err := db.Exec(`UPDATE pending_feed_selections SET created_unix = strftime('%s','now') - 3600 WHERE user_id = ?`, userID)
	if err != nil {
		t.Fatalf("failed to age pending selection: %v", err)
	}

	_, ok, err := storage.GetPendingFeeds(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPendingFeeds failed: %v", err)
	}
	if ok {
		t.Fatal("expected pending feeds to be expired")
	}
}

func TestStorage_PendingRemoveFeeds(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(123)

	feeds := []string{
		"https://example.com/feed1.rss",
		"https://example.com/feed2.rss",
	}

	if err := storage.SetPendingRemoveFeeds(userID, feeds); err != nil {
		t.Fatalf("SetPendingRemoveFeeds failed: %v", err)
	}

	gotFeeds, ok, err := storage.GetPendingRemoveFeeds(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPendingRemoveFeeds failed: %v", err)
	}
	if !ok {
		t.Fatal("expected pending remove feeds to exist")
	}
	if len(gotFeeds) != len(feeds) {
		t.Fatalf("expected %d feeds, got %d", len(feeds), len(gotFeeds))
	}

	if err := storage.DeletePendingRemoveFeeds(userID); err != nil {
		t.Fatalf("DeletePendingRemoveFeeds failed: %v", err)
	}

	_, ok, err = storage.GetPendingRemoveFeeds(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPendingRemoveFeeds after delete failed: %v", err)
	}
	if ok {
		t.Fatal("expected no pending remove feeds after delete")
	}
}

func TestStorage_PendingRemoveFeeds_Expires(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(123)

	feeds := []string{
		"https://example.com/feed.rss",
	}

	if err := storage.SetPendingRemoveFeeds(userID, feeds); err != nil {
		t.Fatalf("SetPendingRemoveFeeds failed: %v", err)
	}

	_, err := db.Exec(`UPDATE pending_remove_selections SET created_unix = strftime('%s','now') - 3600 WHERE user_id = ?`, userID)
	if err != nil {
		t.Fatalf("failed to age pending remove selection: %v", err)
	}

	_, ok, err := storage.GetPendingRemoveFeeds(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPendingRemoveFeeds failed: %v", err)
	}
	if ok {
		t.Fatal("expected pending remove feeds to be expired")
	}
}

func TestStorage_PendingNewsFeeds(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(123)

	feeds := []string{
		"https://example.com/feed1.rss",
		"https://example.com/feed2.rss",
	}

	if err := storage.SetPendingNewsFeeds(userID, feeds); err != nil {
		t.Fatalf("SetPendingNewsFeeds failed: %v", err)
	}

	gotFeeds, ok, err := storage.GetPendingNewsFeeds(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPendingNewsFeeds failed: %v", err)
	}
	if !ok {
		t.Fatal("expected pending news feeds to exist")
	}
	if len(gotFeeds) != len(feeds) {
		t.Fatalf("expected %d feeds, got %d", len(feeds), len(gotFeeds))
	}

	if err := storage.DeletePendingNewsFeeds(userID); err != nil {
		t.Fatalf("DeletePendingNewsFeeds failed: %v", err)
	}

	_, ok, err = storage.GetPendingNewsFeeds(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPendingNewsFeeds after delete failed: %v", err)
	}
	if ok {
		t.Fatal("expected no pending news feeds after delete")
	}
}

func TestStorage_PendingNewsFeeds_Expires(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(123)

	feeds := []string{
		"https://example.com/feed.rss",
	}

	if err := storage.SetPendingNewsFeeds(userID, feeds); err != nil {
		t.Fatalf("SetPendingNewsFeeds failed: %v", err)
	}

	_, err := db.Exec(`UPDATE pending_news_selections SET created_unix = strftime('%s','now') - 3600 WHERE user_id = ?`, userID)
	if err != nil {
		t.Fatalf("failed to age pending news selection: %v", err)
	}

	_, ok, err := storage.GetPendingNewsFeeds(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPendingNewsFeeds failed: %v", err)
	}
	if ok {
		t.Fatal("expected pending news feeds to be expired")
	}
}

func TestStorage_PendingAction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(123)

	if err := storage.SetPendingAction(userID, "add_waiting_url"); err != nil {
		t.Fatalf("SetPendingAction failed: %v", err)
	}

	action, ok, err := storage.GetPendingAction(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPendingAction failed: %v", err)
	}
	if !ok {
		t.Fatal("expected pending action to exist")
	}
	if action != "add_waiting_url" {
		t.Fatalf("expected action %q, got %q", "add_waiting_url", action)
	}

	if err := storage.DeletePendingAction(userID); err != nil {
		t.Fatalf("DeletePendingAction failed: %v", err)
	}

	_, ok, err = storage.GetPendingAction(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPendingAction after delete failed: %v", err)
	}
	if ok {
		t.Fatal("expected no pending action after delete")
	}
}

func TestStorage_PendingAction_Expires(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(123)

	if err := storage.SetPendingAction(userID, "add_waiting_url"); err != nil {
		t.Fatalf("SetPendingAction failed: %v", err)
	}

	_, err := db.Exec(`UPDATE pending_user_actions SET created_unix = strftime('%s','now') - 3600 WHERE user_id = ?`, userID)
	if err != nil {
		t.Fatalf("failed to age pending action: %v", err)
	}

	_, ok, err := storage.GetPendingAction(userID, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPendingAction failed: %v", err)
	}
	if ok {
		t.Fatal("expected pending action to be expired")
	}
}

func TestStorage_UserUpdateSettings_Defaults(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(777)

	if err := storage.EnsureUserUpdateSettings(userID); err != nil {
		t.Fatalf("EnsureUserUpdateSettings failed: %v", err)
	}

	settings, err := storage.GetUserUpdateSettings(userID)
	if err != nil {
		t.Fatalf("GetUserUpdateSettings failed: %v", err)
	}

	if settings.UserID != userID {
		t.Fatalf("expected userID %d, got %d", userID, settings.UserID)
	}
	if !settings.Enabled {
		t.Fatal("expected enabled=true by default")
	}
	if settings.IntervalMinutes != defaultUserUpdateIntervalMinutes {
		t.Fatalf("expected default interval %d, got %d", defaultUserUpdateIntervalMinutes, settings.IntervalMinutes)
	}
}

func TestStorage_SetUserUpdatesEnabled(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(888)

	if err := storage.SetUserUpdatesEnabled(userID, false); err != nil {
		t.Fatalf("SetUserUpdatesEnabled(false) failed: %v", err)
	}

	settings, err := storage.GetUserUpdateSettings(userID)
	if err != nil {
		t.Fatalf("GetUserUpdateSettings failed: %v", err)
	}
	if settings.Enabled {
		t.Fatal("expected enabled=false")
	}

	if err := storage.SetUserUpdatesEnabled(userID, true); err != nil {
		t.Fatalf("SetUserUpdatesEnabled(true) failed: %v", err)
	}

	settings, err = storage.GetUserUpdateSettings(userID)
	if err != nil {
		t.Fatalf("GetUserUpdateSettings failed: %v", err)
	}
	if !settings.Enabled {
		t.Fatal("expected enabled=true")
	}
}

func TestStorage_SetUserUpdateInterval(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(999)

	if err := storage.SetUserUpdateInterval(userID, 60); err != nil {
		t.Fatalf("SetUserUpdateInterval(60) failed: %v", err)
	}

	settings, err := storage.GetUserUpdateSettings(userID)
	if err != nil {
		t.Fatalf("GetUserUpdateSettings failed: %v", err)
	}
	if settings.IntervalMinutes != 60 {
		t.Fatalf("expected interval 60, got %d", settings.IntervalMinutes)
	}
}

func TestStorage_SetUserUpdateInterval_Invalid(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(1001)

	if err := storage.SetUserUpdateInterval(userID, 15); err == nil {
		t.Fatal("expected error for invalid interval 15")
	}
}

func TestStorage_SetNextCheckUnix(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	userID := int64(1002)
	nextCheckUnix := time.Now().Unix() + 1234

	if err := storage.SetNextCheckUnix(userID, nextCheckUnix); err != nil {
		t.Fatalf("SetNextCheckUnix failed: %v", err)
	}

	settings, err := storage.GetUserUpdateSettings(userID)
	if err != nil {
		t.Fatalf("GetUserUpdateSettings failed: %v", err)
	}
	if settings.NextCheckUnix != nextCheckUnix {
		t.Fatalf("expected next_check_unix %d, got %d", nextCheckUnix, settings.NextCheckUnix)
	}
}

func TestStorage_ListDueUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage := NewStorage(db)
	nowUnix := time.Now().Unix()

	dueUser := int64(2001)
	disabledUser := int64(2002)
	futureUser := int64(2003)

	if err := storage.EnsureUserUpdateSettings(dueUser); err != nil {
		t.Fatalf("EnsureUserUpdateSettings failed: %v", err)
	}
	if err := storage.EnsureUserUpdateSettings(disabledUser); err != nil {
		t.Fatalf("EnsureUserUpdateSettings failed: %v", err)
	}
	if err := storage.EnsureUserUpdateSettings(futureUser); err != nil {
		t.Fatalf("EnsureUserUpdateSettings failed: %v", err)
	}

	if err := storage.SetNextCheckUnix(dueUser, nowUnix-1); err != nil {
		t.Fatalf("SetNextCheckUnix due user failed: %v", err)
	}
	if err := storage.SetUserUpdatesEnabled(disabledUser, false); err != nil {
		t.Fatalf("SetUserUpdatesEnabled(false) failed: %v", err)
	}
	if err := storage.SetNextCheckUnix(disabledUser, nowUnix-1); err != nil {
		t.Fatalf("SetNextCheckUnix disabled user failed: %v", err)
	}
	if err := storage.SetNextCheckUnix(futureUser, nowUnix+3600); err != nil {
		t.Fatalf("SetNextCheckUnix future user failed: %v", err)
	}

	dueUsers, err := storage.ListDueUsers(nowUnix)
	if err != nil {
		t.Fatalf("ListDueUsers failed: %v", err)
	}
	if len(dueUsers) != 1 {
		t.Fatalf("expected exactly 1 due user, got %d (%v)", len(dueUsers), dueUsers)
	}
	if dueUsers[0] != dueUser {
		t.Fatalf("expected due user %d, got %d", dueUser, dueUsers[0])
	}
}
