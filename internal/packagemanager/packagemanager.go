package packagemanager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"lax/internal/repository"
	"lax/internal/utils"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

// gets stored with an installed collection ...
type GalaxyYamlMeta struct {
	DownloadUrl   string   `json:"download_url"`
	FormatVersion string   `json:"format_version"`
	Name          string   `json:"name"`
	Namespace     string   `json:"namespace"`
	Server        string   `json:"server"`
	Signatures    []string `json:"signatures"`
	Version       string   `json:"version"`
	VersionUrl    string   `json:"version_url"`
}

type RoleInstallInfo struct {
	InstallDate string `yaml:"install_date"`
	Version string `yaml:"version"`
}

type PackageManager struct {
	BasePath            string
	CachePath           string
	RepoMeta            repository.RepoMetaFile
	CollectionManifests repository.RepoMetaFile
	CollectionFiles     repository.RepoMetaFile
}

func (pkgmgr *PackageManager) Initialize() error {

	// where everything goes
	pkgmgr.BasePath, _ = utils.GetAbsPath(pkgmgr.BasePath)
	utils.MakeDirs(pkgmgr.BasePath)

	// store meta here
	//pkgmgr.CachePath = filepath.Join(pkgmgr.BasePath, ".cache")
	utils.MakeDirs(pkgmgr.CachePath)

	pkgmgr.ReadRepoMeta()
	return nil
}

func (pkgmgr *PackageManager) HasRepoMeta() bool {
	// Construct the full path to the repometa.json file
	filePath := filepath.Join(pkgmgr.CachePath, "repometa.json")
	fmt.Printf("checking %s\n", filePath)
	return utils.IsFile(filePath)
}

func (pkgmgr *PackageManager) ReadRepoMeta() error {

	// Construct the full path to the repometa.json file
	filePath := filepath.Join(pkgmgr.CachePath, "repometa.json")

	// Read the file
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the JSON data
	var repoMeta repository.RepoMeta
	if err := json.Unmarshal(fileData, &repoMeta); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// set ast the manager's RepoMeta ...
	rm := repository.RepoMetaFile{
		Filename: filePath,
		Date:     repoMeta.Date,
	}
	pkgmgr.RepoMeta = rm

	fmt.Printf("repometa: %s\n", repoMeta)
	pkgmgr.CollectionManifests = repoMeta.CollectionManifests
	pkgmgr.CollectionFiles = repoMeta.CollectionFiles

	return nil
}

func (pkgmgr *PackageManager) InstalCollectionFromPath(namespace string, name string, version string, fn string) error {

	cPath := filepath.Join(pkgmgr.BasePath, "collections", "ansible_collections")
	cPath, _ = utils.GetAbsPath(cPath)
	utils.MakeDirs(cPath)

	// Basepath / collections / ansible_collections / namespace / name / ...
	dirPath := filepath.Join(pkgmgr.BasePath, "collections", "ansible_collections", namespace, name)
	dirPath, _ = utils.GetAbsPath(dirPath)
	fmt.Printf("\t%s\n", dirPath)
	utils.MakeDirs(dirPath)
	utils.ExtractTarGz(fn, dirPath)

	// Basepath / collections / ansible_collections / <namespace>.<name>-<version>.info / GALAXY.yml
	ymlDirName := fmt.Sprintf("%s.%s-%s.info", namespace, name, version)
	ymlDirPath := filepath.Join(pkgmgr.BasePath, "collections", "ansible_collections", ymlDirName)
	ymlDirPath, _ = utils.GetAbsPath(ymlDirPath)
	ymlFileName := filepath.Join(ymlDirPath, "GALAXY.tml")
	//fmt.Printf("MAKEDIR %s\n", ymlDirPath)
	utils.MakeDirs(ymlDirPath)

	galaxyYAML := GalaxyYamlMeta{
		Namespace:     namespace,
		Name:          name,
		Version:       version,
		FormatVersion: "1.0.0",
	}

	// Marshal the struct to JSON
	jsonData, err := json.MarshalIndent(galaxyYAML, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return err
	}

	// Write the JSON data to a file
	file, err := os.Create(ymlFileName)
	if err != nil {
		fmt.Printf("Error creating yml file: %v\n", err)
		return err
	}
	defer file.Close()

	if _, err := file.Write(jsonData); err != nil {
		fmt.Printf("Error writing to yml file: %v\n", err)
		return err
	}

	return nil
}

func (pkgmgr *PackageManager) InstallRoleFromPath(namespace string, name string, version string, fn string) error {
	/*
	# roles/geerlingguy.docker/meta/.galaxy_install_info
	1 install_date: 'Thu 13 Jun 2024 02:23:22 PM '                                                                                                                                                                                                                                                                                                               
  	2 version: 7.2.0
	*/

	rPath := filepath.Join(pkgmgr.BasePath, "roles")
	rPath, _ = utils.GetAbsPath(rPath)
	utils.MakeDirs(rPath)

	// Basepath / collections / ansible_collections / namespace / name / ...
	dirPath := filepath.Join(rPath, namespace + "." + name)
	fmt.Printf("\t%s\n", dirPath)
	utils.MakeDirs(dirPath)
	fmt.Printf("extracting %s to %s\n", fn, dirPath)
	err := utils.ExtractRoleTarGz(fn, dirPath)
	if err != nil {
		fmt.Printf("%s\n", err)
		panic("")
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("Mon 02 Jan 2006 03:04:05 PM ")
	infoYAML := RoleInstallInfo{
		InstallDate: formattedTime,
		Version: version,
	}
	yamlData, _ := yaml.Marshal(infoYAML)
	ymlFileName := filepath.Join(dirPath, "meta", ".galaxy_install_info")

	// Write the yaml data to a file
	file, err := os.Create(ymlFileName)
	if err != nil {
		fmt.Printf("Error creating yml file: %v\n", err)
		return err
	}
	defer file.Close()

	if _, err := file.Write(yamlData); err != nil {
		fmt.Printf("Error writing to yml file: %v\n", err)
		return err
	}
	
	return nil
}

func GetPackageManager(cachepath string, basepath string) (PackageManager, error) {

	fmt.Printf("GETPKGMGR: cache:%s basepath:%s\n", cachepath, basepath)

	pkgmgr := PackageManager{
		BasePath:  basepath,
		CachePath: cachepath,
	}
	pkgmgr.BasePath = basepath
	pkgmgr.CachePath = cachepath
	fmt.Printf("created pkgmgr: %s\n", pkgmgr)
	err := pkgmgr.Initialize()
	if err != nil {
		fmt.Printf("%s\n", err)
		return pkgmgr, err
	}

	return pkgmgr, nil
}
