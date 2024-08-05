package github

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/badjware/gitlabfs/config"
	"github.com/badjware/gitlabfs/fstree"
	"github.com/google/go-github/v63/github"
)

type githubClient struct {
	config.GithubClientConfig
	client *github.Client

	logger *slog.Logger

	rootContent map[string]fstree.GroupSource

	// API response cache
	organizationCacheMux    sync.RWMutex
	organizationNameToIDMap map[string]int64
	organizationCache       map[int64]*Organization
	userCacheMux            sync.RWMutex
	userNameToIDMap         map[string]int64
	userCache               map[int64]*User
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

		rootContent: nil,

		organizationNameToIDMap: map[string]int64{},
		organizationCache:       map[int64]*Organization{},
		userNameToIDMap:         map[string]int64{},
		userCache:               map[int64]*User{},
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

func (c *githubClient) FetchRootGroupContent() (map[string]fstree.GroupSource, error) {
	if c.rootContent == nil {
		rootContent := make(map[string]fstree.GroupSource)

		for _, org_name := range c.GithubClientConfig.OrgNames {
			org, err := c.fetchOrganization(org_name)
			if err != nil {
				c.logger.Warn(err.Error())
			} else {
				rootContent[org.Name] = org
			}
		}

		for _, user_name := range c.GithubClientConfig.UserNames {
			user, err := c.fetchUser(user_name)
			if err != nil {
				c.logger.Warn(err.Error())
			} else {
				rootContent[user.Name] = user
			}
		}

		c.rootContent = rootContent
	}
	return c.rootContent, nil
}

func (c *githubClient) FetchGroupContent(gid uint64) (map[string]fstree.GroupSource, map[string]fstree.RepositorySource, error) {
	if org, found := c.organizationCache[int64(gid)]; found {
		return c.fetchOrganizationContent(org)
	}
	if user, found := c.userCache[int64(gid)]; found {
		return c.fetchUserContent(user)
	}
	return nil, nil, fmt.Errorf("invalid gid: %v", gid)
}
