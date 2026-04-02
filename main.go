package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/j4ger/picowatcher/config"
	"github.com/j4ger/picowatcher/feed"
	"github.com/j4ger/picowatcher/llm"
	"github.com/j4ger/picowatcher/logger"
	"github.com/j4ger/picowatcher/notify"
	"github.com/j4ger/picowatcher/state"
	"github.com/j4ger/picowatcher/watcher"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log := logger.Setup(cfg.Log)

	st, err := state.Load(cfg.State.Path)
	if err != nil {
		log.Error("failed to load state", "error", err)
		os.Exit(1)
	}

	// Initialize config watcher
	configWatcher, err := watcher.New(*configPath, log)
	if err != nil {
		log.Error("failed to initialize config watcher", "error", err)
		os.Exit(1)
	}
	defer configWatcher.Close()

	log.Info("picowatcher started",
		"interval_seconds", cfg.IntervalSeconds,
		"feeds", len(cfg.Feeds),
	)

	// Run once immediately, then on interval
	run(cfg, st, log)

	ticker := time.NewTicker(time.Duration(cfg.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			run(cfg, st, log)
		case <-configWatcher.ReloadChan():
			log.Info("reloading configuration")

			// Save current state before reloading
			if err := st.Save(); err != nil {
				log.Error("saving state before reload", "error", err)
			}

			// Load new config
			newCfg, err := config.Load(*configPath)
			if err != nil {
				log.Error("failed to reload config, keeping old config", "error", err)
				continue
			}

			// Reinitialize logger if log config changed
			log = logger.Setup(newCfg.Log)

			// Reload state if path changed
			if newCfg.State.Path != cfg.State.Path {
				newSt, err := state.Load(newCfg.State.Path)
				if err != nil {
					log.Error("failed to load new state file, keeping old state", "error", err)
				} else {
					st = newSt
				}
			}

			// Update config reference
			cfg = newCfg

			// Reset ticker if interval changed
			ticker.Stop()
			ticker = time.NewTicker(time.Duration(cfg.IntervalSeconds) * time.Second)

			log.Info("configuration reloaded successfully",
				"interval_seconds", cfg.IntervalSeconds,
				"feeds", len(cfg.Feeds),
			)

			// Run immediately with new config
			run(cfg, st, log)

		case sig := <-sigs:
			log.Info("shutting down", "signal", sig)
			if err := st.Save(); err != nil {
				log.Error("saving state on shutdown", "error", err)
			}
			return
		}
	}
}

func run(cfg *config.Config, st *state.State, log *slog.Logger) {
	log.Info("starting feed check cycle")
	for _, feedCfg := range cfg.Feeds {
		log.Info("checking feed", "name", feedCfg.Name, "url", feedCfg.URL)
		items, err := feed.FetchNew(feedCfg, st)
		if err != nil {
			log.Error("fetching feed", "name", feedCfg.Name, "error", err)
			continue
		}
		log.Info("found new items", "feed", feedCfg.Name, "count", len(items))
		for _, item := range items {
			// Mark seen early to avoid duplicate processing on partial failures
			st.MarkSeen(feedCfg.URL, item.ID)

			var summary string
			if cfg.LLM.Enabled {
				summary, err = llm.Summarize(cfg.LLM, item)
				if err != nil {
					log.Error("summarizing item", "title", item.Title, "error", err)
					summary = ""
				}
			}

			if cfg.Webhook.URL != "" {
				if err := notify.Send(cfg.Webhook, item, summary); err != nil {
					log.Error("sending webhook", "title", item.Title, "error", err)
				} else {
					log.Info("webhook sent", "feed", feedCfg.Name, "title", item.Title)
				}
			}
		}
	}

	if err := st.Save(); err != nil {
		log.Error("saving state", "error", err)
	}
	log.Info("feed check cycle complete")
}
