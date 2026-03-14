package fstree

import (
	"context"
	"fmt"
	"syscall"

	"github.com/badjware/gitforgefs/types"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

const (
	repositoryBaseInode = 2_000_000_000
)

type repositorySymlinkNode struct {
	fs.Inode
	param *FSParam

	source types.RepositorySource
}

// type repositoryLoopbackNode struct {
// 	*fs.LoopbackNode
// 	param *FSParam

// 	source types.RepositorySource
// }

// Ensure we are implementing the NodeReadlinker interface
var _ = (fs.NodeReadlinker)((*repositorySymlinkNode)(nil))

// Ensure we are implementing the NodeGetattrer interface
var _ = (fs.NodeGetattrer)((*repositorySymlinkNode)(nil))

// Ensure we are implementing the NodeGetattrer interface
// var _ = (fs.NodeWrapChilder)((*repositoryLoopbackNode)(nil))

func newRepositoryNodeFromSource(ctx context.Context, source types.RepositorySource, param *FSParam) (fs.InodeEmbedder, error) {
	if param.UseSymlinks {
		return &repositorySymlinkNode{
			param:  param,
			source: source,
		}, nil
	} else {
		localRepositoryPath, err := param.GitClient.FetchLocalRepositoryPath(ctx, source)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch local repository path: %w", err)
		}
		// var st syscall.Stat_t
		// err := syscall.Stat(localRepositoryPath, &st)
		// rootData := &fs.LoopbackRoot{
		// 	Path: localRepositoryPath,
		// 	Dev:  st.Dev,
		// }
		// rootNode := &fs.LoopbackNode{
		// 	RootData: rootData,
		// }
		// rootData.RootNode = rootNode
		// return rootNode, nil
		return fs.NewLoopbackRoot(localRepositoryPath)
	}
}

func (n *repositorySymlinkNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	// Create the local copy of the repo
	// TODO: cleanup
	localRepositoryPath, err := n.param.GitClient.FetchLocalRepositoryPath(ctx, n.source)
	if err != nil {
		n.param.logger.Error(err.Error())
		if ctx.Err() != nil {
			return nil, syscall.EINTR
		} else {
			// TODO: return the proper errno for the error
			return nil, syscall.EIO
		}
	}
	return []byte(localRepositoryPath), 0
}

func (n *repositorySymlinkNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Atime = uint64(n.source.GetLastModified().Unix())
	out.Mtime = uint64(n.source.GetLastModified().Unix())
	out.Ctime = uint64(n.source.GetLastModified().Unix())
	return 0
}

// func (n *repositoryLoopbackNode) WrapChild(ctx context.Context, ops fs.InodeEmbedder) fs.InodeEmbedder {
// 	return &repositoryLoopbackNode{
// 		LoopbackNode: ops.(*fs.LoopbackNode),

// 		param:  n.param,
// 		source: n.source,
// 	}
// }
