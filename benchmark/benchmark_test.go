package main

import (
	"encoding/json"
	stdHttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	 "github.com/baxromumarov/http"
)

// Test data structures
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Benchmark: JSON processing performance
func BenchmarkJSONProcessing_OurLibrary(b *testing.B) {
	testUser := User{Name: "John Doe", Age: 30}
	jsonData, _ := http.MarshalJSON(testUser)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test JSON marshaling
			_, err := http.MarshalJSON(testUser)
			if err != nil {
				b.Fatalf("Failed to marshal JSON: %v", err)
			}

			// Test JSON unmarshaling
			var user User
			err = http.UnmarshalJSON(jsonData, &user)
			if err != nil {
				b.Fatalf("Failed to unmarshal JSON: %v", err)
			}
		}
	})
}

func BenchmarkJSONProcessing_OfficialHTTP(b *testing.B) {
	testUser := User{Name: "John Doe", Age: 30}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test JSON marshaling
			jsonData, err := json.Marshal(testUser)
			if err != nil {
				b.Fatalf("Failed to marshal JSON: %v", err)
			}

			// Test JSON unmarshaling
			var user User
			err = json.Unmarshal(jsonData, &user)
			if err != nil {
				b.Fatalf("Failed to unmarshal JSON: %v", err)
			}
		}
	})
}

// Benchmark: Memory allocation comparison
func BenchmarkMemoryAllocation_OurLibrary(b *testing.B) {
	testUser := User{Name: "John Doe", Age: 30}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test JSON marshaling with allocation tracking
			jsonData, err := http.MarshalJSON(testUser)
			if err != nil {
				b.Fatalf("Failed to marshal JSON: %v", err)
			}

			// Test JSON unmarshaling with allocation tracking
			var user User
			err = http.UnmarshalJSON(jsonData, &user)
			if err != nil {
				b.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			// Test response creation
			response := Response{
				Success: true,
				Message: "User created",
				Data:    user,
			}

			_, err = http.MarshalJSON(response)
			if err != nil {
				b.Fatalf("Failed to marshal response: %v", err)
			}
		}
	})
}

func BenchmarkMemoryAllocation_OfficialHTTP(b *testing.B) {
	testUser := User{Name: "John Doe", Age: 30}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test JSON marshaling with allocation tracking
			jsonData, err := json.Marshal(testUser)
			if err != nil {
				b.Fatalf("Failed to marshal JSON: %v", err)
			}

			// Test JSON unmarshaling with allocation tracking
			var user User
			err = json.Unmarshal(jsonData, &user)
			if err != nil {
				b.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			// Test response creation
			response := Response{
				Success: true,
				Message: "User created",
				Data:    user,
			}

			_, err = json.Marshal(response)
			if err != nil {
				b.Fatalf("Failed to marshal response: %v", err)
			}
		}
	})
}

// Benchmark: Middleware chain performance
func BenchmarkMiddlewareChain_OurLibrary(b *testing.B) {
	// Create middleware chain
	loggerMiddleware := http.Logger()
	corsMiddleware := http.CORS()

	// Create a simple handler
	handler := func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{http.ContentTypeJSON}},
			Body:       []byte(`{"message": "Hello, World!"}`),
		}
	}

	// Chain middleware
	chainedHandler := loggerMiddleware(corsMiddleware(handler))

	// Create test request
	req := &http.Request{
		Method: http.GET,
		Path:   "/api/hello",
		Header: make(http.Header),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp := chainedHandler(req)
			if resp.StatusCode != 200 {
				b.Fatalf("Expected status 200, got %d", resp.StatusCode)
			}
		}
	})
}

func BenchmarkMiddlewareChain_OfficialHTTP(b *testing.B) {
	// Setup official http server with middleware simulation
	mux := stdHttp.NewServeMux()
	mux.HandleFunc("/api/hello", func(w stdHttp.ResponseWriter, r *stdHttp.Request) {
		// Add CORS headers (simulating middleware)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Hello, World!"}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := &stdHttp.Client{Timeout: 5 * time.Second}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(server.URL + "/api/hello")
			if err != nil {
				b.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				b.Fatalf("Expected status 200, got %d", resp.StatusCode)
			}
		}
	})
}

// Benchmark: HTTP method constants
func BenchmarkHTTPMethodConstants_OurLibrary(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test method comparisons
			_ = string(http.GET) == "GET"
			_ = string(http.POST) == "POST"
			_ = string(http.PUT) == "PUT"
			_ = string(http.DELETE) == "DELETE"
			_ = string(http.OPTIONS) == "OPTIONS"
		}
	})
}

func BenchmarkHTTPMethodConstants_OfficialHTTP(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test method comparisons
			_ = stdHttp.MethodGet == "GET"
			_ = stdHttp.MethodPost == "POST"
			_ = stdHttp.MethodPut == "PUT"
			_ = stdHttp.MethodDelete == "DELETE"
			_ = stdHttp.MethodOptions == "OPTIONS"
		}
	})
}

// Benchmark: Header operations
func BenchmarkHeaderOperations_OurLibrary(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test header operations
			header := make(http.Header)
			header.Set("Content-Type", http.ContentTypeJSON)
			header.Set("Authorization", "Bearer token")
			header.Set("Accept", "application/json")

			_ = header.Get("Content-Type")
			_ = header.Get("Authorization")
			_ = header.Get("Accept")
		}
	})
}

func BenchmarkHeaderOperations_OfficialHTTP(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test header operations
			header := make(stdHttp.Header)
			header.Set("Content-Type", "application/json")
			header.Set("Authorization", "Bearer token")
			header.Set("Accept", "application/json")

			_ = header.Get("Content-Type")
			_ = header.Get("Authorization")
			_ = header.Get("Accept")
		}
	})
}

// Benchmark: Status code operations
func BenchmarkStatusCodeOperations_OurLibrary(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test status code operations
			_ = http.StatusOK == 200
			_ = http.StatusCreated == 201
			_ = http.StatusBadRequest == 400
			_ = http.StatusNotFound == 404
			_ = http.StatusInternalServerError == 500
		}
	})
}

func BenchmarkStatusCodeOperations_OfficialHTTP(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test status code operations
			_ = stdHttp.StatusOK == 200
			_ = stdHttp.StatusCreated == 201
			_ = stdHttp.StatusBadRequest == 400
			_ = stdHttp.StatusNotFound == 404
			_ = stdHttp.StatusInternalServerError == 500
		}
	})
}

// Benchmark: Request creation
func BenchmarkRequestCreation_OurLibrary(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test request creation
			req, err := http.NewRequest(http.GET, "http://localhost:8080/api/hello", nil)
			if err != nil {
				b.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", http.ContentTypeJSON)
			req.Header.Set("Accept", "application/json")

			_ = req.Method
			_ = req.Path
			_ = req.Header.Get("Content-Type")
		}
	})
}

func BenchmarkRequestCreation_OfficialHTTP(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test request creation
			req, err := stdHttp.NewRequest(stdHttp.MethodGet, "http://localhost:8080/api/hello", nil)
			if err != nil {
				b.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			_ = req.Method
			_ = req.URL.Path
			_ = req.Header.Get("Content-Type")
		}
	})
}

// Benchmark: Response creation
func BenchmarkResponseCreation_OurLibrary(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test response creation
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{http.ContentTypeJSON}},
				Body:       []byte(`{"message": "Hello, World!"}`),
			}

			_ = resp.StatusCode
			_ = resp.Header.Get("Content-Type")
			_ = len(resp.Body)
		}
	})
}

func BenchmarkResponseCreation_OfficialHTTP(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test response creation
			resp := &stdHttp.Response{
				StatusCode: stdHttp.StatusOK,
				Header:     stdHttp.Header{"Content-Type": []string{"application/json"}},
				Body:       nil, // Simplified for comparison
			}

			_ = resp.StatusCode
			_ = resp.Header.Get("Content-Type")
		}
	})
}
