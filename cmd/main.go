package main

import (
	"fmt"
	http "github.com/baxromumarov/http-go"
)

func main() {
	srv := http.NewDefaultServer("localhost", 8080)
	http.Route("GET", "/hello", func(request *http.Request) *http.Response {

		return &http.Response{
			StatusCode: 200,
			Status:     "OK",
			Headers: http.Header{
				"Content-Type": {"text/plain"},
			},
			Body: []byte("Hello, World!"),
		}
	})

	if err := srv.StartServer(); err != nil {
		fmt.Println("Here: ", err)
	}
}
