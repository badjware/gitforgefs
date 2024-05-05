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

	// user content cache
	projectCache map[string]fstree.RepositorySource
}

func (u *User) GetGroupID() uint64 {
	return uint64(u.ID)
}

func (u *User) InvalidateCache() {
	u.mux.Lock()
	defer u.mux.Unlock()

	u.projectCache = nil
}

func (c *gitlabClient) fetchUser(uid int) (*User, error) {
	// start by searching the cache
	// TODO: cache invalidation?
	user, found := c.userCache[uid]
	if found {
		return user, nil
	}

	// If not in cache, fetch group infos from API
	gitlabUser, _, err := c.client.Users.GetUser(uid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user with id %v: %v", uid, err)
	}
	newUser := User{
		ID:   gitlabUser.ID,
		Name: gitlabUser.Username,

		projectCache: nil,
	}

	// save in cache
	c.userCache[uid] = &newUser

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

			projectCache: nil,
		}
		c.currentUserCache = &newUser
	}
	return c.currentUserCache, nil
}

func (c *gitlabClient) fetchUserContent(user *User) (map[string]fstree.GroupSource, map[string]fstree.RepositorySource, error) {
	user.mux.Lock()
	defer user.mux.Unlock()

	// Get cached data if available
	// TODO: cache cache invalidation?
	if user.projectCache == nil {
		projectCache := make(map[string]fstree.RepositorySource)

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
				projectCache[project.Name] = &project
			}
			if response.CurrentPage >= response.TotalPages {
				break
			}
			// Get the next page
			listProjectOpt.Page = response.NextPage
		}

		user.projectCache = projectCache
	}
	return make(map[string]fstree.GroupSource), user.projectCache, nil
}
