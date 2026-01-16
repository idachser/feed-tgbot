package main

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	sendMsg(ctx, b, update.Message.Chat.ID, "Hello! I am a bot for RSS feeds.\n\nCommands:\n/add <url> - add a feed\n/list - my feeds\n/news - latest news")
}

func addHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	args := extractArgs(update.Message.Text, "/add")

	if args == "" {
		sendMsg(ctx, b, update.Message.Chat.ID, "Usage: /add <url>")
		return
	}

	urls := splitArgs(args)

	for _, url := range urls {
		if !isValidURL(url) {
			sendMsg(ctx, b, update.Message.Chat.ID, "Wrong URL: "+url)
			continue
		}

		// add to storage
		sendMsg(ctx, b, update.Message.Chat.ID, "Added URL: "+url)
	}
}

func listHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// userID := update.Message.From.ID

	sendMsg(ctx, b, update.Message.Chat.ID, "Your feeds: ...")
}

func newsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	sendMsg(ctx, b, update.Message.Chat.ID, "Loading news...")
}

func removeHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	args := extractArgs(update.Message.Text, "/remove")

	if args == "" {
		sendMsg(ctx, b, update.Message.Chat.ID, "Usage: /remove <url>")
		return
	}

	urls := splitArgs(args)

	for _, url := range urls {
		if !isValidURL(url) {
			sendMsg(ctx, b, update.Message.Chat.ID, "Wrong URL: "+url)
			continue
		}

		// remove from storage
		sendMsg(ctx, b, update.Message.Chat.ID, "Feed "+url+" removed")
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
