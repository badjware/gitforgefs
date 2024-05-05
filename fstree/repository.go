package fstree

import (
	"context"
	"fmt"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
)

type repositoryNode struct {
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
var _ = (fs.NodeReadlinker)((*repositoryNode)(nil))

func newRepositoryNodeFromSource(source RepositorySource, param *FSParam) (*repositoryNode, error) {
	node := &repositoryNode{
		param:  param,
		source: source,
	}
	// Passthrough the error if there is one, nothing to add here
	// Errors on clone/pull are non-fatal
	return node, nil
}

func (n *repositoryNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	// Create the local copy of the repo
	// TODO: cleanup
	localRepositoryPath, err := n.param.GitClient.FetchLocalRepositoryPath(n.source)
	if err != nil {
		fmt.Println(err)
	}
	return []byte(localRepositoryPath), 0
}
