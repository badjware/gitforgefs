package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/badjware/gitlabfs/fstree"
	"github.com/badjware/gitlabfs/utils"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/memqueue"
)

type GitClientParam struct {
	CloneLocation    string `yaml:"clone_location,omitempty"`
	Remote           string `yaml:"remote,omitempty"`
	OnClone          string `yaml:"on_clone,omitempty"`
	AutoPull         bool   `yaml:"auto_pull,omitempty"`
	Depth            int    `yaml:"depth,omitempty"`
	QueueSize        int    `yaml:"queue_size,omitempty"`
	QueueWorkerCount int    `yaml:"worker_count,omitempty"`
}

type gitClient struct {
	GitClientParam

	hostnameProg *regexp.Regexp

	majorVersion int
	minorVersion int
	patchVersion string

	queue     taskq.Queue
	cloneTask *taskq.Task
	pullTask  *taskq.Task
}

func NewClient(p GitClientParam) (*gitClient, error) {
	queueFactory := memqueue.NewFactory()
	// Create the client
	c := &gitClient{
		GitClientParam: p,

		hostnameProg: regexp.MustCompile(`([a-z0-1]+\.)+[a-z0-1]+`),

		queue: queueFactory.RegisterQueue(&taskq.QueueOptions{
			Name:         "git-queue",
			MaxNumWorker: int32(p.QueueWorkerCount),
			BufferSize:   p.QueueSize,
			Storage:      taskq.NewLocalStorage(),
		}),
	}

	// Parse git version
	gitVersionOutput, err := utils.ExecProcess("git", "--version")
	if err != nil {
		return nil, fmt.Errorf("failed to run \"git --version\": %v", err)
	}
	prog := regexp.MustCompile(`([0-9]+)\.([0-9]+)\.(.+)`)
	gitVersionMatches := prog.FindStringSubmatch(gitVersionOutput)
	c.majorVersion, err = strconv.Atoi(gitVersionMatches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse git major version \"%v\": %v", gitVersionOutput, err)
	}
	c.minorVersion, err = strconv.Atoi(gitVersionMatches[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse git minor version \"%v\": %v", gitVersionOutput, err)
	}
	c.patchVersion = gitVersionMatches[3]
	fmt.Printf("Detected git version: major = %v minor = %v patch = %v\n", c.majorVersion, c.minorVersion, c.patchVersion)

	// Register tasks
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

func (c *gitClient) FetchLocalRepositoryPath(source fstree.RepositorySource) (localRepoLoc string, err error) {
	rid := source.GetRepositoryID()
	cloneUrl := source.GetCloneURL()
	defaultBranch := source.GetDefaultBranch()

	// Parse the url
	hostname := c.hostnameProg.FindString(cloneUrl)
	if hostname == "" {
		return "", fmt.Errorf("failed to match a valid hostname from \"%v\"", cloneUrl)
	}

	localRepoLoc = filepath.Join(c.CloneLocation, hostname, strconv.Itoa(int(rid)))
	if _, err := os.Stat(localRepoLoc); os.IsNotExist(err) {
		// Dispatch clone msg
		msg := c.cloneTask.WithArgs(context.Background(), cloneUrl, defaultBranch, localRepoLoc)
		msg.OnceInPeriod(time.Second, rid)
		c.queue.Add(msg)
	} else if c.AutoPull {
		// Dispatch pull msg
		msg := c.pullTask.WithArgs(context.Background(), localRepoLoc, defaultBranch)
		msg.OnceInPeriod(time.Second, rid)
		c.queue.Add(msg)
	}
	return localRepoLoc, nil
}
