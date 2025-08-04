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
    server := http_go.NewDefaultServer("localhost", 8080)
    
    // Add middleware
    server.Use(http_go.Logger())
    server.Use(http_go.CORS())
    
    // Register routes
    http_go.Handle(http_go.GET, "/api/hello", func(req *http_go.Request) *http_go.Response {
        return &http_go.Response{
            StatusCode: 200,
            Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
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
    client := &http_go.Client{Timeout: 10 * time.Second}
    
    // Create request
    req, err := http_go.NewRequest(http_go.GET, "http://localhost:8080/api/hello", nil)
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
server := http_go.NewDefaultServer("localhost", 8080)

// Create a custom server
server := &http_go.Server{
    Host: "0.0.0.0",
    Port: 3000,
    WriteTimeout: 30 * time.Second,
    ReadTimeout:  30 * time.Second,
}
```

#### Registering Routes

```go
// GET request
http_go.Handle(http_go.GET, "/api/users", func(req *http_go.Request) *http_go.Response {
    return &http_go.Response{
        StatusCode: 200,
        Body: []byte("Users list"),
    }
})

// POST request with JSON
http_go.Handle(http_go.POST, "/api/users", func(req *http_go.Request) *http_go.Response {
    var user User
    http_go.UnmarshalJSON(req.Body, &user)
    
    // Process user...
    
    return &http_go.Response{
        StatusCode: 201,
        Header: http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
        Body: jsonData,
    }
})

// Path parameters
http_go.Handle(http_go.GET, "/api/users/:id", func(req *http_go.Request) *http_go.Response {
    id := req.PathValue("id")
    // Use id parameter...
})
```

#### Middleware

```go
// Add middleware to server
server.Use(http_go.Logger())
server.Use(http_go.Recover())
server.Use(http_go.CORS())
server.Use(http_go.BasicAuth(users))

// Custom middleware
func CustomMiddleware() http_go.MiddlewareFunc {
    return func(next http_go.HandlerFunc) http_go.HandlerFunc {
        return func(req *http_go.Request) *http_go.Response {
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
client := &http_go.Client{Timeout: 10 * time.Second}

// GET request
req, err := http_go.NewRequest(http_go.GET, "http://api.example.com/users", nil)
resp, err := client.Send(req)

// POST request with JSON
jsonData, _ := http_go.MarshalJSON(user)
req, err := http_go.NewRequest(http_go.POST, "http://api.example.com/users", jsonData)
req.Header.Set("Content-Type", http_go.ContentTypeJSON)
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
server.Use(http_go.Logger())
```
Logs all requests with timing information.

#### Recover
```go
server.Use(http_go.Recover())
```
Recovers from panics and returns 500 errors.

#### CORS
```go
server.Use(http_go.CORS())
```
Adds CORS headers to all responses.

#### BasicAuth
```go
users := map[string]string{"admin": "password"}
server.Use(http_go.BasicAuth(users))
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

This library is designed for high performance:

- **Zero Allocations**: Minimal memory allocations
- **Fast Routing**: Efficient route matching
- **Lightweight**: Small memory footprint
- **Concurrent**: Thread-safe design

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with ❤️ and Go
- Inspired by modern web frameworks
- Designed for simplicity and performance

## 📞 Support

If you have any questions or need help, please:

1. Check the [examples](examples/) directory
2. Review the test files for usage patterns
3. Open an issue on GitHub

---

**Made with ❤️ by the HTTP-Go community** 