package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}

	b, err := bot.New(os.Getenv("TG_BOT_TOKEN"), opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, startHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/add", bot.MatchTypePrefix, addHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/list", bot.MatchTypeExact, listHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/news", bot.MatchTypeExact, newsHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/remove", bot.MatchTypePrefix, removeHandler)

	b.Start(ctx)
}

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Hello! I am a bot for RSS feeds.\n\nCommands:\n/add <url> - add a feed\n/list - my feeds\n/news - latest news",
	})
}

func addHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	text := update.Message.Text

	if len(text) <= 5 { // "/add " = 5 chars
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.ID,
			Text:   "Usage: /add <url>",
		})
		return
	}

	url := text[5:]

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.ID,
		Text:   "Feed " + url + " added",
	})
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
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "/start - greeting\n/add <url> - add feed\n/list - show my feeds\n/news - get the latest 10 news items from all feeds\n/remove <url> - remove feed",
	})
}
