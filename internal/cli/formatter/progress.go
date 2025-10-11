package formatter

import (
	"fmt"
	"strings"
	"time"
)

// ProgressBar represents a progress bar for long operations
type ProgressBar struct {
	total      int
	current    int
	width      int
	message    string
	startTime  time.Time
	lastUpdate time.Time
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		total:      total,
		width:      50,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}
}

// Update updates the progress bar
func (pb *ProgressBar) Update(current int, message string) {
	pb.current = current
	pb.message = message
	pb.lastUpdate = time.Now()

	// Only update if enough time has passed (throttle updates)
	if time.Since(pb.lastUpdate) < 100*time.Millisecond {
		return
	}

	pb.render()
}

// render renders the progress bar
func (pb *ProgressBar) render() {
	percentage := float64(pb.current) / float64(pb.total)
	filled := int(percentage * float64(pb.width))
	empty := pb.width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	elapsed := time.Since(pb.startTime)

	// Calculate ETA
	var eta time.Duration
	if pb.current > 0 {
		rate := float64(pb.current) / elapsed.Seconds()
		remaining := float64(pb.total-pb.current) / rate
		eta = time.Duration(remaining) * time.Second
	}

	fmt.Printf("\r\033[K") // Clear line
	fmt.Printf("Progress: [%s] %.1f%% (%d/%d) %s ETA: %s",
		bar, percentage*100, pb.current, pb.total, pb.message, formatDuration(eta))
}

// Complete marks the progress bar as complete
func (pb *ProgressBar) Complete(message string) {
	pb.current = pb.total
	pb.message = message
	pb.render()
	fmt.Println() // New line
}

// Spinner represents a loading spinner
type Spinner struct {
	message  string
	frames   []string
	index    int
	running  bool
	stopChan chan bool
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message:  message,
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		index:    0,
		running:  false,
		stopChan: make(chan bool),
	}
}

// Start starts the spinner
func (s *Spinner) Start() {
	if s.running {
		return
	}

	s.running = true
	go s.animate()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	if !s.running {
		return
	}

	s.running = false
	s.stopChan <- true
	fmt.Printf("\r\033[K") // Clear line
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.message = message
}

// animate animates the spinner
func (s *Spinner) animate() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.render()
			s.index = (s.index + 1) % len(s.frames)
		case <-s.stopChan:
			return
		}
	}
}

// render renders the spinner
func (s *Spinner) render() {
	fmt.Printf("\r\033[K") // Clear line
	fmt.Printf("%s %s", s.frames[s.index], s.message)
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.0fh", d.Hours())
	}
}
