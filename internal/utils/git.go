package utils

import (
    "fmt"
    "os"
    "os/exec"

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
func CheckoutTag(path, tagName string) error {
    // Open the repository
    repo, err := git.PlainOpen(path)
    if err != nil {
        return fmt.Errorf("failed to open repository: %v", err)
    }

    // Resolve the tag to a commit
    ref, err := repo.Tag(tagName)
    if err != nil {
        return fmt.Errorf("failed to resolve tag: %v", err)
    }

    // Get the commit object
    commit, err := repo.CommitObject(ref.Hash())
    if err != nil {
        return fmt.Errorf("failed to get commit object: %v", err)
    }

    // Create a worktree to checkout the commit
    worktree, err := repo.Worktree()
    if err != nil {
        return fmt.Errorf("failed to get worktree: %v", err)
    }

    // Checkout the commit
    err = worktree.Checkout(&git.CheckoutOptions{
        Hash: commit.Hash,
    })
    if err != nil {
        return fmt.Errorf("failed to checkout commit: %v", err)
    }

    return nil
}
*/

/*
func CheckoutTag(path, tagName string) error {
    // Open the repository
    repo, err := git.PlainOpen(path)
    if err != nil {
        return fmt.Errorf("failed to open repository: %v", err)
    }

    // Resolve the tag to a commit
    ref, err := repo.Tag(tagName)
    if err != nil {
        return fmt.Errorf("failed to resolve tag: %v", err)
    }

    // Dereference the tag if it's annotated
    tag, err := repo.TagObject(ref.Hash())
    if err == nil {
        ref = tag.Target
    }

    // Get the commit object
    commit, err := repo.CommitObject(ref.Hash())
    if err != nil {
        return fmt.Errorf("failed to get commit object: %v", err)
    }

    // Create a worktree to checkout the commit
    worktree, err := repo.Worktree()
    if err != nil {
        return fmt.Errorf("failed to get worktree: %v", err)
    }

    // Checkout the commit
    err = worktree.Checkout(&git.CheckoutOptions{
        Hash: commit.Hash,
    })
    if err != nil {
        return fmt.Errorf("failed to checkout commit: %v", err)
    }

    return nil
}
*/

func CheckoutTag(path, tagName string) error {
    // Run the git checkout command as a subprocess with the working directory set to the repository path
    cmd := exec.Command("git", "checkout", tagName)
    cmd.Dir = path
    //cmd.Stdout = os.Stdout
    //cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to checkout tag: %v", err)
    }

    return nil
}
