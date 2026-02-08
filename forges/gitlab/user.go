package gitlab

import (
	"context"
	"fmt"
	"time"

	"github.com/badjware/gitforgefs/types"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type User struct {
	ID           int
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

func (c *gitlabClient) fetchUser(ctx context.Context, uid int) (*User, error) {
	gitlabUser, _, err := c.client.Users.GetUser(uid, gitlab.GetUsersOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user with id %v: %v", uid, err)
	}
	lastModified := time.Time{}
	if gitlabUser.CreatedAt != nil {
		lastModified = *gitlabUser.CreatedAt
	}
	return &User{
		ID:           gitlabUser.ID,
		Name:         gitlabUser.Username,
		LastModified: lastModified,
	}, nil
}

func (c *gitlabClient) fetchUserContent(ctx context.Context, uid int) (types.RepositoryGroupContent, error) {
	childProjects := make(map[string]types.RepositorySource)

	// Fetch the user repositories
	listProjectOpt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		}}
	for {
		gitlabProjects, response, err := c.client.Projects.ListUserProjects(uid, listProjectOpt)
		if err != nil {
			return types.RepositoryGroupContent{}, fmt.Errorf("failed to fetch projects in gitlab: %v", err)
		}
		for _, gitlabProject := range gitlabProjects {
			project := c.newProjectFromGitlabProject(gitlabProject)
			if project != nil {
				childProjects[project.GetRepositoryName()] = project
			}
		}
		if response.CurrentPage >= response.TotalPages {
			break
		}
		// Get the next page
		listProjectOpt.Page = response.NextPage
	}
	return types.RepositoryGroupContent{
		Groups:       make(map[string]types.RepositoryGroupSource),
		Repositories: childProjects,
	}, nil
}
