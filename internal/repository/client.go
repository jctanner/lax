package repository

import (
	//"cli/repository"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"lax/internal/utils"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blang/semver/v4"
	//"github.com/Masterminds/semver/v3"
)

type RepoClient interface {
	FetchRepoMeta(cachePath string) error
	//GetRepoMeta(cachePath string) (RepoMeta, error)
	GetRepoMetaDate() (string, error)
	ResolveCollectionDeps(spec utils.InstallSpec) ([]utils.InstallSpec, error)
	ResolveRoleDeps(spec utils.InstallSpec) ([]utils.InstallSpec, error)
	GetCacheFileLocationForInstallSpec(spec utils.InstallSpec) string
}

type FileRepoClient struct {
	CachePath           string
	BasePath            string
	Date                string
	RepoMeta            RepoMetaFile
	CollectionManifests RepoMetaFile
	CollectionFiles     RepoMetaFile
	RoleManifests RepoMetaFile
	RoleFiles     RepoMetaFile
}

type HttpRepoClient struct {
	CachePath           string
	BaseURL             string
	Date                string
	RepoMeta            RepoMetaFile
	CollectionManifests RepoMetaFile
	CollectionFiles     RepoMetaFile
	RoleManifests RepoMetaFile
	RoleFiles     RepoMetaFile
}

func (client *FileRepoClient) InitCache(cachePath string) error {
	return nil
}

func (client *FileRepoClient) FetchRepoMeta(cachePath string) error {

	client.CachePath = cachePath

	fmt.Printf("fetching repometa from %s\n", client.BasePath)

	// Construct the full path to the repometa.json file
	filePath := filepath.Join(client.BasePath, "repometa.json")

	// Read the file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the JSON data
	var repoMeta RepoMeta
	if err := json.Unmarshal(fileData, &repoMeta); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	fmt.Printf("repometa: %s\n", repoMeta)
	client.CollectionManifests = repoMeta.CollectionManifests
	client.CollectionFiles = repoMeta.CollectionFiles
	client.RoleManifests = repoMeta.RoleManifests
	client.RoleFiles = repoMeta.RoleFiles
	client.RepoMeta = RepoMetaFile{
		Filename: filePath,
		Date:     repoMeta.Date,
	}

	// copy repometa.json
	src := filepath.Join(client.BasePath, "repometa.json")
	dst := filepath.Join(cachePath, "repometa.json")
	utils.MakeDirs(cachePath)
	fmt.Printf("repometa: %s -> %s\n", src, dst)
	utils.CopyFile(src, dst)

	// copy the collection manifests
	src = filepath.Join(client.BasePath, client.CollectionManifests.Filename)
	dst = filepath.Join(cachePath, client.CollectionManifests.Filename)
	fmt.Printf("manifests: %s -> %s\n", src, dst)
	utils.CopyFile(src, dst)

	// copy the collection files
	src = filepath.Join(client.BasePath, client.CollectionFiles.Filename)
	dst = filepath.Join(cachePath, client.CollectionFiles.Filename)
	fmt.Printf("files: %s -> %s\n", src, dst)
	utils.CopyFile(src, dst)

	// copy the role manifests
	src = filepath.Join(client.BasePath, client.RoleManifests.Filename)
	dst = filepath.Join(cachePath, client.RoleManifests.Filename)
	fmt.Printf("manifests: %s -> %s\n", src, dst)
	utils.CopyFile(src, dst)

	// copy the role files
	src = filepath.Join(client.BasePath, client.RoleFiles.Filename)
	dst = filepath.Join(cachePath, client.RoleFiles.Filename)
	fmt.Printf("files: %s -> %s\n", src, dst)
	utils.CopyFile(src, dst)

	return nil
}

func (client *FileRepoClient) GetRepoMetaDate() (string, error) {
	// Construct the full path to the repometa.json file
	filePath := filepath.Join(client.BasePath, "repometa.json")

	// Read the file
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the JSON data
	var repoMeta RepoMeta
	if err := json.Unmarshal(fileData, &repoMeta); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	fmt.Printf("repometa: %s\n", repoMeta)
	client.CollectionManifests = repoMeta.CollectionManifests
	client.CollectionFiles = repoMeta.CollectionFiles

	client.RepoMeta = RepoMetaFile{
		Filename: filePath,
		Date:     repoMeta.Date,
	}

	return client.RepoMeta.Date, nil
}

func (client *FileRepoClient) GetCacheFileLocationForInstallSpec(spec utils.InstallSpec) string {
	tarName := fmt.Sprintf("%s-%s-%s.tar.gz", spec.Namespace, spec.Name, spec.Version)
	fileName := filepath.Join(client.BasePath, "collections", tarName)
	return fileName
}

func (client *FileRepoClient) ResolveCollectionDeps(spec utils.InstallSpec) ([]utils.InstallSpec, error) {

	// load the collections manifests
	collectionsManifestsFile := filepath.Join(client.BasePath, client.CollectionManifests.Filename)
	fmt.Printf("reading %s\n", collectionsManifestsFile)
	manifests, _ := ExtractCollectionManifestsFromTarGz(collectionsManifestsFile)
	specs := []utils.InstallSpec{}
	resolveCollectionDeps(spec, &manifests, &specs)

	return specs, nil
}

func (client *FileRepoClient) ResolveRoleDeps(spec utils.InstallSpec) ([]utils.InstallSpec, error) {

	// load the collections manifests
	rolesManifestsFile := filepath.Join(client.BasePath, client.RoleManifests.Filename)
	fmt.Printf("reading %s\n", rolesManifestsFile)
	manifests, _ := ExtractRoleManifestsFromTarGz(rolesManifestsFile)
	fmt.Printf("%s\n", manifests)
	panic("")
	specs := []utils.InstallSpec{}
	resolveRoleDeps(spec, &manifests, &specs)

	return specs, nil
}


func (client *HttpRepoClient) InitCache(cachePath string) error {
	return nil
}
func (client *HttpRepoClient) FetchRepoMeta(cachePath string) error {
	fmt.Printf("fetching repometa from %s\n", client.BaseURL)

	client.CachePath = cachePath

	// Construct the full url to the repometa.json file
	metaUrl := client.BaseURL + "/" + "repometa.json"
	cachedMetaFile := filepath.Join(client.CachePath, "repometa.json")
	fmt.Printf("rm: %s -> %s\n", metaUrl, cachedMetaFile)
	err := DownloadFile(metaUrl, cachedMetaFile)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return fmt.Errorf("failed to download file: %w", err)
	}

	// Read the file
	fileData, err := ioutil.ReadFile(cachedMetaFile)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the JSON data
	var repoMeta RepoMeta
	if err := json.Unmarshal(fileData, &repoMeta); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	fmt.Printf("repometa: %s\n", repoMeta)
	client.CollectionManifests = repoMeta.CollectionManifests
	client.CollectionFiles = repoMeta.CollectionFiles
	client.RoleManifests = repoMeta.RoleManifests
	client.RoleFiles = repoMeta.RoleFiles

	//filesToGet := []string{client.CollectionManifests.Filename, client.CollectionFiles.Filename}
	filesToGet := []string{client.CollectionManifests.Filename, client.RoleManifests.Filename}
	fmt.Printf("%s\n", filesToGet)

	for _, fn := range filesToGet {
		localFile := filepath.Join(client.CachePath, fn)
		url := client.BaseURL + "/" + fn
		fmt.Printf("rm: %s -> %s\n", url, localFile)
		err := DownloadFile(url, localFile)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return fmt.Errorf("failed to download file: %w", err)
		}
	}

	return nil
}

func (client *HttpRepoClient) GetRepoMetaDate() (string, error) {
	return client.RepoMeta.Date, nil
}

func (client *HttpRepoClient) GetCacheFileLocationForInstallSpec(spec utils.InstallSpec) string {
	//fmt.Printf("ERROR: GetCacheFileLocationForInstallSpec NOT YET IMPLEMENTED\n")
	cDir := filepath.Join(client.CachePath, "collections")
	utils.MakeDirs(cDir)

	tarName := fmt.Sprintf("%s-%s-%s.tar.gz", spec.Namespace, spec.Name, spec.Version)
	fmt.Printf("%s\n", tarName)

	cFile := filepath.Join(cDir, tarName)
	if utils.FileExists(cFile) {
		return cFile
	}

	// download it ...
	url := client.BaseURL + "/collections/" + tarName
	DownloadFile(url, cFile)

	return cFile
}

func (client *HttpRepoClient) ResolveCollectionDeps(spec utils.InstallSpec) ([]utils.InstallSpec, error) {
	// load the collections manifests
	collectionsManifestsFile := filepath.Join(client.CachePath, client.CollectionManifests.Filename)
	fmt.Printf("reading %s\n", collectionsManifestsFile)
	manifests, _ := ExtractCollectionManifestsFromTarGz(collectionsManifestsFile)
	specs := []utils.InstallSpec{}
	resolveCollectionDeps(spec, &manifests, &specs)

	return specs, nil
}

func (client *HttpRepoClient) ResolveRoleDeps(spec utils.InstallSpec) ([]utils.InstallSpec, error) {

	// load the collections manifests
	rolesManifestsFile := filepath.Join(client.CachePath, client.RoleManifests.Filename)
	fmt.Printf("reading %s\n", rolesManifestsFile)
	manifests, _ := ExtractRoleManifestsFromTarGz(rolesManifestsFile)
	specs := []utils.InstallSpec{}
	resolveRoleDeps(spec, &manifests, &specs)

	return specs, nil
}

func GetRepoClient(repo string, cachePath string) (RepoClient, error) {
	if utils.IsURL(repo) {
		return &HttpRepoClient{BaseURL: repo, CachePath: cachePath}, nil
	} else if utils.IsDir(repo) {
		return &FileRepoClient{BasePath: repo, CachePath: cachePath}, nil
	} else {
		return nil, fmt.Errorf("unsupported repo format")
	}
}

func resolveCollectionDeps(spec utils.InstallSpec, manifests *[]CollectionManifest, specs *[]utils.InstallSpec) {

	candidates := SpecToManifestCandidates(spec, manifests)

	// exit early if nothing was found
	if len(candidates) == 0 {
		fmt.Printf("ERROR2: found no candidates for %s %s ... %d\n", spec, candidates, len(candidates))
		return
	}

	// sort by version
	sortedManifests, _ := SortManifestsByVersion(candidates)

	if len(sortedManifests) == 0 {
		fmt.Printf("ERROR: found no candidates for %s\n", spec)
		return
	}

	// use the latest version
	thisManifest := sortedManifests[len(sortedManifests)-1]
	fmt.Printf("%s.%s latest: %s\n", thisManifest.CollectionInfo.Namespace, thisManifest.CollectionInfo.Name, thisManifest.CollectionInfo.Version)
	thisSpec := utils.InstallSpec{
		Namespace: thisManifest.CollectionInfo.Namespace,
		Name:      thisManifest.CollectionInfo.Name,
		Version:   thisManifest.CollectionInfo.Version,
	}
	if specListContainsNamespaceName(specs, thisSpec) {
		return
	}
	if specListContainsNamespaceNameVersion(specs, thisSpec) {
		return
	}
	*specs = append(*specs, thisSpec)

	// what are the 1st order deps
	for j, d := range thisManifest.CollectionInfo.Dependencies {
		fmt.Printf("\tdep: %s %s\n", j, d)
		parts := strings.Split(j, ".")
		fmt.Printf("\t\t%s\n", parts)
		dSpec := utils.InstallSpec{
			Namespace: parts[0],
			Name:      parts[1],
			Version:   d,
		}
		fmt.Printf("\t\t%s\n", dSpec)
		resolveCollectionDeps(dSpec, manifests, specs)
	}

	// sort the specs
	SortInstallSpecs(specs)

	// check for duplicates ... ?
	DeduplicateSpecs(specs)

	//return specs

}

func resolveRoleDeps(spec utils.InstallSpec, manifests *[]RoleMeta, specs *[]utils.InstallSpec) {

	candidates := RoleSpecToManifestCandidates(spec, manifests)

	// exit early if nothing was found
	if len(candidates) == 0 {
		fmt.Printf("ERROR2: found no candidates for %s %s ... %d\n", spec, candidates, len(candidates))
		return
	}

	// sort by version
	sortedManifests, _ := SortRoleManifestsByVersion(candidates)

	if len(sortedManifests) == 0 {
		fmt.Printf("ERROR: found no candidates for %s\n", spec)
		return
	}

	// use the latest version
	thisManifest := sortedManifests[len(sortedManifests)-1]
	fmt.Printf("%s.%s latest: %s\n", thisManifest.GalaxyInfo.Namespace, thisManifest.GalaxyInfo.RoleName, thisManifest.GalaxyInfo.Version)
	thisSpec := utils.InstallSpec{
		Namespace: thisManifest.GalaxyInfo.Namespace,
		Name:      thisManifest.GalaxyInfo.RoleName,
		Version:   thisManifest.GalaxyInfo.Version,
	}
	if specListContainsNamespaceName(specs, thisSpec) {
		return
	}
	if specListContainsNamespaceNameVersion(specs, thisSpec) {
		return
	}
	*specs = append(*specs, thisSpec)

	// what are the 1st order deps
	for j, d := range thisManifest.GalaxyInfo.Dependencies {
		fmt.Printf("\tdep: %s %s\n", j, d)
		panic("")
		/*
		parts := strings.Split(j, ".")
		fmt.Printf("\t\t%s\n", parts)
		dSpec := utils.InstallSpec{
			Namespace: parts[0],
			Name:      parts[1],
			Version:   d,
		}
		fmt.Printf("\t\t%s\n", dSpec)
		resolveRoleDeps(dSpec, manifests, specs)
		*/
	}

	// sort the specs
	SortInstallSpecs(specs)

	// check for duplicates ... ?
	DeduplicateSpecs(specs)

	//return specs

}

/*
Given a utils.InstallSpec, reduce a list of repository.Manifest down to the matching
candidates via their namespace, name and version (which can include an operator)
*/
func SpecToManifestCandidates(spec utils.InstallSpec, manifests *[]CollectionManifest) []CollectionManifest {
	// get the meta for the incoming spec
	candidates := []CollectionManifest{}
	for _, manifest := range *manifests {
		if manifest.CollectionInfo.Namespace != spec.Namespace {
			continue
		}
		if manifest.CollectionInfo.Name != spec.Name {
			continue
		}

		// is it a version with an operator or a range?
		// "*" means -any-
		if spec.Version != "" && spec.Version != "*" {

			// make a comparable semver ...
			op, v2, _ := splitVersion(spec.Version)
			specVer, _ := semver.Make(v2)

			// make a comparable semver ...
			_, m2, _ := splitVersion(manifest.CollectionInfo.Version)
			mVer, _ := semver.Make(m2)

			// eval the conditional and skip if false ...
			res, _ := utils.CompareSemVersions(op, &mVer, &specVer)
			if !res {
				continue
			}
		}

		candidates = append(candidates, manifest)
	}

	return candidates
}

func RoleSpecToManifestCandidates(spec utils.InstallSpec, manifests *[]RoleMeta) []RoleMeta {

	fmt.Printf("##############################################\n")

	fmt.Printf("%s\n", manifests)

	// get the meta for the incoming spec
	candidates := []RoleMeta{}
	for _, manifest := range *manifests {
		if manifest.GalaxyInfo.Namespace != spec.Namespace {
			fmt.Printf("skip %s\n", manifest)
			continue
		}
		if manifest.GalaxyInfo.RoleName != spec.Name {
			fmt.Printf("skip %s\n", manifest)
			continue
		}

		// is it a version with an operator or a range?
		// "*" means -any-
		if spec.Version != "" && spec.Version != "*" {

			// make a comparable semver ...
			op, v2, _ := splitVersion(spec.Version)
			specVer, _ := semver.Make(v2)

			// make a comparable semver ...
			_, m2, _ := splitVersion(manifest.GalaxyInfo.Version)
			mVer, _ := semver.Make(m2)

			// eval the conditional and skip if false ...
			res, _ := utils.CompareSemVersions(op, &mVer, &specVer)
			if !res {
				continue
			}
		}

		candidates = append(candidates, manifest)
	}

	fmt.Printf("##############################################\n")
	return candidates
}

/*
Split an arbitrary version string found in collection dependencies.
The first part of the string could be a comparison operator.
*/
func splitVersion(versionStr string) (string, string, error) {
	re := regexp.MustCompile(`(>=|<=|>|<|=)?\s*(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(versionStr)

	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid version string")
	}

	operator := matches[1]
	if operator == "" {
		operator = "=" // Default operator if none is provided
	}
	version := matches[2]

	return operator, version, nil
}

func specListContainsNamespaceNameVersion(specs *[]utils.InstallSpec, newSpec utils.InstallSpec) bool {
	for _, spec := range *specs {
		if spec.Equals(newSpec) {
			return true
		}
	}
	return false
}

func specListContainsNamespaceName(specs *[]utils.InstallSpec, newSpec utils.InstallSpec) bool {
	for _, spec := range *specs {
		if spec.Namespace == newSpec.Namespace && spec.Name == newSpec.Name {
			return true
		}
	}
	return false
}

func specFQNMatch(a utils.InstallSpec, b utils.InstallSpec) bool {
	return false
}

func DeduplicateSpecs(specs *[]utils.InstallSpec) {
	// only keep one version for each namespace.name ...

	//toKeep := []utils.InstallSpec
	//toDrop := []utils.InstallSpec

}

func DownloadFile(url, dest string) error {
	// Define the directory
	destDir := filepath.Dir(dest)
	err := utils.MakeDirs(destDir)
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}

	// Create the file
	outFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create file: %v", err)
	}
	defer outFile.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("get url: %v", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("write to file: %v", err)
	}

	return nil
}
