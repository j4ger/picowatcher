package watcher

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigWatcher(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test config file
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	if err := os.WriteFile(configPath, []byte("test: value1\n"), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Create logger
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Create watcher
	cw, err := New(configPath, log)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer cw.Close()

	// Modify the config file
	time.Sleep(100 * time.Millisecond) // Give watcher time to start
	if err := os.WriteFile(configPath, []byte("test: value2\n"), 0644); err != nil {
		t.Fatalf("failed to modify config: %v", err)
	}

	// Wait for reload signal
	select {
	case <-cw.ReloadChan():
		t.Log("Successfully detected config file change")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for config reload signal")
	}
}

func TestConfigWatcherMultipleChanges(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test config file
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	if err := os.WriteFile(configPath, []byte("test: value1\n"), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Create logger
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Create watcher
	cw, err := New(configPath, log)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer cw.Close()

	// Test multiple changes
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 3; i++ {
		content := []byte("test: value" + string(rune('2'+i)) + "\n")
		if err := os.WriteFile(configPath, content, 0644); err != nil {
			t.Fatalf("failed to modify config (iteration %d): %v", i, err)
		}

		select {
		case <-cw.ReloadChan():
			t.Logf("Successfully detected config file change %d", i+1)
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting for config reload signal (iteration %d)", i)
		}

		time.Sleep(100 * time.Millisecond) // Small delay between changes
	}
}
