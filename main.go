package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/debug"
	"strings"

	"github.com/badjware/gitforgefs/cache"
	"github.com/badjware/gitforgefs/config"
	"github.com/badjware/gitforgefs/forges/gitea"
	"github.com/badjware/gitforgefs/forges/github"
	"github.com/badjware/gitforgefs/forges/gitlab"
	"github.com/badjware/gitforgefs/fstree"
	"github.com/badjware/gitforgefs/git"
	"github.com/badjware/gitforgefs/types"
)

func main() {
	configPath := flag.String("config", "config.yaml", "The config file")
	mountoptionsFlag := flag.String("o", "", "Filesystem mount options. See mount.fuse(8)")
	versionFlag := flag.Bool("version", false, "Print version information and exit")
	debugFlag := flag.Bool("debug", false, "Enable debug logging")
	debugPort := flag.Int("debug-port", 0, "Listen port for debug server. If 0, server is disabled")

	flag.Usage = func() {
		fmt.Println("USAGE:")
		fmt.Printf("    %s MOUNTPOINT\n\n", os.Args[0])
		fmt.Println("OPTIONS:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *versionFlag {
		version := "unknown"
		if info, ok := debug.ReadBuildInfo(); ok {
			version = info.Main.Version
		}
		fmt.Println(version)
		os.Exit(0)
	}

	loadedConfig, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get logger
	var level slog.Level
	if *debugFlag {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	// start pprof server if debug port is set
	if *debugPort != 0 {
		go func() {
			logger.Info("Starting debug server", "port", *debugPort)
			err := http.ListenAndServe(fmt.Sprintf("localhost:%d", *debugPort), nil)
			if err != nil {
				logger.Error("debug server failed", "error", err)
			}
		}()
	}

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

	// setup backend
	var gitForgeClient types.GitForge
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

	// setup cache
	cache := cache.NewForgeCache(gitForgeClient, logger)

	// Start the filesystem
	err = fstree.Start(
		logger,
		mountpoint,
		parsedMountoptions,
		&fstree.FSParam{
			UseSymlinks: loadedConfig.FS.UseSymlinks,
			GitClient:   gitClient,
			Backend:     cache,
		},
		*debugFlag,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
