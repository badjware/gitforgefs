package fs

import (
	"context"
	"syscall"

	"github.com/badjware/gitlabfs/gitlab"
	"github.com/hanwen/go-fuse/v2/fs"
)

type RepositoryNode struct {
	fs.Inode
	param   *FSParam
	project *gitlab.Project
}

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReadlinker)((*RepositoryNode)(nil))

func newRepositoryNode(project *gitlab.Project, param *FSParam) (*RepositoryNode, error) {
	node := &RepositoryNode{
		param:   param,
		project: project,
	}
	// Passthrough the error if there is one, nothing to add here
	// Errors on clone/pull are non-fatal
	return node, nil
}

func (n *RepositoryNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	// Create the local copy of the repo
	localRepoLoc, _ := n.param.Git.CloneOrPull(n.project.CloneURL, n.project.ID, n.project.DefaultBranch)

	return []byte(localRepoLoc), 0
}
