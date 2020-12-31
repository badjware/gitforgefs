package git

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

type gitCloneParam struct {
	url           string
	defaultBranch string
	dst           string
}

func (c *gitClient) cloneWorker() {
	fmt.Println("Started git cloner worker routine")

	for gcp := range c.cloneChan {
		if _, err := os.Stat(gcp.dst); os.IsNotExist(err) {
			if err := c.clone(gcp); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (c *gitClient) clone(gcp *gitCloneParam) error {
	branchRef := plumbing.NewBranchReferenceName(gcp.defaultBranch)

	if c.Clone {
		// Clone the repo
		// TODO: figure out why this operation is so memory intensive...
		fmt.Printf("Cloning %v into %v\n", gcp.url, gcp.dst)
		// fs := osfs.New(gcp.dst)
		// storer := filesystem.NewStorage(fs, cache.NewObjectLRU(0))
		_, err := git.PlainClone(gcp.dst, false, &git.CloneOptions{
			URL:           gcp.url,
			RemoteName:    c.RemoteName,
			ReferenceName: branchRef,
			NoCheckout:    !c.Checkout,
			Depth:         c.PullDepth,
		})
		if err != nil {
			return fmt.Errorf("failed to clone git repo %v to %v: %v", gcp.url, gcp.dst, err)
		}
	} else {
		// "Fake" cloning the repo by never actually talking to the git server
		// This skip a fetch operation that we would do if we where to do a proper clone
		// We can save a lot of time and network i/o doing it this way, at the cost of
		// resulting in a very barebone local copy
		fmt.Printf("Initializing %v into %v\n", gcp.url, gcp.dst)
		r, err := git.PlainInit(gcp.dst, false)
		if err != nil {
			return fmt.Errorf("failed to clone git repo %v to %v: %v", gcp.url, gcp.dst, err)
		}

		// Configure the remote
		_, err = r.CreateRemote(&config.RemoteConfig{
			Name: c.RemoteName,
			URLs: []string{gcp.url},
		})
		if err != nil {
			return fmt.Errorf("failed to setup remote %v in git repo %v: %v", gcp.url, gcp.dst, err)
		}

		// Configure a local branch to track the remote branch
		err = r.CreateBranch(&config.Branch{
			Name:   gcp.defaultBranch,
			Remote: c.RemoteName,
			Merge:  plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", gcp.defaultBranch)),
		})
		if err != nil {
			return fmt.Errorf("failed to create branch %v of git repo %v: %v", gcp.defaultBranch, gcp.dst, err)
		}

		// Checkout the default branch
		w, err := r.Worktree()
		if err != nil {
			return fmt.Errorf("failed to retrieve worktree of git repo %v: %v", gcp.dst, err)
		}
		w.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
		})
	}
	if c.PullAfterClone {
		// Dispatch to pull worker
		select {
		case c.pullChan <- &gitPullParam{
			repoPath:      gcp.dst,
			defaultBranch: gcp.defaultBranch,
		}:
		default:
			return errors.New("failed to pull local repo after clone")
		}
	}
	return nil
}
