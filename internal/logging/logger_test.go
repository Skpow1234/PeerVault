package logging

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureLogger_Debug(t *testing.T) {
	// Capture stdout to verify logging output
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Configure logger
	ConfigureLogger("debug")

	// Create a logger and log something
	logger := Logger("test")
	logger.Debug("test debug message", "key", "value")

	// Restore stdout
	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	// Read the output
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)

	output := string(buf[:n])

	// Verify it's JSON and contains expected fields
	assert.Contains(t, output, `"level":"DEBUG"`)
	assert.Contains(t, output, `"msg":"test debug message"`)
	assert.Contains(t, output, `"component":"test"`)
	assert.Contains(t, output, `"key":"value"`)
}

func TestConfigureLogger_Info(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	ConfigureLogger("info")

	logger := Logger("test")
	logger.Info("test info message")

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)

	output := string(buf[:n])
	assert.Contains(t, output, `"level":"INFO"`)
	assert.Contains(t, output, `"msg":"test info message"`)
}

func TestConfigureLogger_Warn(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	ConfigureLogger("warn")

	logger := Logger("test")
	logger.Warn("test warn message")

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)

	output := string(buf[:n])
	assert.Contains(t, output, `"level":"WARN"`)
	assert.Contains(t, output, `"msg":"test warn message"`)
}

func TestConfigureLogger_Error(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	ConfigureLogger("error")

	logger := Logger("test")
	logger.Error("test error message")

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)

	output := string(buf[:n])
	assert.Contains(t, output, `"level":"ERROR"`)
	assert.Contains(t, output, `"msg":"test error message"`)
}

func TestConfigureLogger_DefaultLevel(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	ConfigureLogger("invalid_level")

	logger := Logger("test")
	logger.Info("test message") // Should work with default level (info)

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)

	output := string(buf[:n])
	assert.Contains(t, output, `"level":"INFO"`)
}

func TestLogger(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	ConfigureLogger("info")
	logger := Logger("test_component")

	logger.Info("test message")

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)

	output := string(buf[:n])
	assert.Contains(t, output, `"component":"test_component"`)
}

func TestWithError(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	ConfigureLogger("info")
	logger := WithError(assert.AnError)

	logger.Error("test error")

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)

	output := string(buf[:n])
	assert.Contains(t, output, `"error":"assert.AnError general error for testing`)
}

func TestWithPeer(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	ConfigureLogger("info")
	logger := WithPeer("peer-123")

	logger.Info("test message")

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)

	output := string(buf[:n])
	assert.Contains(t, output, `"peer":"peer-123"`)
}

func TestWithKey(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	ConfigureLogger("info")
	logger := WithKey("file-key-123")

	logger.Info("test message")

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)

	output := string(buf[:n])
	assert.Contains(t, output, `"key":"file-key-123"`)
}

func TestWithBytes(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	ConfigureLogger("info")
	logger := WithBytes(1024)

	logger.Info("test message")

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)

	output := string(buf[:n])
	assert.Contains(t, output, `"bytes":1024`)
}

func TestJSONOutputFormat(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	ConfigureLogger("info")

	logger := Logger("test")
	logger.Info("test message", "key1", "value1", "key2", 42)

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)

	output := string(buf[:n])

	// Verify it's valid JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(strings.TrimSpace(output)), &jsonData)
	require.NoError(t, err)

	// Check required fields
	assert.Equal(t, "INFO", jsonData["level"])
	assert.Equal(t, "test message", jsonData["msg"])
	assert.Equal(t, "test", jsonData["component"])
	assert.Equal(t, "value1", jsonData["key1"])
	assert.Equal(t, float64(42), jsonData["key2"])
}

func TestLogLevelFiltering(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	ConfigureLogger("warn") // Only warn and above

	logger := Logger("test")
	logger.Debug("debug message") // Should be filtered out
	logger.Info("info message")   // Should be filtered out
	logger.Warn("warn message")   // Should appear
	logger.Error("error message") // Should appear

	assert.NoError(t, w.Close())
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)

	output := string(buf[:n])

	// Should contain warn and error, but not debug or info
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
	assert.NotContains(t, output, "debug message")
	assert.NotContains(t, output, "info message")
}
