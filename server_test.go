package http_go

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseRequest_ValidGETRequest(t *testing.T) {
	data := []byte("GET /hello?name=go HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n")
	t.Run("valid GET request with headers", func(t *testing.T) {
		req, err := parseRequest(data)
		assert.NoError(t, err)

		assert.Equal(t, GET, req.Method)
		assert.Equal(t, "/hello", req.Path)
		assert.Equal(t, map[string][]string{"name": {"go"}}, req.Query)
		assert.Equal(t, "HTTP/1.1", req.Version)
		assert.Equal(t, "localhost:8080", req.Header.Get("Host"))
		assert.Equal(t, "curl/7.81.0", req.Header.Get("User-Agent"))
		assert.Equal(t, "*/*", req.Header.Get("Accept"))
	})
}
func TestMatchRoute(t *testing.T) {
	// Setup routes with patterns
	routes = map[Method][]route{}

	// Register some routes
	Handle(GET, "/hello/:id", func(req *Request, params map[string]string) *Response {
		return nil
	})
	Handle(GET, "/user/:userId/profile", func(req *Request, params map[string]string) *Response {
		return nil
	})
	Handle(POST, "/post/:postId/comment/:commentId", func(req *Request, params map[string]string) *Response {
		return nil
	})

	tests := []struct {
		method     Method
		path       string
		wantParams map[string]string
		wantFound  bool
		wantIndex  int // index of matched route
	}{
		{
			method:    GET,
			path:      "/hello/123",
			wantFound: true,
			wantParams: map[string]string{
				"id": "123",
			},
			wantIndex: 0,
		},
		{
			method:    GET,
			path:      "/user/456/profile",
			wantFound: true,
			wantParams: map[string]string{
				"userId": "456",
			},
			wantIndex: 1,
		},
		{
			method:    POST,
			path:      "/post/789/comment/10",
			wantFound: true,
			wantParams: map[string]string{
				"postId":    "789",
				"commentId": "10",
			},
			wantIndex: 2,
		},
		{
			method:    GET,
			path:      "/hello", // missing param
			wantFound: false,
		},
		{
			method:    GET,
			path:      "/hello/123/extra", // too many parts
			wantFound: false,
		},
		{
			method:    GET,
			path:      "/user/123", // incomplete path
			wantFound: false,
		},
		{
			method:    POST,
			path:      "/post/123/comment", // missing commentId param
			wantFound: false,
		},
		{
			method:    DELETE,
			path:      "/hello/123", // no DELETE route registered
			wantFound: false,
		},
	}

	for _, tt := range tests {
		handler, params := matchRoute(tt.method, tt.path)

		if (handler != nil) != tt.wantFound {
			t.Errorf("method %v path %q: got found=%v, want %v", tt.method, tt.path, handler != nil, tt.wantFound)
		}

		if tt.wantFound {
			// check params
			for k, v := range tt.wantParams {
				if params[k] != v {
					t.Errorf("method %v path %q: param %q = %q; want %q", tt.method, tt.path, k, params[k], v)
				}
			}
		}
	}
}

func TestParseRequest_QueryParameters(t *testing.T) {
	tests := []struct {
		name     string
		request  string
		expected map[string][]string
	}{
		{
			name:     "single query param",
			request:  "GET /search?q=test HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expected: map[string][]string{"q": {"test"}},
		},
		{
			name:    "multiple query params",
			request: "GET /search?q=test&sort=desc&page=2 HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expected: map[string][]string{
				"q":    {"test"},
				"sort": {"desc"},
				"page": {"2"},
			},
		},
		{
			name:    "duplicate query params",
			request: "GET /search?q=test&q=another&sort=asc HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expected: map[string][]string{
				"q":    {"test", "another"},
				"sort": {"asc"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := parseRequest([]byte(tt.request))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, req.Query)
		})
	}
}

func TestParseRequest_InvalidRequests(t *testing.T) {
	tests := []struct {
		name        string
		request     string
		expectedErr string
	}{
		{
			name:        "empty request",
			request:     "",
			expectedErr: "empty request",
		},
		{
			name:        "malformed request line",
			request:     "GET /test\r\nHost: example.com\r\n\r\n",
			expectedErr: "malformed request line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseRequest([]byte(tt.request))
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), tt.expectedErr)
			}
		})
	}
}

func TestParseRequest_DifferentHTTPMethods(t *testing.T) {
	tests := []struct {
		method     string
		shouldPass bool
	}{
		{"GET", true},
		{"POST", true},
		{"PUT", true},
		{"DELETE", true},
		{"PATCH", false},
		{"HEAD", false},    // Not implemented
		{"OPTIONS", false}, // Not implemented
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			request := []byte(tt.method + " / HTTP/1.1\r\nHost: example.com\r\n\r\n")
			_, err := parseRequest(request)
			if tt.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestParseRequest_Header(t *testing.T) {
	tests := []struct {
		name    string
		request string
		testFn  func(*testing.T, *Request)
	}{
		{
			name:    "multiple headers",
			request: "GET / HTTP/1.1\r\nHost: example.com\r\nAccept: application/json\r\nAccept: text/plain\r\nContent-Type: application/json\r\n\r\n",
			testFn: func(t *testing.T, req *Request) {
				assert.Equal(t, "example.com", req.Header.Get("Host"))
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				assert.Equal(t, []string{"application/json", "text/plain"}, req.Header.Values("Accept"))
			},
		},
		{
			name:    "empty headers",
			request: "GET / HTTP/1.1\r\n\r\n",
			testFn: func(t *testing.T, req *Request) {
				assert.Empty(t, req.Header)
			},
		},
		{
			name:    "header with spaces",
			request: "GET / HTTP/1.1\r\nUser-Agent:   My User Agent  \r\n\r\n",
			testFn: func(t *testing.T, req *Request) {
				assert.Equal(t, "My User Agent", req.Header.Get("User-Agent"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := parseRequest([]byte(tt.request))
			assert.NoError(t, err)
			tt.testFn(t, req)
		})
	}
}

func TestParseRequest_RequestURI(t *testing.T) {
	tests := []struct {
		name     string
		request  string
		expected string
	}{
		{
			name:     "root path",
			request:  "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expected: "/",
		},
		{
			name:     "nested path",
			request:  "GET /api/v1/users/123 HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expected: "/api/v1/users/123",
		},
		{
			name:     "path with special chars",
			request:  "GET /api/v1/users/name%20with%20spaces HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expected: "/api/v1/users/name with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := parseRequest([]byte(tt.request))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, req.Path)
		})
	}
}
