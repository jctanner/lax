package repository

import (
	//"cli/repository"
	"lax/internal/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	ResolveDeps(spec utils.InstallSpec) ([]utils.InstallSpec, error)
}

type FileRepoClient struct {
	CachePath           string
	BasePath            string
	Date                string
	RepoMeta            RepoMetaFile
	CollectionManifests RepoMetaFile
	CollectionFiles     RepoMetaFile
}

type HttpRepoClient struct {
	CachePath           string
	BaseURL             string
	Date                string
	RepoMeta            RepoMetaFile
	CollectionManifests RepoMetaFile
	CollectionFiles     RepoMetaFile
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
	fileData, err := ioutil.ReadFile(filePath)
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

	client.RepoMeta = RepoMetaFile{
		Filename: filePath,
		Date:     repoMeta.Date,
	}

	// copy repometa.json
	src := filepath.Join(client.BasePath, "repometa.json")
	dst := filepath.Join(cachePath, "repometa.json")
	fmt.Printf("repometa: %s -> %s\n", src, dst)
	utils.CopyFile(src, dst)

	// copy the manifests
	src = filepath.Join(client.BasePath, client.CollectionManifests.Filename)
	dst = filepath.Join(cachePath, client.CollectionManifests.Filename)
	fmt.Printf("manifests: %s -> %s\n", src, dst)
	utils.CopyFile(src, dst)

	// copy the files
	src = filepath.Join(client.BasePath, client.CollectionFiles.Filename)
	dst = filepath.Join(cachePath, client.CollectionFiles.Filename)
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

func (client *FileRepoClient) ResolveDeps(spec utils.InstallSpec) ([]utils.InstallSpec, error) {

	// load the collections manifests
	collectionsManifestsFile := filepath.Join(client.BasePath, client.CollectionManifests.Filename)
	fmt.Printf("reading %s\n", collectionsManifestsFile)
	manifests, _ := ExtractManifestsFromTarGz(collectionsManifestsFile)
	specs := []utils.InstallSpec{}
	resolveDeps(spec, &manifests, &specs)

	return specs, nil
}

func (client *HttpRepoClient) InitCache(cachePath string) error {
	return nil
}
func (client *HttpRepoClient) FetchRepoMeta(cachePath string) error {
	fmt.Printf("fetching repometa from %s\n", client.BaseURL)
	return nil
}

func (client *HttpRepoClient) GetRepoMetaDate() (string, error) {
	return client.RepoMeta.Date, nil
}

func (client *HttpRepoClient) ResolveDeps(spec utils.InstallSpec) ([]utils.InstallSpec, error) {
	specs := []utils.InstallSpec{}
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

func resolveDeps(spec utils.InstallSpec, manifests *[]Manifest, specs *[]utils.InstallSpec) {
	//fmt.Printf("PROCESS: %s\n", spec)
	//specs := []utils.InstallSpec{}

	// get the meta for the incoming spec
	candidates := []Manifest{}
	for _, manifest := range *manifests {
		if manifest.CollectionInfo.Namespace != spec.Namespace {
			continue
		}
		if manifest.CollectionInfo.Name != spec.Name {
			continue
		}

		//fmt.Printf("\tcheck %s\n", manifest)

		// is it a version with an operator or a range?
		if spec.Version != "" && spec.Version != "*" {

			op, v2, _ := splitVersion(spec.Version)
			specVer, _ := semver.Make(v2)

			_, m2, _ := splitVersion(manifest.CollectionInfo.Version)
			mVer, _ := semver.Make(m2)

			//fmt.Printf("\t\t%s %s %s\n", specVer, op, mVer)

			/*
				if spec.Version != "" && manifest.CollectionInfo.Version != spec.Version {
					fmt.Printf("\tignoring %s.%s %s\n", manifest.CollectionInfo.Namespace, manifest.CollectionInfo.Name, manifest.CollectionInfo.Version)
					continue
				}
			*/

			//res, _ := compareVersions(op, mVer, specVer)
			res, _ := utils.CompareSemVersions(op, &mVer, &specVer)
			if !res {
				//fmt.Printf("\tignoring %s.%s %s\n", manifest.CollectionInfo.Namespace, manifest.CollectionInfo.Name, manifest.CollectionInfo.Version)
				continue
			}
		}
		candidates = append(candidates, manifest)
	}

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
	if specListContains(specs, thisSpec) {
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
		resolveDeps(dSpec, manifests, specs)
	}

	// sort the specs
	SortInstallSpecs(specs)

	//return specs

}

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

/*
func compareVersions(op string, v1 *semver.Version, v2 *semver.Version) (bool, error) {
	switch op {
	case ">":
		return v1.GreaterThan(v2), nil
	case ">=":
		return v1.GreaterThan(v2) || v1.Equal(v2), nil
	case "<":
		return v1.LessThan(v2), nil
	case "<=":
		return v1.LessThan(v2) || v1.Equal(v2), nil
	case "=":
		return v1.Equal(v2), nil
	default:
		return false, fmt.Errorf("invalid operator: %s", op)
	}
}
*/

func specListContains(specs *[]utils.InstallSpec, newSpec utils.InstallSpec) bool {
	for _, spec := range *specs {
		if spec.Equals(newSpec) {
			return true
		}
	}
	return false
}
