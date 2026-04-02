package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/j4ger/picowatcher/config"
	"github.com/j4ger/picowatcher/feed"
)

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// httpClient is shared across all Summarize calls to enable connection pooling.
var httpClient = &http.Client{}

// templateCache stores parsed prompt templates keyed by their source string.
var (
	tmplMu    sync.RWMutex
	tmplCache = make(map[string]*template.Template)
)

func Summarize(cfg config.LLMConfig, item feed.Item) (string, error) {
	if !cfg.Enabled {
		return "", nil
	}

	userPrompt, err := renderTemplate(cfg.UserPromptTmpl, item)
	if err != nil {
		return "", fmt.Errorf("rendering user prompt template: %w", err)
	}

	reqBody := chatRequest{
		Model: cfg.Model,
		Messages: []message{
			{Role: "system", Content: cfg.SystemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling LLM request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()

	url := cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("creating LLM request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling LLM API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("LLM API returned non-2xx status: %d", resp.StatusCode)
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("decoding LLM response: %w", err)
	}
	if chatResp.Error != nil {
		return "", fmt.Errorf("LLM API error: %s", chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}
	return chatResp.Choices[0].Message.Content, nil
}

func renderTemplate(tmpl string, item feed.Item) (string, error) {
	tmplMu.RLock()
	t, ok := tmplCache[tmpl]
	tmplMu.RUnlock()
	if !ok {
		tmplMu.Lock()
		// Re-check under write lock to avoid double-parse.
		t, ok = tmplCache[tmpl]
		if !ok {
			var err error
			t, err = template.New("prompt").Parse(tmpl)
			if err != nil {
				tmplMu.Unlock()
				return "", err
			}
			tmplCache[tmpl] = t
		}
		tmplMu.Unlock()
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, item); err != nil {
		return "", err
	}
	return buf.String(), nil
}
