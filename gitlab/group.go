package gitlab

import (
	"fmt"
	"sync"

	"github.com/xanzy/go-gitlab"
)

type GroupFetcher interface {
	FetchGroup(gid int) (*Group, error)
	FetchGroupContent(group *Group) (*GroupContent, error)
}

type GroupContent struct {
	Groups   map[string]*Group
	Projects map[string]*Project
}

type Group struct {
	ID   int
	Name string

	mux     sync.Mutex
	content *GroupContent
}

func NewGroupFromGitlabGroup(group *gitlab.Group) Group {
	// https://godoc.org/github.com/xanzy/go-gitlab#Group
	return Group{
		ID:   group.ID,
		Name: group.Path,
	}
}

func (g *Group) InvalidateCache() {
	g.mux.Lock()
	defer g.mux.Unlock()

	g.content = nil
}

func (c *gitlabClient) FetchGroup(gid int) (*Group, error) {
	gitlabGroup, _, err := c.client.Groups.GetGroup(gid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group with id %v: %v", gid, err)
	}
	group := NewGroupFromGitlabGroup(gitlabGroup)
	return &group, nil
}

func (c *gitlabClient) FetchGroupContent(group *Group) (*GroupContent, error) {
	group.mux.Lock()
	defer group.mux.Unlock()

	// Get cached data if available
	if group.content != nil {
		return group.content, nil
	}

	content := &GroupContent{
		Groups:   map[string]*Group{},
		Projects: map[string]*Project{},
	}

	// List subgroups in path
	ListGroupsOpt := &gitlab.ListSubgroupsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
		AllAvailable: gitlab.Bool(true),
	}
	for {
		gitlabGroups, response, err := c.client.Groups.ListSubgroups(group.ID, ListGroupsOpt)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch groups in gitlab: %v", err)
		}
		for _, gitlabGroup := range gitlabGroups {
			group := NewGroupFromGitlabGroup(gitlabGroup)
			content.Groups[group.Name] = &group
		}
		if response.CurrentPage >= response.TotalPages {
			break
		}
		// Get the next page
		ListGroupsOpt.Page = response.NextPage
	}

	// List projects in path
	listProjectOpt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		}}
	for {
		gitlabProjects, response, err := c.client.Groups.ListGroupProjects(group.ID, listProjectOpt)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch projects in gitlab: %v", err)
		}
		for _, gitlabProject := range gitlabProjects {
			project := c.newProjectFromGitlabProject(gitlabProject)
			content.Projects[project.Name] = &project
		}
		if response.CurrentPage >= response.TotalPages {
			break
		}
		// Get the next page
		listProjectOpt.Page = response.NextPage
	}

	group.content = content
	return content, nil
}
