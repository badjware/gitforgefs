package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/badjware/gitlabfs/fs"
	"github.com/badjware/gitlabfs/git"
	"github.com/badjware/gitlabfs/gitlab"
)

func main() {
	gitlabURL := flag.String("gitlab-url", "https://gitlab.com", "the gitlab url")
	gitlabToken := flag.String("gitlab-token", "", "the gitlab authentication token")
	gitlabRootGroupID := flag.Int("gitlab-group-id", 9970, "the group id of the groups at the root of the filesystem")
	// gitlabNamespace := flag.String()
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Printf("usage: %s MOUNTPOINT\n", os.Args[0])
		os.Exit(2)
	}
	mountpoint := flag.Arg(0)
	parsedGitlabURL, err := url.Parse(*gitlabURL)
	if err != nil {
		fmt.Printf("%v is not a valid url: %v\n", *gitlabURL, err)
		os.Exit(1)
	}

	// Create the gitlab client
	gitlabClientParam := gitlab.GitlabClientParam{}
	gitlabClient, _ := gitlab.NewClient(*gitlabURL, *gitlabToken, gitlabClientParam)

	// Create the git client
	gitClientParam := git.GitClientParam{
		RemoteURL:    parsedGitlabURL,
		AutoClone:    true,
		AutoPull:     false,
		Fetch:        false,
		Checkout:     false,
		SingleBranch: true,
		PullDepth:    0,
	}
	gitClient, _ := git.NewClient(gitClientParam)

	// Start the filesystem
	fs.Start(mountpoint, []int{*gitlabRootGroupID}, []int{}, &fs.FSParam{Gf: gitlabClient, Gcp: gitClient})
}
