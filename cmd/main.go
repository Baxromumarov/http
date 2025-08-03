package main

import (
	"fmt"
	http "github.com/baxromumarov/http-go"
	"log"
)

// Example of registering multiple API endpoints using your server's http.Handle

func registerAPIs() {
	// GET /hello/:id
	http.Handle(http.GET, "/hello/:id", func(req *http.Request, params map[string]string) *http.Response {
		id := params["id"]
		body := fmt.Sprintf("Hello, user with ID %s", id)
		return &http.Response{
			StatusCode: 200,
			Headers: http.Header{
				"Content-Type": {"text/plain"},
			},
			Body: []byte(body),
		}
	})

	// GET /user/:userId/profile
	http.Handle(http.GET, "/user/:userId/profile", func(req *http.Request, params map[string]string) *http.Response {
		userId := params["userId"]
		// Imagine loading user profile from DB here
		jsonBody := fmt.Sprintf(`{"userId": "%s", "profile": "This is user profile data"}`, userId)
		return &http.Response{
			StatusCode: 200,
			Headers: http.Header{
				"Content-Type": {"application/json"},
			},
			Body: []byte(jsonBody),
		}
	})

	// POST /post/:postId/comment
	http.Handle(http.POST, "/post/:postId/comment", func(req *http.Request, params map[string]string) *http.Response {
		postId := params["postId"]
		comment := string(req.Body) // assuming raw comment in body
		body := fmt.Sprintf("Comment added to post %s: %s", postId, comment)
		return &http.Response{
			StatusCode: 201,
			Headers: http.Header{
				"Content-Type": {"text/plain"},
			},
			Body: []byte(body),
		}
	})

	// DELETE /post/:postId
	http.Handle(http.DELETE, "/post/:postId", func(req *http.Request, params map[string]string) *http.Response {
		postId := params["postId"]
		// Imagine deleting post from DB here
		body := fmt.Sprintf("Deleted post with ID %s", postId)
		return &http.Response{
			StatusCode: 200,
			Headers: http.Header{
				"Content-Type": {"text/plain"},
			},
			Body: []byte(body),
		}
	})
}

func main() {
	registerAPIs()

	// Start your server as usual (e.g. srv.StartServer())
	// It will use your existing `handleConn` logic that calls your route matching and handlers
	srv := http.NewDefaultServer("localhost", 8080)
	if err := srv.StartServer(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
