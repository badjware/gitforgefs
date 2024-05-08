package git

import (
	"fmt"
	"strconv"

	"github.com/badjware/gitlabfs/utils"
)

func (c *gitClient) clone(url string, defaultBranch string, dst string) error {
	if c.GitClientParam.OnClone == "init" {
		// "Fake" cloning the repo by never actually talking to the git server
		// This skip a fetch operation that we would do if we where to do a proper clone
		// We can save a lot of time and network i/o doing it this way, at the cost of
		// resulting in a very barebone local copy

		// Init the local repo
		fmt.Printf("Initializing %v into %v\n", url, dst)
		args := []string{
			"init",
		}
		if c.majorVersion > 2 || c.majorVersion == 2 && c.minorVersion >= 28 {
			args = append(args, "--initial-branch", defaultBranch)
		} else {
			fmt.Printf("Version of git is too old to support --initial-branch. Consider upgrading git to version >= 2.28.0")
		}
		args = append(args,
			"--",
			dst, // directory
		)
		_, err := utils.ExecProcess("git", args...)
		if err != nil {
			return fmt.Errorf("failed to init git repo %v to %v: %v", url, dst, err)
		}

		// Configure the remote
		_, err = utils.ExecProcessInDir(
			dst, // workdir
			"git", "remote", "add",
			"-m", defaultBranch,
			"--",
			c.GitClientParam.Remote, // name
			url,                     // url
		)
		if err != nil {
			return fmt.Errorf("failed to setup remote %v in git repo %v: %v", url, dst, err)
		}

		// Configure the default branch
		_, err = utils.ExecProcessInDir(
			dst, // workdir
			"git", "config", "--local",
			"--",
			fmt.Sprintf("branch.%s.remote", defaultBranch), // key
			c.GitClientParam.Remote,                        // value

		)
		if err != nil {
			return fmt.Errorf("failed to setup default branch remote in git repo %v: %v", dst, err)
		}
		_, err = utils.ExecProcessInDir(
			dst, // workdir
			"git", "config", "--local",
			"--",
			fmt.Sprintf("branch.%s.merge", defaultBranch), // key
			fmt.Sprintf("refs/heads/%s", defaultBranch),   // value

		)
		if err != nil {
			return fmt.Errorf("failed to setup default branch merge in git repo %v: %v", dst, err)
		}
	} else {
		// Clone the repo
		args := []string{
			"clone",
			"--origin", c.GitClientParam.Remote,
		}
		if c.GitClientParam.Depth != 0 {
			args = append(args, "--depth", strconv.Itoa(c.GitClientParam.Depth))
		}
		args = append(args,
			"--",
			url, // repository
			dst, // directory
		)

		_, err := utils.ExecProcess("git", args...)
		if err != nil {
			return fmt.Errorf("failed to clone git repo %v to %v: %v", url, dst, err)
		}
	}
	return nil
}
