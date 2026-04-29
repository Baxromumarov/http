# HTTP-Go: A Simple and Fast HTTP Library for Go

HTTP library built from scratch with zero external (for learning purpose)
[![Go Version](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-Passing-brightgreen.svg)](https://github.com/baxromumarov/http-go)


##  Quick Start

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


## API Reference

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

## Middleware

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
