package galaxy_sync

import (
	"fmt"
	"lax/internal/utils"
	"path"
	"sync"
)

func GalaxySync(server string, dest string, collections_only bool, roles_only bool) error {

	fmt.Printf("syncing %s to %s\n", server, dest)
	
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

	rolesDir := path.Join(dest, "roles")
	utils.MakeDirs(rolesDir)

	// make the api client
	apiClient := CachedGalaxyClient{
		baseUrl: server,
		cachePath: cacheDir,
	}

	if roles_only || !collections_only {
		roles, err := syncRoles(server, dest, apiClient)
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

		maxConcurrent := 4

		var wg sync.WaitGroup
		sem := make(chan struct{}, maxConcurrent) // semaphore to limit concurrency
	
		for ix, role := range roles {
			wg.Add(1)
			go func(ix int, role Role) {
				defer wg.Done()
	
				fmt.Printf("%d: %s\n", ix, role)
				if len(role.SummaryFields.Versions) == 0 {
					return
				}
	
				for _, roleVersion := range role.SummaryFields.Versions {
					sem <- struct{}{} // acquire a slot
					go func(role Role, roleVersion RoleVersion) {
						defer func() { <-sem }() // release the slot
	
						fn, _ := GetRoleVersionArtifact(role, roleVersion, rolesDir)
						fmt.Printf("\t\t%s\n", fn)
					}(role, roleVersion)
				}
			}(ix, role)
		}
	
		wg.Wait()
		close(sem) // close the semaphore channel


	}

	

	/*
	if collections_only || !roles_only {
		err := syncCollections(server, dest, apiClient)
		if err != nil {
			return err
		}
	}
	*/

	return nil
}
