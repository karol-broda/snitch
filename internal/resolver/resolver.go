package resolver

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/thediveo/netdb"
)

var debugTiming = os.Getenv("SNITCH_DEBUG_TIMING") != ""

// Resolver handles DNS and service name resolution with caching and timeouts
type Resolver struct {
	timeout time.Duration
	cache   map[string]string
	mutex   sync.RWMutex
	noCache bool
}

// New creates a new resolver with the specified timeout
func New(timeout time.Duration) *Resolver {
	return &Resolver{
		timeout: timeout,
		cache:   make(map[string]string),
		noCache: false,
	}
}

// SetNoCache disables caching - each lookup will hit DNS directly
func (r *Resolver) SetNoCache(noCache bool) {
	r.noCache = noCache
}

// ResolveAddr resolves an IP address to a hostname, with caching
func (r *Resolver) ResolveAddr(addr string) string {
	// check cache first (unless caching is disabled)
	if !r.noCache {
		r.mutex.RLock()
		if cached, exists := r.cache[addr]; exists {
			r.mutex.RUnlock()
			return cached
		}
		r.mutex.RUnlock()
	}

	// parse ip to validate it
	ip := net.ParseIP(addr)
	if ip == nil {
		return addr
	}

	// perform resolution with timeout
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	names, err := net.DefaultResolver.LookupAddr(ctx, addr)

	resolved := addr
	if err == nil && len(names) > 0 {
		resolved = names[0]
		// remove trailing dot if present
		if len(resolved) > 0 && resolved[len(resolved)-1] == '.' {
			resolved = resolved[:len(resolved)-1]
		}
	}

	elapsed := time.Since(start)
	if debugTiming && elapsed > 50*time.Millisecond {
		fmt.Fprintf(os.Stderr, "[timing] slow DNS lookup: %s -> %s (%v)\n", addr, resolved, elapsed)
	}

	// cache the result (unless caching is disabled)
	if !r.noCache {
		r.mutex.Lock()
		r.cache[addr] = resolved
		r.mutex.Unlock()
	}

	return resolved
}

// ResolvePort resolves a port number to a service name
func (r *Resolver) ResolvePort(port int, proto string) string {
	if port == 0 {
		return "0"
	}

	cacheKey := strconv.Itoa(port) + "/" + proto

	// check cache first (unless caching is disabled)
	if !r.noCache {
		r.mutex.RLock()
		if cached, exists := r.cache[cacheKey]; exists {
			r.mutex.RUnlock()
			return cached
		}
		r.mutex.RUnlock()
	}

	// perform resolution with timeout
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	service, err := net.DefaultResolver.LookupPort(ctx, proto, strconv.Itoa(port))

	resolved := strconv.Itoa(port) // fallback to port number
	if err == nil && service != 0 {
		// try to get service name
		if serviceName := getServiceName(port, proto); serviceName != "" {
			resolved = serviceName
		}
	}

	// cache the result (unless caching is disabled)
	if !r.noCache {
		r.mutex.Lock()
		r.cache[cacheKey] = resolved
		r.mutex.Unlock()
	}

	return resolved
}

// ResolveAddrPort resolves both address and port
func (r *Resolver) ResolveAddrPort(addr string, port int, proto string) (string, string) {
	resolvedAddr := r.ResolveAddr(addr)
	resolvedPort := r.ResolvePort(port, proto)
	return resolvedAddr, resolvedPort
}

// ClearCache clears the resolution cache
func (r *Resolver) ClearCache() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.cache = make(map[string]string)
}

// GetCacheSize returns the number of cached entries
func (r *Resolver) GetCacheSize() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.cache)
}

// getServiceName returns well-known service names for common ports
func getServiceName(port int, proto string) string {
	// standardize the protocols to tcp/udp
	switch proto {
	case "tcp4", "tcp6":
		proto = "tcp"
	case "udp4", "udp6":
		proto = "udp"
	}

	svc := netdb.ServiceByPort(port, proto)

	if svc != nil {
		return svc.Name
	}

	return ""
}

// global resolver instance
var globalResolver *Resolver

// ResolverOptions configures the global resolver
type ResolverOptions struct {
	Timeout time.Duration
	NoCache bool
}

// SetGlobalResolver sets the global resolver instance with options
func SetGlobalResolver(opts ResolverOptions) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 200 * time.Millisecond
	}
	globalResolver = New(timeout)
	globalResolver.SetNoCache(opts.NoCache)
}

// GetGlobalResolver returns the global resolver instance
func GetGlobalResolver() *Resolver {
	if globalResolver == nil {
		globalResolver = New(200 * time.Millisecond)
	}
	return globalResolver
}

// SetNoCache configures whether the global resolver bypasses cache
func SetNoCache(noCache bool) {
	GetGlobalResolver().SetNoCache(noCache)
}

// ResolveAddr is a convenience function using the global resolver
func ResolveAddr(addr string) string {
	return GetGlobalResolver().ResolveAddr(addr)
}

// ResolvePort is a convenience function using the global resolver
func ResolvePort(port int, proto string) string {
	return GetGlobalResolver().ResolvePort(port, proto)
}

// ResolveAddrPort is a convenience function using the global resolver
func ResolveAddrPort(addr string, port int, proto string) (string, string) {
	return GetGlobalResolver().ResolveAddrPort(addr, port, proto)
}

// ResolveAddrsParallel resolves multiple addresses concurrently and caches results.
// This should be called before rendering to pre-warm the cache.
func (r *Resolver) ResolveAddrsParallel(addrs []string) {
	// dedupe and filter addresses that need resolution
	unique := make(map[string]struct{})
	for _, addr := range addrs {
		if addr == "" || addr == "*" {
			continue
		}
		// skip if already cached
		r.mutex.RLock()
		_, exists := r.cache[addr]
		r.mutex.RUnlock()
		if exists {
			continue
		}
		unique[addr] = struct{}{}
	}

	if len(unique) == 0 {
		return
	}

	var wg sync.WaitGroup
	// limit concurrency to avoid overwhelming dns
	sem := make(chan struct{}, 32)

	for addr := range unique {
		wg.Add(1)
		go func(a string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			r.ResolveAddr(a)
		}(addr)
	}

	wg.Wait()
}

// ResolveAddrsParallel is a convenience function using the global resolver
func ResolveAddrsParallel(addrs []string) {
	GetGlobalResolver().ResolveAddrsParallel(addrs)
}
