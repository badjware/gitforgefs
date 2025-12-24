package gitea

import (
	"context"
	"fmt"
	"log/slog"

	"code.gitea.io/sdk/gitea"
	"github.com/badjware/gitforgefs/config"
	"github.com/badjware/gitforgefs/types"
)

type giteaClient struct {
	config.GiteaClientConfig
	client *gitea.Client

	logger *slog.Logger

	// use a map without values for efficient lookups
	users map[string]struct{}
}

func NewClient(logger *slog.Logger, config config.GiteaClientConfig) (*giteaClient, error) {
	client, err := gitea.NewClient(config.URL, gitea.SetToken(config.Token))
	if err != nil {
		return nil, fmt.Errorf("failed to create the gitea client: %v", err)
	}

	giteaClient := &giteaClient{
		GiteaClientConfig: config,
		client:            client,

		logger: logger,

		users: make(map[string]struct{}),
	}

	// Fetch current user and add it to the list
	currentUser, _, err := client.GetMyUserInfo()
	if err != nil {
		logger.Warn("failed to fetch the current user:", "error", err.Error())
	} else {
		giteaClient.UserNames = append(giteaClient.UserNames, currentUser.UserName)
	}

	return giteaClient, nil
}

func (c *giteaClient) FetchRootGroupContent(ctx context.Context) (map[string]types.RepositoryGroupSource, error) {
	rootContent := make(map[string]types.RepositoryGroupSource)

	for _, orgName := range c.GiteaClientConfig.OrgNames {
		org, err := c.fetchOrganization(ctx, orgName)
		if err != nil {
			c.logger.Warn(err.Error())
		} else {
			rootContent[org.Name] = org
		}
	}

	for _, userName := range c.GiteaClientConfig.UserNames {
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

func (c *giteaClient) FetchGroupContent(ctx context.Context, source types.RepositoryGroupSource) (types.RepositoryGroupContent, error) {
	if _, found := c.users[source.GetGroupPath()]; found {
		return c.fetchUserContent(ctx, source.GetGroupPath())
	} else {
		return c.fetchOrganizationContent(ctx, source.GetGroupPath())
	}
}
