package utils

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5" // with go modules disabled
	"github.com/go-git/go-git/v5/plumbing"
)

// CloneRepo clones a Git repository and ensures it's fully cloned
func CloneRepo(url, path string) error {
	// Ensure the target directory exists
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Clone the repository
	repo, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	// Verify the repository has at least one commit
	headRef, err := repo.Head()
	if err != nil || headRef.Hash() == plumbing.ZeroHash {
		return fmt.Errorf("cloned repository is empty or invalid")
	}

	return nil
}

// listTags lists the tags in the given repository path
func ListTags(path string) ([]string, error) {
	// Open the repository
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %v", err)
	}

	// Get the tag references
	tagRefs, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %v", err)
	}

	// Iterate over the tags and collect their names
	var tags []string
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tags = append(tags, ref.Name().Short())
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %v", err)
	}

	return tags, nil
}

func GetCommitDate(repoPath string, commitHash string) (string, error) {
	// Open the repository at the specified path
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %v", err)
	}

	// Get the commit object using the commit hash
	hash := plumbing.NewHash(commitHash)
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return "", fmt.Errorf("failed to get commit object: %v", err)
	}

	// Get the commit date
	commitDate := commit.Committer.When
	//return commitDate.Format(time.RFC3339), nil
	return commitDate.Format("20060102150405"), nil
}

func GetLatestCommitHash(repoPath string) (string, error) {
	// Open the repository at the specified path
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %v", err)
	}

	// Get the HEAD reference
	headRef, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %v", err)
	}

	// Get the latest commit hash
	commitHash := headRef.Hash().String()
	return commitHash, nil
}

func CheckoutBranch(path, branchName string) error {
	// Open the repository at the specified path
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %v", err)
	}

	// Get the working tree of the repository
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get working tree: %v", err)
	}

	// Perform the checkout
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %v", err)
	}

	return nil
}

func CheckoutTag(path, tagName string) error {
	// Open the repository at the specified path
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %v", err)
	}

	// Get the tag reference
	tagRef, err := repo.Tag(tagName)
	if err != nil {
		return fmt.Errorf("failed to get tag reference: %v", err)
	}

	// Get the commit hash for the tag
	tagCommit, err := repo.CommitObject(tagRef.Hash())
	if err != nil {
		return fmt.Errorf("failed to get commit for tag: %v", err)
	}

	// Get the working tree of the repository
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get working tree: %v", err)
	}

	// Perform the checkout to the commit hash
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: tagCommit.Hash,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout tag: %v", err)
	}

	return nil
}

func IsValidGitRepo(repoPath string) (bool, error) {
	// Check if the directory exists
	info, err := os.Stat(repoPath)
	if os.IsNotExist(err) {
		return false, fmt.Errorf("directory does not exist")
	}
	if !info.IsDir() {
		return false, fmt.Errorf("path is not a directory")
	}

	// Try to open the repository
	_, err = git.PlainOpen(repoPath)
	if err != nil {
		return false, nil // Not a git repository
	}

	return true, nil // Valid git repository
}
