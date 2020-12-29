package fs

import (
	"context"

	"github.com/hanwen/go-fuse/v2/fs"
)

type usersNode struct {
	fs.Inode
	param *FSParam

	userIds []int
}

// Ensure we are implementing the NodeOnAdder interface
var _ = (fs.NodeOnAdder)((*usersNode)(nil))

func NewUsersNode(userIds []int, param *FSParam) *usersNode {
	return &usersNode{
		param:   param,
		userIds: userIds,
	}
}

func (n *usersNode) OnAdd(ctx context.Context) {
	// for _, userId := range n.userIds {
	// 	userNode, err := newRootUserNode(userId, n.param)
	// 	if err != nil {
	// 		fmt.Printf("user fetch fail: %v\n", err)
	// 	}
	// 	inode := n.NewPersistentInode(
	// 		ctx,
	// 		userNode,
	// 		fs.StableAttr{
	// 			Ino:  <-n.param.staticInoChan,
	// 			Mode: fuse.S_IFDIR,
	// 		},
	// 	)
	// 	n.AddChild(userNode.user.Path, inode, false)
	// }
}
