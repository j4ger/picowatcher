package notify

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/j4ger/picowatcher/config"
	"github.com/j4ger/picowatcher/feed"
)

type TemplateData struct {
	FeedName    string
	FeedURL     string
	Title       string
	Link        string
	Description string
	Content     string
	Summary     string
	Published   time.Time
}

// httpClient is shared across all Send calls to enable connection pooling.
var httpClient = &http.Client{}

// templateCache stores parsed payload templates keyed by their source string.
var (
	tmplMu    sync.RWMutex
	tmplCache = make(map[string]*template.Template)
)

func Send(cfg config.WebhookConfig, item feed.Item, summary string) error {
	payload, err := RenderPayload(cfg, item, summary)
	if err != nil {
		return err
	}

	return SendPayload(cfg, payload)
}

func RenderPayload(cfg config.WebhookConfig, item feed.Item, summary string) (string, error) {
	data := TemplateData{
		FeedName:    item.FeedName,
		FeedURL:     item.FeedURL,
		Title:       item.Title,
		Link:        item.Link,
		Description: item.Description,
		Content:     item.Content,
		Summary:     summary,
		Published:   item.Published,
	}

	payload, err := renderTemplate(cfg.PayloadTemplate, data)
	if err != nil {
		return "", fmt.Errorf("rendering webhook payload template: %w", err)
	}
	return payload, nil
}

func SendPayload(cfg config.WebhookConfig, payload string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.URL, bytes.NewBufferString(payload))
	if err != nil {
		return fmt.Errorf("creating webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	client := httpClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sending webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}
	return nil
}

func renderTemplate(tmpl string, data TemplateData) (string, error) {
	tmplMu.RLock()
	t, ok := tmplCache[tmpl]
	tmplMu.RUnlock()
	if !ok {
		tmplMu.Lock()
		// Re-check under write lock to avoid double-parse.
		t, ok = tmplCache[tmpl]
		if !ok {
			var err error
			t, err = template.New("webhook").Parse(tmpl)
			if err != nil {
				tmplMu.Unlock()
				return "", err
			}
			tmplCache[tmpl] = t
		}
		tmplMu.Unlock()
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
