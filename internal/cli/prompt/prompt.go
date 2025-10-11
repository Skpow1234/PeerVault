package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Skpow1234/Peervault/internal/cli/completion"
	"github.com/Skpow1234/Peervault/internal/cli/config"
	"github.com/Skpow1234/Peervault/internal/cli/history"
)

var ErrEOF = fmt.Errorf("EOF")

// Prompt represents an interactive prompt with advanced features
type Prompt struct {
	config        *config.Config
	history       *history.History
	reader        *bufio.Reader
	currentLine   string
	cursorPos     int
	historyIdx    int
	completions   []string
	completionIdx int
	inCompletion  bool
	multiLine     bool
	context       string
	completer     *completion.Completer
}

// New creates a new prompt instance
func New(cfg *config.Config, hist *history.History) *Prompt {
	return &Prompt{
		config:        cfg,
		history:       hist,
		reader:        bufio.NewReader(os.Stdin),
		historyIdx:    -1,
		completionIdx: -1,
		completer:     completion.New(),
	}
}

// ReadLine reads a line from the user with advanced features
func (p *Prompt) ReadLine() (string, error) {
	// Set terminal to raw mode for advanced input handling
	if err := p.setRawMode(); err != nil {
		// Fallback to simple input if raw mode fails
		return p.readLineSimple()
	}
	defer p.restoreMode()

	p.currentLine = ""
	p.cursorPos = 0
	p.historyIdx = -1
	p.completions = nil
	p.completionIdx = -1
	p.inCompletion = false
	p.multiLine = false

	for {
		char, err := p.readChar()
		if err != nil {
			return "", err
		}

		switch char {
		case '\r', '\n':
			if p.multiLine && !p.isCompleteCommand() {
				p.currentLine += "\n"
				p.cursorPos = len(p.currentLine)
				p.printPrompt()
				continue
			}
			return p.currentLine, nil

		case '\t':
			p.handleTabCompletion()
			continue

		case '\b', 127: // Backspace
			p.handleBackspace()
			continue

		case 3: // Ctrl+C
			return "", ErrEOF

		case 4: // Ctrl+D
			if p.currentLine == "" {
				return "", ErrEOF
			}

		case 27: // Escape sequence
			if err := p.handleEscapeSequence(); err != nil {
				continue
			}

		default:
			if char >= 32 { // Printable characters
				p.insertChar(char)
			}
		}

		p.refreshLine()
	}
}

// readLineSimple provides fallback simple input reading
func (p *Prompt) readLineSimple() (string, error) {
	p.printPrompt()
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", ErrEOF
	}
	return strings.TrimSpace(line), nil
}

// setRawMode sets the terminal to raw mode for advanced input
func (p *Prompt) setRawMode() error {
	// This is a simplified implementation
	// In a real implementation, you'd use termios or similar
	return nil
}

// restoreMode restores the terminal to normal mode
func (p *Prompt) restoreMode() {
	// This is a simplified implementation
	// In a real implementation, you'd restore termios settings
}

// readChar reads a single character from input
func (p *Prompt) readChar() (rune, error) {
	char, _, err := p.reader.ReadRune()
	return char, err
}

// handleTabCompletion handles tab completion
func (p *Prompt) handleTabCompletion() {
	if p.inCompletion {
		p.cycleCompletion()
		return
	}

	// Get completions based on current input
	p.completions = p.getCompletions()
	if len(p.completions) == 0 {
		return
	}

	if len(p.completions) == 1 {
		// Single completion - apply it
		p.applyCompletion(p.completions[0])
	} else {
		// Multiple completions - show them
		p.showCompletions()
		p.inCompletion = true
		p.completionIdx = 0
	}
}

// getCompletions returns possible completions for current input
func (p *Prompt) getCompletions() []string {
	return p.completer.GetCompletions(p.currentLine)
}

// applyCompletion applies a completion to the current line
func (p *Prompt) applyCompletion(completion string) {
	words := strings.Fields(p.currentLine)
	if len(words) == 0 {
		p.currentLine = completion + " "
		p.cursorPos = len(p.currentLine)
		return
	}

	// Replace the last word with the completion
	words[len(words)-1] = completion
	p.currentLine = strings.Join(words, " ") + " "
	p.cursorPos = len(p.currentLine)
}

// showCompletions displays available completions
func (p *Prompt) showCompletions() {
	fmt.Println()
	for i, comp := range p.completions {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Print(comp)
	}
	fmt.Println()
}

// cycleCompletion cycles through available completions
func (p *Prompt) cycleCompletion() {
	if len(p.completions) == 0 {
		return
	}

	p.completionIdx = (p.completionIdx + 1) % len(p.completions)
	p.applyCompletion(p.completions[p.completionIdx])
}

// handleBackspace handles backspace key
func (p *Prompt) handleBackspace() {
	if p.cursorPos > 0 {
		// Remove character before cursor
		p.currentLine = p.currentLine[:p.cursorPos-1] + p.currentLine[p.cursorPos:]
		p.cursorPos--
	}
}

// handleEscapeSequence handles escape sequences (arrow keys, etc.)
func (p *Prompt) handleEscapeSequence() error {
	// Read the next two characters to determine the sequence
	char1, err := p.readChar()
	if err != nil {
		return err
	}

	if char1 == '[' {
		char2, err := p.readChar()
		if err != nil {
			return err
		}

		switch char2 {
		case 'A': // Up arrow - history
			p.navigateHistory(-1)
		case 'B': // Down arrow - history
			p.navigateHistory(1)
		case 'C': // Right arrow
			p.moveCursor(1)
		case 'D': // Left arrow
			p.moveCursor(-1)
		case 'H': // Home
			p.cursorPos = 0
		case 'F': // End
			p.cursorPos = len(p.currentLine)
		}
	}
	return nil
}

// navigateHistory navigates through command history
func (p *Prompt) navigateHistory(direction int) {
	if p.history == nil {
		return
	}

	history := p.history.GetAll()
	if len(history) == 0 {
		return
	}

	p.historyIdx += direction
	if p.historyIdx < 0 {
		p.historyIdx = 0
	} else if p.historyIdx >= len(history) {
		p.historyIdx = len(history) - 1
	}

	if p.historyIdx >= 0 && p.historyIdx < len(history) {
		p.currentLine = history[p.historyIdx]
		p.cursorPos = len(p.currentLine)
	}
}

// moveCursor moves the cursor left or right
func (p *Prompt) moveCursor(direction int) {
	newPos := p.cursorPos + direction
	if newPos >= 0 && newPos <= len(p.currentLine) {
		p.cursorPos = newPos
	}
}

// insertChar inserts a character at the current cursor position
func (p *Prompt) insertChar(char rune) {
	before := p.currentLine[:p.cursorPos]
	after := p.currentLine[p.cursorPos:]
	p.currentLine = before + string(char) + after
	p.cursorPos++
}

// refreshLine refreshes the display of the current line
func (p *Prompt) refreshLine() {
	// Clear current line
	fmt.Print("\r\033[K")

	// Print prompt
	p.printPrompt()

	// Print current line
	fmt.Print(p.currentLine)

	// Position cursor
	if p.cursorPos < len(p.currentLine) {
		fmt.Printf("\033[%dD", len(p.currentLine)-p.cursorPos)
	}
}

// printPrompt prints the current prompt
func (p *Prompt) printPrompt() {
	if p.context != "" {
		fmt.Printf("peervault[%s]> ", p.context)
	} else {
		fmt.Print("peervault> ")
	}
}

// isCompleteCommand checks if the current line forms a complete command
func (p *Prompt) isCompleteCommand() bool {
	// Simple check - in a real implementation, you'd parse the command
	return !strings.HasSuffix(p.currentLine, "\\")
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

// SetContext sets the prompt context (e.g., current node)
func (p *Prompt) SetContext(context string) {
	p.context = context
}

// GetContext returns the current prompt context
func (p *Prompt) GetContext() string {
	return p.context
}

// SetMultiLine enables or disables multi-line mode
func (p *Prompt) SetMultiLine(enabled bool) {
	p.multiLine = enabled
}

// IsMultiLine returns whether multi-line mode is enabled
func (p *Prompt) IsMultiLine() bool {
	return p.multiLine
}
