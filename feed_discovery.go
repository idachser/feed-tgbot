package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type DiscoveredFeed struct {
	Title string
	URL   string
	Type  string // rss or atom
}

func DiscoverFeeds(pageURL string) ([]DiscoveredFeed, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	res, err := client.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	doc, err := html.Parse(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	feeds := extractFeeds(doc, pageURL)

	return feeds, nil
}

func extractFeeds(n *html.Node, baseURL string) []DiscoveredFeed {
	var feeds []DiscoveredFeed

	if n.Type == html.ElementNode && n.Data == "link" {
		feed := parseLinkTag(n, baseURL)
		if feed != nil {
			feeds = append(feeds, *feed)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		feeds = append(feeds, extractFeeds(c, baseURL)...)
	}

	return feeds
}

func parseLinkTag(n *html.Node, baseURL string) *DiscoveredFeed {
	var rel, feedType, href, title string

	for _, attr := range n.Attr {
		switch attr.Key {
		case "rel":
			rel = attr.Val
		case "type":
			feedType = attr.Val
		case "href":
			href = attr.Val
		case "title":
			title = attr.Val
		}
	}

	if rel != "alternate" {
		return nil
	}

	feedType = strings.ToLower(feedType)
	if !strings.Contains(feedType, "rss") && !strings.Contains(feedType, "atom") {
		return nil
	}

	var typeLabel string
	if strings.Contains(feedType, "atom") {
		typeLabel = "Atom"
	} else {
		typeLabel = "RSS"
	}

	absoluteURL, err := makeAbsoluteURL(baseURL, href)
	if err != nil {
		return nil
	}

	if title == "" {
		title = typeLabel + " Feed"
	}

	return &DiscoveredFeed{
		Title: title,
		URL:   absoluteURL,
		Type:  typeLabel,
	}
}

func makeAbsoluteURL(baseURL, relativeURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	rel, err := url.Parse(relativeURL)
	if err != nil {
		return "", err
	}

	if rel.IsAbs() {
		return relativeURL, nil
	}

	absolute := base.ResolveReference(rel)
	return absolute.String(), nil
}
