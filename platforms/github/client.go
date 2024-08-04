package github

import (
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
		// TODO: user + current user

		c.rootContent = rootContent
	}
	return c.rootContent, nil
}

func (c *githubClient) FetchGroupContent(gid uint64) (map[string]fstree.GroupSource, map[string]fstree.RepositorySource, error) {
	if org, found := c.organizationCache[int64(gid)]; found {
		return c.fetchOrganizationContent(org)
	}
	return nil, nil, fmt.Errorf("invalid gid: %v", gid)
}
