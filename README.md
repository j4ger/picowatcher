# picowatcher

A production-quality Go daemon that monitors RSS/Atom feeds, summarizes new items with an OpenAI-compatible LLM, and delivers notifications via HTTP webhooks.

## Features

- **Feed monitoring** – Polls any number of RSS/Atom feeds on a configurable interval
- **Change detection** – Persists per-feed seen-item state (by GUID → link → title) to a JSON file so only genuinely new items trigger notifications
- **LLM summarization** – Calls any OpenAI-compatible API (OpenAI, Azure OpenAI, Ollama, etc.) to produce a short summary for each new item
- **Webhook delivery** – Renders a configurable Go `text/template` payload and POSTs it to any HTTP endpoint
- **Rolling log files** – Uses [lumberjack](https://github.com/natefinish/lumberjack) for size- and age-based log rotation, with simultaneous output to stdout
- **Graceful shutdown** – Handles `SIGINT`/`SIGTERM`, saves state before exiting

## Requirements

- Go 1.22+

## Installation

```bash
git clone https://github.com/j4ger/picowatcher.git
cd picowatcher
go build -o picowatcher .
```

## Configuration

Copy `config.example.yaml` to `config.yaml` and edit:

```bash
cp config.example.yaml config.yaml
```

### Full Configuration Reference

```yaml
# Poll interval in seconds (default: 3600)
interval_seconds: 3600

# Feeds to monitor
feeds:
  - name: "Hacker News"
    url: "https://news.ycombinator.com/rss"
  - name: "Go Blog"
    url: "https://go.dev/blog/feed.atom"

# LLM (OpenAI-compatible)
llm:
  enabled: true                          # set false to skip summarization
  base_url: "https://api.openai.com/v1"  # change for Ollama / Azure / etc.
  api_key: "sk-..."
  model: "gpt-4o-mini"
  timeout_seconds: 30
  system_prompt: "Summarize in 2-3 sentences."
  user_prompt_template: "Title: {{.Title}}\n\nContent: {{.Content}}"

# Webhook delivery
webhook:
  url: "https://hooks.example.com/notify"
  method: "POST"          # default: POST
  timeout_seconds: 10
  headers:
    Authorization: "Bearer token"
  # Available template variables:
  #   .FeedName  .FeedURL  .Title  .Link
  #   .Description  .Content  .Summary  .Published
  payload_template: |
    {
      "feed":    "{{.FeedName}}",
      "title":   "{{.Title}}",
      "link":    "{{.Link}}",
      "summary": "{{.Summary}}"
    }

# State file path (default: state.json)
state:
  path: "state.json"

# Log rotation (lumberjack)
log:
  path: "picowatcher.log"
  max_size_mb: 100
  max_backups: 3
  max_age_days: 30
  compress: true
```

## Running

```bash
# Use default config.yaml in the current directory
./picowatcher

# Specify a custom config path
./picowatcher -config /etc/picowatcher/config.yaml
```

### systemd unit (example)

```ini
[Unit]
Description=picowatcher RSS feed watcher
After=network-online.target

[Service]
ExecStart=/usr/local/bin/picowatcher -config /etc/picowatcher/config.yaml
Restart=on-failure
WorkingDirectory=/var/lib/picowatcher

[Install]
WantedBy=multi-user.target
```

## Project Structure

```
picowatcher/
├── main.go               # Entry point, ticker loop, graceful shutdown
├── config/
│   └── config.go         # YAML config loading and defaults
├── feed/
│   └── feed.go           # RSS/Atom fetching and new-item detection
├── llm/
│   └── llm.go            # OpenAI-compatible chat completion client
├── notify/
│   └── webhook.go        # HTTP webhook delivery with template rendering
├── state/
│   └── state.go          # JSON-backed seen-item persistence
├── logger/
│   └── logger.go         # slog + lumberjack rolling log setup
└── config.example.yaml   # Annotated sample configuration
```

## License

MIT