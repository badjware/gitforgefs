package github

import (
	"path"

	"github.com/badjware/gitlabfs/config"
	"github.com/google/go-github/v63/github"
)

type Repository struct {
	ID            int64
	Path          string
	CloneURL      string
	DefaultBranch string
}

func (r *Repository) GetRepositoryID() uint64 {
	return uint64(r.ID)
}

func (r *Repository) GetCloneURL() string {
	return r.CloneURL
}

func (r *Repository) GetDefaultBranch() string {
	return r.DefaultBranch
}

func (c *githubClient) newRepositoryFromGithubRepository(repository *github.Repository) *Repository {
	if c.ArchivedRepoHandling == config.ArchivedProjectIgnore && *repository.Archived {
		return nil
	}
	r := Repository{
		ID:            *repository.ID,
		Path:          *repository.Name,
		DefaultBranch: *repository.DefaultBranch,
	}
	if r.DefaultBranch == "" {
		r.DefaultBranch = "master"
	}
	if c.PullMethod == config.PullMethodSSH {
		r.CloneURL = *repository.SSHURL
	} else {
		r.CloneURL = *repository.CloneURL
	}
	if c.ArchivedRepoHandling == config.ArchivedProjectHide && *repository.Archived {
		r.Path = path.Join(path.Dir(r.Path), "."+path.Base(r.Path))
	}
	return &r
}
