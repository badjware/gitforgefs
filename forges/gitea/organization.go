package gitea

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/badjware/gitforgefs/types"
)

type Organization struct {
	ID   int64
	Name string
}

func (o *Organization) GetGroupID() uint64 {
	return uint64(o.ID)
}

func (o *Organization) GetGroupName() string {
	return o.Name
}

func (o *Organization) GetGroupPath() string {
	return o.Name
}

func (c *giteaClient) fetchOrganization(ctx context.Context, orgName string) (*Organization, error) {
	giteaOrg, _, err := c.client.GetOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organization with name %v: %v", orgName, err)
	}
	newOrg := Organization{
		ID:   giteaOrg.ID,
		Name: giteaOrg.UserName,
	}

	return &newOrg, nil
}

func (c *giteaClient) fetchOrganizationContent(ctx context.Context, orgName string) (types.RepositoryGroupContent, error) {
	org, err := c.fetchOrganization(ctx, orgName)
	if err != nil {
		return types.RepositoryGroupContent{}, err
	}

	repositories := make(map[string]types.RepositorySource)

	// Fetch the organization repositories
	listReposOptions := gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{PageSize: 100},
	}
	for {
		giteaRepositories, response, err := c.client.ListOrgRepos(org.Name, gitea.ListOrgReposOptions(listReposOptions))
		if err != nil {
			return types.RepositoryGroupContent{}, fmt.Errorf("failed to fetch repository in gitea: %v", err)
		}
		for _, giteaRepository := range giteaRepositories {
			repository := c.newRepositoryFromGiteaRepository(giteaRepository)
			if repository != nil {
				repositories[repository.GetRepositoryName()] = repository
			}
		}
		if response.NextPage == 0 {
			break
		}
		// Get the next page
		listReposOptions.Page = response.NextPage
	}

	return types.RepositoryGroupContent{
		Groups:       make(map[string]types.RepositoryGroupSource),
		Repositories: repositories,
	}, nil
}
