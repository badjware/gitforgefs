package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/badjware/gitlabfs/fs"
	"github.com/badjware/gitlabfs/git"
	"github.com/badjware/gitlabfs/gitlab"
	"gopkg.in/yaml.v2"
)

const (
	OnCloneInit       = "init"
	OnCloneNoCheckout = "no-checkout"
	OnCloneCheckout   = "checkout"
)

type (
	Config struct {
		FS     FSConfig     `yaml:"fs,omitempty"`
		Gitlab GitlabConfig `yaml:"gitlab,omitempty"`
		Git    GitConfig    `yaml:"git,omitempty"`
	}
	FSConfig struct {
		Mountpoint string `yaml:"mountpoint,omitempty"`
	}
	GitlabConfig struct {
		URL                string `yaml:"url,omitempty"`
		Token              string `yaml:"token,omitempty"`
		GroupIDs           []int  `yaml:"group_ids,omitempty"`
		UserIDs            []int  `yaml:"user_ids,omitempty"`
		IncludeCurrentUser bool   `yaml:"include_current_user,omitempty"`
	}
	GitConfig struct {
		CloneLocation string `yaml:"clone_location,omitempty"`
		Remote        string `yaml:"remote,omitempty"`
		PullMethod    string `yaml:"pull_method,omitempty"`
		OnClone       string `yaml:"on_clone,omitempty"`
		AutoPull      bool   `yaml:"auto_pull,omitempty"`
		Depth         int    `yaml:"depth,omitempty"`
	}
)

func loadConfig(configPath string) (*Config, error) {
	// defaults
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(os.Getenv("HOME"), ".local/share")
	}
	defaultCloneLocation := filepath.Join(dataHome, "gitlabfs")

	config := &Config{
		FS: FSConfig{
			Mountpoint: "",
		},
		Gitlab: GitlabConfig{
			URL:                "https://gitlab.com",
			Token:              "",
			GroupIDs:           []int{9970},
			UserIDs:            []int{},
			IncludeCurrentUser: true,
		},
		Git: GitConfig{
			CloneLocation: defaultCloneLocation,
			Remote:        "origin",
			PullMethod:    "http",
			OnClone:       "init",
			AutoPull:      false,
			Depth:         0,
		},
	}

	if configPath != "" {
		f, err := os.Open(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file: %v", err)
		}
		defer f.Close()

		d := yaml.NewDecoder(f)
		if err := d.Decode(config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %v", err)
		}
	}

	return config, nil
}

func makeGitlabConfig(config *Config) (*gitlab.GitlabClientParam, error) {
	// parse pull_method
	if config.Git.PullMethod != gitlab.PullMethodHTTP && config.Git.PullMethod != gitlab.PullMethodSSH {
		return nil, fmt.Errorf("pull_method must be either \"%v\" or \"%v\"", gitlab.PullMethodHTTP, gitlab.PullMethodSSH)
	}

	return &gitlab.GitlabClientParam{
		PullMethod: config.Git.PullMethod,
	}, nil
}

func makeGitConfig(config *Config) (*git.GitClientParam, error) {
	// Parse the gilab url
	parsedGitlabURL, err := url.Parse(config.Gitlab.URL)
	if err != nil {
		return nil, err
	}

	// parse on_clone
	fetch := false
	checkout := false
	if config.Git.OnClone == OnCloneInit {
		fetch = false
		checkout = false
	} else if config.Git.OnClone == OnCloneNoCheckout {
		fetch = true
		checkout = false
	} else if config.Git.OnClone == OnCloneCheckout {
		fetch = true
		checkout = true
	} else {
		return nil, fmt.Errorf("on_clone must be either \"%v\", \"%v\" or \"%V\"", OnCloneInit, OnCloneNoCheckout, OnCloneCheckout)
	}

	return &git.GitClientParam{
		CloneLocation: config.Git.CloneLocation,
		RemoteName:    config.Git.Remote,
		RemoteURL:     parsedGitlabURL,
		Fetch:         fetch,
		Checkout:      checkout,
		AutoPull:      config.Git.AutoPull,
		PullDepth:     config.Git.Depth,
	}, nil
}

func main() {
	configPath := flag.String("config", "", "the config file")
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Configure mountpoint
	mountpoint := config.FS.Mountpoint
	if flag.NArg() == 1 {
		mountpoint = flag.Arg(0)
	}
	if mountpoint == "" {
		fmt.Printf("usage: %s MOUNTPOINT\n", os.Args[0])
		os.Exit(2)
	}

	// Create the git client
	gitClientParam, err := makeGitConfig(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	gitClient, _ := git.NewClient(*gitClientParam)

	// Create the gitlab client
	gitlabClientParam, err := makeGitlabConfig(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	gitlabClient, _ := gitlab.NewClient(config.Gitlab.URL, config.Gitlab.Token, *gitlabClientParam)

	// Start the filesystem
	fs.Start(
		mountpoint,
		config.Gitlab.GroupIDs,
		config.Gitlab.UserIDs,
		&fs.FSParam{Git: gitClient, Gitlab: gitlabClient},
		*debug,
	)
}
