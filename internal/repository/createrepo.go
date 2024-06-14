package repository

import (
	"encoding/json"
	"fmt"
	"lax/internal/utils"
	"os"
	"path/filepath"
	"time"
)

func CreateRepo(dest string, roles_only bool, collectios_only bool) error {

	fmt.Printf("Create repo in %s\n", dest)

	// find the full path
	apath, err := utils.GetAbsPath(dest)
	if err != nil {
		return err
	}

	// assert it's a dir that exists
	if !utils.IsDir(apath) {
		fmt.Printf("%s is not a directory\n", apath)
		return nil
	}

	// assert it has a collections subdir
	collectionsPath := filepath.Join(apath, "collections")
	err = processCollections(apath, collectionsPath)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}

	// assert it has a collections subdir
	rolesPath := filepath.Join(apath, "roles")
	err = processRoles(apath, rolesPath)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}

	// write repodata.json
	currentTime := time.Now().UTC()
	isoFormattedCurrent := currentTime.Format(time.RFC3339)
	rMeta := RepoMeta{
		Date: isoFormattedCurrent,
		CollectionManifests: RepoMetaFile{
			Date:     isoFormattedCurrent,
			Filename: "collection_manifests.tar.gz",
		},
		CollectionFiles: RepoMetaFile{
			Date:     isoFormattedCurrent,
			Filename: "collection_files.tar.gz",
		},
		RoleManifests: RepoMetaFile{
			Date:     isoFormattedCurrent,
			Filename: "role_manifests.tar.gz",
		},
		RoleFiles: RepoMetaFile{
			Date:     isoFormattedCurrent,
			Filename: "role_files.tar.gz",
		},
	}

	// Marshal the RepoMeta instance to JSON
	jsonData, err := json.MarshalIndent(rMeta, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return err
	}

	// Write the JSON data to a file
	fn := filepath.Join(apath, "repometa.json")
	file, err := os.Create(fn)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	return nil
}

func processCollections(basePath string, collectionsPath string) error {

	if !utils.IsDir(collectionsPath) {
		fmt.Printf("%s is not a directory\n", collectionsPath)
		return nil
		//hasCollections = false
	}

	// make a list of tarballs
	collectionTarBalls, err := utils.ListTarGzFiles(collectionsPath)
	if err != nil {
		return err
	}

	// we need the metdata from each file
	metadataDir := filepath.Join(basePath, "metadata")
	err = utils.MakeDirs(metadataDir)
	if err != nil {
		return err
	}

	// store all collectionManifests
	collectionManifests := []CollectionManifest{}
	collectionFilesCache := []CollectionCachedFileInfo{}

	for _, file := range collectionTarBalls {
		fmt.Printf("%s\n", file)

		// get MANIFEST.json + FILES.json
		fmap, err := utils.ExtractJSONFilesFromTarGz(file, []string{"MANIFEST.json", "FILES.json"})
		if err != nil {
			fmt.Printf("ERROR extracting %s\n", err)
			continue
		}

		var manifest CollectionManifest
		var filesdata CollectionFilesMeta

		err = json.Unmarshal(fmap["MANIFEST.json"], &manifest)
		if err != nil {
			fmt.Printf("ERROR %s\n", err)
			continue
		}
		//fmt.Printf("%s\n", manifest)
		collectionManifests = append(collectionManifests, manifest)

		err = json.Unmarshal(fmap["FILES.json"], &filesdata)
		if err != nil {
			fmt.Printf("ERROR %s\n", err)
			continue
		}

		for _, f := range filesdata.Files {
			fDs := CollectionCachedFileInfo{
				Namespace:      manifest.CollectionInfo.Namespace,
				Name:           manifest.CollectionInfo.Name,
				Version:        manifest.CollectionInfo.Version,
				FileName:       f.Name,
				FileType:       f.FType,
				CheckSumSHA256: f.CheckSumSHA256,
			}

			collectionFilesCache = append(collectionFilesCache, fDs)

		}

	}

	// write manifests.tar.gz
	collectionManifestsFilePath := filepath.Join(basePath, "collection_manifests.tar.gz")
	fmt.Printf("write %s\n", collectionManifestsFilePath)
	createCollectionManifestsTarGz(collectionManifests, collectionManifestsFilePath)

	// write files.tar.gz
	fmt.Printf("total files %d\n", len(collectionFilesCache))
	collectionsCachedFilesPath := filepath.Join(basePath, "collection_files.tar.gz")
	saveCachedCollectionFilesToGzippedFile(collectionFilesCache, collectionsCachedFilesPath, 1000000)

	return nil
}

func processRoles(basePath string, rolesPath string) error {
	if !utils.IsDir(rolesPath) {
		fmt.Printf("%s is not a directory\n", rolesPath)
		return nil
		//hasCollections = false
	}

	// make a list of tarballs
	roleTarBalls, err := utils.ListTarGzFiles(rolesPath)
	if err != nil {
		return err
	}

	// we need the metdata from each file
	metadataDir := filepath.Join(basePath, "metadata")
	err = utils.MakeDirs(metadataDir)
	if err != nil {
		return err
	}

	// store all role manifests
	rolesMeta := []RoleMeta{}
	roleFilesCache := []RoleCachedFileInfo{}

	for _, f := range roleTarBalls {
		fmt.Printf("tar: %s\n", f)

		rmeta, _ := GetRoleMetaFromTarball(f)

		// -always- define the namespace even if the role
		// author did not.
		if rmeta.GalaxyInfo.Namespace == "" {
			rmeta.GalaxyInfo.Namespace = extractRoleNamespaceFromTarName(f)
		}

		if rmeta.GalaxyInfo.RoleName == "" {
			rmeta.GalaxyInfo.RoleName = extractRoleNameFromTarName(f)
		}

		if rmeta.GalaxyInfo.Version == "" {
			rmeta.GalaxyInfo.Version = extractRoleVersionFromTarName(f)
		}

		pretty, _ := utils.PrettyPrint(rmeta)
		fmt.Println(pretty)

		rolesMeta = append(rolesMeta, rmeta)

		tarFileNames, _ := utils.ListFilenamesInTarGz(f)
		for _, tfn := range tarFileNames {
			fmt.Printf("%s\n", tfn)
			relativePath := utils.RemoveFirstPathElement(tfn)
			fmt.Printf("\t%s\n", relativePath)

			cf := RoleCachedFileInfo{
				Namespace: rmeta.GalaxyInfo.Namespace,
				Name:      rmeta.GalaxyInfo.RoleName,
				Version:   rmeta.GalaxyInfo.Version,
				FileName:  tfn,
			}
			roleFilesCache = append(roleFilesCache, cf)
		}

	}

	// write manifests.tar.gz
	roleMetaFilePath := filepath.Join(basePath, "role_manifests.tar.gz")
	fmt.Printf("write %s\n", roleMetaFilePath)
	createRoleMetaTarGz(rolesMeta, roleMetaFilePath)

	// write files.tar.gz
	fmt.Printf("total files %d\n", len(roleFilesCache))
	roleCachedFilesPath := filepath.Join(basePath, "role_files.tar.gz")
	saveCachedRoleFilesToGzippedFile(roleFilesCache, roleCachedFilesPath, 1000000)

	return nil
}
