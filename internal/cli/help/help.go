package help

import (
	"fmt"
	"strings"
)

// HelpManager manages help content and examples
type HelpManager struct {
	commands  map[string]*CommandHelp
	examples  map[string][]string
	tutorials map[string]*Tutorial
}

// CommandHelp represents help information for a command
type CommandHelp struct {
	Name        string
	Description string
	Usage       string
	Examples    []string
	Options     []*Option
	Subcommands []*Subcommand
	Tips        []string
	Related     []string
}

// Option represents a command option
type Option struct {
	Name        string
	Short       string
	Description string
	Required    bool
	Default     string
}

// Subcommand represents a subcommand
type Subcommand struct {
	Name        string
	Description string
	Usage       string
}

// Tutorial represents a tutorial
type Tutorial struct {
	Title         string
	Description   string
	Steps         []*TutorialStep
	Prerequisites []string
}

// TutorialStep represents a tutorial step
type TutorialStep struct {
	Title       string
	Description string
	Command     string
	Explanation string
}

// New creates a new help manager
func New() *HelpManager {
	hm := &HelpManager{
		commands:  make(map[string]*CommandHelp),
		examples:  make(map[string][]string),
		tutorials: make(map[string]*Tutorial),
	}

	hm.initializeHelp()
	return hm
}

// GetCommandHelp returns help for a specific command
func (hm *HelpManager) GetCommandHelp(command string) *CommandHelp {
	return hm.commands[command]
}

// GetExamples returns examples for a command
func (hm *HelpManager) GetExamples(command string) []string {
	return hm.examples[command]
}

// GetTutorial returns a tutorial for a command
func (hm *HelpManager) GetTutorial(command string) *Tutorial {
	return hm.tutorials[command]
}

// initializeHelp initializes all help content
func (hm *HelpManager) initializeHelp() {
	// Store command help
	hm.commands["store"] = &CommandHelp{
		Name:        "store",
		Description: "Store a file in the PeerVault network",
		Usage:       "store <file_path> [options]",
		Examples: []string{
			"store document.pdf",
			"store /path/to/large-file.zip --compress",
			"store image.jpg --encrypt --key mykey",
		},
		Options: []*Option{
			{Name: "compress", Short: "c", Description: "Compress file before storing", Required: false, Default: "false"},
			{Name: "encrypt", Short: "e", Description: "Encrypt file before storing", Required: false, Default: "false"},
			{Name: "key", Short: "k", Description: "Encryption key to use", Required: false, Default: ""},
		},
		Tips: []string{
			"Large files are automatically chunked for better performance",
			"Use --compress for text files to save space",
			"Encrypted files require the same key for retrieval",
		},
		Related: []string{"get", "list", "delete"},
	}

	hm.commands["get"] = &CommandHelp{
		Name:        "get",
		Description: "Retrieve a file from the PeerVault network",
		Usage:       "get <file_id> [output_path]",
		Examples: []string{
			"get abc123def456",
			"get abc123def456 ./downloaded-file.pdf",
			"get abc123def456 --decrypt --key mykey",
		},
		Options: []*Option{
			{Name: "decrypt", Short: "d", Description: "Decrypt file after downloading", Required: false, Default: "false"},
			{Name: "key", Short: "k", Description: "Decryption key to use", Required: false, Default: ""},
		},
		Tips: []string{
			"If no output path is specified, file is saved with original name",
			"Use --decrypt if the file was encrypted during storage",
			"Download progress is shown for large files",
		},
		Related: []string{"store", "list"},
	}

	hm.commands["list"] = &CommandHelp{
		Name:        "list",
		Description: "List files in the PeerVault network",
		Usage:       "list [options]",
		Examples: []string{
			"list",
			"list --format=table",
			"list --filter=*.pdf",
			"list --sort=size --limit=10",
		},
		Options: []*Option{
			{Name: "format", Short: "f", Description: "Output format (table, json, yaml, csv)", Required: false, Default: "table"},
			{Name: "filter", Short: "", Description: "Filter files by pattern", Required: false, Default: ""},
			{Name: "sort", Short: "s", Description: "Sort by field (name, size, date)", Required: false, Default: "date"},
			{Name: "limit", Short: "l", Description: "Limit number of results", Required: false, Default: "100"},
		},
		Tips: []string{
			"Use --filter to find specific file types",
			"JSON format is useful for scripting",
			"Large result sets are paginated automatically",
		},
		Related: []string{"store", "get", "delete"},
	}

	hm.commands["batch"] = &CommandHelp{
		Name:        "batch",
		Description: "Perform batch operations on files",
		Usage:       "batch <operation> [options]",
		Subcommands: []*Subcommand{
			{Name: "upload", Description: "Upload multiple files", Usage: "batch upload <file1> [file2] [file3] ..."},
			{Name: "download", Description: "Download multiple files", Usage: "batch download <output_dir> <file_id1> [file_id2] ..."},
			{Name: "sync", Description: "Synchronize directory", Usage: "batch sync <local_dir> <remote_prefix>"},
		},
		Examples: []string{
			"batch upload *.pdf",
			"batch download ./backup/ file1 file2 file3",
			"batch sync ./documents/ /my-docs/",
		},
		Tips: []string{
			"Batch operations run in parallel for better performance",
			"Progress is shown for each file in the batch",
			"Failed files are reported but don't stop the batch",
		},
		Related: []string{"store", "get", "list"},
	}

	hm.commands["monitor"] = &CommandHelp{
		Name:        "monitor",
		Description: "Monitor system health and performance",
		Usage:       "monitor <subcommand> [options]",
		Subcommands: []*Subcommand{
			{Name: "start", Description: "Start monitoring", Usage: "monitor start [interval]"},
			{Name: "stop", Description: "Stop monitoring", Usage: "monitor stop"},
			{Name: "status", Description: "Show monitoring status", Usage: "monitor status"},
			{Name: "alerts", Description: "Show active alerts", Usage: "monitor alerts"},
			{Name: "dashboard", Description: "Show system dashboard", Usage: "monitor dashboard"},
		},
		Examples: []string{
			"monitor start 30s",
			"monitor dashboard",
			"monitor alerts",
		},
		Tips: []string{
			"Monitoring runs in the background",
			"Alerts are shown in real-time",
			"Dashboard provides system overview",
		},
		Related: []string{"health", "metrics", "status"},
	}

	// Initialize tutorials
	hm.tutorials["store"] = &Tutorial{
		Title:         "Storing Files in PeerVault",
		Description:   "Learn how to store files securely in the PeerVault network",
		Prerequisites: []string{"Connected to PeerVault network", "Have files to store"},
		Steps: []*TutorialStep{
			{
				Title:       "Basic File Storage",
				Description: "Store a simple file",
				Command:     "store my-document.pdf",
				Explanation: "This stores the file and returns a unique file ID",
			},
			{
				Title:       "Compressed Storage",
				Description: "Store a file with compression",
				Command:     "store large-file.txt --compress",
				Explanation: "Compression reduces storage space for text files",
			},
			{
				Title:       "Encrypted Storage",
				Description: "Store a file with encryption",
				Command:     "store sensitive-data.txt --encrypt --key mysecretkey",
				Explanation: "Encryption protects your data with a secret key",
			},
		},
	}

	hm.tutorials["batch"] = &Tutorial{
		Title:         "Batch Operations",
		Description:   "Learn how to perform operations on multiple files",
		Prerequisites: []string{"Multiple files to process", "Understanding of basic commands"},
		Steps: []*TutorialStep{
			{
				Title:       "Batch Upload",
				Description: "Upload multiple files at once",
				Command:     "batch upload file1.pdf file2.pdf file3.pdf",
				Explanation: "All files are uploaded in parallel for better performance",
			},
			{
				Title:       "Batch Download",
				Description: "Download multiple files to a directory",
				Command:     "batch download ./backup/ id1 id2 id3",
				Explanation: "Files are downloaded to the specified directory",
			},
			{
				Title:       "Directory Sync",
				Description: "Synchronize a local directory",
				Command:     "batch sync ./documents/ /my-docs/",
				Explanation: "Only changed files are uploaded, saving time and bandwidth",
			},
		},
	}
}

// FormatCommandHelp formats help for a command
func (hm *HelpManager) FormatCommandHelp(command string, showExamples bool, showTutorial bool) string {
	help := hm.commands[command]
	if help == nil {
		return fmt.Sprintf("No help available for command: %s", command)
	}

	var result strings.Builder

	// Title
	result.WriteString(fmt.Sprintf("📖 %s - %s\n", help.Name, help.Description))
	result.WriteString(strings.Repeat("=", 50) + "\n")

	// Usage
	result.WriteString(fmt.Sprintf("Usage: %s\n\n", help.Usage))

	// Options
	if len(help.Options) > 0 {
		result.WriteString("Options:\n")
		for _, option := range help.Options {
			required := ""
			if option.Required {
				required = " (required)"
			}
			defaultVal := ""
			if option.Default != "" {
				defaultVal = fmt.Sprintf(" (default: %s)", option.Default)
			}

			if option.Short != "" {
				result.WriteString(fmt.Sprintf("  -%s, --%s%s%s\n", option.Short, option.Name, required, defaultVal))
			} else {
				result.WriteString(fmt.Sprintf("  --%s%s%s\n", option.Name, required, defaultVal))
			}
			result.WriteString(fmt.Sprintf("      %s\n", option.Description))
		}
		result.WriteString("\n")
	}

	// Subcommands
	if len(help.Subcommands) > 0 {
		result.WriteString("Subcommands:\n")
		for _, subcmd := range help.Subcommands {
			result.WriteString(fmt.Sprintf("  %s - %s\n", subcmd.Name, subcmd.Description))
			result.WriteString(fmt.Sprintf("      Usage: %s\n", subcmd.Usage))
		}
		result.WriteString("\n")
	}

	// Examples
	if showExamples && len(help.Examples) > 0 {
		result.WriteString("Examples:\n")
		for i, example := range help.Examples {
			result.WriteString(fmt.Sprintf("  %d. %s\n", i+1, example))
		}
		result.WriteString("\n")
	}

	// Tips
	if len(help.Tips) > 0 {
		result.WriteString("💡 Tips:\n")
		for _, tip := range help.Tips {
			result.WriteString(fmt.Sprintf("  • %s\n", tip))
		}
		result.WriteString("\n")
	}

	// Related commands
	if len(help.Related) > 0 {
		result.WriteString("Related commands: ")
		result.WriteString(strings.Join(help.Related, ", "))
		result.WriteString("\n\n")
	}

	// Tutorial
	if showTutorial {
		tutorial := hm.tutorials[command]
		if tutorial != nil {
			result.WriteString("🎓 Tutorial: " + tutorial.Title + "\n")
			result.WriteString(tutorial.Description + "\n\n")

			if len(tutorial.Prerequisites) > 0 {
				result.WriteString("Prerequisites:\n")
				for _, prereq := range tutorial.Prerequisites {
					result.WriteString(fmt.Sprintf("  • %s\n", prereq))
				}
				result.WriteString("\n")
			}

			result.WriteString("Steps:\n")
			for i, step := range tutorial.Steps {
				result.WriteString(fmt.Sprintf("  %d. %s\n", i+1, step.Title))
				result.WriteString(fmt.Sprintf("     %s\n", step.Description))
				result.WriteString(fmt.Sprintf("     $ %s\n", step.Command))
				result.WriteString(fmt.Sprintf("     %s\n\n", step.Explanation))
			}
		}
	}

	return result.String()
}

// GetAvailableCommands returns all available commands
func (hm *HelpManager) GetAvailableCommands() []string {
	var commands []string
	for cmd := range hm.commands {
		commands = append(commands, cmd)
	}
	return commands
}

// SearchHelp searches for commands matching a query
func (hm *HelpManager) SearchHelp(query string) []string {
	var matches []string
	query = strings.ToLower(query)

	for cmd, help := range hm.commands {
		if strings.Contains(strings.ToLower(cmd), query) ||
			strings.Contains(strings.ToLower(help.Description), query) {
			matches = append(matches, cmd)
		}
	}

	return matches
}
