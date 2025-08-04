package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	http_go "github.com/baxromumarov/http-go"
)

// User represents a user in our system
type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Age     int    `json:"age"`
	Created string `json:"created"`
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// In-memory storage for users
var users = map[int]User{
	1: {ID: 1, Name: "John Doe", Email: "john@example.com", Age: 30, Created: time.Now().Format(time.RFC3339)},
	2: {ID: 2, Name: "Jane Smith", Email: "jane@example.com", Age: 25, Created: time.Now().Format(time.RFC3339)},
	3: {ID: 3, Name: "Bob Johnson", Email: "bob@example.com", Age: 35, Created: time.Now().Format(time.RFC3339)},
}

var nextID = 4

func main() {
	// Create a new server
	server := http_go.NewDefaultServer("localhost", 8080)

	// Add middleware
	server.Use(http_go.Logger())
	server.Use(http_go.Recover())
	server.Use(http_go.CORS())

	// Register routes
	registerRoutes()

	fmt.Println("🚀 Starting backend server on http://localhost:8080")
	fmt.Println("📚 Available endpoints:")
	fmt.Println("  GET  /api/health     - Health check")
	fmt.Println("  GET  /api/users      - List all users")
	fmt.Println("  GET  /api/users/:id  - Get user by ID")
	fmt.Println("  POST /api/users      - Create new user")
	fmt.Println("  PUT  /api/users/:id  - Update user")
	fmt.Println("  DELETE /api/users/:id - Delete user")
	fmt.Println("  GET  /api/stats      - Get server statistics")

	// Start the server
	if err := server.StartServer(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func registerRoutes() {
	// Health check endpoint
	http_go.Handle(http_go.GET, "/api/health", func(req *http_go.Request) *http_go.Response {
		response := Response{
			Success: true,
			Message: "Server is healthy",
			Data: map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"uptime":    "running",
			},
		}

		jsonData, _ := http_go.MarshalJSON(response)
		return &http_go.Response{
			StatusCode: 200,
			Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
			Body:       jsonData,
		}
	})

	// Get all users
	http_go.Handle(http_go.GET, "/api/users", func(req *http_go.Request) *http_go.Response {
		userList := make([]User, 0, len(users))
		for _, user := range users {
			userList = append(userList, user)
		}

		response := Response{
			Success: true,
			Message: fmt.Sprintf("Found %d users", len(userList)),
			Data:    userList,
		}

		jsonData, _ := http_go.MarshalJSON(response)
		return &http_go.Response{
			StatusCode: 200,
			Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
			Body:       jsonData,
		}
	})

	// Get user by ID
	http_go.Handle(http_go.GET, "/api/users/:id", func(req *http_go.Request) *http_go.Response {
		idStr := req.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			response := Response{
				Success: false,
				Error:   "Invalid user ID",
			}
			jsonData, _ := http_go.MarshalJSON(response)
			return &http_go.Response{
				StatusCode: 400,
				Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
				Body:       jsonData,
			}
		}

		user, exists := users[id]
		if !exists {
			response := Response{
				Success: false,
				Error:   fmt.Sprintf("User with ID %d not found", id),
			}
			jsonData, _ := http_go.MarshalJSON(response)
			return &http_go.Response{
				StatusCode: 404,
				Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
				Body:       jsonData,
			}
		}

		response := Response{
			Success: true,
			Message: "User found",
			Data:    user,
		}

		jsonData, _ := http_go.MarshalJSON(response)
		return &http_go.Response{
			StatusCode: 200,
			Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
			Body:       jsonData,
		}
	})

	// Create new user
	http_go.Handle(http_go.POST, "/api/users", func(req *http_go.Request) *http_go.Response {
		var newUser User
		if err := http_go.UnmarshalJSON(req.Body, &newUser); err != nil {
			response := Response{
				Success: false,
				Error:   "Invalid JSON data",
			}
			jsonData, _ := http_go.MarshalJSON(response)
			return &http_go.Response{
				StatusCode: 400,
				Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
				Body:       jsonData,
			}
		}

		// Validate required fields
		if newUser.Name == "" || newUser.Email == "" {
			response := Response{
				Success: false,
				Error:   "Name and email are required",
			}
			jsonData, _ := http_go.MarshalJSON(response)
			return &http_go.Response{
				StatusCode: 400,
				Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
				Body:       jsonData,
			}
		}

		// Assign new ID and creation time
		newUser.ID = nextID
		newUser.Created = time.Now().Format(time.RFC3339)
		users[nextID] = newUser
		nextID++

		response := Response{
			Success: true,
			Message: "User created successfully",
			Data:    newUser,
		}

		jsonData, _ := http_go.MarshalJSON(response)
		return &http_go.Response{
			StatusCode: 201,
			Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
			Body:       jsonData,
		}
	})

	// Update user
	http_go.Handle(http_go.PUT, "/api/users/:id", func(req *http_go.Request) *http_go.Response {
		idStr := req.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			response := Response{
				Success: false,
				Error:   "Invalid user ID",
			}
			jsonData, _ := http_go.MarshalJSON(response)
			return &http_go.Response{
				StatusCode: 400,
				Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
				Body:       jsonData,
			}
		}

		// Check if user exists
		if _, exists := users[id]; !exists {
			response := Response{
				Success: false,
				Error:   fmt.Sprintf("User with ID %d not found", id),
			}
			jsonData, _ := http_go.MarshalJSON(response)
			return &http_go.Response{
				StatusCode: 404,
				Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
				Body:       jsonData,
			}
		}

		var updateData User
		if err := http_go.UnmarshalJSON(req.Body, &updateData); err != nil {
			response := Response{
				Success: false,
				Error:   "Invalid JSON data",
			}
			jsonData, _ := http_go.MarshalJSON(response)
			return &http_go.Response{
				StatusCode: 400,
				Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
				Body:       jsonData,
			}
		}

		// Update user data
		existingUser := users[id]
		if updateData.Name != "" {
			existingUser.Name = updateData.Name
		}
		if updateData.Email != "" {
			existingUser.Email = updateData.Email
		}
		if updateData.Age > 0 {
			existingUser.Age = updateData.Age
		}
		users[id] = existingUser

		response := Response{
			Success: true,
			Message: "User updated successfully",
			Data:    existingUser,
		}

		jsonData, _ := http_go.MarshalJSON(response)
		return &http_go.Response{
			StatusCode: 200,
			Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
			Body:       jsonData,
		}
	})

	// Delete user
	http_go.Handle(http_go.DELETE, "/api/users/:id", func(req *http_go.Request) *http_go.Response {
		idStr := req.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			response := Response{
				Success: false,
				Error:   "Invalid user ID",
			}
			jsonData, _ := http_go.MarshalJSON(response)
			return &http_go.Response{
				StatusCode: 400,
				Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
				Body:       jsonData,
			}
		}

		// Check if user exists
		if _, exists := users[id]; !exists {
			response := Response{
				Success: false,
				Error:   fmt.Sprintf("User with ID %d not found", id),
			}
			jsonData, _ := http_go.MarshalJSON(response)
			return &http_go.Response{
				StatusCode: 404,
				Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
				Body:       jsonData,
			}
		}

		// Delete user
		delete(users, id)

		response := Response{
			Success: true,
			Message: fmt.Sprintf("User with ID %d deleted successfully", id),
		}

		jsonData, _ := http_go.MarshalJSON(response)
		return &http_go.Response{
			StatusCode: 200,
			Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
			Body:       jsonData,
		}
	})

	// Get server statistics
	http_go.Handle(http_go.GET, "/api/stats", func(req *http_go.Request) *http_go.Response {
		stats := map[string]interface{}{
			"total_users": len(users),
			"server_time": time.Now().Format(time.RFC3339),
			"uptime":      "running",
		}

		response := Response{
			Success: true,
			Message: "Server statistics",
			Data:    stats,
		}

		jsonData, _ := http_go.MarshalJSON(response)
		return &http_go.Response{
			StatusCode: 200,
			Header:     http_go.Header{"Content-Type": {http_go.ContentTypeJSON}},
			Body:       jsonData,
		}
	})
}
