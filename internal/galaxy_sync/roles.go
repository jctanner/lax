package galaxy_sync

import (
	"fmt"
	"lax/internal/repository"
	"lax/internal/utils"
	"log"
	"os"
	"path/filepath"
)

func syncRoles(apiClient CachedGalaxyClient, namespace string, name string) ([]Role, error) {

	// iterate roles ...
	roles, err := apiClient.GetRoles(namespace, name)
	if err != nil {
		log.Fatalf("Error fetching roles: %v", err)
	}

	/*
	for _, role := range roles {
		fmt.Printf("Role ID: %d, GUser:%s, GRepo: %s, Name: %s\n", role.ID, role.GithubUser, role.GithubRepo, role.Name)
	}
	*/

	return roles, nil
}


func GetRoleVersionArtifact(role Role, version RoleVersion, destdir string) (string, error) {
	// is there a release tarball ... ?
	// https://github.com/0ccupi3R/ansible-kibana/archive/refs/tags/7.6.1.tar.gz

	tarName := fmt.Sprintf("%s-%s-%s.tar.gz", role.SummaryFields.Namespace.Name, role.Name, version.Name)
	tarFilePath := filepath.Join(destdir, tarName)
	if utils.IsFile(tarFilePath) {
		return tarFilePath, nil
	}
	//fmt.Printf("NEEED TO FETCH %s\n", tarFilePath)
	//panic("check that tar!")

	tarUrl := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", role.GithubUser, role.GithubRepo, version.Name)
	//fmt.Printf("\t%s -> %s\n", baseUrl, tarFilePath)
	fmt.Printf("\tHEAD %s\n", tarUrl)
	if !utils.IsURLGood(tarUrl) {
		return "", fmt.Errorf("url failed http.HEAD check")
	}

	utils.DownloadBinaryFileToPath(tarUrl, tarFilePath)

	if !utils.IsFile(tarFilePath) {
		fmt.Printf("%s DID NOT ACTUALLY DOWNLOAD!!!\n", tarFilePath)
		panic("")
	}

	newNamespace := role.SummaryFields.Namespace.Name
	newName := role.Name
	needsRename := false

	meta, err := repository.GetRoleMetaFromTarball(tarFilePath)
	fmt.Printf("meta: %s\n", meta)
	if err != nil {
		fmt.Printf("%s\n", err)
		panic("")
	}

	/*
	if meta.GalaxyInfo.Namespace != "" && meta.GalaxyInfo.Namespace != role.GithubUser {
		fmt.Printf("role meta namespace DOES NOT MATCH github_user!!!\n")
		fmt.Printf("\t '%s' != '%s' \n", meta.GalaxyInfo.Namespace, role.GithubUser)
		panic("")
	}

	if meta.GalaxyInfo.RoleName != role.Name {
		fmt.Printf("role meta.name DOES NOT MATCH role.name!!!\n")
		fmt.Printf("\t'%s' != '%s' \n", meta.GalaxyInfo.RoleName, role.Name)
		panic("")
	}
	*/

	if meta.GalaxyInfo.Namespace != "" && meta.GalaxyInfo.Namespace != role.SummaryFields.Namespace.Name {
		needsRename = true
		newNamespace = meta.GalaxyInfo.Namespace
	}

	if meta.GalaxyInfo.RoleName != "" && meta.GalaxyInfo.RoleName != role.Name {
		needsRename = true
		newName = meta.GalaxyInfo.RoleName
	}

	if !needsRename {
		return "", nil
	}

	// construct the new filename
	newFn := fmt.Sprintf("%s-%s-%s.tar.gz", newNamespace, newName, version.Name)
	newFp := filepath.Join(destdir, newFn)

	fmt.Printf("rename %s -> %s\n", tarFilePath, newFp)
	fmt.Printf("symlink %s -> %s", tarFilePath, newFp)
	err = os.Rename(tarFilePath, newFp)
	if err != nil {
		fmt.Printf("%s\n", err)
		panic("")
	}

	return "", nil
}