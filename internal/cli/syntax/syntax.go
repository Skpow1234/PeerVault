package syntax

import (
	"strings"
)

// Highlighter provides syntax highlighting for CLI output
type Highlighter struct {
	colors bool
}

// New creates a new syntax highlighter
func New(colors bool) *Highlighter {
	return &Highlighter{
		colors: colors,
	}
}

// HighlightCommand highlights a command with colors
func (h *Highlighter) HighlightCommand(input string) string {
	if !h.colors {
		return input
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return input
	}

	// Highlight command name
	parts[0] = h.highlightCommand(parts[0])

	// Highlight options and arguments
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if strings.HasPrefix(part, "--") {
			parts[i] = h.highlightOption(part)
		} else if strings.HasPrefix(part, "-") {
			parts[i] = h.highlightShortOption(part)
		} else if strings.Contains(part, ".") {
			parts[i] = h.highlightFile(part)
		} else if strings.Contains(part, ":") {
			parts[i] = h.highlightAddress(part)
		} else {
			parts[i] = h.highlightArgument(part)
		}
	}

	return strings.Join(parts, " ")
}

// highlightCommand highlights command names
func (h *Highlighter) highlightCommand(cmd string) string {
	return "\033[1;32m" + cmd + "\033[0m" // Bold green
}

// highlightOption highlights long options
func (h *Highlighter) highlightOption(opt string) string {
	return "\033[1;36m" + opt + "\033[0m" // Bold cyan
}

// highlightShortOption highlights short options
func (h *Highlighter) highlightShortOption(opt string) string {
	return "\033[1;33m" + opt + "\033[0m" // Bold yellow
}

// highlightFile highlights file paths
func (h *Highlighter) highlightFile(file string) string {
	return "\033[1;35m" + file + "\033[0m" // Bold magenta
}

// highlightAddress highlights network addresses
func (h *Highlighter) highlightAddress(addr string) string {
	return "\033[1;34m" + addr + "\033[0m" // Bold blue
}

// highlightArgument highlights regular arguments
func (h *Highlighter) highlightArgument(arg string) string {
	return "\033[1;37m" + arg + "\033[0m" // Bold white
}

// HighlightJSON highlights JSON syntax
func (h *Highlighter) HighlightJSON(json string) string {
	if !h.colors {
		return json
	}

	// Simple JSON highlighting
	json = strings.ReplaceAll(json, "\"", "\033[32m\"\033[0m") // Green for strings
	json = strings.ReplaceAll(json, ":", "\033[36m:\033[0m")   // Cyan for colons
	json = strings.ReplaceAll(json, ",", "\033[33m,\033[0m")   // Yellow for commas
	json = strings.ReplaceAll(json, "{", "\033[35m{\033[0m")   // Magenta for braces
	json = strings.ReplaceAll(json, "}", "\033[35m}\033[0m")   // Magenta for braces
	json = strings.ReplaceAll(json, "[", "\033[35m[\033[0m")   // Magenta for brackets
	json = strings.ReplaceAll(json, "]", "\033[35m]\033[0m")   // Magenta for brackets

	return json
}

// HighlightYAML highlights YAML syntax
func (h *Highlighter) HighlightYAML(yaml string) string {
	if !h.colors {
		return yaml
	}

	lines := strings.Split(yaml, "\n")
	for i, line := range lines {
		// Highlight keys
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				lines[i] = "\033[1;36m" + parts[0] + "\033[0m:" + parts[1]
			}
		}

		// Highlight list items
		if strings.HasPrefix(strings.TrimSpace(line), "-") {
			lines[i] = strings.Replace(line, "-", "\033[1;33m-\033[0m", 1)
		}
	}

	return strings.Join(lines, "\n")
}

// HighlightError highlights error messages
func (h *Highlighter) HighlightError(err string) string {
	if !h.colors {
		return err
	}
	return "\033[1;31m" + err + "\033[0m" // Bold red
}

// HighlightSuccess highlights success messages
func (h *Highlighter) HighlightSuccess(msg string) string {
	if !h.colors {
		return msg
	}
	return "\033[1;32m" + msg + "\033[0m" // Bold green
}

// HighlightWarning highlights warning messages
func (h *Highlighter) HighlightWarning(msg string) string {
	if !h.colors {
		return msg
	}
	return "\033[1;33m" + msg + "\033[0m" // Bold yellow
}

// HighlightInfo highlights info messages
func (h *Highlighter) HighlightInfo(msg string) string {
	if !h.colors {
		return msg
	}
	return "\033[1;34m" + msg + "\033[0m" // Bold blue
}

// HighlightCode highlights code blocks
func (h *Highlighter) HighlightCode(code string) string {
	if !h.colors {
		return code
	}
	return "\033[2m" + code + "\033[0m" // Dim
}

// HighlightStatus highlights status indicators
func (h *Highlighter) HighlightStatus(status string) string {
	if !h.colors {
		return status
	}

	switch strings.ToLower(status) {
	case "healthy", "success", "ok", "active", "online":
		return "\033[1;32m" + status + "\033[0m" // Bold green
	case "warning", "degraded", "pending":
		return "\033[1;33m" + status + "\033[0m" // Bold yellow
	case "error", "failed", "unhealthy", "inactive", "offline":
		return "\033[1;31m" + status + "\033[0m" // Bold red
	case "info", "unknown":
		return "\033[1;34m" + status + "\033[0m" // Bold blue
	default:
		return "\033[1;37m" + status + "\033[0m" // Bold white
	}
}

// HighlightNumber highlights numeric values
func (h *Highlighter) HighlightNumber(num string) string {
	if !h.colors {
		return num
	}
	return "\033[1;35m" + num + "\033[0m" // Bold magenta
}

// HighlightTimestamp highlights timestamps
func (h *Highlighter) HighlightTimestamp(ts string) string {
	if !h.colors {
		return ts
	}
	return "\033[2m" + ts + "\033[0m" // Dim
}

// HighlightBytes highlights byte values
func (h *Highlighter) HighlightBytes(bytes string) string {
	if !h.colors {
		return bytes
	}
	return "\033[1;36m" + bytes + "\033[0m" // Bold cyan
}

// HighlightPercentage highlights percentage values
func (h *Highlighter) HighlightPercentage(pct string) string {
	if !h.colors {
		return pct
	}

	// Color based on percentage value
	if strings.Contains(pct, "100") || strings.Contains(pct, "9") {
		return "\033[1;31m" + pct + "\033[0m" // Red for high values
	} else if strings.Contains(pct, "8") || strings.Contains(pct, "7") {
		return "\033[1;33m" + pct + "\033[0m" // Yellow for medium-high values
	} else {
		return "\033[1;32m" + pct + "\033[0m" // Green for low values
	}
}

// SetColors enables or disables colors
func (h *Highlighter) SetColors(enabled bool) {
	h.colors = enabled
}

// IsColorsEnabled returns whether colors are enabled
func (h *Highlighter) IsColorsEnabled() bool {
	return h.colors
}
