// package integration
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestLaxInstall(t *testing.T) {

	// Print the current working directory
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %s", err)
	}
	fmt.Printf("Current working directory: %s\n", pwd)

	// Setup
	tempDir := "/tmp/lax-test"
	os.Mkdir(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// Start local server
	cmd := exec.Command("python3", "-m", "http.server", "--directory", tempDir, "8000")
	err = cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start local server: %v", err)
	}
	defer cmd.Process.Kill()

	// Run lax command
	laxCmd, _ := filepath.Abs("../../lax")
	installCmd := exec.Command(laxCmd, "collection", "install", "--server=http://localhost:8000", "test.collection")
	installCmd.Dir = tempDir
	output, err := installCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("lax install failed: %v, output: %s", err, output)
	}
	fmt.Println(output)

	// Verify installation
	if _, err := os.Stat(tempDir + "/test/collection"); os.IsNotExist(err) {
		t.Fatalf("Collection was not installed correctly")
	}

	// Cleanup is handled by defer statements
}

//func main() {
//
//}
