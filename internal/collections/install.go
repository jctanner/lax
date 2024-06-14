package collections

import (
	"fmt"
	"lax/internal/packagemanager"
	"lax/internal/repository"
	"lax/internal/utils"
	"time"
)

func Install(dest string, cachedir string, server string, requirements_file string, namespace string, name string, version string, args []string) error {

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
