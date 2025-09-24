package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/mocking"
	"gopkg.in/yaml.v3"
)

func main() {
	var (
		configFile = flag.String("config", "config/mock-server.yaml", "Configuration file path")
		port       = flag.Int("port", 3001, "Server port")
		host       = flag.String("host", "localhost", "Server host")
		generate   = flag.Bool("generate", false, "Generate mock scenarios from OpenAPI spec")
		spec       = flag.String("spec", "docs/api/peervault-rest-api.yaml", "OpenAPI specification file")
		output     = flag.String("output", "tests/api/mock-data", "Output directory for generated scenarios")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Setup logging
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Load configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Override config with command line flags
	if *port != 3001 {
		config.Port = *port
	}
	if *host != "localhost" {
		config.Host = *host
	}

	// Handle generation mode
	if *generate {
		if err := generateMockScenarios(*spec, *output, logger); err != nil {
			logger.Error("Failed to generate mock scenarios", "error", err)
			os.Exit(1)
		}
		logger.Info("Mock scenarios generated successfully", "output", *output)
		return
	}

	// Create and start mock server
	mockServer := mocking.NewMockServer(config, logger)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutdown signal received")
		cancel()
	}()

	// Start server
	if err := mockServer.Start(ctx); err != nil {
		logger.Error("Failed to start mock server", "error", err)
		os.Exit(1)
	}

	// Wait for shutdown
	<-ctx.Done()

	// Stop server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := mockServer.Stop(shutdownCtx); err != nil {
		logger.Error("Failed to stop mock server", "error", err)
		os.Exit(1)
	}

	logger.Info("Mock server stopped")
}

// loadConfig loads configuration from file
func loadConfig(configFile string) (*mocking.MockConfig, error) {
	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return mocking.DefaultMockConfig(), nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config mocking.MockConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// generateMockScenarios generates mock scenarios from OpenAPI specification
func generateMockScenarios(specFile, outputDir string, logger *slog.Logger) error {
	logger.Info("Generating mock scenarios", "spec", specFile, "output", outputDir)

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Read OpenAPI spec
	specData, err := os.ReadFile(specFile)
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI spec: %w", err)
	}

	// Parse OpenAPI spec (simplified - in production, use proper OpenAPI parser)
	var spec map[string]interface{}
	if err := yaml.Unmarshal(specData, &spec); err != nil {
		return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	// Generate scenarios
	scenarios := generateScenariosFromSpec(spec, logger)

	// Write scenarios to files
	for name, scenario := range scenarios {
		filename := fmt.Sprintf("%s/%s.yaml", outputDir, name)
		if err := writeScenarioToFile(scenario, filename); err != nil {
			logger.Error("Failed to write scenario", "name", name, "error", err)
			continue
		}
		logger.Info("Generated scenario", "name", name, "file", filename)
	}

	return nil
}

// generateScenariosFromSpec generates scenarios from OpenAPI specification
func generateScenariosFromSpec(spec map[string]interface{}, logger *slog.Logger) map[string]*mocking.Scenario {
	scenarios := make(map[string]*mocking.Scenario)

	// Extract paths from spec
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		logger.Warn("No paths found in OpenAPI spec")
		return scenarios
	}

	// Generate scenarios for each path
	for path, pathItem := range paths {
		pathItemMap, ok := pathItem.(map[string]interface{})
		if !ok {
			continue
		}

		// Generate scenarios for each HTTP method
		for method, operation := range pathItemMap {
			if !isValidHTTPMethod(method) {
				continue
			}

			operationMap, ok := operation.(map[string]interface{})
			if !ok {
				continue
			}

			// Generate success scenario
			successScenario := generateSuccessScenario(path, method, operationMap)
			scenarioName := fmt.Sprintf("%s-%s-success", method, sanitizePath(path))
			scenarios[scenarioName] = successScenario

			// Generate error scenarios
			errorScenarios := generateErrorScenarios(path, method, operationMap)
			for errorName, errorScenario := range errorScenarios {
				scenarios[errorName] = errorScenario
			}
		}
	}

	return scenarios
}

// generateSuccessScenario generates a success scenario for an operation
func generateSuccessScenario(path, method string, operation map[string]interface{}) *mocking.Scenario {
	// Extract operation ID
	operationID, _ := operation["operationId"].(string)
	if operationID == "" {
		operationID = fmt.Sprintf("%s-%s", method, sanitizePath(path))
	}

	// Generate response
	response := &mocking.MockResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: generateMockResponseBody(operation),
	}

	// Generate conditions
	conditions := []mocking.Condition{
		{
			Type:     "path",
			Key:      "path",
			Value:    path,
			Operator: "equals",
			Required: true,
		},
	}

	return &mocking.Scenario{
		Name:        fmt.Sprintf("%s-%s-success", method, sanitizePath(path)),
		Description: fmt.Sprintf("Success response for %s %s", method, path),
		Conditions:  conditions,
		Response:    response,
		Enabled:     true,
	}
}

// generateErrorScenarios generates error scenarios for an operation
func generateErrorScenarios(path, method string, operation map[string]interface{}) map[string]*mocking.Scenario {
	scenarios := make(map[string]*mocking.Scenario)

	// Common error scenarios
	errorCodes := []int{400, 401, 403, 404, 500}
	errorMessages := map[int]string{
		400: "Bad Request",
		401: "Unauthorized",
		403: "Forbidden",
		404: "Not Found",
		500: "Internal Server Error",
	}

	for _, code := range errorCodes {
		scenarioName := fmt.Sprintf("%s-%s-error-%d", method, sanitizePath(path), code)
		scenarios[scenarioName] = &mocking.Scenario{
			Name:        scenarioName,
			Description: fmt.Sprintf("Error %d response for %s %s", code, method, path),
			Conditions: []mocking.Condition{
				{
					Type:     "path",
					Key:      "path",
					Value:    path,
					Operator: "equals",
					Required: true,
				},
			},
			Response: &mocking.MockResponse{
				StatusCode: code,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: map[string]interface{}{
					"error":   errorMessages[code],
					"code":    code,
					"message": errorMessages[code],
				},
			},
			Enabled: true,
		}
	}

	return scenarios
}

// generateMockResponseBody generates a mock response body
func generateMockResponseBody(operation map[string]interface{}) interface{} {
	// Try to extract response schema from OpenAPI spec
	responses, ok := operation["responses"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{
			"message":   "Mock response",
			"timestamp": time.Now().UTC(),
		}
	}

	// Look for 200 response
	successResponse, ok := responses["200"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{
			"message":   "Mock response",
			"timestamp": time.Now().UTC(),
		}
	}

	// Extract content schema
	if _, ok := successResponse["content"].(map[string]interface{}); !ok {
		return map[string]interface{}{
			"message":   "Mock response",
			"timestamp": time.Now().UTC(),
		}
	}

	// Generate mock data based on schema (simplified)
	return map[string]interface{}{
		"id":        "mock-id",
		"message":   "Mock response",
		"timestamp": time.Now().UTC(),
		"data":      "mock-data",
	}
}

// writeScenarioToFile writes a scenario to a YAML file
func writeScenarioToFile(scenario *mocking.Scenario, filename string) error {
	data, err := yaml.Marshal([]*mocking.Scenario{scenario})
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// isValidHTTPMethod checks if a string is a valid HTTP method
func isValidHTTPMethod(method string) bool {
	validMethods := map[string]bool{
		"get":     true,
		"post":    true,
		"put":     true,
		"patch":   true,
		"delete":  true,
		"head":    true,
		"options": true,
	}
	return validMethods[method]
}

// sanitizePath sanitizes a path for use in scenario names
func sanitizePath(path string) string {
	// Replace path parameters and special characters
	path = strings.ReplaceAll(path, "/", "-")
	path = strings.ReplaceAll(path, "{", "")
	path = strings.ReplaceAll(path, "}", "")
	path = strings.ReplaceAll(path, " ", "-")
	return strings.Trim(path, "-")
}
