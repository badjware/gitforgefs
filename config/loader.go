package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
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
		FS     FSConfig           `koanf:"fs"`
		Gitlab GitlabClientConfig `koanf:"gitlab"`
		Github GithubClientConfig `koanf:"github"`
		Gitea  GiteaClientConfig  `koanf:"gitea"`
		Git    GitClientConfig    `koanf:"git"`
	}
	FSConfig struct {
		Mountpoint   string `koanf:"mountpoint"`
		MountOptions string `koanf:"mountoptions"`
		UseSymlinks  bool   `koanf:"use_symlinks"`
		Forge        string `koanf:"forge"`
	}
	GitlabClientConfig struct {
		URL   string `koanf:"url"`
		Token string `koanf:"token"`

		GroupIDs  []int64  `koanf:"group_ids"`
		UserNames []string `koanf:"user_names"`

		ArchivedProjectHandling string `koanf:"archived_project_handling"`
		IncludeCurrentUser      bool   `koanf:"include_current_user"`
		PullMethod              string `koanf:"pull_method"`
	}
	GithubClientConfig struct {
		Token string `koanf:"token"`

		OrgNames  []string `koanf:"org_names"`
		UserNames []string `koanf:"user_names"`

		ArchivedRepoHandling string `koanf:"archived_repo_handling"`
		IncludeCurrentUser   bool   `koanf:"include_current_user"`
		PullMethod           string `koanf:"pull_method"`
	}
	GiteaClientConfig struct {
		URL   string `koanf:"url"`
		Token string `koanf:"token"`

		OrgNames  []string `koanf:"org_names"`
		UserNames []string `koanf:"user_names"`

		ArchivedRepoHandling string `koanf:"archived_repo_handling"`
		IncludeCurrentUser   bool   `koanf:"include_current_user"`
		PullMethod           string `koanf:"pull_method"`
	}
	GitClientConfig struct {
		CloneLocation    string `koanf:"clone_location"`
		Remote           string `koanf:"remote"`
		OnClone          string `koanf:"on_clone"`
		AutoPull         bool   `koanf:"auto_pull"`
		Depth            int    `koanf:"depth"`
		QueueSize        int    `koanf:"queue_size"`
		QueueWorkerCount int    `koanf:"worker_count"`
	}
)

func Default() Config {
	// defaults
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(os.Getenv("HOME"), ".local/share")
	}
	defaultCloneLocation := filepath.Join(dataHome, "gitforgefs")

	return Config{
		FS: FSConfig{
			Mountpoint:   "",
			MountOptions: "nodev,nosuid",
			UseSymlinks:  true,
			Forge:        "",
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
		Gitea: GiteaClientConfig{
			URL:                  "https://codeberg.org",
			Token:                "",
			PullMethod:           "http",
			OrgNames:             []string{},
			UserNames:            []string{},
			ArchivedRepoHandling: "hide",
			IncludeCurrentUser:   true,
		},
		Gitlab: GitlabClientConfig{
			URL:                     "https://gitlab.com",
			Token:                   "",
			PullMethod:              "http",
			GroupIDs:                []int64{9970},
			UserNames:               []string{},
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
	}
}

func LoadConfig(configPath string) (Config, error) {
	k := koanf.New(".")

	// load file
	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return Default(), fmt.Errorf("failed to load config file: %v", err)
	}

	// unmarshal on top of defaults
	config := Default()
	if err := k.Unmarshal("", &config); err != nil {
		return config, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// validate
	if err := Validate(config); err != nil {
		return config, fmt.Errorf("failed to validate config: %v", err)
	}

	return config, nil
}

func Validate(config Config) error {
	// fs

	// validate forge is set
	if config.FS.Forge != ForgeGithub && config.FS.Forge != ForgeGitlab && config.FS.Forge != ForgeGitea {
		return fmt.Errorf("fs.forge must be either \"%v\", \"%v\", or \"%v\"", ForgeGitlab, ForgeGithub, ForgeGitea)
	}

	// git

	// parse on_clone
	if config.Git.OnClone != "init" && config.Git.OnClone != "clone" {
		return fmt.Errorf("git.on_clone must be either \"init\" or \"clone\"")
	}

	// gitlab

	if config.Gitlab.PullMethod != PullMethodHTTP && config.Gitlab.PullMethod != PullMethodSSH {
		return fmt.Errorf("gitlab.pull_method must be either \"%v\" or \"%v\"", PullMethodHTTP, PullMethodSSH)
	}

	// parse archive_handing
	if config.Gitlab.ArchivedProjectHandling != ArchivedProjectShow && config.Gitlab.ArchivedProjectHandling != ArchivedProjectHide && config.Gitlab.ArchivedProjectHandling != ArchivedProjectIgnore {
		return fmt.Errorf("gitlab.archived_project_handling must be either \"%v\", \"%v\" or \"%v\"", ArchivedProjectShow, ArchivedProjectHide, ArchivedProjectIgnore)
	}

	// github

	// parse pull_method
	if config.Github.PullMethod != PullMethodHTTP && config.Github.PullMethod != PullMethodSSH {
		return fmt.Errorf("github.pull_method must be either \"%v\" or \"%v\"", PullMethodHTTP, PullMethodSSH)
	}

	// parse archive_handing
	if config.Github.ArchivedRepoHandling != ArchivedProjectShow && config.Github.ArchivedRepoHandling != ArchivedProjectHide && config.Github.ArchivedRepoHandling != ArchivedProjectIgnore {
		return fmt.Errorf("github.archived_repo_handling must be either \"%v\", \"%v\" or \"%v\"", ArchivedProjectShow, ArchivedProjectHide, ArchivedProjectIgnore)
	}

	// gitea

	// parse pull_method
	if config.Gitea.PullMethod != PullMethodHTTP && config.Gitea.PullMethod != PullMethodSSH {
		return fmt.Errorf("gitea.pull_method must be either \"%v\" or \"%v\"", PullMethodHTTP, PullMethodSSH)
	}

	// parse archive_handing
	if config.Gitea.ArchivedRepoHandling != ArchivedProjectShow && config.Gitea.ArchivedRepoHandling != ArchivedProjectHide && config.Gitea.ArchivedRepoHandling != ArchivedProjectIgnore {
		return fmt.Errorf("gitea.archived_repo_handling must be either \"%v\", \"%v\" or \"%v\"", ArchivedProjectShow, ArchivedProjectHide, ArchivedProjectIgnore)
	}

	return nil
}
