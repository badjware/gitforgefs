package github

import (
	"context"
	"log/slog"

	"github.com/badjware/gitforgefs/config"
	"github.com/badjware/gitforgefs/types"
	"github.com/google/go-github/v63/github"
)

type githubClient struct {
	config.GithubClientConfig
	client *github.Client

	logger *slog.Logger

	// use a map without values for efficient lookups
	users map[string]struct{}
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

		users: make(map[string]struct{}),
	}

	// Fetch current user and add it to the list
	currentUser, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		logger.Warn("failed to fetch the current user:", "error", err.Error())
	} else {
		gitHubClient.UserNames = append(gitHubClient.UserNames, *currentUser.Login)
	}

	return gitHubClient, nil
}

func (c *githubClient) FetchRootGroupContent(ctx context.Context) (map[string]types.GroupSource, error) {
	rootContent := make(map[string]types.GroupSource)

	for _, orgName := range c.GithubClientConfig.OrgNames {
		org, err := c.fetchOrganization(ctx, orgName)
		if err != nil {
			c.logger.Warn(err.Error())
		} else {
			rootContent[org.Name] = org
		}
	}

	for _, userName := range c.GithubClientConfig.UserNames {
		user, err := c.fetchUser(ctx, userName)
		if err != nil {
			c.logger.Warn(err.Error())
		} else {
			rootContent[user.Name] = user
			c.users[user.Name] = struct{}{}
		}
	}

	return rootContent, nil
}

func (c *githubClient) FetchGroupContent(ctx context.Context, source types.GroupSource) (types.GroupContent, error) {
	if _, found := c.users[source.GetGroupPath()]; found {
		return c.fetchUserContent(ctx, source.GetGroupPath())
	} else {
		return c.fetchOrganizationContent(ctx, source.GetGroupPath())
	}
}
