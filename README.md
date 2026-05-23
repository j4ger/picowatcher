# picowatcher

A production-quality Go daemon that monitors RSS/Atom feeds, summarizes new items with an OpenAI-compatible LLM, and delivers notifications via HTTP webhooks.

## Features

- **Feed monitoring** – Polls any number of RSS/Atom feeds on a configurable interval
- **Change detection** – Persists per-feed seen-item state (by GUID → link → title) to a JSON file so only genuinely new items trigger notifications
- **LLM summarization** – Calls any OpenAI-compatible API (OpenAI, Azure OpenAI, Ollama, etc.) to produce a short summary for each new item
- **Webhook delivery** – Renders a configurable Go `text/template` payload and POSTs it to any HTTP endpoint
- **Rolling log files** – Uses [lumberjack](https://github.com/natefinch/lumberjack) for size- and age-based log rotation, with simultaneous output to stdout
- **Graceful shutdown** – Handles `SIGINT`/`SIGTERM`, saves state before exiting

## Requirements

- Go 1.22+

### Nix Development Shell

If you use [Nix](https://nixos.org/), you can quickly set up a complete development environment:

```bash
# Enter the development shell
nix develop

# Or use direnv for automatic activation
echo "use flake" > .envrc
direnv allow
```

The flake provides:
- Go toolchain
- gopls (language server)
- delve (debugger)
- Additional Go development tools

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

# Dry-run toggles (for config testing)
dry_run:
  fetch_only: false        # stop after fetching items (no summaries, no webhooks, no state writes)
  skip_llm: false          # render and log the prompt instead of calling the LLM
  skip_webhook: false      # render and log the webhook payload instead of sending it
  skip_state_save: false   # do not write the state file after a cycle

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

# Trigger dry run from CLI (overrides config dry_run fields)
# Full dry run: fetch only, skip LLM/webhook/state writes
./picowatcher -dry-run

# Selective dry-run toggles
./picowatcher -dry-fetch-only          # list new items only
./picowatcher -dry-skip-llm            # render prompt but skip API call
./picowatcher -dry-skip-webhook        # render payload but skip HTTP send
./picowatcher -dry-skip-state-save     # skip writing state.json
```

## Running with Docker Compose

1. Create a runtime config from the example:

```bash
cp config.example.yaml config.yaml
```

2. Start picowatcher:

```bash
docker compose up -d --build
```

3. Check logs:

```bash
docker compose logs -f picowatcher
```

4. Stop it:

```bash
docker compose down
```

### Use a custom config file with Docker Compose

The compose setup mounts `config.yaml` by default.  
To use a different config file, set `PICOWATCHER_CONFIG` when starting:

```bash
PICOWATCHER_CONFIG=./configs/prod.yaml docker compose up -d --build
```

State and log files are persisted in the `picowatcher-data` Docker volume (`/data` in the container).

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
