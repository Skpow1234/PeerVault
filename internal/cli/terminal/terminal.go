package terminal

import (
	"fmt"
	"os"
	"runtime"
)

// Terminal represents terminal utilities
type Terminal struct {
}

// New creates a new terminal instance
func New() *Terminal {
	return &Terminal{}
}

// GetSize returns the terminal size (simplified cross-platform implementation)
func (t *Terminal) GetSize() (width, height int, err error) {
	// Default terminal size - in a real implementation, you'd use platform-specific code
	return 80, 24, nil
}

// ClearScreen clears the terminal screen
func (t *Terminal) ClearScreen() {
	if runtime.GOOS == "windows" {
		fmt.Print("\033[2J\033[H")
	} else {
		fmt.Print("\033[2J\033[H")
	}
}

// MoveCursor moves the cursor to a specific position
func (t *Terminal) MoveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

// HideCursor hides the cursor
func (t *Terminal) HideCursor() {
	fmt.Print("\033[?25l")
}

// ShowCursor shows the cursor
func (t *Terminal) ShowCursor() {
	fmt.Print("\033[?25h")
}

// SaveCursor saves the current cursor position
func (t *Terminal) SaveCursor() {
	fmt.Print("\033[s")
}

// RestoreCursor restores the saved cursor position
func (t *Terminal) RestoreCursor() {
	fmt.Print("\033[u")
}

// ClearLine clears the current line
func (t *Terminal) ClearLine() {
	fmt.Print("\033[2K")
}

// ClearLineFromCursor clears from cursor to end of line
func (t *Terminal) ClearLineFromCursor() {
	fmt.Print("\033[K")
}

// ClearLineToCursor clears from beginning of line to cursor
func (t *Terminal) ClearLineToCursor() {
	fmt.Print("\033[1K")
}

// ScrollUp scrolls the terminal up
func (t *Terminal) ScrollUp(lines int) {
	fmt.Printf("\033[%dS", lines)
}

// ScrollDown scrolls the terminal down
func (t *Terminal) ScrollDown(lines int) {
	fmt.Printf("\033[%dT", lines)
}

// SetTitle sets the terminal title
func (t *Terminal) SetTitle(title string) {
	fmt.Printf("\033]0;%s\007", title)
}

// SetForegroundColor sets the foreground color
func (t *Terminal) SetForegroundColor(color Color) {
	fmt.Printf("\033[%dm", color)
}

// SetBackgroundColor sets the background color
func (t *Terminal) SetBackgroundColor(color Color) {
	fmt.Printf("\033[%dm", color+10)
}

// ResetColors resets all colors
func (t *Terminal) ResetColors() {
	fmt.Print("\033[0m")
}

// SetBold sets bold text
func (t *Terminal) SetBold() {
	fmt.Print("\033[1m")
}

// SetDim sets dim text
func (t *Terminal) SetDim() {
	fmt.Print("\033[2m")
}

// SetItalic sets italic text
func (t *Terminal) SetItalic() {
	fmt.Print("\033[3m")
}

// SetUnderline sets underlined text
func (t *Terminal) SetUnderline() {
	fmt.Print("\033[4m")
}

// SetBlink sets blinking text
func (t *Terminal) SetBlink() {
	fmt.Print("\033[5m")
}

// SetReverse sets reverse video
func (t *Terminal) SetReverse() {
	fmt.Print("\033[7m")
}

// SetStrikethrough sets strikethrough text
func (t *Terminal) SetStrikethrough() {
	fmt.Print("\033[9m")
}

// ResetFormatting resets all formatting
func (t *Terminal) ResetFormatting() {
	fmt.Print("\033[0m")
}

// Color represents terminal colors
type Color int

const (
	ColorBlack Color = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
	ColorDefault
)

// IsTerminal checks if stdout is a terminal (simplified implementation)
func IsTerminal() bool {
	// Simplified check - in a real implementation, you'd use platform-specific code
	return true
}

// GetTerminalType returns the terminal type
func GetTerminalType() string {
	term := os.Getenv("TERM")
	if term == "" {
		return "unknown"
	}
	return term
}

// GetTerminalProgram returns the terminal program name
func GetTerminalProgram() string {
	if runtime.GOOS == "windows" {
		return "cmd" // Simplified for Windows
	}

	// Try to get terminal program from environment
	term := os.Getenv("TERM_PROGRAM")
	if term != "" {
		return term
	}

	// Try to get from parent process
	term = os.Getenv("TERM_PROGRAM_VERSION")
	if term != "" {
		return "unknown"
	}

	return "unknown"
}

// SupportsColors checks if the terminal supports colors
func SupportsColors() bool {
	if !IsTerminal() {
		return false
	}

	// Check environment variables
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check terminal type
	term := GetTerminalType()
	switch term {
	case "xterm", "xterm-256color", "screen", "screen-256color":
		return true
	case "vt100", "vt220":
		return false
	default:
		return true // Assume modern terminals support colors
	}
}

// GetColorDepth returns the color depth supported by the terminal
func GetColorDepth() int {
	if !SupportsColors() {
		return 1
	}

	term := GetTerminalType()
	switch term {
	case "xterm-256color", "screen-256color":
		return 256
	case "xterm", "screen":
		return 16
	default:
		return 8
	}
}
