package git

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

type GitClonerPuller interface {
	CloneOrPull(url string, pid int, defaultBranch string) (localRepoLoc string, err error)
}

type GitClientParam struct {
	CloneLocation  string
	RemoteName     string
	RemoteURL      *url.URL
	PullAfterClone bool
	Clone          bool
	Checkout       bool
	PullDepth      int
	AutoPull       bool

	CloneBuffSize    int
	CloneWorkerCount int
	PullBuffSize     int
	PullWorkerCount  int
}

type gitClient struct {
	GitClientParam
	cloneChan chan *gitCloneParam
	pullChan  chan *gitPullParam
}

func NewClient(p GitClientParam) (*gitClient, error) {
	// Create the client
	c := &gitClient{
		GitClientParam: p,
		cloneChan:      make(chan *gitCloneParam, p.CloneBuffSize),
		pullChan:       make(chan *gitPullParam, p.PullBuffSize),
	}

	// Start worker goroutines
	for i := 0; i < p.CloneWorkerCount; i++ {
		go c.cloneWorker()
	}
	for i := 0; i < p.PullWorkerCount; i++ {
		go c.pullWorker()
	}

	return c, nil
}

func (c *gitClient) getLocalRepoLoc(pid int) string {
	return filepath.Join(c.CloneLocation, c.RemoteURL.Hostname(), strconv.Itoa(pid))
}

func (c *gitClient) CloneOrPull(url string, pid int, defaultBranch string) (localRepoLoc string, err error) {
	localRepoLoc = c.getLocalRepoLoc(pid)
	// TODO: Better manage concurrency, filter out duplicate requests
	if _, err := os.Stat(localRepoLoc); os.IsNotExist(err) {
		// Dispatch to clone worker
		select {
		case c.cloneChan <- &gitCloneParam{
			url:           url,
			defaultBranch: defaultBranch,
			dst:           localRepoLoc,
		}:
		default:
			return localRepoLoc, errors.New("failed to clone local repo")
		}
	} else if c.AutoPull {
		// Dispatch to pull worker
		select {
		case c.pullChan <- &gitPullParam{
			repoPath:      localRepoLoc,
			defaultBranch: defaultBranch,
		}:
		default:
			return localRepoLoc, errors.New("failed to pull local repo")
		}
	}
	return localRepoLoc, nil
}
