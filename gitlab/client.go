package gitlab

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
)

type GroupContentFetcher interface {
	FetchGroupContent(path string) (GroupContent, error)
}

type GroupContent struct {
	Repositories []Repository
	Groups       []Group
}

type Repository struct {
	ID       int
	Name     string
	Path     string
	CloneURL string
}

type Group struct {
	ID   int
	Name string
	Path string
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

func NewRepositryFromGitlabProject(project *gitlab.Project) Repository {
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

func (g GitlabClient) FetchGroupContent(group *Group) (*GroupContent, error) {
	content := &GroupContent{}

	// List repositories in path
	listProjectOpt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page: 1,
		}}
	for {
		projects, response, err := g.Client.Groups.ListGroupProjects(group.ID, listProjectOpt)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch projects in gitlab: %w", err)
		}
		for _, project := range projects {
			content.Repositories = append(content.Repositories, NewRepositryFromGitlabProject(project))
		}
		if response.CurrentPage >= response.TotalPages {
			break
		}
		// Get the next page
		listProjectOpt.Page = response.NextPage
	}

	// List subgroups in path
	ListGroupsOpt := &gitlab.ListSubgroupsOptions{
		ListOptions: gitlab.ListOptions{
			Page: 1,
		}}
	for {
		groups, response, err := g.Client.Groups.ListSubgroups(group.ID, ListGroupsOpt)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch groups in gitlab: %w", err)
		}
		for _, group := range groups {
			content.Groups = append(content.Groups, NewGroupFromGitlabGroup(group))
		}
		if response.CurrentPage >= response.TotalPages {
			break
		}
		// Get the next page
		ListGroupsOpt.Page = response.NextPage
	}

	return content, nil
}
