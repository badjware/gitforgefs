package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/badjware/gitlabfs/config"
	"github.com/badjware/gitlabfs/fstree"
	"github.com/badjware/gitlabfs/git"
	"github.com/badjware/gitlabfs/platforms/github"
	"github.com/badjware/gitlabfs/platforms/gitlab"
)

func main() {
	configPath := flag.String("config", "", "The config file")
	mountoptionsFlag := flag.String("o", "", "Filesystem mount options. See mount.fuse(8)")
	debug := flag.Bool("debug", false, "Enable debug logging")

	flag.Usage = func() {
		fmt.Println("USAGE:")
		fmt.Printf("    %s MOUNTPOINT\n\n", os.Args[0])
		fmt.Println("OPTIONS:")
		flag.PrintDefaults()
	}
	flag.Parse()

	loadedConfig, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get logger
	logger := slog.Default()

	// Configure mountpoint
	mountpoint := loadedConfig.FS.Mountpoint
	if flag.NArg() == 1 {
		mountpoint = flag.Arg(0)
	}
	if mountpoint == "" {
		fmt.Println("Mountpoint is not configured in config file and missing from command-line arguments")
		flag.Usage()
		os.Exit(2)
	}

	// Configure mountoptions
	mountoptions := loadedConfig.FS.MountOptions
	if *mountoptionsFlag != "" {
		mountoptions = *mountoptionsFlag
	}
	parsedMountoptions := make([]string, 0)
	if mountoptions != "" {
		parsedMountoptions = strings.Split(mountoptions, ",")
	}

	// Create the git client
	gitClientParam, err := config.MakeGitConfig(loadedConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	gitClient, _ := git.NewClient(logger, *gitClientParam)

	var gitPlatformClient fstree.GitPlatform
	if loadedConfig.FS.Platform == config.PlatformGitlab {
		// Create the gitlab client
		GitlabClientConfig, err := config.MakeGitlabConfig(loadedConfig)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		gitPlatformClient, _ = gitlab.NewClient(logger, *GitlabClientConfig)
	} else if loadedConfig.FS.Platform == config.PlatformGithub {
		// Create the github client
		GithubClientConfig, err := config.MakeGithubConfig(loadedConfig)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		gitPlatformClient, _ = github.NewClient(logger, *GithubClientConfig)
	}

	// Start the filesystem
	err = fstree.Start(
		logger,
		mountpoint,
		parsedMountoptions,
		&fstree.FSParam{GitClient: gitClient, GitPlatform: gitPlatformClient},
		*debug,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
