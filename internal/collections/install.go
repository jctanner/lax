package collections

import (
	"fmt"
	"time"

	"github.com/jctanner/lax/internal/packagemanager"
	"github.com/jctanner/lax/internal/repository"
	"github.com/jctanner/lax/internal/types"
	"github.com/jctanner/lax/internal/utils"
)

// func Install(dest string, cachedir string, server string, requirements_file string, namespace string, name string, version string, args []string) error {
func Install(kwargs *types.CmdKwargs, args []string) error {

	dest := kwargs.DestDir
	cachedir := kwargs.CacheDir
	server := kwargs.Server
	requirements_file := kwargs.RequirementsFile
	if requirements_file != "" {
		panic("collection install with requirements file not yet implemented!!!")
	}
	namespace := kwargs.Namespace
	name := kwargs.Name
	version := kwargs.Version

	fmt.Printf("INSTALL2: cachedir:%s dest:%s\n", cachedir, dest)

	// does dest have a repodata.json file, read it in?
	repoClient, _ := repository.GetRepoClient(server, cachedir)
	fmt.Printf("repoclient: %s\n", repoClient)
	if repoClient == nil {
		return fmt.Errorf("no suitable repostiory found")
	}

	// Make the local package manager client
	pkgMgr, err := packagemanager.GetPackageManager(cachedir, dest)
	fmt.Printf("packagemanager: %s\n", pkgMgr)
	if err != nil {
		return err
	}

	// Is the package manager's meta older? Re-download if so ...
	if !pkgMgr.HasRepoMeta() {
		fmt.Printf("need to fetch repo meta ...\n")
		repoClient.FetchRepoMeta(pkgMgr.CachePath)
	} else {
		// Is it up to date?
		pDate := pkgMgr.RepoMeta.Date
		fmt.Printf("pdate: %s\n", pDate)
		rDate, _ := repoClient.GetRepoMetaDate()
		fmt.Printf("rdate: %s\n", rDate)

		d1, _ := time.Parse(time.RFC3339, pDate)
		d2, _ := time.Parse(time.RFC3339, rDate)

		if d1.Before(d2) {
			fmt.Printf("need to update local cache\n")
			repoClient.FetchRepoMeta(pkgMgr.CachePath)
		} else {
			fmt.Printf("do not need to update local cache\n")
		}

	}

	if len(args) > 0 {
		fqn := args[0]
		spec := utils.SplitSpec(fqn)
		//fmt.Printf("spec split .. %s\n", spec)

		if len(spec) == 3 {
			//server = spec[0]
			namespace = spec[1]
			name = spec[2]
		} else if len(spec) == 2 {
			namespace = spec[0]
			name = spec[1]
		}

	}

	ispec := utils.InstallSpec{
		Namespace: namespace,
		Name:      name,
		Version:   version,
	}

	fmt.Printf("spec: %s\n", ispec)

	specs, err := repoClient.ResolveCollectionDeps(ispec)

	if err != nil {
		fmt.Printf("error solving dep tree %s\n", err)
		return err
	}

	fmt.Printf("-----------------------------\n")
	for _, spec := range specs {
		fmt.Printf("install: %s.%s==%s\n", spec.Namespace, spec.Name, spec.Version)
	}

	fmt.Printf("-----------------------------\n")
	for _, spec := range specs {
		fmt.Printf("installing: %s.%s==%s\n", spec.Namespace, spec.Name, spec.Version)

		// get a local cache file from the repo to install ...
		fn := repoClient.GetCacheFileLocationForInstallSpec(spec)
		fmt.Printf("\tfrom %s\n", fn)

		// extract it to the right place ...
		pkgMgr.InstalCollectionFromPath(spec.Namespace, spec.Name, spec.Version, fn)

	}

	return nil
}
