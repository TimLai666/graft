package textmetrics

import "testing"

func key(text string) widthKey {
	return widthKey{family: "F", size: 14, text: text}
}

func TestWidthCachePutGet(t *testing.T) {
	c := newWidthCache(4)
	c.put(key("a"), 10)
	if w, ok := c.get(key("a")); !ok || w != 10 {
		t.Errorf("get(a) = (%v, %v), want (10, true)", w, ok)
	}
	if _, ok := c.get(key("missing")); ok {
		t.Error("get(missing) = hit, want miss")
	}
}

func TestWidthCacheUpdateExisting(t *testing.T) {
	c := newWidthCache(4)
	c.put(key("a"), 10)
	c.put(key("a"), 20)
	if w, _ := c.get(key("a")); w != 20 {
		t.Errorf("updated value = %v, want 20", w)
	}
	if c.len() != 1 {
		t.Errorf("len = %d, want 1", c.len())
	}
}

func TestWidthCacheEvictsLeastRecentlyUsed(t *testing.T) {
	c := newWidthCache(2)
	c.put(key("a"), 1)
	c.put(key("b"), 2)
	// Touch "a" so "b" becomes the LRU entry.
	if _, ok := c.get(key("a")); !ok {
		t.Fatal("get(a) missed")
	}
	c.put(key("c"), 3)

	if c.len() != 2 {
		t.Fatalf("len = %d, want 2", c.len())
	}
	if _, ok := c.get(key("b")); ok {
		t.Error("b survived eviction, want it evicted as LRU")
	}
	if _, ok := c.get(key("a")); !ok {
		t.Error("a evicted, want it retained (recently used)")
	}
	if _, ok := c.get(key("c")); !ok {
		t.Error("c evicted, want it retained (just inserted)")
	}
}

func TestWidthCacheClear(t *testing.T) {
	c := newWidthCache(4)
	c.put(key("a"), 1)
	c.put(key("b"), 2)
	c.clear()
	if c.len() != 0 {
		t.Errorf("len after clear = %d, want 0", c.len())
	}
	if _, ok := c.get(key("a")); ok {
		t.Error("entry survived clear")
	}
}
