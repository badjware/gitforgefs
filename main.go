package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/badjware/gitlabfs/gitlab"
)

func main() {
	gitlabURL := flag.String("gitlab-url", "https://gitlab.com", "the gitlab url")
	gitlabToken := flag.String("gitlab-token", "", "the gitlab authentication token")
	gitlabRootGroupID := flag.Int("gitlab-group-id", 9970, "the group id of the groups at the root of the filesystem")
	// gitlabNamespace := flag.String()

	flag.Parse()

	gitlabClient, _ := gitlab.NewClient(*gitlabURL, *gitlabToken)

	// TODO: move this
	group, _, err := gitlabClient.Client.Groups.GetGroup(*gitlabRootGroupID)
	if err != nil {
		fmt.Printf("failed to fetch root group with id %v: %w\n", *gitlabRootGroupID, err)
		os.Exit(1)
	}
	rootGroup := gitlab.NewGroupFromGitlabGroup(group)
	fmt.Printf("Root group: %v\n", rootGroup.Name)

	content, err := gitlabClient.FetchGroupContent(&rootGroup)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Projects")
	for _, r := range content.Repositories {
		fmt.Println(r.Name, r.Path, r.CloneURL)
	}
	fmt.Println("Groups")
	for _, g := range content.Groups {
		fmt.Println(g.Name, g.Path)
	}
}
