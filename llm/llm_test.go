package llm

import (
	"testing"

	"github.com/j4ger/picowatcher/config"
	"github.com/j4ger/picowatcher/feed"
)

func TestRenderPrompt(t *testing.T) {
	cfg := config.LLMConfig{
		UserPromptTmpl: "Title: {{.Title}}\nContent: {{.Content}}",
	}
	item := feed.Item{
		Title:   "Example Title",
		Content: "Example Content",
	}

	prompt, err := RenderPrompt(cfg, item)
	if err != nil {
		t.Fatalf("RenderPrompt returned error: %v", err)
	}
	expected := "Title: Example Title\nContent: Example Content"
	if prompt != expected {
		t.Fatalf("unexpected prompt.\nexpected: %q\ngot:      %q", expected, prompt)
	}
}
