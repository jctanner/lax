package integration

import (
    "os"
    "os/exec"
    "testing"
)

func TestLaxInstall(t *testing.T) {
    // Setup
    tempDir := "/tmp/lax-test"
    os.Mkdir(tempDir, 0755)
    defer os.RemoveAll(tempDir)

    // Start local server
    cmd := exec.Command("python3", "-m", "http.server", "--directory", tempDir, "8000")
    err := cmd.Start()
    if err != nil {
        t.Fatalf("Failed to start local server: %v", err)
    }
    defer cmd.Process.Kill()

    // Run lax command
    installCmd := exec.Command("lax", "collection", "install", "--server=http://localhost:8000", "test.collection")
    installCmd.Dir = tempDir
    output, err := installCmd.CombinedOutput()
    if err != nil {
        t.Fatalf("lax install failed: %v, output: %s", err, output)
    }

    // Verify installation
    if _, err := os.Stat(tempDir + "/test/collection"); os.IsNotExist(err) {
        t.Fatalf("Collection was not installed correctly")
    }

    // Cleanup is handled by defer statements
}
