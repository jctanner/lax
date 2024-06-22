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

/*
func IsDir(path string) bool {
    info, err := os.Stat(path)
    if err != nil {
        return false
    }
    return info.IsDir()
}

func IsFile(path string) bool {
    abspath, _ := GetAbsPath(path)
    _, err := os.Stat(abspath)
    //if err != nil {
    if errors.Is(err, fs.ErrNotExist) {
        fmt.Printf("%s does not exist1 %s\n", abspath, err)
        return false
    }
    if IsDir(abspath) {
        fmt.Printf("%s does not exist2\n", abspath)
        return false
    }
    return true
}
*/

// GetFileInfo runs the `stat` command on the provided path and parses the output
func GetFileInfo(path string) (*FileInfo, error) {
	// Execute the stat command
	cmd := exec.Command("stat", "-c", "%F", path)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute stat command: %w", err)
	}

	// Parse the output to determine the file type
	fileType := strings.TrimSpace(string(output))

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
	if finfo.Type == "symbolic link" {
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

/*
// Unmarshal JSON data into a map
func UnmarshalJSONData(jsonData []byte) (map[string]interface{}, error) {
	unmarshaledData := make(map[string]interface{})
	for filename, data := range jsonData {
		var temp map[string]interface{}
		if err := json.Unmarshal(data, &temp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON file %s: %w", filename, err)
		}
		unmarshaledData[filename] = temp
	}
	return unmarshaledData, nil
}
*/

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

/*
// ExtractTarGz extracts a tar.gz file to the specified destination
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
		case tar.TypeSymlink:
			// Create symlink
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("create symlink: %v", err)
			}
		case tar.TypeLink:
			// Create hard link
			if err := os.Link(header.Linkname, target); err != nil {
				return fmt.Errorf("create hard link: %v", err)
			}
		default:
			return fmt.Errorf("unsupported file type: %v", header.Typeflag)
		}
	}
	return nil
}
*/

/*
func ExtractTarGz(tarGzPath, dest string) error {

	// extract to a tempdir first ...
	tempDir, _ := CreateTempDirectory()

	// Ensure the destination directory exists
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("create destination directory: %v", err)
	}

	// Get the absolute path of the tar.gz file
	absTarGzPath, err := filepath.Abs(tarGzPath)
	if err != nil {
		return fmt.Errorf("get absolute path of tar.gz file: %v", err)
	}

	// Change the current working directory to the destination directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current working directory: %v", err)
	}
	defer os.Chdir(originalDir) // Ensure we return to the original directory

	if err := os.Chdir(dest); err != nil {
		return fmt.Errorf("change directory to destination: %v", err)
	}

	// Execute the tar xzvf command
	cmd := exec.Command("tar", "xzvf", absTarGzPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute tar command: %v", err)
	}

	return nil
}
*/

// ExtractTarGz extracts a tar.gz file to the specified destination
func ExtractRoleTarGz(tarGzPath, dest string) error {
	// Create a temporary directory
	tempDir, err := CreateTempDirectory()
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir) // Clean up the temporary directory

	// Execute the tar xzvf command to extract files to the temporary directory
	cmd := exec.Command("tar", "xzvf", tarGzPath, "-C", tempDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute tar command: %v", err)
	}

	/*
		// Copy all files from the temporary directory to the destination directory
		if err := CopyDir(tempDir, dest); err != nil {
			return fmt.Errorf("copy files to destination: %v", err)
		}
	*/

	fmt.Printf("%s\n", tempDir)

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
	fmt.Printf("found meta dir at %s\n", metaDir)
	srcDir := filepath.Dir(metaDir)
	fmt.Printf("setting source dir as %s\n", srcDir)

	// copy the src to the dst ...
	cmd = exec.Command("cp", "-Ra", srcDir+"/.", dest)

	// Set the command's output to the standard output and error to the standard error
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute cp command: %v", err)
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

/*
func FindMatchingFiles(directory, pattern string) ([]string, error) {
	// Combine the directory and pattern to create a full search pattern
	searchPattern := filepath.Join(directory, pattern)

	// Use filepath.Glob to find matching files
	logrus.Debugf("%s search starting\n", pattern)
	matches, err := filepath.Glob(searchPattern)
	logrus.Debugf("%s search finished\n", pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search for files: %w", err)
		//panic("")
	}

	return matches, nil
}
*/

func FindMatchingFiles(directory string, pattern string) ([]string, error) {
	searchPattern := filepath.Join(directory, pattern)
	logrus.Debugf("%s search starting\n", patsearchPatterntern)
	cmd := exec.Command("find", ".", "-maxdepth", "1", "-name", searchPattern)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	logrus.Debugf("%s search finished\n", searchPattern)
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, line := range bytes.Split(out.Bytes(), []byte{'\n'}) {
		if len(line) > 0 {
			matches = append(matches, string(line))
		}
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
