package feed

import (
	"context"
	"fmt"
	"time"

	"github.com/j4ger/picowatcher/config"
	"github.com/j4ger/picowatcher/state"
	"github.com/mmcdole/gofeed"
)

type Item struct {
	FeedName    string
	FeedURL     string
	ID          string
	Title       string
	Link        string
	Description string
	Content     string
	Published   time.Time
}

func itemID(item *gofeed.Item) string {
	if item.GUID != "" {
		return item.GUID
	}
	if item.Link != "" {
		return item.Link
	}
	return item.Title
}

func FetchNew(feedCfg config.FeedConfig, st *state.State) ([]Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(feedCfg.URL, ctx)
	if err != nil {
		return nil, fmt.Errorf("parsing feed %s: %w", feedCfg.URL, err)
	}

	var newItems []Item
	for _, fi := range feed.Items {
		id := itemID(fi)
		if !st.IsNew(feedCfg.URL, id) {
			continue
		}
		item := Item{
			FeedName:    feedCfg.Name,
			FeedURL:     feedCfg.URL,
			ID:          id,
			Title:       fi.Title,
			Link:        fi.Link,
			Description: fi.Description,
			Content:     fi.Content,
		}
		if fi.PublishedParsed != nil {
			item.Published = *fi.PublishedParsed
		}
		if item.Content == "" {
			item.Content = fi.Description
		}
		newItems = append(newItems, item)
	}
	return newItems, nil
}
