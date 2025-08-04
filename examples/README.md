# HTTP Library Examples

This directory contains examples demonstrating how to use our custom HTTP library for building backend services and clients.

## 🏗️ Project Structure

```
examples/
├── backend/
│   └── main.go          # Backend API server
├── client/
│   └── main.go          # HTTP client demo
└── README.md            # This file
```

## 🚀 Quick Start

### 1. Start the Backend Server

First, start the backend server:

```bash
cd examples/backend
go run main.go
```

You should see output like:
```
🚀 Starting backend server on http://localhost:8080
📚 Available endpoints:
  GET  /api/health     - Health check
  GET  /api/users      - List all users
  GET  /api/users/:id  - Get user by ID
  POST /api/users      - Create new user
  PUT  /api/users/:id  - Update user
  DELETE /api/users/:id - Delete user
  GET  /api/stats      - Get server statistics
```

### 2. Run the Client Demo

In a new terminal, run the client demo:

```bash
cd examples/client
go run main.go
```

The client will demonstrate all the API endpoints:
- Health check
- Get all users
- Get user by ID
- Create new user
- Update user
- Get server statistics
- Delete user

## 📋 API Endpoints

### Health Check
- **GET** `/api/health`
- Returns server health status and timestamp

### Users Management
- **GET** `/api/users` - List all users
- **GET** `/api/users/:id` - Get user by ID
- **POST** `/api/users` - Create new user
- **PUT** `/api/users/:id` - Update user
- **DELETE** `/api/users/:id` - Delete user

### Server Statistics
- **GET** `/api/stats` - Get server statistics

## 🔧 Features Demonstrated

### Backend Features
- ✅ **Simplified HandlerFunc**: Clean, simple handler functions
- ✅ **Middleware Support**: Logger, Recover, and CORS middleware
- ✅ **Path Parameters**: Dynamic route parameters (`:id`)
- ✅ **JSON Handling**: Request/response JSON marshaling
- ✅ **Error Handling**: Proper HTTP status codes and error responses
- ✅ **In-Memory Storage**: Simple data persistence

### Client Features
- ✅ **HTTP Client**: Making requests to the backend
- ✅ **Request Building**: Creating HTTP requests with different methods
- ✅ **Response Handling**: Parsing JSON responses
- ✅ **Error Handling**: Graceful error handling and logging
- ✅ **Complete CRUD**: Create, Read, Update, Delete operations

## 📝 Example Usage

### Backend Handler Example
```go
http.Handle(http.GET, "/api/users/:id", func(req *http.Request) *http.Response {
    idStr := req.PathValue("id")
    // ... handle request
    return &http.Response{
        StatusCode: 200,
        Header:     http.Header{"Content-Type": {http.ContentTypeJSON}},
        Body:       jsonData,
    }
})
```

### Client Request Example
```go
req, err := http.NewRequest(http.GET, "http://localhost:8080/api/users", nil)
resp, err := client.Send(req)
var response Response
resp.Unmarshal(&response)
```

## 🛠️ Building and Running

### Prerequisites
- Go 1.19 or later
- Our HTTP library installed

### Build Backend
```bash
cd examples/backend
go build -o backend main.go
./backend
```

### Build Client
```bash
cd examples/client
go build -o client main.go
./client
```

## 🧪 Testing the API

You can also test the API using curl:

```bash
# Health check
curl http://localhost:8080/api/health

# Get all users
curl http://localhost:8080/api/users

# Get user by ID
curl http://localhost:8080/api/users/1

# Create new user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","age":25}'

# Update user
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Name"}'

# Delete user
curl -X DELETE http://localhost:8080/api/users/1

# Get server stats
curl http://localhost:8080/api/stats
```

## 🔍 What You'll Learn

This example demonstrates:

1. **Server Setup**: How to create and configure an HTTP server
2. **Route Registration**: How to register handlers for different HTTP methods and paths
3. **Middleware Usage**: How to add logging, recovery, and CORS middleware
4. **Request Handling**: How to handle different types of requests (GET, POST, PUT, DELETE)
5. **Path Parameters**: How to extract and use dynamic route parameters
6. **JSON Processing**: How to marshal/unmarshal JSON data
7. **Error Handling**: How to return appropriate HTTP status codes and error messages
8. **Client Usage**: How to create HTTP clients and make requests
9. **Response Processing**: How to handle and parse API responses

## 🎯 Next Steps

After running these examples, you can:

1. **Extend the API**: Add more endpoints and functionality
2. **Add Database**: Replace in-memory storage with a real database
3. **Add Authentication**: Implement user authentication and authorization
4. **Add Validation**: Add input validation and sanitization
5. **Add Testing**: Write unit and integration tests
6. **Add Documentation**: Generate API documentation
7. **Add Monitoring**: Add metrics and monitoring

This example provides a solid foundation for building real-world applications with our HTTP library! 