package http_go

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestIntegration_ServerAndClient(t *testing.T) {
	// Start a test server
	server := NewDefaultServer("localhost", 0) // Use port 0 to get a random available port

	// Register test routes
	Handle(GET, "/api/test", func(req *Request) *Response {
		return &Response{
			StatusCode: 200,
			Header:     Header{"Content-Type": {ContentTypeJSON}},
			Body:       []byte(`{"message": "success", "method": "GET"}`),
		}
	})

	Handle(POST, "/api/test", func(req *Request) *Response {
		return &Response{
			StatusCode: 201,
			Header:     Header{"Content-Type": {ContentTypeJSON}},
			Body:       []byte(fmt.Sprintf(`{"message": "created", "body": "%s"}`, string(req.Body))),
		}
	})

	Handle(GET, "/api/users/:id", func(req *Request) *Response {
		userID := req.PathValue("id")
		return &Response{
			StatusCode: 200,
			Header:     Header{"Content-Type": {ContentTypeJSON}},
			Body:       []byte(fmt.Sprintf(`{"id": "%s", "name": "User %s"}`, userID, userID)),
		}
	})

	// Start server in background
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	serverPort := listener.Addr().(*net.TCPAddr).Port

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConn(conn)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test client
	client := &Client{Timeout: 5 * time.Second}

	t.Run("GET Request", func(t *testing.T) {
		req, err := NewRequest(GET, fmt.Sprintf("http://localhost:%d/api/test", serverPort), nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := client.Send(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var data map[string]interface{}
		err = resp.Unmarshal(&data)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if data["message"] != "success" {
			t.Errorf("Expected message 'success', got %v", data["message"])
		}
		if data["method"] != "GET" {
			t.Errorf("Expected method 'GET', got %v", data["method"])
		}
	})

	t.Run("POST Request", func(t *testing.T) {
		body := []byte(`{"name": "test user"}`)
		req, err := NewRequest(POST, fmt.Sprintf("http://localhost:%d/api/test", serverPort), body)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", ContentTypeJSON)

		resp, err := client.Send(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != 201 {
			t.Errorf("Expected status 201, got %d", resp.StatusCode)
		}

		var data map[string]interface{}
		err = resp.Unmarshal(&data)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if data["message"] != "created" {
			t.Errorf("Expected message 'created', got %v", data["message"])
		}
	})

	t.Run("Path Parameters", func(t *testing.T) {
		req, err := NewRequest(GET, fmt.Sprintf("http://localhost:%d/api/users/123", serverPort), nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := client.Send(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var data map[string]interface{}
		err = resp.Unmarshal(&data)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if data["id"] != "123" {
			t.Errorf("Expected id '123', got %v", data["id"])
		}
		if data["name"] != "User 123" {
			t.Errorf("Expected name 'User 123', got %v", data["name"])
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		req, err := NewRequest(GET, fmt.Sprintf("http://localhost:%d/api/notfound", serverPort), nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := client.Send(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != 404 {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}

		if string(resp.Body) != "404 Not found" {
			t.Errorf("Expected body '404 Not found', got '%s'", string(resp.Body))
		}
	})
}

func TestIntegration_Middleware(t *testing.T) {
	// Start a test server with middleware
	server := NewDefaultServer("localhost", 0)

	// Add middleware to the server
	server.Use(CORS())

	// Create a handler that returns user info
	userHandler := func(req *Request) *Response {
		username := req.Context().Value("username")
		return &Response{
			StatusCode: 200,
			Header:     Header{"Content-Type": {ContentTypeJSON}},
			Body:       []byte(fmt.Sprintf(`{"message": "Hello %v"}`, username)),
		}
	}

	// Register the handler with middleware
	Handle(GET, "/api/user", userHandler)

	// Start server
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	serverPort := listener.Addr().(*net.TCPAddr).Port

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConn(conn)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Test client
	client := &Client{Timeout: 5 * time.Second}

	t.Run("Request with CORS", func(t *testing.T) {
		req, err := NewRequest(GET, fmt.Sprintf("http://localhost:%d/api/user", serverPort), nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := client.Send(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check CORS headers
		if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("Expected CORS header Access-Control-Allow-Origin *, got %s", resp.Header.Get("Access-Control-Allow-Origin"))
		}

		var data map[string]interface{}
		err = resp.Unmarshal(&data)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if data["message"] != "Hello <nil>" {
			t.Errorf("Expected message 'Hello <nil>', got %s", data["message"])
		}
	})
}

func TestIntegration_LargeRequest(t *testing.T) {
	// Start a test server
	server := NewDefaultServer("localhost", 0)

	// Register handler that echoes back the request body
	Handle(POST, "/api/echo", func(req *Request) *Response {
		return &Response{
			StatusCode: 200,
			Header:     Header{"Content-Type": {ContentTypeJSON}},
			Body:       []byte(fmt.Sprintf(`{"length": %d, "body": "%s"}`, len(req.Body), string(req.Body))),
		}
	})

	// Start server
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	serverPort := listener.Addr().(*net.TCPAddr).Port

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConn(conn)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Test client
	client := &Client{Timeout: 10 * time.Second}

	t.Run("Large Request Body", func(t *testing.T) {
		// Create a large body (10KB) with printable characters
		largeBody := make([]byte, 10240)
		for i := range largeBody {
			largeBody[i] = byte('A' + (i % 26)) // Use only printable ASCII characters
		}

		req, err := NewRequest(POST, fmt.Sprintf("http://localhost:%d/api/echo", serverPort), largeBody)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", ContentTypeJSON)
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(largeBody)))

		resp, err := client.Send(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var data map[string]interface{}
		err = resp.Unmarshal(&data)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Check that the server received the full body
		length := int(data["length"].(float64))
		if length != len(largeBody) {
			t.Errorf("Expected body length %d, got %d", len(largeBody), length)
		}
	})
}

func TestIntegration_ConcurrentRequests(t *testing.T) {
	// Start a test server
	server := NewDefaultServer("localhost", 0)

	// Register a simple handler
	Handle(GET, "/api/test", func(req *Request) *Response {
		return &Response{
			StatusCode: 200,
			Header:     Header{"Content-Type": {ContentTypeJSON}},
			Body:       []byte(`{"message": "success"}`),
		}
	})

	// Start server
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	serverPort := listener.Addr().(*net.TCPAddr).Port

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConn(conn)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Test client
	client := &Client{Timeout: 5 * time.Second}

	t.Run("Concurrent Requests", func(t *testing.T) {
		const numRequests = 10
		results := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(id int) {
				req, err := NewRequest(GET, fmt.Sprintf("http://localhost:%d/api/test", serverPort), nil)
				if err != nil {
					results <- fmt.Errorf("request %d: failed to create request: %v", id, err)
					return
				}

				resp, err := client.Send(req)
				if err != nil {
					results <- fmt.Errorf("request %d: failed to send request: %v", id, err)
					return
				}

				if resp.StatusCode != 200 {
					results <- fmt.Errorf("request %d: expected status 200, got %d", id, resp.StatusCode)
					return
				}

				results <- nil
			}(i)
		}

		// Collect results
		for i := 0; i < numRequests; i++ {
			if err := <-results; err != nil {
				t.Errorf("Concurrent request failed: %v", err)
			}
		}
	})
}
