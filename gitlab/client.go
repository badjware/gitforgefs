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

type GitlabClient struct {
	Client *gitlab.Client
}

func NewClient(gitlabUrl string, gitlabToken string) (*GitlabClient, error) {
	client, err := gitlab.NewClient(
		gitlabToken,
		gitlab.WithBaseURL(gitlabUrl),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gitlab client: %w", err)
	}

	gitlabClient := &GitlabClient{
		Client: client,
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

func (c GitlabClient) FetchGroup(gid int) (*Group, error) {
	gitlabGroup, _, err := c.Client.Groups.GetGroup(gid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch root group with id %v: %w\n", gid, err)
	}
	group := NewGroupFromGitlabGroup(gitlabGroup)
	return &group, nil
}

func (c GitlabClient) FetchGroupContent(group *Group) (*GroupContent, error) {
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
			Page: 1,
		}}
	for {
		gitlabGroups, response, err := c.Client.Groups.ListSubgroups(group.ID, ListGroupsOpt)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch groups in gitlab: %w", err)
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
			Page: 1,
		}}
	for {
		gitlabProjects, response, err := c.Client.Groups.ListGroupProjects(group.ID, listProjectOpt)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch projects in gitlab: %w", err)
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
