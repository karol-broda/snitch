//go:build linux

package collector

// clearUserCache clears the user lookup cache for testing
func clearUserCache() {
	userCache.Lock()
	userCache.m = make(map[int]string)
	userCache.Unlock()
}

// userCacheSize returns the number of cached user entries
func userCacheSize() int {
	userCache.RLock()
	defer userCache.RUnlock()
	return len(userCache.m)
}

