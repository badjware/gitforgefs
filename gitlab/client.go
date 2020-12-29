package gitlab

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
)

type GroupFetcher interface {
	FetchGroup(gid int) (*Group, error)
	FetchGroupContent(group *Group) (*GroupContent, error)
}

type GroupContent struct {
	Groups       map[string]*Group
	Repositories map[string]*Repository
}

type Group struct {
	ID      int
	Name    string
	Path    string
	Content *GroupContent
}

type Repository struct {
	ID       int
	Name     string
	Path     string
	CloneURL string
}

type GitlabClientParam struct {
}

type gitlabClient struct {
	GitlabClientParam
	client *gitlab.Client
}

func NewClient(gitlabUrl string, gitlabToken string, p GitlabClientParam) (*gitlabClient, error) {
	client, err := gitlab.NewClient(
		gitlabToken,
		gitlab.WithBaseURL(gitlabUrl),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gitlab client: %v", err)
	}

	gitlabClient := &gitlabClient{
		client: client,
	}
	return gitlabClient, nil
}

func NewRepositoryFromGitlabProject(project *gitlab.Project) Repository {
	// https://godoc.org/github.com/xanzy/go-gitlab#Project
	return Repository{
		ID:       project.ID,
		Name:     project.Name,
		Path:     project.Path,
		CloneURL: project.HTTPURLToRepo,
		// CloneUrl: project.SSHURLToRepo,
	}
}

func NewGroupFromGitlabGroup(group *gitlab.Group) Group {
	// https://godoc.org/github.com/xanzy/go-gitlab#Group
	return Group{
		ID:   group.ID,
		Name: group.Name,
		Path: group.Path,
	}
}

func (c gitlabClient) FetchGroup(gid int) (*Group, error) {
	gitlabGroup, _, err := c.client.Groups.GetGroup(gid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group with id %v: %v\n", gid, err)
	}
	group := NewGroupFromGitlabGroup(gitlabGroup)
	return &group, nil
}

func (c gitlabClient) FetchGroupContent(group *Group) (*GroupContent, error) {
	if group.Content != nil {
		return group.Content, nil
	}

	content := &GroupContent{
		Groups:       map[string]*Group{},
		Repositories: map[string]*Repository{},
	}

	// List subgroups in path
	ListGroupsOpt := &gitlab.ListSubgroupsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 1000,
		}}
	for {
		gitlabGroups, response, err := c.client.Groups.ListSubgroups(group.ID, ListGroupsOpt)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch groups in gitlab: %v", err)
		}
		for _, gitlabGroup := range gitlabGroups {
			group := NewGroupFromGitlabGroup(gitlabGroup)
			content.Groups[group.Path] = &group
		}
		if response.CurrentPage >= response.TotalPages {
			break
		}
		// Get the next page
		ListGroupsOpt.Page = response.NextPage
	}

	// List repositories in path
	listProjectOpt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 1000,
		}}
	for {
		gitlabProjects, response, err := c.client.Groups.ListGroupProjects(group.ID, listProjectOpt)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch projects in gitlab: %v", err)
		}
		for _, gitlabProject := range gitlabProjects {
			repository := NewRepositoryFromGitlabProject(gitlabProject)
			content.Repositories[repository.Path] = &repository
		}
		if response.CurrentPage >= response.TotalPages {
			break
		}
		// Get the next page
		listProjectOpt.Page = response.NextPage
	}

	group.Content = content
	return content, nil
}
