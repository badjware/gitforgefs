package gitea

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/badjware/gitforgefs/types"
)

type User struct {
	ID   int64
	Name string
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

func (c *giteaClient) fetchUser(ctx context.Context, userName string) (*User, error) {
	giteaUser, _, err := c.client.GetUserInfo(userName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user with name %v: %v", userName, err)
	}
	newUser := User{
		ID:   giteaUser.ID,
		Name: giteaUser.UserName,
	}

	return &newUser, nil
}

func (c *giteaClient) fetchUserContent(ctx context.Context, userName string) (types.RepositoryGroupContent, error) {
	user, err := c.fetchUser(ctx, userName)
	if err != nil {
		return types.RepositoryGroupContent{}, err
	}

	repositories := make(map[string]types.RepositorySource)

	// Fetch the user repositories
	listReposOptions := gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{PageSize: 100},
	}
	for {
		giteaRepositories, response, err := c.client.ListUserRepos(user.Name, listReposOptions)
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
