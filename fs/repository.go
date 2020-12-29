package fs

import (
	"context"
	"syscall"

	"github.com/badjware/gitlabfs/git"
	"github.com/badjware/gitlabfs/gitlab"
	"github.com/hanwen/go-fuse/v2/fs"
)

type RepositoryNode struct {
	fs.Inode
	repository *gitlab.Repository

	gcp git.GitClonerPuller
}

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReadlinker)((*RepositoryNode)(nil))

func newRepositoryNode(gcp git.GitClonerPuller, repository *gitlab.Repository) (*RepositoryNode, error) {

	node := &RepositoryNode{
		repository: repository,
		gcp:        gcp,
	}
	// Passthrough the error if there is one, nothing to add here
	// Errors on clone/pull are non-fatal
	return node, nil
}

func (n *RepositoryNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	// Create the local copy of the repo
	localRepoLoc, _ := n.gcp.CloneOrPull(n.repository.CloneURL, n.repository.ID, "master")

	return []byte(localRepoLoc), 0
}
