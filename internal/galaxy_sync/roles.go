package galaxy_sync

import (
	"fmt"
	"lax/internal/repository"
	"lax/internal/utils"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"
)

func syncRoles(apiClient CachedGalaxyClient, namespace string, name string, latest_only bool) ([]Role, error) {

	// iterate roles ...
	roles, err := apiClient.GetRoles(namespace, name, latest_only)
	if err != nil {
		log.Fatalf("Error fetching roles: %v", err)
	}

	// Sort the list by GithubUser and then by GithubRepo
	sort.Slice(roles, func(i, j int) bool {
		if roles[i].GithubUser == roles[j].GithubUser {
			return roles[i].GithubRepo < roles[j].GithubRepo
		}
		return roles[i].GithubUser < roles[j].GithubUser
	})

	return roles, nil
}

func GetRoleVersionArtifact(role Role, version RoleVersion, destdir string) (string, error) {
	// is there a release tarball ... ?
	// https://github.com/0ccupi3R/ansible-kibana/archive/refs/tags/7.6.1.tar.gz

	tarName := fmt.Sprintf("%s-%s-%s.tar.gz", role.SummaryFields.Namespace.Name, role.Name, version.Name)
	tarFilePath := filepath.Join(destdir, tarName)
	if utils.IsFile(tarFilePath) || utils.IsLink(tarFilePath) {
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
	fmt.Printf("symlink %s -> %s\n", tarFilePath, newFp)
	err = os.Rename(tarFilePath, newFp)

	if err != nil {
		fmt.Printf("%s\n", err)
		panic("")
	}

	err = utils.CreateSymlink(newFp, tarFilePath)
	//err = utils.CreateSymlink(tarFilePath, newFp)

	if err != nil {
		fmt.Printf("%s\n", err)
		panic("")
	}

	//panic("")

	return "", nil
}

func MakeRoleVersionArtifact(role Role, rolesDir string, cacheDir string) (string, error) {

	// how can we make a sortable version from a commit?
	// how can we make a sortable version from a branch name?
	// how can we quickly determine if a tar already exists for the hash?
	// how can we quickly get the latest commit on a branch?

	// Can get a tarball of a specific commit like this ...
	// 	https://github.com/github/codeql/archive/aef66c462abe817e33aad91d97aa782a1e2ad2c7.zip or .tar.gz
	// Or of a specific branch like this ...
	//	https://github.com/github/codeql/archive/refs/heads/main.tar.gz
	// Otherwise get the specified branch ...

	// short circuit if the role has a commit and there's a relevant tarball
	globPattern := fmt.Sprintf("%s-%s-*-%s.tar.gz", role.SummaryFields.Namespace.Name, role.Name, role.Commit)
	fmt.Printf("looking for files matching %s\n", globPattern)
	matches, _ := utils.FindMatchingFiles(rolesDir, globPattern)
	fmt.Printf("%s\n", matches)
	if len(matches) > 0 {
		return matches[0], nil
	}
	//panic("")

	// make a shallow clone ...
	repoUrl := fmt.Sprintf("https://github.com/%s/%s", role.GithubUser, role.GithubRepo)
	gitDir := path.Join(cacheDir, "git")
	utils.MakeDirs(gitDir)
	repoPath := path.Join(gitDir, fmt.Sprintf("%s.%s", role.GithubUser, role.GithubRepo))
	if !utils.IsDir(repoPath) {
		fmt.Printf("clone %s -> %s\n", repoUrl, repoPath)
		err := utils.CloneRepo(repoUrl, repoPath)
		if err != nil {
			fmt.Printf("failed to clone %s to %s ::%s\n", repoUrl, repoPath, err)
			//panic("")
			return "", err
		}
	}

	if role.GithubBranch != "" {
		utils.CheckoutBranch(repoPath, role.GithubBranch)
	}

	if role.Commit == "" {
		// get the latest commit hash
		role.Commit, _ = utils.GetLatestCommitHash(repoPath)
	}

	rawDate, _ := utils.GetCommitDate(repoPath, role.Commit)
	fmt.Printf("%s == %s\n", role.Commit, rawDate)
	date, _ := time.Parse("2006-01-02 15:04:05 -0700", rawDate)
	fmt.Printf("%s == %s\n", role.Commit, date)
	formattedDate := date.Format("20060102150405")
	fmt.Printf("%s == %s\n", role.Commit, formattedDate)

	//panic("")

	//currentTime := time.Now()
	version := "0.0.0+" + formattedDate + "-" + role.Commit

	dstFile := fmt.Sprintf("%s-%s-%s.tar.gz", role.SummaryFields.Namespace.Name, role.Name, version)
	dstFile = path.Join(rolesDir, dstFile)
	if utils.IsFile(dstFile) {
		fmt.Printf("%s exists\n", dstFile)
		return dstFile, nil
	}

	dstFile = fmt.Sprintf("%s-%s-%s.tar.gz", role.SummaryFields.Namespace.Name, role.Name, version)
	dstFile = path.Join(rolesDir, dstFile)
	tarUrl := fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", role.GithubUser, role.GithubRepo, role.Commit)
	fmt.Printf("%s -> %s\n", tarUrl, dstFile)
	_, err := utils.DownloadBinaryFileToPath(tarUrl, dstFile)
	if err != nil {
		fmt.Printf("%s\n", err)
		//panic("")
		return "", err
	}
	return dstFile, nil

	/*
		if role.Commit != "" {
			version = "0.0.0+" + role.Commit
			dstFile = fmt.Sprintf("%s-%s-%s.tar.gz", role.SummaryFields.Namespace.Name, role.Name, version)
			dstFile = path.Join(rolesDir, dstFile)
			tarUrl := fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", role.GithubUser, role.GithubRepo, role.Commit)
			fmt.Printf("%s -> %s\n", tarUrl, dstFile)
			_, err := utils.DownloadBinaryFileToPath(tarUrl, dstFile)
			if err != nil {
				fmt.Printf("%s\n", err)
				panic("")
			}
			return dstFile, nil

		} else if role.GithubBranch != "" {
			tarUrl := fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/%s.tar.gz", role.GithubUser, role.GithubRepo, role.GithubBranch)
			fmt.Printf("%s -> %s\n", tarUrl, dstFile)
			_, err := utils.DownloadBinaryFileToPath(tarUrl, dstFile)
			if err != nil {
				fmt.Printf("%s\n", err)
				panic("")
			}
			return dstFile, nil
		}

		fmt.Printf("role has no commit nor branch!!!")
		panic("")
	*/

	//return "", nil
}
