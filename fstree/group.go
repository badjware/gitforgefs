package fstree

import (
	"context"
	"sync/atomic"
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
	gen         atomic.Uint64
}

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReaddirer)((*groupNode)(nil))

// Ensure we are implementing the NodeLookuper interface
var _ = (fs.NodeLookuper)((*groupNode)(nil))

// Ensure we are implementing the NodeGetattrer interface
var _ = (fs.NodeGetattrer)((*groupNode)(nil))

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
			Mode: fuse.S_IFDIR,
			Ino:  group.GetGroupID() + groupBaseInode,
		})
	}
	for repositoryName, repository := range content.Repositories {
		if n.param.UseSymlinks {
			entries = append(entries, fuse.DirEntry{
				Name: repositoryName,
				Mode: fuse.S_IFLNK,
				Ino:  repository.GetRepositoryID() + repositoryBaseInode,
			})
		} else {
			entries = append(entries, fuse.DirEntry{
				Name: repositoryName,
				Mode: fuse.S_IFDIR,
				Ino:  repository.GetRepositoryID() + repositoryBaseInode,
			})
		}

	}
	for name, staticNode := range n.staticNodes {
		entries = append(entries, fuse.DirEntry{
			Name: name,
			Mode: staticNode.Mode(),
		})
	}
	return fs.NewListDirStream(entries), 0
}

func (n *groupNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	content, err := n.param.Backend.FetchGroupContent(ctx, n.source)
	if err != nil {
		n.param.logger.Error(err.Error())
		return nil, syscall.EIO
	} else {
		// Check if the map of groups contains it
		group, found := content.Groups[name]
		if found {
			attrs := fs.StableAttr{
				Mode: fuse.S_IFDIR,
				Ino:  group.GetGroupID() + groupBaseInode,
			}
			out.Atime = uint64(group.GetLastModified().Unix())
			out.Mtime = uint64(group.GetLastModified().Unix())
			out.Ctime = uint64(group.GetLastModified().Unix())

			groupNode, _ := newGroupNodeFromSource(ctx, group, n.param)
			return n.NewInode(ctx, groupNode, attrs), 0
		}

		// Check if the map of projects contains it
		repository, found := content.Repositories[name]
		if found {
			attrs := fs.StableAttr{
				Ino: repository.GetRepositoryID() + repositoryBaseInode,
			}
			out.Atime = uint64(repository.GetLastModified().Unix())
			out.Mtime = uint64(repository.GetLastModified().Unix())
			out.Ctime = uint64(repository.GetLastModified().Unix())

			if n.param.UseSymlinks {
				attrs.Mode = fuse.S_IFLNK
			} else {
				attrs.Mode = fuse.S_IFDIR

				// fetch and mirror attrs from loopback
				var st syscall.Stat_t
				localRepositoryPath, err := n.param.GitClient.FetchLocalRepositoryPath(ctx, repository)
				if err != nil {
					n.param.logger.Error(err.Error())
					if ctx.Err() != nil {
						return nil, syscall.EINTR
					} else {
						// TODO: return the proper errno for the error
						return nil, syscall.EIO
					}
				}
				if err := syscall.Stat(localRepositoryPath, &st); err != nil {
					n.param.logger.Warn("Failed to stat", "path", localRepositoryPath)
					return nil, syscall.EIO
				}
				out.Size = uint64(st.Size)
				out.Blocks = uint64(st.Blocks)
				// wrong ownership may trigger "detected dubious ownership in repository" in git
				out.Uid = st.Uid
				out.Gid = st.Gid
				out.Blksize = uint32(st.Blksize)

				// Set gen as a workaround for ino collisions when using loopback nodes. See
				// https://github.com/hanwen/go-fuse/issues/592#issuecomment-3650851207
				// attrs.Gen = n.gen.Add(1)
			}

			repositoryNode, err := newRepositoryNodeFromSource(ctx, repository, n.param)
			if err != nil {
				n.param.logger.Error(err.Error())
				if ctx.Err() != nil {
					return nil, syscall.EINTR
				} else {
					// TODO: return the proper errno for the error
					return nil, syscall.EIO
				}
			}
			return n.NewInode(ctx, repositoryNode, attrs), 0
		}

		// Check if the map of static nodes contains it
		staticNode, ok := n.staticNodes[name]
		if ok {
			attrs := fs.StableAttr{
				Mode: staticNode.Mode(),
			}
			return n.NewInode(ctx, staticNode, attrs), 0
		}
	}

	return nil, syscall.ENOENT
}

func (n *groupNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Atime = uint64(n.source.GetLastModified().Unix())
	out.Mtime = uint64(n.source.GetLastModified().Unix())
	out.Ctime = uint64(n.source.GetLastModified().Unix())
	return 0
}
