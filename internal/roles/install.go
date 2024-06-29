package roles

import (
	"fmt"
	"time"

	"github.com/jctanner/lax/internal/packagemanager"
	"github.com/jctanner/lax/internal/repository"
	"github.com/jctanner/lax/internal/types"
	"github.com/jctanner/lax/internal/utils"
	"github.com/sirupsen/logrus"
)

// func Install(dest string, cachedir string, server string, requirements_file string, namespace string, name string, version string, args []string) error {
func Install(kwargs *types.CmdKwargs, args []string) error {

	dest := kwargs.DestDir
	cachedir := kwargs.CacheDir
	server := kwargs.Server
	requirements_file := kwargs.RequirementsFile
	if requirements_file != "" {
		panic("role install with requirements file not yet implemented!!!")
	}
	namespace := kwargs.Namespace
	name := kwargs.Name
	version := kwargs.Version

	logrus.Debugf("ROLE INSTALL COMMAND: cachedir:%s dest:%s server:%s namespace:%s name:%s version:%s\n",
		cachedir, dest, server, namespace, name, version)

	// does dest have a repodata.json file, read it in?
	repoClient, _ := repository.GetRepoClient(server, cachedir)
	logrus.Debugf("created repo client: %s", repoClient)
	if repoClient == nil {
		return fmt.Errorf("no suitable repostiory found")
	}

	// Make the local package manager client
	pkgMgr, err := packagemanager.GetPackageManager(cachedir, dest)
	logrus.Debugf("created package manager: %s", pkgMgr)
	if err != nil {
		return err
	}

	// Is the package manager's meta older? Re-download if so ...
	if !pkgMgr.HasRepoMeta() {
		logrus.Debugf("no repo meta found on disk, fetching ...")
		repoClient.FetchRepoMeta(pkgMgr.CachePath)
	} else {
		// Is it up to date?
		pDate := pkgMgr.RepoMeta.Date
		logrus.Debugf("package manager meta date: %s", pDate)
		rDate, _ := repoClient.GetRepoMetaDate()
		logrus.Debugf("repo client meta date: %s\n", rDate)

		d1, _ := time.Parse(time.RFC3339, pDate)
		d2, _ := time.Parse(time.RFC3339, rDate)

		if d1.Before(d2) {
			logrus.Debugf("updating local meta cache")
			repoClient.FetchRepoMeta(pkgMgr.CachePath)
		} else {
			logrus.Debugf("not updating local meta cache")
		}

	}

	// split the last argument into namespace/name/etc
	if len(args) > 0 {
		fqn := args[0]
		spec := utils.SplitSpec(fqn)

		if len(spec) == 3 {
			//server = spec[0]
			namespace = spec[1]
			name = spec[2]
		} else if len(spec) == 2 {
			namespace = spec[0]
			name = spec[1]
		}

	}

	// define the top level install spec
	ispec := utils.InstallSpec{
		Namespace: namespace,
		Name:      name,
		Version:   version,
	}

	logrus.Infof("initial spec: %s.%s==%s", ispec.Namespace, ispec.Name, ispec.Version)

	specs, err := repoClient.ResolveRoleDeps(ispec)

	if err != nil {
		logrus.Errorf("error solving dep tree %s", err)
		return err
	}

	logrus.Infof("-----------------------------------------------------")
	for _, spec := range specs {
		logrus.Infof("install: %s.%s==%s\n", spec.Namespace, spec.Name, spec.Version)
	}

	logrus.Infof("-----------------------------------------------------")
	for _, spec := range specs {
		logrus.Infof("installing: %s.%s==%s\n", spec.Namespace, spec.Name, spec.Version)

		// get a local cache file from the repo to install ...
		fn := repoClient.GetCacheRoleFileLocationForInstallSpec(spec)
		logrus.Debugf("install %s from %s", spec, fn)

		// extract it to the right place ...
		pkgMgr.InstallRoleFromPath(spec.Namespace, spec.Name, spec.Version, fn)

	}

	return nil
}
