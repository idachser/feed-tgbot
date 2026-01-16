package main

import (
	"sync"
)

type Storage struct {
	mu           sync.RWMutex
	subscription map[int64][]string
}

func NewStorage() *Storage {
	return &Storage{
		subscription: make(map[int64][]string),
	}
}

func (s *Storage) AddFeed(userID int64, feedURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	feeds := s.subscription[userID]
	for _, url := range feeds {
		if url == feedURL {
			return nil
		}
	}

	s.subscription[userID] = append(feeds, feedURL)
	return nil
}

func (s *Storage) RemoveFeed(userID int64, feedURL string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	feeds := s.subscription[userID]
	for i, url := range feeds {
		if url == feedURL {
			s.subscription[userID] = append(feeds[:i], feeds[i+1:]...)
			return true
		}
	}

	return false
}

func (s *Storage) GetFeeds(userID int64) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	feeds := s.subscription[userID]
	result := make([]string, len(feeds))
	copy(result, feeds)

	return result
}

func (s *Storage) GetAllUsers() []int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]int64, len(s.subscription))
	for userID := range s.subscription {
		users = append(users, userID)
	}

	return users
}
