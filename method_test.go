package http

import (
	"strings"
	"testing"
)

func TestMethod_Constants(t *testing.T) {
	// Test that all method constants are defined correctly
	expectedMethods := map[string]Method{
		"GET":    GET,
		"POST":   POST,
		"PUT":    PUT,
		"DELETE": DELETE,
	}

	for name, method := range expectedMethods {
		t.Run(name, func(t *testing.T) {
			if string(method) != name {
				t.Errorf("Expected method %s to be %s, got %s", name, name, string(method))
			}
		})
	}
}

func TestMethod_String(t *testing.T) {
	tests := []struct {
		name     string
		method   Method
		expected string
	}{
		{"GET", GET, "GET"},
		{"POST", POST, "POST"},
		{"PUT", PUT, "PUT"},
		{"DELETE", DELETE, "DELETE"},
		{"Empty method", "", ""},
		{"Custom method", "CUSTOM", "CUSTOM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.method); got != tt.expected {
				t.Errorf("Method.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMethod_Comparison(t *testing.T) {
	// Test method equality
	if GET != "GET" {
		t.Errorf("GET should equal 'GET'")
	}
	if POST != "POST" {
		t.Errorf("POST should equal 'POST'")
	}
	if GET == POST {
		t.Errorf("GET should not equal POST")
	}

	// Test method comparison
	if string(GET) >= string(POST) {
		t.Errorf("GET should be less than POST in string comparison")
	}
	if string(POST) <= string(GET) {
		t.Errorf("POST should be greater than GET in string comparison")
	}
}

func TestMethod_EdgeCases(t *testing.T) {
	// Test with very long method names
	longMethod := Method("VERY_LONG_METHOD_NAME_THAT_EXCEEDS_NORMAL_LENGTH")
	if string(longMethod) != "VERY_LONG_METHOD_NAME_THAT_EXCEEDS_NORMAL_LENGTH" {
		t.Errorf("Long method name should be preserved")
	}

	// Test with special characters
	specialMethod := Method("METHOD-WITH-HYPHENS")
	if string(specialMethod) != "METHOD-WITH-HYPHENS" {
		t.Errorf("Method with hyphens should be preserved")
	}

	// Test with numbers
	numericMethod := Method("METHOD123")
	if string(numericMethod) != "METHOD123" {
		t.Errorf("Method with numbers should be preserved")
	}

	// Test with spaces
	spaceMethod := Method("METHOD WITH SPACES")
	if string(spaceMethod) != "METHOD WITH SPACES" {
		t.Errorf("Method with spaces should be preserved")
	}
}

func TestMethod_HTTPCompliance(t *testing.T) {
	// Test that our methods comply with HTTP/1.1 specification
	httpMethods := []Method{GET, POST, PUT, DELETE}

	for _, method := range httpMethods {
		t.Run(string(method), func(t *testing.T) {
			// Test that method string representation is uppercase
			if string(method) != strings.ToUpper(string(method)) {
				t.Errorf("HTTP method should be uppercase: %s", method)
			}
		})
	}
}

func TestMethod_ZeroValue(t *testing.T) {
	var zeroMethod Method

	// Test zero value behavior
	if string(zeroMethod) != "" {
		t.Errorf("Zero value method should be empty string")
	}
}

func TestMethod_TypeConversion(t *testing.T) {
	// Test conversion from string to Method
	methodStr := "GET"
	method := Method(methodStr)
	if string(method) != methodStr {
		t.Errorf("Method conversion failed: expected %s, got %s", methodStr, string(method))
	}

	// Test conversion from Method to string
	method = GET
	methodStr = string(method)
	if methodStr != "GET" {
		t.Errorf("String conversion failed: expected GET, got %s", methodStr)
	}
}
