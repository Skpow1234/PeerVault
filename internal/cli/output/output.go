package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Format represents output format types
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatCSV   Format = "csv"
	FormatText  Format = "text"
)

// Formatter handles different output formats
type Formatter struct {
	format Format
}

// New creates a new output formatter
func New(format Format) *Formatter {
	return &Formatter{format: format}
}

// SetFormat sets the output format
func (f *Formatter) SetFormat(format Format) {
	f.format = format
}

// GetFormat returns the current format
func (f *Formatter) GetFormat() Format {
	return f.format
}

// FormatData formats data according to the current format
func (f *Formatter) FormatData(data interface{}) (string, error) {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(data)
	case FormatYAML:
		return f.formatYAML(data)
	case FormatCSV:
		return f.formatCSV(data)
	case FormatTable:
		return f.formatTable(data)
	case FormatText:
		return f.formatText(data)
	default:
		return f.formatText(data)
	}
}

// formatJSON formats data as JSON
func (f *Formatter) formatJSON(data interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonData), nil
}

// formatYAML formats data as YAML (simplified)
func (f *Formatter) formatYAML(data interface{}) (string, error) {
	// Simplified YAML formatting - in a real implementation, use a proper YAML library
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	// Convert JSON-like formatting to YAML-like (very basic)
	yamlStr := strings.ReplaceAll(string(jsonData), "\"", "")
	yamlStr = strings.ReplaceAll(yamlStr, ":", ": ")
	yamlStr = strings.ReplaceAll(yamlStr, ",", "")
	return yamlStr, nil
}

// formatCSV formats data as CSV
func (f *Formatter) formatCSV(data interface{}) (string, error) {
	var result strings.Builder
	writer := csv.NewWriter(&result)

	// Handle different data types
	switch v := data.(type) {
	case []map[string]interface{}:
		if len(v) == 0 {
			return "", nil
		}

		// Get headers from first row
		var headers []string
		for key := range v[0] {
			headers = append(headers, key)
		}
		_ = writer.Write(headers)

		// Write data rows
		for _, row := range v {
			var values []string
			for _, header := range headers {
				value := fmt.Sprintf("%v", row[header])
				values = append(values, value)
			}
			_ = writer.Write(values)
		}

	case map[string]interface{}:
		// Single row
		var headers []string
		var values []string
		for key, value := range v {
			headers = append(headers, key)
			values = append(values, fmt.Sprintf("%v", value))
		}
		_ = writer.Write(headers)
		_ = writer.Write(values)

	default:
		return "", fmt.Errorf("unsupported data type for CSV: %T", data)
	}

	writer.Flush()
	return result.String(), nil
}

// formatTable formats data as a table
func (f *Formatter) formatTable(data interface{}) (string, error) {
	var result strings.Builder

	switch v := data.(type) {
	case []map[string]interface{}:
		if len(v) == 0 {
			return "No data available", nil
		}

		// Get all unique keys
		keys := make([]string, 0)
		keySet := make(map[string]bool)
		for _, row := range v {
			for key := range row {
				if !keySet[key] {
					keys = append(keys, key)
					keySet[key] = true
				}
			}
		}

		// Calculate column widths
		widths := make(map[string]int)
		for _, key := range keys {
			widths[key] = len(key)
		}

		for _, row := range v {
			for _, key := range keys {
				value := fmt.Sprintf("%v", row[key])
				if len(value) > widths[key] {
					widths[key] = len(value)
				}
			}
		}

		// Print header
		for i, key := range keys {
			if i > 0 {
				result.WriteString(" | ")
			}
			result.WriteString(fmt.Sprintf("%-*s", widths[key], key))
		}
		result.WriteString("\n")

		// Print separator
		for i, key := range keys {
			if i > 0 {
				result.WriteString("-+-")
			}
			result.WriteString(strings.Repeat("-", widths[key]))
		}
		result.WriteString("\n")

		// Print rows
		for _, row := range v {
			for i, key := range keys {
				if i > 0 {
					result.WriteString(" | ")
				}
				value := fmt.Sprintf("%v", row[key])
				result.WriteString(fmt.Sprintf("%-*s", widths[key], value))
			}
			result.WriteString("\n")
		}

	case map[string]interface{}:
		// Single row as key-value pairs
		for key, value := range v {
			result.WriteString(fmt.Sprintf("%-20s: %v\n", key, value))
		}

	default:
		return "", fmt.Errorf("unsupported data type for table: %T", data)
	}

	return result.String(), nil
}

// formatText formats data as plain text
func (f *Formatter) formatText(data interface{}) (string, error) {
	return fmt.Sprintf("%+v", data), nil
}

// FormatFileList formats a list of files
func (f *Formatter) FormatFileList(files []FileInfo) (string, error) {
	switch f.format {
	case FormatTable:
		return f.formatFileListTable(files)
	case FormatJSON:
		return f.formatJSON(files)
	case FormatYAML:
		return f.formatYAML(files)
	case FormatCSV:
		return f.formatFileListCSV(files)
	default:
		return f.formatFileListTable(files)
	}
}

// FileInfo represents file information
type FileInfo struct {
	ID        string    `json:"id" yaml:"id"`
	Name      string    `json:"name" yaml:"name"`
	Size      int64     `json:"size" yaml:"size"`
	Hash      string    `json:"hash" yaml:"hash"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	Owner     string    `json:"owner" yaml:"owner"`
}

// formatFileListTable formats file list as a table
func (f *Formatter) formatFileListTable(files []FileInfo) (string, error) {
	if len(files) == 0 {
		return "No files found", nil
	}

	var result strings.Builder

	// Header
	result.WriteString(fmt.Sprintf("%-12s %-30s %-10s %-16s %-20s %-15s\n",
		"ID", "Name", "Size", "Hash", "Created", "Owner"))
	result.WriteString(strings.Repeat("-", 120) + "\n")

	// Rows
	for _, file := range files {
		size := formatBytes(file.Size)
		created := file.CreatedAt.Format("2006-01-02 15:04:05")
		hash := file.Hash
		if len(hash) > 16 {
			hash = hash[:16] + "..."
		}

		result.WriteString(fmt.Sprintf("%-12s %-30s %-10s %-16s %-20s %-15s\n",
			file.ID[:12], truncateString(file.Name, 30), size, hash, created, file.Owner))
	}

	return result.String(), nil
}

// formatFileListCSV formats file list as CSV
func (f *Formatter) formatFileListCSV(files []FileInfo) (string, error) {
	var result strings.Builder
	writer := csv.NewWriter(&result)

	// Header
	_ = writer.Write([]string{"ID", "Name", "Size", "Hash", "Created", "Owner"})

	// Rows
	for _, file := range files {
		_ = writer.Write([]string{
			file.ID,
			file.Name,
			fmt.Sprintf("%d", file.Size),
			file.Hash,
			file.CreatedAt.Format("2006-01-02 15:04:05"),
			file.Owner,
		})
	}

	writer.Flush()
	return result.String(), nil
}

// FormatPeerList formats a list of peers
func (f *Formatter) FormatPeerList(peers []PeerInfo) (string, error) {
	switch f.format {
	case FormatTable:
		return f.formatPeerListTable(peers)
	case FormatJSON:
		return f.formatJSON(peers)
	case FormatYAML:
		return f.formatYAML(peers)
	case FormatCSV:
		return f.formatPeerListCSV(peers)
	default:
		return f.formatPeerListTable(peers)
	}
}

// PeerInfo represents peer information
type PeerInfo struct {
	ID       string    `json:"id" yaml:"id"`
	Address  string    `json:"address" yaml:"address"`
	Status   string    `json:"status" yaml:"status"`
	Latency  int64     `json:"latency" yaml:"latency"`
	Storage  int64     `json:"storage" yaml:"storage"`
	LastSeen time.Time `json:"last_seen" yaml:"last_seen"`
}

// formatPeerListTable formats peer list as a table
func (f *Formatter) formatPeerListTable(peers []PeerInfo) (string, error) {
	if len(peers) == 0 {
		return "No peers found", nil
	}

	var result strings.Builder

	// Header
	result.WriteString(fmt.Sprintf("%-12s %-20s %-10s %-10s %-12s %-20s\n",
		"ID", "Address", "Status", "Latency", "Storage", "Last Seen"))
	result.WriteString(strings.Repeat("-", 100) + "\n")

	// Rows
	for _, peer := range peers {
		latency := fmt.Sprintf("%dms", peer.Latency)
		storage := formatBytes(peer.Storage)
		lastSeen := peer.LastSeen.Format("2006-01-02 15:04:05")

		result.WriteString(fmt.Sprintf("%-12s %-20s %-10s %-10s %-12s %-20s\n",
			peer.ID[:12], peer.Address, peer.Status, latency, storage, lastSeen))
	}

	return result.String(), nil
}

// formatPeerListCSV formats peer list as CSV
func (f *Formatter) formatPeerListCSV(peers []PeerInfo) (string, error) {
	var result strings.Builder
	writer := csv.NewWriter(&result)

	// Header
	_ = writer.Write([]string{"ID", "Address", "Status", "Latency", "Storage", "Last Seen"})

	// Rows
	for _, peer := range peers {
		_ = writer.Write([]string{
			peer.ID,
			peer.Address,
			peer.Status,
			fmt.Sprintf("%d", peer.Latency),
			fmt.Sprintf("%d", peer.Storage),
			peer.LastSeen.Format("2006-01-02 15:04:05"),
		})
	}

	writer.Flush()
	return result.String(), nil
}

// Utility functions
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
