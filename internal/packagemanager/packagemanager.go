package packagemanager

import (
    "fmt"
    "lax/internal/repository"
    "lax/internal/utils"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
)

type PackageManager struct {
    BasePath string
	CachePath string
    RepoMeta repository.RepoMetaFile
    CollectionManifests repository.RepoMetaFile
    CollectionFiles repository.RepoMetaFile
}

func (pkgmgr *PackageManager) Initialize() error {

	// where everything goes
	pkgmgr.BasePath, _ = utils.GetAbsPath(pkgmgr.BasePath)
	utils.MakeDirs(pkgmgr.BasePath)

	// store meta here
    pkgmgr.CachePath = filepath.Join(pkgmgr.BasePath, ".cache")
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
        Date: repoMeta.Date,
    }
    pkgmgr.RepoMeta = rm

    fmt.Printf("repometa: %s\n", repoMeta)
    pkgmgr.CollectionManifests = repoMeta.CollectionManifests
    pkgmgr.CollectionFiles = repoMeta.CollectionFiles

    return nil
}

func GetPackageManager(basepath string) (PackageManager, error) {

    pkgmgr := PackageManager{
        BasePath: basepath,
    }
    err := pkgmgr.Initialize()
    if err != nil {
        fmt.Errorf("%s\n", err)
        return pkgmgr, err
    }

    return pkgmgr, nil
}
