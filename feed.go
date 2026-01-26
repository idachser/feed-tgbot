package main

import (
	"errors"
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
