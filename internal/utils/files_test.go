package utils

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestFindByGlob(t *testing.T) {
	// Arrange: Create a FileStore with some test data
	fs := &FileStore{
		Files: []FileInfo{
			{Path: "file1.txt", Type: "text"},
			{Path: "file2.log", Type: "log"},
			{Path: "subdir/file3.txt", Type: "text"},
			{Path: "subdir/file4.log", Type: "log"},
		},
	}

	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "Match all .txt files",
			pattern:  "*.txt",
			expected: []string{"file1.txt"},
		},
		{
			name:     "Match all files in subdir",
			pattern:  "subdir/*",
			expected: []string{"subdir/file3.txt", "subdir/file4.log"},
		},
		{
			name:     "Match all .log files in subdir",
			pattern:  "subdir/*.log",
			expected: []string{"subdir/file4.log"},
		},
		{
			name:     "No matches",
			pattern:  "*.json",
			expected: []string{},
		},
	}

	// Act & Assert: Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fs.FindByGlob(tt.pattern)
			if !equalSlices(result, tt.expected) {
				t.Errorf("FindByGlob(%q) = %v, want %v", tt.pattern, result, tt.expected)
			}
		})
	}
}

// Helper function to compare two slices for equality
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

/**************************************************************************
/*************************************************************************/

func TestExpandUser(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get user home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Home Directory Expansion",
			input:    "~/testdir",
			expected: filepath.Join(homeDir, "testdir"),
		},
		{
			name:     "No Home Directory Expansion",
			input:    "/usr/local/test",
			expected: "/usr/local/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandUser(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandUser(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetAbsPath(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "Relative Path",
			input:     ".",
			expectErr: false,
		},
		{
			name:      "Absolute Path",
			input:     "/usr/local",
			expectErr: false,
		},
		//{
		//	name:      "Invalid Path",
		//	input:     string([]byte{0x7f}), // Invalid path with an ASCII control character
		//	expectErr: true,
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetAbsPath(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("GetAbsPath(%q) error = %v, expectErr %v", tt.input, err, tt.expectErr)
			}
			if err == nil && !filepath.IsAbs(result) {
				t.Errorf("GetAbsPath(%q) = %v, expected an absolute path", tt.input, result)
			}
		})
	}
}

func TestGetFileInfo(t *testing.T) {
	// Create a temporary file to use for testing
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up

	// Create a temporary directory to use for testing
	tmpDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("could not create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up

	tests := []struct {
		name      string
		input     string
		expectErr bool
		expected  string
	}{
		{
			name:      "Regular File",
			input:     tmpFile.Name(),
			expectErr: false,
			expected:  "regular file",
		},
		{
			name:      "Directory",
			input:     tmpDir,
			expectErr: false,
			expected:  "directory",
		},
		{
			name:      "Non-Existent File",
			input:     "/non/existent/file",
			expectErr: true,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileInfo, err := GetFileInfo(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("GetFileInfo(%q) error = %v, expectErr %v", tt.input, err, tt.expectErr)
			}
			if err == nil && fileInfo.Type != tt.expected {
				t.Errorf("GetFileInfo(%q) = %v, want %v", tt.input, fileInfo.Type, tt.expected)
			}
		})
	}
}

func TestIsDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("could not create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Valid Directory",
			path:     tmpDir,
			expected: true,
		},
		{
			name:     "Non-Existent Path",
			path:     "/non/existent/path",
			expected: false,
		},
		{
			name:     "Regular File",
			path:     "testfile.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If testing a regular file, create it
			if tt.name == "Regular File" {
				file, err := os.Create(tt.path)
				if err != nil {
					t.Fatalf("could not create temp file: %v", err)
				}
				defer os.Remove(file.Name()) // Clean up
				file.Close()
			}

			result := IsDir(tt.path)
			if result != tt.expected {
				t.Errorf("IsDir(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsFile(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Valid Regular File",
			path:     tmpFile.Name(),
			expected: true,
		},
		{
			name:     "Non-Existent Path",
			path:     "/non/existent/path",
			expected: false,
		},
		{
			name:     "Directory",
			path:     ".",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFile(tt.path)
			if result != tt.expected {
				t.Errorf("IsFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsLink(t *testing.T) {
	// Create a temporary file and a symbolic link for testing
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up

	linkPath := tmpFile.Name() + "_link"
	err = os.Symlink(tmpFile.Name(), linkPath)
	if err != nil {
		t.Fatalf("could not create symbolic link: %v", err)
	}
	defer os.Remove(linkPath) // Clean up

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Valid Symbolic Link",
			path:     linkPath,
			expected: true,
		},
		{
			name:     "Regular File",
			path:     tmpFile.Name(),
			expected: false,
		},
		{
			name:     "Non-Existent Path",
			path:     "/non/existent/path",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//finfo, _ := GetFileInfo(tt.path)
			//t.Logf("finfo %v", finfo)
			result := IsLink(tt.path)
			if result != tt.expected {
				t.Errorf("IsLink(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

/****************************************************************************************
****************************************************************************************/

func TestMakeDirs(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		expectErr bool
	}{
		{
			name:      "Create Non-Existent Directory",
			path:      "testdir/subdir",
			expectErr: false,
		},
		{
			name:      "Existing Directory",
			path:      "testdir",
			expectErr: false,
		},
		{
			name:      "Path Exists But Is File",
			path:      "testfile",
			expectErr: true,
		},
	}

	// Setup: Create a file for the "Path Exists But Is File" test case
	err := os.WriteFile("testfile", []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("could not create test file: %v", err)
	}
	defer os.Remove("testfile") // Cleanup

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure cleanup of directories after test
			defer os.RemoveAll("testdir")

			err := MakeDirs(tt.path)
			if (err != nil) != tt.expectErr {
				t.Errorf("MakeDirs(%q) error = %v, expectErr %v", tt.path, err, tt.expectErr)
			}

			// If no error is expected, verify directory existence
			if !tt.expectErr {
				info, statErr := os.Stat(tt.path)
				if statErr != nil || !info.IsDir() {
					t.Errorf("MakeDirs(%q) did not create directory", tt.path)
				}
			}
		})
	}
}

func TestEndsWithMetaMainYAML(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "Valid meta/main.yml",
			filename: "some/path/meta/main.yml",
			expected: true,
		},
		{
			name:     "Valid meta/main.yaml",
			filename: "some/path/meta/main.yaml",
			expected: true,
		},
		{
			name:     "Invalid meta/main.txt",
			filename: "some/path/meta/main.txt",
			expected: false,
		},
		{
			name:     "Non-meta file",
			filename: "some/path/not_meta.yml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndsWithMetaMainYAML(tt.filename)
			if result != tt.expected {
				t.Errorf("EndsWithMetaMainYAML(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestListFilenamesInTarGz(t *testing.T) {
	// Create a tar.gz file for testing
	testTarGz := "test.tar.gz"
	files := []struct {
		Name string
		Body string
	}{
		{Name: "file1.txt", Body: "content of file1"},
		{Name: "file2.txt", Body: "content of file2"},
		{Name: "dir/file3.txt", Body: "content of file3"},
	}

	createTestTarGz(testTarGz, files)
	defer os.Remove(testTarGz) // Cleanup

	tests := []struct {
		name      string
		filepath  string
		expected  []string
		expectErr bool
	}{
		{
			name:      "Valid TarGz File",
			filepath:  testTarGz,
			expected:  []string{"file1.txt", "file2.txt", "dir/file3.txt"},
			expectErr: false,
		},
		{
			name:      "Non-Existent File",
			filepath:  "nonexistent.tar.gz",
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ListFilenamesInTarGz(tt.filepath)
			if (err != nil) != tt.expectErr {
				t.Errorf("ListFilenamesInTarGz(%q) error = %v, expectErr %v", tt.filepath, err, tt.expectErr)
			}
			if !tt.expectErr && !equalSlices(result, tt.expected) {
				t.Errorf("ListFilenamesInTarGz(%q) = %v, want %v", tt.filepath, result, tt.expected)
			}
		})
	}
}

// Helper function to create a tar.gz file
func createTestTarGz(filepath string, files []struct {
	Name string
	Body string
}) {
	file, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	for _, f := range files {
		header := &tar.Header{
			Name: f.Name,
			Size: int64(len(f.Body)),
			Mode: 0600,
		}
		err := tarWriter.WriteHeader(header)
		if err != nil {
			panic(err)
		}
		_, err = tarWriter.Write([]byte(f.Body))
		if err != nil {
			panic(err)
		}
	}
}
