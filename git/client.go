package git

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/badjware/gitforgefs/config"
	"github.com/badjware/gitforgefs/queue"
	"github.com/badjware/gitforgefs/types"
	"github.com/badjware/gitforgefs/utils"
)

type gitClient struct {
	config.GitClientConfig

	logger *slog.Logger

	hostnameProg *regexp.Regexp

	majorVersion int
	minorVersion int
	patchVersion string

	queue queue.TaskQueue
}

func NewClient(logger *slog.Logger, p config.GitClientConfig) (*gitClient, error) {
	// Create the client
	c := &gitClient{
		GitClientConfig: p,

		logger: logger,

		hostnameProg: regexp.MustCompile(`([a-z0-1\-]+\.)+[a-z0-1\-]+`),

		queue: queue.NewMemoryQueue(logger, "git-queue", p.QueueWorkerCount, p.QueueSize),
	}

	// Parse git version
	gitVersionOutput, err := utils.ExecProcess(logger, "git", "--version")
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
	logger.Info("Detected git version", "major", c.majorVersion, "minor", c.minorVersion, "patch", c.patchVersion)

	return c, nil
}

func (c *gitClient) FetchLocalRepositoryPath(ctx context.Context, source types.RepositorySource) (localRepoLoc string, err error) {
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
		// Dispatch clone task
		c.queue.AddTask(func() {
			c.clone(cloneUrl, defaultBranch, localRepoLoc)
		})
	} else if c.AutoPull {
		// Dispatch pull task
		c.queue.AddTask(func() {
			c.pull(localRepoLoc, defaultBranch)
		})
	}
	return localRepoLoc, nil
}
