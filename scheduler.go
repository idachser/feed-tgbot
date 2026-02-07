package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-telegram/bot"
)

func startScheduler(ctx context.Context, b *bot.Bot, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("scheduler started with interval: %v", interval)

	checkAndSendNews(ctx, b)

	for {
		select {
		case <-ticker.C:
			checkAndSendNews(ctx, b)

		case <-ctx.Done():
			log.Printf("scheduler stopped")
			return
		}
	}
}

func checkAndSendNews(ctx context.Context, b *bot.Bot) {
	log.Println("check feeds for all users...")

	users, err := storage.GetAllUsers()
	if err != nil {
		log.Printf("error getting users: %v", err)
		return
	}

	for _, userID := range users {
		feeds, err := storage.GetFeeds(userID)
		if err != nil {
			log.Printf("error getting urls of feeds for user %d: %v", userID, err)
			continue
		}

		for _, feedURL := range feeds {
			lastSent, err := storage.GetLastSent(userID, feedURL)
			if err != nil {
				log.Printf("error getting last sent for user %d: %v", userID, err)
				continue
			}

			newItems, err := getNewItems(feedURL, lastSent)
			if err != nil {
				log.Printf("error fetching feed %s: %v", feedURL, err)
				continue
			}

			if len(newItems) == 0 {
				continue
			}

			sendNewsToUser(ctx, b, userID, feedURL, newItems)
		}
	}
}

func sendNewsToUser(ctx context.Context, b *bot.Bot, userID int64, feedURL string, items []FeedItem) {
	for _, item := range items {
		message := fmt.Sprintf("🆕 *New from %s*\n\n📰 *%s*\n\n%s\n\n🔗 %s",
			feedURL,
			item.Title,
			item.Description,
			item.Link,
		)

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: userID,
			Text:   message,
		})
		if err != nil {
			log.Printf("error sending message to user %d: %v", userID, err)
			continue
		}

		err = storage.SetLastSent(userID, feedURL, item.Link)
		if err != nil {
			log.Printf("error setting last sent: %v", err)
		}
	}
	log.Printf("sent %d news items to user %d from %s", len(items), userID, feedURL)
}
