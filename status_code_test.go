package http

import (
	"fmt"
	"testing"
)

func TestStatusText(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{200, "OK"},
		{201, "Created"},
		{204, "No Content"},
		{301, "Moved Permanently"},
		{302, "Found"},
		{400, "Bad Request"},
		{401, "Unauthorized"},
		{403, "Forbidden"},
		{404, "Not Found"},
		{405, "Method Not Allowed"},
		{500, "Internal Server Error"},
		{501, "Not Implemented"},
		{502, "Bad Gateway"},
		{503, "Service Unavailable"},
		{999, ""},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.code), func(t *testing.T) {
			result := StatusText(tt.code)
			if result != tt.expected {
				t.Errorf("Expected status text '%s' for code %d, got '%s'", tt.expected, tt.code, result)
			}
		})
	}
}

func TestContentTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"JSON", ContentTypeJSON, "application/json"},
		{"HTML", ContentTypeHTML, "text/html"},
		{"Text", ContentTypeText, "text/plain"},
		{"XML", ContentTypeXML, "application/xml"},
		{"FormURLEncoded", ContentTypeFormURLEncoded, "application/x-www-form-urlencoded"},
		{"MultipartForm", ContentTypeMultipartForm, "multipart/form-data"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.content != tt.expected {
				t.Errorf("Expected content type '%s', got '%s'", tt.expected, tt.content)
			}
		})
	}
}

func TestResponse_WithStatus(t *testing.T) {
	resp := &Response{
		StatusCode: 200,
	}

	// Test that status code is set correctly
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Test status text
	statusText := StatusText(resp.StatusCode)
	if statusText != "OK" {
		t.Errorf("Expected status text 'OK', got '%s'", statusText)
	}
}

func TestResponse_WithHeaders(t *testing.T) {
	resp := &Response{
		Header: make(Header),
	}

	// Set headers
	resp.Header.Set("Content-Type", ContentTypeJSON)
	resp.Header.Set("Cache-Control", "no-cache")

	// Test headers
	if resp.Header.Get("Content-Type") != ContentTypeJSON {
		t.Errorf("Expected Content-Type %s, got %s", ContentTypeJSON, resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("Cache-Control") != "no-cache" {
		t.Errorf("Expected Cache-Control 'no-cache', got %s", resp.Header.Get("Cache-Control"))
	}
}

func TestResponse_WithBody(t *testing.T) {
	body := []byte(`{"message": "test"}`)
	resp := &Response{
		Header: Header{"Content-Type": {ContentTypeJSON}},
		Body:   body,
	}

	// Set Content-Length header
	resp.Header["Content-Length"] = []string{fmt.Sprintf("%d", len(resp.Body))}

	// Test body
	if string(resp.Body) != string(body) {
		t.Errorf("Expected body '%s', got '%s'", string(body), string(resp.Body))
	}

	// Test Content-Length header should be set automatically
	if resp.Header.Get("Content-Length") != "19" {
		t.Errorf("Expected Content-Length '19', got %s", resp.Header.Get("Content-Length"))
	}
}

func TestResponse_EmptyBody(t *testing.T) {
	resp := &Response{
		Header: make(Header),
		Body:   []byte{},
	}

	// Set Content-Length header
	resp.Header["Content-Length"] = []string{fmt.Sprintf("%d", len(resp.Body))}

	// Test empty body
	if len(resp.Body) != 0 {
		t.Errorf("Expected empty body, got %d bytes", len(resp.Body))
	}

	// Test Content-Length for empty body
	if resp.Header.Get("Content-Length") != "0" {
		t.Errorf("Expected Content-Length '0', got %s", resp.Header.Get("Content-Length"))
	}
}

func TestResponse_ErrorResponses(t *testing.T) {
	tests := []struct {
		statusCode int
		statusText string
		body       string
	}{
		{400, "Bad Request", "Bad Request"},
		{401, "Unauthorized", "Unauthorized"},
		{403, "Forbidden", "Forbidden"},
		{404, "Not Found", "Not Found"},
		{405, "Method Not Allowed", "Method Not Allowed"},
		{500, "Internal Server Error", "Internal Server Error"},
		{501, "Not Implemented", "Not Implemented"},
		{502, "Bad Gateway", "Bad Gateway"},
		{503, "Service Unavailable", "Service Unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.statusText, func(t *testing.T) {
			resp := &Response{
				StatusCode: tt.statusCode,
				Body:       []byte(tt.body),
			}

			// Test status code
			if resp.StatusCode != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, resp.StatusCode)
			}

			// Test status text
			statusText := StatusText(resp.StatusCode)
			if statusText != tt.statusText {
				t.Errorf("Expected status text '%s', got '%s'", tt.statusText, statusText)
			}

			// Test body
			if string(resp.Body) != tt.body {
				t.Errorf("Expected body '%s', got '%s'", tt.body, string(resp.Body))
			}
		})
	}
}

func TestResponse_RedirectResponses(t *testing.T) {
	tests := []struct {
		statusCode int
		statusText string
		location   string
	}{
		{301, "Moved Permanently", "https://example.com/new"},
		{302, "Found", "https://example.com/temp"},
		{307, "Temporary Redirect", "https://example.com/temp"},
		{308, "Permanent Redirect", "https://example.com/new"},
	}

	for _, tt := range tests {
		t.Run(tt.statusText, func(t *testing.T) {
			resp := &Response{
				StatusCode: tt.statusCode,
				Header:     Header{"Location": {tt.location}},
			}

			// Test status code
			if resp.StatusCode != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, resp.StatusCode)
			}

			// Test Location header
			if resp.Header.Get("Location") != tt.location {
				t.Errorf("Expected Location '%s', got '%s'", tt.location, resp.Header.Get("Location"))
			}

			// Test status text
			statusText := StatusText(resp.StatusCode)
			if statusText != tt.statusText {
				t.Errorf("Expected status text '%s', got '%s'", tt.statusText, statusText)
			}
		})
	}
}

func TestResponse_SuccessResponses(t *testing.T) {
	tests := []struct {
		statusCode int
		statusText string
		body       string
	}{
		{200, "OK", "Success"},
		{201, "Created", "Resource created"},
		{202, "Accepted", "Request accepted"},
		{204, "No Content", ""},
	}

	for _, tt := range tests {
		t.Run(tt.statusText, func(t *testing.T) {
			resp := &Response{
				StatusCode: tt.statusCode,
				Body:       []byte(tt.body),
			}

			// Test status code
			if resp.StatusCode != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, resp.StatusCode)
			}

			// Test status text
			statusText := StatusText(resp.StatusCode)
			if statusText != tt.statusText {
				t.Errorf("Expected status text '%s', got '%s'", tt.statusText, statusText)
			}

			// Test body
			if string(resp.Body) != tt.body {
				t.Errorf("Expected body '%s', got '%s'", tt.body, string(resp.Body))
			}
		})
	}
}

func TestResponse_JSONResponse(t *testing.T) {
	jsonBody := `{"message": "success", "data": {"id": 1, "name": "test"}}`
	resp := &Response{
		Header: Header{"Content-Type": {ContentTypeJSON}},
		Body:   []byte(jsonBody),
	}

	// Set Content-Length header
	resp.Header["Content-Length"] = []string{fmt.Sprintf("%d", len(resp.Body))}

	// Test Content-Type
	if resp.Header.Get("Content-Type") != ContentTypeJSON {
		t.Errorf("Expected Content-Type %s, got %s", ContentTypeJSON, resp.Header.Get("Content-Type"))
	}

	// Test body
	if string(resp.Body) != jsonBody {
		t.Errorf("Expected body '%s', got '%s'", jsonBody, string(resp.Body))
	}

	// Test Content-Length
	expectedLength := len(jsonBody)
	if resp.Header.Get("Content-Length") != fmt.Sprintf("%d", expectedLength) {
		t.Errorf("Expected Content-Length '%d', got %s", expectedLength, resp.Header.Get("Content-Length"))
	}
}

func TestResponse_HTMLResponse(t *testing.T) {
	htmlBody := `<html><head><title>Test</title></head><body><h1>Hello World</h1></body></html>`
	resp := &Response{
		Header: Header{"Content-Type": {ContentTypeHTML}},
		Body:   []byte(htmlBody),
	}

	// Test Content-Type
	if resp.Header.Get("Content-Type") != ContentTypeHTML {
		t.Errorf("Expected Content-Type %s, got %s", ContentTypeHTML, resp.Header.Get("Content-Type"))
	}

	// Test body
	if string(resp.Body) != htmlBody {
		t.Errorf("Expected body '%s', got '%s'", htmlBody, string(resp.Body))
	}
}

func TestResponse_PlainTextResponse(t *testing.T) {
	textBody := "Hello, World! This is plain text."
	resp := &Response{
		Header: Header{"Content-Type": {ContentTypeText}},
		Body:   []byte(textBody),
	}

	// Test Content-Type
	if resp.Header.Get("Content-Type") != ContentTypeText {
		t.Errorf("Expected Content-Type %s, got %s", ContentTypeText, resp.Header.Get("Content-Type"))
	}

	// Test body
	if string(resp.Body) != textBody {
		t.Errorf("Expected body '%s', got '%s'", textBody, string(resp.Body))
	}
}

func TestResponse_WithCustomHeaders(t *testing.T) {
	resp := &Response{
		Header: make(Header),
	}

	// Set custom headers
	resp.Header.Set("X-Custom-Header", "custom-value")
	resp.Header.Set("X-Request-ID", "req-123")
	resp.Header.Add("X-Multiple-Values", "value1")
	resp.Header.Add("X-Multiple-Values", "value2")

	// Test custom headers
	if resp.Header.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("Expected X-Custom-Header 'custom-value', got %s", resp.Header.Get("X-Custom-Header"))
	}
	if resp.Header.Get("X-Request-ID") != "req-123" {
		t.Errorf("Expected X-Request-ID 'req-123', got %s", resp.Header.Get("X-Request-ID"))
	}

	// Test multiple values
	values := resp.Header.Values("X-Multiple-Values")
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
	if values[0] != "value1" || values[1] != "value2" {
		t.Errorf("Expected values ['value1', 'value2'], got %v", values)
	}
}
