package git

import (
	"fmt"
	"strconv"

	"github.com/badjware/gitlabfs/utils"
)

func (c *gitClient) pull(repoPath string, defaultBranch string) error {
	// Check if the local repo is on default branch
	branchName, err := utils.ExecProcessInDir(
		repoPath, // workdir
		"git", "branch",
		"--show-current",
	)
	if err != nil {
		return fmt.Errorf("failed to retrieve HEAD of git repo %v: %v", repoPath, err)
	}

	if branchName == defaultBranch {
		// Pull the repo
		_, err = utils.ExecProcessInDir(
			repoPath, // workdir
			"git", "pull",
			"--depth", strconv.Itoa(c.GitClientParam.Depth),
			"--",
			c.GitClientParam.Remote, // repository
			defaultBranch,           // refspec
		)
		if err != nil {
			return fmt.Errorf("failed to pull git repo %v: %v", repoPath, err)
		}
	} else {
		fmt.Printf("%v != %v, skipping pull", branchName, defaultBranch)
	}

	return nil
}
