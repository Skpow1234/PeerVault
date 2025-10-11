package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Skpow1234/Peervault/internal/cli/config"
	"github.com/Skpow1234/Peervault/internal/cli/history"
)

var ErrEOF = fmt.Errorf("EOF")

// Prompt represents an interactive prompt
type Prompt struct {
	config  *config.Config
	history *history.History
	reader  *bufio.Reader
}

// New creates a new prompt instance
func New(cfg *config.Config, hist *history.History) *Prompt {
	return &Prompt{
		config:  cfg,
		history: hist,
		reader:  bufio.NewReader(os.Stdin),
	}
}

// ReadLine reads a line from the user with basic completion
func (p *Prompt) ReadLine() (string, error) {
	fmt.Print("peervault> ")

	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", ErrEOF
	}

	return strings.TrimSpace(line), nil
}

// AddToHistory adds a command to history
func (p *Prompt) AddToHistory(command string) {
	if p.history != nil {
		p.history.Add(command)
	}
}

// GetHistory returns command history
func (p *Prompt) GetHistory() []string {
	if p.history != nil {
		return p.history.GetAll()
	}
	return nil
}

// SetPrompt sets a custom prompt
func (p *Prompt) SetPrompt(prompt string) {
	// This would be used for context-aware prompts like:
	// peervault[node1]>
	// peervault[grpc]>
	// etc.
}
