package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/badjware/gitlabfs/fs"
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

	gitlabClient, _ := gitlab.NewClient(*gitlabURL, *gitlabToken)

	fs.Start(gitlabClient, mountpoint, *gitlabRootGroupID)

	// content, err := gitlabClient.FetchGroupContent(&rootGroup)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println("Projects")
	// for _, r := range content.Repositories {
	// 	fmt.Println(r.Name, r.Path, r.CloneURL)
	// }
	// fmt.Println("Groups")
	// for _, g := range content.Groups {
	// 	fmt.Println(g.Name, g.Path)
	// }
}
