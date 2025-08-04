# 🚀 HTTP-Go: A Simple and Fast HTTP Library for Go

A lightweight, high-performance HTTP library for Go that provides a clean and simple API for building HTTP servers and clients. Built from scratch with zero external dependencies.

[![Go Version](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-Passing-brightgreen.svg)](https://github.com/baxromumarov/http-go)

## ✨ Features

- 🎯 **Simple API**: Clean, intuitive handler functions
- ⚡ **High Performance**: Built for speed and efficiency
- 🔧 **Middleware Support**: Logger, Recover, CORS, and BasicAuth
- 🛣️ **Dynamic Routing**: Path parameters and flexible route matching
- 📦 **Zero Dependencies**: No external packages required
- 🔒 **Type Safe**: Full Go type safety and compile-time checks
- 🧪 **Well Tested**: Comprehensive test coverage
- 📚 **Complete Examples**: Backend and client examples included

## 🚀 Quick Start

### Installation

```bash
go get github.com/baxromumarov/http-go
```

### Simple Server Example

```go
package main

import (
    "fmt"
    "github.com/baxromumarov/http-go"
)

func main() {
    // Create a new server
    server := http.NewDefaultServer("localhost", 8080)
    
    // Add middleware
    server.Use(http.Logger())
    server.Use(http.CORS())
    
    // Register routes
    http.Handle(http.GET, "/api/hello", func(req *http.Request) *http.Response {
        return &http.Response{
            StatusCode: 200,
            Header:     http.Header{"Content-Type": {http.ContentTypeJSON}},
            Body:       []byte(`{"message": "Hello, World!"}`),
        }
    })
    
    // Start the server
    fmt.Println("🚀 Server starting on http://localhost:8080")
    server.StartServer()
}
```

### Simple Client Example

```go
package main

import (
    "fmt"
    "github.com/baxromumarov/http-go"
)

func main() {
    // Create HTTP client
    client := &http.Client{Timeout: 10 * time.Second}
    
    // Create request
    req, err := http.NewRequest(http.GET, "http://localhost:8080/api/hello", nil)
    if err != nil {
        panic(err)
    }
    
    // Send request
    resp, err := client.Send(req)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Response: %s\n", string(resp.Body))
}
```

## 📚 API Reference

### Server

#### Creating a Server

```go
// Create a new server with default settings
server := http.NewDefaultServer("localhost", 8080)

// Create a custom server
server := &http.Server{
    Host: "0.0.0.0",
    Port: 3000,
    WriteTimeout: 30 * time.Second,
    ReadTimeout:  30 * time.Second,
}
```

#### Registering Routes

```go
// GET request
http.Handle(http.GET, "/api/users", func(req *http.Request) *http.Response {
    return &http.Response{
        StatusCode: 200,
        Body: []byte("Users list"),
    }
})

// POST request with JSON
http.Handle(http.POST, "/api/users", func(req *http.Request) *http.Response {
    var user User
    http.UnmarshalJSON(req.Body, &user)
    
    // Process user...
    
    return &http.Response{
        StatusCode: 201,
        Header: http.Header{"Content-Type": {http.ContentTypeJSON}},
        Body: jsonData,
    }
})

// Path parameters
http.Handle(http.GET, "/api/users/:id", func(req *http.Request) *http.Response {
    id := req.PathValue("id")
    // Use id parameter...
})
```

#### Middleware

```go
// Add middleware to server
server.Use(http.Logger())
server.Use(http.Recover())
server.Use(http.CORS())
server.Use(http.BasicAuth(users))

// Custom middleware
func CustomMiddleware() http.MiddlewareFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(req *http.Request) *http.Response {
            // Pre-processing
            fmt.Printf("Request: %s %s\n", req.Method, req.Path)
            
            // Call next handler
            resp := next(req)
            
            // Post-processing
            fmt.Printf("Response: %d\n", resp.StatusCode)
            
            return resp
        }
    }
}
```

### Client

#### Making Requests

```go
client := &http.Client{Timeout: 10 * time.Second}

// GET request
req, err := http.NewRequest(http.GET, "http://api.example.com/users", nil)
resp, err := client.Send(req)

// POST request with JSON
jsonData, _ := http.MarshalJSON(user)
req, err := http.NewRequest(http.POST, "http://api.example.com/users", jsonData)
req.Header.Set("Content-Type", http.ContentTypeJSON)
resp, err := client.Send(req)

// Parse JSON response
var response Response
resp.Unmarshal(&response)
```

## 🏗️ Examples

### Complete Backend and Client

Check out the `examples/` directory for a complete REST API implementation:

```bash
# Start the backend server
cd examples/backend
go run main.go

# In another terminal, run the client demo
cd examples/client
go run main.go
```

The examples demonstrate:
- ✅ Complete CRUD operations
- ✅ Middleware usage
- ✅ Path parameters
- ✅ JSON handling
- ✅ Error handling
- ✅ Client-server communication

### Available Endpoints

- `GET /api/health` - Health check
- `GET /api/users` - List all users
- `GET /api/users/:id` - Get user by ID
- `POST /api/users` - Create new user
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user
- `GET /api/stats` - Get server statistics

## 🔧 Middleware

### Built-in Middleware

#### Logger
```go
server.Use(http.Logger())
```
Logs all requests with timing information.

#### Recover
```go
server.Use(http.Recover())
```
Recovers from panics and returns 500 errors.

#### CORS
```go
server.Use(http.CORS())
```
Adds CORS headers to all responses.

#### BasicAuth
```go
users := map[string]string{"admin": "password"}
server.Use(http.BasicAuth(users))
```
Adds HTTP Basic Authentication.

## 🧪 Testing

Run the test suite:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```

## 📊 Performance

This library is designed for high performance and has been benchmarked against the official Go `net/http` package:

### 🏆 **Benchmark Results vs Official Go HTTP**

| Category | HTTP-Go | Official HTTP | Improvement |
|----------|---------|---------------|-------------|
| **JSON Processing** | 110,719 ns/op | 120,610 ns/op | **8.2% faster** |
| **Memory Usage** | 659 B/op | 28,511 B/op | **97.7% less memory** |
| **Header Operations** | 17.21 ns/op | 28.64 ns/op | **39.9% faster** |
| **Response Creation** | 6.435 ns/op | 8.587 ns/op | **25.1% faster** |
| **Allocation Count** | 14 allocs/op | 158 allocs/op | **91.1% fewer allocations** |

### 🚀 **Performance Highlights**

- **Memory Efficiency**: Up to **97.7% less memory usage**
- **Faster Processing**: Up to **39.9% faster** in key operations
- **Fewer Allocations**: Up to **91.1% fewer allocations**
- **Zero-Cost Abstractions**: Minimal overhead for constants and basic operations

### 📈 **Detailed Results**

For comprehensive benchmark results, see [BENCHMARK_RESULTS.md](benchmark/BENCHMARK_RESULTS.md).

**🏆 HTTP-Go wins in 13 out of 18 performance categories!**

### 🧪 **Running Benchmarks**

Run the benchmark suite:

```bash
# Quick benchmark summary
./benchmark/run_benchmarks.sh

# Detailed benchmark results
go test ./benchmark -bench=. -benchmem

# Run specific benchmarks
go test ./benchmark -bench=BenchmarkJSONProcessing -benchmem
```