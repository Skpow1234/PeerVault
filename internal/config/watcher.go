package config

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// ConfigWatcher watches for configuration file changes and triggers reload callbacks
type ConfigWatcher struct {
	filePath string
	callback func()
	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex
	running  bool
}

// NewConfigWatcher creates a new configuration file watcher
func NewConfigWatcher(filePath string, callback func()) *ConfigWatcher {
	return &ConfigWatcher{
		filePath: filePath,
		callback: callback,
		stopChan: make(chan struct{}),
	}
}

// Start starts watching the configuration file for changes
func (w *ConfigWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return fmt.Errorf("watcher is already running")
	}

	// Check if file exists
	if _, err := os.Stat(w.filePath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", w.filePath)
	}

	w.running = true
	w.wg.Add(1)

	go w.watch()

	return nil
}

// Stop stops watching the configuration file
func (w *ConfigWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	close(w.stopChan)
	w.wg.Wait()
	w.running = false
}

// watch monitors the configuration file for changes
func (w *ConfigWatcher) watch() {
	defer w.wg.Done()

	var lastModTime time.Time
	var lastSize int64

	// Get initial file info
	if info, err := os.Stat(w.filePath); err == nil {
		lastModTime = info.ModTime()
		lastSize = info.Size()
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			// Check if file has changed
			if info, err := os.Stat(w.filePath); err == nil {
				if info.ModTime().After(lastModTime) || info.Size() != lastSize {
					lastModTime = info.ModTime()
					lastSize = info.Size()

					// Wait a bit to ensure file is fully written
					time.Sleep(100 * time.Millisecond)

					// Trigger callback
					if w.callback != nil {
						w.callback()
					}
				}
			}
		}
	}
}

// IsRunning returns true if the watcher is currently running
func (w *ConfigWatcher) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

// GetFilePath returns the path of the file being watched
func (w *ConfigWatcher) GetFilePath() string {
	return w.filePath
}
