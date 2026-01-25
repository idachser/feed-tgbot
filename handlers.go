package main

import (
	"context"
	"fmt"
	"sort"
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
	for i, url := range feeds {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, url))
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

	var allNews []FeedItem

	for _, url := range feeds {
		news, err := getFeeds(url)
		if err != nil {
			continue
		}
		allNews = append(allNews, news...)
	}

	if len(allNews) == 0 {
		sendMsg(ctx, b, chatID, "No news found")
		return
	}

	sort.Slice(allNews, func(i, j int) bool {
		if allNews[i].Published == nil {
			return false
		}
		if allNews[j].Published == nil {
			return true
		}
		return allNews[i].Published.After(*allNews[j].Published)
	})

	limit := 10
	if len(allNews) < limit {
		limit = len(allNews)
	}

	for i := 0; i < limit; i++ {
		item := allNews[i]

		pubStr := "Unknown date"
		if item.Published != nil {
			pubStr = item.Published.Format("2006-01-02 15:04")
		}

		message := fmt.Sprintf("📰 *%s*\n\n%s\n\n🔗 %s\n📅 %s",
			item.Title,
			item.Description,
			item.Link,
			pubStr,
		)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      message,
			ParseMode: models.ParseModeMarkdown,
		})
	}
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
