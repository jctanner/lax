package repository

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jctanner/lax/internal/types"
	"github.com/jctanner/lax/internal/utils"
	"github.com/sirupsen/logrus"

	"encoding/gob"
	"encoding/json"

	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"

	"github.com/blang/semver/v4"
	"gopkg.in/yaml.v2"
)

/***************************************************************
GLOBAL
***************************************************************/

type RepoMeta struct {
	Date                string       `json:"date"`
	CollectionManifests RepoMetaFile `json:"collection_manifests"`
	CollectionFiles     RepoMetaFile `json:"collection_files"`
	RoleManifests       RepoMetaFile `json:"role_manifests"`
	RoleFiles           RepoMetaFile `json:"role_files"`
}

type RepoMetaFile struct {
	Date     string `json:"date"`
	Filename string `json:"filename"`
}

/***************************************************************
COLLECTIONS
***************************************************************/

type CollectionManifest struct {
	CollectionInfo CollectionInfo `json:"collection_info"`
}

type CollectionInfo struct {
	Namespace    string            `json:"namespace"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

type CollectionFilesMeta struct {
	Files []CollectionFileInfo `json:"files"`
}

type CollectionFileInfo struct {
	Name           string `json:"name"`
	FType          string `json:"ftype"`
	CheckSumType   string `json:"chksum_type"`
	CheckSumSHA256 string `json:"chksum_sha256"`
}

type CollectionCachedFileInfo struct {
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	FileName       string `json:"filename"`
	FileType       string `json:"filetype"`
	CheckSumSHA256 string `json:"chksum_sha256"`
}

/***************************************************************
ROLES
***************************************************************/

type RoleCachedFileInfo struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	FileName  string `json:"filename"`
	FileType  string `json:"filetype"`
	//CheckSumType   string `json:"chksum_type"`
	//CheckSumSHA256 string `json:"chksum_sha256"`
}

func createCollectionManifestsTarGz(manifests []CollectionManifest, tarGzPath string) error {
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

func ExtractCollectionManifestsFromTarGz(tarGzPath string) ([]CollectionManifest, error) {
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

	var manifests []CollectionManifest

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
		var manifest CollectionManifest
		if err := json.Unmarshal(jsonData, &manifest); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
		}

		// Append the manifest to the slice
		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

func ExtractRoleManifestsFromTarGz(tarGzPath string) ([]types.RoleMeta, error) {
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

	var manifests []types.RoleMeta

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
		var manifest types.RoleMeta
		if err := json.Unmarshal(jsonData, &manifest); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
		}

		// Append the manifest to the slice
		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

func createRoleMetaTarGz(manifests []types.RoleMeta, tarGzPath string) error {
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

func saveCachedCollectionFilesToGzippedFile(manifests []CollectionCachedFileInfo, filePath string, chunkSize int) error {

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

func saveCachedRoleFilesToGzippedFile(manifests []RoleCachedFileInfo, filePath string, chunkSize int) error {

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

func SortManifestsByVersion(manifests []CollectionManifest) ([]CollectionManifest, error) {
	// Define a custom type for sorting
	type semverManifest struct {
		version  semver.Version
		manifest CollectionManifest
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
	sortedManifests := make([]CollectionManifest, len(semverManifests))
	for i, semverManifest := range semverManifests {
		sortedManifests[i] = semverManifest.manifest
	}

	return sortedManifests, nil
}

func SortRoleManifestsByVersion(manifests []types.RoleMeta) ([]types.RoleMeta, error) {
	// Define a custom type for sorting
	type semverManifest struct {
		version  semver.Version
		manifest types.RoleMeta
	}

	// Convert manifests to semverManifests
	var semverManifests []semverManifest
	for _, manifest := range manifests {
		version, err := semver.Parse(manifest.GalaxyInfo.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid version %s: %w", manifest.GalaxyInfo.Version, err)
		}
		semverManifests = append(semverManifests, semverManifest{version: version, manifest: manifest})
	}

	// Sort the semverManifests by version
	sort.Slice(semverManifests, func(i, j int) bool {
		return semverManifests[i].version.LT(semverManifests[j].version)
	})

	// Extract sorted manifests
	sortedManifests := make([]types.RoleMeta, len(semverManifests))
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

/*
func displayLinedYaml(rawyaml string) {
	fmt.Printf("===========================\n")
	lines := strings.Split(rawyaml, "\n")
	for ix, line := range lines {
		fmt.Printf("%d:%s\n", ix+1, line)
	}
}
*/

func displayLinedYaml(text string) {
	fmt.Printf("===========================\n")
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		fmt.Printf("%3d: %q\n", i+1, line)
	}
}

func GetRoleMetaFromTarball(f string) (types.RoleMeta, error) {

	var meta types.RoleMeta

	tarFileNames, _ := utils.ListFilenamesInTarGz(f)

	metaFile := ""
	for _, tfn := range tarFileNames {
		if utils.EndsWithMetaMainYAML(tfn) {
			metaFile = tfn
			break
		}
	}

	fmap, err := utils.ExtractFilesFromTarGz(f, []string{metaFile})
	if err != nil {
		fmt.Printf("ERROR extracting %s\n", err)
		return meta, err
	}

	//fmt.Printf("raw:\n%s\n", fmap[metaFile])

	err = yaml.Unmarshal(fmap[metaFile], &meta)

	if err != nil {
		fmt.Printf("ERROR_0 %s %s %s\n", f, metaFile, err)
		rawstring := string(fmap[metaFile])
		/*
			lines := strings.Split(rawstring, "\n")
				for ix, line := range lines {
					fmt.Printf("%d:%s\n", ix+1, line)
				}
		*/
		displayLinedYaml(rawstring)
		logrus.Warnf("unmarshalling error %s %s %s\n", f, metaFile, err)
	}

	if err != nil {
		// fix indentation if possible ...
		if strings.Contains(err.Error(), "did not find expected key") || strings.Contains(err.Error(), " mapping values are not allowed in this context") || strings.Contains(err.Error(), " cannot unmarshal !!str") || true {
			fmt.Printf("FIXING YAML IN MEMORY...\n")
			rawstring := string(fmap[metaFile])

			// these were hiding in someone's depependency definition ...
			rawstring = strings.ReplaceAll(rawstring, "\u00a0", " ")

			//displayLinedYaml(rawstring)
			newstring, _ := utils.FixGalaxyIndentation(rawstring)
			//displayLinedYaml(newstring)
			newstring = utils.AddQuotesToDescription(newstring)
			//displayLinedYaml(newstring)
			newstring = utils.AddLiteralBlockScalarToTags(newstring)
			//displayLinedYaml(newstring)

			if strings.Contains(err.Error(), "did not find expected key") {
				newstring = utils.FixPlatformVersion(newstring)
				displayLinedYaml(newstring)
			}

			newstring = utils.ReplaceDependencyRoleWithName(newstring)

			newstring = utils.RemoveComments(newstring)
			//displayLinedYaml(newstring)

			displayLinedYaml(newstring)
			//fmt.Println("trying to unmarshall munged data ...")

			var meta2 types.RoleMeta
			err2 := yaml.Unmarshal([]byte(newstring), &meta2)
			if err2 == nil {
				return meta2, nil
			}

			//fmt.Printf("ERROR_2 %s %s (%s)\n", f, metaFile, err2)
			/*
				lines := strings.Split(newstring, "\n")
				for ix, line := range lines {
					fmt.Printf("%d:%s\n", ix+1, line)
				}
			*/
			displayLinedYaml(newstring)
			fmt.Printf("ERROR_2 %s %s [[[%s]]]\n", f, metaFile, err2)
			//return meta2, err
			panic("COUNT NOT UNMARSHALL")
		}
	}

	if err != nil {

		fmt.Printf("ERROR_1 %s %s %s\n", f, metaFile, err)
		//fmt.Printf("RAW:\n%s\n", fmap[metaFile])
		rawstring := string(fmap[metaFile])
		lines := strings.Split(rawstring, "\n")
		for ix, line := range lines {
			fmt.Printf("%d:%s\n", ix+1, line)
		}
		fmt.Printf("ERROR_2 %s %s %s\n", f, metaFile, err)
		os.Remove(f)
		panic("")
		//return meta, err
	}

	return meta, nil
}

/*
func extractRoleVersionFromTarName(path string) string {
	// Split the path into components
	components := strings.Split(path, "/")
	// Get the last component (filename)
	filename := components[len(components)-1]
	// Split the filename by hyphens
	parts := strings.Split(filename, "-")
	// Get the part before the last part and the extension
	versionWithExt := parts[len(parts)-1]
	// Remove the extension from the version
	version := strings.TrimSuffix(versionWithExt, filepath.Ext(versionWithExt))
	return version
}
*/

func extractRoleVersionFromTarName(path string) string {
	// Split the path into components
	components := strings.Split(path, "/")
	// Get the last component (filename)
	filename := components[len(components)-1]
	// Remove the ".tar.gz" extension
	filenameWithoutExt := strings.TrimSuffix(filename, ".tar.gz")
	// Split the filename by hyphens
	parts := strings.Split(filenameWithoutExt, "-")
	// Get the last part which should be the version
	version := parts[len(parts)-1]
	return version
}

func extractRoleNamespaceFromTarName(path string) string {
	// Split the path into components
	components := strings.Split(path, "/")
	// Get the last component (filename)
	filename := components[len(components)-1]
	// Split the filename by hyphens
	parts := strings.Split(filename, "-")
	return parts[0]
}

func extractRoleNameFromTarName(path string) string {
	// Split the path into components
	components := strings.Split(path, "/")
	// Get the last component (filename)
	filename := components[len(components)-1]
	// Split the filename by hyphens
	parts := strings.Split(filename, "-")
	return parts[1]
}
