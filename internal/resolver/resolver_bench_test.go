package resolver

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkResolveAddr_CacheHit(b *testing.B) {
	r := New(100 * time.Millisecond)
	addr := "127.0.0.1"

	// pre-populate cache
	r.ResolveAddr(addr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ResolveAddr(addr)
	}
}

func BenchmarkResolveAddr_CacheMiss(b *testing.B) {
	r := New(10 * time.Millisecond) // short timeout for faster benchmarks

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// use different addresses to avoid cache hits
		addr := fmt.Sprintf("127.0.0.%d", i%256)
		r.ClearCache() // clear cache to force miss
		r.ResolveAddr(addr)
	}
}

func BenchmarkResolveAddr_NoCache(b *testing.B) {
	r := New(10 * time.Millisecond)
	r.SetNoCache(true)
	addr := "127.0.0.1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ResolveAddr(addr)
	}
}

func BenchmarkResolvePort_CacheHit(b *testing.B) {
	r := New(100 * time.Millisecond)

	// pre-populate cache
	r.ResolvePort(80, "tcp")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ResolvePort(80, "tcp")
	}
}

func BenchmarkResolvePort_WellKnown(b *testing.B) {
	r := New(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ClearCache()
		r.ResolvePort(443, "tcp")
	}
}

func BenchmarkGetServiceName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getServiceName(80, "tcp")
	}
}

func BenchmarkGetServiceName_NotFound(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getServiceName(12345, "tcp")
	}
}

func BenchmarkResolveAddrsParallel_10(b *testing.B) {
	benchmarkResolveAddrsParallel(b, 10)
}

func BenchmarkResolveAddrsParallel_100(b *testing.B) {
	benchmarkResolveAddrsParallel(b, 100)
}

func BenchmarkResolveAddrsParallel_1000(b *testing.B) {
	benchmarkResolveAddrsParallel(b, 1000)
}

func benchmarkResolveAddrsParallel(b *testing.B, count int) {
	addrs := make([]string, count)
	for i := 0; i < count; i++ {
		addrs[i] = fmt.Sprintf("127.0.%d.%d", i/256, i%256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := New(10 * time.Millisecond)
		r.ResolveAddrsParallel(addrs)
	}
}

func BenchmarkConcurrentResolveAddr(b *testing.B) {
	r := New(100 * time.Millisecond)
	addr := "127.0.0.1"

	// pre-populate cache
	r.ResolveAddr(addr)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.ResolveAddr(addr)
		}
	})
}

func BenchmarkConcurrentResolvePort(b *testing.B) {
	r := New(100 * time.Millisecond)

	// pre-populate cache
	r.ResolvePort(80, "tcp")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.ResolvePort(80, "tcp")
		}
	})
}

func BenchmarkGetCacheSize(b *testing.B) {
	r := New(100 * time.Millisecond)

	// populate with some entries
	for i := 0; i < 100; i++ {
		r.ResolvePort(i+1, "tcp")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.GetCacheSize()
	}
}

func BenchmarkClearCache(b *testing.B) {
	r := New(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// populate and clear
		for j := 0; j < 10; j++ {
			r.ResolvePort(j+1, "tcp")
		}
		r.ClearCache()
	}
}

