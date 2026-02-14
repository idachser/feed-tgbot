package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const helpText = "/start - greeting\n/add <url> - add feed\n/list - show my feeds\n/news - choose a feed source and get latest 10 news items\n/remove <url> - remove feed\n/updates - manage automatic updates\n/help - show help"

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	sendMsgWithKeyboard(ctx, b, update.Message.Chat.ID, "Hello! I am a bot for RSS feeds.\n\nCommands:\n"+helpText)
}

func helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	sendMsgWithKeyboard(ctx, b, update.Message.Chat.ID, helpText)
}

// handler for add RSS
func addHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	args := extractArgs(update.Message.Text, "/add")
	chatID := update.Message.Chat.ID

	if args == "" {
		sendMsg(ctx, b, chatID, "Usage: /add <url>")
		return
	}

	userID := update.Message.From.ID
	if err := storage.DeletePendingAction(userID); err != nil {
		log.Printf("error deleting pending action: %v", err)
	}
	processAddInput(ctx, b, chatID, userID, args)
}

func processAddInput(ctx context.Context, b *bot.Bot, chatID, userID int64, args string) {
	urls := splitArgs(args)

	for _, url := range urls {
		if !isValidURL(url) {
			sendMsg(ctx, b, chatID, "❌ Wrong URL: "+url)
			continue
		}

		_, err := getFeedsWithContext(ctx, url)
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
		feeds, err := DiscoverFeedsWithContext(ctx, url)
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

func addButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	if err := storage.SetPendingAction(userID, addPendingAction); err != nil {
		log.Printf("error setting pending action: %v", err)
		sendMsg(ctx, b, chatID, "❌ Error preparing add action. Try again.")
		return
	}

	sendMsg(ctx, b, chatID, "Send feed URL (or page URL) to add.")
}

// handler for showing inline keyboard feed selection
const pendingSelectionMaxAge = 15 * time.Minute
const addPendingAction = "add_waiting_url"
const callbackAddFeed = "add_feed"
const callbackRemoveFeed = "remove_feed"
const callbackNewsFeed = "news_feed"
const callbackUpdatesToggle = "updates_toggle"
const callbackUpdatesInterval = "updates_interval"

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
		callbackData := fmt.Sprintf("%s:%d:%d", callbackAddFeed, userID, i)

		button := models.InlineKeyboardButton{
			Text:         fmt.Sprintf("%s (%s)", feed.Title, feed.Type),
			CallbackData: callbackData,
		}

		buttons = append(buttons, []models.InlineKeyboardButton{button})
	}

	keyboard := models.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	if err := storage.SetPendingFeeds(userID, feeds); err != nil {
		log.Printf("Error saving pending feeds: %v", err)
		sendMsg(ctx, b, chatID, "❌ Error preparing feed selection")
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        fmt.Sprintf("Found %d feeds. Choose one:", len(feeds)),
		ReplyMarkup: keyboard,
	})
}

func callbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	callback := update.CallbackQuery

	parts := strings.Split(callback.Data, ":")
	if len(parts) != 3 {
		return
	}

	callbackType := parts[0]
	if callbackType != callbackAddFeed && callbackType != callbackRemoveFeed && callbackType != callbackNewsFeed && callbackType != callbackUpdatesToggle && callbackType != callbackUpdatesInterval {
		return
	}

	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}

	if userID != callback.From.ID {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "This button is not for you",
			ShowAlert:       true,
		})
		return
	}

	switch callbackType {
	case callbackAddFeed:
		feedIndex, err := strconv.Atoi(parts[2])
		if err != nil || feedIndex < 0 {
			return
		}
		handleAddFeedCallback(ctx, b, callback, userID, feedIndex)
	case callbackRemoveFeed:
		feedIndex, err := strconv.Atoi(parts[2])
		if err != nil || feedIndex < 0 {
			return
		}
		handleRemoveFeedCallback(ctx, b, callback, userID, feedIndex)
	case callbackNewsFeed:
		feedIndex, err := strconv.Atoi(parts[2])
		if err != nil || feedIndex < 0 {
			return
		}
		handleNewsFeedCallback(ctx, b, callback, userID, feedIndex)
	case callbackUpdatesToggle:
		handleUpdatesToggleCallback(ctx, b, callback, userID, parts[2])
	case callbackUpdatesInterval:
		intervalMinutes, err := strconv.Atoi(parts[2])
		if err != nil {
			return
		}
		handleUpdatesIntervalCallback(ctx, b, callback, userID, intervalMinutes)
	}
}

func handleAddFeedCallback(ctx context.Context, b *bot.Bot, callback *models.CallbackQuery, userID int64, feedIndex int) {
	feeds, ok, err := storage.GetPendingFeeds(userID, pendingSelectionMaxAge)
	if err != nil {
		log.Printf("Error loading pending feeds: %v", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Error loading selection. Try /add again",
			ShowAlert:       true,
		})
		return
	}
	if !ok || feedIndex >= len(feeds) {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Feed selection expired. Try /add again",
			ShowAlert:       true,
		})
		return
	}

	selectedFeed := feeds[feedIndex]
	err = storage.AddFeed(userID, selectedFeed.URL)
	if err != nil {
		log.Printf("Error adding feed: %v", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Error adding feed",
			ShowAlert:       true,
		})
		return
	}

	if err := storage.DeletePendingFeeds(userID); err != nil {
		log.Printf("Error deleting pending feeds: %v", err)
	}

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

func handleRemoveFeedCallback(ctx context.Context, b *bot.Bot, callback *models.CallbackQuery, userID int64, feedIndex int) {
	feeds, ok, err := storage.GetPendingRemoveFeeds(userID, pendingSelectionMaxAge)
	if err != nil {
		log.Printf("Error loading pending remove feeds: %v", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Error loading selection. Try Remove again",
			ShowAlert:       true,
		})
		return
	}
	if !ok || feedIndex >= len(feeds) {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Selection expired. Tap Remove again",
			ShowAlert:       true,
		})
		return
	}

	selectedFeed := feeds[feedIndex]
	removed, err := storage.RemoveFeed(userID, selectedFeed)
	if err != nil {
		log.Printf("Error removing feed: %v", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Error removing feed",
			ShowAlert:       true,
		})
		return
	}
	if !removed {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Feed not found anymore",
			ShowAlert:       true,
		})
		return
	}

	if err := storage.DeletePendingRemoveFeeds(userID); err != nil {
		log.Printf("Error deleting pending remove feeds: %v", err)
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callback.ID,
		Text:            "✅ Feed removed!",
	})

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    callback.Message.Message.Chat.ID,
		MessageID: callback.Message.Message.ID,
		Text:      fmt.Sprintf("✅ Removed:\n%s", selectedFeed),
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

	if err := storage.SetPendingNewsFeeds(userID, feeds); err != nil {
		log.Printf("error saving pending news feeds: %v", err)
		sendMsg(ctx, b, chatID, "❌ Error preparing news source selection")
		return
	}

	showNewsSourceSelection(ctx, b, chatID, userID, feeds)
}

func showNewsSourceSelection(ctx context.Context, b *bot.Bot, chatID, userID int64, feeds []string) {
	keyboard := newsSourceSelectionKeyboard(userID, feeds)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Choose feed source for latest news:",
		ReplyMarkup: keyboard,
	})
}

func newsSourceSelectionKeyboard(userID int64, feeds []string) models.InlineKeyboardMarkup {
	var buttons [][]models.InlineKeyboardButton
	for i, feed := range feeds {
		buttons = append(buttons, []models.InlineKeyboardButton{
			{
				Text:         "📰 " + truncate(feed, 56),
				CallbackData: fmt.Sprintf("%s:%d:%d", callbackNewsFeed, userID, i),
			},
		})
	}

	return models.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}
}

func handleNewsFeedCallback(ctx context.Context, b *bot.Bot, callback *models.CallbackQuery, userID int64, feedIndex int) {
	feeds, ok, err := storage.GetPendingNewsFeeds(userID, pendingSelectionMaxAge)
	if err != nil {
		log.Printf("Error loading pending news feeds: %v", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Error loading selection. Try /news again",
			ShowAlert:       true,
		})
		return
	}
	if !ok || feedIndex >= len(feeds) {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Selection expired. Run /news again",
			ShowAlert:       true,
		})
		return
	}

	selectedFeedURL := feeds[feedIndex]

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callback.ID,
		Text:            "Loading news...",
	})

	news, err := getFeedsWithContext(ctx, selectedFeedURL)
	if err != nil {
		log.Printf("error fetching selected feed %s: %v", selectedFeedURL, err)
		sendMsg(ctx, b, callback.Message.Message.Chat.ID, "❌ Failed to fetch news from selected source")
		return
	}

	if len(news) == 0 {
		sendMsg(ctx, b, callback.Message.Message.Chat.ID, "No news found for selected source.")
		return
	}

	sort.Slice(news, func(i, j int) bool {
		if news[i].Published == nil {
			return false
		}
		if news[j].Published == nil {
			return true
		}
		return news[i].Published.After(*news[j].Published)
	})

	limit := 10
	if len(news) < limit {
		limit = len(news)
	}

	for i := 0; i < limit; i++ {
		item := news[i]
		pubStr := "Unknown date"
		if item.Published != nil {
			pubStr = item.Published.Format("2006-01-02 15:04")
		}

		message := buildNewsMessage(item, pubStr)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    callback.Message.Message.Chat.ID,
			Text:      message,
			ParseMode: models.ParseModeHTML,
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
	if err := storage.DeletePendingAction(userID); err != nil {
		log.Printf("error deleting pending action: %v", err)
	}

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

func removeButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	if err := storage.DeletePendingAction(userID); err != nil {
		log.Printf("error deleting pending action: %v", err)
	}

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

	if err := storage.SetPendingRemoveFeeds(userID, feeds); err != nil {
		log.Printf("error saving pending remove feeds: %v", err)
		sendMsg(ctx, b, chatID, "❌ Error preparing remove list")
		return
	}

	showRemoveSelection(ctx, b, chatID, userID, feeds)
}

func showRemoveSelection(ctx context.Context, b *bot.Bot, chatID, userID int64, feeds []string) {
	keyboard := removeSelectionKeyboard(userID, feeds)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Choose feed to remove:",
		ReplyMarkup: keyboard,
	})
}

func removeSelectionKeyboard(userID int64, feeds []string) models.InlineKeyboardMarkup {
	var buttons [][]models.InlineKeyboardButton
	for i, feed := range feeds {
		buttons = append(buttons, []models.InlineKeyboardButton{
			{
				Text:         "❌ " + truncate(feed, 56),
				CallbackData: fmt.Sprintf("%s:%d:%d", callbackRemoveFeed, userID, i),
			},
		})
	}

	return models.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}
}

func updatesHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	if err := storage.DeletePendingAction(userID); err != nil {
		log.Printf("error deleting pending action: %v", err)
	}

	settings, err := storage.GetUserUpdateSettings(userID)
	if err != nil {
		log.Printf("error loading user update settings: %v", err)
		sendMsg(ctx, b, chatID, "Error loading update settings.")
		return
	}

	sendUpdatesSettingsMessage(ctx, b, chatID, settings)
}

func sendUpdatesSettingsMessage(ctx context.Context, b *bot.Bot, chatID int64, settings UserUpdateSettings) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        updatesSettingsText(settings),
		ReplyMarkup: updatesSettingsKeyboard(settings),
	})
}

func updatesSettingsText(settings UserUpdateSettings) string {
	status := "OFF"
	if settings.Enabled {
		status = "ON"
	}

	return fmt.Sprintf(
		"Automatic updates: %s\nCheck frequency: every %d minutes\n\nUse buttons to change settings.",
		status,
		settings.IntervalMinutes,
	)
}

func updatesSettingsKeyboard(settings UserUpdateSettings) models.InlineKeyboardMarkup {
	enableText := "Enable"
	disableText := "Disable"
	if settings.Enabled {
		enableText += " ✅"
	} else {
		disableText += " ✅"
	}

	var intervalButtons []models.InlineKeyboardButton
	for _, interval := range []int{30, 60, 360} {
		text := fmt.Sprintf("%dm", interval)
		if settings.IntervalMinutes == interval {
			text += " ✅"
		}

		intervalButtons = append(intervalButtons, models.InlineKeyboardButton{
			Text:         text,
			CallbackData: fmt.Sprintf("%s:%d:%d", callbackUpdatesInterval, settings.UserID, interval),
		})
	}

	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         enableText,
					CallbackData: fmt.Sprintf("%s:%d:on", callbackUpdatesToggle, settings.UserID),
				},
				{
					Text:         disableText,
					CallbackData: fmt.Sprintf("%s:%d:off", callbackUpdatesToggle, settings.UserID),
				},
			},
			intervalButtons,
		},
	}
}

func handleUpdatesToggleCallback(ctx context.Context, b *bot.Bot, callback *models.CallbackQuery, userID int64, toggleValue string) {
	var enabled bool
	switch toggleValue {
	case "on":
		enabled = true
	case "off":
		enabled = false
	default:
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Invalid toggle value",
			ShowAlert:       true,
		})
		return
	}

	if err := storage.SetUserUpdatesEnabled(userID, enabled); err != nil {
		log.Printf("error setting updates enabled: %v", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Could not update setting",
			ShowAlert:       true,
		})
		return
	}

	settings, err := storage.GetUserUpdateSettings(userID)
	if err != nil {
		log.Printf("error loading update settings: %v", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Could not load updated settings",
			ShowAlert:       true,
		})
		return
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callback.ID,
		Text:            "Settings updated",
	})

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      callback.Message.Message.Chat.ID,
		MessageID:   callback.Message.Message.ID,
		Text:        updatesSettingsText(settings),
		ReplyMarkup: updatesSettingsKeyboard(settings),
	})
}

func handleUpdatesIntervalCallback(ctx context.Context, b *bot.Bot, callback *models.CallbackQuery, userID int64, intervalMinutes int) {
	if err := storage.SetUserUpdateInterval(userID, intervalMinutes); err != nil {
		log.Printf("error setting update interval: %v", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Invalid interval",
			ShowAlert:       true,
		})
		return
	}

	settings, err := storage.GetUserUpdateSettings(userID)
	if err != nil {
		log.Printf("error loading update settings: %v", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Could not load updated settings",
			ShowAlert:       true,
		})
		return
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callback.ID,
		Text:            "Interval updated",
	})

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      callback.Message.Message.Chat.ID,
		MessageID:   callback.Message.Message.ID,
		Text:        updatesSettingsText(settings),
		ReplyMarkup: updatesSettingsKeyboard(settings),
	})
}

// send message with help instructions
func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	action, ok, err := storage.GetPendingAction(userID, pendingSelectionMaxAge)
	if err != nil {
		log.Printf("error getting pending action: %v", err)
		sendMsg(ctx, b, chatID, "Error loading pending action.")
		return
	}
	if ok && action == addPendingAction {
		urls := splitArgs(strings.TrimSpace(update.Message.Text))
		if len(urls) == 0 {
			sendMsg(ctx, b, chatID, "Send feed URL (or page URL) to add.")
			return
		}
		for _, url := range urls {
			if !isValidURL(url) {
				sendMsg(ctx, b, chatID, "Please send a valid URL starting with http:// or https://")
				return
			}
		}

		if err := storage.DeletePendingAction(userID); err != nil {
			log.Printf("error deleting pending action: %v", err)
		}
		processAddInput(ctx, b, chatID, userID, strings.TrimSpace(update.Message.Text))
		return
	}

	sendMsg(ctx, b, update.Message.Chat.ID, helpText)
}

func commandReplyKeyboard() models.ReplyKeyboardMarkup {
	return models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: "Add"},
				{Text: "Remove"},
			},
			{
				{Text: "News"},
				{Text: "List"},
			},
			{
				{Text: "Help"},
				{Text: "Updates"},
			},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
}

func sendMsgWithKeyboard(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	keyboard := commandReplyKeyboard()
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: keyboard,
	})
}

func sendMsg(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
}
