package fs

import (
	"context"
	"syscall"

	"github.com/badjware/gitlabfs/gitlab"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type groupNode struct {
	fs.Inode
	group *gitlab.Group
	gf    gitlab.GroupFetcher
}

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReaddirer)((*groupNode)(nil))

// Ensure we are implementing the NodeLookuper interface
var _ = (fs.NodeLookuper)((*groupNode)(nil))

func newRootGroupNode(gf gitlab.GroupFetcher, gid int) (*groupNode, error) {
	group, err := gf.FetchGroup(gid)
	if err != nil {
		return nil, err
	}
	node := &groupNode{
		group: group,
		gf:    gf,
	}
	return node, nil
}

func newGroupNode(gf gitlab.GroupFetcher, group *gitlab.Group) (*groupNode, error) {
	node := &groupNode{
		group: group,
		gf:    gf,
	}
	return node, nil
}

func (n *groupNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	groupContent, _ := n.gf.FetchGroupContent(n.group)
	entries := make([]fuse.DirEntry, 0, len(groupContent.Groups)+len(groupContent.Repositories))
	for _, group := range groupContent.Groups {
		entries = append(entries, fuse.DirEntry{
			Name: group.Path,
			Ino:  uint64(group.ID),
			Mode: fuse.S_IFDIR,
		})
	}
	for _, repository := range groupContent.Repositories {
		entries = append(entries, fuse.DirEntry{
			Name: repository.Path,
			Ino:  uint64(repository.ID),
			Mode: fuse.S_IFLNK,
		})
	}
	return fs.NewListDirStream(entries), 0
}

func (n *groupNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	groupContent, _ := n.gf.FetchGroupContent(n.group)

	// Check if the map of groups contains it
	group, ok := groupContent.Groups[name]
	if ok {
		attrs := fs.StableAttr{
			Ino:  uint64(group.ID),
			Mode: fuse.S_IFDIR,
		}
		groupNode, _ := newGroupNode(n.gf, group)
		return n.NewInode(ctx, groupNode, attrs), 0
	}

	// Check if the map of repositories contains it
	repository, ok := groupContent.Repositories[name]
	if ok {
		attrs := fs.StableAttr{
			Ino:  uint64(repository.ID),
			Mode: fuse.S_IFLNK,
		}
		repositoryNode, _ := newRepositoryNode(repository) // TODO
		return n.NewInode(ctx, repositoryNode, attrs), 0
	}
	return nil, syscall.ENOENT
}
