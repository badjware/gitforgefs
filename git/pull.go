package git

import (
	"fmt"
	"strconv"

	"github.com/badjware/gitlabfs/utils"
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
	// r, err := git.PlainOpen(gpp.repoPath)
	// if err != nil {
	// 	return fmt.Errorf("failed to open git repo %v: %v", gpp.repoPath, err)
	// }
	// w, err := r.Worktree()
	// if err != nil {
	// 	return fmt.Errorf("failed to retrieve worktree of git repo %v: %v", gpp.repoPath, err)
	// }

	// Check if the local repo is on default branch
	// headRef, err := r.Head()
	branchName, err := utils.ExecProcessInDir(
		gpp.repoPath, // workdir
		"git", "branch",
		"--show-current",
	)
	if err != nil {
		return fmt.Errorf("failed to retrieve HEAD of git repo %v: %v", gpp.repoPath, err)
	}

	if branchName == gpp.defaultBranch {
		_, err = utils.ExecProcessInDir(
			gpp.repoPath, // workdir
			"git", "pull",
			"--depth", strconv.Itoa(c.PullDepth),
			"--",
			c.RemoteName,      // repository
			gpp.defaultBranch, // refspec
		)
		if err != nil {
			return fmt.Errorf("failed to pull git repo %v: %v", gpp.repoPath, err)
		}
	} else {
		fmt.Printf("%v != %v, skipping pull", branchName, gpp.defaultBranch)
	}

	return nil
}
