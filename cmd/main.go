package main

import (
	"fmt"
	"github.com/baxromumarov/http-go" // Replace with your actual module path
	"log"
)

func main() {
	// Create a new server instance
	server := http_go.NewDefaultServer("", 8080)

	// Add global middleware (applies to all routes)
	server.Use(http_go.Logger())  // Log all requests
	server.Use(http_go.Recover()) // Recover from panics
	server.Use(http_go.CORS())    // Enable CORS

	// Public routes
	server.GET("/", func(req *http_go.Request, params map[string]string, next http_go.HandlerFunc) *http_go.Response {
		return &http_go.Response{
			StatusCode: 200,
			Header:     http_go.Header{"Content-Type": {"application/json"}},
			Body:       []byte(`{"message": "Welcome to the API"}`),
		}
	})

	server.GET("/public", func(req *http_go.Request, params map[string]string, next http_go.HandlerFunc) *http_go.Response {
		return &http_go.Response{
			StatusCode: 200,
			Header:     http_go.Header{"Content-Type": {"application/json"}},
			Body:       []byte(`{"message": "This is a public endpoint"}`),
		}
	})

	// Protected routes
	server.GET(
		"/protected",
		func(req *http_go.Request, params map[string]string, next http_go.HandlerFunc) *http_go.Response {
			// Get username from context (set by BasicAuth middleware)
			username := req.Context().Value("username").(string)

			return &http_go.Response{
				StatusCode: 200,
				Header:     http_go.Header{"Content-Type": {"application/json"}},
				Body:       []byte(`{"message": "Hello, ` + username + `! This is a protected endpoint"}`),
			}
		})

	// Example of a POST endpoint with JSON response
	server.POST("/echo", func(req *http_go.Request, params map[string]string, next http_go.HandlerFunc) *http_go.Response {
		// Access request body
		body := string(req.Body)
		fmt.Println("LENGTH: ", len(body))
		return &http_go.Response{
			StatusCode: 200,
			Header:     http_go.Header{"Content-Type": {"application/json"}},
			Body:       []byte(`{"echo": "` + body + `"}`),
		}
	})

	// Start the server
	log.Println("Server starting on :8080")
	if err := server.StartServer(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
