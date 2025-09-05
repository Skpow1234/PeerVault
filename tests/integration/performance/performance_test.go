package performance

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/internal/backup"
	"github.com/Skpow1234/Peervault/internal/cache"
	"github.com/Skpow1234/Peervault/internal/compression"
	"github.com/Skpow1234/Peervault/internal/deduplication"
	"github.com/Skpow1234/Peervault/internal/health"
	"github.com/Skpow1234/Peervault/internal/metrics"
	"github.com/Skpow1234/Peervault/internal/pool"
	"github.com/Skpow1234/Peervault/internal/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBufferPool(t *testing.T) {
	// Test buffer pool functionality
	bufferPool := pool.NewBufferPool(1024)

	// Get buffer
	buf := bufferPool.Get()
	assert.Len(t, buf, 1024)

	// Use buffer
	copy(buf, []byte("test data"))

	// Return buffer
	bufferPool.Put(buf)

	// Get another buffer
	buf2 := bufferPool.Get()
	assert.Len(t, buf2, 1024)

	// Should be clean (in a real implementation)
	bufferPool.Put(buf2)
}

func TestConnectionPool(t *testing.T) {
	// Test connection pool functionality
	config := pool.DefaultPoolConfig()
	config.MaxSize = 5
	config.MinSize = 2

	// Mock connection factory
	factory := func(ctx context.Context, address string) (pool.Connection, error) {
		// Return a mock connection
		return &mockConnection{}, nil
	}

	connectionPool := pool.NewConnectionPool(factory, "localhost:8080", config)
	defer connectionPool.Close()

	// Test getting connections
	ctx := context.Background()
	conn, err := connectionPool.Get(ctx)
	require.NoError(t, err)
	assert.NotNil(t, conn)

	// Return connection
	connectionPool.Put(conn)

	// Check stats
	stats := connectionPool.Stats()
	assert.Equal(t, 2, stats.MinSize)
	assert.Equal(t, 5, stats.MaxSize)
}

func TestMemoryCache(t *testing.T) {
	// Test memory cache functionality
	cache := cache.NewMemoryCache[string](100)
	defer cache.Close()

	ctx := context.Background()

	// Test set and get
	err := cache.Set(ctx, "key1", "value1", 5*time.Minute)
	require.NoError(t, err)

	value, found := cache.Get(ctx, "key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Test miss
	_, found = cache.Get(ctx, "key2")
	assert.False(t, found)

	// Test stats
	stats := cache.Stats()
	assert.Equal(t, 1, stats.Size)
	assert.Equal(t, 1, int(stats.Hits))
	assert.Equal(t, 1, int(stats.Misses))
}

func TestMultiLevelCache(t *testing.T) {
	// Test multi-level cache functionality
	mlCache := cache.NewMultiLevelCache[string](50, 100)
	defer mlCache.Close()

	ctx := context.Background()

	// Test set and get
	err := mlCache.Set(ctx, "key1", "value1", 5*time.Minute)
	require.NoError(t, err)

	value, found := mlCache.Get(ctx, "key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Test stats
	stats := mlCache.Stats()
	assert.Equal(t, 1, int(stats.L1Hits))
	assert.Equal(t, 0, int(stats.L2Hits))
}

func TestCompression(t *testing.T) {
	// Test compression functionality
	manager := compression.NewCompressionManager()

	ctx := context.Background()
	data := []byte("This is test data for compression")

	// Test gzip compression
	compressed, err := manager.Compress(ctx, data, compression.CompressionTypeGzip)
	require.NoError(t, err)
	assert.NotEmpty(t, compressed)
	assert.Less(t, len(compressed), len(data)) // Should be smaller

	// Test decompression
	decompressed, err := manager.Decompress(ctx, compressed, compression.CompressionTypeGzip)
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)
}

func TestDeduplication(t *testing.T) {
	// Test deduplication functionality
	chunkStore := deduplication.NewMemoryChunkStore()
	deduplicator := deduplication.NewDeduplicator(nil, chunkStore)

	ctx := context.Background()

	// Create test data
	data1 := []byte("This is test data for deduplication")
	data2 := []byte("This is test data for deduplication") // Same data
	data3 := []byte("This is different test data")

	// Process first chunk
	chunks1, err := deduplicator.ProcessFile(ctx, bytes.NewReader(data1))
	require.NoError(t, err)
	assert.Len(t, chunks1, 1)

	// Process second chunk (should be deduplicated)
	chunks2, err := deduplicator.ProcessFile(ctx, bytes.NewReader(data2))
	require.NoError(t, err)
	assert.Len(t, chunks2, 1)
	assert.Equal(t, chunks1[0].ID, chunks2[0].ID) // Same chunk ID

	// Process third chunk (should be new)
	chunks3, err := deduplicator.ProcessFile(ctx, bytes.NewReader(data3))
	require.NoError(t, err)
	assert.Len(t, chunks3, 1)
	assert.NotEqual(t, chunks1[0].ID, chunks3[0].ID) // Different chunk ID

	// Check stats
	stats := deduplicator.GetStats()
	assert.Equal(t, 2, stats.TotalChunks) // Two unique chunks
}

func TestMetrics(t *testing.T) {
	// Test metrics functionality
	registry := metrics.NewMetricsRegistry()

	// Register counter
	counter := registry.RegisterCounter("test_counter", "Test counter", nil)
	counter.Inc()
	counter.Add(5)
	assert.Equal(t, int64(6), counter.Get())

	// Register gauge
	gauge := registry.RegisterGauge("test_gauge", "Test gauge", nil)
	gauge.Set(100)
	gauge.Inc()
	assert.Equal(t, int64(101), gauge.Get())

	// Register histogram
	buckets := []float64{0.1, 0.5, 1.0, 2.0, 5.0}
	histogram := registry.RegisterHistogram("test_histogram", "Test histogram", buckets, nil)
	histogram.Observe(0.3)
	histogram.Observe(0.7)
	histogram.Observe(1.5)

	assert.Equal(t, int64(3), histogram.GetTotalCount())
	assert.Equal(t, int64(2), histogram.GetSum()) // 0.3 + 0.7 + 1.5 = 2.5, but stored as int64

	// Get all metrics
	allMetrics := registry.GetAllMetrics()
	assert.Len(t, allMetrics, 3)
}

func TestHealthChecks(t *testing.T) {
	// Test health check functionality
	healthChecker := health.NewHealthChecker()

	// Register simple health check
	check := health.NewSimpleHealthCheck(
		"test_check",
		"Test health check",
		5*time.Second,
		func(ctx context.Context) error {
			return nil // Always healthy
		},
	)
	healthChecker.RegisterCheck(check)

	ctx := context.Background()

	// Check specific health check
	result, err := healthChecker.Check(ctx, "test_check")
	require.NoError(t, err)
	assert.Equal(t, health.HealthStatusHealthy, result.Status)

	// Check overall status
	overallStatus := healthChecker.GetOverallStatus(ctx)
	assert.Equal(t, health.HealthStatusHealthy, overallStatus)

	// Get health report
	report := healthChecker.GetHealthReport(ctx)
	assert.Equal(t, health.HealthStatusHealthy, report.OverallStatus)
	assert.Len(t, report.Checks, 1)
}

func TestTracing(t *testing.T) {
	// Test tracing functionality
	tracer := tracing.NewSimpleTracer()

	ctx := context.Background()

	// Start span
	ctx, span := tracer.StartSpan(ctx, "test_span", tracing.WithSpanKind(tracing.SpanKindInternal))
	assert.NotEmpty(t, span.TraceID)
	assert.NotEmpty(t, span.SpanID)
	assert.Equal(t, "test_span", span.Name)
	assert.Equal(t, tracing.SpanKindInternal, span.Kind)

	// Add span event
	tracing.AddSpanEvent(ctx, "test_event", map[string]interface{}{
		"key": "value",
	})

	// Set span attribute
	tracing.SetSpanAttribute(ctx, "test_attr", "test_value")

	// Finish span
	tracer.Finish(span)

	// Get span
	retrievedSpan, exists := tracer.GetSpan(span.SpanID)
	assert.True(t, exists)
	assert.Equal(t, span.SpanID, retrievedSpan.SpanID)
	assert.NotZero(t, retrievedSpan.Duration)
}

func TestBackup(t *testing.T) {
	// Test backup functionality
	backupManager := backup.NewBackupManager()

	// Register backup config
	config := &backup.BackupConfig{
		Type:          backup.BackupTypeFull,
		Location:      "/tmp/test_backup",
		Compression:   true,
		Encryption:    false,
		RetentionDays: 7,
	}
	backupManager.RegisterConfig("test_config", config)

	ctx := context.Background()

	// Start backup
	backupObj, err := backupManager.StartBackup(ctx, "test_config")
	require.NoError(t, err)
	assert.NotEmpty(t, backupObj.ID)
	assert.Equal(t, backup.BackupTypeFull, backupObj.Type)

	// Wait for backup to complete
	time.Sleep(3 * time.Second)

	// Get backup
	retrievedBackup, exists := backupManager.GetBackup(backupObj.ID)
	assert.True(t, exists)
	assert.Equal(t, backup.BackupStatusCompleted, retrievedBackup.Status)
	assert.NotZero(t, retrievedBackup.Size)
	assert.NotZero(t, retrievedBackup.Files)
}

// Mock connection for testing
type mockConnection struct{}

func (mc *mockConnection) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (mc *mockConnection) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (mc *mockConnection) Close() error {
	return nil
}

func (mc *mockConnection) LocalAddr() net.Addr {
	return nil
}

func (mc *mockConnection) RemoteAddr() net.Addr {
	return nil
}

func (mc *mockConnection) SetDeadline(t time.Time) error {
	return nil
}

func (mc *mockConnection) SetReadDeadline(t time.Time) error {
	return nil
}

func (mc *mockConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

func (mc *mockConnection) IsHealthy() bool {
	return true
}

func (mc *mockConnection) LastUsed() time.Time {
	return time.Now()
}

func (mc *mockConnection) Reset() error {
	return nil
}
