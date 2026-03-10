package core

import (
	"testing"
)

func TestNewConcurrentDict(t *testing.T) {
	dict := NewConcurrentDict()
	if dict == nil {
		t.Error("NewConcurrentDict should not return nil")
	}
	if dict.count != 0 {
		t.Errorf("Expected count 0, got %d", dict.count)
	}
}

func TestConcurrentDictSet(t *testing.T) {
	dict := NewConcurrentDict()

	existed := dict.Set("key1", "value1")
	if !existed {
		t.Error("Expected true for new key (returns true if newly added)")
	}

	existed = dict.Set("key1", "value2")
	if existed {
		t.Error("Expected false for existing key (returns false if key already existed)")
	}
}

func TestConcurrentDictGet(t *testing.T) {
	dict := NewConcurrentDict()

	dict.Set("key1", "value1")

	val, ok := dict.Get("key1")
	if !ok {
		t.Error("Expected key to exist")
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got '%v'", val)
	}

	_, ok = dict.Get("nonexistent")
	if ok {
		t.Error("Expected non-existent key to return false")
	}
}

func TestConcurrentDictSetIfAbsent(t *testing.T) {
	dict := NewConcurrentDict()

	result := dict.SetIfAbsent("key1", "value1")
	if !result {
		t.Error("Expected true for new key")
	}

	result = dict.SetIfAbsent("key1", "value2")
	if result {
		t.Error("Expected false for existing key")
	}
}

func TestConcurrentDictDelete(t *testing.T) {
	dict := NewConcurrentDict()

	dict.Set("key1", "value1")

	existed := dict.Delete("key1")
	if !existed {
		t.Error("Expected true for deleted key")
	}

	existed = dict.Delete("nonexistent")
	if existed {
		t.Error("Expected false for non-existent key")
	}
}

func TestConcurrentDictLen(t *testing.T) {
	dict := NewConcurrentDict()

	if dict.Len() != 0 {
		t.Error("Expected 0 for empty dict")
	}

	dict.Set("key1", "value1")
	dict.Set("key2", "value2")

	if dict.Len() != 2 {
		t.Errorf("Expected 2, got %d", dict.Len())
	}
}

func TestConcurrentDictClear(t *testing.T) {
	dict := NewConcurrentDict()

	dict.Set("key1", "value1")
	dict.Set("key2", "value2")

	dict.Clear()

	if dict.Len() != 0 {
		t.Errorf("Expected 0 after clear, got %d", dict.Len())
	}
}

func TestConcurrentDictKeys(t *testing.T) {
	dict := NewConcurrentDict()

	dict.Set("key1", "value1")
	dict.Set("key2", "value2")

	keys := dict.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}
}

func TestConcurrentDictForEach(t *testing.T) {
	dict := NewConcurrentDict()

	dict.Set("key1", "value1")
	dict.Set("key2", "value2")

	count := 0
	dict.ForEach(func(key string, val interface{}) bool {
		count++
		return true
	})

	if count != 2 {
		t.Errorf("Expected 2, got %d", count)
	}
}

func TestConcurrentDictForEachBreak(t *testing.T) {
	dict := NewConcurrentDict()

	dict.Set("key1", "value1")
	dict.Set("key2", "value2")

	count := 0
	dict.ForEach(func(key string, val interface{}) bool {
		count++
		return false
	})

	if count != 1 {
		t.Errorf("Expected 1 (should stop after first), got %d", count)
	}
}

func TestDictHash(t *testing.T) {
	h1 := hash("test")
	h2 := hash("test")
	if h1 != h2 {
		t.Error("Same string should produce same hash")
	}

	h3 := hash("different")
	if h1 == h3 {
		t.Error("Different strings should produce different hash")
	}
}
