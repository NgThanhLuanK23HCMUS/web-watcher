package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func makeRequest(client *http.Client, wg *sync.WaitGroup, id int) {
	defer wg.Done()

	resp, err := client.Get("http://localhost:8081/search?q=test")
	if err != nil {
		fmt.Printf("Request %d failed: %v\n", id, err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Request %d: Status: %s\n", id, resp.Status)
}

func main() {
	// Create a client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Number of concurrent requests
	concurrency := 5
	// Number of requests per goroutine
	requestsPerRoutine := 10

	for {
		var wg sync.WaitGroup
		fmt.Println("\n--- Starting new batch of requests ---")

		// Launch multiple goroutines to make concurrent requests
		for i := 0; i < concurrency; i++ {
			for j := 0; j < requestsPerRoutine; j++ {
				wg.Add(1)
				go makeRequest(client, &wg, i*requestsPerRoutine+j)
			}
		}

		wg.Wait()
		fmt.Println("--- Batch complete, waiting 2 seconds ---")
		time.Sleep(2 * time.Second) // Wait before starting next batch
	}
}
