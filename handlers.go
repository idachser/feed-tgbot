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
		}

		sendMsg(ctx, b, update.Message.Chat.ID, "Added URL: "+url)
	}
}

func listHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// userID := update.Message.From.ID

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Your feeds: ...",
	})
}

func newsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Loading news...",
	})
}

func removeHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	text := update.Message.Text

	if len(text) <= 8 { // "/remove " = 8 chars
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.ID,
			Text:   "Usage: /remove <url>",
		})
		return
	}

	url := text[5:]

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.ID,
		Text:   "Feed " + url + " deleted",
	})
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
