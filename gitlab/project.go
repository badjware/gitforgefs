package gitlab

import "github.com/xanzy/go-gitlab"

type Project struct {
	ID       int
	Name     string
	CloneURL string
}

func NewProjectFromGitlabProject(project *gitlab.Project) Project {
	// https://godoc.org/github.com/xanzy/go-gitlab#Project
	return Project{
		ID:   project.ID,
		Name: project.Path,
		// CloneURL: project.HTTPURLToRepo,
		CloneURL: project.SSHURLToRepo,
	}
}
