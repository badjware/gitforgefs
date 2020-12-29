package fs

import (
	"context"
	"fmt"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type projectsNode struct {
	fs.Inode
	param *FSParam

	rootGroupIds []int
}

// Ensure we are implementing the NodeOnAdder interface
var _ = (fs.NodeOnAdder)((*projectsNode)(nil))

func NewProjectsNode(rootGroupIds []int, param *FSParam) *projectsNode {
	return &projectsNode{
		param:        param,
		rootGroupIds: rootGroupIds,
	}
}

func (n *projectsNode) OnAdd(ctx context.Context) {
	for _, groupID := range n.rootGroupIds {
		groupNode, err := newRootGroupNode(groupID, n.param)
		if err != nil {
			fmt.Printf("root group fetch fail: %v\n", err)
		}
		inode := n.NewPersistentInode(
			ctx,
			groupNode,
			fs.StableAttr{
				Ino:  <-n.param.staticInoChan,
				Mode: fuse.S_IFDIR,
			},
		)
		n.AddChild(groupNode.group.Path, inode, false)
	}
}
