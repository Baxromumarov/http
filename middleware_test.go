package http_go

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestLogger(t *testing.T) {
	// Create a test request
	req := &Request{
		Method: GET,
		Path:   "/api/test",
		Header: make(Header),
	}

	// Create a test handler
	testHandler := func(req *Request, params map[string]string, next HandlerFunc) *Response {
		return &Response{
			StatusCode: 200,
			Body:       []byte("test response"),
		}
	}

	// Apply logger middleware
	loggerHandler := Logger()

	// Execute the handler
	resp := loggerHandler(req, nil, testHandler)

	// Check that response is correct
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "test response" {
		t.Errorf("Expected body 'test response', got '%s'", string(resp.Body))
	}
}

func TestRecover(t *testing.T) {
	// Create a test request
	req := &Request{
		Method: GET,
		Path:   "/api/test",
		Header: make(Header),
	}

	// Create a handler that panics
	panicHandler := func(req *Request, params map[string]string, next HandlerFunc) *Response {
		panic("test panic")
	}

	// Apply recover middleware
	recoverHandler := Recover()

	// Execute the handler - should not panic
	resp := recoverHandler(req, nil, panicHandler)

	// Check that we get a 500 response
	if resp.StatusCode != 500 {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "Internal Server Error" {
		t.Errorf("Expected body 'Internal Server Error', got '%s'", string(resp.Body))
	}
}

func TestCORS(t *testing.T) {
	// Create a test request
	req := &Request{
		Method: GET,
		Path:   "/api/test",
		Header: make(Header),
	}

	// Create a test handler
	testHandler := func(req *Request, params map[string]string, next HandlerFunc) *Response {
		return &Response{
			StatusCode: 200,
			Body:       []byte("test response"),
		}
	}

	// Apply CORS middleware
	corsHandler := CORS()

	// Execute the handler
	resp := corsHandler(req, nil, testHandler)

	// Check that CORS headers are set
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin *, got %s", resp.Header.Get("Access-Control-Allow-Origin"))
	}
	if resp.Header.Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("Expected Access-Control-Allow-Methods, got %s", resp.Header.Get("Access-Control-Allow-Methods"))
	}
	if resp.Header.Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
		t.Errorf("Expected Access-Control-Allow-Headers, got %s", resp.Header.Get("Access-Control-Allow-Headers"))
	}

	// Check that response is correct
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestBasicAuth(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
	}{
		{
			name:           "Valid credentials",
			username:       "admin",
			password:       "password",
			expectedStatus: 200,
		},
		{
			name:           "Invalid username",
			username:       "wrong",
			password:       "password",
			expectedStatus: 401,
		},
		{
			name:           "Invalid password",
			username:       "admin",
			password:       "wrong",
			expectedStatus: 401,
		},
		{
			name:           "No credentials",
			username:       "",
			password:       "",
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := &Request{
				Method: GET,
				Path:   "/api/test",
				Header: make(Header),
			}

			// Add Authorization header if credentials provided
			if tt.username != "" && tt.password != "" {
				credentials := fmt.Sprintf("%s:%s", tt.username, tt.password)
				encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
				req.Header.Set("Authorization", "Basic "+encoded)
			}

			// Create a test handler
			testHandler := func(req *Request, params map[string]string, next HandlerFunc) *Response {
				return &Response{
					StatusCode: 200,
					Body:       []byte("authenticated"),
				}
			}

			// Apply BasicAuth middleware
			users := map[string]string{"admin": "password"}
			authHandler := BasicAuth(users)

			// Execute the handler
			resp := authHandler(req, nil, testHandler)

			// Check response status
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Check WWW-Authenticate header for 401
			if tt.expectedStatus == 401 {
				if resp.Header.Get("WWW-Authenticate") != "Basic realm=\"Restricted\"" {
					t.Errorf("Expected WWW-Authenticate header, got %s", resp.Header.Get("WWW-Authenticate"))
				}
			}
		})
	}
}

func TestMiddlewareChain(t *testing.T) {
	// Create a test request
	req := &Request{
		Method: GET,
		Path:   "/api/test",
		Header: make(Header),
	}

	// Create a test handler
	testHandler := func(req *Request, params map[string]string, next HandlerFunc) *Response {
		return &Response{
			StatusCode: 200,
			Body:       []byte("test response"),
		}
	}

	// Apply multiple middleware
	loggerHandler := Logger()
	recoverHandler := Recover()
	corsHandler := CORS()

	// Execute the handler chain
	resp := loggerHandler(req, nil, func(req *Request, params map[string]string, next HandlerFunc) *Response {
		return recoverHandler(req, params, func(req *Request, params map[string]string, next HandlerFunc) *Response {
			return corsHandler(req, params, testHandler)
		})
	})

	// Check that response is correct
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "test response" {
		t.Errorf("Expected body 'test response', got '%s'", string(resp.Body))
	}

	// Check that CORS headers are present
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS headers to be present")
	}
}

func TestMiddlewareWithPanic(t *testing.T) {
	// Create a test request
	req := &Request{
		Method: GET,
		Path:   "/api/test",
		Header: make(Header),
	}

	// Create a handler that panics
	panicHandler := func(req *Request, params map[string]string, next HandlerFunc) *Response {
		panic("test panic")
	}

	// Apply middleware chain with panic handler
	loggerHandler := Logger()
	recoverHandler := Recover()
	corsHandler := CORS()

	// Execute the handler - should not panic
	resp := loggerHandler(req, nil, func(req *Request, params map[string]string, next HandlerFunc) *Response {
		return recoverHandler(req, params, func(req *Request, params map[string]string, next HandlerFunc) *Response {
			return corsHandler(req, params, panicHandler)
		})
	})

	// Check that we get a 500 response
	if resp.StatusCode != 500 {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	// Check that CORS headers are still present
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS headers to be present even after panic")
	}
}

func TestCORS_OptionsRequest(t *testing.T) {
	// Create a test OPTIONS request
	req := &Request{
		Method: "OPTIONS",
		Path:   "/api/test",
		Header: make(Header),
	}

	// Apply CORS middleware
	corsHandler := CORS()

	// Execute the handler
	resp := corsHandler(req, nil, nil)

	// Check that we get a 204 response for preflight
	if resp.StatusCode != 204 {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Check CORS headers
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin *, got %s", resp.Header.Get("Access-Control-Allow-Origin"))
	}
	if resp.Header.Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("Expected Access-Control-Allow-Methods, got %s", resp.Header.Get("Access-Control-Allow-Methods"))
	}
}

func TestBasicAuth_Context(t *testing.T) {
	// Create a test request
	req := &Request{
		Method: GET,
		Path:   "/api/test",
		Header: make(Header),
	}
	req.Header.Set("Authorization", "Basic YWRtaW46cGFzc3dvcmQ=") // admin:password

	// Create a test handler that checks context
	testHandler := func(req *Request, params map[string]string, next HandlerFunc) *Response {
		username := req.Context().Value("username")
		if username != "admin" {
			t.Errorf("Expected username 'admin' in context, got %v", username)
		}
		return &Response{
			StatusCode: 200,
			Body:       []byte("authenticated"),
		}
	}

	// Apply BasicAuth middleware
	users := map[string]string{"admin": "password"}
	authHandler := BasicAuth(users)

	// Execute the handler
	resp := authHandler(req, nil, testHandler)

	// Check response
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
