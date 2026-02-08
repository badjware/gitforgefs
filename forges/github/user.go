package github

import (
	"context"
	"fmt"
	"time"

	"github.com/badjware/gitforgefs/types"
	"github.com/google/go-github/v63/github"
)

type User struct {
	ID           int64
	Name         string
	LastModified time.Time
}

func (u *User) GetGroupID() uint64 {
	return uint64(u.ID)
}

func (u *User) GetGroupName() string {
	return u.Name
}

func (u *User) GetGroupPath() string {
	return u.Name
}

func (u *User) GetLastModified() time.Time {
	return u.LastModified
}

func (c *githubClient) fetchUser(ctx context.Context, userName string) (*User, error) {
	githubUser, _, err := c.client.Users.Get(ctx, userName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user with name %v: %v", userName, err)
	}
	return &User{
		ID:           *githubUser.ID,
		Name:         *githubUser.Login,
		LastModified: githubUser.UpdatedAt.Time,
	}, nil
}

func (c *githubClient) fetchUserContent(ctx context.Context, userName string) (types.RepositoryGroupContent, error) {
	user, err := c.fetchUser(ctx, userName)
	if err != nil {
		return types.RepositoryGroupContent{}, err
	}

	repositories := make(map[string]types.RepositorySource)

	// Fetch the user repositories
	repositoryListOpt := &github.RepositoryListByUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		githubRepositories, response, err := c.client.Repositories.ListByUser(ctx, user.Name, repositoryListOpt)
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
