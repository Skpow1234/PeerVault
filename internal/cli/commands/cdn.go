package commands

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
	"github.com/Skpow1234/Peervault/internal/cli/network"
)

// CDNCommand handles CDN operations
type CDNCommand struct {
	BaseCommand
	cdnManager *network.CDNManager
}

// NewCDNCommand creates a new CDN command
func NewCDNCommand(client *client.Client, formatter *formatter.Formatter, cdnManager *network.CDNManager) *CDNCommand {
	return &CDNCommand{
		BaseCommand: BaseCommand{
			client:    client,
			formatter: formatter,
		},
		cdnManager: cdnManager,
	}
}

// Execute executes the CDN command
func (c *CDNCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cdn <subcommand> [args...]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "add-node":
		return c.addNode(subArgs)
	case "remove-node":
		return c.removeNode(subArgs)
	case "list-nodes":
		return c.listNodes()
	case "get-node":
		return c.getNode(subArgs)
	case "nearest-node":
		return c.getNearestNode(subArgs)
	case "cache-file":
		return c.cacheFile(subArgs)
	case "get-cached-file":
		return c.getCachedFile(subArgs)
	case "sync-node":
		return c.syncNode(subArgs)
	case "stats":
		return c.getStats()
	case "config":
		return c.getConfig(subArgs)
	case "update-config":
		return c.updateConfig(subArgs)
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func (c *CDNCommand) addNode(args []string) error {
	if len(args) < 6 {
		return fmt.Errorf("usage: cdn add-node <id> <name> <location> <url> <region> <country> [latitude] [longitude]")
	}

	id := args[0]
	name := args[1]
	location := args[2]
	url := args[3]
	region := args[4]
	country := args[5]

	latitude := 0.0
	longitude := 0.0
	if len(args) > 6 {
		var err error
		latitude, err = strconv.ParseFloat(args[6], 64)
		if err != nil {
			return fmt.Errorf("invalid latitude: %w", err)
		}
	}
	if len(args) > 7 {
		var err error
		longitude, err = strconv.ParseFloat(args[7], 64)
		if err != nil {
			return fmt.Errorf("invalid longitude: %w", err)
		}
	}

	node := &network.CDNNode{
		ID:             id,
		Name:           name,
		Location:       location,
		URL:            url,
		Region:         region,
		Country:        country,
		Latitude:       latitude,
		Longitude:      longitude,
		IsActive:       true,
		Bandwidth:      100 * 1024 * 1024,       // 100MB/s
		Storage:        10 * 1024 * 1024 * 1024, // 10GB
		UsedStorage:    0,
		LastSync:       time.Now(),
		ResponseTime:   0,
		CacheHitRate:   0.0,
		TotalRequests:  0,
		FailedRequests: 0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Metadata:       make(map[string]interface{}),
	}

	err := c.cdnManager.AddNode(node)
	if err != nil {
		return fmt.Errorf("failed to add node: %w", err)
	}

	c.formatter.PrintSuccess("CDN node added successfully")
	c.formatter.PrintInfo("ID: " + node.ID)
	c.formatter.PrintInfo("Name: " + node.Name)
	c.formatter.PrintInfo("Location: " + node.Location)
	c.formatter.PrintInfo("URL: " + node.URL)
	c.formatter.PrintInfo("Region: " + node.Region)
	c.formatter.PrintInfo("Country: " + node.Country)
	c.formatter.PrintInfo("Latitude: " + fmt.Sprintf("%.6f", node.Latitude))
	c.formatter.PrintInfo("Longitude: " + fmt.Sprintf("%.6f", node.Longitude))

	return nil
}

func (c *CDNCommand) removeNode(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cdn remove-node <node_id>")
	}

	nodeID := args[0]
	err := c.cdnManager.RemoveNode(nodeID)
	if err != nil {
		return fmt.Errorf("failed to remove node: %w", err)
	}

	c.formatter.PrintSuccess("CDN node removed successfully")
	c.formatter.PrintInfo("Node ID: " + nodeID)

	return nil
}

func (c *CDNCommand) listNodes() error {
	nodes := c.cdnManager.ListNodes()

	if len(nodes) == 0 {
		c.formatter.PrintInfo("No CDN nodes found")
		return nil
	}

	c.formatter.PrintInfo("CDN Nodes:")
	c.formatter.PrintTable([]string{"ID", "Name", "Location", "Region", "Country", "Active", "Bandwidth", "Storage", "Used", "Hit Rate"},
		func() [][]string {
			var rows [][]string
			for _, node := range nodes {
				active := "Yes"
				if !node.IsActive {
					active = "No"
				}
				rows = append(rows, []string{
					node.ID,
					node.Name,
					node.Location,
					node.Region,
					node.Country,
					active,
					c.formatter.FormatBytes(node.Bandwidth) + "/s",
					c.formatter.FormatBytes(node.Storage),
					c.formatter.FormatBytes(node.UsedStorage),
					fmt.Sprintf("%.2f%%", node.CacheHitRate*100),
				})
			}
			return rows
		}())

	return nil
}

func (c *CDNCommand) getNode(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cdn get-node <node_id>")
	}

	nodeID := args[0]
	node, err := c.cdnManager.GetNode(nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	c.formatter.PrintInfo("CDN Node details:")
	c.formatter.PrintInfo("ID: " + node.ID)
	c.formatter.PrintInfo("Name: " + node.Name)
	c.formatter.PrintInfo("Location: " + node.Location)
	c.formatter.PrintInfo("URL: " + node.URL)
	c.formatter.PrintInfo("Region: " + node.Region)
	c.formatter.PrintInfo("Country: " + node.Country)
	c.formatter.PrintInfo("Latitude: " + fmt.Sprintf("%.6f", node.Latitude))
	c.formatter.PrintInfo("Longitude: " + fmt.Sprintf("%.6f", node.Longitude))
	c.formatter.PrintInfo("Is Active: " + strconv.FormatBool(node.IsActive))
	c.formatter.PrintInfo("Bandwidth: " + c.formatter.FormatBytes(node.Bandwidth) + "/s")
	c.formatter.PrintInfo("Storage: " + c.formatter.FormatBytes(node.Storage))
	c.formatter.PrintInfo("Used Storage: " + c.formatter.FormatBytes(node.UsedStorage))
	c.formatter.PrintInfo("Response Time: " + fmt.Sprintf("%d ms", node.ResponseTime))
	c.formatter.PrintInfo("Cache Hit Rate: " + fmt.Sprintf("%.2f%%", node.CacheHitRate*100))
	c.formatter.PrintInfo("Total Requests: " + strconv.FormatInt(node.TotalRequests, 10))
	c.formatter.PrintInfo("Failed Requests: " + strconv.FormatInt(node.FailedRequests, 10))
	c.formatter.PrintInfo("Last Sync: " + node.LastSync.Format(time.RFC3339))
	c.formatter.PrintInfo("Created: " + node.CreatedAt.Format(time.RFC3339))
	c.formatter.PrintInfo("Updated: " + node.UpdatedAt.Format(time.RFC3339))

	return nil
}

func (c *CDNCommand) getNearestNode(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: cdn nearest-node <latitude> <longitude>")
	}

	latitude, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid latitude: %w", err)
	}
	longitude, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid longitude: %w", err)
	}

	node, err := c.cdnManager.GetNearestNode(latitude, longitude)
	if err != nil {
		return fmt.Errorf("failed to get nearest node: %w", err)
	}

	c.formatter.PrintSuccess("Nearest CDN node:")
	c.formatter.PrintInfo("ID: " + node.ID)
	c.formatter.PrintInfo("Name: " + node.Name)
	c.formatter.PrintInfo("Location: " + node.Location)
	c.formatter.PrintInfo("URL: " + node.URL)
	c.formatter.PrintInfo("Region: " + node.Region)
	c.formatter.PrintInfo("Country: " + node.Country)
	c.formatter.PrintInfo("Latitude: " + fmt.Sprintf("%.6f", node.Latitude))
	c.formatter.PrintInfo("Longitude: " + fmt.Sprintf("%.6f", node.Longitude))

	return nil
}

func (c *CDNCommand) cacheFile(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: cdn cache-file <file_id> <node_id>")
	}

	fileID := args[0]
	nodeID := args[1]

	err := c.cdnManager.CacheFile(fileID, nodeID)
	if err != nil {
		return fmt.Errorf("failed to cache file: %w", err)
	}

	c.formatter.PrintSuccess("File cached successfully")
	c.formatter.PrintInfo("File ID: " + fileID)
	c.formatter.PrintInfo("Node ID: " + nodeID)

	return nil
}

func (c *CDNCommand) getCachedFile(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cdn get-cached-file <file_id>")
	}

	fileID := args[0]
	cdnFile, err := c.cdnManager.GetCachedFile(fileID)
	if err != nil {
		return fmt.Errorf("failed to get cached file: %w", err)
	}

	c.formatter.PrintSuccess("Cached file found")
	c.formatter.PrintInfo("File ID: " + cdnFile.FileID)
	c.formatter.PrintInfo("Node ID: " + cdnFile.NodeID)
	c.formatter.PrintInfo("Size: " + c.formatter.FormatBytes(cdnFile.Size))
	c.formatter.PrintInfo("Checksum: " + cdnFile.Checksum)
	c.formatter.PrintInfo("Cached At: " + cdnFile.CachedAt.Format(time.RFC3339))
	c.formatter.PrintInfo("Last Accessed: " + cdnFile.LastAccessed.Format(time.RFC3339))
	c.formatter.PrintInfo("Access Count: " + strconv.FormatInt(cdnFile.AccessCount, 10))
	c.formatter.PrintInfo("Expires At: " + cdnFile.ExpiresAt.Format(time.RFC3339))

	return nil
}

func (c *CDNCommand) syncNode(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cdn sync-node <node_id>")
	}

	nodeID := args[0]
	err := c.cdnManager.SyncNode(nodeID)
	if err != nil {
		return fmt.Errorf("failed to sync node: %w", err)
	}

	c.formatter.PrintSuccess("Node synchronized successfully")
	c.formatter.PrintInfo("Node ID: " + nodeID)

	return nil
}

func (c *CDNCommand) getStats() error {
	stats := c.cdnManager.GetStats()

	c.formatter.PrintInfo("CDN Statistics:")
	c.formatter.PrintInfo("Total Nodes: " + strconv.Itoa(stats.TotalNodes))
	c.formatter.PrintInfo("Active Nodes: " + strconv.Itoa(stats.ActiveNodes))
	c.formatter.PrintInfo("Inactive Nodes: " + strconv.Itoa(stats.InactiveNodes))
	c.formatter.PrintInfo("Total Bandwidth: " + c.formatter.FormatBytes(stats.TotalBandwidth) + "/s")
	c.formatter.PrintInfo("Used Bandwidth: " + c.formatter.FormatBytes(stats.UsedBandwidth) + "/s")
	c.formatter.PrintInfo("Total Storage: " + c.formatter.FormatBytes(stats.TotalStorage))
	c.formatter.PrintInfo("Used Storage: " + c.formatter.FormatBytes(stats.UsedStorage))
	c.formatter.PrintInfo("Total Requests: " + strconv.FormatInt(stats.TotalRequests, 10))
	c.formatter.PrintInfo("Cache Hits: " + strconv.FormatInt(stats.CacheHits, 10))
	c.formatter.PrintInfo("Cache Misses: " + strconv.FormatInt(stats.CacheMisses, 10))
	c.formatter.PrintInfo("Cache Hit Rate: " + fmt.Sprintf("%.2f%%", stats.CacheHitRate*100))
	c.formatter.PrintInfo("Average Response Time: " + fmt.Sprintf("%.2f ms", stats.AverageResponseTime))
	c.formatter.PrintInfo("Last Updated: " + stats.LastUpdated.Format(time.RFC3339))

	return nil
}

func (c *CDNCommand) getConfig(args []string) error {
	config := c.cdnManager.GetConfig()

	c.formatter.PrintInfo("CDN Configuration:")
	c.formatter.PrintInfo("Auto Sync: " + strconv.FormatBool(config.AutoSync))
	c.formatter.PrintInfo("Sync Interval: " + config.SyncInterval.String())
	c.formatter.PrintInfo("Cache TTL: " + config.CacheTTL.String())
	c.formatter.PrintInfo("Max Bandwidth: " + c.formatter.FormatBytes(config.MaxBandwidth) + "/s")
	c.formatter.PrintInfo("Max Storage: " + c.formatter.FormatBytes(config.MaxStorage))
	c.formatter.PrintInfo("Compression Level: " + strconv.Itoa(config.CompressionLevel))
	c.formatter.PrintInfo("Enable SSL: " + strconv.FormatBool(config.EnableSSL))
	c.formatter.PrintInfo("Enable Gzip: " + strconv.FormatBool(config.EnableGzip))
	c.formatter.PrintInfo("Enable Brotli: " + strconv.FormatBool(config.EnableBrotli))

	return nil
}

func (c *CDNCommand) updateConfig(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: cdn update-config <setting> <value>")
	}

	setting := args[0]
	value := args[1]

	config := c.cdnManager.GetConfig()

	switch setting {
	case "auto_sync":
		autoSync, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid auto_sync: %w", err)
		}
		config.AutoSync = autoSync
	case "sync_interval":
		interval, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid sync_interval: %w", err)
		}
		config.SyncInterval = interval
	case "cache_ttl":
		ttl, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid cache_ttl: %w", err)
		}
		config.CacheTTL = ttl
	case "max_bandwidth":
		bandwidth, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid max_bandwidth: %w", err)
		}
		config.MaxBandwidth = bandwidth
	case "max_storage":
		storage, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid max_storage: %w", err)
		}
		config.MaxStorage = storage
	case "compression_level":
		level, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid compression_level: %w", err)
		}
		if level < 1 || level > 9 {
			return fmt.Errorf("compression level must be between 1 and 9")
		}
		config.CompressionLevel = level
	case "enable_ssl":
		ssl, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid enable_ssl: %w", err)
		}
		config.EnableSSL = ssl
	case "enable_gzip":
		gzip, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid enable_gzip: %w", err)
		}
		config.EnableGzip = gzip
	case "enable_brotli":
		brotli, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid enable_brotli: %w", err)
		}
		config.EnableBrotli = brotli
	default:
		return fmt.Errorf("unknown setting: %s", setting)
	}

	err := c.cdnManager.UpdateConfig(config)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	c.formatter.PrintSuccess("Configuration updated successfully")
	c.formatter.PrintInfo(setting + " = " + value)

	return nil
}

// BandwidthCommand handles bandwidth operations
type BandwidthCommand struct {
	BaseCommand
	bandwidthManager *network.BandwidthManager
}

// NewBandwidthCommand creates a new bandwidth command
func NewBandwidthCommand(client *client.Client, formatter *formatter.Formatter, bandwidthManager *network.BandwidthManager) *BandwidthCommand {
	return &BandwidthCommand{
		BaseCommand: BaseCommand{
			client:    client,
			formatter: formatter,
		},
		bandwidthManager: bandwidthManager,
	}
}

// Execute executes the bandwidth command
func (c *BandwidthCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: bandwidth <subcommand> [args...]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create-policy":
		return c.createPolicy(subArgs)
	case "update-policy":
		return c.updatePolicy(subArgs)
	case "delete-policy":
		return c.deletePolicy(subArgs)
	case "list-policies":
		return c.listPolicies()
	case "get-policy":
		return c.getPolicy(subArgs)
	case "get-user-policy":
		return c.getUserPolicy(subArgs)
	case "check-bandwidth":
		return c.checkBandwidth(subArgs)
	case "record-usage":
		return c.recordUsage(subArgs)
	case "reset-usage":
		return c.resetUsage(subArgs)
	case "get-monitor":
		return c.getMonitor(subArgs)
	case "list-monitors":
		return c.listMonitors()
	case "stats":
		return c.getStats()
	case "config":
		return c.getConfig(subArgs)
	case "update-config":
		return c.updateConfig(subArgs)
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func (c *BandwidthCommand) createPolicy(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: bandwidth create-policy <id> <name> <user_id> <max_bandwidth> [description] [priority]")
	}

	id := args[0]
	name := args[1]
	userID := args[2]
	maxBandwidth, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid max_bandwidth: %w", err)
	}

	description := ""
	priority := 5
	if len(args) > 4 {
		description = args[4]
	}
	if len(args) > 5 {
		priority, err = strconv.Atoi(args[5])
		if err != nil {
			return fmt.Errorf("invalid priority: %w", err)
		}
	}

	policy := &network.BandwidthPolicy{
		ID:             id,
		Name:           name,
		Description:    description,
		UserID:         userID,
		MaxBandwidth:   maxBandwidth,
		BurstBandwidth: maxBandwidth * 2, // 2x burst
		Priority:       priority,
		TimeWindow:     1 * time.Hour,
		AllowedHours:   []int{}, // All hours
		AllowedDays:    []int{}, // All days
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Metadata:       make(map[string]interface{}),
	}

	err = c.bandwidthManager.CreatePolicy(policy)
	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	c.formatter.PrintSuccess("Bandwidth policy created successfully")
	c.formatter.PrintInfo("ID: " + policy.ID)
	c.formatter.PrintInfo("Name: " + policy.Name)
	c.formatter.PrintInfo("User ID: " + policy.UserID)
	c.formatter.PrintInfo("Max Bandwidth: " + c.formatter.FormatBytes(policy.MaxBandwidth) + "/s")
	c.formatter.PrintInfo("Burst Bandwidth: " + c.formatter.FormatBytes(policy.BurstBandwidth) + "/s")
	c.formatter.PrintInfo("Priority: " + strconv.Itoa(policy.Priority))

	return nil
}

func (c *BandwidthCommand) updatePolicy(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: bandwidth update-policy <policy_id> <field> <value>")
	}

	policyID := args[0]
	field := args[1]
	value := args[2]

	updates := &network.BandwidthPolicy{}

	switch field {
	case "name":
		updates.Name = value
	case "description":
		updates.Description = value
	case "max_bandwidth":
		bandwidth, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid max_bandwidth: %w", err)
		}
		updates.MaxBandwidth = bandwidth
	case "burst_bandwidth":
		bandwidth, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid burst_bandwidth: %w", err)
		}
		updates.BurstBandwidth = bandwidth
	case "priority":
		priority, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid priority: %w", err)
		}
		updates.Priority = priority
	case "is_active":
		active, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid is_active: %w", err)
		}
		updates.IsActive = active
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	err := c.bandwidthManager.UpdatePolicy(policyID, updates)
	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	c.formatter.PrintSuccess("Policy updated successfully")
	c.formatter.PrintInfo("Policy ID: " + policyID)
	c.formatter.PrintInfo(field + " = " + value)

	return nil
}

func (c *BandwidthCommand) deletePolicy(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: bandwidth delete-policy <policy_id>")
	}

	policyID := args[0]
	err := c.bandwidthManager.DeletePolicy(policyID)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	c.formatter.PrintSuccess("Policy deleted successfully")
	c.formatter.PrintInfo("Policy ID: " + policyID)

	return nil
}

func (c *BandwidthCommand) listPolicies() error {
	policies := c.bandwidthManager.ListPolicies()

	if len(policies) == 0 {
		c.formatter.PrintInfo("No bandwidth policies found")
		return nil
	}

	c.formatter.PrintInfo("Bandwidth Policies:")
	c.formatter.PrintTable([]string{"ID", "Name", "User ID", "Max Bandwidth", "Burst Bandwidth", "Priority", "Active"},
		func() [][]string {
			var rows [][]string
			for _, policy := range policies {
				active := "Yes"
				if !policy.IsActive {
					active = "No"
				}
				rows = append(rows, []string{
					policy.ID,
					policy.Name,
					policy.UserID,
					c.formatter.FormatBytes(policy.MaxBandwidth) + "/s",
					c.formatter.FormatBytes(policy.BurstBandwidth) + "/s",
					strconv.Itoa(policy.Priority),
					active,
				})
			}
			return rows
		}())

	return nil
}

func (c *BandwidthCommand) getPolicy(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: bandwidth get-policy <policy_id>")
	}

	policyID := args[0]
	policy, err := c.bandwidthManager.GetPolicy(policyID)
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	c.formatter.PrintInfo("Bandwidth Policy details:")
	c.formatter.PrintInfo("ID: " + policy.ID)
	c.formatter.PrintInfo("Name: " + policy.Name)
	c.formatter.PrintInfo("Description: " + policy.Description)
	c.formatter.PrintInfo("User ID: " + policy.UserID)
	c.formatter.PrintInfo("Max Bandwidth: " + c.formatter.FormatBytes(policy.MaxBandwidth) + "/s")
	c.formatter.PrintInfo("Burst Bandwidth: " + c.formatter.FormatBytes(policy.BurstBandwidth) + "/s")
	c.formatter.PrintInfo("Priority: " + strconv.Itoa(policy.Priority))
	c.formatter.PrintInfo("Time Window: " + policy.TimeWindow.String())
	c.formatter.PrintInfo("Is Active: " + strconv.FormatBool(policy.IsActive))
	c.formatter.PrintInfo("Created: " + policy.CreatedAt.Format(time.RFC3339))
	c.formatter.PrintInfo("Updated: " + policy.UpdatedAt.Format(time.RFC3339))

	return nil
}

func (c *BandwidthCommand) getUserPolicy(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: bandwidth get-user-policy <user_id>")
	}

	userID := args[0]
	policy, err := c.bandwidthManager.GetUserPolicy(userID)
	if err != nil {
		return fmt.Errorf("failed to get user policy: %w", err)
	}

	c.formatter.PrintInfo("User Bandwidth Policy:")
	c.formatter.PrintInfo("ID: " + policy.ID)
	c.formatter.PrintInfo("Name: " + policy.Name)
	c.formatter.PrintInfo("User ID: " + policy.UserID)
	c.formatter.PrintInfo("Max Bandwidth: " + c.formatter.FormatBytes(policy.MaxBandwidth) + "/s")
	c.formatter.PrintInfo("Burst Bandwidth: " + c.formatter.FormatBytes(policy.BurstBandwidth) + "/s")
	c.formatter.PrintInfo("Priority: " + strconv.Itoa(policy.Priority))
	c.formatter.PrintInfo("Is Active: " + strconv.FormatBool(policy.IsActive))

	return nil
}

func (c *BandwidthCommand) checkBandwidth(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: bandwidth check-bandwidth <user_id> <requested_bandwidth>")
	}

	userID := args[0]
	requestedBandwidth, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid requested_bandwidth: %w", err)
	}

	result, err := c.bandwidthManager.CheckBandwidth(userID, requestedBandwidth)
	if err != nil {
		return fmt.Errorf("failed to check bandwidth: %w", err)
	}

	c.formatter.PrintInfo("Bandwidth Check Result:")
	c.formatter.PrintInfo("Allowed: " + strconv.FormatBool(result.Allowed))
	c.formatter.PrintInfo("Current Usage: " + c.formatter.FormatBytes(result.CurrentUsage) + "/s")
	c.formatter.PrintInfo("Max Bandwidth: " + c.formatter.FormatBytes(result.MaxBandwidth) + "/s")
	c.formatter.PrintInfo("Utilization: " + fmt.Sprintf("%.2f%%", result.Utilization))
	c.formatter.PrintInfo("Throttled: " + strconv.FormatBool(result.Throttled))
	c.formatter.PrintInfo("Message: " + result.Message)

	return nil
}

func (c *BandwidthCommand) recordUsage(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: bandwidth record-usage <user_id> <bytes_used>")
	}

	userID := args[0]
	bytesUsed, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid bytes_used: %w", err)
	}

	err = c.bandwidthManager.RecordUsage(userID, bytesUsed)
	if err != nil {
		return fmt.Errorf("failed to record usage: %w", err)
	}

	c.formatter.PrintSuccess("Usage recorded successfully")
	c.formatter.PrintInfo("User ID: " + userID)
	c.formatter.PrintInfo("Bytes Used: " + c.formatter.FormatBytes(bytesUsed))

	return nil
}

func (c *BandwidthCommand) resetUsage(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: bandwidth reset-usage <user_id>")
	}

	userID := args[0]
	err := c.bandwidthManager.ResetUsage(userID)
	if err != nil {
		return fmt.Errorf("failed to reset usage: %w", err)
	}

	c.formatter.PrintSuccess("Usage reset successfully")
	c.formatter.PrintInfo("User ID: " + userID)

	return nil
}

func (c *BandwidthCommand) getMonitor(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: bandwidth get-monitor <user_id>")
	}

	userID := args[0]
	monitor, err := c.bandwidthManager.GetMonitor(userID)
	if err != nil {
		return fmt.Errorf("failed to get monitor: %w", err)
	}

	c.formatter.PrintInfo("Bandwidth Monitor:")
	c.formatter.PrintInfo("User ID: " + monitor.UserID)
	c.formatter.PrintInfo("Current Usage: " + c.formatter.FormatBytes(monitor.CurrentUsage) + "/s")
	c.formatter.PrintInfo("Peak Usage: " + c.formatter.FormatBytes(monitor.PeakUsage) + "/s")
	c.formatter.PrintInfo("Total Usage: " + c.formatter.FormatBytes(monitor.TotalUsage))
	c.formatter.PrintInfo("Last Reset: " + monitor.LastReset.Format(time.RFC3339))
	c.formatter.PrintInfo("Last Updated: " + monitor.LastUpdated.Format(time.RFC3339))
	c.formatter.PrintInfo("Usage History Entries: " + strconv.Itoa(len(monitor.UsageHistory)))

	return nil
}

func (c *BandwidthCommand) listMonitors() error {
	monitors := c.bandwidthManager.ListMonitors()

	if len(monitors) == 0 {
		c.formatter.PrintInfo("No bandwidth monitors found")
		return nil
	}

	c.formatter.PrintInfo("Bandwidth Monitors:")
	c.formatter.PrintTable([]string{"User ID", "Current Usage", "Peak Usage", "Total Usage", "Last Reset", "History Entries"},
		func() [][]string {
			var rows [][]string
			for _, monitor := range monitors {
				rows = append(rows, []string{
					monitor.UserID,
					c.formatter.FormatBytes(monitor.CurrentUsage) + "/s",
					c.formatter.FormatBytes(monitor.PeakUsage) + "/s",
					c.formatter.FormatBytes(monitor.TotalUsage),
					monitor.LastReset.Format("2006-01-02 15:04:05"),
					strconv.Itoa(len(monitor.UsageHistory)),
				})
			}
			return rows
		}())

	return nil
}

func (c *BandwidthCommand) getStats() error {
	stats := c.bandwidthManager.GetStats()

	c.formatter.PrintInfo("Bandwidth Statistics:")
	c.formatter.PrintInfo("Total Policies: " + strconv.Itoa(stats.TotalPolicies))
	c.formatter.PrintInfo("Active Policies: " + strconv.Itoa(stats.ActivePolicies))
	c.formatter.PrintInfo("Total Monitors: " + strconv.Itoa(stats.TotalMonitors))
	c.formatter.PrintInfo("Total Bandwidth: " + c.formatter.FormatBytes(stats.TotalBandwidth) + "/s")
	c.formatter.PrintInfo("Used Bandwidth: " + c.formatter.FormatBytes(stats.UsedBandwidth) + "/s")
	c.formatter.PrintInfo("Available Bandwidth: " + c.formatter.FormatBytes(stats.AvailableBandwidth) + "/s")
	c.formatter.PrintInfo("Utilization Rate: " + fmt.Sprintf("%.2f%%", stats.UtilizationRate))
	c.formatter.PrintInfo("Throttled Users: " + strconv.Itoa(stats.ThrottledUsers))
	c.formatter.PrintInfo("Alerted Users: " + strconv.Itoa(stats.AlertedUsers))
	c.formatter.PrintInfo("Last Updated: " + stats.LastUpdated.Format(time.RFC3339))

	return nil
}

func (c *BandwidthCommand) getConfig(args []string) error {
	config := c.bandwidthManager.GetConfig()

	c.formatter.PrintInfo("Bandwidth Configuration:")
	c.formatter.PrintInfo("Monitoring Interval: " + config.MonitoringInterval.String())
	c.formatter.PrintInfo("History Retention: " + config.HistoryRetention.String())
	c.formatter.PrintInfo("Default Policy: " + config.DefaultPolicy)
	c.formatter.PrintInfo("Enable Throttling: " + strconv.FormatBool(config.EnableThrottling))
	c.formatter.PrintInfo("Throttle Threshold: " + fmt.Sprintf("%.1f%%", config.ThrottleThreshold*100))
	c.formatter.PrintInfo("Alert Threshold: " + fmt.Sprintf("%.1f%%", config.AlertThreshold*100))

	return nil
}

func (c *BandwidthCommand) updateConfig(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: bandwidth update-config <setting> <value>")
	}

	setting := args[0]
	value := args[1]

	config := c.bandwidthManager.GetConfig()

	switch setting {
	case "monitoring_interval":
		interval, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid monitoring_interval: %w", err)
		}
		config.MonitoringInterval = interval
	case "history_retention":
		retention, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid history_retention: %w", err)
		}
		config.HistoryRetention = retention
	case "default_policy":
		config.DefaultPolicy = value
	case "enable_throttling":
		throttling, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid enable_throttling: %w", err)
		}
		config.EnableThrottling = throttling
	case "throttle_threshold":
		threshold, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid throttle_threshold: %w", err)
		}
		config.ThrottleThreshold = threshold
	case "alert_threshold":
		threshold, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid alert_threshold: %w", err)
		}
		config.AlertThreshold = threshold
	default:
		return fmt.Errorf("unknown setting: %s", setting)
	}

	err := c.bandwidthManager.UpdateConfig(config)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	c.formatter.PrintSuccess("Configuration updated successfully")
	c.formatter.PrintInfo(setting + " = " + value)

	return nil
}
