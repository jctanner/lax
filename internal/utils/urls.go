package utils

import (
	"fmt"
	"strings"
	"time"

	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// isURL checks if a given string is a valid URL
func IsURL(str string) bool {
	parsedURL, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}

	// Ensure the scheme and host are present
	return strings.HasPrefix(parsedURL.Scheme, "http") && parsedURL.Host != ""
}

func DownloadFile(urlStr string) (string, error) {
	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %v", err)
	}

	// Get the basename of the URL's path
	filename := path.Base(parsedURL.Path)
	if filename == "" {
		return "", fmt.Errorf("invalid URL: no filename in path")
	}

	// Get the path to the system's temporary directory
	tmpDir := os.TempDir()

	// Create the full file path
	filePath := filepath.Join(tmpDir, filename)

	// Make the HTTP GET request
	resp, err := http.Get(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
	}

	// Create the file in the temporary directory
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	// Copy the content from the response to the file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	fmt.Printf("File downloaded successfully: %s\n", filePath)
	return filePath, nil
}

func DownloadBinaryFileToPathWithBearerToken(urlStr string, token string, filePath string) (string, error) {
	logrus.Debugf("Downloading w/ token %s -> %s\n", urlStr, filePath)

	/*
		// Create a new HTTP request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logrus.Errorf("Error creating HTTP request: %v", err)
			return nil, err
		}

		// Add Authorization header if accessToken is non-empty
		if c.accessToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
			logrus.Infof("Added Authorization header with token")
		}

		// Use the default HTTP client to send the request
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			logrus.Errorf("Error making HTTP request: %v", err)
			return nil, err
		}

		return resp, nil
	*/

	// Make the HTTP GET request
	/*
		resp, err := http.Get(urlStr)
		if err != nil {
			return "", fmt.Errorf("failed to download file: %v", err)
		}
		defer resp.Body.Close()
	*/

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logrus.Errorf("Error creating HTTP request: %v", err)
		return "", err
	}

	// add the token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	/*
		// Check if the request was successful
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
		}
	*/

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Errorf("Error making HTTP request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// Create the file
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	// Copy the content from the response to the file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	logrus.Debugf("File downloaded successfully: %s\n", filePath)
	return filePath, nil

}

func DownloadBinaryFileToPath(urlStr string, filePath string) (string, error) {

	logrus.Debugf("Downloading %s -> %s\n", urlStr, filePath)

	// Make the HTTP GET request
	resp, err := http.Get(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
	}

	// Create the file
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	// Copy the content from the response to the file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	logrus.Debugf("File downloaded successfully: %s\n", filePath)
	return filePath, nil
}

func IsURLGood(url string) bool {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Head(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
