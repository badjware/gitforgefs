package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	ForgeGitlab = "gitlab"
	ForgeGithub = "github"
	ForgeGitea  = "gitea"

	PullMethodHTTP = "http"
	PullMethodSSH  = "ssh"

	ArchivedProjectShow   = "show"
	ArchivedProjectHide   = "hide"
	ArchivedProjectIgnore = "ignore"
)

type (
	Config struct {
		FS     FSConfig           `yaml:"fs,omitempty"`
		Gitlab GitlabClientConfig `yaml:"gitlab,omitempty"`
		Github GithubClientConfig `yaml:"github,omitempty"`
		Gitea  GiteaClientConfig  `yaml:"gitea,omitempty"`
		Git    GitClientConfig    `yaml:"git,omitempty"`
	}
	FSConfig struct {
		Mountpoint   string `yaml:"mountpoint,omitempty"`
		MountOptions string `yaml:"mountoptions,omitempty"`
		Forge        string `yaml:"forge,omitempty"`
	}
	GitlabClientConfig struct {
		URL   string `yaml:"url,omitempty"`
		Token string `yaml:"token,omitempty"`

		GroupIDs []int `yaml:"group_ids,omitempty"`
		UserIDs  []int `yaml:"user_ids,omitempty"`

		ArchivedProjectHandling string `yaml:"archived_project_handling,omitempty"`
		IncludeCurrentUser      bool   `yaml:"include_current_user,omitempty"`
		PullMethod              string `yaml:"pull_method,omitempty"`
	}
	GithubClientConfig struct {
		Token string `yaml:"token,omitempty"`

		OrgNames  []string `yaml:"org_names,omitempty"`
		UserNames []string `yaml:"user_names,omitempty"`

		ArchivedRepoHandling string `yaml:"archived_repo_handling,omitempty"`
		IncludeCurrentUser   bool   `yaml:"include_current_user,omitempty"`
		PullMethod           string `yaml:"pull_method,omitempty"`
	}
	GiteaClientConfig struct {
		URL   string `yaml:"url,omitempty"`
		Token string `yaml:"token,omitempty"`

		OrgNames  []string `yaml:"org_names,omitempty"`
		UserNames []string `yaml:"user_names,omitempty"`

		ArchivedRepoHandling string `yaml:"archived_repo_handling,omitempty"`
		IncludeCurrentUser   bool   `yaml:"include_current_user,omitempty"`
		PullMethod           string `yaml:"pull_method,omitempty"`
	}
	GitClientConfig struct {
		CloneLocation    string `yaml:"clone_location,omitempty"`
		Remote           string `yaml:"remote,omitempty"`
		OnClone          string `yaml:"on_clone,omitempty"`
		AutoPull         bool   `yaml:"auto_pull,omitempty"`
		Depth            int    `yaml:"depth,omitempty"`
		QueueSize        int    `yaml:"queue_size,omitempty"`
		QueueWorkerCount int    `yaml:"worker_count,omitempty"`
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
			Forge:        "",
		},
		Gitlab: GitlabClientConfig{
			URL:                     "https://gitlab.com",
			Token:                   "",
			PullMethod:              "http",
			GroupIDs:                []int{9970},
			UserIDs:                 []int{},
			ArchivedProjectHandling: "hide",
			IncludeCurrentUser:      true,
		},
		Github: GithubClientConfig{
			Token:                "",
			PullMethod:           "http",
			OrgNames:             []string{},
			UserNames:            []string{},
			ArchivedRepoHandling: "hide",
			IncludeCurrentUser:   true,
		},
		Git: GitClientConfig{
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

	// validate forge is set
	if config.FS.Forge != ForgeGithub && config.FS.Forge != ForgeGitlab && config.FS.Forge != ForgeGitea {
		return nil, fmt.Errorf("fs.forge must be either \"%v\", \"%v\", or \"%v\"", ForgeGitlab, ForgeGithub, ForgeGitea)
	}

	return config, nil
}

func MakeGitlabConfig(config *Config) (*GitlabClientConfig, error) {
	// parse pull_method
	if config.Gitlab.PullMethod != PullMethodHTTP && config.Gitlab.PullMethod != PullMethodSSH {
		return nil, fmt.Errorf("gitlab.pull_method must be either \"%v\" or \"%v\"", PullMethodHTTP, PullMethodSSH)
	}

	// parse archive_handing
	if config.Gitlab.ArchivedProjectHandling != ArchivedProjectShow && config.Gitlab.ArchivedProjectHandling != ArchivedProjectHide && config.Gitlab.ArchivedProjectHandling != ArchivedProjectIgnore {
		return nil, fmt.Errorf("gitlab.archived_project_handling must be either \"%v\", \"%v\" or \"%v\"", ArchivedProjectShow, ArchivedProjectHide, ArchivedProjectIgnore)
	}

	return &config.Gitlab, nil
}

func MakeGithubConfig(config *Config) (*GithubClientConfig, error) {
	// parse pull_method
	if config.Github.PullMethod != PullMethodHTTP && config.Github.PullMethod != PullMethodSSH {
		return nil, fmt.Errorf("github.pull_method must be either \"%v\" or \"%v\"", PullMethodHTTP, PullMethodSSH)
	}

	// parse archive_handing
	if config.Github.ArchivedRepoHandling != ArchivedProjectShow && config.Github.ArchivedRepoHandling != ArchivedProjectHide && config.Github.ArchivedRepoHandling != ArchivedProjectIgnore {
		return nil, fmt.Errorf("github.archived_repo_handling must be either \"%v\", \"%v\" or \"%v\"", ArchivedProjectShow, ArchivedProjectHide, ArchivedProjectIgnore)
	}

	return &config.Github, nil
}

func MakeGiteaConfig(config *Config) (*GiteaClientConfig, error) {
	// parse pull_method
	if config.Gitea.PullMethod != PullMethodHTTP && config.Gitea.PullMethod != PullMethodSSH {
		return nil, fmt.Errorf("gitea.pull_method must be either \"%v\" or \"%v\"", PullMethodHTTP, PullMethodSSH)
	}

	// parse archive_handing
	if config.Gitea.ArchivedRepoHandling != ArchivedProjectShow && config.Gitea.ArchivedRepoHandling != ArchivedProjectHide && config.Gitea.ArchivedRepoHandling != ArchivedProjectIgnore {
		return nil, fmt.Errorf("gitea.archived_repo_handling must be either \"%v\", \"%v\" or \"%v\"", ArchivedProjectShow, ArchivedProjectHide, ArchivedProjectIgnore)
	}

	return &config.Gitea, nil
}

func MakeGitConfig(config *Config) (*GitClientConfig, error) {
	// parse on_clone
	if config.Git.OnClone != "init" && config.Git.OnClone != "clone" {
		return nil, fmt.Errorf("git.on_clone must be either \"init\" or \"clone\"")
	}

	return &config.Git, nil
}
