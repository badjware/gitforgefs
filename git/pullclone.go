package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type GitClonerPuller interface {
	CloneOrPull(url string, pid int, defaultBranch string) (localRepoLoc string, err error)
}

type gitClonePullParam struct {
	url           string
	defaultBranch string
	dst           string
}

func (c *gitClient) getLocalRepoLoc(pid int) string {
	return filepath.Join(c.CloneLocation, c.RemoteURL.Hostname(), strconv.Itoa(pid))
}

func (c *gitClient) CloneOrPull(url string, pid int, defaultBranch string) (localRepoLoc string, err error) {
	localRepoLoc = c.getLocalRepoLoc(pid)
	select {
	case c.clonePullChan <- &gitClonePullParam{
		url:           url,
		defaultBranch: defaultBranch,
		dst:           localRepoLoc,
	}:
	default:
		return localRepoLoc, errors.New("failed to clone/pull local repo")
	}
	return localRepoLoc, nil
}

func (c *gitClient) clonePullWorker() {
	fmt.Println("Started git cloner/puller worker routine")

	for gpp := range c.clonePullChan {
		if _, err := os.Stat(gpp.dst); os.IsNotExist(err) {
			if err := c.clone(gpp); err != nil {
				fmt.Println(err)
			}
		} else if c.AutoPull {
			if err := c.pull(gpp); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (c *gitClient) clone(gpp *gitClonePullParam) error {
	branchRef := plumbing.NewBranchReferenceName(gpp.defaultBranch)

	if c.Fetch {
		// Clone the repo
		// TODO: figure out why this operation is so memory intensive...
		fmt.Printf("Cloning %v into %v\n", gpp.url, gpp.dst)
		fs := osfs.New(gpp.dst)
		storer := filesystem.NewStorage(fs, cache.NewObjectLRU(0))
		_, err := git.Clone(storer, fs, &git.CloneOptions{
			URL:           gpp.url,
			RemoteName:    c.RemoteName,
			ReferenceName: branchRef,
			NoCheckout:    !c.Checkout,
			Depth:         c.PullDepth,
		})
		if err != nil {
			return fmt.Errorf("failed to clone git repo %v to %v: %v", gpp.url, gpp.dst, err)
		}
	} else {
		// "Fake" cloning the repo by never actually talking to the git server
		// This skip a fetch operation that we would do if we where to do a proper clone
		// We can save a lot of time and network i/o doing it this way, at the cost of
		// resulting in a very barebone local copy
		fmt.Printf("Initializing %v into %v\n", gpp.url, gpp.dst)
		r, err := git.PlainInit(gpp.dst, false)
		if err != nil {
			return fmt.Errorf("failed to clone git repo %v to %v: %v", gpp.url, gpp.dst, err)
		}

		// Configure the remote
		_, err = r.CreateRemote(&config.RemoteConfig{
			Name: c.RemoteName,
			URLs: []string{gpp.url},
		})
		if err != nil {
			return fmt.Errorf("failed to setup remote %v in git repo %v: %v", gpp.url, gpp.dst, err)
		}

		// Configure a local branch to track the remote branch
		err = r.CreateBranch(&config.Branch{
			Name:   gpp.defaultBranch,
			Remote: c.RemoteName,
			Merge:  plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", gpp.defaultBranch)),
		})
		if err != nil {
			return fmt.Errorf("failed to create branch %v of git repo %v: %v", gpp.defaultBranch, gpp.dst, err)
		}

		// Checkout the default branch
		w, err := r.Worktree()
		if err != nil {
			return fmt.Errorf("failed to retrieve worktree of git repo %v: %v", gpp.dst, err)
		}
		w.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
		})
	}
	return nil
}

func (c *gitClient) pull(gpp *gitClonePullParam) error {
	// Check if the local repo is on default branch
	// Check if the local repo is dirty
	// Checkout the remote default branch
	// TODO
	return nil
}
