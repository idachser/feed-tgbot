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

func TestSanitizeFeedText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strips html tags and decodes entities",
			input:    "<p>Hello <b>world</b> &amp; team</p>",
			expected: "Hello world & team",
		},
		{
			name:     "removes script and style blocks",
			input:    `<style>.x{color:red}</style><script>alert("x")</script><div>Safe text</div>`,
			expected: "Safe text",
		},
		{
			name:     "normalizes whitespace and non breaking spaces",
			input:    "Line1&nbsp;&nbsp;\n\tLine2   Line3",
			expected: "Line1 Line2 Line3",
		},
		{
			name:     "handles malformed html fragments",
			input:    "<p>One<div>Two</p>Three",
			expected: "One Two Three",
		},
		{
			name:     "html only becomes empty",
			input:    "<p><br/></p>",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFeedText(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeFeedText(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCleanFeedTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "uses cleaned title when available",
			input:    "Tom &amp; Jerry",
			expected: "Tom & Jerry",
		},
		{
			name:     "falls back when empty",
			input:    "<p><br/></p>",
			expected: "Untitled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanFeedTitle(tt.input)
			if got != tt.expected {
				t.Errorf("cleanFeedTitle(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCleanFeedDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "uses cleaned description when available",
			input:    "<div>Summary &amp; details</div>",
			expected: "Summary & details",
		},
		{
			name:     "falls back when empty",
			input:    "<style>p{}</style>",
			expected: "(No description provided)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanFeedDescription(tt.input)
			if got != tt.expected {
				t.Errorf("cleanFeedDescription(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatTelegramLink(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		label    string
		expected string
	}{
		{
			name:     "valid url returns clickable html link",
			url:      "https://example.com?q=1&x=2",
			label:    "Read article",
			expected: `<a href="https://example.com?q=1&amp;x=2">Read article</a>`,
		},
		{
			name:     "invalid url falls back to plain escaped text",
			url:      "javascript:alert(1)",
			label:    "Read article",
			expected: "Read article: javascript:alert(1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTelegramLink(tt.url, tt.label)
			if got != tt.expected {
				t.Errorf("formatTelegramLink(%q, %q) = %q; want %q", tt.url, tt.label, got, tt.expected)
			}
		})
	}
}

func TestBuildNewsMessage(t *testing.T) {
	item := FeedItem{
		Title:       "A < B &amp; C",
		Description: "<p>Hello <b>world</b></p>",
		Link:        "https://example.com/post?id=1&src=rss",
	}

	got := buildNewsMessage(item, "2026-02-13 10:30")
	expected := "📰 <b>A &lt; B &amp; C</b>\n\nHello world\n\n🔗 <a href=\"https://example.com/post?id=1&amp;src=rss\">Read article</a>\n📅 2026-02-13 10:30"

	if got != expected {
		t.Errorf("buildNewsMessage() = %q; want %q", got, expected)
	}
}

func TestBuildAutoNewsMessage(t *testing.T) {
	item := FeedItem{
		Title:       "Title",
		Description: "",
		Link:        "https://example.com/post",
	}

	got := buildAutoNewsMessage("https://example.com/feed.xml", item)
	expected := "🆕 New from <a href=\"https://example.com/feed.xml\">Feed</a>\n\n📰 <b>Title</b>\n\n(No description provided)\n\n🔗 <a href=\"https://example.com/post\">Read article</a>"

	if got != expected {
		t.Errorf("buildAutoNewsMessage() = %q; want %q", got, expected)
	}
}
