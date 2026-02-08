package github

import (
	"context"
	"fmt"
	"time"

	"github.com/badjware/gitforgefs/types"
	"github.com/google/go-github/v63/github"
)

type Organization struct {
	ID           int64
	Name         string
	LastModified time.Time
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

func (o *Organization) GetLastModified() time.Time {
	return o.LastModified
}

func (c *githubClient) fetchOrganization(ctx context.Context, orgName string) (*Organization, error) {
	githubOrg, _, err := c.client.Organizations.Get(ctx, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organization with name %v: %v", orgName, err)
	}
	return &Organization{
		ID:           *githubOrg.ID,
		Name:         *githubOrg.Login,
		LastModified: githubOrg.UpdatedAt.Time,
	}, nil
}

func (c *githubClient) fetchOrganizationContent(ctx context.Context, orgName string) (types.RepositoryGroupContent, error) {
	org, err := c.fetchOrganization(ctx, orgName)
	if err != nil {
		return types.RepositoryGroupContent{}, err
	}

	repositories := make(map[string]types.RepositorySource)

	// Fetch the organization repositories
	repositoryListOpt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		githubRepositories, response, err := c.client.Repositories.ListByOrg(ctx, org.Name, repositoryListOpt)
		if err != nil {
			return types.RepositoryGroupContent{}, fmt.Errorf("failed to fetch repository in github: %v", err)
		}
		for _, githubRepository := range githubRepositories {
			repository := c.newRepositoryFromGithubRepository(githubRepository)
			if repository != nil {
				repositories[repository.GetRepositoryName()] = repository
			}
		}
		if response.NextPage == 0 {
			break
		}
		// Get the next page
		repositoryListOpt.Page = response.NextPage
	}

	return types.RepositoryGroupContent{
		Groups:       make(map[string]types.RepositoryGroupSource),
		Repositories: repositories,
	}, nil
}
