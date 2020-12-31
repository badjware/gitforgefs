package fs

import (
	"context"
	"fmt"
	"syscall"

	"github.com/badjware/gitlabfs/gitlab"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type usersNode struct {
	fs.Inode
	param *FSParam

	userIds []int
}

// Ensure we are implementing the NodeOnAdder interface
var _ = (fs.NodeOnAdder)((*usersNode)(nil))

func newUsersNode(userIds []int, param *FSParam) *usersNode {
	return &usersNode{
		param:   param,
		userIds: userIds,
	}
}

func (n *usersNode) OnAdd(ctx context.Context) {
	// Fetch the current logged user
	currentUser, err := n.param.Gitlab.FetchCurrentUser()
	// Skip if we are anonymous (or the call fails for some reason...)
	if err != nil {
		fmt.Println(err)
	} else {
		currentUserNode, _ := newUserNode(currentUser, n.param)
		inode := n.NewPersistentInode(
			ctx,
			currentUserNode,
			fs.StableAttr{
				Ino:  <-n.param.staticInoChan,
				Mode: fuse.S_IFDIR,
			},
		)
		n.AddChild(currentUserNode.user.Name, inode, false)
	}

	for _, userID := range n.userIds {
		if currentUser != nil && currentUser.ID == userID {
			// We already added the current user, we can skip it
			continue
		}

		userNode, err := newUserNodeByID(userID, n.param)
		if err != nil {
			fmt.Printf("user fetch fail: %v\n", err)
		}
		inode := n.NewPersistentInode(
			ctx,
			userNode,
			fs.StableAttr{
				Ino:  <-n.param.staticInoChan,
				Mode: fuse.S_IFDIR,
			},
		)
		n.AddChild(userNode.user.Name, inode, false)
	}
}

type userNode struct {
	fs.Inode
	param *FSParam

	user        *gitlab.User
	staticNodes map[string]staticNode
}

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReaddirer)((*userNode)(nil))

// Ensure we are implementing the NodeLookuper interface
var _ = (fs.NodeLookuper)((*userNode)(nil))

func newUserNodeByID(uid int, param *FSParam) (*userNode, error) {
	user, err := param.Gitlab.FetchUser(uid)
	if err != nil {
		return nil, err
	}
	node := &userNode{
		param: param,
		user:  user,
		staticNodes: map[string]staticNode{
			".refresh": newRefreshNode(user, param),
		},
	}
	return node, nil
}

func newUserNode(user *gitlab.User, param *FSParam) (*userNode, error) {
	node := &userNode{
		param: param,
		user:  user,
		staticNodes: map[string]staticNode{
			".refresh": newRefreshNode(user, param),
		},
	}
	return node, nil
}

func (n *userNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	userContent, _ := n.param.Gitlab.FetchUserContent(n.user)
	entries := make([]fuse.DirEntry, 0, len(userContent.Projects)+len(n.staticNodes))
	for _, project := range userContent.Projects {
		entries = append(entries, fuse.DirEntry{
			Name: project.Name,
			Ino:  uint64(project.ID),
			Mode: fuse.S_IFLNK,
		})
	}
	for name, staticNode := range n.staticNodes {
		entries = append(entries, fuse.DirEntry{
			Name: name,
			Ino:  staticNode.Ino(),
			Mode: staticNode.Mode(),
		})
	}
	return fs.NewListDirStream(entries), 0
}

func (n *userNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	userContent, _ := n.param.Gitlab.FetchUserContent(n.user)

	// Check if the map of projects contains it
	project, ok := userContent.Projects[name]
	if ok {
		attrs := fs.StableAttr{
			Ino:  uint64(project.ID),
			Mode: fuse.S_IFLNK,
		}
		repositoryNode, _ := newRepositoryNode(project, n.param)
		return n.NewInode(ctx, repositoryNode, attrs), 0
	}

	// Check if the map of static nodes contains it
	staticNode, ok := n.staticNodes[name]
	if ok {
		attrs := fs.StableAttr{
			Ino:  staticNode.Ino(),
			Mode: staticNode.Mode(),
		}
		return n.NewInode(ctx, staticNode, attrs), 0
	}

	return nil, syscall.ENOENT
}
