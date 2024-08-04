package git

import (
	"fmt"
	"strconv"

	"github.com/badjware/gitlabfs/utils"
)

func (c *gitClient) pull(repoPath string, defaultBranch string) error {
	// Check if the local repo is on default branch
	branchName, err := utils.ExecProcessInDir(
		c.logger,
		repoPath, // workdir
		"git", "branch",
		"--show-current",
	)
	if err != nil {
		return fmt.Errorf("failed to retrieve HEAD of git repo %v: %v", repoPath, err)
	}

	if branchName == defaultBranch {
		// Pull the repo
		args := []string{
			"pull",
		}
		if c.GitClientConfig.Depth != 0 {
			args = append(args, "--depth", strconv.Itoa(c.GitClientConfig.Depth))
		}
		args = append(args,
			"--",
			c.GitClientConfig.Remote, // repository
			defaultBranch,            // refspec
		)

		_, err = utils.ExecProcessInDir(c.logger, repoPath, "git", args...)
		if err != nil {
			return fmt.Errorf("failed to pull git repo %v: %v", repoPath, err)
		}
	} else {
		c.logger.Info("Skipping pull because local is not on default branch", "currentBranch", branchName, "defaultBranch", defaultBranch)
	}

	return nil
}
