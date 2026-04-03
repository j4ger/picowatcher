package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	IntervalSeconds int           `yaml:"interval_seconds"`
	Feeds           []FeedConfig  `yaml:"feeds"`
	LLM             LLMConfig     `yaml:"llm"`
	Webhook         WebhookConfig `yaml:"webhook"`
	DryRun          DryRunConfig  `yaml:"dry_run"`
	State           StateConfig   `yaml:"state"`
	Log             LogConfig     `yaml:"log"`
}

type FeedConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type DryRunConfig struct {
	// FetchOnly stops after downloading and parsing feeds; no state updates,
	// summarization, or webhooks are executed.
	FetchOnly bool `yaml:"fetch_only"`
	// SkipLLM renders the prompt for validation but does not call the LLM API.
	SkipLLM bool `yaml:"skip_llm"`
	// SkipWebhook renders the payload for validation but does not perform the HTTP request.
	SkipWebhook bool `yaml:"skip_webhook"`
	// SkipStateSave prevents writing the state file to disk after a cycle.
	SkipStateSave bool `yaml:"skip_state_save"`
}

type LLMConfig struct {
	BaseURL        string `yaml:"base_url"`
	APIKey         string `yaml:"api_key"`
	Model          string `yaml:"model"`
	SystemPrompt   string `yaml:"system_prompt"`
	UserPromptTmpl string `yaml:"user_prompt_template"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
	Enabled        bool   `yaml:"enabled"`
}

type WebhookConfig struct {
	URL             string            `yaml:"url"`
	Method          string            `yaml:"method"`
	PayloadTemplate string            `yaml:"payload_template"`
	Headers         map[string]string `yaml:"headers"`
	TimeoutSeconds  int               `yaml:"timeout_seconds"`
}

type StateConfig struct {
	Path string `yaml:"path"`
}

type LogConfig struct {
	Path       string `yaml:"path"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
	Compress   bool   `yaml:"compress"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	cfg.applyDefaults()
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.IntervalSeconds <= 0 {
		c.IntervalSeconds = 3600
	}
	if c.LLM.TimeoutSeconds <= 0 {
		c.LLM.TimeoutSeconds = 30
	}
	if c.LLM.Model == "" {
		c.LLM.Model = "gpt-4o-mini"
	}
	if c.LLM.SystemPrompt == "" {
		c.LLM.SystemPrompt = "You are a concise summarizer. Summarize the provided article in 2-3 sentences."
	}
	if c.LLM.UserPromptTmpl == "" {
		c.LLM.UserPromptTmpl = "Title: {{.Title}}\n\nContent: {{.Content}}"
	}
	if c.Webhook.Method == "" {
		c.Webhook.Method = "POST"
	}
	if c.Webhook.TimeoutSeconds <= 0 {
		c.Webhook.TimeoutSeconds = 10
	}
	if c.Webhook.PayloadTemplate == "" {
		c.Webhook.PayloadTemplate = `{"feed":"{{.FeedName}}","title":"{{.Title}}","link":"{{.Link}}","summary":"{{.Summary}}"}`
	}
	if c.State.Path == "" {
		c.State.Path = "state.json"
	}
	if c.Log.Path == "" {
		c.Log.Path = "picowatcher.log"
	}
	if c.Log.MaxSizeMB <= 0 {
		c.Log.MaxSizeMB = 100
	}
	if c.Log.MaxBackups <= 0 {
		c.Log.MaxBackups = 3
	}
	if c.Log.MaxAgeDays <= 0 {
		c.Log.MaxAgeDays = 30
	}
}
