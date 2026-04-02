package watcher

import (
	"log/slog"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// ConfigWatcher watches a config file for changes
type ConfigWatcher struct {
	watcher  *fsnotify.Watcher
	filePath string
	reloadCh chan struct{}
	log      *slog.Logger
}

// New creates a new ConfigWatcher for the specified file path
func New(filePath string, log *slog.Logger) (*ConfigWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Get absolute path and watch the directory (more reliable than watching the file directly)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		w.Close()
		return nil, err
	}
	dir := filepath.Dir(absPath)

	if err := w.Add(dir); err != nil {
		w.Close()
		return nil, err
	}

	cw := &ConfigWatcher{
		watcher:  w,
		filePath: absPath,
		reloadCh: make(chan struct{}, 1),
		log:      log,
	}

	go cw.watch()

	return cw, nil
}

// ReloadChan returns a channel that signals when the config should be reloaded
func (cw *ConfigWatcher) ReloadChan() <-chan struct{} {
	return cw.reloadCh
}

// Close stops the watcher
func (cw *ConfigWatcher) Close() error {
	return cw.watcher.Close()
}

func (cw *ConfigWatcher) watch() {
	for {
		select {
		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}

			// Check if the event is for our config file
			eventPath, err := filepath.Abs(event.Name)
			if err != nil {
				continue
			}

			if eventPath != cw.filePath {
				continue
			}

			// Handle Write and Create events (covers most editors)
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				cw.log.Info("config file changed, triggering reload", "file", cw.filePath)
				// Non-blocking send to reload channel
				select {
				case cw.reloadCh <- struct{}{}:
				default:
					// Already pending reload
				}
			}

		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}
			cw.log.Error("config watcher error", "error", err)
		}
	}
}
