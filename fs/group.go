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
	param *FSParam

	group       *gitlab.Group
	staticNodes map[string]staticNode
}

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReaddirer)((*groupNode)(nil))

// Ensure we are implementing the NodeLookuper interface
var _ = (fs.NodeLookuper)((*groupNode)(nil))

func newGroupNodeByID(gid int, param *FSParam) (*groupNode, error) {
	group, err := param.Gitlab.FetchGroup(gid)
	if err != nil {
		return nil, err
	}
	node := &groupNode{
		param: param,
		group: group,
		staticNodes: map[string]staticNode{
			".refresh": newRefreshNode(group, param),
		},
	}
	return node, nil
}

func newGroupNode(group *gitlab.Group, param *FSParam) (*groupNode, error) {
	node := &groupNode{
		param: param,
		group: group,
		staticNodes: map[string]staticNode{
			".refresh": newRefreshNode(group, param),
		},
	}
	return node, nil
}

func (n *groupNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	groupContent, _ := n.param.Gitlab.FetchGroupContent(n.group)
	entries := make([]fuse.DirEntry, 0, len(groupContent.Groups)+len(groupContent.Projects)+len(n.staticNodes))
	for _, group := range groupContent.Groups {
		entries = append(entries, fuse.DirEntry{
			Name: group.Name,
			Ino:  uint64(group.ID),
			Mode: fuse.S_IFDIR,
		})
	}
	for _, project := range groupContent.Projects {
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

func (n *groupNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	groupContent, _ := n.param.Gitlab.FetchGroupContent(n.group)

	// Check if the map of groups contains it
	group, ok := groupContent.Groups[name]
	if ok {
		attrs := fs.StableAttr{
			Ino:  uint64(group.ID),
			Mode: fuse.S_IFDIR,
		}
		groupNode, _ := newGroupNode(group, n.param)
		return n.NewInode(ctx, groupNode, attrs), 0
	}

	// Check if the map of projects contains it
	project, ok := groupContent.Projects[name]
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
