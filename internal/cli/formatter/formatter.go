package formatter

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
)

// Formatter handles output formatting
type Formatter struct {
	outputFormat string
	verbose      bool
}

// New creates a new formatter
func New() *Formatter {
	return &Formatter{
		outputFormat: "table",
		verbose:      false,
	}
}

// SetOutputFormat sets the output format
func (f *Formatter) SetOutputFormat(format string) {
	f.outputFormat = format
}

// SetVerbose sets verbose mode
func (f *Formatter) SetVerbose(verbose bool) {
	f.verbose = verbose
}

// PrintError prints an error message
func (f *Formatter) PrintError(err error) {
	fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
}

// PrintSuccess prints a success message
func (f *Formatter) PrintSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// PrintInfo prints an info message
func (f *Formatter) PrintInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// PrintWarning prints a warning message
func (f *Formatter) PrintWarning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// PrintFileInfo prints file information
func (f *Formatter) PrintFileInfo(file *client.FileInfo) {
	switch f.outputFormat {
	case "json":
		f.printFileInfoJSON(file)
	case "yaml":
		f.printFileInfoYAML(file)
	default:
		f.printFileInfoTable(file)
	}
}

// PrintFileList prints a list of files
func (f *Formatter) PrintFileList(files *client.FileListResponse) {
	switch f.outputFormat {
	case "json":
		f.printFileListJSON(files)
	case "yaml":
		f.printFileListYAML(files)
	default:
		f.printFileListTable(files)
	}
}

// PrintPeerInfo prints peer information
func (f *Formatter) PrintPeerInfo(peer *client.PeerInfo) {
	switch f.outputFormat {
	case "json":
		f.printPeerInfoJSON(peer)
	case "yaml":
		f.printPeerInfoYAML(peer)
	default:
		f.printPeerInfoTable(peer)
	}
}

// PrintPeerList prints a list of peers
func (f *Formatter) PrintPeerList(peers *client.PeerListResponse) {
	switch f.outputFormat {
	case "json":
		f.printPeerListJSON(peers)
	case "yaml":
		f.printPeerListYAML(peers)
	default:
		f.printPeerListTable(peers)
	}
}

// PrintHealth prints health status
func (f *Formatter) PrintHealth(health *client.HealthStatus) {
	switch f.outputFormat {
	case "json":
		f.printHealthJSON(health)
	case "yaml":
		f.printHealthYAML(health)
	default:
		f.printHealthTable(health)
	}
}

// PrintMetrics prints system metrics
func (f *Formatter) PrintMetrics(metrics *client.Metrics) {
	switch f.outputFormat {
	case "json":
		f.printMetricsJSON(metrics)
	case "yaml":
		f.printMetricsYAML(metrics)
	default:
		f.printMetricsTable(metrics)
	}
}

// Table formatting methods
func (f *Formatter) printFileInfoTable(file *client.FileInfo) {
	fmt.Printf("📁 File Information\n")
	fmt.Printf("┌─────────────────┬─────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ Field           │ Value                                                       │\n")
	fmt.Printf("├─────────────────┼─────────────────────────────────────────────────────────────┤\n")
	fmt.Printf("│ ID              │ %-61s │\n", file.ID)
	fmt.Printf("│ Key             │ %-61s │\n", file.Key)
	fmt.Printf("│ Size            │ %-61s │\n", f.formatBytes(file.Size))
	fmt.Printf("│ Hash            │ %-61s │\n", file.Hash)
	fmt.Printf("│ Created At      │ %-61s │\n", file.CreatedAt.Format(time.RFC3339))
	fmt.Printf("│ Owner           │ %-61s │\n", file.Owner)
	fmt.Printf("└─────────────────┴─────────────────────────────────────────────────────────────┘\n")
}

func (f *Formatter) printFileListTable(files *client.FileListResponse) {
	if len(files.Files) == 0 {
		fmt.Println("📁 No files found")
		return
	}

	fmt.Printf("📁 Files (%d total)\n", files.Total)
	fmt.Printf("┌─────────────────────────────────────────────────────────────┬─────────────┬─────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ Key                                                         │ Size        │ Created At                                               │\n")
	fmt.Printf("├─────────────────────────────────────────────────────────────┼─────────────┼─────────────────────────────────────────────────────────────┤\n")

	for _, file := range files.Files {
		key := file.Key
		if len(key) > 60 {
			key = key[:57] + "..."
		}
		fmt.Printf("│ %-61s │ %-11s │ %-61s │\n", key, f.formatBytes(file.Size), file.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("└─────────────────────────────────────────────────────────────┴─────────────┴─────────────────────────────────────────────────────────────┘\n")
}

func (f *Formatter) printPeerInfoTable(peer *client.PeerInfo) {
	fmt.Printf("🌐 Peer Information\n")
	fmt.Printf("┌─────────────────┬─────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ Field           │ Value                                                       │\n")
	fmt.Printf("├─────────────────┼─────────────────────────────────────────────────────────────┤\n")
	fmt.Printf("│ ID              │ %-61s │\n", peer.ID)
	fmt.Printf("│ Address         │ %-61s │\n", peer.Address)
	fmt.Printf("│ Status          │ %-61s │\n", f.getStatusEmoji(peer.Status)+" "+peer.Status)
	fmt.Printf("│ Latency         │ %-61s │\n", f.formatLatency(peer.Latency))
	fmt.Printf("│ Storage         │ %-61s │\n", f.formatBytes(peer.Storage))
	fmt.Printf("│ Last Seen       │ %-61s │\n", f.formatTimeAgo(peer.LastSeen))
	fmt.Printf("└─────────────────┴─────────────────────────────────────────────────────────────┘\n")
}

func (f *Formatter) printPeerListTable(peers *client.PeerListResponse) {
	if len(peers.Peers) == 0 {
		fmt.Println("🌐 No peers found")
		return
	}

	fmt.Printf("🌐 Peers (%d total)\n", peers.Total)
	fmt.Printf("┌─────────────────────────────────────────────────────────────┬─────────────┬─────────────┬─────────────┬─────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ Address                                                     │ Status      │ Latency     │ Storage     │ Last Seen                                               │\n")
	fmt.Printf("├─────────────────────────────────────────────────────────────┼─────────────┼─────────────┼─────────────┼─────────────────────────────────────────────────────────────┤\n")

	for _, peer := range peers.Peers {
		address := peer.Address
		if len(address) > 60 {
			address = address[:57] + "..."
		}
		status := f.getStatusEmoji(peer.Status) + " " + peer.Status
		fmt.Printf("│ %-61s │ %-11s │ %-11s │ %-11s │ %-61s │\n",
			address, status, f.formatLatency(peer.Latency), f.formatBytes(peer.Storage), f.formatTimeAgo(peer.LastSeen))
	}

	fmt.Printf("└─────────────────────────────────────────────────────────────┴─────────────┴─────────────┴─────────────┴─────────────────────────────────────────────────────────────┘\n")
}

func (f *Formatter) printHealthTable(health *client.HealthStatus) {
	fmt.Printf("🏥 System Health\n")
	fmt.Printf("┌─────────────────┬─────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ Field           │ Value                                                       │\n")
	fmt.Printf("├─────────────────┼─────────────────────────────────────────────────────────────┤\n")
	fmt.Printf("│ Overall Status  │ %-61s │\n", f.getStatusEmoji(health.Status)+" "+health.Status)
	fmt.Printf("│ Timestamp       │ %-61s │\n", health.Timestamp.Format(time.RFC3339))
	fmt.Printf("└─────────────────┴─────────────────────────────────────────────────────────────┘\n")

	if len(health.Services) > 0 {
		fmt.Printf("\n🔧 Service Status\n")
		fmt.Printf("┌─────────────────────────────────────────────────────────────┬─────────────────────────────────────────────────────────────┐\n")
		fmt.Printf("│ Service                                                     │ Status                                                       │\n")
		fmt.Printf("├─────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────┤\n")

		for service, status := range health.Services {
			fmt.Printf("│ %-61s │ %-61s │\n", service, f.getStatusEmoji(status)+" "+status)
		}

		fmt.Printf("└─────────────────────────────────────────────────────────────┴─────────────────────────────────────────────────────────────┘\n")
	}
}

func (f *Formatter) printMetricsTable(metrics *client.Metrics) {
	fmt.Printf("📊 System Metrics\n")
	fmt.Printf("┌─────────────────────────────────────────────────────────────┬─────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ Metric                                                      │ Value                                                       │\n")
	fmt.Printf("├─────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────┤\n")
	fmt.Printf("│ Files Stored                                                │ %-61d │\n", metrics.FilesStored)
	fmt.Printf("│ Network Traffic (MB/s)                                      │ %-61.2f │\n", metrics.NetworkTraffic)
	fmt.Printf("│ Active Peers                                                │ %-61d │\n", metrics.ActivePeers)
	fmt.Printf("│ Storage Used                                                │ %-61s │\n", f.formatBytes(metrics.StorageUsed))
	fmt.Printf("└─────────────────────────────────────────────────────────────┴─────────────────────────────────────────────────────────────┘\n")
}

// JSON formatting methods
func (f *Formatter) printFileInfoJSON(file *client.FileInfo) {
	// Simple JSON output - in a real implementation, you'd use proper JSON marshaling
	fmt.Printf("{\n")
	fmt.Printf("  \"id\": \"%s\",\n", file.ID)
	fmt.Printf("  \"key\": \"%s\",\n", file.Key)
	fmt.Printf("  \"size\": %d,\n", file.Size)
	fmt.Printf("  \"hash\": \"%s\",\n", file.Hash)
	fmt.Printf("  \"created_at\": \"%s\",\n", file.CreatedAt.Format(time.RFC3339))
	fmt.Printf("  \"owner\": \"%s\"\n", file.Owner)
	fmt.Printf("}\n")
}

func (f *Formatter) printFileListJSON(files *client.FileListResponse) {
	fmt.Printf("{\n")
	fmt.Printf("  \"files\": [\n")
	for i, file := range files.Files {
		fmt.Printf("    {\n")
		fmt.Printf("      \"id\": \"%s\",\n", file.ID)
		fmt.Printf("      \"key\": \"%s\",\n", file.Key)
		fmt.Printf("      \"size\": %d,\n", file.Size)
		fmt.Printf("      \"hash\": \"%s\",\n", file.Hash)
		fmt.Printf("      \"created_at\": \"%s\",\n", file.CreatedAt.Format(time.RFC3339))
		fmt.Printf("      \"owner\": \"%s\"\n", file.Owner)
		if i < len(files.Files)-1 {
			fmt.Printf("    },\n")
		} else {
			fmt.Printf("    }\n")
		}
	}
	fmt.Printf("  ],\n")
	fmt.Printf("  \"total\": %d\n", files.Total)
	fmt.Printf("}\n")
}

func (f *Formatter) printPeerInfoJSON(peer *client.PeerInfo) {
	fmt.Printf("{\n")
	fmt.Printf("  \"id\": \"%s\",\n", peer.ID)
	fmt.Printf("  \"address\": \"%s\",\n", peer.Address)
	fmt.Printf("  \"status\": \"%s\",\n", peer.Status)
	fmt.Printf("  \"latency\": %d,\n", peer.Latency)
	fmt.Printf("  \"storage\": %d,\n", peer.Storage)
	fmt.Printf("  \"last_seen\": \"%s\"\n", peer.LastSeen.Format(time.RFC3339))
	fmt.Printf("}\n")
}

func (f *Formatter) printPeerListJSON(peers *client.PeerListResponse) {
	fmt.Printf("{\n")
	fmt.Printf("  \"peers\": [\n")
	for i, peer := range peers.Peers {
		fmt.Printf("    {\n")
		fmt.Printf("      \"id\": \"%s\",\n", peer.ID)
		fmt.Printf("      \"address\": \"%s\",\n", peer.Address)
		fmt.Printf("      \"status\": \"%s\",\n", peer.Status)
		fmt.Printf("      \"latency\": %d,\n", peer.Latency)
		fmt.Printf("      \"storage\": %d,\n", peer.Storage)
		fmt.Printf("      \"last_seen\": \"%s\"\n", peer.LastSeen.Format(time.RFC3339))
		if i < len(peers.Peers)-1 {
			fmt.Printf("    },\n")
		} else {
			fmt.Printf("    }\n")
		}
	}
	fmt.Printf("  ],\n")
	fmt.Printf("  \"total\": %d\n", peers.Total)
	fmt.Printf("}\n")
}

func (f *Formatter) printHealthJSON(health *client.HealthStatus) {
	fmt.Printf("{\n")
	fmt.Printf("  \"status\": \"%s\",\n", health.Status)
	fmt.Printf("  \"timestamp\": \"%s\",\n", health.Timestamp.Format(time.RFC3339))
	fmt.Printf("  \"services\": {\n")
	i := 0
	for service, status := range health.Services {
		if i < len(health.Services)-1 {
			fmt.Printf("    \"%s\": \"%s\",\n", service, status)
		} else {
			fmt.Printf("    \"%s\": \"%s\"\n", service, status)
		}
		i++
	}
	fmt.Printf("  }\n")
	fmt.Printf("}\n")
}

func (f *Formatter) printMetricsJSON(metrics *client.Metrics) {
	fmt.Printf("{\n")
	fmt.Printf("  \"files_stored\": %d,\n", metrics.FilesStored)
	fmt.Printf("  \"network_traffic\": %.2f,\n", metrics.NetworkTraffic)
	fmt.Printf("  \"active_peers\": %d,\n", metrics.ActivePeers)
	fmt.Printf("  \"storage_used\": %d\n", metrics.StorageUsed)
	fmt.Printf("}\n")
}

// YAML formatting methods (simplified)
func (f *Formatter) printFileInfoYAML(file *client.FileInfo) {
	fmt.Printf("id: %s\n", file.ID)
	fmt.Printf("key: %s\n", file.Key)
	fmt.Printf("size: %d\n", file.Size)
	fmt.Printf("hash: %s\n", file.Hash)
	fmt.Printf("created_at: %s\n", file.CreatedAt.Format(time.RFC3339))
	fmt.Printf("owner: %s\n", file.Owner)
}

func (f *Formatter) printFileListYAML(files *client.FileListResponse) {
	fmt.Printf("files:\n")
	for _, file := range files.Files {
		fmt.Printf("  - id: %s\n", file.ID)
		fmt.Printf("    key: %s\n", file.Key)
		fmt.Printf("    size: %d\n", file.Size)
		fmt.Printf("    hash: %s\n", file.Hash)
		fmt.Printf("    created_at: %s\n", file.CreatedAt.Format(time.RFC3339))
		fmt.Printf("    owner: %s\n", file.Owner)
	}
	fmt.Printf("total: %d\n", files.Total)
}

func (f *Formatter) printPeerInfoYAML(peer *client.PeerInfo) {
	fmt.Printf("id: %s\n", peer.ID)
	fmt.Printf("address: %s\n", peer.Address)
	fmt.Printf("status: %s\n", peer.Status)
	fmt.Printf("latency: %d\n", peer.Latency)
	fmt.Printf("storage: %d\n", peer.Storage)
	fmt.Printf("last_seen: %s\n", peer.LastSeen.Format(time.RFC3339))
}

func (f *Formatter) printPeerListYAML(peers *client.PeerListResponse) {
	fmt.Printf("peers:\n")
	for _, peer := range peers.Peers {
		fmt.Printf("  - id: %s\n", peer.ID)
		fmt.Printf("    address: %s\n", peer.Address)
		fmt.Printf("    status: %s\n", peer.Status)
		fmt.Printf("    latency: %d\n", peer.Latency)
		fmt.Printf("    storage: %d\n", peer.Storage)
		fmt.Printf("    last_seen: %s\n", peer.LastSeen.Format(time.RFC3339))
	}
	fmt.Printf("total: %d\n", peers.Total)
}

func (f *Formatter) printHealthYAML(health *client.HealthStatus) {
	fmt.Printf("status: %s\n", health.Status)
	fmt.Printf("timestamp: %s\n", health.Timestamp.Format(time.RFC3339))
	fmt.Printf("services:\n")
	for service, status := range health.Services {
		fmt.Printf("  %s: %s\n", service, status)
	}
}

func (f *Formatter) printMetricsYAML(metrics *client.Metrics) {
	fmt.Printf("files_stored: %d\n", metrics.FilesStored)
	fmt.Printf("network_traffic: %.2f\n", metrics.NetworkTraffic)
	fmt.Printf("active_peers: %d\n", metrics.ActivePeers)
	fmt.Printf("storage_used: %d\n", metrics.StorageUsed)
}

// Utility methods
func (f *Formatter) formatBytes(bytes int64) string {
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

func (f *Formatter) formatLatency(latency int64) string {
	if latency < 1000 {
		return fmt.Sprintf("%d μs", latency)
	} else if latency < 1000000 {
		return fmt.Sprintf("%.1f ms", float64(latency)/1000)
	} else {
		return fmt.Sprintf("%.1f s", float64(latency)/1000000)
	}
}

func (f *Formatter) formatTimeAgo(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return fmt.Sprintf("%.0fs ago", duration.Seconds())
	} else if duration < time.Hour {
		return fmt.Sprintf("%.0fm ago", duration.Minutes())
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%.0fh ago", duration.Hours())
	} else {
		return fmt.Sprintf("%.0fd ago", duration.Hours()/24)
	}
}

func (f *Formatter) getStatusEmoji(status string) string {
	switch strings.ToLower(status) {
	case "healthy", "active", "online":
		return "🟢"
	case "degraded", "warning":
		return "🟡"
	case "unhealthy", "inactive", "offline", "error":
		return "🔴"
	default:
		return "⚪"
	}
}
