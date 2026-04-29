package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"

	http "github.com/baxromumarov/http"
)

// User represents a user in our system.
type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Age     int    `json:"age"`
	Created string `json:"created"`
}

// Response represents a standard API response.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type api struct {
	store *userStore
}

type userStore struct {
	mu     sync.RWMutex
	users  map[int]User
	nextID int
}

func main() {
	server := http.NewDefaultServer("localhost", 8080)
	server.Use(
		http.Logger(),
		http.Recover(),
		http.CORS(),
	)

	app := &api{
		store: newUserStore(seedUsers()),
	}

	registerRoutes(server, app)

	if err := server.StartServer(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func registerRoutes(server *http.Server, app *api) {
	server.GET("/v1/health", app.health)
	server.GET("/v1/users", app.listUsers)
	server.GET("/v1/users/:id", app.getUser)
	server.POST("/v1/users", app.createUser)
	server.PUT("/v1/users/:id", app.updateUser)
	server.DELETE("/v1/users/:id", app.deleteUser)
	server.GET("/v1/stats", app.stats)
}

func (a *api) health(req *http.Request) *http.Response {
	return ok("Server is healthy", map[string]any{
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    "running",
	})
}

func (a *api) listUsers(req *http.Request) *http.Response {
	users := a.store.list()

	return ok(
		fmt.Sprintf("Found %d users", len(users)),
		users,
	)
}

func (a *api) getUser(req *http.Request) *http.Response {
	id, resp := userIDFromRequest(req)
	if resp != nil {
		return resp
	}

	user, found := a.store.get(id)
	if !found {
		return userNotFound(id)
	}

	return ok("User found", user)
}

func (a *api) createUser(req *http.Request) *http.Response {
	var user User
	if err := http.UnmarshalJSON(req.Body, &user); err != nil {
		return badRequest("Invalid JSON data")
	}

	if err := validateUser(user); err != nil {
		return badRequest(err.Error())
	}

	createdUser := a.store.create(user)
	return jsonResponse(http.StatusCreated, Response{
		Success: true,
		Message: "User created successfully",
		Data:    createdUser,
	})
}

func (a *api) updateUser(req *http.Request) *http.Response {
	id, resp := userIDFromRequest(req)
	if resp != nil {
		return resp
	}

	var update User
	if err := http.UnmarshalJSON(req.Body, &update); err != nil {
		return badRequest("Invalid JSON data")
	}

	updatedUser, found := a.store.update(id, update)
	if !found {
		return userNotFound(id)
	}

	return ok("User updated successfully", updatedUser)
}

func (a *api) deleteUser(req *http.Request) *http.Response {
	id, resp := userIDFromRequest(req)
	if resp != nil {
		return resp
	}

	if deleted := a.store.delete(id); !deleted {
		return userNotFound(id)
	}

	return ok(
		fmt.Sprintf("User with ID %d deleted successfully", id),
		nil,
	)
}

func (a *api) stats(req *http.Request) *http.Response {
	return ok("Server statistics", map[string]any{
		"total_users": a.store.count(),
		"server_time": time.Now().Format(time.RFC3339),
		"uptime":      "running",
	})
}

func seedUsers() map[int]User {
	now := time.Now().Format(time.RFC3339)

	return map[int]User{
		1: {
			ID:      1,
			Name:    "John Doe",
			Email:   "john@example.com",
			Age:     30,
			Created: now,
		},
		2: {
			ID:      2,
			Name:    "Jane Smith",
			Email:   "jane@example.com",
			Age:     25,
			Created: now,
		},
		3: {
			ID:      3,
			Name:    "Bob Johnson",
			Email:   "bob@example.com",
			Age:     35,
			Created: now,
		},
	}
}

func newUserStore(users map[int]User) *userStore {
	nextID := 1
	for id := range users {
		if id >= nextID {
			nextID = id + 1
		}
	}

	return &userStore{
		users:  users,
		nextID: nextID,
	}
}

func (s *userStore) list() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	return users
}

func (s *userStore) get(id int) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, found := s.users[id]
	return user, found
}

func (s *userStore) create(user User) User {
	s.mu.Lock()
	defer s.mu.Unlock()

	user.ID = s.nextID
	user.Created = time.Now().Format(time.RFC3339)
	s.users[user.ID] = user
	s.nextID++

	return user
}

func (s *userStore) update(id int, update User) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, found := s.users[id]
	if !found {
		return User{}, false
	}

	if update.Name != "" {
		user.Name = update.Name
	}
	if update.Email != "" {
		user.Email = update.Email
	}
	if update.Age > 0 {
		user.Age = update.Age
	}

	s.users[id] = user

	return user, true
}

func (s *userStore) delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.users[id]; !found {
		return false
	}

	delete(s.users, id)
	return true
}

func (s *userStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.users)
}

func validateUser(user User) error {
	if user.Name == "" || user.Email == "" {
		return fmt.Errorf("Name and email are required")
	}

	return nil
}

func userIDFromRequest(req *http.Request) (int, *http.Response) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		return 0, badRequest("Invalid user ID")
	}

	return id, nil
}

func userNotFound(id int) *http.Response {
	return jsonResponse(http.StatusNotFound, Response{
		Success: false,
		Error:   fmt.Sprintf("User with ID %d not found", id),
	})
}

func badRequest(message string) *http.Response {
	return jsonResponse(http.StatusBadRequest, Response{
		Success: false,
		Error:   message,
	})
}

func ok(message string, data any) *http.Response {
	return jsonResponse(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func jsonResponse(statusCode int, response Response) *http.Response {
	jsonData, _ := http.MarshalJSON(response)

	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": {http.ContentTypeJSON}},
		Body:       jsonData,
	}
}
