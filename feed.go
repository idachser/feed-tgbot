package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

type FeedItem struct {
	Title       string
	Description string
	Link        string
	Published   *time.Time
}

const feedRequestTimeout = 10 * time.Second

func getFeeds(url string) ([]FeedItem, error) {
	return getFeedsWithContext(context.Background(), url)
}

func getFeedsWithContext(parentCtx context.Context, url string) ([]FeedItem, error) {
	ctx, cancel := context.WithTimeout(parentCtx, feedRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create feed request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
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

func getNewItems(feedURL, lastItemLink string) ([]FeedItem, error) {
	return getNewItemsWithContext(context.Background(), feedURL, lastItemLink)
}

func getNewItemsWithContext(ctx context.Context, feedURL, lastItemLink string) ([]FeedItem, error) {
	allItems, err := getFeedsWithContext(ctx, feedURL)
	if err != nil {
		return nil, err
	}

	if lastItemLink == "" {
		return allItems, nil
	}

	var newItems []FeedItem
	for _, item := range allItems {
		if item.Link == lastItemLink {
			break
		}
		newItems = append(newItems, item)
	}

	return newItems, nil
}
