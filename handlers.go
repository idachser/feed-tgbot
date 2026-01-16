package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	sendMsg(ctx, b, update.Message.Chat.ID, "Hello! I am a bot for RSS feeds.\n\nCommands:\n/add <url> - add a feed\n/list - my feeds\n/news - latest news")
}

func addHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	args := extractArgs(update.Message.Text, "/add")
	chatID := update.Message.Chat.ID

	if args == "" {
		sendMsg(ctx, b, chatID, "Usage: /add <url>")
		return
	}

	urls := splitArgs(args)
	userID := update.Message.From.ID

	for _, url := range urls {
		if !isValidURL(url) {
			sendMsg(ctx, b, chatID, "Wrong URL: "+url)
			continue
		}

		err := storage.AddFeed(userID, url)
		if err != nil {
			sendMsg(ctx, b, chatID, "Error adding feed: "+url)
			continue
		}

		sendMsg(ctx, b, chatID, "Added URL: "+url)
	}
}

func listHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	feeds := storage.GetFeeds(userID)
	if len(feeds) == 0 {
		sendMsg(ctx, b, chatID, "You have no feeds yet. Use /add <url>")
		return
	}

	var sb strings.Builder
	sb.WriteString("Your feeds:\n\n")
	for i, feed := range feeds {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, feed))
	}

	sendMsg(ctx, b, chatID, sb.String())
}

func newsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	feeds := storage.GetFeeds(userID)
	if len(feeds) == 0 {
		sendMsg(ctx, b, chatID, "You have no feeds yet. Use /add <url>")
		return
	}

	sendMsg(ctx, b, chatID, "Loading news...")

	// get news
	sendMsg(ctx, b, chatID, "Here will be news")
}

func removeHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	args := extractArgs(update.Message.Text, "/remove")
	chatID := update.Message.Chat.ID

	if args == "" {
		sendMsg(ctx, b, chatID, "Usage: /remove <url>")
		return
	}

	urls := splitArgs(args)
	userID := update.Message.From.ID

	for _, url := range urls {
		if !isValidURL(url) {
			sendMsg(ctx, b, chatID, "Wrong URL: "+url)
			continue
		}

		removed := storage.RemoveFeed(userID, url)
		if !removed {
			sendMsg(ctx, b, chatID, "Feed "+url+" not found")
			continue
		}

		sendMsg(ctx, b, chatID, "Feed "+url+" removed")
	}
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	sendMsg(ctx, b, update.Message.Chat.ID, "/start - greeting\n/add <url> - add feed\n/list - show my feeds\n/news - gvet the latest 10 news items from all feeds\n/remove <url> - remove feed")
}

func sendMsg(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
}
