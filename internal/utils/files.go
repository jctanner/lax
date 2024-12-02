package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"

	"github.com/sirupsen/logrus"
	//"encoding/json"
)

// FileInfo holds information about the file type
type FileInfo struct {
	Path string
	Type string
}

// cache known files on disk
type FileStore struct {
	Files []FileInfo
}

func (fs *FileStore) AddFile(file FileInfo) {
	fs.Files = append(fs.Files, file)
}

// FindByGlob finds and returns all FileInfo objects with paths matching the given glob pattern
func (fs *FileStore) FindByGlob(pattern string) []string {
	var matches []string
	for _, file := range fs.Files {
		matched, err := filepath.Match(pattern, file.Path)
		if err != nil {
			fmt.Println("Error matching pattern:", err)
			continue
		}
		if matched {
			matches = append(matches, file.Path)
		}
	}
	return matches
}

func ExpandUser(path string) string {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		path = strings.Replace(path, "~", homeDir, 1)
	}
	return path
}

func GetAbsPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("%s\n", err)
		return "", err
	}
	return absPath, err
}

func GetFileInfo(path string) (*FileInfo, error) {
	// Use os.Stat to get file information
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist: %w", err)
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Determine the file type
	var fileType string
	switch mode := info.Mode(); {
	case mode.IsRegular():
		fileType = "regular file"
	case mode.IsDir():
		fileType = "directory"
	case mode&os.ModeSymlink != 0:
		fileType = "symlink"
	default:
		fileType = "other"
	}

	return &FileInfo{
		Path: path,
		Type: fileType,
	}, nil
}

func IsDir(path string) bool {
	finfo, err := GetFileInfo(path)
	if err != nil {
		return false
	}
	if finfo.Type == "directory" {
		return true
	}
	return false
}

func IsFile(path string) bool {
	finfo, err := GetFileInfo(path)
	if err != nil {
		return false
	}
	if finfo.Type == "regular file" || finfo.Type == "regular empty file" {
		return true
	}
	return false
}

func IsLink(path string) bool {
	finfo, err := GetFileInfo(path)
	if err != nil {
		return false
	}
	if finfo.Type == "symbolic link" || finfo.Type == "symlink" {
		return true
	}
	return false
}

func MakeDirs(path string) error {
	//fmt.Printf("MAKEDIRS: %s\n", path)
	// Check if the path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Path does not exist, create the directory
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to check directory: %v", err)
	}

	// Path exists, check if it is a directory
	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory")
	}

	// Path exists and is a directory
	return nil
}

func EndsWithMetaMainYAML(filename string) bool {
	return strings.HasSuffix(filename, "meta/main.yml") || strings.HasSuffix(filename, "meta/main.yaml")
}

func ListFilenamesInTarGz(filepath string) ([]string, error) {
	// Open the tar.gz file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a new gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create a new tar reader
	tarReader := tar.NewReader(gzReader)

	var filenames []string

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar entry: %w", err)
		}

		filenames = append(filenames, header.Name)
	}

	return filenames, nil
}

func ListTarGzFiles(dir string) ([]string, error) {
	var tarGzFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".gz" && filepath.Ext(filepath.Base(path[:len(path)-len(filepath.Ext(path))])) == ".tar" {
			tarGzFiles = append(tarGzFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tarGzFiles, nil
}

func ExtractJSONFilesFromTarGz(tarGzPath string, jsonFileNames []string) (map[string][]byte, error) {
	// Open the .tar.gz file
	file, err := os.Open(tarGzPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer file.Close()

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Create a map to store the results
	result := make(map[string][]byte)

	// Create a set of filenames for quick lookup
	fileSet := make(map[string]struct{})
	for _, name := range jsonFileNames {
		fileSet[name] = struct{}{}
	}

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}

		// Check if this is one of the JSON files we are looking for
		if _, found := fileSet[header.Name]; found {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tarReader); err != nil {
				return nil, fmt.Errorf("failed to read JSON file from tar archive: %w", err)
			}
			result[header.Name] = buf.Bytes()

			// Remove the found file from the set
			delete(fileSet, header.Name)

			// If all files are found, break the loop
			if len(fileSet) == 0 {
				break
			}
		}
	}

	return result, nil
}

func ExtractFilesFromTarGz(filepath string, filenamesToExtract []string) (map[string][]byte, error) {
	// Open the tar.gz file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a new gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create a new tar reader
	tarReader := tar.NewReader(gzReader)

	filesContent := make(map[string][]byte)
	filenamesToExtractMap := make(map[string]struct{}, len(filenamesToExtract))
	for _, filename := range filenamesToExtract {
		filenamesToExtractMap[filename] = struct{}{}
	}

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Check if the current file is in the list of files to extract
		if _, shouldExtract := filenamesToExtractMap[header.Name]; shouldExtract {
			// Read the file content
			content, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("failed to read file content: %w", err)
			}

			// Store the file content in the map
			filesContent[header.Name] = content
		}
	}

	return filesContent, nil
}

func CopyFile(src string, dst string) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Flush the writer
	err = dstFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to flush to disk: %w", err)
	}

	return nil
}

func ExtractTarGz(tarGzPath, dest string) error {
	// Open the tar.gz file
	file, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("open tar.gz file: %v", err)
	}
	defer file.Close()

	// Create gzip reader
	uncompressedStream, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("create gzip reader: %v", err)
	}
	defer uncompressedStream.Close()

	// Create tar reader
	tarReader := tar.NewReader(uncompressedStream)

	// Iterate through the files in the archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("read tar header: %v", err)
		}

		// Determine the target file path
		target := filepath.Join(dest, header.Name)

		// Ensure the parent directory exists
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("create directory: %v", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory if it does not exist
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
					return fmt.Errorf("create directory: %v", err)
				}
			}
		case tar.TypeReg:
			// Create and write to file
			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("create file: %v", err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("write file: %v", err)
			}
			outFile.Close()

			// Set file permissions
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("set file permissions: %v", err)
			}
		default:
			return fmt.Errorf("unsupported file type: %v", header.Typeflag)
		}
	}
	return nil
}

// ExtractTarGz extracts a tar.gz file to the specified destination
func ExtractRoleTarGz(tarGzPath, dest string) error {
	// Create a temporary directory
	tempDir, err := CreateTempDirectory()
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir) // Clean up the temporary directory

	// Execute the tar xzvf command to extract files to the temporary directory
	var stdout bytes.Buffer
	cmd := exec.Command("tar", "xzvf", tarGzPath, "-C", tempDir)
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		stdoutLines := strings.Split(stdout.String(), "\n")
		for _, line := range stdoutLines {
			logrus.Errorf(line)
		}
		return fmt.Errorf("failed to execute tar command: %v", err)
	}

	stdoutLines := strings.Split(stdout.String(), "\n")
	for _, line := range stdoutLines {
		logrus.Debugf(line)
	}

	/*
		// Copy all files from the temporary directory to the destination directory
		if err := CopyDir(tempDir, dest); err != nil {
			return fmt.Errorf("copy files to destination: %v", err)
		}
	*/

	logrus.Debugf("finding 'meta' dir in tmpdir:%s", tempDir)

	// find the meta dir ...
	metaDir := ""
	_ = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the current path is a directory
		if info.IsDir() && info.Name() == "meta" {
			//directories = append(directories, path)
			metaDir = path
			return nil
		}
		return nil
	})
	logrus.Debugf("found meta dir at %s", metaDir)
	srcDir := filepath.Dir(metaDir)
	logrus.Debugf("setting source dir as %s", srcDir)

	// copy the src to the dst ...
	logrus.Debugf("cp -Ra %s/. %s", srcDir, dest)
	cmd = exec.Command("cp", "-Ra", srcDir+"/.", dest)

	// Set the command's output to the standard output and error to the standard error
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failure to execute cp command: %v", err)
	}

	return nil
}

func RemoveFirstPathElement(path string) string {
	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) > 1 {
		return filepath.Join(parts[1:]...)
	}
	return path // Return the original path if there's only one part
}

func CreateSymlink(srcFile string, linkName string) error {
	// Get the absolute path of the source file
	srcPath, err := filepath.Abs(srcFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of source file: %w", err)
	}

	// Get the directory and base name of the source file
	srcDir := filepath.Dir(srcPath)
	srcBase := filepath.Base(srcPath)

	linkName = filepath.Base(linkName)

	// Construct the shell command to create the symlink
	cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && ln -s %s %s", srcDir, srcBase, linkName))

	// Run the shell command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create symlink: %w, output: %s", err, output)
	}

	return nil
}

func FindMatchingFiles(directory string, pattern string) ([]string, error) {

	maxDepth := 1

	// Check if the directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %v", directory)
	}

	// Helper function to calculate depth
	calculateDepth := func(base, target string) int {
		base = filepath.Clean(base)
		target = filepath.Clean(target)
		if base == target {
			return 0
		}
		rel, err := filepath.Rel(base, target)
		if err != nil {
			return -1
		}
		return len(strings.Split(rel, string(os.PathSeparator)))
	}

	var matches []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		//fmt.Printf("walk: %s\n", path)

		// Skip directories and only check files
		if !info.IsDir() {
			// Check the depth of the file
			if calculateDepth(directory, path) <= maxDepth {
				// Check if the file matches the pattern
				matched, err := filepath.Match(pattern, filepath.Base(path))
				if err != nil {
					return err
				}
				if matched {
					matches = append(matches, path)
				}
			}
		}

		// Skip directories that exceed the max depth
		if info.IsDir() && calculateDepth(directory, path) >= maxDepth {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return matches, nil
}

func CreateTempDirectory() (string, error) {
	// Create a temporary directory with a specified prefix
	tempDir, err := os.MkdirTemp("", "lax_extract_")
	if err != nil {
		return "", fmt.Errorf("create temporary directory: %v", err)
	}
	return tempDir, nil
}

func IsDirEmpty(dirPath string) (bool, error) {
	// Open the directory
	f, err := os.Open(dirPath)
	if err != nil {
		return false, fmt.Errorf("failed to open directory: %v", err)
	}
	defer f.Close()

	// Read directory contents
	entries, err := f.Readdirnames(1) // Read at least one entry
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("failed to read directory: %v", err)
	}

	// If entries slice is empty, directory is empty
	return len(entries) == 0, nil
}
