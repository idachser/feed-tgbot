package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
)

type FeedItem struct {
	Title       string
	Description string
	Link        string
	Published   *time.Time
}

func getFeeds(url string) ([]FeedItem, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	if len(feed.Items) == 0 {
		return nil, errors.New("feed is empty")
	}

	news := make([]FeedItem, 0, len(feed.Items))

	for _, item := range feed.Items {
		f := FeedItem{
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Published:   item.PublishedParsed,
		}
		news = append(news, f)
	}
	return news, nil
}

func scan() {
	feedURLs := []string{
		"https://www.golem.de/sonstiges/rss.html",
		"https://www.wissenschaft.de/feed-wissenschaft",
	}

	gol, _ := getFeeds(feedURLs[0])
	wiss, _ := getFeeds(feedURLs[1])

	fmt.Printf("========== Golem.de\n")
	for _, news := range gol {
		fmt.Printf("Title: %s\n", news.Title)
		fmt.Printf("Description: %s\n", news.Description)
		fmt.Printf("Link: %s\n", news.Link)
		fmt.Printf("Published: %s\n", news.Published)
	}

	fmt.Printf("========== Wissenschaft.de\n")
	for _, news := range wiss {
		fmt.Printf("Title: %s\n", news.Title)
		fmt.Printf("Description: %s\n", news.Description)
		fmt.Printf("Link: %s\n", news.Link)
		fmt.Printf("Published: %s\n", news.Published)
	}
}
