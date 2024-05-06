package gitlab

import (
	"fmt"
	"sync"

	"github.com/badjware/gitlabfs/fstree"
	"github.com/xanzy/go-gitlab"
)

type Group struct {
	ID   int
	Name string

	mux sync.Mutex

	// group content cache
	groupCache   map[string]fstree.GroupSource
	projectCache map[string]fstree.RepositorySource
}

func (g *Group) GetGroupID() uint64 {
	return uint64(g.ID)
}

func (g *Group) InvalidateCache() {
	g.mux.Lock()
	defer g.mux.Unlock()

	g.groupCache = nil
	g.projectCache = nil
}

func (c *gitlabClient) fetchGroup(gid int) (*Group, error) {
	// start by searching the cache
	// TODO: cache invalidation?
	group, found := c.groupCache[gid]
	if found {
		return group, nil
	}

	// If not in cache, fetch group infos from API
	gitlabGroup, _, err := c.client.Groups.GetGroup(gid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group with id %v: %v", gid, err)
	}
	newGroup := Group{
		ID:   gitlabGroup.ID,
		Name: gitlabGroup.Path,

		groupCache:   nil,
		projectCache: nil,
	}

	// save in cache
	c.groupCache[gid] = &newGroup

	return &newGroup, nil
}

func (c *gitlabClient) fetchGroupContent(group *Group) (map[string]fstree.GroupSource, map[string]fstree.RepositorySource, error) {
	group.mux.Lock()
	defer group.mux.Unlock()

	// Get cached data if available
	// TODO: cache cache invalidation?
	if group.groupCache == nil || group.projectCache == nil {
		groupCache := make(map[string]fstree.GroupSource)
		projectCache := make(map[string]fstree.RepositorySource)

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
				return nil, nil, fmt.Errorf("failed to fetch groups in gitlab: %v", err)
			}
			for _, gitlabGroup := range gitlabGroups {
				group, _ := c.fetchGroup(gitlabGroup.ID)
				groupCache[group.Name] = group
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

		group.groupCache = groupCache
		group.projectCache = projectCache
	}
	return group.groupCache, group.projectCache, nil
}