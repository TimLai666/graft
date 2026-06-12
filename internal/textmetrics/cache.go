package textmetrics

import (
	"container/list"
	"sync"
)

// widthKey identifies one memoized measurement.
type widthKey struct {
	family string
	size   float32
	text   string
}

// widthEntry is the value stored in the LRU list.
type widthEntry struct {
	key   widthKey
	width float32
}

// widthCache is a mutex-guarded LRU cache of measured text widths.
//
// Text measurement happens on every layout pass, so lookups must be cheap
// and the cache must stay bounded for long-running apps with churning text.
type widthCache struct {
	mu  sync.Mutex
	cap int
	ll  *list.List // front = most recently used; values are *widthEntry
	m   map[widthKey]*list.Element
}

// newWidthCache creates a cache bounded to capacity entries.
func newWidthCache(capacity int) *widthCache {
	return &widthCache{
		cap: capacity,
		ll:  list.New(),
		m:   make(map[widthKey]*list.Element, capacity),
	}
}

// get returns the cached width for key, marking it most recently used.
func (c *widthCache) get(key widthKey) (float32, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.m[key]
	if !ok {
		return 0, false
	}
	c.ll.MoveToFront(el)
	return el.Value.(*widthEntry).width, true
}

// put stores the width for key, evicting the least recently used entries
// when the cache exceeds its capacity.
func (c *widthCache) put(key widthKey, width float32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.m[key]; ok {
		el.Value.(*widthEntry).width = width
		c.ll.MoveToFront(el)
		return
	}
	c.m[key] = c.ll.PushFront(&widthEntry{key: key, width: width})
	for c.ll.Len() > c.cap {
		oldest := c.ll.Back()
		if oldest == nil {
			break
		}
		c.ll.Remove(oldest)
		delete(c.m, oldest.Value.(*widthEntry).key)
	}
}

// clear drops every cached entry.
func (c *widthCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ll.Init()
	c.m = make(map[widthKey]*list.Element, c.cap)
}

// len returns the current number of cached entries.
func (c *widthCache) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}
