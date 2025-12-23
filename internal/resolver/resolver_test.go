package resolver

import (
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	r := New(100 * time.Millisecond)
	if r == nil {
		t.Fatal("expected non-nil resolver")
	}
	if r.timeout != 100*time.Millisecond {
		t.Errorf("expected timeout 100ms, got %v", r.timeout)
	}
	if r.cache == nil {
		t.Error("expected cache to be initialized")
	}
	if r.noCache {
		t.Error("expected noCache to be false by default")
	}
}

func TestSetNoCache(t *testing.T) {
	r := New(100 * time.Millisecond)

	r.SetNoCache(true)
	if !r.noCache {
		t.Error("expected noCache to be true")
	}

	r.SetNoCache(false)
	if r.noCache {
		t.Error("expected noCache to be false")
	}
}

func TestResolveAddr_InvalidIP(t *testing.T) {
	r := New(100 * time.Millisecond)

	// invalid ip should return as-is
	result := r.ResolveAddr("not-an-ip")
	if result != "not-an-ip" {
		t.Errorf("expected 'not-an-ip', got %q", result)
	}

	// empty string should return as-is
	result = r.ResolveAddr("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestResolveAddr_Caching(t *testing.T) {
	r := New(100 * time.Millisecond)

	// first call should cache
	addr := "127.0.0.1"
	result1 := r.ResolveAddr(addr)

	// verify cache is populated
	if r.GetCacheSize() != 1 {
		t.Errorf("expected cache size 1, got %d", r.GetCacheSize())
	}

	// second call should use cache
	result2 := r.ResolveAddr(addr)
	if result1 != result2 {
		t.Errorf("expected same result from cache, got %q and %q", result1, result2)
	}
}

func TestResolveAddr_NoCacheMode(t *testing.T) {
	r := New(100 * time.Millisecond)
	r.SetNoCache(true)

	addr := "127.0.0.1"
	r.ResolveAddr(addr)

	// cache should remain empty when noCache is enabled
	if r.GetCacheSize() != 0 {
		t.Errorf("expected cache size 0 with noCache, got %d", r.GetCacheSize())
	}
}

func TestResolvePort_Zero(t *testing.T) {
	r := New(100 * time.Millisecond)

	result := r.ResolvePort(0, "tcp")
	if result != "0" {
		t.Errorf("expected '0' for port 0, got %q", result)
	}
}

func TestResolvePort_WellKnown(t *testing.T) {
	r := New(100 * time.Millisecond)

	tests := []struct {
		port     int
		proto    string
		expected string
	}{
		{80, "tcp", "http"},
		{443, "tcp", "https"},
		{22, "tcp", "ssh"},
		{53, "udp", "domain"},
		{5432, "tcp", "postgresql"},
	}

	for _, tt := range tests {
		result := r.ResolvePort(tt.port, tt.proto)
		if result != tt.expected {
			t.Errorf("ResolvePort(%d, %q) = %q, want %q", tt.port, tt.proto, result, tt.expected)
		}
	}
}

func TestResolvePort_Caching(t *testing.T) {
	r := New(100 * time.Millisecond)

	r.ResolvePort(80, "tcp")
	r.ResolvePort(443, "tcp")

	if r.GetCacheSize() != 2 {
		t.Errorf("expected cache size 2, got %d", r.GetCacheSize())
	}

	// same port/proto should not add new entry
	r.ResolvePort(80, "tcp")
	if r.GetCacheSize() != 2 {
		t.Errorf("expected cache size still 2, got %d", r.GetCacheSize())
	}
}

func TestResolveAddrPort(t *testing.T) {
	r := New(100 * time.Millisecond)

	addr, port := r.ResolveAddrPort("127.0.0.1", 80, "tcp")

	if addr == "" {
		t.Error("expected non-empty address")
	}
	if port != "http" {
		t.Errorf("expected port 'http', got %q", port)
	}
}

func TestClearCache(t *testing.T) {
	r := New(100 * time.Millisecond)

	r.ResolveAddr("127.0.0.1")
	r.ResolvePort(80, "tcp")

	if r.GetCacheSize() == 0 {
		t.Error("expected non-empty cache before clear")
	}

	r.ClearCache()

	if r.GetCacheSize() != 0 {
		t.Errorf("expected empty cache after clear, got %d", r.GetCacheSize())
	}
}

func TestGetCacheSize(t *testing.T) {
	r := New(100 * time.Millisecond)

	if r.GetCacheSize() != 0 {
		t.Errorf("expected initial cache size 0, got %d", r.GetCacheSize())
	}

	r.ResolveAddr("127.0.0.1")
	if r.GetCacheSize() != 1 {
		t.Errorf("expected cache size 1, got %d", r.GetCacheSize())
	}
}

func TestGetServiceName(t *testing.T) {
	tests := []struct {
		port     int
		proto    string
		expected string
	}{
		{80, "tcp", "http"},
		{443, "tcp", "https"},
		{22, "tcp", "ssh"},
		{53, "tcp", "domain"},
		{53, "udp", "domain"},
		{12345, "tcp", ""},
		{0, "tcp", ""},
	}

	for _, tt := range tests {
		result := getServiceName(tt.port, tt.proto)
		if result != tt.expected {
			t.Errorf("getServiceName(%d, %q) = %q, want %q", tt.port, tt.proto, result, tt.expected)
		}
	}
}

func TestResolveAddrsParallel(t *testing.T) {
	r := New(100 * time.Millisecond)

	addrs := []string{
		"127.0.0.1",
		"127.0.0.2",
		"127.0.0.3",
		"", // should be skipped
		"*", // should be skipped
	}

	r.ResolveAddrsParallel(addrs)

	// should have cached 3 addresses (excluding empty and *)
	if r.GetCacheSize() != 3 {
		t.Errorf("expected cache size 3, got %d", r.GetCacheSize())
	}
}

func TestResolveAddrsParallel_Dedupe(t *testing.T) {
	r := New(100 * time.Millisecond)

	addrs := []string{
		"127.0.0.1",
		"127.0.0.1",
		"127.0.0.1",
		"127.0.0.2",
	}

	r.ResolveAddrsParallel(addrs)

	// should have cached 2 unique addresses
	if r.GetCacheSize() != 2 {
		t.Errorf("expected cache size 2, got %d", r.GetCacheSize())
	}
}

func TestResolveAddrsParallel_SkipsCached(t *testing.T) {
	r := New(100 * time.Millisecond)

	// pre-cache one address
	r.ResolveAddr("127.0.0.1")

	addrs := []string{
		"127.0.0.1", // already cached
		"127.0.0.2", // not cached
	}

	initialSize := r.GetCacheSize()
	r.ResolveAddrsParallel(addrs)

	// should have added 1 more
	if r.GetCacheSize() != initialSize+1 {
		t.Errorf("expected cache size %d, got %d", initialSize+1, r.GetCacheSize())
	}
}

func TestResolveAddrsParallel_Empty(t *testing.T) {
	r := New(100 * time.Millisecond)

	// should not panic with empty input
	r.ResolveAddrsParallel([]string{})
	r.ResolveAddrsParallel(nil)

	if r.GetCacheSize() != 0 {
		t.Errorf("expected cache size 0, got %d", r.GetCacheSize())
	}
}

func TestGlobalResolver(t *testing.T) {
	// reset global resolver
	globalResolver = nil

	r := GetGlobalResolver()
	if r == nil {
		t.Fatal("expected non-nil global resolver")
	}

	// should return same instance
	r2 := GetGlobalResolver()
	if r != r2 {
		t.Error("expected same global resolver instance")
	}
}

func TestSetGlobalResolver(t *testing.T) {
	SetGlobalResolver(ResolverOptions{
		Timeout: 500 * time.Millisecond,
		NoCache: true,
	})

	r := GetGlobalResolver()
	if r.timeout != 500*time.Millisecond {
		t.Errorf("expected timeout 500ms, got %v", r.timeout)
	}
	if !r.noCache {
		t.Error("expected noCache to be true")
	}

	// reset for other tests
	globalResolver = nil
}

func TestSetGlobalResolver_DefaultTimeout(t *testing.T) {
	SetGlobalResolver(ResolverOptions{
		Timeout: 0, // should use default
	})

	r := GetGlobalResolver()
	if r.timeout != 200*time.Millisecond {
		t.Errorf("expected default timeout 200ms, got %v", r.timeout)
	}

	// reset for other tests
	globalResolver = nil
}

func TestGlobalConvenienceFunctions(t *testing.T) {
	globalResolver = nil

	// test global ResolveAddr
	result := ResolveAddr("127.0.0.1")
	if result == "" {
		t.Error("expected non-empty result from global ResolveAddr")
	}

	// test global ResolvePort
	port := ResolvePort(80, "tcp")
	if port != "http" {
		t.Errorf("expected 'http', got %q", port)
	}

	// test global ResolveAddrPort
	addr, portStr := ResolveAddrPort("127.0.0.1", 443, "tcp")
	if addr == "" {
		t.Error("expected non-empty address")
	}
	if portStr != "https" {
		t.Errorf("expected 'https', got %q", portStr)
	}

	// test global SetNoCache
	SetNoCache(true)
	if !GetGlobalResolver().noCache {
		t.Error("expected global noCache to be true")
	}

	// reset
	globalResolver = nil
}

func TestConcurrentAccess(t *testing.T) {
	r := New(100 * time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			addr := "127.0.0.1"
			r.ResolveAddr(addr)
			r.ResolvePort(80+n%10, "tcp")
			r.GetCacheSize()
		}(i)
	}

	wg.Wait()

	// should not panic and cache should have entries
	if r.GetCacheSize() == 0 {
		t.Error("expected non-empty cache after concurrent access")
	}
}

func TestResolveAddr_TrailingDot(t *testing.T) {
	// this test verifies the trailing dot removal logic
	// by checking the internal logic works correctly
	r := New(100 * time.Millisecond)

	// localhost should resolve and have trailing dot removed
	result := r.ResolveAddr("127.0.0.1")
	if len(result) > 0 && result[len(result)-1] == '.' {
		t.Error("expected trailing dot to be removed")
	}
}

