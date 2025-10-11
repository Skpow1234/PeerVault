package history

import (
	"bufio"
	"os"
	"strings"
)

// History manages command history
type History struct {
	commands []string
	file     string
	maxSize  int
}

// New creates a new history instance
func New(file string) *History {
	h := &History{
		commands: make([]string, 0),
		file:     file,
		maxSize:  1000,
	}

	// Load existing history
	h.Load()

	return h
}

// Add adds a command to history
func (h *History) Add(command string) {
	// Don't add empty commands or duplicates
	if command == "" || (len(h.commands) > 0 && h.commands[len(h.commands)-1] == command) {
		return
	}

	h.commands = append(h.commands, command)

	// Trim history if it exceeds max size
	if len(h.commands) > h.maxSize {
		h.commands = h.commands[1:]
	}

	// Save to file
	h.Save()
}

// GetAll returns all commands in history
func (h *History) GetAll() []string {
	return append([]string(nil), h.commands...)
}

// GetLast returns the last N commands
func (h *History) GetLast(n int) []string {
	if n <= 0 || len(h.commands) == 0 {
		return nil
	}

	start := len(h.commands) - n
	if start < 0 {
		start = 0
	}

	return append([]string(nil), h.commands[start:]...)
}

// Search searches for commands containing the given text
func (h *History) Search(text string) []string {
	var results []string
	text = strings.ToLower(text)

	for _, cmd := range h.commands {
		if strings.Contains(strings.ToLower(cmd), text) {
			results = append(results, cmd)
		}
	}

	return results
}

// Clear clears the history
func (h *History) Clear() {
	h.commands = h.commands[:0]
	h.Save()
}

// Load loads history from file
func (h *History) Load() error {
	if h.file == "" {
		return nil
	}

	file, err := os.Open(h.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's okay
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			h.commands = append(h.commands, line)
		}
	}

	return scanner.Err()
}

// Save saves history to file
func (h *History) Save() error {
	if h.file == "" {
		return nil
	}

	// Create directory if it doesn't exist
	dir := strings.TrimSuffix(h.file, "/"+strings.Split(h.file, "/")[len(strings.Split(h.file, "/"))-1])
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	file, err := os.Create(h.file)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, cmd := range h.commands {
		if _, err := writer.WriteString(cmd + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// GetCompletions returns possible completions based on history
func (h *History) GetCompletions(input string) []string {
	var completions []string
	input = strings.ToLower(input)

	// Get unique commands that start with the input
	seen := make(map[string]bool)
	for _, cmd := range h.commands {
		parts := strings.Fields(cmd)
		if len(parts) > 0 {
			firstWord := strings.ToLower(parts[0])
			if strings.HasPrefix(firstWord, input) && !seen[firstWord] {
				completions = append(completions, firstWord)
				seen[firstWord] = true
			}
		}
	}

	return completions
}
