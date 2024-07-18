package gitlab

import (
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/badjware/gitlabfs/fstree"
	"github.com/xanzy/go-gitlab"
)

const (
	PullMethodHTTP = "http"
	PullMethodSSH  = "ssh"

	ArchivedProjectShow   = "show"
	ArchivedProjectHide   = "hide"
	ArchivedProjectIgnore = "ignore"
)

type GitlabClientConfig struct {
	URL                     string `yaml:"url,omitempty"`
	Token                   string `yaml:"token,omitempty"`
	GroupIDs                []int  `yaml:"group_ids,omitempty"`
	UserIDs                 []int  `yaml:"user_ids,omitempty"`
	ArchivedProjectHandling string `yaml:"archived_project_handling,omitempty"`
	IncludeCurrentUser      bool   `yaml:"include_current_user,omitempty"`
	PullMethod              string `yaml:"pull_method,omitempty"`
}

type gitlabClient struct {
	GitlabClientConfig
	client *gitlab.Client

	logger *slog.Logger

	// root group cache
	rootGroupCache   map[string]fstree.GroupSource
	currentUserCache *User

	// API response cache
	groupCacheMux sync.RWMutex
	groupCache    map[int]*Group
	userCacheMux  sync.RWMutex
	userCache     map[int]*User
}

func NewClient(logger *slog.Logger, gitlabUrl string, gitlabToken string, p GitlabClientConfig) (*gitlabClient, error) {
	client, err := gitlab.NewClient(
		gitlabToken,
		gitlab.WithBaseURL(gitlabUrl),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gitlab client: %v", err)
	}

	gitlabClient := &gitlabClient{
		GitlabClientConfig: p,
		client:             client,

		logger: logger,

		rootGroupCache:   nil,
		currentUserCache: nil,

		groupCache: map[int]*Group{},
		userCache:  map[int]*User{},
	}
	return gitlabClient, nil
}

func (c *gitlabClient) FetchRootGroupContent() (map[string]fstree.GroupSource, error) {
	// use cached values if available
	if c.rootGroupCache == nil {
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
		// fetch current user if configured
		if c.IncludeCurrentUser {
			currentUser, err := c.fetchCurrentUser()
			if err != nil {
				c.logger.Warn(err.Error())
			} else {
				rootGroupCache[currentUser.Name] = currentUser
			}
		}

		c.rootGroupCache = rootGroupCache
	}
	return c.rootGroupCache, nil
}

func (c *gitlabClient) FetchGroupContent(gid uint64) (map[string]fstree.GroupSource, map[string]fstree.RepositorySource, error) {
	if slices.Contains[[]int, int](c.UserIDs, int(gid)) || (c.currentUserCache != nil && c.currentUserCache.ID == int(gid)) {
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
