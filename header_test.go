package http

import (
	"testing"
)

func TestHeader_Set(t *testing.T) {
	h := make(Header)

	// Test setting a single value
	h.Set("Content-Type", "application/json")
	if h.Get("Content-Type") != "application/json" {
		t.Errorf("Expected 'application/json', got '%s'", h.Get("Content-Type"))
	}

	// Test overwriting existing value
	h.Set("Content-Type", "text/html")
	if h.Get("Content-Type") != "text/html" {
		t.Errorf("Expected 'text/html', got '%s'", h.Get("Content-Type"))
	}

	// Test setting multiple headers
	h.Set("Accept", "application/json")
	h.Set("User-Agent", "test-client")

	if h.Get("Accept") != "application/json" {
		t.Errorf("Expected 'application/json', got '%s'", h.Get("Accept"))
	}
	if h.Get("User-Agent") != "test-client" {
		t.Errorf("Expected 'test-client', got '%s'", h.Get("User-Agent"))
	}
}

func TestHeader_Add(t *testing.T) {
	h := make(Header)

	// Test adding first value
	h.Add("Accept", "application/json")
	if h.Get("Accept") != "application/json" {
		t.Errorf("Expected 'application/json', got '%s'", h.Get("Accept"))
	}

	// Test adding multiple values
	h.Add("Accept", "text/html")
	h.Add("Accept", "text/plain")

	values := h.Values("Accept")
	expected := []string{"application/json", "text/html", "text/plain"}

	if len(values) != len(expected) {
		t.Errorf("Expected %d values, got %d", len(expected), len(values))
	}

	for i, v := range expected {
		if values[i] != v {
			t.Errorf("Expected '%s' at index %d, got '%s'", v, i, values[i])
		}
	}
}

func TestHeader_Get(t *testing.T) {
	h := make(Header)

	// Test getting non-existent header
	if h.Get("NonExistent") != "" {
		t.Errorf("Expected empty string for non-existent header, got '%s'", h.Get("NonExistent"))
	}

	// Test getting existing header
	h.Set("Content-Type", "application/json")
	if h.Get("Content-Type") != "application/json" {
		t.Errorf("Expected 'application/json', got '%s'", h.Get("Content-Type"))
	}

	// Test getting first value when multiple exist
	h.Add("Accept", "application/json")
	h.Add("Accept", "text/html")
	if h.Get("Accept") != "application/json" {
		t.Errorf("Expected first value 'application/json', got '%s'", h.Get("Accept"))
	}
}

func TestHeader_Delete(t *testing.T) {
	h := make(Header)

	// Test deleting non-existent header
	h.Delete("NonExistent") // Should not panic

	// Test deleting existing header
	h.Set("Content-Type", "application/json")
	h.Delete("Content-Type")

	if h.Get("Content-Type") != "" {
		t.Errorf("Expected empty string after deletion, got '%s'", h.Get("Content-Type"))
	}

	// Test deleting header with multiple values
	h.Add("Accept", "application/json")
	h.Add("Accept", "text/html")
	h.Delete("Accept")

	if h.Get("Accept") != "" {
		t.Errorf("Expected empty string after deletion, got '%s'", h.Get("Accept"))
	}
}

func TestHeader_Values(t *testing.T) {
	h := make(Header)

	// Test getting values for non-existent header
	values := h.Values("NonExistent")
	if len(values) != 0 {
		t.Errorf("Expected empty slice for non-existent header, got %d values", len(values))
	}

	// Test getting single value
	h.Set("Content-Type", "application/json")
	values = h.Values("Content-Type")
	if len(values) != 1 || values[0] != "application/json" {
		t.Errorf("Expected ['application/json'], got %v", values)
	}

	// Test getting multiple values
	h.Add("Accept", "application/json")
	h.Add("Accept", "text/html")
	h.Add("Accept", "text/plain")
	values = h.Values("Accept")
	expected := []string{"application/json", "text/html", "text/plain"}

	if len(values) != len(expected) {
		t.Errorf("Expected %d values, got %d", len(expected), len(values))
	}

	for i, v := range expected {
		if values[i] != v {
			t.Errorf("Expected '%s' at index %d, got '%s'", v, i, values[i])
		}
	}
}

func TestHeader_Len(t *testing.T) {
	h := make(Header)

	// Test empty header
	if h.Len() != 0 {
		t.Errorf("Expected length 0, got %d", h.Len())
	}

	// Test single header
	h.Set("Content-Type", "application/json")
	if h.Len() != 1 {
		t.Errorf("Expected length 1, got %d", h.Len())
	}

	// Test multiple headers
	h.Set("Accept", "application/json")
	h.Set("User-Agent", "test-client")
	if h.Len() != 3 {
		t.Errorf("Expected length 3, got %d", h.Len())
	}

	// Test after deletion
	h.Delete("Content-Type")
	if h.Len() != 2 {
		t.Errorf("Expected length 2 after deletion, got %d", h.Len())
	}
}

func TestHeader_Keys(t *testing.T) {
	h := make(Header)

	// Test empty header
	keys := h.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected empty keys slice, got %d keys", len(keys))
	}

	// Test with headers
	h.Set("Content-Type", "application/json")
	h.Set("Accept", "application/json")
	h.Set("User-Agent", "test-client")

	keys = h.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Check that all expected keys are present
	expectedKeys := map[string]bool{
		"Content-Type": true,
		"Accept":       true,
		"User-Agent":   true,
	}

	for _, key := range keys {
		if !expectedKeys[key] {
			t.Errorf("Unexpected key: %s", key)
		}
	}
}

func TestHeader_Clone(t *testing.T) {
	h := make(Header)
	h.Set("Content-Type", "application/json")
	h.Add("Accept", "application/json")
	h.Add("Accept", "text/html")

	cloned := h.Clone()

	// Test that cloned header has same values
	if cloned.Get("Content-Type") != h.Get("Content-Type") {
		t.Errorf("Cloned header has different Content-Type")
	}

	values := cloned.Values("Accept")
	expected := []string{"application/json", "text/html"}
	if len(values) != len(expected) {
		t.Errorf("Expected %d Accept values in clone, got %d", len(expected), len(values))
	}

	// Test that modifying original doesn't affect clone
	h.Set("Content-Type", "text/html")
	if cloned.Get("Content-Type") == "text/html" {
		t.Errorf("Modifying original header affected clone")
	}

	// Test that modifying clone doesn't affect original
	cloned.Set("X-Custom", "test")
	if h.Get("X-Custom") != "" {
		t.Errorf("Modifying clone affected original header")
	}
}

func TestHeader_Exists(t *testing.T) {
	h := make(Header)

	// Test non-existent header
	if h.Exists("NonExistent") {
		t.Errorf("Non-existent header should return false")
	}

	// Test existing header
	h.Set("Content-Type", "application/json")
	if !h.Exists("Content-Type") {
		t.Errorf("Existing header should return true")
	}

	// Test header with empty value
	h.Set("Empty-Header", "")
	if !h.Exists("Empty-Header") {
		t.Errorf("Header with empty value should return true")
	}

	// Test after deletion
	h.Delete("Content-Type")
	if h.Exists("Content-Type") {
		t.Errorf("Deleted header should return false")
	}
}

func TestHeader_CaseSensitive(t *testing.T) {
	h := make(Header)

	// Test case sensitivity
	h.Set("Content-Type", "application/json")

	// Only exact case should work
	if h.Get("Content-Type") != "application/json" {
		t.Errorf("Exact case should work")
	}
	if h.Get("content-type") == "application/json" {
		t.Errorf("Lowercase key should not work")
	}
	if h.Get("CONTENT-TYPE") == "application/json" {
		t.Errorf("Uppercase key should not work")
	}

	// Test setting with different cases
	h.Set("ACCEPT", "application/json")
	if h.Get("ACCEPT") != "application/json" {
		t.Errorf("Setting with uppercase and getting with uppercase should work")
	}
	if h.Get("accept") == "application/json" {
		t.Errorf("Getting with lowercase should not work")
	}
}

func TestHeader_EdgeCases(t *testing.T) {
	h := make(Header)

	// Test empty key
	h.Set("", "value")
	if h.Get("") != "value" {
		t.Errorf("Empty key should work")
	}

	// Test empty value
	h.Set("Empty-Value", "")
	if h.Get("Empty-Value") != "" {
		t.Errorf("Empty value should work")
	}

	// Test very long key and value
	longKey := string(make([]byte, 1000))
	longValue := string(make([]byte, 1000))
	h.Set(longKey, longValue)
	if h.Get(longKey) != longValue {
		t.Errorf("Long key and value should work")
	}

	// Test special characters in key and value
	h.Set("Special-Key!@#$%", "Special-Value!@#$%")
	if h.Get("Special-Key!@#$%") != "Special-Value!@#$%" {
		t.Errorf("Special characters should work")
	}
}
