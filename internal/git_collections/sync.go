package git_collections

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jctanner/lax/internal/utils"
)

// Meta contains metadata about the collection
type Meta struct {
	Count int `json:"count"`
}

// CollectionResponse represents the structure of the response
type CollectionResponse struct {
	Meta Meta             `json:"meta"`
	Data []CollectionData `json:"data"`
}

// CollectionData represents individual collection data
type CollectionData struct {
	VersionsURL string `json:"versions_url"`
}

// VersionsResponse represents the structure of the versions response
type VersionsResponse struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data"`
}

// FetchPage fetches a single page of collections
func FetchPage(server string, limit, offset int) (*CollectionResponse, error) {
	url := fmt.Sprintf("%s/api/v3/collections/?limit=%d&offset=%d", server, limit, offset)
	fmt.Printf("fetching %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch page: %s", resp.Status)
	}

	var result CollectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// FetchVersionsPage fetches a single page of versions data
func FetchVersionsPage(server, versionsURL string, limit, offset int) (*VersionsResponse, error) {
	url := fmt.Sprintf("%s%s?limit=%d&offset=%d", server, versionsURL, limit, offset)
	fmt.Printf("fetching %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch versions: %s", resp.Status)
	}

	var result VersionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func FetchVersionPage(versionURL string) error {
	fmt.Printf("fetching %s\n", versionURL)
	resp, err := http.Get(versionURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("rate limit exceeded")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch versions: %s", resp.Status)
	}

	var result VersionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	return nil
}

// SaveJSON saves JSON data to a file
func SaveJSON(data interface{}, dest, filename string) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(dest, filename)
	return ioutil.WriteFile(filePath, bytes, 0644)
}

// PageAlreadyFetched checks if a page is already fetched
func PageAlreadyFetched(dest, filename string) bool {
	filePath := filepath.Join(dest, filename)
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	return false
}

// CollectionWorker function to fetch and save a collection page
func CollectionWorker(server, dest string, jobs <-chan int, results chan<- error, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	for pageNumber := range jobs {
		offset := (pageNumber - 1) * 10
		pageFilename := fmt.Sprintf("collections_page_%d.json", pageNumber)

		// Check if the page is already fetched
		if PageAlreadyFetched(dest, pageFilename) {
			results <- nil
			continue
		}

		for {
			page, err := FetchPage(server, 10, offset)
			if err != nil {
				if err.Error() == "rate limit exceeded" {
					fmt.Println("Rate limit exceeded, pausing for 60 seconds...")
					time.Sleep(60 * time.Second)
					continue
				} else {
					results <- fmt.Errorf("failed to fetch page %d: %v", pageNumber, err)
					return
				}
			}

			mu.Lock()
			if err := SaveJSON(page, dest, pageFilename); err != nil {
				mu.Unlock()
				results <- fmt.Errorf("failed to save page %d: %v", pageNumber, err)
				return
			}
			mu.Unlock()
			break
		}
		results <- nil
	}
}

// VersionsJob represents a job to fetch a specific versions page
type VersionsJob struct {
	VersionsURL string
	PageNumber  int
}

// VersionsWorker function to fetch and save versions pages
func VersionsWorker(server, dest string, jobs <-chan VersionsJob, results chan<- error, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	for job := range jobs {
		vOffset := (job.PageNumber - 1) * 10
		versionsFilename := "cv_page" + strings.ReplaceAll(job.VersionsURL, "/", "_") + "_page_" + strconv.Itoa(job.PageNumber) + ".json"

		if PageAlreadyFetched(dest, versionsFilename) {
			results <- nil
			continue
		}

		for {
			versions, err := FetchVersionsPage(server, job.VersionsURL, 10, vOffset)
			if err != nil {
				if err.Error() == "rate limit exceeded" {
					fmt.Println("Rate limit exceeded, pausing for 60 seconds...")
					time.Sleep(60 * time.Second)
					continue
				} else {
					results <- fmt.Errorf("failed to fetch versions for %s page %d: %v", job.VersionsURL, job.PageNumber, err)
					return
				}
			}

			mu.Lock()
			if err := SaveJSON(versions, dest, versionsFilename); err != nil {
				mu.Unlock()
				results <- fmt.Errorf("failed to save versions for %s page %d: %v", job.VersionsURL, job.PageNumber, err)
				return
			}
			mu.Unlock()
			break
		}
		results <- nil
	}
}

// QueueVersionsJobs queues all versions jobs for a given versions URL
func QueueVersionsJobs(server, dest string, versionsURL string, jobs chan<- VersionsJob, mu *sync.Mutex) error {
	// Fetch the first page to determine the total number of pages
	vOffset := 0
	versions, err := FetchVersionsPage(server, versionsURL, 10, vOffset)
	if err != nil {
		return err
	}

	// Save the first page
	versionsFilename := "cv_page" + strings.ReplaceAll(versionsURL, "/", "_") + "_page_1.json"
	mu.Lock()
	if err := SaveJSON(versions, dest, versionsFilename); err != nil {
		mu.Unlock()
		return err
	}
	mu.Unlock()

	// Determine the total number of pages
	totalItems := versions.Meta.Count
	totalPages := (totalItems + 10 - 1) / 10

	// Queue the remaining pages
	for vPage := 2; vPage <= totalPages; vPage++ {
		job := VersionsJob{
			VersionsURL: versionsURL,
			PageNumber:  vPage,
		}
		jobs <- job
	}
	return nil
}

// SyncCollections syncs collections from the server to the destination folder
func SyncCollections(server, dest string) error {
	// Ensure the destination folder exists
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// Fetch the first page to get the count of total items
	firstPage, err := FetchPage(server, 10, 0)
	if err != nil {
		return err
	}

	totalItems := firstPage.Meta.Count
	limit := 10
	totalPages := (totalItems + limit - 1) / limit

	jobs := make(chan int, totalPages)
	results := make(chan error, totalPages)

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Start 10 workers for fetching collection pages
	for w := 1; w <= 10; w++ {
		wg.Add(1)
		go CollectionWorker(server, dest, jobs, results, &wg, &mu)
	}

	// Send page numbers to the jobs channel
	for i := 1; i <= totalPages; i++ {
		jobs <- i
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
	close(results)

	// Check for errors
	for err := range results {
		if err != nil {
			return err
		}
	}

	return nil
}

// SyncVersions syncs versions from the server to the destination folder
func SyncVersions(server, dest string) error {
	files, err := ioutil.ReadDir(dest)
	if err != nil {
		return err
	}

	jobs := make(chan VersionsJob, len(files)*10)
	results := make(chan error, len(files)*10)

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Start 10 workers for fetching versions pages
	for w := 1; w <= 10; w++ {
		wg.Add(1)
		go VersionsWorker(server, dest, jobs, results, &wg, &mu)
	}

	// Read collection pages and queue the first page of versions URLs
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "collections_page_") {
			data, err := ioutil.ReadFile(filepath.Join(dest, file.Name()))
			if err != nil {
				return err
			}

			var page CollectionResponse
			if err := json.Unmarshal(data, &page); err != nil {
				return err
			}

			for _, item := range page.Data {
				if err := QueueVersionsJobs(server, dest, item.VersionsURL, jobs, &mu); err != nil {
					return err
				}
			}
		}
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
	close(results)

	// Check for errors
	for err := range results {
		if err != nil {
			return err
		}
	}

	return nil
}

func SyncArtifacts(server, dest string) error {
	files, err := ioutil.ReadDir(dest)
	if err != nil {
		return err
	}

	var filteredFiles []string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "cv_") && !strings.HasPrefix(file.Name(), "cv_detail") {
			filteredFiles = append(filteredFiles, dest+"/"+file.Name())
		}
	}

	var versionHrefs []string

	for _, filePath := range filteredFiles {

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePath, err)
			continue
		}

		var jsonData map[string]interface{}
		if err := json.Unmarshal(content, &jsonData); err != nil {
			fmt.Printf("Error unmarshalling JSON from file %s: %v\n", filePath, err)
			continue
		}

		data, ok := jsonData["data"].([]interface{})
		if !ok {
			fmt.Printf("No 'data' key found or 'data' is not a list in file %s\n", filePath)
			continue
		}

		for _, item := range data {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				fmt.Printf("Item is not a map in file %s\n", filePath)
				continue
			}

			href, ok := itemMap["href"].(string)
			if !ok {
				fmt.Printf("No 'href' key found or 'href' is not a string in item in file %s\n", filePath)
				continue
			}

			versionHrefs = append(versionHrefs, href)
		}
	}

	var detailsToDownload []utils.DownloadItem
	for _, versionHref := range versionHrefs {
		versionFilename := "cv_detail" + strings.ReplaceAll(versionHref, "/", "_") + ".json"
		if PageAlreadyFetched(dest, versionFilename) {
			continue
		}
		url := server + versionHref

		di := utils.DownloadItem{
			URL:      url,
			FilePath: dest + "/" + versionFilename,
		}
		detailsToDownload = append(detailsToDownload, di)
	}

	fmt.Printf("total detail pages to fetch: %d\n", len(detailsToDownload))
	utils.DownloadJSONFilesConcurrently(detailsToDownload, 10)

	detailFiles, err := ioutil.ReadDir(dest)
	if err != nil {
		return err
	}

	var filteredDetailsFiles []string
	for _, file := range detailFiles {
		if strings.HasPrefix(file.Name(), "cv_detail") {
			filteredDetailsFiles = append(filteredDetailsFiles, dest+"/"+file.Name())
		}
	}

	var tarsToDownload []utils.DownloadItem
	for _, file := range filteredDetailsFiles {
		//fmt.Printf("check %s\n", file)
		// Read the JSON file
		jsonFile, err := os.Open(file)
		if err != nil {
			//log.Fatalf("Failed to open file %s: %v\n", file, err)
			fmt.Printf("error %s\n", err)
		}
		defer jsonFile.Close()

		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			//log.Fatalf("Failed to read file %s: %v\n", file, err)
			fmt.Printf("error %s\n", err)
		}

		// Unmarshal the JSON into the CollectionVersionDetail struct
		var detail CollectionVersionDetail
		if err := json.Unmarshal(byteValue, &detail); err != nil {
			//log.Fatalf("Failed to unmarshal JSON: %v\n", err)
			fmt.Printf("error %s\n", err)
		}

		fmt.Printf("%s\n", detail)

		tarball := dest + "/" + detail.Artifact.FileName
		if !utils.FileExists(tarball) {
			di := utils.DownloadItem{
				URL:      detail.DownloadUrl,
				FilePath: tarball,
			}
			tarsToDownload = append(tarsToDownload, di)
		}
	}

	utils.DownloadBinaryFilesConcurrently(tarsToDownload, 5)

	return nil
}
