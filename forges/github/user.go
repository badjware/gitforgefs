package github

import (
	"context"
	"fmt"
	"sync"

	"github.com/badjware/gitforgefs/fstree"
	"github.com/google/go-github/v63/github"
)

type User struct {
	ID   int64
	Name string

	mux sync.Mutex

	// hold user content
	childRepositories map[string]fstree.RepositorySource
}

func (u *User) GetGroupID() uint64 {
	return uint64(u.ID)
}

func (u *User) InvalidateContentCache() {
	u.mux.Lock()
	defer u.mux.Unlock()

	// clear child repositories from cache
	u.childRepositories = nil
}

func (c *githubClient) fetchUser(userName string) (*User, error) {
	c.userCacheMux.RLock()
	cachedId, found := c.userNameToIDMap[userName]
	if found {
		cachedUser := c.userCache[cachedId]
		c.userCacheMux.RUnlock()

		// if found in cache, return the cached reference
		c.logger.Debug("User cache hit", "user_name", userName)
		return cachedUser, nil
	} else {
		c.userCacheMux.RUnlock()

		c.logger.Debug("User cache miss", "user_name", userName)
	}

	// If not found in cache, fetch user infos from API
	githubUser, _, err := c.client.Users.Get(context.Background(), userName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user with name %v: %v", userName, err)
	}
	newUser := User{
		ID:   *githubUser.ID,
		Name: *githubUser.Login,

		childRepositories: nil,
	}

	// save in cache
	c.userCacheMux.Lock()
	c.userCache[newUser.ID] = &newUser
	c.userNameToIDMap[newUser.Name] = newUser.ID
	c.userCacheMux.Unlock()

	return &newUser, nil
}

func (c *githubClient) fetchUserContent(user *User) (map[string]fstree.GroupSource, map[string]fstree.RepositorySource, error) {
	user.mux.Lock()
	defer user.mux.Unlock()

	// Get cached data if available
	// TODO: cache cache invalidation?
	if user.childRepositories == nil {
		childRepositories := make(map[string]fstree.RepositorySource)

		// Fetch the user repositories
		repositoryListOpt := &github.RepositoryListByUserOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		}
		for {
			githubRepositories, response, err := c.client.Repositories.ListByUser(context.Background(), user.Name, repositoryListOpt)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to fetch repository in github: %v", err)
			}
			for _, githubRepository := range githubRepositories {
				repository := c.newRepositoryFromGithubRepository(githubRepository)
				if repository != nil {
					childRepositories[repository.Path] = repository
				}
			}
			if response.NextPage == 0 {
				break
			}
			// Get the next page
			repositoryListOpt.Page = response.NextPage
		}

		user.childRepositories = childRepositories
	}
	return make(map[string]fstree.GroupSource), user.childRepositories, nil
}
