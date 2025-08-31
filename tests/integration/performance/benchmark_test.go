package performance

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/tests/utils"
)

// BenchmarkStoreGet measures the performance of store and get operations
func BenchmarkStoreGet(b *testing.B) {
	// Create test network
	config := utils.NetworkConfig{
		BootstrapNodes: []utils.NodeConfig{
			{Name: "bootstrap", ListenAddr: ":7001"},
		},
		ClientNodes: []utils.NodeConfig{
			{Name: "client", ListenAddr: ":7002", BootstrapNodes: []string{":7001"}},
		},
	}

	manager := utils.CreateTestNetwork(&testing.T{}, config)
	defer manager.StopAll()

	// Get servers
	bootstrap := manager.GetServer("bootstrap")
	client := manager.GetServer("client")

	// Create test data generator
	dataGen := utils.NewTestDataGenerator()
	testData := dataGen.GenerateSmallFile()

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := "benchmark_file_" + string(rune(i))

		// Store operation
		err := bootstrap.Store(ctx, key, bytes.NewReader(testData))
		if err != nil {
			b.Fatalf("Store failed: %v", err)
		}

		// Get operation
		reader, err := client.Get(ctx, key)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}

		// Read data
		_, err = io.ReadAll(reader)
		if err != nil {
			b.Fatalf("Read failed: %v", err)
		}

		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	}
}

// BenchmarkLargeFileTransfer measures performance with large files
func BenchmarkLargeFileTransfer(b *testing.B) {
	// Create test network
	config := utils.NetworkConfig{
		BootstrapNodes: []utils.NodeConfig{
			{Name: "bootstrap", ListenAddr: ":7003"},
		},
		ClientNodes: []utils.NodeConfig{
			{Name: "client", ListenAddr: ":7004", BootstrapNodes: []string{":7003"}},
		},
	}

	manager := utils.CreateTestNetwork(&testing.T{}, config)
	defer manager.StopAll()

	// Get servers
	bootstrap := manager.GetServer("bootstrap")
	client := manager.GetServer("client")

	// Create large test data (100KB)
	dataGen := utils.NewTestDataGenerator()
	testData := dataGen.GenerateMediumFile()

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := "large_benchmark_file_" + string(rune(i))

		// Store operation
		start := time.Now()
		err := bootstrap.Store(ctx, key, bytes.NewReader(testData))
		if err != nil {
			b.Fatalf("Store failed: %v", err)
		}
		storeDuration := time.Since(start)

		// Get operation
		start = time.Now()
		reader, err := client.Get(ctx, key)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}

		// Read data
		_, err = io.ReadAll(reader)
		if err != nil {
			b.Fatalf("Read failed: %v", err)
		}
		getDuration := time.Since(start)

		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}

		// Report metrics
		b.ReportMetric(float64(storeDuration.Milliseconds()), "store_ms")
		b.ReportMetric(float64(getDuration.Milliseconds()), "get_ms")
		b.ReportMetric(float64(len(testData))/1024, "file_size_kb")
	}
}

// BenchmarkConcurrentOperations measures performance under concurrent load
func BenchmarkConcurrentOperations(b *testing.B) {
	// Create test network with multiple nodes
	config := utils.NetworkConfig{
		BootstrapNodes: []utils.NodeConfig{
			{Name: "bootstrap", ListenAddr: ":7005"},
		},
		ClientNodes: []utils.NodeConfig{
			{Name: "client1", ListenAddr: ":7006", BootstrapNodes: []string{":7005"}},
			{Name: "client2", ListenAddr: ":7007", BootstrapNodes: []string{":7005"}},
			{Name: "client3", ListenAddr: ":7008", BootstrapNodes: []string{":7005"}},
		},
	}

	manager := utils.CreateTestNetwork(&testing.T{}, config)
	defer manager.StopAll()

	// Get servers
	bootstrap := manager.GetServer("bootstrap")
	client1 := manager.GetServer("client1")
	client2 := manager.GetServer("client2")
	client3 := manager.GetServer("client3")

	// Create test data
	dataGen := utils.NewTestDataGenerator()
	testData := dataGen.GenerateSmallFile()

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b.ResetTimer()
	b.ReportAllocs()

	// Run concurrent operations
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			key := "concurrent_file_" + string(rune(counter))
			counter++

			// Store from bootstrap
			err := bootstrap.Store(ctx, key, bytes.NewReader(testData))
			if err != nil {
				b.Fatalf("Store failed: %v", err)
			}

			// Get from random client
			var reader io.Reader
			switch counter % 3 {
			case 0:
				reader, err = client1.Get(ctx, key)
			case 1:
				reader, err = client2.Get(ctx, key)
			case 2:
				reader, err = client3.Get(ctx, key)
			}

			if err != nil {
				b.Fatalf("Get failed: %v", err)
			}

			// Read data
			_, err = io.ReadAll(reader)
			if err != nil {
				b.Fatalf("Read failed: %v", err)
			}

			if closer, ok := reader.(io.Closer); ok {
				closer.Close()
			}
		}
	})
}

// BenchmarkResourceLimits measures performance under resource constraints
func BenchmarkResourceLimits(b *testing.B) {
	// Create test network
	config := utils.NetworkConfig{
		BootstrapNodes: []utils.NodeConfig{
			{Name: "bootstrap", ListenAddr: ":7009"},
		},
		ClientNodes: []utils.NodeConfig{
			{Name: "client", ListenAddr: ":7010", BootstrapNodes: []string{":7009"}},
		},
	}

	manager := utils.CreateTestNetwork(&testing.T{}, config)
	defer manager.StopAll()

	// Get servers
	bootstrap := manager.GetServer("bootstrap")
	client := manager.GetServer("client")

	// Create test data
	dataGen := utils.NewTestDataGenerator()
	testData := dataGen.GenerateSmallFile()

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b.ResetTimer()
	b.ReportAllocs()

	// Test with many small operations to stress resource limits
	for i := 0; i < b.N; i++ {
		key := "resource_test_file_" + string(rune(i))

		// Store operation
		err := bootstrap.Store(ctx, key, bytes.NewReader(testData))
		if err != nil {
			b.Fatalf("Store failed: %v", err)
		}

		// Get operation
		reader, err := client.Get(ctx, key)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}

		// Read data
		_, err = io.ReadAll(reader)
		if err != nil {
			b.Fatalf("Read failed: %v", err)
		}

		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	}
}

// TestPerformanceMetrics tests various performance metrics
func TestPerformanceMetrics(t *testing.T) {
	// Create test network
	config := utils.NetworkConfig{
		BootstrapNodes: []utils.NodeConfig{
			{Name: "bootstrap", ListenAddr: ":7011"},
		},
		ClientNodes: []utils.NodeConfig{
			{Name: "client", ListenAddr: ":7012", BootstrapNodes: []string{":7011"}},
		},
	}

	manager := utils.CreateTestNetwork(t, config)
	defer manager.StopAll()

	// Get servers
	bootstrap := manager.GetServer("bootstrap")
	client := manager.GetServer("client")

	// Create test data
	dataGen := utils.NewTestDataGenerator()
	testData := dataGen.GenerateSmallFile()

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test store performance
	t.Run("Store Performance", func(t *testing.T) {
		start := time.Now()
		err := bootstrap.Store(ctx, "perf_test_file", bytes.NewReader(testData))
		duration := time.Since(start)

		utils.AssertNoError(t, err, "Store operation failed")
		slog.Info("Store performance", "duration", duration, "size", len(testData))
	})

	// Test get performance
	t.Run("Get Performance", func(t *testing.T) {
		start := time.Now()
		reader, err := client.Get(ctx, "perf_test_file")
		duration := time.Since(start)

		utils.AssertNoError(t, err, "Get operation failed")
		defer func() {
			if closer, ok := reader.(io.Closer); ok {
				closer.Close()
			}
		}()

		// Read data
		readStart := time.Now()
		data, err := io.ReadAll(reader)
		readDuration := time.Since(readStart)

		utils.AssertNoError(t, err, "Read operation failed")
		utils.AssertEqual(t, len(testData), len(data), "Data size mismatch")

		slog.Info("Get performance", 
			"get_duration", duration, 
			"read_duration", readDuration, 
			"total_duration", duration+readDuration,
			"size", len(data))
	})

	// Test throughput
	t.Run("Throughput Test", func(t *testing.T) {
		const numFiles = 10
		start := time.Now()

		for i := 0; i < numFiles; i++ {
			key := "throughput_file_" + string(rune(i))
			err := bootstrap.Store(ctx, key, bytes.NewReader(testData))
			utils.AssertNoError(t, err, "Store operation failed")
		}

		totalDuration := time.Since(start)
		throughput := float64(numFiles) / totalDuration.Seconds()

		slog.Info("Throughput test", 
			"files", numFiles, 
			"duration", totalDuration, 
			"throughput_files_per_second", throughput)
	})
}
