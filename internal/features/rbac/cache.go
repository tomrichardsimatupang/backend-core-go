package rbac

import (
	"sync"
	"time"
)

type PermissionSet map[string]struct{}

func NewPermissionSet(codes ...string) PermissionSet {
	s := make(PermissionSet, len(codes))
	for _, c := range codes {
		s[c] = struct{}{}
	}
	return s
}

func (s PermissionSet) Has(code string) bool {
	_, ok := s[code]
	return ok
}

// Cache menyimpan permission per user. Implementasi in-memory hanya aman
// untuk satu instance; untuk multi-instance buat implementasi Redis.
type Cache interface {
	Get(userID int64) (PermissionSet, bool)
	Set(userID int64, perms PermissionSet)
	Invalidate(userID int64)
	InvalidateAll()
}

type cacheEntry struct {
	perms     PermissionSet
	expiresAt time.Time
}

type memoryCache struct {
	mu    sync.Mutex
	ttl   time.Duration
	now   func() time.Time
	items map[int64]cacheEntry
}

// NewMemoryCache membuat cache in-memory. ttl <= 0 berarti cache nonaktif.
func NewMemoryCache(ttl time.Duration) Cache {
	return &memoryCache{ttl: ttl, now: time.Now, items: make(map[int64]cacheEntry)}
}

func (c *memoryCache) Get(userID int64) (PermissionSet, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.items[userID]
	if !ok {
		return nil, false
	}
	if c.now().After(e.expiresAt) {
		delete(c.items, userID)
		return nil, false
	}
	return e.perms, true
}

func (c *memoryCache) Set(userID int64, perms PermissionSet) {
	if c.ttl <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[userID] = cacheEntry{perms: perms, expiresAt: c.now().Add(c.ttl)}
}

func (c *memoryCache) Invalidate(userID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, userID)
}

func (c *memoryCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[int64]cacheEntry)
}
