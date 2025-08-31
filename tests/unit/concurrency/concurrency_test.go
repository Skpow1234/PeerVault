package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrencySafety verifies basic concurrency safety patterns
func TestConcurrencySafety(t *testing.T) {
	// Test RWMutex behavior
	var mu sync.RWMutex
	var data map[string]int = make(map[string]int)

	// Test concurrent reads
	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer wg.Done()
			mu.RLock()
			_ = data["test"]
			mu.RUnlock()
		}(i)
	}

	wg.Wait()
	t.Log("Concurrent reads completed safely")

	// Test concurrent writes
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer wg.Done()
			mu.Lock()
			data[fmt.Sprintf("key%d", id)] = id
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	t.Log("Concurrent writes completed safely")
}

// TestRaceConditionPrevention verifies that race conditions are prevented
func TestRaceConditionPrevention(t *testing.T) {
	var mu sync.RWMutex
	var counter int

	// Simulate concurrent access patterns similar to our peer management
	var wg sync.WaitGroup
	wg.Add(20)

	// Mix of reads and writes
	for i := 0; i < 10; i++ {
		// Readers
		go func() {
			defer wg.Done()
			mu.RLock()
			_ = counter
			time.Sleep(1 * time.Millisecond)
			mu.RUnlock()
		}()

		// Writers
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			time.Sleep(1 * time.Millisecond)
			mu.Unlock()
		}()
	}

	wg.Wait()
	t.Logf("Final counter value: %d (should be 10)", counter)

	if counter != 10 {
		t.Errorf("Expected counter to be 10, got %d", counter)
	}
}

// TestCopyUnderLock verifies the pattern of copying data under lock
func TestCopyUnderLock(t *testing.T) {
	var mu sync.RWMutex
	source := map[string]int{"a": 1, "b": 2, "c": 3}

	// Copy under read lock (similar to our peer list copying)
	mu.RLock()
	copy := make(map[string]int, len(source))
	for k, v := range source {
		copy[k] = v
	}
	mu.RUnlock()

	// Verify copy is correct
	if len(copy) != 3 {
		t.Errorf("Expected copy to have 3 elements, got %d", len(copy))
	}

	t.Log("Data copying under lock completed successfully")
}
