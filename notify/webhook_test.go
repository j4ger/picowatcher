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

	func TestRenderPayloadTruncate(t *testing.T) {
		cfg := config.WebhookConfig{
			PayloadTemplate: `{"summary":"{{truncate .Summary 5}}","title":"{{truncate .Title 4}}"}`,
		}
		item := feed.Item{
			Title: "Picowatcher",
		}

		payload, err := RenderPayload(cfg, item, "こんにちは世界")
		if err != nil {
			t.Fatalf("RenderPayload returned error: %v", err)
		}
		expected := `{"summary":"こんにちは","title":"Pico"}`
		if payload != expected {
			t.Fatalf("unexpected payload.\nexpected: %q\ngot:      %q", expected, payload)
		}
	}

	func TestRenderPayloadTruncateNonPositiveLimit(t *testing.T) {
		cfg := config.WebhookConfig{
			PayloadTemplate: `{"summary":"{{truncate .Summary 0}}"}`,
		}

		payload, err := RenderPayload(cfg, feed.Item{}, "Summary text")
		if err != nil {
			t.Fatalf("RenderPayload returned error: %v", err)
		}
		expected := `{"summary":""}`
		if payload != expected {
			t.Fatalf("unexpected payload.\nexpected: %q\ngot:      %q", expected, payload)
		}
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
