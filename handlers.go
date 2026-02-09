package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	sendMsg(ctx, b, update.Message.Chat.ID, "Hello! I am a bot for RSS feeds.\n\nCommands:\n/add <url> - add a feed\n/list - my feeds\n/news - latest news")
}

// handler for add RSS
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
			sendMsg(ctx, b, chatID, "❌ Wrong URL: "+url)
			continue
		}

		_, err := getFeeds(url)
		if err == nil {
			err := storage.AddFeed(userID, url)
			if err != nil {
				log.Printf("error adding feed: %v", err)
				sendMsg(ctx, b, chatID, "❌ Error adding feed: "+url)
				continue
			}

			sendMsg(ctx, b, chatID, "✅ Added URL: "+url)
			continue
		}

		log.Printf("Not a direct feed, discovering feeds on %s", url)
		feeds, err := DiscoverFeeds(url)
		if err != nil {
			log.Printf("Error discovering feeds: %v", err)
			sendMsg(ctx, b, chatID, "❌ Could not find feeds on this page")
			continue
		}

		if len(feeds) == 0 {
			sendMsg(ctx, b, chatID, "❌ No RSS/Atom feeds found on this page")
			continue
		}

		showFeedSelection(ctx, b, chatID, userID, feeds)
	}
}

// handler for showing inline keyboard feed selection
var pendingFeeds = make(map[int64][]DiscoveredFeed)

func showFeedSelection(ctx context.Context, b *bot.Bot, chatID, userID int64, feeds []DiscoveredFeed) {
	if len(feeds) == 1 {
		feed := feeds[0]
		err := storage.AddFeed(userID, feed.URL)
		if err != nil {
			log.Printf("Error adding feed: %v", err)
			sendMsg(ctx, b, chatID, "❌ Error adding feed")
			return
		}
		sendMsg(ctx, b, chatID, fmt.Sprintf("✅ Added: %s\n🔗 %s", feed.Title, feed.URL))
		return
	}

	var buttons [][]models.InlineKeyboardButton

	for i, feed := range feeds {
		callbackData := fmt.Sprintf("add_feed:%d:%d", userID, i)

		button := models.InlineKeyboardButton{
			Text:         fmt.Sprintf("%s (%s)", feed.Title, feed.Type),
			CallbackData: callbackData,
		}

		buttons = append(buttons, []models.InlineKeyboardButton{button})
	}

	keyboard := models.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	pendingFeeds[userID] = feeds

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        fmt.Sprintf("Found %d feeds. Choose one:", len(feeds)),
		ReplyMarkup: keyboard,
	})
}

func callbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	callback := update.CallbackQuery

	parts := strings.Split(callback.Data, ":")
	if len(parts) != 3 || parts[0] != "add_feed" {
		return
	}

	var userID, feedIndex int
	fmt.Sscanf(parts[1], "%d", &userID)
	fmt.Sscanf(parts[2], "%d", &feedIndex)

	if int64(userID) != callback.From.ID {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "This button is not for you",
			ShowAlert:       true,
		})
		return
	}

	feeds, ok := pendingFeeds[int64(userID)]
	if !ok || feedIndex >= len(feeds) {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Feed selection expired. Try /add again",
			ShowAlert:       true,
		})
		return
	}

	selectedFeed := feeds[feedIndex]

	err := storage.AddFeed(int64(userID), selectedFeed.URL)
	if err != nil {
		log.Printf("Error adding feed: %v", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Error adding feed",
			ShowAlert:       true,
		})
		return
	}

	delete(pendingFeeds, int64(userID))

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callback.ID,
		Text:            "✅ Feed added!",
	})

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    callback.Message.Message.Chat.ID,
		MessageID: callback.Message.Message.ID,
		Text:      fmt.Sprintf("✅ Added: %s\n🔗 %s", selectedFeed.Title, selectedFeed.URL),
	})
}

// handler for get list of RSS subscripted URLs
func listHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	feeds, err := storage.GetFeeds(userID)
	if err != nil {
		log.Printf("error getting feeds: %v", err)
		sendMsg(ctx, b, chatID, "Error loading feeds.")
		return
	}

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

// handler for get last 10 news from RSS
func newsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	feeds, err := storage.GetFeeds(userID)
	if err != nil {
		log.Printf("error getting feeds: %v", err)
		sendMsg(ctx, b, chatID, "Error loading feeds.")
		return
	}

	if len(feeds) == 0 {
		sendMsg(ctx, b, chatID, "You have no feeds yet. Use /add <url>")
		return
	}

	sendMsg(ctx, b, chatID, "Loading news...")

	var allNews []FeedItem

	for _, url := range feeds {
		news, err := getFeeds(url)
		if err != nil {

			log.Printf("error fetching feed %s: %v", url, err)

			continue
		}
		allNews = append(allNews, news...)
	}

	log.Printf("total news fetched = %d", len(allNews))

	if len(allNews) == 0 {
		sendMsg(ctx, b, chatID, "No news found.")
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

		log.Printf("item: %s", item.Title)

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

		log.Printf("message: %s", message)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   message,
		})
	}
}

// remove RSS
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

		removed, err := storage.RemoveFeed(userID, url)
		if err != nil {
			log.Printf("error removing feed %s: %v", url, err)
			sendMsg(ctx, b, chatID, "Error removing feed.")
			continue
		}

		if !removed {
			sendMsg(ctx, b, chatID, "Feed "+url+" not found.")
			continue
		}

		sendMsg(ctx, b, chatID, "Feed "+url+" removed.")
	}
}

// send message with help instructions
func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	sendMsg(ctx, b, update.Message.Chat.ID, "/start - greeting\n/add <url> - add feed\n/list - show my feeds\n/news - gvet the latest 10 news items from all feeds\n/remove <url> - remove feed")
}

func sendMsg(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
}
