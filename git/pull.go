package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type gitPullParam struct {
	repoPath      string
	defaultBranch string
}

func (c *gitClient) pullWorker() {
	fmt.Println("Started git puller worker routine")

	for gpp := range c.pullChan {
		if err := c.pull(gpp); err != nil {
			fmt.Println(err)
		}
	}
}

func (c *gitClient) pull(gpp *gitPullParam) error {
	r, err := git.PlainOpen(gpp.repoPath)
	if err != nil {
		return fmt.Errorf("failed to open git repo %v: %v", gpp.repoPath, err)
	}
	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to retrieve worktree of git repo %v: %v", gpp.repoPath, err)
	}

	// Check if the local repo is on default branch
	headRef, err := r.Head()
	if err != nil {
		// We ignore "reference not found" as this occurs when the local branch
		// has never been checked out when we are in init-pull mode
		if err.Error() != "reference not found" {
			return fmt.Errorf("failed to retrieve HEAD of git repo %v: %v", gpp.repoPath, err)
		}
	} else {
		branchRef := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", gpp.defaultBranch))
		if headRef.Name() != branchRef {
			// default branch is not checked out, nothing to do
			fmt.Printf("Repo %v is not on default branch (%v != %v), skipping pull\n", gpp.repoPath, branchRef, headRef.Name())
			return nil
		}
	}

	// Pull the remote
	// TODO: Just like clone, this is very memory intensive for some reasons...
	fmt.Printf("Pulling %v\n", gpp.repoPath)
	if err := w.Pull(&git.PullOptions{RemoteName: c.RemoteName, Depth: c.PullDepth}); err != nil {
		return fmt.Errorf("failed to pull git repo %v: %v", gpp.repoPath, err)
	}

	return nil
}
