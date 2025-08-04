package http

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name        string
		method      Method
		url         string
		body        []byte
		expectError bool
	}{
		{
			name:        "Valid GET request",
			method:      GET,
			url:         "http://localhost:8080/api",
			body:        nil,
			expectError: false,
		},
		{
			name:        "Valid POST request with body",
			method:      POST,
			url:         "http://localhost:8080/api",
			body:        []byte(`{"key": "value"}`),
			expectError: false,
		},
		{
			name:        "Invalid URL",
			method:      GET,
			url:         "://invalid-url",
			body:        nil,
			expectError: true,
		},
		{
			name:        "Empty URL",
			method:      GET,
			url:         "",
			body:        nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := NewRequest(tt.method, tt.url, tt.body)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if req.Method != tt.method {
				t.Errorf("Expected method %s, got %s", tt.method, req.Method)
			}

			if req.URL.String() != tt.url {
				t.Errorf("Expected URL %s, got %s", tt.url, req.URL.String())
			}

			if !bytes.Equal(req.Body, tt.body) {
				t.Errorf("Expected body %v, got %v", tt.body, req.Body)
			}

			if req.Header == nil {
				t.Errorf("Header should be initialized")
			}
		})
	}
}

func TestRequest_Raw(t *testing.T) {
	req, err := NewRequest(POST, "http://localhost:8080/api", []byte(`{"test": "data"}`))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	raw := req.Raw()

	// Check that raw request contains expected components
	rawStr := string(raw)

	// Check request line
	if !bytes.Contains(raw, []byte("POST /api HTTP/1.1")) {
		t.Errorf("Raw request should contain request line: %s", rawStr)
	}

	// Check Host header
	if !bytes.Contains(raw, []byte("Host: localhost:8080")) {
		t.Errorf("Raw request should contain Host header: %s", rawStr)
	}

	// Check Content-Type header
	if !bytes.Contains(raw, []byte("Content-Type: application/json")) {
		t.Errorf("Raw request should contain Content-Type header: %s", rawStr)
	}

	// Check Accept header
	if !bytes.Contains(raw, []byte("Accept: application/json")) {
		t.Errorf("Raw request should contain Accept header: %s", rawStr)
	}

	// Check body
	if !bytes.Contains(raw, []byte(`{"test": "data"}`)) {
		t.Errorf("Raw request should contain body: %s", rawStr)
	}

	// Check proper line endings
	if !bytes.Contains(raw, []byte("\r\n\r\n")) {
		t.Errorf("Raw request should contain proper header-body separator: %s", rawStr)
	}
}

func TestRequest_Raw_WithQueryParams(t *testing.T) {
	req, err := NewRequest(GET, "http://localhost:8080/api?param1=value1&param2=value2", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	raw := req.Raw()
	rawStr := string(raw)

	// Check that query parameters are included in request line
	if !bytes.Contains(raw, []byte("GET /api?param1=value1&param2=value2 HTTP/1.1")) {
		t.Errorf("Raw request should contain query parameters: %s", rawStr)
	}
}

func TestRequest_Raw_WithPathParams(t *testing.T) {
	req, err := NewRequest(GET, "http://localhost:8080/api/users/123", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	raw := req.Raw()
	rawStr := string(raw)

	// Check that path is included correctly
	if !bytes.Contains(raw, []byte("GET /api/users/123 HTTP/1.1")) {
		t.Errorf("Raw request should contain correct path: %s", rawStr)
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name        string
		rawResponse string
		expectError bool
		checkFunc   func(*Response) error
	}{
		{
			name: "Valid JSON response",
			rawResponse: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"Content-Length: 23\r\n" +
				"\r\n" +
				`{"message": "success"}`,
			expectError: false,
			checkFunc: func(resp *Response) error {
				if resp.StatusCode != 200 {
					return fmt.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
				if resp.Header.Get("Content-Type") != ContentTypeJSON {
					return fmt.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
				}
				if string(resp.Body) != `{"message": "success"}` {
					return fmt.Errorf("Expected body %s, got %s", `{"message": "success"}`, string(resp.Body))
				}
				return nil
			},
		},
		{
			name: "Response without Content-Length",
			rawResponse: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: text/plain\r\n" +
				"\r\n" +
				"Hello, World!",
			expectError: false,
			checkFunc: func(resp *Response) error {
				if resp.StatusCode != 200 {
					return fmt.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
				if string(resp.Body) != "Hello, World!" {
					return fmt.Errorf("Expected body 'Hello, World!', got '%s'", string(resp.Body))
				}
				return nil
			},
		},
		{
			name:        "Empty response",
			rawResponse: "",
			expectError: true,
		},
		{
			name:        "Invalid response format",
			rawResponse: "Invalid HTTP response",
			expectError: true,
		},
		{
			name: "Response with chunked encoding",
			rawResponse: "HTTP/1.1 200 OK\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"\r\n" +
				"5\r\n" +
				"Hello\r\n" +
				"6\r\n" +
				" World\r\n" +
				"0\r\n" +
				"\r\n",
			expectError: false,
			checkFunc: func(resp *Response) error {
				if string(resp.Body) != "Hello World" {
					return fmt.Errorf("Expected body 'Hello World', got '%s'", string(resp.Body))
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := parseResponse([]byte(tt.rawResponse))

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkFunc != nil {
				if err := tt.checkFunc(resp); err != nil {
					t.Errorf("Response validation failed: %v", err)
				}
			}
		})
	}
}

func TestResponse_Unmarshal(t *testing.T) {
	resp := &Response{
		Body: []byte(`{"name": "John", "age": 30, "active": true}`),
	}

	var data struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}

	err := resp.Unmarshal(&data)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if data.Name != "John" {
		t.Errorf("Expected name 'John', got '%s'", data.Name)
	}
	if data.Age != 30 {
		t.Errorf("Expected age 30, got %d", data.Age)
	}
	if !data.Active {
		t.Errorf("Expected active true, got %v", data.Active)
	}
}

func TestResponse_Unmarshal_InvalidJSON(t *testing.T) {
	resp := &Response{
		Body: []byte(`{"name": "John", "age": 30, "active": true`), // Missing closing brace
	}

	var data map[string]interface{}

	err := resp.Unmarshal(&data)
	if err == nil {
		t.Errorf("Expected error for invalid JSON, got none")
	}
}

func TestClient_Send_Timeout(t *testing.T) {
	client := &Client{
		Timeout: 100 * time.Millisecond,
	}

	req, err := NewRequest(GET, "http://localhost:9999/timeout", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	_, err = client.Send(req)
	if err == nil {
		t.Errorf("Expected timeout error, got none")
	}

	// Check that error contains connection information
	if !bytes.Contains([]byte(err.Error()), []byte("connection refused")) {
		t.Errorf("Expected connection error, got: %v", err)
	}
}

func TestClient_Send_InvalidHost(t *testing.T) {
	client := &Client{
		Timeout: 5 * time.Second,
	}

	req, err := NewRequest(GET, "http://invalid-host-that-does-not-exist:8080/api", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	_, err = client.Send(req)
	if err == nil {
		t.Errorf("Expected connection error, got none")
	}
}

func TestClient_DefaultTimeout(t *testing.T) {
	client := &Client{} // No timeout set

	// Should use default timeout
	if client.Timeout != 0 {
		t.Errorf("Expected zero timeout initially, got %v", client.Timeout)
	}

	req, err := NewRequest(GET, "http://localhost:9999/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// The Send method should set a default timeout
	_, err = client.Send(req)
	if err == nil {
		t.Errorf("Expected error, got none")
	}

	// After Send, timeout should be set
	if client.Timeout == 0 {
		t.Errorf("Expected timeout to be set after Send, got %v", client.Timeout)
	}
}

func TestRequest_WithHeaders(t *testing.T) {
	req, err := NewRequest(POST, "http://localhost:8080/api", []byte(`{"data": "test"}`))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set multiple headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Accept", "text/plain")

	raw := req.Raw()
	rawStr := string(raw)

	// Check all headers are present
	expectedHeaders := []string{
		"Content-Type: application/json",
		"Authorization: Bearer token123",
		"Accept: application/json, text/plain",
	}

	for _, header := range expectedHeaders {
		if !bytes.Contains(raw, []byte(header)) {
			t.Errorf("Raw request should contain header '%s': %s", header, rawStr)
		}
	}
}

func TestRequest_WithEmptyBody(t *testing.T) {
	req, err := NewRequest(GET, "http://localhost:8080/api", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	raw := req.Raw()
	rawStr := string(raw)

	// Should end with double CRLF for empty body
	if !bytes.HasSuffix(raw, []byte("\r\n\r\n")) {
		t.Errorf("Raw request should end with double CRLF for empty body: %s", rawStr)
	}
}

func TestRequest_WithLargeBody(t *testing.T) {
	// Create a large body
	largeBody := make([]byte, 10000)
	for i := range largeBody {
		largeBody[i] = byte(i % 256)
	}

	req, err := NewRequest(POST, "http://localhost:8080/api", largeBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	raw := req.Raw()

	// Check that body is included
	if len(raw) < len(largeBody) {
		t.Errorf("Raw request should include body, got length %d, expected at least %d", len(raw), len(largeBody))
	}

	// Check that body is at the end
	bodyStart := bytes.LastIndex(raw, []byte("\r\n\r\n"))
	if bodyStart == -1 {
		t.Errorf("Raw request should contain header-body separator")
	}

	body := raw[bodyStart+4:]
	if !bytes.Equal(body, largeBody) {
		t.Errorf("Body in raw request doesn't match original body")
	}
}
