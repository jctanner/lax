package galaxy_sync

import (
	"fmt"
	"lax/internal/utils"
	"log"
	"path/filepath"
)

func syncRoles(server string, dest string, apiClient CachedGalaxyClient) ([]Role, error) {

	// iterate roles ...
	roles, err := apiClient.GetRoles()
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

	tarUrl := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", role.GithubUser, role.GithubRepo, version.Name)
	//fmt.Printf("\t%s -> %s\n", baseUrl, tarFilePath)
	fmt.Printf("\tHEAD %s\n", tarUrl)
	if !utils.IsURLGood(tarUrl) {
		return "", fmt.Errorf("url failed http.HEAD check")
	}

	utils.DownloadBinaryFileToPath(tarUrl, tarFilePath)

	return "", nil
}