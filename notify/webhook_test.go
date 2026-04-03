package notify

import (
	"testing"
	"time"

	"github.com/j4ger/picowatcher/config"
	"github.com/j4ger/picowatcher/feed"
)

func TestRenderPayload(t *testing.T) {
	cfg := config.WebhookConfig{
		PayloadTemplate: `{"title":"{{.Title}}","summary":"{{.Summary}}","feed":"{{.FeedName}}"}`,
	}
	item := feed.Item{
		FeedName:  "Feed",
		Title:     "Item Title",
		Content:   "Body",
		Published: time.Date(2024, time.February, 3, 0, 0, 0, 0, time.UTC),
	}

	payload, err := RenderPayload(cfg, item, "Summary text")
	if err != nil {
		t.Fatalf("RenderPayload returned error: %v", err)
	}
	expected := `{"title":"Item Title","summary":"Summary text","feed":"Feed"}`
	if payload != expected {
		t.Fatalf("unexpected payload.\nexpected: %q\ngot:      %q", expected, payload)
	}
}
