package gitea

import (
	"path"

	"code.gitea.io/sdk/gitea"
	"github.com/badjware/gitforgefs/config"
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

func (c *giteaClient) newRepositoryFromGiteaRepository(repository *gitea.Repository) *Repository {
	if c.ArchivedRepoHandling == config.ArchivedProjectIgnore && repository.Archived {
		return nil
	}
	r := Repository{
		ID:            repository.ID,
		Path:          repository.Name,
		DefaultBranch: repository.DefaultBranch,
	}
	if r.DefaultBranch == "" {
		r.DefaultBranch = "master"
	}
	if c.PullMethod == config.PullMethodSSH {
		r.CloneURL = repository.SSHURL
	} else {
		r.CloneURL = repository.CloneURL
	}
	if c.ArchivedRepoHandling == config.ArchivedProjectHide && repository.Archived {
		r.Path = path.Join(path.Dir(r.Path), "."+path.Base(r.Path))
	}
	return &r
}
