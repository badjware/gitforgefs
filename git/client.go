package git

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/memqueue"
)

const (
	CloneInit  = iota
	CloneClone = iota
)

type GitClonerPuller interface {
	CloneOrPull(url string, pid int, defaultBranch string) (localRepoLoc string, err error)
}

type GitClientParam struct {
	CloneLocation string
	RemoteName    string
	RemoteURL     *url.URL
	CloneMethod   int
	PullDepth     int
	AutoPull      bool

	QueueSize        int
	QueueWorkerCount int
}

type gitClient struct {
	GitClientParam
	queue     taskq.Queue
	cloneTask *taskq.Task
	pullTask  *taskq.Task
}

func NewClient(p GitClientParam) (*gitClient, error) {
	queueFactory := memqueue.NewFactory()
	// Create the client
	c := &gitClient{
		GitClientParam: p,

		queue: queueFactory.RegisterQueue(&taskq.QueueOptions{
			Name:         "git-queue",
			MaxNumWorker: int32(p.QueueWorkerCount),
			BufferSize:   p.QueueSize,
			Storage:      taskq.NewLocalStorage(),
		}),
	}

	c.cloneTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name:       "git-clone",
		Handler:    c.clone,
		RetryLimit: 1,
	})
	c.pullTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name:       "git-pull",
		Handler:    c.pull,
		RetryLimit: 1,
	})

	return c, nil
}

func (c *gitClient) getLocalRepoLoc(pid int) string {
	return filepath.Join(c.CloneLocation, c.RemoteURL.Hostname(), strconv.Itoa(pid))
}

func (c *gitClient) CloneOrPull(url string, pid int, defaultBranch string) (localRepoLoc string, err error) {
	localRepoLoc = c.getLocalRepoLoc(pid)
	if _, err := os.Stat(localRepoLoc); os.IsNotExist(err) {
		// Dispatch clone msg
		msg := c.cloneTask.WithArgs(context.Background(), url, defaultBranch, localRepoLoc)
		msg.OnceInPeriod(time.Second, pid)
		c.queue.Add(msg)
	} else if c.AutoPull {
		// Dispatch pull msg
		msg := c.pullTask.WithArgs(context.Background(), localRepoLoc, defaultBranch)
		msg.OnceInPeriod(time.Second, pid)
		c.queue.Add(msg)
	}
	return localRepoLoc, nil
}
