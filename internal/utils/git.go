package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5" // with go modules disabled
	"github.com/go-git/go-git/v5/plumbing"
)

// cloneRepo clones a Git repository to the specified path
func CloneRepo(url, path string) error {
	// Ensure the target directory exists
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Clone the repository
	_, err = git.PlainClone(path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
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

/*
func GetCommitDate(repoPath string, commitHash string) (string, error) {
	// Create the git command
	cmd := exec.Command("git", "-C", repoPath, "show", "-s", "--format=%ci", commitHash)

	// Run the command and capture the output
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute git command: %w", err)
	}

	// Trim any surrounding whitespace from the output
	date := strings.TrimSpace(string(output))
	return date, nil
}
*/

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
	return commitDate.Format(time.RFC3339), nil
}

/*
func GetLatestCommitHash(repoPath string) (string, error) {
	// Create the git command
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")

	// Run the command and capture the output
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute git command: %w", err)
	}

	// Trim any surrounding whitespace from the output
	commitHash := strings.TrimSpace(string(output))
	return commitHash, nil
}
*/

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

/*
func CheckoutBranch(path, branchName string) error {
	// Run the git checkout command as a subprocess with the working directory set to the repository path
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout branch: %v", err)
	}
	return nil
}
*/

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

/*
func CheckoutTag(path, tagName string) error {
	// Run the git checkout command as a subprocess with the working directory set to the repository path
	cmd := exec.Command("git", "checkout", tagName)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout tag: %v", err)
	}
	return nil
}
*/

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
