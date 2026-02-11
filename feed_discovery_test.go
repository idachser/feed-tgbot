package main

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestMakeAbsoluteURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		relativeURL string
		expected    string
		shouldError bool
	}{
		{
			name:        "absolute url stays absolute",
			baseURL:     "https://example.com",
			relativeURL: "https://other.com/feed.rss",
			expected:    "https://other.com/feed.rss",
			shouldError: false,
		},
		{
			name:        "relative path",
			baseURL:     "https://example.com",
			relativeURL: "/feed.rss",
			expected:    "https://example.com/feed.rss",
			shouldError: false,
		},
		{
			name:        "relative path with subdirectory",
			baseURL:     "https://example.com/blog/",
			relativeURL: "feed.rss",
			expected:    "https://example.com/blog/feed.rss",
			shouldError: false,
		},
		{
			name:        "parent directory",
			baseURL:     "https://example.com/blog/post",
			relativeURL: "../feed.rss",
			expected:    "https://example.com/feed.rss",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := makeAbsoluteURL(tt.baseURL, tt.relativeURL)

			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("makeAbsoluteURL(%q, %q) = %q; want %q",
					tt.baseURL, tt.relativeURL, result, tt.expected)
			}
		})
	}
}

func TestParseLinkTag(t *testing.T) {
	tests := []struct {
		name        string
		htmlTag     string
		baseURL     string
		expectNil   bool
		expectedURL string
	}{
		{
			name:        "valid RSS feed",
			htmlTag:     `<link rel="alternate" type="application/rss+xml" href="/feed.rss" title="RSS Feed">`,
			baseURL:     "https://example.com",
			expectNil:   false,
			expectedURL: "https://example.com/feed.rss",
		},
		{
			name:        "valid Atom feed",
			htmlTag:     `<link rel="alternate" type="application/atom+xml" href="/atom.xml">`,
			baseURL:     "https://example.com",
			expectNil:   false,
			expectedURL: "https://example.com/atom.xml",
		},
		{
			name:      "not alternate link",
			htmlTag:   `<link rel="stylesheet" type="text/css" href="/style.css">`,
			baseURL:   "https://example.com",
			expectNil: true,
		},
		{
			name:      "wrong type",
			htmlTag:   `<link rel="alternate" type="text/html" href="/page.html">`,
			baseURL:   "https://example.com",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.htmlTag))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			var linkNode *html.Node
			var findLink func(*html.Node)
			findLink = func(n *html.Node) {
				if n.Type == html.ElementNode && n.Data == "link" {
					linkNode = n
					return
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					findLink(c)
				}
			}
			findLink(doc)

			if linkNode == nil {
				t.Fatal("link tag not found in HTML")
			}

			result := parseLinkTag(linkNode, tt.baseURL)

			if tt.expectNil && result != nil {
				t.Errorf("expected nil but got: %+v", result)
				return
			}

			if !tt.expectNil && result == nil {
				t.Error("expected feed but got nil")
				return
			}

			if !tt.expectNil && result.URL != tt.expectedURL {
				t.Errorf("parseLinkTag URL = %q; want %q", result.URL, tt.expectedURL)
			}
		})
	}
}
