package roles

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-resty/resty/v2"
)

func SyncRoles(server string, dest string) error {
	client := resty.New()
	var wg sync.WaitGroup
	pageCh := make(chan int, 100) // Buffered channel to limit the number of concurrent requests
	errCh := make(chan error, 1)  // Channel to capture errors

	// Create destination folder if it doesn't exist
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		os.MkdirAll(dest, os.ModePerm)
	}

	// Fetch the first page to get the total count
	firstPageURL := fmt.Sprintf("%s/api/v1/roles/?page=1", server)
	resp, err := client.R().Get(firstPageURL)
	if err != nil {
		return fmt.Errorf("failed to fetch the first page: %v", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to fetch the first page: status code %d", resp.StatusCode())
	}

	var firstPageData map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &firstPageData); err != nil {
		return fmt.Errorf("failed to parse the first page response: %v", err)
	}

	count, ok := firstPageData["count"].(float64)
	if !ok {
		return fmt.Errorf("count key not found or not a number in the first page response")
	}

	totalPages := int(count / 10)
	if int(count)%10 != 0 {
		totalPages++
	}

	fmt.Printf("Total number of pages to fetch: %d\n", totalPages)

	// Worker function to process pages
	worker := func() {
		defer wg.Done()
		for page := range pageCh {
			fileName := filepath.Join(dest, fmt.Sprintf("roles_page_%d.json", page))

			// Skip downloading if the file already exists
			if _, err := os.Stat(fileName); err == nil {
				fmt.Printf("File for page %d already exists, skipping download.\n", page)
				continue
			}

			fmt.Printf("Starting download for page %d.\n", page)
			url := fmt.Sprintf("%s/api/v1/roles/?page=%d", server, page)
			resp, err := client.R().Get(url)
			if err != nil {
				errCh <- fmt.Errorf("failed to fetch page %d: %v", page, err)
				return
			}

			if resp.StatusCode() != 200 {
				errCh <- fmt.Errorf("failed to fetch page %d: status code %d", page, resp.StatusCode())
				return
			}

			if len(resp.Body()) == 0 {
				// If the response body is empty, assume we've reached the last page
				close(pageCh)
				return
			}

			// Write the response body to a file
			err = ioutil.WriteFile(fileName, resp.Body(), 0644)
			if err != nil {
				errCh <- fmt.Errorf("failed to write page %d to file: %v", page, err)
				return
			}

			fmt.Printf("Successfully fetched and saved page %d\n", page)
		}
	}

	// Start workers
	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker()
	}

	// Feed pages to workers
	go func() {
		for page := 1; page <= totalPages; page++ {
			select {
			case err := <-errCh:
				close(pageCh)
				fmt.Println("Error:", err)
				return
			default:
				pageCh <- page
			}
		}
		close(pageCh)
	}()

	// Wait for all workers to complete
	wg.Wait()
	close(errCh)

	// Check if there were any errors
	if len(errCh) > 0 {
		return <-errCh
	}

	return nil
}
