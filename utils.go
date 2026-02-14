package main

import (
	"fmt"
	"html"
	u "net/url"
	"regexp"
	"strings"
)

var (
	feedScriptTagRe = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	feedStyleTagRe  = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	feedBlockTagRe  = regexp.MustCompile(`(?is)</?(p|br|div|li|ul|ol|h[1-6]|tr|td|th|blockquote|article|section)[^>]*>`)
	feedAnyTagRe    = regexp.MustCompile(`(?s)<[^>]+>`)
	feedWhitespace  = regexp.MustCompile(`\s+`)
)

func extractArgs(text, command string) string {
	args := strings.TrimPrefix(text, command)
	return strings.TrimSpace(args)
}

func splitArgs(args string) []string {
	return strings.Fields(args)
}

func isValidURL(url string) bool {
	_, err := u.Parse(url)
	if err != nil {
		return false
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}

	return true
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func sourceNameFromURL(raw string) string {
	parsedURL, err := u.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}

	host := strings.ToLower(strings.TrimSpace(parsedURL.Hostname()))
	host = strings.TrimPrefix(host, "www.")
	if host == "" {
		return ""
	}

	parts := strings.Split(host, ".")
	if len(parts) == 1 {
		return formatSourceToken(parts[0])
	}

	token := parts[len(parts)-2]
	if len(token) <= 2 && len(parts) >= 3 {
		token = parts[len(parts)-3]
	}

	name := formatSourceToken(token)
	if name != "" {
		return name
	}

	return formatSourceToken(host)
}

func formatSourceToken(token string) string {
	token = strings.Trim(strings.ToLower(token), ".-_ ")
	if token == "" {
		return ""
	}

	segments := strings.FieldsFunc(token, func(r rune) bool {
		return r == '-' || r == '_'
	})
	if len(segments) == 0 {
		return ""
	}

	words := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment == "" {
			continue
		}

		if len(segment) <= 4 {
			words = append(words, strings.ToUpper(segment))
			continue
		}

		words = append(words, strings.ToUpper(segment[:1])+segment[1:])
	}

	if len(words) == 0 {
		return ""
	}

	return strings.Join(words, " ")
}

func sourceButtonLabel(prefix, rawURL string, maxLen int) string {
	sourceName := sourceNameFromURL(rawURL)
	if sourceName == "" {
		sourceName = truncate(rawURL, maxLen)
	}

	return truncate(strings.TrimSpace(prefix+" "+sourceName), maxLen)
}

func sanitizeFeedText(raw string) string {
	if raw == "" {
		return ""
	}

	text := feedScriptTagRe.ReplaceAllString(raw, " ")
	text = feedStyleTagRe.ReplaceAllString(text, " ")
	text = feedBlockTagRe.ReplaceAllString(text, " ")
	text = feedAnyTagRe.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = feedWhitespace.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

func cleanFeedTitle(raw string) string {
	title := sanitizeFeedText(raw)
	if title == "" {
		return "Untitled"
	}
	return title
}

func cleanFeedDescription(raw string) string {
	description := sanitizeFeedText(raw)
	if description == "" {
		return "(No description provided)"
	}
	return description
}

func escapeTelegramHTML(text string) string {
	return html.EscapeString(text)
}

func formatTelegramLink(url, label string) string {
	safeLabel := escapeTelegramHTML(label)
	if isValidURL(url) {
		return fmt.Sprintf(`<a href="%s">%s</a>`, escapeTelegramHTML(url), safeLabel)
	}
	return fmt.Sprintf("%s: %s", safeLabel, escapeTelegramHTML(url))
}

func buildNewsMessage(item FeedItem, pubStr string) string {
	title := escapeTelegramHTML(cleanFeedTitle(item.Title))
	description := escapeTelegramHTML(cleanFeedDescription(item.Description))
	articleLink := formatTelegramLink(item.Link, "Read article")
	dateText := escapeTelegramHTML(pubStr)

	return fmt.Sprintf(
		"📰 <b>%s</b>\n\n%s\n\n🔗 %s\n📅 %s",
		title,
		description,
		articleLink,
		dateText,
	)
}

func buildAutoNewsMessage(feedURL string, item FeedItem) string {
	feedLink := formatTelegramLink(feedURL, "Feed")
	title := escapeTelegramHTML(cleanFeedTitle(item.Title))
	description := escapeTelegramHTML(cleanFeedDescription(item.Description))
	articleLink := formatTelegramLink(item.Link, "Read article")

	return fmt.Sprintf(
		"🆕 New from %s\n\n📰 <b>%s</b>\n\n%s\n\n🔗 %s",
		feedLink,
		title,
		description,
		articleLink,
	)
}
