package fs

import (
	"context"
	"strconv"
	"syscall"

	"github.com/badjware/gitlabfs/gitlab"
	"github.com/hanwen/go-fuse/v2/fs"
)

type RepositoryNode struct {
	fs.Inode
	repository *gitlab.Repository
}

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReadlinker)((*RepositoryNode)(nil))

func newRepositoryNode(repository *gitlab.Repository) (*RepositoryNode, error) {
	node := &RepositoryNode{
		repository: repository,
	}
	return node, nil
}

func (n *RepositoryNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	// TODO
	return []byte(strconv.Itoa(n.repository.ID)), 0
}
