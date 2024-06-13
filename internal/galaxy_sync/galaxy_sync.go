package galaxy_sync

import (
	"fmt"
	"lax/internal/utils"
	"os"
	"path"
	"sync"
	"time"
)

func GalaxySync(server string, dest string, download_concurrency int, collections_only bool, roles_only bool, latest_only bool, namespace string, name string) error {

	fmt.Printf("syncing %s to %s collections:%s roles:%s latest:%s\n", server, dest, collections_only, roles_only, latest_only)
	
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
		baseUrl: server,
		cachePath: cacheDir,
	}

	if roles_only || !collections_only {
		roles, err := syncRoles(apiClient, namespace, name, latest_only)
		if err != nil {
			return err
		}
		fmt.Printf("%d total roles\n", len(roles))

		/*
		// get or make all the release tarballs ..
		for ix, role := range roles {
			fmt.Printf("%d: %s\n", ix, role)
			if len(role.SummaryFields.Versions) == 0 {
				continue
			}
			for _, roleVersion := range role.SummaryFields.Versions {
				//fmt.Printf("\t%s\n", roleVersion)
				fn, _ := GetRoleVersionArtifact(role, roleVersion, rolesDir)
				fmt.Printf("\t\t%s\n", fn)
			}
		}
			*/

		// store all the role data into a tar.gz file ...

		maxConcurrent := download_concurrency

		var wg sync.WaitGroup
		sem := make(chan struct{}, maxConcurrent) // semaphore to limit concurrency
	
		for ix, role := range roles {
			wg.Add(1)
			go func(ix int, role Role) {
				defer wg.Done()
	
				fmt.Printf("%d: %s\n", ix, role)
				//if len(role.SummaryFields.Versions) == 0 {
				//	return
				//}

				badFile := fmt.Sprintf("%s-%s.bad", role.GithubUser, role.GithubRepo)
				badFile = path.Join(rolesDir, badFile)
				if utils.IsFile(badFile) {
					fmt.Printf("found %s, skipping\n", badFile)
					return
				}
	
				if len(role.SummaryFields.Versions) > 0 {
					versions := role.SummaryFields.Versions
					if latest_only {
						versions,_ = reduceRoleVersionsToHighest(versions)
					}
					for _, roleVersion := range versions{
						sem <- struct{}{} // acquire a slot
						go func(role Role, roleVersion RoleVersion) {
							defer func() { <-sem }() // release the slot
		
							vBadFile := fmt.Sprintf("%s-%s-%s.bad", role.GithubUser, role.GithubRepo, roleVersion.Name)
							vBadFile = path.Join(rolesDir, vBadFile)
							if utils.IsFile(vBadFile) {
								fmt.Printf("found %s, skipping\n", vBadFile)
								return
							}

							time.Sleep(2 * time.Second)
							fn, err := GetRoleVersionArtifact(role, roleVersion, rolesDir)
							fmt.Printf("\t\t%s\n", fn)

							if err == nil {
								fmt.Printf("\t\t%s\n", fn)
							} else {
								// can we mark this as "BAD" somehow ... ?
								file, _ := os.Create(vBadFile)
								defer file.Close() // Ensure the file is closed
							}

						}(role, roleVersion)
					}
				} else {
					//fmt.Printf("NO VERSIONS!!!\n")
					time.Sleep(2 * time.Second)
					fmt.Printf("Enumerting virtual role version ...\n")
					fn, err := MakeRoleVersionArtifact(role, rolesDir, cacheDir)
					if err == nil {
						fmt.Printf("\t\t%s\n", fn)
					} else {
						// can we mark this as "BAD" somehow ... ?
						file, _ := os.Create(badFile)
						defer file.Close() // Ensure the file is closed
					}
				}
			}(ix, role)
		}
	
		wg.Wait()
		close(sem) // close the semaphore channel


	}
	
	if collections_only || !roles_only {
		collections, err := syncCollections(server, dest, apiClient, namespace, name, latest_only)
		if err != nil {
			return err
		}

		fmt.Printf("%d total collection versions\n", len(collections))

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
					fmt.Printf("downloading %s\n", col.DownloadUrl)
					utils.DownloadBinaryFileToPath(col.DownloadUrl, fp)
				}
	
			}(ix, cv)
		}
	
		wg.Wait()

	}

	return nil
}

