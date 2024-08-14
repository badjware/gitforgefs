package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/badjware/gitlabfs/config"
	"github.com/badjware/gitlabfs/forges/gitea"
	"github.com/badjware/gitlabfs/forges/github"
	"github.com/badjware/gitlabfs/forges/gitlab"
	"github.com/badjware/gitlabfs/fstree"
	"github.com/badjware/gitlabfs/git"
)

func main() {
	configPath := flag.String("config", "config.yaml", "The config file")
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

	var gitForgeClient fstree.GitForge
	if loadedConfig.FS.Forge == config.ForgeGitlab {
		// Create the gitlab client
		gitlabClientConfig, err := config.MakeGitlabConfig(loadedConfig)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		gitForgeClient, _ = gitlab.NewClient(logger, *gitlabClientConfig)
	} else if loadedConfig.FS.Forge == config.ForgeGithub {
		// Create the github client
		githubClientConfig, err := config.MakeGithubConfig(loadedConfig)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		gitForgeClient, _ = github.NewClient(logger, *githubClientConfig)
	} else if loadedConfig.FS.Forge == config.ForgeGitea {
		// Create the gitea client
		giteaClientConfig, err := config.MakeGiteaConfig(loadedConfig)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		gitForgeClient, _ = gitea.NewClient(logger, *giteaClientConfig)
	}

	// Start the filesystem
	err = fstree.Start(
		logger,
		mountpoint,
		parsedMountoptions,
		&fstree.FSParam{GitClient: gitClient, GitForge: gitForgeClient},
		*debug,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
