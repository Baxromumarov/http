package http_go

import (
	"reflect"
	"testing"
)

func TestHeader_Set(t *testing.T) {
	h := make(Header)
	h.Set("Content-Type", "application/json")

	if h["Content-Type"][0] != "application/json" {
		t.Errorf("Expected Content-Type=application/json, got %s", h["Content-Type"])
	}

	// Test overwrite
	h.Set("Content-Type", "text/plain")
	if h["Content-Type"][0] != "text/plain" {
		t.Errorf("Expected Content-Type to be overwritten with text/plain, got %s", h["Content-Type"])
	}
}

func TestHeader_Add(t *testing.T) {
	h := make(Header)
	h.Add("Accept", "application/json")

	if h["Accept"][0] != "application/json" {
		t.Errorf("Expected first Accept value to be application/json, got %s", h["Accept"])
	}

	// Test adding another value to the same key
	h.Add("Accept", "text/plain")
	if len(h["Accept"]) != 2 || h["Accept"][1] != "text/plain" {
		t.Errorf("Expected two Accept values, got %v", h["Accept"])
	}
}

func TestHeader_Get(t *testing.T) {
	h := Header{
		"Content-Type": {"application/json"},
	}

	// Test existing key
	if val := h.Get("Content-Type"); val != "application/json" {
		t.Errorf("Expected application/json, got %s", val)
	}

	// Test non-existent key
	if val := h.Get("X-Non-Existent"); val != "" {
		t.Errorf("Expected empty string for non-existent key, got %s", val)
	}
}

func TestHeader_Delete(t *testing.T) {
	h := Header{
		"X-Test-Header": {"test-value"},
	}

	h.Delete("X-Test-Header")

	if _, exists := h["X-Test-Header"]; exists {
		t.Error("Expected X-Test-Header to be deleted")
	}

	// Test deleting non-existent key (should not panic)
	h.Delete("Non-Existent-Key")
}

func TestHeader_Values(t *testing.T) {
	h := Header{
		"Accept": {"application/json", "text/plain"},
	}

	values := h.Values("Accept")
	expected := []string{"application/json", "text/plain"}

	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Expected %v, got %v", expected, values)
	}

	// Test non-existent key
	values = h.Values("X-Non-Existent")
	if values != nil {
		t.Errorf("Expected nil for non-existent key, got %v", values)
	}
}

func TestHeader_Len(t *testing.T) {
	h := Header{
		"Content-Type": {"application/json"},
		"Accept":       {"application/json", "text/plain"},
	}

	if h.Len() != 2 {
		t.Errorf("Expected length 2, got %d", h.Len())
	}

	// Test empty header
	h = Header{}
	if h.Len() != 0 {
		t.Errorf("Expected length 0, got %d", h.Len())
	}
}

func TestHeader_Keys(t *testing.T) {
	h := Header{
		"Content-Type": {"application/json"},
		"Accept":       {"application/json"},
		"User-Agent":   {"test-agent"},
	}

	keys := h.Keys()
	if len(keys) != 3 {
		t.Fatalf("Expected 3 keys, got %d", len(keys))
	}

	// Convert to map for easier checking
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	expectedKeys := []string{"Content-Type", "Accept", "User-Agent"}
	for _, k := range expectedKeys {
		if !keyMap[k] {
			t.Errorf("Expected key %s not found in %v", k, keys)
		}
	}
}

func TestHeader_Clone(t *testing.T) {
	h := Header{
		"Content-Type": {"application/json"},
		"Accept":       {"application/json", "text/plain"},
	}

	clone := h.Clone()

	// Check if the clone has the same content
	if !reflect.DeepEqual(h, clone) {
		t.Errorf("Clone does not match original. Got %v, want %v", clone, h)
	}

	// Modify the clone and ensure original is not affected
	clone.Set("X-Test", "value")
	if _, exists := h["X-Test"]; exists {
		t.Error("Modifying clone affected the original header")
	}
}

func TestHeader_Exists(t *testing.T) {
	h := Header{
		"Content-Type": {"application/json"},
	}

	// Test existing key
	if !h.Exists("Content-Type") {
		t.Error("Expected Content-Type to exist")
	}

	// Test non-existent key
	if h.Exists("X-Non-Existent") {
		t.Error("Expected X-Non-Existent to not exist")
	}

	// Test with empty value
	h["Empty-Value"] = nil
	if h.Exists("Empty-Value") {
		t.Error("Expected Empty-Value with nil slice to not exist")
	}
}

func TestHeader_ConcurrentAccess(t *testing.T) {
	h := make(Header)
	done := make(chan bool)

	// Concurrent writes
	go func() {
		for i := 0; i < 1000; i++ {
			h.Set("Test", "value")
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 1000; i++ {
			_ = h.Get("Test")
		}
		done <- true
	}()

	// Wait for goroutines to complete
	<-done
	<-done
}
