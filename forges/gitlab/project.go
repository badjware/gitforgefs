package gitlab

import (
	"path"

	"github.com/badjware/gitforgefs/config"
	"github.com/xanzy/go-gitlab"
)

type Project struct {
	ID            int
	Path          string
	CloneURL      string
	DefaultBranch string
}

func (p *Project) GetRepositoryID() uint64 {
	return uint64(p.ID)
}

func (p *Project) GetCloneURL() string {
	return p.CloneURL
}

func (p *Project) GetDefaultBranch() string {
	return p.DefaultBranch
}

func (c *gitlabClient) newProjectFromGitlabProject(project *gitlab.Project) *Project {
	// https://godoc.org/github.com/xanzy/go-gitlab#Project
	if c.ArchivedProjectHandling == config.ArchivedProjectIgnore && project.Archived {
		return nil
	}
	p := Project{
		ID:            project.ID,
		Path:          project.Path,
		DefaultBranch: project.DefaultBranch,
	}
	if p.DefaultBranch == "" {
		p.DefaultBranch = "master"
	}
	if c.PullMethod == config.PullMethodSSH {
		p.CloneURL = project.SSHURLToRepo
	} else {
		p.CloneURL = project.HTTPURLToRepo
	}
	if c.ArchivedProjectHandling == config.ArchivedProjectHide && project.Archived {
		p.Path = path.Join(path.Dir(p.Path), "."+path.Base(p.Path))
	}
	return &p
}
