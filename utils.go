package main

import (
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
