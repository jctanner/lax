package galaxy_sync

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/jctanner/lax/internal/types"
	"github.com/jctanner/lax/internal/utils"
	"github.com/sirupsen/logrus"
)

func GalaxySync(kwargs *types.CmdKwargs) error {

	server := kwargs.Server
	dest := kwargs.DestDir
	download_concurrency := kwargs.DownloadConcurrency
	collections_only := kwargs.CollectionsOnly
	roles_only := kwargs.RolesOnly
	latest_only := kwargs.LatestOnly
	namespace := kwargs.Namespace
	name := kwargs.Name
	requirements_file := kwargs.RequirementsFile

	logrus.Infof("syncing %s to %s collections:%t roles:%t latest:%t\n", server, dest, collections_only, roles_only, latest_only)

	// need to make sure the dest exists
	dest = utils.ExpandUser(dest)
	utils.MakeDirs(dest)

	// define the api cache dir
	cacheDir := path.Join(dest, ".cache")
	utils.MakeDirs(dest)

	rolesCacheDir := path.Join(cacheDir, "roles")
	utils.MakeDirs(rolesCacheDir)
	collectionsCacheDir := path.Join(cacheDir, "collections")
	utils.MakeDirs(collectionsCacheDir)

	collectionsDir := path.Join(dest, "collections")
	utils.MakeDirs(collectionsDir)

	rolesDir := path.Join(dest, "roles")
	utils.MakeDirs(rolesDir)

	// make the api client
	apiClient := CachedGalaxyClient{
		baseUrl:   server,
		cachePath: cacheDir,
	}

	var requirements *Requirements

	if requirements_file != "" {
		if !utils.IsFile(requirements_file) {
			return fmt.Errorf("ERROR: %s does not exist", requirements_file)
		}
		requirements_, _ := parseRequirements(requirements_file)
		pretty, _ := utils.PrettyPrint(requirements_)
		fmt.Println(pretty)
		//return nil
		requirements = requirements_
	}

	if roles_only || !collections_only {

		var roles []Role

		if requirements_file == "" {

			_roles, err := syncRoles(apiClient, namespace, name, latest_only)
			if err != nil {
				return err
			}
			roles = append(roles, _roles...)

		} else {

			for _, rrole := range requirements.Roles {
				parts := strings.Split(rrole.Name, ".")
				_namespace := parts[0]
				_name := parts[1]
				_roles, err := syncRoles(apiClient, _namespace, _name, latest_only)
				if err != nil {
					return err
				}
				roles = append(roles, _roles...)
			}
		}

		logrus.Infof("%d total roles\n", len(roles))

		maxConcurrent := download_concurrency

		var wg sync.WaitGroup
		sem := make(chan struct{}, maxConcurrent) // semaphore to limit concurrency

		for ix, role := range roles {
			wg.Add(1)
			go func(ix int, role Role) {
				defer wg.Done()

				badFile := path.Join(rolesDir, fmt.Sprintf("%s-%s.bad", role.GithubUser, role.GithubRepo))
				if utils.IsFile(badFile) {
					logrus.Debugf("found %s, skipping\n", badFile)
					return
				}

				if len(role.SummaryFields.Versions) > 0 {
					versions := role.SummaryFields.Versions
					if latest_only {
						versions, _ = reduceRoleVersionsToHighest(versions)
					}
					for _, roleVersion := range versions {
						sem <- struct{}{} // acquire a slot
						go func(role Role, roleVersion RoleVersion) {
							defer func() { <-sem }() // release the slot

							vBadFile := path.Join(rolesDir, fmt.Sprintf("%s-%s-%s.bad", role.GithubUser, role.GithubRepo, roleVersion.Name))
							vBadFile, _ = utils.GetAbsPath(vBadFile)
							logrus.Debugf("checking for %s\n", vBadFile)
							if utils.IsFile(vBadFile) {
								fmt.Printf("found %s, skipping\n", vBadFile)
								return
							}

							logrus.Infof("%s not found\n", vBadFile)
							logrus.Infof("sleeping 1s before GET ...")
							time.Sleep(1 * time.Second)
							logrus.Infof("GET %s %s\n", role, roleVersion)
							fn, err := GetRoleVersionArtifact(role, roleVersion, rolesDir)
							//logrus.Debugf("\t\tnew-file: %s\n", fn)

							if err != nil {
								// mark as "BAD"
								logrus.Errorf("\t\tmarking %s. %s as 'bad'\n", role, roleVersion)
								file, _ := os.Create(vBadFile)
								file.Write([]byte(fmt.Sprintf("%s\n", err)))
								defer file.Close()
							} else {
								//logrus.Debugf("\t\t%s\n", fn)
								logrus.Debugf("\t\tsaved: %s\n", fn)
							}

						}(role, roleVersion)
					}
				} else {
					time.Sleep(2 * time.Second)
					logrus.Debugf("Enumerating virtual role version ...\n")
					fn, err := MakeRoleVersionArtifact(role, rolesDir, cacheDir)
					if err != nil {
						logrus.Errorf("marking as bad due to %s\n", err)
						file, _ := os.Create(badFile)
						file.Write([]byte(fmt.Sprintf("%s\n", err)))
						defer file.Close()
						return
					}
					fmt.Printf("\t\t%s\n", fn)
				}
			}(ix, role)
		}

		wg.Wait()
		close(sem) // close the semaphore channel

	}

	if collections_only || !roles_only {

		var collections []CollectionVersionDetail

		if requirements_file == "" {

			_collections, err := syncCollections(server, dest, apiClient, namespace, name, latest_only)
			if err != nil {
				return err
			}
			collections = append(collections, _collections...)

		} else {

			for _, _col := range requirements.Collections {
				parts := strings.Split(_col.Name, ".")
				_namespace := parts[0]
				_name := parts[1]
				_cols, err := syncCollections(server, dest, apiClient, _namespace, _name, latest_only)
				if err != nil {
					return err
				}
				collections = append(collections, _cols...)
			}
		}

		logrus.Infof("%d total collection versions\n", len(collections))

		maxConcurrent := download_concurrency

		var wg sync.WaitGroup
		sem := make(chan struct{}, maxConcurrent) // semaphore to limit concurrency

		for ix, cv := range collections {
			sem <- struct{}{} // acquire a slot
			wg.Add(1)
			go func(ix int, col CollectionVersionDetail) {
				defer wg.Done()
				defer func() { <-sem }() // release the slot

				fn := path.Base(col.Artifact.FileName)
				fp := path.Join(collectionsDir, fn)
				if !utils.IsFile(fp) {
					logrus.Infof("call download of %s to %s\n", col.DownloadUrl, fp)
					utils.DownloadBinaryFileToPath(col.DownloadUrl, fp)
				}

			}(ix, cv)
		}

		wg.Wait()

	}

	return nil
}
