package git

import (
	"fmt"
	"os"
	"strconv"

	"github.com/badjware/gitlabfs/utils"
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
	if c.CloneMethod == CloneInit {
		// "Fake" cloning the repo by never actually talking to the git server
		// This skip a fetch operation that we would do if we where to do a proper clone
		// We can save a lot of time and network i/o doing it this way, at the cost of
		// resulting in a very barebone local copy

		// Init the local repo
		fmt.Printf("Initializing %v into %v\n", gcp.url, gcp.dst)
		_, err := utils.ExecProcess(
			"git", "init",
			"--initial-branch", gcp.defaultBranch,
			"--",
			gcp.dst, // directory
		)
		if err != nil {
			return fmt.Errorf("failed to init git repo %v to %v: %v", gcp.url, gcp.dst, err)
		}

		// Configure the remote
		_, err = utils.ExecProcessInDir(
			gcp.dst, // workdir
			"git", "remote", "add",
			"-m", gcp.defaultBranch,
			"--",
			c.RemoteName, // name
			gcp.url,      // url
		)
		if err != nil {
			return fmt.Errorf("failed to setup remote %v in git repo %v: %v", gcp.url, gcp.dst, err)
		}

		// Configure the default branch
		_, err = utils.ExecProcessInDir(
			gcp.dst, // workdir
			"git", "config", "--local",
			"--",
			fmt.Sprintf("branch.%s.remote", gcp.defaultBranch), // key
			c.RemoteName, // value

		)
		if err != nil {
			return fmt.Errorf("failed to setup default branch remote in git repo %v: %v", gcp.dst, err)
		}
		_, err = utils.ExecProcessInDir(
			gcp.dst, // workdir
			"git", "config", "--local",
			"--",
			fmt.Sprintf("branch.%s.merge", gcp.defaultBranch), // key
			fmt.Sprintf("refs/heads/%s", gcp.defaultBranch),   // value

		)
		if err != nil {
			return fmt.Errorf("failed to setup default branch merge in git repo %v: %v", gcp.dst, err)
		}
	} else {
		// Clone the repo
		_, err := utils.ExecProcess(
			"git", "clone",
			"--origin", c.RemoteName,
			"--depth", strconv.Itoa(c.PullDepth),
			"--",
			gcp.url, // repository
			gcp.dst, // directory
		)
		if err != nil {
			return fmt.Errorf("failed to clone git repo %v to %v: %v", gcp.url, gcp.dst, err)
		}
	}
	return nil
}
