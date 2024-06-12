package repository

import (
	"fmt"
	"io"
	"lax/internal/utils"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"encoding/gob"
	"encoding/json"

	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"

	"time"

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
	RoleManifests RepoMetaFile `json:"role_manifests"`
	RoleFiles     RepoMetaFile `json:"role_files"`
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
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	FileName           string `json:"filename"`
	FileType          string `json:"filetype"`
	//CheckSumType   string `json:"chksum_type"`
	//CheckSumSHA256 string `json:"chksum_sha256"`
}

type RoleMeta struct {
	GalaxyInfo GalaxyInfo  `yaml:"galaxy_info"`
}

type GalaxyInfo struct {
	Author string `yaml:"author"`
	Namespace string `yaml:"namespace"`
	RoleName string `yaml:"role_name"`

	// this doesn't actually exist
	// but we want to have a settable
	// property for the index files
	Version string `yaml:"version"`
	
	Description string `yaml:"description"`
	License RoleLicense `yaml:"license"`
	MinAnsibleVersion string `json:"min_ansible_version"`
	Platforms        []RolePlatform `yaml:"platforms"`
	GalaxyTags       []string   `yaml:"galaxy_tags"`
	Dependencies []RoleDependency `yaml:"dependencies"`
}

type RolePlatform struct {
    Name     string   `yaml:"name"`
    Versions []string `yaml:"versions"`
}

type RolePlatformVersions []string

func (rp *RolePlatform) UnmarshalYAML(unmarshal func(interface{}) error) error {
    var temp struct {
        Name     string      `yaml:"name"`
        Versions interface{} `yaml:"versions"`
    }

    if err := unmarshal(&temp); err != nil {
        return err
    }

    rp.Name = temp.Name

    switch v := temp.Versions.(type) {
    case string:
        rp.Versions = RolePlatformVersions{v}
    case int:
        rp.Versions = RolePlatformVersions{strconv.Itoa(v)}
	case float64:
		vstring := strconv.FormatFloat(v, 'f', -1, 64)
		rp.Versions = RolePlatformVersions{vstring}
	case nil:
		rp.Versions = RolePlatformVersions{}
    case []interface{}:
        var versions RolePlatformVersions
        for _, version := range v {
            switch v := version.(type) {
            case string:
                versions = append(versions, v)
            case int:
                versions = append(versions, strconv.Itoa(v))
			case float64:
				vstring := strconv.FormatFloat(v, 'f', -1, 64)
				rp.Versions = append(versions, vstring)
            default:
                return fmt.Errorf("unexpected type for platform-version: %T", version)
            }
        }
        rp.Versions = versions
    default:
        return fmt.Errorf("unexpected type for platform-versions: %T", temp.Versions)
    }

    return nil
}

type RoleLicense []string

func (l *RoleLicense) UnmarshalYAML(unmarshal func(interface{}) error) error {
    var singleLicense string
    if err := unmarshal(&singleLicense); err == nil {
        *l = RoleLicense{singleLicense}
        return nil
    }

    var licenseList []string
    if err := unmarshal(&licenseList); err == nil {
        *l = RoleLicense(licenseList)
        return nil
    }

    return fmt.Errorf("failed to unmarshal License")
}

// Dependency can be either a string or a map
type RoleDependency struct {
	Src  string
	Name string
}

func (d *RoleDependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var depString string
	if err := unmarshal(&depString); err == nil {
		d.Src = depString
		d.Name = depString
		return nil
	}

	var depMap map[string]string
	if err := unmarshal(&depMap); err == nil {
		d.Src = depMap["src"]
		d.Name = depMap["name"]
		return nil
	}

	return fmt.Errorf("failed to unmarshal Dependency")
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


func createRoleMetaTarGz(manifests []RoleMeta, tarGzPath string) error {
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


func CreateRepo(dest string, roles_only bool, collectios_only bool) error {

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
	err = processCollections(apath, collectionsPath)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}

	// assert it has a collections subdir
	rolesPath := filepath.Join(apath, "roles")
	err = processRoles(apath, rolesPath)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}

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

func processCollections(basePath string, collectionsPath string) (error) {

	if !utils.IsDir(collectionsPath) {
		fmt.Printf("%s is not a directory\n", collectionsPath)
		return nil
		//hasCollections = false
	}

	// make a list of tarballs
	collectionTarBalls, err := utils.ListTarGzFiles(collectionsPath)
	if err != nil {
		return err
	}

	// we need the metdata from each file
	metadataDir := filepath.Join(basePath, "metadata")
	err = utils.MakeDirs(metadataDir)
	if err != nil {
		return err
	}

	// store all collectionManifests
	collectionManifests := []CollectionManifest{}
	collectionFilesCache := []CollectionCachedFileInfo{}

	for _, file := range collectionTarBalls {
		fmt.Printf("%s\n", file)

		// get MANIFEST.json + FILES.json
		fmap, err := utils.ExtractJSONFilesFromTarGz(file, []string{"MANIFEST.json", "FILES.json"})
		if err != nil {
			fmt.Printf("ERROR extracting %s\n", err)
			continue
		}

		var manifest CollectionManifest
		var filesdata CollectionFilesMeta

		err = json.Unmarshal(fmap["MANIFEST.json"], &manifest)
		if err != nil {
			fmt.Printf("ERROR %s\n", err)
			continue
		}
		//fmt.Printf("%s\n", manifest)
		collectionManifests = append(collectionManifests, manifest)

		err = json.Unmarshal(fmap["FILES.json"], &filesdata)
		if err != nil {
			fmt.Printf("ERROR %s\n", err)
			continue
		}

		for _, f := range filesdata.Files {
			fDs := CollectionCachedFileInfo{
				Namespace:      manifest.CollectionInfo.Namespace,
				Name:           manifest.CollectionInfo.Name,
				Version:        manifest.CollectionInfo.Version,
				FileName:       f.Name,
				FileType:       f.FType,
				CheckSumSHA256: f.CheckSumSHA256,
			}

			collectionFilesCache = append(collectionFilesCache, fDs)

		}

	}

	// write manifests.tar.gz
	collectionManifestsFilePath := filepath.Join(basePath, "collection_manifests.tar.gz")
	fmt.Printf("write %s\n", collectionManifestsFilePath)
	createCollectionManifestsTarGz(collectionManifests, collectionManifestsFilePath)

	// write files.tar.gz
	fmt.Printf("total files %d\n", len(collectionFilesCache))
	collectionsCachedFilesPath := filepath.Join(basePath, "collection_files.tar.gz")
	saveCachedCollectionFilesToGzippedFile(collectionFilesCache, collectionsCachedFilesPath, 1000000)

	return nil
}

func processRoles(basePath string, rolesPath string) (error) {
	if !utils.IsDir(rolesPath) {
		fmt.Printf("%s is not a directory\n", rolesPath)
		return nil
		//hasCollections = false
	}

	// make a list of tarballs
	roleTarBalls, err := utils.ListTarGzFiles(rolesPath)
	if err != nil {
		return err
	}

	// we need the metdata from each file
	metadataDir := filepath.Join(basePath, "metadata")
	err = utils.MakeDirs(metadataDir)
	if err != nil {
		return err
	}

	// store all collectionManifests
	roleMeta := []RoleMeta{}
	roleFilesCache := []RoleCachedFileInfo{}

	for _, f := range roleTarBalls {
		fmt.Printf("tar: %s\n", f)

		version := extractRoleVersionFromTarName(f)

		rmeta, _ := GetRoleMetaFromTarball(f)
		rmeta.GalaxyInfo.Version = version

		tarFileNames, _ := utils.ListFilenamesInTarGz(f)
		for _, tfn := range tarFileNames {
			fmt.Printf("%s\n", tfn)
			relativePath := utils.RemoveFirstPathElement(tfn)
			fmt.Printf("\t%s\n", relativePath)

			cf := RoleCachedFileInfo{
				Namespace: rmeta.GalaxyInfo.Namespace,
				Name: rmeta.GalaxyInfo.Namespace,
				Version: version,
				FileName: tfn,
			}
			roleFilesCache = append(roleFilesCache, cf)
		}

	}

	// write manifests.tar.gz
	roleMetaFilePath := filepath.Join(basePath, "role_manifests.tar.gz")
	fmt.Printf("write %s\n", roleMetaFilePath)
	createRoleMetaTarGz(roleMeta, roleMetaFilePath)

	// write files.tar.gz
	fmt.Printf("total files %d\n", len(roleFilesCache))
	roleCachedFilesPath := filepath.Join(basePath, "role_files.tar.gz")
	saveCachedRoleFilesToGzippedFile(roleFilesCache, roleCachedFilesPath, 1000000)

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


func GetRoleMetaFromTarball(f string) (RoleMeta, error){

	var meta RoleMeta

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
		// fix indentation if possible ...
		if strings.Contains(err.Error(), "did not find expected key") || strings.Contains(err.Error(), " mapping values are not allowed in this context") {
			fmt.Printf("FIXING YAML IN MEMORY...\n")
			rawstring := string(fmap[metaFile])
			newstring, _ := utils.FixGalaxyIndentation(rawstring)
			newstring = utils.AddQuotesToDescription(newstring)

			var meta2 RoleMeta
			err = yaml.Unmarshal([]byte(newstring), &meta2)
			if err == nil {
				return meta2, nil
			}

			fmt.Printf("ERROR %s %s %s\n", f, metaFile, err)
			lines := strings.Split(newstring, "\n")
			for ix, line := range lines {
				fmt.Printf("%d:%s\n", ix + 1, line)
			}
			fmt.Printf("ERROR %s %s %s\n", f, metaFile, err)
			//return meta2, err
			panic("")
		}
	}

	if err != nil {

		fmt.Printf("ERROR %s %s %s\n", f, metaFile, err)
		//fmt.Printf("RAW:\n%s\n", fmap[metaFile])
		rawstring := string(fmap[metaFile])
		lines := strings.Split(rawstring, "\n")
		for ix, line := range lines {
			fmt.Printf("%d:%s\n", ix + 1, line)
		}
		fmt.Printf("ERROR %s %s %s\n", f, metaFile, err)
		os.Remove(f)
		panic("")
		//return meta, err
	}

	return meta, nil
}




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
