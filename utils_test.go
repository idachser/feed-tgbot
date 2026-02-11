package main

import (
	"testing"
)

func TestExtractArgs(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		command  string
		expected string
	}{
		{
			name:     "simple url",
			text:     "/add https://example.com/feed.rss",
			command:  "/add",
			expected: "https://example.com/feed.rss",
		},
		{
			name:     "no args",
			text:     "/add",
			command:  "/add",
			expected: "",
		},
		{
			name:     "multiple args",
			text:     "/add https://example.com/feed.rss MyFeed",
			command:  "/add",
			expected: "https://example.com/feed.rss MyFeed",
		},
		{
			name:     "extra spaces",
			text:     "/add    https://example.com/feed.rss   ",
			command:  "/add",
			expected: "https://example.com/feed.rss",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractArgs(tt.text, tt.command)
			if result != tt.expected {
				t.Errorf("extractArgs(%q, %q) = %q; want %q",
					tt.text, tt.command, result, tt.expected)
			}
		})
	}
}

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		expected []string
	}{
		{
			name:     "single arg",
			args:     "https://example.com",
			expected: []string{"https://example.com"},
		},
		{
			name:     "multiple args",
			args:     "https://example.com MyFeed",
			expected: []string{"https://example.com", "MyFeed"},
		},
		{
			name:     "empty string",
			args:     "",
			expected: []string{},
		},
		{
			name:     "multiple spaces",
			args:     "arg1    arg2   arg3",
			expected: []string{"arg1", "arg2", "arg3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitArgs(tt.args)

			if len(result) != len(tt.expected) {
				t.Errorf("splitArgs(%q) length = %d; want %d",
					tt.args, len(result), len(tt.expected))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitArgs(%q)[%d] = %q; want %q",
						tt.args, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid http",
			url:      "http://example.com",
			expected: true,
		},
		{
			name:     "valid https",
			url:      "https://example.com/feed.rss",
			expected: true,
		},
		{
			name:     "no protocol",
			url:      "example.com",
			expected: false,
		},
		{
			name:     "invalid protocol",
			url:      "ftp://example.com",
			expected: false,
		},
		{
			name:     "empty string",
			url:      "",
			expected: false,
		},
		{
			name:     "malformed url",
			url:      "ht!tp://exam ple.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidURL(tt.url)
			if result != tt.expected {
				t.Errorf("isValidURL(%q) = %v; want %v",
					tt.url, result, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLen   int
		expected string
	}{
		{
			name:     "shorter than limit",
			text:     "Hello",
			maxLen:   10,
			expected: "Hello",
		},
		{
			name:     "exactly at limit",
			text:     "Hello",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "longer than limit",
			text:     "Hello, World!",
			maxLen:   5,
			expected: "Hello...",
		},
		{
			name:     "empty string",
			text:     "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.text, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q; want %q",
					tt.text, tt.maxLen, result, tt.expected)
			}
		})
	}
}
