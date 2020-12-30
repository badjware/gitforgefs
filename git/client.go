package git

import (
	"net/url"
)

type GitClientParam struct {
	CloneLocation string
	RemoteName    string
	RemoteURL     *url.URL
	Fetch         bool
	Checkout      bool
	PullDepth     int
	AutoPull      bool

	ChanBuffSize    int
	ChanWorkerCount int
}

type gitClient struct {
	GitClientParam
	clonePullChan chan *gitClonePullParam
}

func NewClient(p GitClientParam) (*gitClient, error) {
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
