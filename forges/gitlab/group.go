package gitlab

import (
	"context"
	"fmt"

	"github.com/badjware/gitforgefs/types"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Group struct {
	ID   int
	Name string
	Path string
}

func (g *Group) GetGroupID() uint64 {
	return uint64(g.ID)
}

func (g *Group) GetGroupName() string {
	return g.Name
}

func (g *Group) GetGroupPath() string {
	return g.Path
}

func (c *gitlabClient) newGroupFromGitlabGroup(gitlabGroup *gitlab.Group) *Group {
	return &Group{
		ID:   gitlabGroup.ID,
		Name: gitlabGroup.Path,
		Path: gitlabGroup.FullPath,
	}
}

func (c *gitlabClient) fetchGroup(ctx context.Context, gid uint64) (*Group, error) {
	gitlabGroup, _, err := c.client.Groups.GetGroup(gid, &gitlab.GetGroupOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group with id %v: %v", gid, err)
	}
	c.logger.Debug("Fetched group", "gid", gid)
	return c.newGroupFromGitlabGroup(gitlabGroup), nil
}

func (c *gitlabClient) fetchGroupContent(ctx context.Context, gid uint64) (types.GroupContent, error) {
	childGroups := make(map[string]types.GroupSource)
	childProjects := make(map[string]types.RepositorySource)

	// List subgroups in path
	listGroupsOpt := &gitlab.ListSubGroupsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
		AllAvailable: gitlab.Ptr(true),
	}
	for {
		gitlabGroups, response, err := c.client.Groups.ListSubGroups(gid, listGroupsOpt)
		if err != nil {
			return types.GroupContent{}, fmt.Errorf("failed to fetch groups in gitlab: %v", err)
		}
		for _, gitlabGroup := range gitlabGroups {
			group := c.newGroupFromGitlabGroup(gitlabGroup)
			if group != nil {
				childGroups[group.GetGroupName()] = group
			}
		}
		if response.CurrentPage >= response.TotalPages {
			break
		}
		// Get the next page
		listGroupsOpt.Page = response.NextPage
	}

	// List projects in path
	listProjectOpt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		}}
	for {
		gitlabProjects, response, err := c.client.Groups.ListGroupProjects(gid, listProjectOpt)
		if err != nil {
			return types.GroupContent{}, fmt.Errorf("failed to fetch projects in gitlab: %v", err)
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
	return types.GroupContent{
		Groups:       childGroups,
		Repositories: childProjects,
	}, nil
}
