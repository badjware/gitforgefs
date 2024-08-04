package github

import (
	"log/slog"

	"github.com/badjware/gitlabfs/config"
	"github.com/badjware/gitlabfs/fstree"
	"github.com/google/go-github/v63/github"
)

type githubClient struct {
	config.GithubClientConfig
	client *github.Client

	logger *slog.Logger

	rootContent map[string]fstree.GroupSource
}

func NewClient(logger *slog.Logger, config config.GithubClientConfig) (*githubClient, error) {
	client := github.NewClient(nil)
	if config.Token != "" {
		client = client.WithAuthToken(config.Token)
	}

	gitHubClient := &githubClient{
		GithubClientConfig: config,
		client:             client,

		logger: logger,
	}

	return gitHubClient, nil
}

func (c *githubClient) FetchRootGroupContent() (map[string]fstree.GroupSource, error) {
	return nil, nil
}

func (c *githubClient) FetchGroupContent(gid uint64) (map[string]fstree.GroupSource, map[string]fstree.RepositorySource, error) {
	return nil, nil, nil
}
