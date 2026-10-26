package rbac

import (
	"testing"
	"time"
)

func TestMemoryCache_ExpiresAfterTTL(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c := &memoryCache{ttl: time.Minute, now: func() time.Time { return now }, items: map[int64]cacheEntry{}}

	c.Set(1, NewPermissionSet("user:read"))
	if set, ok := c.Get(1); !ok || !set.Has("user:read") {
		t.Fatal("expected cached permissions")
	}

	now = now.Add(2 * time.Minute)
	if _, ok := c.Get(1); ok {
		t.Fatal("entry should expire after TTL")
	}
}

func TestMemoryCache_Invalidate(t *testing.T) {
	c := NewMemoryCache(time.Minute)
	c.Set(1, NewPermissionSet("a:b"))
	c.Set(2, NewPermissionSet("a:b"))

	c.Invalidate(1)
	if _, ok := c.Get(1); ok {
		t.Fatal("user 1 should be invalidated")
	}
	if _, ok := c.Get(2); !ok {
		t.Fatal("user 2 should still be cached")
	}

	c.InvalidateAll()
	if _, ok := c.Get(2); ok {
		t.Fatal("all entries should be invalidated")
	}
}

func TestMemoryCache_DisabledWithZeroTTL(t *testing.T) {
	c := NewMemoryCache(0)
	c.Set(1, NewPermissionSet("a:b"))
	if _, ok := c.Get(1); ok {
		t.Fatal("cache with zero TTL should not store entries")
	}
}
