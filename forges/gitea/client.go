package gitea

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"code.gitea.io/sdk/gitea"
	"github.com/badjware/gitforgefs/config"
	"github.com/badjware/gitforgefs/fstree"
)

type giteaClient struct {
	config.GiteaClientConfig
	client *gitea.Client

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

func NewClient(logger *slog.Logger, config config.GiteaClientConfig) (*giteaClient, error) {
	client, err := gitea.NewClient(config.URL, gitea.SetToken(config.Token))
	if err != nil {
		return nil, fmt.Errorf("failed to create the gitea client: %v", err)
	}

	giteaClient := &giteaClient{
		GiteaClientConfig: config,
		client:            client,

		logger: logger,

		rootContent: nil,

		organizationNameToIDMap: map[string]int64{},
		organizationCache:       map[int64]*Organization{},
		userNameToIDMap:         map[string]int64{},
		userCache:               map[int64]*User{},
	}

	// Fetch current user and add it to the list
	currentUser, _, err := client.GetMyUserInfo()
	if err != nil {
		logger.Warn("failed to fetch the current user:", "error", err.Error())
	} else {
		giteaClient.UserNames = append(giteaClient.UserNames, *&currentUser.UserName)
	}

	return giteaClient, nil
}

func (c *giteaClient) FetchRootGroupContent(ctx context.Context) (map[string]fstree.GroupSource, error) {
	if c.rootContent == nil {
		rootContent := make(map[string]fstree.GroupSource)

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
			}
		}

		c.rootContent = rootContent
	}
	return c.rootContent, nil
}

func (c *giteaClient) FetchGroupContent(ctx context.Context, gid uint64) (map[string]fstree.GroupSource, map[string]fstree.RepositorySource, error) {
	if org, found := c.organizationCache[int64(gid)]; found {
		return c.fetchOrganizationContent(ctx, org)
	}
	if user, found := c.userCache[int64(gid)]; found {
		return c.fetchUserContent(ctx, user)
	}
	return nil, nil, fmt.Errorf("invalid gid: %v", gid)
}
