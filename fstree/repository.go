package fstree

import (
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

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

// Ensure we are implementing the NodeReadlinker interface
var _ = (fs.NodeReadlinker)((*repositorySymlinkNode)(nil))

// Ensure we are implementing the NodeGetattrer interface
var _ = (fs.NodeGetattrer)((*repositorySymlinkNode)(nil))

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
		// The path must exist to successfully create a loopback. We poll the filesystem until its created.
		// This of course add latency, maybe we should think of a way of mitigating it in the future.
		// We do not care in the case of a symlink. A symlink pointing on nothing is still a valid symlink.
		for ctx.Err() == nil {
			_, err := os.Stat(localRepositoryPath)
			if err == nil {
				return fs.NewLoopbackRoot(localRepositoryPath)
			} else if errors.Is(err, os.ErrNotExist) {
				// wait for the file to be created
				// TODO: think of a more efficient way of archiving this
				time.Sleep(100 * time.Millisecond)
			} else {
				// error, filesystem
				return nil, fmt.Errorf("error while waiting for the local repository to be created: %w", err)
			}
		}
		// error, context cancelled
		return nil, fmt.Errorf("context cancelled while waiting for the local repository to be created: %w", ctx.Err())
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
	out.Mtime = uint64(n.source.GetLastModified().Unix())
	return 0
}
