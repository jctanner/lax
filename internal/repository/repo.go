package repository

import (
	"lax/internal/utils"
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"encoding/gob"
	"encoding/json"

	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"

	"time"

	"github.com/blang/semver/v4"
)

type RepoMeta struct {
	Date                string       `json:"date"`
	CollectionManifests RepoMetaFile `json:"collection_manifests"`
	CollectionFiles     RepoMetaFile `json:"collection_files"`
}

type RepoMetaFile struct {
	Date     string `json:"date"`
	Filename string `json:"filename"`
}

type Manifest struct {
	CollectionInfo CollectionInfo `json:"collection_info"`
}

type CollectionInfo struct {
	Namespace    string            `json:"namespace"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

type FilesMeta struct {
	Files []FileInfo `json:"files"`
}

type FileInfo struct {
	Name           string `json:"name"`
	FType          string `json:"ftype"`
	CheckSumType   string `json:"chksum_type"`
	CheckSumSHA256 string `json:"chksum_sha256"`
}

type CachedFileInfo struct {
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	FileName       string `json:"filename"`
	FileType       string `json:"filetype"`
	CheckSumSHA256 string `json:"chksum_sha256"`
}

func createManifestsTarGz(manifests []Manifest, tarGzPath string) error {
	// Create a buffer to write the tar archive
	var buf bytes.Buffer

	// Create a tar writer
	tarWriter := tar.NewWriter(&buf)

	// Add each manifest as a JSON file to the tar archive
	for i, manifest := range manifests {
		// Marshal the manifest to JSON
		jsonData, err := json.Marshal(manifest)
		if err != nil {
			return fmt.Errorf("failed to marshal manifest to JSON: %w", err)
		}

		// Create a tar header for the JSON file
		header := &tar.Header{
			Name: fmt.Sprintf("manifest_%d.json", i),
			Mode: 0600,
			Size: int64(len(jsonData)),
		}

		// Write the header to the tar archive
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		// Write the JSON data to the tar archive
		if _, err := tarWriter.Write(jsonData); err != nil {
			return fmt.Errorf("failed to write JSON data to tar archive: %w", err)
		}
	}

	// Close the tar writer
	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Create the output file for the tar.gz archive
	outFile, err := os.Create(tarGzPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create a gzip writer
	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	// Write the tar archive to the gzip writer
	if _, err := gzipWriter.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write tar data to gzip writer: %w", err)
	}

	return nil
}

func ExtractManifestsFromTarGz(tarGzPath string) ([]Manifest, error) {
	// Open the tar.gz file
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

	var manifests []Manifest

	// Iterate over the files in the tar archive
	for {
		// Get the next header
		header, err := tarReader.Next()
		if err == io.EOF {
			break // end of tar archive
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}

		// Read the JSON data
		jsonData := make([]byte, header.Size)
		if _, err := io.ReadFull(tarReader, jsonData); err != nil {
			return nil, fmt.Errorf("failed to read JSON data: %w", err)
		}

		// Unmarshal the JSON data into a Manifest object
		var manifest Manifest
		if err := json.Unmarshal(jsonData, &manifest); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
		}

		// Append the manifest to the slice
		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

/*
func createFilesCacheTarGz(manifests []CachedFileInfo, tarGzPath string) error {
	// Create a buffer to write the tar archive
	var buf bytes.Buffer

	// Create a tar writer
	tarWriter := tar.NewWriter(&buf)

	// Add each manifest as a JSON file to the tar archive
	for i, manifest := range manifests {
		// Marshal the manifest to JSON
		jsonData, err := json.Marshal(manifest)
		if err != nil {
			return fmt.Errorf("failed to marshal manifest to JSON: %w", err)
		}

		// Create a tar header for the JSON file
		header := &tar.Header{
			Name: fmt.Sprintf("manifest_%d.json", i),
			Mode: 0600,
			Size: int64(len(jsonData)),
		}

		// Write the header to the tar archive
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		// Write the JSON data to the tar archive
		if _, err := tarWriter.Write(jsonData); err != nil {
			return fmt.Errorf("failed to write JSON data to tar archive: %w", err)
		}
	}

	// Close the tar writer
	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Create the output file for the tar.gz archive
	outFile, err := os.Create(tarGzPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create a gzip writer
	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	// Write the tar archive to the gzip writer
	if _, err := gzipWriter.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write tar data to gzip writer: %w", err)
	}

	return nil
}
*/

func saveCachedFilesToGzippedFile(manifests []CachedFileInfo, filePath string, chunkSize int) error {
	// Create the output file
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create a gzip writer
	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	// Create a buffer and an encoder for the gob
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	for start := 0; start < len(manifests); start += chunkSize {
		end := start + chunkSize
		if end > len(manifests) {
			end = len(manifests)
		}

		// Encode the chunk to the buffer
		buf.Reset() // Clear the buffer
		if err := encoder.Encode(manifests[start:end]); err != nil {
			return fmt.Errorf("failed to encode chunk: %w", err)
		}

		// Write the buffer to the gzip writer
		if _, err := gzipWriter.Write(buf.Bytes()); err != nil {
			return fmt.Errorf("failed to write chunk to gzip writer: %w", err)
		}
	}

	return nil
}

func CreateRepo(dest string) error {

	fmt.Printf("Create repo in %s\n", dest)

	// find the full path
	apath, err := utils.GetAbsPath(dest)
	if err != nil {
		return err
	}

	// assert it's a dir that exists
	if !utils.IsDir(apath) {
		fmt.Printf("%s is not a directory\n", apath)
		return nil
	}

	// assert it has a collections subdir
	collectionsPath := filepath.Join(apath, "collections")
	if !utils.IsDir(collectionsPath) {
		fmt.Printf("%s is not a directory\n", collectionsPath)
		return nil
	}

	// make a list of tarballs
	tarBalls, err := utils.ListTarGzFiles(collectionsPath)
	if err != nil {
		return err
	}

	// we need the metdata from each file
	metadataDir := filepath.Join(apath, "metadata")
	err = utils.MakeDirs(metadataDir)
	if err != nil {
		return err
	}

	// store all manifests
	manifests := []Manifest{}
	filescache := []CachedFileInfo{}

	for _, file := range tarBalls {
		fmt.Printf("%s\n", file)

		// get MANIFEST.json + FILES.json
		fmap, err := utils.ExtractJSONFilesFromTarGz(file, []string{"MANIFEST.json", "FILES.json"})
		if err != nil {
			fmt.Printf("ERROR extracting %s\n", err)
			continue
		}

		var manifest Manifest
		var filesdata FilesMeta

		err = json.Unmarshal(fmap["MANIFEST.json"], &manifest)
		if err != nil {
			fmt.Printf("ERROR %s\n", err)
			continue
		}
		//fmt.Printf("%s\n", manifest)
		manifests = append(manifests, manifest)

		err = json.Unmarshal(fmap["FILES.json"], &filesdata)
		if err != nil {
			fmt.Printf("ERROR %s\n", err)
			continue
		}

		//fmt.Printf("%s\n", filesdata.Files[0])
		//namespace = manifest.Namespace
		//name = manifest.Name
		//version = manifest.Version

		for _, f := range filesdata.Files {
			//fmt.Printf("\t%s\n", f.Name)

			fDs := CachedFileInfo{
				Namespace:      manifest.CollectionInfo.Namespace,
				Name:           manifest.CollectionInfo.Name,
				Version:        manifest.CollectionInfo.Version,
				FileName:       f.Name,
				FileType:       f.FType,
				CheckSumSHA256: f.CheckSumSHA256,
			}

			filescache = append(filescache, fDs)

		}

	}

	// write manifests.tar.gz
	manifestsFilePath := filepath.Join(apath, "collection_manifests.tar.gz")
	fmt.Printf("write %s\n", manifestsFilePath)
	createManifestsTarGz(manifests, manifestsFilePath)

	// write files.tar.gz
	fmt.Printf("total files %d\n", len(filescache))
	cachedFilesPath := filepath.Join(apath, "collection_files.tar.gz")
	saveCachedFilesToGzippedFile(filescache, cachedFilesPath, 1000000)

	// write repodata.json
	currentTime := time.Now().UTC()
	isoFormattedCurrent := currentTime.Format(time.RFC3339)
	rMeta := RepoMeta{
		Date: isoFormattedCurrent,
		CollectionManifests: RepoMetaFile{
			Date:     isoFormattedCurrent,
			Filename: "collection_manifests.tar.gz",
		},
		CollectionFiles: RepoMetaFile{
			Date:     isoFormattedCurrent,
			Filename: "collection_files.tar.gz",
		},
	}

	// Marshal the RepoMeta instance to JSON
	jsonData, err := json.MarshalIndent(rMeta, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return err
	}

	// Write the JSON data to a file
	fn := filepath.Join(apath, "repometa.json")
	file, err := os.Create(fn)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	return nil
}

func SortManifestsByVersion(manifests []Manifest) ([]Manifest, error) {
	// Define a custom type for sorting
	type semverManifest struct {
		version  semver.Version
		manifest Manifest
	}

	// Convert manifests to semverManifests
	var semverManifests []semverManifest
	for _, manifest := range manifests {
		version, err := semver.Parse(manifest.CollectionInfo.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid version %s: %w", manifest.CollectionInfo.Version, err)
		}
		semverManifests = append(semverManifests, semverManifest{version: version, manifest: manifest})
	}

	// Sort the semverManifests by version
	sort.Slice(semverManifests, func(i, j int) bool {
		return semverManifests[i].version.LT(semverManifests[j].version)
	})

	// Extract sorted manifests
	sortedManifests := make([]Manifest, len(semverManifests))
	for i, semverManifest := range semverManifests {
		sortedManifests[i] = semverManifest.manifest
	}

	return sortedManifests, nil
}

func SortInstallSpecs(specs *[]utils.InstallSpec) {
	sort.Slice(*specs, func(i, j int) bool {
		if (*specs)[i].Namespace != (*specs)[j].Namespace {
			return (*specs)[i].Namespace < (*specs)[j].Namespace
		}
		if (*specs)[i].Name != (*specs)[j].Name {
			return (*specs)[i].Name < (*specs)[j].Name
		}
		return (*specs)[i].Version < (*specs)[j].Version
	})
}
