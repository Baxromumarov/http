package main

import (
	"encoding/json"
	"fmt"
	http_go "github.com/baxromumarov/http-go" // Make sure this import path is correct
	"io/ioutil"
	"log"
)

func main() {
	// Read the test data file
	filePath := "/home/bakhromumarov/go/src/github.com/baxromumarov/http-go/test_data.json"
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Create a new request
	req, err := http_go.NewRequest("POST", "http://localhost:8080/echo", data)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set required headers
	req.Header = http_go.Header{
		"Content-Type":   {"application/json"},
		"Content-Length": {fmt.Sprintf("%d", len(data))},
		"Connection":     {"keep-alive"},
	}

	// Send the request
	resp, err := http_go.DefaultClient.Send(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	if resp == nil {
		log.Fatal("Received nil response")
	}

	// Print response status
	fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.Status)

	// Try to unmarshal the response body
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		fmt.Printf("Raw response: %s\n", string(resp.Body))
		return
	}

	// Print the response data
	fmt.Println("Response data:", result)
}
