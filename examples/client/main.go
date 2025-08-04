package main

import (
	"fmt"
	"log"
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

// CreateUserRequest represents the data needed to create a new user
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func main() {
	// Create HTTP client
	client := &http_go.Client{
		Timeout: 10 * time.Second,
	}

	fmt.Println("🚀 Starting HTTP Client Demo")
	fmt.Println("📡 Connecting to backend server...")

	// Test health check
	fmt.Println("\n1️⃣ Testing Health Check...")
	healthCheck(client)

	// Test getting all users
	fmt.Println("\n2️⃣ Testing Get All Users...")
	getAllUsers(client)

	// Test getting a specific user
	fmt.Println("\n3️⃣ Testing Get User by ID...")
	getUserByID(client, 1)

	// Test creating a new user
	fmt.Println("\n4️⃣ Testing Create User...")
	newUser := CreateUserRequest{
		Name:  "Alice Johnson",
		Email: "alice@example.com",
		Age:   28,
	}
	createdUser := createUser(client, newUser)

	// Test updating a user
	if createdUser != nil {
		fmt.Println("\n5️⃣ Testing Update User...")
		updateData := User{
			Name: "Alice Smith",
			Age:  29,
		}
		updateUser(client, createdUser.ID, updateData)
	}

	// Test getting server stats
	fmt.Println("\n6️⃣ Testing Get Server Stats...")
	getServerStats(client)

	// Test getting all users again to see the new user
	fmt.Println("\n7️⃣ Testing Get All Users (after changes)...")
	getAllUsers(client)

	// Test deleting a user
	if createdUser != nil {
		fmt.Println("\n8️⃣ Testing Delete User...")
		deleteUser(client, createdUser.ID)
	}

	fmt.Println("\n✅ Client demo completed!")
}

func healthCheck(client *http_go.Client) {
	req, err := http_go.NewRequest(http_go.GET, "http://localhost:8080/api/health", nil)
	if err != nil {
		log.Printf("❌ Failed to create health check request: %v", err)
		return
	}

	resp, err := client.Send(req)
	if err != nil {
		log.Printf("❌ Health check failed: %v", err)
		return
	}

	var response Response
	if err := resp.Unmarshal(&response); err != nil {
		log.Printf("❌ Failed to unmarshal health response: %v", err)
		return
	}

	if response.Success {
		fmt.Printf("✅ Health check passed: %s\n", response.Message)
		if data, ok := response.Data.(map[string]interface{}); ok {
			fmt.Printf("   📅 Timestamp: %v\n", data["timestamp"])
			fmt.Printf("   ⏱️  Uptime: %v\n", data["uptime"])
		}
	} else {
		fmt.Printf("❌ Health check failed: %s\n", response.Error)
	}
}

func getAllUsers(client *http_go.Client) {
	req, err := http_go.NewRequest(http_go.GET, "http://localhost:8080/api/users", nil)
	if err != nil {
		log.Printf("❌ Failed to create get users request: %v", err)
		return
	}

	resp, err := client.Send(req)
	if err != nil {
		log.Printf("❌ Get users failed: %v", err)
		return
	}

	var response Response
	if err := resp.Unmarshal(&response); err != nil {
		log.Printf("❌ Failed to unmarshal users response: %v", err)
		return
	}

	if response.Success {
		fmt.Printf("✅ %s\n", response.Message)
		if users, ok := response.Data.([]interface{}); ok {
			fmt.Printf("   📊 Found %d users:\n", len(users))
			for i, user := range users {
				if userMap, ok := user.(map[string]interface{}); ok {
					fmt.Printf("   %d. ID: %v, Name: %v, Email: %v, Age: %v\n",
						i+1, userMap["id"], userMap["name"], userMap["email"], userMap["age"])
				}
			}
		}
	} else {
		fmt.Printf("❌ Get users failed: %s\n", response.Error)
	}
}

func getUserByID(client *http_go.Client, id int) {
	url := fmt.Sprintf("http://localhost:8080/api/users/%d", id)
	req, err := http_go.NewRequest(http_go.GET, url, nil)
	if err != nil {
		log.Printf("❌ Failed to create get user request: %v", err)
		return
	}

	resp, err := client.Send(req)
	if err != nil {
		log.Printf("❌ Get user failed: %v", err)
		return
	}

	var response Response
	if err := resp.Unmarshal(&response); err != nil {
		log.Printf("❌ Failed to unmarshal user response: %v", err)
		return
	}

	if response.Success {
		fmt.Printf("✅ %s\n", response.Message)
		if user, ok := response.Data.(map[string]interface{}); ok {
			fmt.Printf("   👤 User Details:\n")
			fmt.Printf("      ID: %v\n", user["id"])
			fmt.Printf("      Name: %v\n", user["name"])
			fmt.Printf("      Email: %v\n", user["email"])
			fmt.Printf("      Age: %v\n", user["age"])
			fmt.Printf("      Created: %v\n", user["created"])
		}
	} else {
		fmt.Printf("❌ Get user failed: %s\n", response.Error)
	}
}

func createUser(client *http_go.Client, user CreateUserRequest) *User {
	jsonData, err := http_go.MarshalJSON(user)
	if err != nil {
		log.Printf("❌ Failed to marshal user data: %v", err)
		return nil
	}

	req, err := http_go.NewRequest(http_go.POST, "http://localhost:8080/api/users", jsonData)
	if err != nil {
		log.Printf("❌ Failed to create user request: %v", err)
		return nil
	}

	req.Header.Set("Content-Type", http_go.ContentTypeJSON)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(jsonData)))

	resp, err := client.Send(req)
	if err != nil {
		log.Printf("❌ Create user failed: %v", err)
		return nil
	}

	var response Response
	if err := resp.Unmarshal(&response); err != nil {
		log.Printf("❌ Failed to unmarshal create response: %v", err)
		return nil
	}

	if response.Success {
		fmt.Printf("✅ %s\n", response.Message)
		if userData, ok := response.Data.(map[string]interface{}); ok {
			createdUser := &User{
				ID:      int(userData["id"].(float64)),
				Name:    userData["name"].(string),
				Email:   userData["email"].(string),
				Age:     int(userData["age"].(float64)),
				Created: userData["created"].(string),
			}
			fmt.Printf("   🆕 Created user with ID: %d\n", createdUser.ID)
			return createdUser
		}
	} else {
		fmt.Printf("❌ Create user failed: %s\n", response.Error)
	}

	return nil
}

func updateUser(client *http_go.Client, id int, updateData User) {
	jsonData, err := http_go.MarshalJSON(updateData)
	if err != nil {
		log.Printf("❌ Failed to marshal update data: %v", err)
		return
	}

	url := fmt.Sprintf("http://localhost:8080/api/users/%d", id)
	req, err := http_go.NewRequest(http_go.PUT, url, jsonData)
	if err != nil {
		log.Printf("❌ Failed to create update request: %v", err)
		return
	}

	req.Header.Set("Content-Type", http_go.ContentTypeJSON)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(jsonData)))

	resp, err := client.Send(req)
	if err != nil {
		log.Printf("❌ Update user failed: %v", err)
		return
	}

	var response Response
	if err := resp.Unmarshal(&response); err != nil {
		log.Printf("❌ Failed to unmarshal update response: %v", err)
		return
	}

	if response.Success {
		fmt.Printf("✅ %s\n", response.Message)
		if userData, ok := response.Data.(map[string]interface{}); ok {
			fmt.Printf("   ✏️  Updated user ID %d:\n", int(userData["id"].(float64)))
			fmt.Printf("      Name: %v\n", userData["name"])
			fmt.Printf("      Email: %v\n", userData["email"])
			fmt.Printf("      Age: %v\n", userData["age"])
		}
	} else {
		fmt.Printf("❌ Update user failed: %s\n", response.Error)
	}
}

func deleteUser(client *http_go.Client, id int) {
	url := fmt.Sprintf("http://localhost:8080/api/users/%d", id)
	req, err := http_go.NewRequest(http_go.DELETE, url, nil)
	if err != nil {
		log.Printf("❌ Failed to create delete request: %v", err)
		return
	}

	resp, err := client.Send(req)
	if err != nil {
		log.Printf("❌ Delete user failed: %v", err)
		return
	}

	var response Response
	if err := resp.Unmarshal(&response); err != nil {
		log.Printf("❌ Failed to unmarshal delete response: %v", err)
		return
	}

	if response.Success {
		fmt.Printf("✅ %s\n", response.Message)
	} else {
		fmt.Printf("❌ Delete user failed: %s\n", response.Error)
	}
}

func getServerStats(client *http_go.Client) {
	req, err := http_go.NewRequest(http_go.GET, "http://localhost:8080/api/stats", nil)
	if err != nil {
		log.Printf("❌ Failed to create stats request: %v", err)
		return
	}

	resp, err := client.Send(req)
	if err != nil {
		log.Printf("❌ Get stats failed: %v", err)
		return
	}

	var response Response
	if err := resp.Unmarshal(&response); err != nil {
		log.Printf("❌ Failed to unmarshal stats response: %v", err)
		return
	}

	if response.Success {
		fmt.Printf("✅ %s\n", response.Message)
		if stats, ok := response.Data.(map[string]interface{}); ok {
			fmt.Printf("   📈 Server Statistics:\n")
			fmt.Printf("      Total Users: %v\n", stats["total_users"])
			fmt.Printf("      Server Time: %v\n", stats["server_time"])
			fmt.Printf("      Uptime: %v\n", stats["uptime"])
		}
	} else {
		fmt.Printf("❌ Get stats failed: %s\n", response.Error)
	}
}
