package fstree

import (
	"context"
	"fmt"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type groupNode struct {
	fs.Inode
	param *FSParam

	source      GroupSource
	staticNodes map[string]staticNode
}

type GroupSource interface {
	GetGroupID() uint64
	InvalidateCache()
}

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReaddirer)((*groupNode)(nil))

// Ensure we are implementing the NodeLookuper interface
var _ = (fs.NodeLookuper)((*groupNode)(nil))

func newGroupNodeFromSource(source GroupSource, param *FSParam) (*groupNode, error) {
	node := &groupNode{
		param:  param,
		source: source,
		staticNodes: map[string]staticNode{
			".refresh": newRefreshNode(source, param),
		},
	}
	return node, nil
}

func (n *groupNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	groups, repositories, err := n.param.GitPlatform.FetchGroupContent(n.source.GetGroupID())
	if err != nil {
		fmt.Errorf("%v", err)
	}

	entries := make([]fuse.DirEntry, 0, len(groups)+len(repositories)+len(n.staticNodes))
	for groupName, group := range groups {
		entries = append(entries, fuse.DirEntry{
			Name: groupName,
			Ino:  group.GetGroupID(),
			Mode: fuse.S_IFDIR,
		})
	}
	for repositoryName, repository := range repositories {
		entries = append(entries, fuse.DirEntry{
			Name: repositoryName,
			Ino:  repository.GetRepositoryID(),
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

func (n *groupNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	groups, repositories, _ := n.param.GitPlatform.FetchGroupContent(n.source.GetGroupID())

	// Check if the map of groups contains it
	group, found := groups[name]
	if found {
		attrs := fs.StableAttr{
			Ino:  group.GetGroupID(),
			Mode: fuse.S_IFDIR,
		}
		groupNode, _ := newGroupNodeFromSource(group, n.param)
		return n.NewInode(ctx, groupNode, attrs), 0
	}

	// Check if the map of projects contains it
	repository, found := repositories[name]
	if found {
		attrs := fs.StableAttr{
			Ino:  repository.GetRepositoryID(),
			Mode: fuse.S_IFLNK,
		}
		repositoryNode, _ := newRepositoryNodeFromSource(repository, n.param)
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
