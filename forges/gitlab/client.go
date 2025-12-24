package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/badjware/gitforgefs/config"
	"github.com/badjware/gitforgefs/types"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type gitlabClient struct {
	config.GitlabClientConfig
	client *gitlab.Client

	logger *slog.Logger

	// use a map without values for efficient lookups
	users map[string]int
}

func NewClient(logger *slog.Logger, config config.GitlabClientConfig) (*gitlabClient, error) {
	client, err := gitlab.NewClient(
		config.Token,
		gitlab.WithBaseURL(config.URL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gitlab client: %v", err)
	}

	gitlabClient := &gitlabClient{
		GitlabClientConfig: config,
		client:             client,

		logger: logger,

		users: make(map[string]int),
	}

	// Fetch current user and add it to the list
	currentUser, _, err := client.Users.CurrentUser()
	if err != nil {
		logger.Warn("failed to fetch the current user:", "error", err.Error())
	} else {
		gitlabClient.users[currentUser.Username] = currentUser.ID
	}

	// Fetch the configured users and add them to the list
	for _, userName := range config.UserNames {
		user, _, err := client.Users.ListUsers(&gitlab.ListUsersOptions{Username: &userName})
		if err != nil || len(user) != 1 {
			logger.Warn("failed to fetch the user", "userName", userName, "error", err.Error())
		} else {
			gitlabClient.users[userName] = user[0].ID
		}
	}

	return gitlabClient, nil
}

func (c *gitlabClient) FetchRootGroupContent(ctx context.Context) (map[string]types.GroupSource, error) {
	rootContent := make(map[string]types.GroupSource)

	// fetch root groups
	for _, gid := range c.GroupIDs {
		group, err := c.fetchGroup(ctx, gid)
		if err != nil {
			return nil, err
		}
		rootContent[group.Name] = group
	}
	// fetch users
	for _, uid := range c.users {
		user, err := c.fetchUser(ctx, uid)
		if err != nil {
			return nil, err
		}
		rootContent[user.Name] = user
	}
	return rootContent, nil
}

func (c *gitlabClient) FetchGroupContent(ctx context.Context, source types.GroupSource) (types.GroupContent, error) {
	if _, found := c.users[source.GetGroupPath()]; found {
		return c.fetchUserContent(ctx, source.GetGroupID())
	} else {
		return c.fetchGroupContent(ctx, source.GetGroupID())
	}
}
