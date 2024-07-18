package gitlab

import (
	"fmt"
	"sync"

	"github.com/badjware/gitlabfs/fstree"
	"github.com/xanzy/go-gitlab"
)

type User struct {
	ID   int
	Name string

	mux sync.Mutex

	// hold user content
	childProjects map[string]fstree.RepositorySource
}

func (u *User) GetGroupID() uint64 {
	return uint64(u.ID)
}

func (u *User) InvalidateContentCache() {
	u.mux.Lock()
	defer u.mux.Unlock()

	// clear child repositories from cache
	u.childProjects = nil
}

func (c *gitlabClient) fetchUser(uid int) (*User, error) {
	// start by searching the cache
	// TODO: cache invalidation?
	c.userCacheMux.RLock()
	user, found := c.userCache[uid]
	c.userCacheMux.RUnlock()
	if found {
		// if found in cache, return the cached reference
		c.logger.Debug("User cache hit", "uid", uid)
		return user, nil
	} else {
		c.logger.Debug("User cache miss", "uid", uid)
	}

	// If not found in cache, fetch group infos from API
	gitlabUser, _, err := c.client.Users.GetUser(uid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user with id %v: %v", uid, err)
	}
	newUser := User{
		ID:   gitlabUser.ID,
		Name: gitlabUser.Username,

		childProjects: nil,
	}

	// save in cache
	c.userCache[uid] = &newUser
	c.userCacheMux.Unlock()

	return &newUser, nil
}

func (c *gitlabClient) fetchCurrentUser() (*User, error) {
	if c.currentUserCache == nil {
		gitlabUser, _, err := c.client.Users.CurrentUser()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch current user: %v", err)
		}
		newUser := User{
			ID:   gitlabUser.ID,
			Name: gitlabUser.Username,

			childProjects: nil,
		}
		c.currentUserCache = &newUser
	}
	return c.currentUserCache, nil
}

func (c *gitlabClient) fetchUserContent(user *User) (map[string]fstree.GroupSource, map[string]fstree.RepositorySource, error) {
	// Only a single routine can fetch the user content at the time.
	// We lock for the whole duration of the function to avoid fetching the same data from the API
	// multiple times if concurrent calls where to occur.
	user.mux.Lock()
	defer user.mux.Unlock()

	// Get cached data if available
	// TODO: cache cache invalidation?
	if user.childProjects == nil {
		childProjects := make(map[string]fstree.RepositorySource)

		// Fetch the user repositories
		listProjectOpt := &gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    1,
				PerPage: 100,
			}}
		for {
			gitlabProjects, response, err := c.client.Projects.ListUserProjects(user.ID, listProjectOpt)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to fetch projects in gitlab: %v", err)
			}
			for _, gitlabProject := range gitlabProjects {
				project := c.newProjectFromGitlabProject(gitlabProject)
				if project != nil {
					childProjects[project.Path] = project
				}
			}
			if response.CurrentPage >= response.TotalPages {
				break
			}
			// Get the next page
			listProjectOpt.Page = response.NextPage
		}

		user.childProjects = childProjects
	}
	return make(map[string]fstree.GroupSource), user.childProjects, nil
}
