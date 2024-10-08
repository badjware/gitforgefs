package gitlab

import (
	"fmt"
	"sync"

	"github.com/badjware/gitforgefs/fstree"
	"github.com/xanzy/go-gitlab"
)

type Group struct {
	ID   int
	Name string

	gitlabClient *gitlabClient

	mux sync.Mutex

	// hold group content
	childGroups   map[string]fstree.GroupSource
	childProjects map[string]fstree.RepositorySource
}

func (g *Group) GetGroupID() uint64 {
	return uint64(g.ID)
}

func (g *Group) InvalidateContentCache() {
	g.mux.Lock()
	defer g.mux.Unlock()

	// clear child group from cache
	g.gitlabClient.groupCacheMux.Lock()
	for _, childGroup := range g.childGroups {
		gid := int(childGroup.GetGroupID())
		delete(g.gitlabClient.groupCache, gid)
	}
	g.gitlabClient.groupCacheMux.Unlock()
	g.childGroups = nil

	// clear child repositories from cache
	g.childGroups = nil
}

func (c *gitlabClient) fetchGroup(gid int) (*Group, error) {
	// start by searching the cache
	// TODO: cache invalidation?
	c.groupCacheMux.RLock()
	group, found := c.groupCache[gid]
	c.groupCacheMux.RUnlock()
	if found {
		c.logger.Debug("Group cache hit", "gid", gid)
		return group, nil
	} else {
		c.logger.Debug("Group cache miss; fetching group", "gid", gid)
	}

	// If not in cache, fetch group infos from API
	gitlabGroup, _, err := c.client.Groups.GetGroup(gid, &gitlab.GetGroupOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group with id %v: %v", gid, err)
	}
	c.logger.Debug("Fetched group", "gid", gid)
	newGroup := Group{
		ID:   gitlabGroup.ID,
		Name: gitlabGroup.Path,

		gitlabClient: c,

		childGroups:   nil,
		childProjects: nil,
	}

	// save in cache
	c.groupCacheMux.Lock()
	c.groupCache[gid] = &newGroup
	c.groupCacheMux.Unlock()

	return &newGroup, nil
}

func (c *gitlabClient) newGroupFromGitlabGroup(gitlabGroup *gitlab.Group) (*Group, error) {
	gid := gitlabGroup.ID

	// start by searching the cache
	c.groupCacheMux.RLock()
	group, found := c.groupCache[gid]
	c.groupCacheMux.RUnlock()
	if found {
		// if found in cache, return the cached reference
		c.logger.Debug("Group cache hit", "gid", gid)
		return group, nil
	} else {
		c.logger.Debug("Group cache miss; registering group", "gid", gid)
	}

	// if not found in cache, convert and save to cache now
	newGroup := Group{
		ID:   gitlabGroup.ID,
		Name: gitlabGroup.Path,

		gitlabClient: c,

		childGroups:   nil,
		childProjects: nil,
	}

	// save in cache
	c.groupCacheMux.Lock()
	c.groupCache[gid] = &newGroup
	c.groupCacheMux.Unlock()

	return &newGroup, nil
}

func (c *gitlabClient) fetchGroupContent(group *Group) (map[string]fstree.GroupSource, map[string]fstree.RepositorySource, error) {
	// Only a single routine can fetch the group content at the time.
	// We lock for the whole duration of the function to avoid fetching the same data from the API
	// multiple times if concurrent calls where to occur.
	group.mux.Lock()
	defer group.mux.Unlock()

	// Get cached data if available
	// TODO: cache cache invalidation?
	if group.childGroups == nil || group.childProjects == nil {
		childGroups := make(map[string]fstree.GroupSource)
		childProjects := make(map[string]fstree.RepositorySource)

		// List subgroups in path
		listGroupsOpt := &gitlab.ListSubGroupsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    1,
				PerPage: 100,
			},
			AllAvailable: gitlab.Ptr(true),
		}
		for {
			gitlabGroups, response, err := c.client.Groups.ListSubGroups(group.ID, listGroupsOpt)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to fetch groups in gitlab: %v", err)
			}
			for _, gitlabGroup := range gitlabGroups {
				group, _ := c.newGroupFromGitlabGroup(gitlabGroup)
				childGroups[group.Name] = group
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
			gitlabProjects, response, err := c.client.Groups.ListGroupProjects(group.ID, listProjectOpt)
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

		group.childGroups = childGroups
		group.childProjects = childProjects
	}
	return group.childGroups, group.childProjects, nil
}
