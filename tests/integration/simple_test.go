// package integration
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func runCommandInDir(command string, args []string, dir string) error {
	// Create the command
	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set the working directory, if specified
	if dir != "" {
		cmd.Dir = dir
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}
	defer cmd.Process.Kill()

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for command to finish: %w", err)
	}

	return nil
}

func TestLaxInstallFromHttp(t *testing.T) {

	laxCmd, _ := filepath.Abs("../../lax")

	// Print the current working directory
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %s", err)
	}
	fmt.Printf("Current working directory: %s\n", pwd)

	// Setup
	tempDir := "/tmp/lax-test"
	installDir := tempDir + "/install"
	repoDir := tempDir + "/repo"
	os.Mkdir(tempDir, 0755)
	os.Mkdir(installDir, 0755)
	os.Mkdir(repoDir, 0755)
	//defer os.RemoveAll(tempDir)

	// Make a collection (use ansible[-core] for now ...)
	galaxy_cli := "ansible-galaxy"
	if !commandExists(galaxy_cli) {
		// fmt.Printf("%s is not available in PATH\n", galaxy_cli)
		t.Fatalf("%s command is not available in PATH\n", galaxy_cli)
	}
	initErr := runCommandInDir("ansible-galaxy", []string{"collection", "init", "testn.col"}, repoDir)
	if initErr != nil {
		t.Fatalf("init failed: %v", initErr)
	}
	buildErr := runCommandInDir("ansible-galaxy", []string{"collection", "build", "--output-path=" + repoDir + "/collections", "."}, repoDir+"/"+"testn"+"/"+"col")
	if buildErr != nil {
		t.Fatalf("init failed: %v", buildErr)
	}

	// create the lax repo ... ?
	repoInitErr := runCommandInDir(laxCmd, []string{"createrepo"}, repoDir)
	if repoInitErr != nil {
		t.Fatalf("createrepo failed: %v", repoInitErr)
	}

	// Start local http server
	cmd := exec.Command("python3", "-m", "http.server", "--directory", repoDir, "8000")
	err = cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start local server: %v", err)
	}
	defer cmd.Process.Kill()
	time.Sleep(3 * time.Second)

	// Run lax install command
	installArgs := []string{
		"collection",
		"install",
		"--server=http://localhost:8000",
		"--cachedir=" + installDir + "/.cache",
		"--dest=" + installDir + "/.ansible",
		"testn.col",
	}
	installErr := runCommandInDir(laxCmd, installArgs, installDir)
	if installErr != nil {
		t.Fatalf("install failed: %v", installErr)
	}

	// Verify installation
	if _, err := os.Stat(installDir + "/.ansible/collections/ansible_collections/testn/col/MANIFEST.json"); os.IsNotExist(err) {
		t.Fatalf("Collection was not installed correctly")
	}

	// Cleanup is handled by defer statements
}

//func main() {
//
//}
