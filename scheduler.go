package main

import (
	"context"
	"log"
	"time"

	"github.com/go-telegram/bot"
)

func startScheduler(ctx context.Context, b *bot.Bot, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("scheduler started with interval: %v", interval)

	sendNews(ctx, b)

	for {
		select {
		case <-ticker.C:

			sendNews(ctx, b)

		case <-ctx.Done():
			log.Printf("scheduler stopped")
			return
		}
	}
}

func sendNews(ctx context.Context, b *bot.Bot) {
	// send news
}
