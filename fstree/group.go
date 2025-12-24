package fstree

import (
	"context"
	"syscall"

	"github.com/badjware/gitforgefs/types"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

const (
	groupBaseInode = 1_000_000_000
)

type groupNode struct {
	fs.Inode
	param *FSParam

	source      types.RepositoryGroupSource
	staticNodes map[string]staticNode
}

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReaddirer)((*groupNode)(nil))

// Ensure we are implementing the NodeLookuper interface
var _ = (fs.NodeLookuper)((*groupNode)(nil))

func newGroupNodeFromSource(ctx context.Context, source types.RepositoryGroupSource, param *FSParam) (fs.InodeEmbedder, error) {
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
	content, err := n.param.Backend.FetchGroupContent(ctx, n.source)
	if err != nil {
		n.param.logger.Error(err.Error())
	}

	entries := make([]fuse.DirEntry, 0, len(content.Groups)+len(content.Repositories)+len(n.staticNodes))
	for groupName, group := range content.Groups {
		entries = append(entries, fuse.DirEntry{
			Name: groupName,
			Ino:  group.GetGroupID() + groupBaseInode,
			Mode: fuse.S_IFDIR,
		})
	}
	for repositoryName, repository := range content.Repositories {
		entries = append(entries, fuse.DirEntry{
			Name: repositoryName,
			Ino:  repository.GetRepositoryID() + repositoryBaseInode,
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
	content, err := n.param.Backend.FetchGroupContent(ctx, n.source)
	if err != nil {
		n.param.logger.Error(err.Error())
	} else {
		// Check if the map of groups contains it
		group, found := content.Groups[name]
		if found {
			attrs := fs.StableAttr{
				Ino:  group.GetGroupID() + groupBaseInode,
				Mode: fuse.S_IFDIR,
			}
			groupNode, _ := newGroupNodeFromSource(ctx, group, n.param)
			return n.NewInode(ctx, groupNode, attrs), 0
		}

		// Check if the map of projects contains it
		repository, found := content.Repositories[name]
		if found {
			attrs := fs.StableAttr{
				Ino: repository.GetRepositoryID() + repositoryBaseInode,
			}
			if n.param.UseSymlinks {
				attrs.Mode = fuse.S_IFLNK
			} else {
				attrs.Mode = fuse.S_IFDIR
			}
			repositoryNode, err := newRepositoryNodeFromSource(ctx, repository, n.param)
			if err != nil {
				n.param.logger.Error(err.Error())
				// TODO: return the proper errno for the error
				return nil, syscall.EIO
			}
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
	}

	return nil, syscall.ENOENT
}
