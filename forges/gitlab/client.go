package gitlab

import (
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/badjware/gitforgefs/config"
	"github.com/badjware/gitforgefs/fstree"
	"github.com/xanzy/go-gitlab"
)

type gitlabClient struct {
	config.GitlabClientConfig
	client *gitlab.Client

	logger *slog.Logger

	rootContent map[string]fstree.GroupSource

	// API response cache
	groupCacheMux sync.RWMutex
	groupCache    map[int]*Group
	userCacheMux  sync.RWMutex
	userCache     map[int]*User
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

		rootContent: nil,

		groupCache: map[int]*Group{},
		userCache:  map[int]*User{},
	}

	// Fetch current user and add it to the list
	currentUser, _, err := client.Users.CurrentUser()
	if err != nil {
		logger.Warn("failed to fetch the current user:", "error", err.Error())
	} else {
		gitlabClient.UserIDs = append(gitlabClient.UserIDs, currentUser.ID)
	}

	return gitlabClient, nil
}

func (c *gitlabClient) FetchRootGroupContent() (map[string]fstree.GroupSource, error) {
	// use cached values if available
	if c.rootContent == nil {
		rootGroupCache := make(map[string]fstree.GroupSource)

		// fetch root groups
		for _, gid := range c.GroupIDs {
			group, err := c.fetchGroup(gid)
			if err != nil {
				return nil, err
			}
			rootGroupCache[group.Name] = group
		}
		// fetch users
		for _, uid := range c.UserIDs {
			user, err := c.fetchUser(uid)
			if err != nil {
				return nil, err
			}
			rootGroupCache[user.Name] = user
		}

		c.rootContent = rootGroupCache
	}
	return c.rootContent, nil
}

func (c *gitlabClient) FetchGroupContent(gid uint64) (map[string]fstree.GroupSource, map[string]fstree.RepositorySource, error) {
	if slices.Contains[[]int, int](c.UserIDs, int(gid)) {
		// gid is a user
		user, err := c.fetchUser(int(gid))
		if err != nil {
			return nil, nil, err
		}
		return c.fetchUserContent(user)
	} else {
		// gid is a group
		group, err := c.fetchGroup(int(gid))
		if err != nil {
			return nil, nil, err
		}
		return c.fetchGroupContent(group)
	}
}
