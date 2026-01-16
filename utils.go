package main

import (
	u "net/url"
	"strings"
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
