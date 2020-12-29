package git

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
)

type GitClientParam struct {
	CloneLocation string
	RemoteName    string
	RemoteURL     *url.URL
	Fetch         bool
	Checkout      bool
	SingleBranch  bool
	PullDepth     int
	AutoClone     bool
	AutoPull      bool

	ChanBuffSize    int
	ChanWorkerCount int
}

type gitClient struct {
	GitClientParam
	clonePullChan chan *gitClonePullParam
}

func NewClient(p GitClientParam) (*gitClient, error) {
	// Some validations
	if p.RemoteURL == nil {
		return nil, errors.New("required param RemoteURL is nil")
	}

	// Setup defaults
	if p.CloneLocation == "" {
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			dataHome = filepath.Join(os.Getenv("HOME"), ".local/share")
		}
		p.CloneLocation = filepath.Join(dataHome, "gitlabfs")
	}
	if p.RemoteName == "" {
		p.RemoteName = "origin"
	}
	if p.ChanBuffSize == 0 {
		p.ChanBuffSize = 500
	}
	if p.ChanWorkerCount == 0 {
		p.ChanWorkerCount = 5
	}

	// Create the client
	c := &gitClient{
		GitClientParam: p,
		clonePullChan:  make(chan *gitClonePullParam, p.ChanBuffSize),
	}

	// Start worker goroutines
	for i := 0; i < p.ChanWorkerCount; i++ {
		go c.clonePullWorker()
	}

	return c, nil
}
