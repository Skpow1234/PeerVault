package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
	"github.com/Skpow1234/Peervault/internal/cli/network"
)

// LoadBalancerCommand handles load balancer operations
type LoadBalancerCommand struct {
	BaseCommand
	loadBalancer *network.LoadBalancer
}

// NewLoadBalancerCommand creates a new load balancer command
func NewLoadBalancerCommand(client *client.Client, formatter *formatter.Formatter, loadBalancer *network.LoadBalancer) *LoadBalancerCommand {
	return &LoadBalancerCommand{
		BaseCommand: BaseCommand{
			client:    client,
			formatter: formatter,
		},
		loadBalancer: loadBalancer,
	}
}

// Execute executes the load balancer command
func (c *LoadBalancerCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: lb <subcommand> [args...]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "add-server":
		return c.addServer(subArgs)
	case "remove-server":
		return c.removeServer(subArgs)
	case "list-servers":
		return c.listServers()
	case "get-server":
		return c.getServer(subArgs)
	case "select-server":
		return c.selectServer()
	case "update-health":
		return c.updateHealth(subArgs)
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

func (c *LoadBalancerCommand) addServer(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: lb add-server <id> <name> <url> <weight>")
	}

	id := args[0]
	name := args[1]
	url := args[2]
	weight, err := strconv.Atoi(args[3])
	if err != nil {
		return fmt.Errorf("invalid weight: %w", err)
	}

	server := &network.Server{
		ID:          id,
		Name:        name,
		URL:         url,
		Weight:      weight,
		HealthCheck: "/health",
		IsHealthy:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = c.loadBalancer.AddServer(server)
	if err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	c.formatter.PrintSuccess("Server added successfully")
	c.formatter.PrintInfo("ID: " + server.ID)
	c.formatter.PrintInfo("Name: " + server.Name)
	c.formatter.PrintInfo("URL: " + server.URL)
	c.formatter.PrintInfo("Weight: " + strconv.Itoa(server.Weight))

	return nil
}

func (c *LoadBalancerCommand) removeServer(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: lb remove-server <server_id>")
	}

	serverID := args[0]
	err := c.loadBalancer.RemoveServer(serverID)
	if err != nil {
		return fmt.Errorf("failed to remove server: %w", err)
	}

	c.formatter.PrintSuccess("Server removed successfully")
	c.formatter.PrintInfo("Server ID: " + serverID)

	return nil
}

func (c *LoadBalancerCommand) listServers() error {
	servers := c.loadBalancer.ListServers()

	if len(servers) == 0 {
		c.formatter.PrintInfo("No servers found")
		return nil
	}

	c.formatter.PrintInfo("Load Balancer Servers:")
	c.formatter.PrintTable([]string{"ID", "Name", "URL", "Weight", "Healthy", "Response Time", "Connections", "Requests"},
		func() [][]string {
			var rows [][]string
			for _, server := range servers {
				healthy := "Yes"
				if !server.IsHealthy {
					healthy = "No"
				}
				rows = append(rows, []string{
					server.ID,
					server.Name,
					server.URL,
					strconv.Itoa(server.Weight),
					healthy,
					fmt.Sprintf("%d ms", server.ResponseTime),
					strconv.Itoa(server.ActiveConnections),
					strconv.FormatInt(server.TotalRequests, 10),
				})
			}
			return rows
		}())

	return nil
}

func (c *LoadBalancerCommand) getServer(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: lb get-server <server_id>")
	}

	serverID := args[0]
	server, err := c.loadBalancer.GetServer(serverID)
	if err != nil {
		return fmt.Errorf("failed to get server: %w", err)
	}

	c.formatter.PrintInfo("Server details:")
	c.formatter.PrintInfo("ID: " + server.ID)
	c.formatter.PrintInfo("Name: " + server.Name)
	c.formatter.PrintInfo("URL: " + server.URL)
	c.formatter.PrintInfo("Weight: " + strconv.Itoa(server.Weight))
	c.formatter.PrintInfo("Health Check: " + server.HealthCheck)
	c.formatter.PrintInfo("Is Healthy: " + strconv.FormatBool(server.IsHealthy))
	c.formatter.PrintInfo("Response Time: " + fmt.Sprintf("%d ms", server.ResponseTime))
	c.formatter.PrintInfo("Active Connections: " + strconv.Itoa(server.ActiveConnections))
	c.formatter.PrintInfo("Total Requests: " + strconv.FormatInt(server.TotalRequests, 10))
	c.formatter.PrintInfo("Failed Requests: " + strconv.FormatInt(server.FailedRequests, 10))
	c.formatter.PrintInfo("Created: " + server.CreatedAt.Format(time.RFC3339))
	c.formatter.PrintInfo("Updated: " + server.UpdatedAt.Format(time.RFC3339))

	return nil
}

func (c *LoadBalancerCommand) selectServer() error {
	server, err := c.loadBalancer.SelectServer()
	if err != nil {
		return fmt.Errorf("failed to select server: %w", err)
	}

	c.formatter.PrintSuccess("Selected server:")
	c.formatter.PrintInfo("ID: " + server.ID)
	c.formatter.PrintInfo("Name: " + server.Name)
	c.formatter.PrintInfo("URL: " + server.URL)
	c.formatter.PrintInfo("Weight: " + strconv.Itoa(server.Weight))

	return nil
}

func (c *LoadBalancerCommand) updateHealth(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: lb update-health <server_id> <is_healthy> <response_time>")
	}

	serverID := args[0]
	isHealthy, err := strconv.ParseBool(args[1])
	if err != nil {
		return fmt.Errorf("invalid is_healthy value: %w", err)
	}
	responseTime, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid response_time value: %w", err)
	}

	c.loadBalancer.UpdateServerHealth(serverID, isHealthy, responseTime)

	c.formatter.PrintSuccess("Server health updated")
	c.formatter.PrintInfo("Server ID: " + serverID)
	c.formatter.PrintInfo("Is Healthy: " + strconv.FormatBool(isHealthy))
	c.formatter.PrintInfo("Response Time: " + fmt.Sprintf("%d ms", responseTime))

	return nil
}

func (c *LoadBalancerCommand) getStats() error {
	stats := c.loadBalancer.GetStats()

	c.formatter.PrintInfo("Load Balancer Statistics:")
	c.formatter.PrintInfo("Total Requests: " + strconv.FormatInt(stats.TotalRequests, 10))
	c.formatter.PrintInfo("Successful Requests: " + strconv.FormatInt(stats.SuccessfulRequests, 10))
	c.formatter.PrintInfo("Failed Requests: " + strconv.FormatInt(stats.FailedRequests, 10))
	c.formatter.PrintInfo("Average Response Time: " + fmt.Sprintf("%.2f ms", stats.AverageResponseTime))
	c.formatter.PrintInfo("Active Connections: " + strconv.Itoa(stats.ActiveConnections))
	c.formatter.PrintInfo("Healthy Servers: " + strconv.Itoa(stats.HealthyServers))
	c.formatter.PrintInfo("Unhealthy Servers: " + strconv.Itoa(stats.UnhealthyServers))
	c.formatter.PrintInfo("Last Updated: " + stats.LastUpdated.Format(time.RFC3339))

	return nil
}

func (c *LoadBalancerCommand) getConfig(args []string) error {
	config := c.loadBalancer.GetConfig()

	c.formatter.PrintInfo("Load Balancer Configuration:")
	c.formatter.PrintInfo("Algorithm: " + string(config.Algorithm))
	c.formatter.PrintInfo("Health Check Interval: " + config.HealthCheckInterval.String())
	c.formatter.PrintInfo("Health Check Timeout: " + config.HealthCheckTimeout.String())
	c.formatter.PrintInfo("Max Retries: " + strconv.Itoa(config.MaxRetries))
	c.formatter.PrintInfo("Retry Delay: " + config.RetryDelay.String())
	c.formatter.PrintInfo("Sticky Session: " + strconv.FormatBool(config.StickySession))
	c.formatter.PrintInfo("Session Timeout: " + config.SessionTimeout.String())

	return nil
}

func (c *LoadBalancerCommand) updateConfig(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: lb update-config <setting> <value>")
	}

	setting := args[0]
	value := args[1]

	config := c.loadBalancer.GetConfig()

	switch setting {
	case "algorithm":
		config.Algorithm = network.LoadBalancingAlgorithm(value)
	case "health_check_interval":
		interval, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid interval: %w", err)
		}
		config.HealthCheckInterval = interval
	case "health_check_timeout":
		timeout, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
		config.HealthCheckTimeout = timeout
	case "max_retries":
		retries, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid max_retries: %w", err)
		}
		config.MaxRetries = retries
	case "retry_delay":
		delay, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid retry_delay: %w", err)
		}
		config.RetryDelay = delay
	case "sticky_session":
		sticky, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid sticky_session: %w", err)
		}
		config.StickySession = sticky
	case "session_timeout":
		timeout, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid session_timeout: %w", err)
		}
		config.SessionTimeout = timeout
	default:
		return fmt.Errorf("unknown setting: %s", setting)
	}

	err := c.loadBalancer.UpdateConfig(config)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	c.formatter.PrintSuccess("Configuration updated successfully")
	c.formatter.PrintInfo(setting + " = " + value)

	return nil
}

// CacheCommand handles cache operations
type CacheCommand struct {
	BaseCommand
	cacheManager *network.CacheManager
}

// NewCacheCommand creates a new cache command
func NewCacheCommand(client *client.Client, formatter *formatter.Formatter, cacheManager *network.CacheManager) *CacheCommand {
	return &CacheCommand{
		BaseCommand: BaseCommand{
			client:    client,
			formatter: formatter,
		},
		cacheManager: cacheManager,
	}
}

// Execute executes the cache command
func (c *CacheCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cache <subcommand> [args...]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "set":
		return c.setCache(subArgs)
	case "get":
		return c.getCache(subArgs)
	case "delete":
		return c.deleteCache(subArgs)
	case "clear":
		return c.clearCache()
	case "list":
		return c.listCache()
	case "stats":
		return c.getStats()
	case "config":
		return c.getConfig(subArgs)
	case "update-config":
		return c.updateConfig(subArgs)
	case "cache-file":
		return c.cacheFile(subArgs)
	case "get-cached-file":
		return c.getCachedFile(subArgs)
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func (c *CacheCommand) setCache(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: cache set <key> <value> <ttl> [tags...]")
	}

	key := args[0]
	value := args[1]
	ttl, err := time.ParseDuration(args[2])
	if err != nil {
		return fmt.Errorf("invalid TTL: %w", err)
	}
	tags := args[3:]

	err = c.cacheManager.Set(key, value, ttl, tags)
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	c.formatter.PrintSuccess("Cache entry set successfully")
	c.formatter.PrintInfo("Key: " + key)
	c.formatter.PrintInfo("Value: " + value)
	c.formatter.PrintInfo("TTL: " + ttl.String())
	if len(tags) > 0 {
		c.formatter.PrintInfo("Tags: " + strings.Join(tags, ", "))
	}

	return nil
}

func (c *CacheCommand) getCache(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cache get <key>")
	}

	key := args[0]
	result, err := c.cacheManager.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get cache: %w", err)
	}

	c.formatter.PrintSuccess("Cache entry found")
	c.formatter.PrintInfo("Key: " + key)
	c.formatter.PrintInfo("Value: " + fmt.Sprintf("%v", result.Value))
	c.formatter.PrintInfo("Size: " + c.formatter.FormatBytes(result.Size))
	c.formatter.PrintInfo("Created: " + result.CreatedAt.Format(time.RFC3339))
	c.formatter.PrintInfo("Expires: " + result.ExpiresAt.Format(time.RFC3339))

	return nil
}

func (c *CacheCommand) deleteCache(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cache delete <key>")
	}

	key := args[0]
	err := c.cacheManager.Delete(key)
	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	c.formatter.PrintSuccess("Cache entry deleted successfully")
	c.formatter.PrintInfo("Key: " + key)

	return nil
}

func (c *CacheCommand) clearCache() error {
	err := c.cacheManager.Clear()
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	c.formatter.PrintSuccess("Cache cleared successfully")

	return nil
}

func (c *CacheCommand) listCache() error {
	entries := c.cacheManager.ListEntries()

	if len(entries) == 0 {
		c.formatter.PrintInfo("No cache entries found")
		return nil
	}

	c.formatter.PrintInfo("Cache Entries:")
	c.formatter.PrintTable([]string{"Key", "Size", "Created", "Expires", "Access Count", "Tags"},
		func() [][]string {
			var rows [][]string
			for _, entry := range entries {
				rows = append(rows, []string{
					entry.Key,
					c.formatter.FormatBytes(entry.Size),
					entry.CreatedAt.Format("2006-01-02 15:04:05"),
					entry.ExpiresAt.Format("2006-01-02 15:04:05"),
					strconv.FormatInt(entry.AccessCount, 10),
					strings.Join(entry.Tags, ", "),
				})
			}
			return rows
		}())

	return nil
}

func (c *CacheCommand) getStats() error {
	stats := c.cacheManager.GetStats()

	c.formatter.PrintInfo("Cache Statistics:")
	c.formatter.PrintInfo("Total Entries: " + strconv.FormatInt(stats.TotalEntries, 10))
	c.formatter.PrintInfo("Current Size: " + c.formatter.FormatBytes(stats.CurrentSize))
	c.formatter.PrintInfo("Max Size: " + c.formatter.FormatBytes(stats.MaxSize))
	c.formatter.PrintInfo("Hit Count: " + strconv.FormatInt(stats.HitCount, 10))
	c.formatter.PrintInfo("Miss Count: " + strconv.FormatInt(stats.MissCount, 10))
	c.formatter.PrintInfo("Hit Rate: " + fmt.Sprintf("%.2f%%", stats.HitRate*100))
	c.formatter.PrintInfo("Eviction Count: " + strconv.FormatInt(stats.EvictionCount, 10))
	c.formatter.PrintInfo("Expiration Count: " + strconv.FormatInt(stats.ExpirationCount, 10))
	c.formatter.PrintInfo("Last Cleanup: " + stats.LastCleanup.Format(time.RFC3339))
	c.formatter.PrintInfo("Last Updated: " + stats.LastUpdated.Format(time.RFC3339))

	return nil
}

func (c *CacheCommand) getConfig(args []string) error {
	config := c.cacheManager.GetConfig()

	c.formatter.PrintInfo("Cache Configuration:")
	c.formatter.PrintInfo("Max Size: " + c.formatter.FormatBytes(config.MaxSize))
	c.formatter.PrintInfo("Max Entries: " + strconv.Itoa(config.MaxEntries))
	c.formatter.PrintInfo("Default TTL: " + config.DefaultTTL.String())
	c.formatter.PrintInfo("Cleanup Interval: " + config.CleanupInterval.String())
	c.formatter.PrintInfo("Eviction Policy: " + string(config.EvictionPolicy))
	c.formatter.PrintInfo("Compression: " + strconv.FormatBool(config.Compression))
	c.formatter.PrintInfo("Persistence: " + strconv.FormatBool(config.Persistence))

	return nil
}

func (c *CacheCommand) updateConfig(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: cache update-config <setting> <value>")
	}

	setting := args[0]
	value := args[1]

	config := c.cacheManager.GetConfig()

	switch setting {
	case "max_size":
		size, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid size: %w", err)
		}
		config.MaxSize = size
	case "max_entries":
		entries, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid max_entries: %w", err)
		}
		config.MaxEntries = entries
	case "default_ttl":
		ttl, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid TTL: %w", err)
		}
		config.DefaultTTL = ttl
	case "cleanup_interval":
		interval, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid interval: %w", err)
		}
		config.CleanupInterval = interval
	case "eviction_policy":
		config.EvictionPolicy = network.EvictionPolicy(value)
	case "compression":
		compression, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid compression: %w", err)
		}
		config.Compression = compression
	case "persistence":
		persistence, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid persistence: %w", err)
		}
		config.Persistence = persistence
	default:
		return fmt.Errorf("unknown setting: %s", setting)
	}

	err := c.cacheManager.UpdateConfig(config)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	c.formatter.PrintSuccess("Configuration updated successfully")
	c.formatter.PrintInfo(setting + " = " + value)

	return nil
}

func (c *CacheCommand) cacheFile(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: cache cache-file <file_id> <ttl>")
	}

	fileID := args[0]
	ttl, err := time.ParseDuration(args[1])
	if err != nil {
		return fmt.Errorf("invalid TTL: %w", err)
	}

	err = c.cacheManager.CacheFile(fileID, ttl)
	if err != nil {
		return fmt.Errorf("failed to cache file: %w", err)
	}

	c.formatter.PrintSuccess("File cached successfully")
	c.formatter.PrintInfo("File ID: " + fileID)
	c.formatter.PrintInfo("TTL: " + ttl.String())

	return nil
}

func (c *CacheCommand) getCachedFile(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cache get-cached-file <file_id>")
	}

	fileID := args[0]
	content, err := c.cacheManager.GetCachedFile(fileID)
	if err != nil {
		return fmt.Errorf("failed to get cached file: %w", err)
	}

	c.formatter.PrintSuccess("Cached file retrieved")
	c.formatter.PrintInfo("File ID: " + fileID)
	c.formatter.PrintInfo("Size: " + c.formatter.FormatBytes(int64(len(content))))

	return nil
}
