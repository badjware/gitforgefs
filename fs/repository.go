package fs

import (
	"context"
	"syscall"

	"github.com/badjware/gitlabfs/gitlab"
	"github.com/hanwen/go-fuse/v2/fs"
)

type RepositoryNode struct {
	fs.Inode
	param      *FSParam
	repository *gitlab.Repository
}

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReadlinker)((*RepositoryNode)(nil))

func newRepositoryNode(repository *gitlab.Repository, param *FSParam) (*RepositoryNode, error) {

	node := &RepositoryNode{
		param:      param,
		repository: repository,
	}
	// Passthrough the error if there is one, nothing to add here
	// Errors on clone/pull are non-fatal
	return node, nil
}

func (n *RepositoryNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	// Create the local copy of the repo
	localRepoLoc, _ := n.param.Gcp.CloneOrPull(n.repository.CloneURL, n.repository.ID, "master")

	return []byte(localRepoLoc), 0
}
