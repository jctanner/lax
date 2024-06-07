package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	return false
}

type DownloadItem struct {
	URL      string
	FilePath string
}

func downloadJSONFile(item DownloadItem, wg *sync.WaitGroup, results chan<- error, sem chan struct{}) {

	defer wg.Done()
	defer func() { <-sem }() 

	// Download the data
    fmt.Printf("fetching %s to %s\n", item.URL, item.FilePath)

	resp, err := http.Get(item.URL)
	if err != nil {
		results <- fmt.Errorf("error downloading %s: %v", item.URL, err)
		return
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		results <- fmt.Errorf("bad status: %s", resp.Status)
		return
	}

	// Read the body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		results <- fmt.Errorf("error reading response body from %s: %v", item.URL, err)
		return
	}

	// Unmarshal the JSON to check if it's valid
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		results <- fmt.Errorf("error unmarshalling JSON from %s: %v", item.URL, err)
		return
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(item.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		results <- fmt.Errorf("error creating directory %s: %v", dir, err)
		return
	}

	// Write the body to file
	if err := ioutil.WriteFile(item.FilePath, body, 0644); err != nil {
		results <- fmt.Errorf("error saving file %s: %v", item.FilePath, err)
		return
	}

	results <- nil
}

func downloadBinaryFile(item DownloadItem, wg *sync.WaitGroup, results chan<- error, sem chan struct{}) {

	defer wg.Done()
	defer func() { <-sem }() 

	// Download the data
    fmt.Printf("fetching %s to %s\n", item.URL, item.FilePath)

	resp, err := http.Get(item.URL)
	if err != nil {
		results <- fmt.Errorf("error downloading %s: %v", item.URL, err)
		return
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		results <- fmt.Errorf("bad status: %s", resp.Status)
		return
	}

	// Read the body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		results <- fmt.Errorf("error reading response body from %s: %v", item.URL, err)
		return
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(item.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		results <- fmt.Errorf("error creating directory %s: %v", dir, err)
		return
	}

	// Write the body to file
	if err := ioutil.WriteFile(item.FilePath, body, 0644); err != nil {
		results <- fmt.Errorf("error saving file %s: %v", item.FilePath, err)
		return
	}

	results <- nil
}

func DownloadJSONFilesConcurrently(items []DownloadItem, maxWorkers int) {

	var wg sync.WaitGroup
	results := make(chan error, len(items))
	sem := make(chan struct{}, maxWorkers) // Semaphore to limit concurrency

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore slot
		go downloadJSONFile(item, &wg, results, sem)
	}

	wg.Wait()
	close(results)

	for err := range results {
		if err != nil {
			fmt.Println(err)
		}
	}

}

func DownloadBinaryFilesConcurrently(items []DownloadItem, maxWorkers int) {

	var wg sync.WaitGroup
	results := make(chan error, len(items))
	sem := make(chan struct{}, maxWorkers) // Semaphore to limit concurrency

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore slot
		go downloadBinaryFile(item, &wg, results, sem)
	}

	wg.Wait()
	close(results)

	for err := range results {
		if err != nil {
			fmt.Println(err)
		}
	}

}
