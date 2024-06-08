package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/badjware/gitlabfs/git"
	"github.com/badjware/gitlabfs/platforms/gitlab"
	"gopkg.in/yaml.v2"
)

type (
	Config struct {
		FS     FSConfig                  `yaml:"fs,omitempty"`
		Gitlab gitlab.GitlabClientConfig `yaml:"gitlab,omitempty"`
		Git    git.GitClientParam        `yaml:"git,omitempty"`
	}
	FSConfig struct {
		Mountpoint   string `yaml:"mountpoint,omitempty"`
		MountOptions string `yaml:"mountoptions,omitempty"`
	}
)

func LoadConfig(configPath string) (*Config, error) {
	// defaults
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(os.Getenv("HOME"), ".local/share")
	}
	defaultCloneLocation := filepath.Join(dataHome, "gitlabfs")

	config := &Config{
		FS: FSConfig{
			Mountpoint:   "",
			MountOptions: "nodev,nosuid",
		},
		Gitlab: gitlab.GitlabClientConfig{
			URL:                "https://gitlab.com",
			Token:              "",
			PullMethod:         "http",
			GroupIDs:           []int{9970},
			UserIDs:            []int{},
			IncludeCurrentUser: true,
		},
		Git: git.GitClientParam{
			CloneLocation:    defaultCloneLocation,
			Remote:           "origin",
			OnClone:          "init",
			AutoPull:         false,
			Depth:            0,
			QueueSize:        200,
			QueueWorkerCount: 5,
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

func MakeGitConfig(config *Config) (*git.GitClientParam, error) {
	// parse on_clone
	if config.Git.OnClone != "init" && config.Git.OnClone != "clone" {
		return nil, fmt.Errorf("on_clone must be either \"init\" or \"clone\"")
	}

	return &config.Git, nil
}

func MakeGitlabConfig(config *Config) (*gitlab.GitlabClientConfig, error) {
	// parse pull_method
	if config.Gitlab.PullMethod != gitlab.PullMethodHTTP && config.Gitlab.PullMethod != gitlab.PullMethodSSH {
		return nil, fmt.Errorf("pull_method must be either \"%v\" or \"%v\"", gitlab.PullMethodHTTP, gitlab.PullMethodSSH)
	}

	return &config.Gitlab, nil
}
