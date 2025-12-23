//go:build linux

package collector

import (
	"testing"
	"time"
)

func TestGetConnections(t *testing.T) {
	// integration test to verify /proc parsing works
	conns, err := GetConnections()
	if err != nil {
		t.Fatalf("GetConnections() returned an error: %v", err)
	}

	// connections are dynamic, so just verify function succeeded
	t.Logf("Successfully got %d connections", len(conns))
}

func TestGetConnectionsPerformance(t *testing.T) {
	// measures performance to catch regressions
	// run with: go test -v -run TestGetConnectionsPerformance

	const maxDuration = 500 * time.Millisecond
	const iterations = 5

	// warm up caches first
	_, err := GetConnections()
	if err != nil {
		t.Fatalf("warmup failed: %v", err)
	}

	var total time.Duration
	var maxSeen time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		conns, err := GetConnections()
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("iteration %d failed: %v", i, err)
		}

		total += elapsed
		if elapsed > maxSeen {
			maxSeen = elapsed
		}

		t.Logf("iteration %d: %v (%d connections)", i+1, elapsed, len(conns))
	}

	avg := total / time.Duration(iterations)
	t.Logf("average: %v, max: %v", avg, maxSeen)

	if maxSeen > maxDuration {
		t.Errorf("slowest iteration took %v, expected < %v", maxSeen, maxDuration)
	}
}

func TestGetConnectionsColdCache(t *testing.T) {
	// tests performance with cold user cache
	// this simulates first run or after cache invalidation

	const maxDuration = 2 * time.Second

	clearUserCache()

	start := time.Now()
	conns, err := GetConnections()
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("GetConnections() failed: %v", err)
	}

	t.Logf("cold cache: %v (%d connections, %d cached users after)",
		elapsed, len(conns), userCacheSize())

	if elapsed > maxDuration {
		t.Errorf("cold cache took %v, expected < %v", elapsed, maxDuration)
	}
}

func BenchmarkGetConnections(b *testing.B) {
	// warm cache benchmark - measures typical runtime
	// run with: go test -bench=BenchmarkGetConnections -benchtime=5s

	// warm up
	_, _ = GetConnections()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetConnections()
	}
}

func BenchmarkGetConnectionsColdCache(b *testing.B) {
	// cold cache benchmark - measures worst-case with cache cleared each iteration
	// run with: go test -bench=BenchmarkGetConnectionsColdCache -benchtime=10s

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clearUserCache()
		_, _ = GetConnections()
	}
}

func BenchmarkBuildInodeMap(b *testing.B) {
	// benchmarks just the inode map building (most expensive part)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = buildInodeToProcessMap()
	}
}