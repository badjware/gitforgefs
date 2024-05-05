package fs

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
)

type RepositoryNode struct {
	fs.Inode
	param *FSParam

	source RepositorySource
}

type RepositorySource interface {
	// GetName() string
	GetRepositoryID() uint64
	GetCloneURL() string
	GetDefaultBranch() string
}

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReadlinker)((*RepositoryNode)(nil))

func newRepositoryNodeFromSource(source RepositorySource, param *FSParam) (*RepositoryNode, error) {
	node := &RepositoryNode{
		param:  param,
		source: source,
	}
	// Passthrough the error if there is one, nothing to add here
	// Errors on clone/pull are non-fatal
	return node, nil
}

func (n *RepositoryNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	// Create the local copy of the repo
	// TODO: cleanup
	localRepositoryPath, _ := n.param.GitImplementation.CloneOrPull(n.source.GetCloneURL(), int(n.source.GetRepositoryID()), n.source.GetDefaultBranch())

	return []byte(localRepositoryPath), 0
}
