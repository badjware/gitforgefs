package config_test

import (
	"reflect"
	"testing"

	"github.com/badjware/gitforgefs/config"
)

// TestDefaultMatchesExample checks that the values in config.example.yaml are the defaults.
func TestDefaultMatchesExample(t *testing.T) {
	got, err := config.LoadConfig("../config.example.yaml")
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}

	expected := config.Default()
	// fs.forge has no default and is set in the example
	expected.FS.Forge = config.ForgeGitlab

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("config.example.yaml does not match the defaults\ngot: %+v\nexpected: %+v", got, expected)
	}
}

// TestLoadConfig checks that LoadConfig correctly loads a config file and returns the expected Config struct.
func TestLoadConfig(t *testing.T) {
	got, err := config.LoadConfig("config.test.yaml")
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}

	expected := config.Config{
		FS: config.FSConfig{
			Mountpoint:   "/tmp/gitforgefs/test/mnt/gitlab",
			MountOptions: "nodev",
			Forge:        "gitlab",
			UseSymlinks:  true,
		},
		Gitlab: config.GitlabClientConfig{
			URL:                     "https://example.com",
			Token:                   "12345",
			PullMethod:              "ssh",
			GroupIDs:                []int64{123},
			UserNames:               []string{"test-user"},
			ArchivedProjectHandling: "hide",
			IncludeCurrentUser:      true,
		},
		Github: config.GithubClientConfig{
			Token:                "12345",
			PullMethod:           "http",
			OrgNames:             []string{"test-org"},
			UserNames:            []string{"test-user"},
			ArchivedRepoHandling: "hide",
			IncludeCurrentUser:   true,
		},
		Gitea: config.GiteaClientConfig{
			URL:                  "https://example.com",
			Token:                "12345",
			PullMethod:           "http",
			OrgNames:             []string{"test-org"},
			UserNames:            []string{"test-user"},
			ArchivedRepoHandling: "hide",
			IncludeCurrentUser:   true,
		},
		Git: config.GitClientConfig{
			CloneLocation:    "/tmp/gitforgefs/test/cache/gitlab",
			Remote:           "origin",
			OnClone:          "clone",
			AutoPull:         false,
			Depth:            0,
			QueueSize:        100,
			QueueWorkerCount: 1,
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("LoadConfig returned an unexpected config\ngot: %+v\nexpected: %+v", got, expected)
	}
}

// TestValidate checks that Validate correctly validates the configuration and returns an error for invalid configurations.
func TestValidate(t *testing.T) {
	tests := map[string]struct {
		mutate  func(c *config.Config)
		wantErr bool
	}{
		"Valid":        {func(c *config.Config) {}, false},
		"ForgeGithub":  {func(c *config.Config) { c.FS.Forge = config.ForgeGithub }, false},
		"ForgeGitea":   {func(c *config.Config) { c.FS.Forge = config.ForgeGitea }, false},
		"ForgeEmpty":   {func(c *config.Config) { c.FS.Forge = "" }, true},
		"ForgeInvalid": {func(c *config.Config) { c.FS.Forge = "invalid" }, true},

		"GitOnCloneClone":   {func(c *config.Config) { c.Git.OnClone = "clone" }, false},
		"GitOnCloneInvalid": {func(c *config.Config) { c.Git.OnClone = "invalid" }, true},

		"GitlabPullMethodSSH":     {func(c *config.Config) { c.Gitlab.PullMethod = config.PullMethodSSH }, false},
		"GitlabPullMethodInvalid": {func(c *config.Config) { c.Gitlab.PullMethod = "invalid" }, true},
		"GitlabArchivedShow":      {func(c *config.Config) { c.Gitlab.ArchivedProjectHandling = config.ArchivedProjectShow }, false},
		"GitlabArchivedIgnore":    {func(c *config.Config) { c.Gitlab.ArchivedProjectHandling = config.ArchivedProjectIgnore }, false},
		"GitlabArchivedInvalid":   {func(c *config.Config) { c.Gitlab.ArchivedProjectHandling = "invalid" }, true},

		"GithubPullMethodSSH":     {func(c *config.Config) { c.Github.PullMethod = config.PullMethodSSH }, false},
		"GithubPullMethodInvalid": {func(c *config.Config) { c.Github.PullMethod = "invalid" }, true},
		"GithubArchivedShow":      {func(c *config.Config) { c.Github.ArchivedRepoHandling = config.ArchivedProjectShow }, false},
		"GithubArchivedIgnore":    {func(c *config.Config) { c.Github.ArchivedRepoHandling = config.ArchivedProjectIgnore }, false},
		"GithubArchivedInvalid":   {func(c *config.Config) { c.Github.ArchivedRepoHandling = "invalid" }, true},

		"GiteaPullMethodSSH":     {func(c *config.Config) { c.Gitea.PullMethod = config.PullMethodSSH }, false},
		"GiteaPullMethodInvalid": {func(c *config.Config) { c.Gitea.PullMethod = "invalid" }, true},
		"GiteaArchivedShow":      {func(c *config.Config) { c.Gitea.ArchivedRepoHandling = config.ArchivedProjectShow }, false},
		"GiteaArchivedIgnore":    {func(c *config.Config) { c.Gitea.ArchivedRepoHandling = config.ArchivedProjectIgnore }, false},
		"GiteaArchivedInvalid":   {func(c *config.Config) { c.Gitea.ArchivedRepoHandling = "invalid" }, true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			c := config.Default()
			c.FS.Forge = config.ForgeGitlab
			test.mutate(&c)
			err := config.Validate(c)
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate returned error %v; expected an error: %v", err, test.wantErr)
			}
		})
	}
}
