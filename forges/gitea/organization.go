package gitea

import (
	"fmt"
	"sync"

	"code.gitea.io/sdk/gitea"
	"github.com/badjware/gitforgefs/fstree"
)

type Organization struct {
	ID   int64
	Name string

	mux sync.Mutex

	// hold org content
	childRepositories map[string]fstree.RepositorySource
}

func (o *Organization) GetGroupID() uint64 {
	return uint64(o.ID)
}

func (o *Organization) InvalidateContentCache() {
	o.mux.Lock()
	defer o.mux.Unlock()

	// clear child repositories from cache
	o.childRepositories = nil
}

func (c *giteaClient) fetchOrganization(orgName string) (*Organization, error) {
	c.organizationCacheMux.RLock()
	cachedId, found := c.organizationNameToIDMap[orgName]
	if found {
		cachedOrg := c.organizationCache[cachedId]
		c.organizationCacheMux.RUnlock()

		// if found in cache, return the cached reference
		c.logger.Debug("Organization cache hit", "org_name", orgName)
		return cachedOrg, nil
	} else {
		c.organizationCacheMux.RUnlock()

		c.logger.Debug("Organization cache miss", "org_name", orgName)
	}

	// If not found in cache, fetch organization infos from API
	giteaOrg, _, err := c.client.GetOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organization with name %v: %v", orgName, err)
	}
	newOrg := Organization{
		ID:   giteaOrg.ID,
		Name: giteaOrg.UserName,

		childRepositories: nil,
	}

	// save in cache
	c.organizationCacheMux.Lock()
	c.organizationCache[newOrg.ID] = &newOrg
	c.organizationNameToIDMap[newOrg.Name] = newOrg.ID
	c.organizationCacheMux.Unlock()

	return &newOrg, nil
}

func (c *giteaClient) fetchOrganizationContent(org *Organization) (map[string]fstree.GroupSource, map[string]fstree.RepositorySource, error) {
	org.mux.Lock()
	defer org.mux.Unlock()

	// Get cached data if available
	// TODO: cache cache invalidation?
	if org.childRepositories == nil {
		childRepositories := make(map[string]fstree.RepositorySource)

		// Fetch the organization repositories
		listReposOptions := gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{PageSize: 100},
		}
		for {
			giteaRepositories, response, err := c.client.ListOrgRepos(org.Name, gitea.ListOrgReposOptions(listReposOptions))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to fetch repository in gitea: %v", err)
			}
			for _, giteaRepository := range giteaRepositories {
				repository := c.newRepositoryFromGiteaRepository(giteaRepository)
				if repository != nil {
					childRepositories[repository.Path] = repository
				}
			}
			if response.NextPage == 0 {
				break
			}
			// Get the next page
			listReposOptions.Page = response.NextPage
		}

		org.childRepositories = childRepositories
	}
	return make(map[string]fstree.GroupSource), org.childRepositories, nil
}
