package http

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestParseRequest(t *testing.T) {
	tests := []struct {
		name        string
		rawRequest  string
		expectError bool
		checkFunc   func(*Request) error
	}{
		{
			name: "Valid GET request",
			rawRequest: "GET /api/users HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Accept: application/json\r\n" +
				"\r\n",
			expectError: false,
			checkFunc: func(req *Request) error {
				if req.Method != GET {
					return fmt.Errorf("Expected method GET, got %s", req.Method)
				}
				if req.Path != "/api/users" {
					return fmt.Errorf("Expected path /api/users, got %s", req.Path)
				}
				if req.Version != "HTTP/1.1" {
					return fmt.Errorf("Expected version HTTP/1.1, got %s", req.Version)
				}
				if req.Header.Get("Host") != "localhost:8080" {
					return fmt.Errorf("Expected Host localhost:8080, got %s", req.Header.Get("Host"))
				}
				if req.Header.Get("Accept") != "application/json" {
					return fmt.Errorf("Expected Accept application/json, got %s", req.Header.Get("Accept"))
				}
				return nil
			},
		},
		{
			name: "Valid POST request with body",
			rawRequest: "POST /api/users HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Type: application/json\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				`{"name": "John Doe"}`,
			expectError: false,
			checkFunc: func(req *Request) error {
				if req.Method != POST {
					return fmt.Errorf("Expected method POST, got %s", req.Method)
				}
				if string(req.Body) != `{"name": "John Doe"}` {
					return fmt.Errorf("Expected body %s, got %s", `{"name": "John Doe"}`, string(req.Body))
				}
				if req.Header.Get("Content-Type") != "application/json" {
					return fmt.Errorf("Expected Content-Type application/json, got %s", req.Header.Get("Content-Type"))
				}
				return nil
			},
		},
		{
			name: "Request with query parameters",
			rawRequest: "GET /api/users?page=1&limit=10 HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"\r\n",
			expectError: false,
			checkFunc: func(req *Request) error {
				if req.Path != "/api/users" {
					return fmt.Errorf("Expected path /api/users, got %s", req.Path)
				}
				pageValues := req.QueryValue("page")
				if len(pageValues) == 0 || pageValues[0] != "1" {
					return fmt.Errorf("Expected query page=1, got %v", pageValues)
				}
				limitValues := req.QueryValue("limit")
				if len(limitValues) == 0 || limitValues[0] != "10" {
					return fmt.Errorf("Expected query limit=10, got %v", limitValues)
				}
				return nil
			},
		},
		{
			name:        "Empty request",
			rawRequest:  "",
			expectError: true,
		},
		{
			name:        "Invalid request line",
			rawRequest:  "INVALID REQUEST\r\n\r\n",
			expectError: true,
		},
		{
			name:        "Unsupported HTTP version",
			rawRequest:  "GET /api HTTP/2.0\r\n\r\n",
			expectError: true,
		},
		{
			name:        "Unsupported method",
			rawRequest:  "PATCH /api HTTP/1.1\r\n\r\n",
			expectError: true,
		},
		{
			name: "Request with chunked encoding",
			rawRequest: "POST /api HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"\r\n" +
				"5\r\n" +
				"Hello\r\n" +
				"6\r\n" +
				" World\r\n" +
				"0\r\n" +
				"\r\n",
			expectError: false,
			checkFunc: func(req *Request) error {
				if string(req.Body) != "Hello World" {
					return fmt.Errorf("Expected body 'Hello World', got '%s'", string(req.Body))
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := parseRequest([]byte(tt.rawRequest))

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
				if err := tt.checkFunc(req); err != nil {
					t.Errorf("Request validation failed: %v", err)
				}
			}
		})
	}
}

func TestMatchRoute(t *testing.T) {
	// Clear existing routes
	routes = make(map[Method][]route)

	// Register test routes
	Handle(GET, "/api/users", func(req *Request) *Response {
		return &Response{StatusCode: 200, Body: []byte("users list")}
	})

	Handle(GET, "/api/users/:id", func(req *Request) *Response {
		return &Response{StatusCode: 200, Body: []byte("user " + req.PathValue("id"))}
	})

	Handle(POST, "/api/users", func(req *Request) *Response {
		return &Response{StatusCode: 201, Body: []byte("user created")}
	})

	tests := []struct {
		name     string
		method   Method
		path     string
		expected string
		found    bool
	}{
		{
			name:     "Exact match",
			method:   GET,
			path:     "/api/users",
			expected: "users list",
			found:    true,
		},
		{
			name:     "Path parameter match",
			method:   GET,
			path:     "/api/users/123",
			expected: "user 123",
			found:    true,
		},
		{
			name:     "POST method match",
			method:   POST,
			path:     "/api/users",
			expected: "user created",
			found:    true,
		},
		{
			name:     "Method mismatch",
			method:   PUT,
			path:     "/api/users",
			expected: "",
			found:    false,
		},
		{
			name:     "Path not found",
			method:   GET,
			path:     "/api/posts",
			expected: "",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, params := matchRoute(tt.method, tt.path)

			if tt.found {
				if handler == nil {
					t.Errorf("Expected handler to be found")
					return
				}

				// Create a mock request
				req := &Request{Method: tt.method, Path: tt.path}
				req.pathParams = params
				resp := handler(req)

				if string(resp.Body) != tt.expected {
					t.Errorf("Expected response '%s', got '%s'", tt.expected, string(resp.Body))
				}
			} else {
				if handler != nil {
					t.Errorf("Expected no handler to be found")
				}
			}
		})
	}
}

func TestServer_NewDefaultServer(t *testing.T) {
	server := NewDefaultServer("localhost", 8080)

	if server.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", server.Host)
	}
	if server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", server.Port)
	}
	if server.conn == nil {
		t.Errorf("Expected conn map to be initialized")
	}
}

func TestServer_TimeoutHandler(t *testing.T) {
	server := NewDefaultServer("localhost", 8080)

	// Create a mock connection
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	err = server.timeoutHandler(conn)
	if err != nil {
		t.Errorf("Expected no error from timeoutHandler, got %v", err)
	}
}

func TestServer_SetupConnection(t *testing.T) {
	server := NewDefaultServer("localhost", 8080)

	// Create a mock connection
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	err = server.setupConnection(conn)
	if err != nil {
		t.Errorf("Expected no error from setupConnection, got %v", err)
	}

	// Check that connection is tracked
	server.mu.Lock()
	if len(server.conn) != 1 {
		t.Errorf("Expected 1 tracked connection, got %d", len(server.conn))
	}
	server.mu.Unlock()
}

func TestServer_ReadHeaders(t *testing.T) {
	server := NewDefaultServer("localhost", 8080)

	// Create a mock connection with headers
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Start a goroutine to accept and send headers
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		headers := "GET /api HTTP/1.1\r\n" +
			"Host: localhost:8080\r\n" +
			"Content-Type: application/json\r\n" +
			"\r\n"

		conn.Write([]byte(headers))
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	headersPart, bodyStart, err := server.readHeaders(conn)
	if err != nil {
		t.Errorf("Expected no error from readHeaders, got %v", err)
	}

	// Check that headers were read correctly
	headersStr := string(headersPart)
	if !strings.Contains(headersStr, "GET /api HTTP/1.1") {
		t.Errorf("Expected headers to contain request line: %s", headersStr)
	}
	if !strings.Contains(headersStr, "Host: localhost:8080") {
		t.Errorf("Expected headers to contain Host header: %s", headersStr)
	}
	if !strings.Contains(headersStr, "Content-Type: application/json") {
		t.Errorf("Expected headers to contain Content-Type header: %s", headersStr)
	}

	// Body start should be empty for this request
	if len(bodyStart) != 0 {
		t.Errorf("Expected empty body start, got %d bytes", len(bodyStart))
	}
}

func TestServer_GetContentLength(t *testing.T) {
	server := NewDefaultServer("localhost", 8080)

	headersPart := []byte("GET /api HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: application/json\r\n" +
		"Content-Length: 25\r\n" +
		"\r\n")

	contentLength, err := server.getContentLength(headersPart)
	if err != nil {
		t.Errorf("Expected no error from getContentLength, got %v", err)
	}

	if contentLength != 25 {
		t.Errorf("Expected content length 25, got %d", contentLength)
	}
}

func TestServer_GetContentLength_NoContentLength(t *testing.T) {
	server := NewDefaultServer("localhost", 8080)

	headersPart := []byte("GET /api HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n")

	contentLength, err := server.getContentLength(headersPart)
	if err != nil {
		t.Errorf("Expected no error from getContentLength, got %v", err)
	}

	if contentLength != 0 {
		t.Errorf("Expected content length 0, got %d", contentLength)
	}
}

func TestServer_ReadBody(t *testing.T) {
	server := NewDefaultServer("localhost", 8080)

	// Test the readBody function with a simple case
	bodyStart := []byte("Hello")
	contentLength := 12 // "Hello, World!" length (5 + 7)

	// Create a buffer with the remaining body data
	remainingBody := []byte(" World!") // 7 bytes

	// Create a mock connection using bytes.Buffer
	mockConn := &mockConn{
		buffer: bytes.NewBuffer(remainingBody),
	}

	fullBody, err := server.readBody(mockConn, bodyStart, contentLength)
	if err != nil {
		t.Errorf("Expected no error from readBody, got %v", err)
	}

	expectedBody := "Hello World!"
	if string(fullBody) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(fullBody))
	}
}

// mockConn implements net.Conn for testing
type mockConn struct {
	buffer *bytes.Buffer
}

func (m *mockConn) Read(p []byte) (n int, err error) {
	return m.buffer.Read(p)
}

func (m *mockConn) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return nil
}

func (m *mockConn) RemoteAddr() net.Addr {
	return nil
}

func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestServer_ProcessRequest(t *testing.T) {
	server := NewDefaultServer("localhost", 8080)

	// Clear existing routes
	routes = make(map[Method][]route)

	// Register a test route
	Handle(GET, "/api/test", func(req *Request) *Response {
		return &Response{
			StatusCode: 200,
			Header:     Header{"Content-Type": {"application/json"}},
			Body:       []byte(`{"message": "success"}`),
		}
	})

	req := &Request{
		Method: GET,
		Path:   "/api/test",
		Header: make(Header),
	}

	resp := server.processRequest(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}
	if string(resp.Body) != `{"message": "success"}` {
		t.Errorf("Expected body %s, got %s", `{"message": "success"}`, string(resp.Body))
	}
}

func TestServer_ProcessRequest_NotFound(t *testing.T) {
	server := NewDefaultServer("localhost", 8080)

	// Clear existing routes
	routes = make(map[Method][]route)

	req := &Request{
		Method: GET,
		Path:   "/api/notfound",
		Header: make(Header),
	}

	resp := server.processRequest(req)

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "404 Not found" {
		t.Errorf("Expected body '404 Not found', got '%s'", string(resp.Body))
	}
}

func TestServer_CreateHandlerChain(t *testing.T) {
	server := NewDefaultServer("localhost", 8080)

	// Create a test handler
	testHandler := func(req *Request) *Response {
		return &Response{StatusCode: 200, Body: []byte("test response")}
	}

	handlerChain := server.createHandlerChain(testHandler)

	if handlerChain == nil {
		t.Errorf("Expected handler chain to be created")
	}

	// Test the handler chain
	req := &Request{Method: GET, Path: "/test"}
	resp := handlerChain(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "test response" {
		t.Errorf("Expected body 'test response', got '%s'", string(resp.Body))
	}
}

func TestServer_EnsureResponseHeaders(t *testing.T) {
	server := NewDefaultServer("localhost", 8080)

	// Test with nil header
	resp := &Response{
		StatusCode: 200,
		Body:       []byte("test"),
	}

	server.ensureResponseHeaders(resp)

	if resp.Header == nil {
		t.Errorf("Expected header to be initialized")
	}

	// Test with existing Content-Type
	resp = &Response{
		StatusCode: 200,
		Header:     Header{"Content-Type": {"text/plain"}},
		Body:       []byte("test"),
	}

	server.ensureResponseHeaders(resp)

	if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type to remain text/plain, got %s", resp.Header.Get("Content-Type"))
	}

	// Test with no Content-Type and empty body
	resp = &Response{
		StatusCode: 200,
		Header:     make(Header),
		Body:       []byte{},
	}

	server.ensureResponseHeaders(resp)

	if resp.Header.Get("Content-Type") != "" {
		t.Errorf("Expected no Content-Type for empty body, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestDecodeChunked(t *testing.T) {
	tests := []struct {
		name        string
		chunkedData string
		expected    string
		expectError bool
	}{
		{
			name: "Simple chunked data",
			chunkedData: "5\r\n" +
				"Hello\r\n" +
				"6\r\n" +
				" World\r\n" +
				"0\r\n" +
				"\r\n",
			expected:    "Hello World",
			expectError: false,
		},
		{
			name: "Single chunk",
			chunkedData: "5\r\n" +
				"Hello\r\n" +
				"0\r\n" +
				"\r\n",
			expected:    "Hello",
			expectError: false,
		},
		{
			name:        "Empty chunked data",
			chunkedData: "0\r\n\r\n",
			expected:    "",
			expectError: false,
		},
		{
			name:        "Invalid chunk size",
			chunkedData: "invalid\r\nHello\r\n0\r\n\r\n",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Incomplete chunk data",
			chunkedData: "10\r\nHello",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeChunked([]byte(tt.chunkedData))

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

			if string(result) != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, string(result))
			}
		})
	}
}
