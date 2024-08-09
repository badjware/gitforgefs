package fstree

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type refreshNode struct {
	fs.Inode
	ino uint64

	source GroupSource
}

// Ensure we are implementing the NodeSetattrer interface
var _ = (fs.NodeSetattrer)((*refreshNode)(nil))

// Ensure we are implementing the NodeOpener interface
var _ = (fs.NodeOpener)((*refreshNode)(nil))

func newRefreshNode(source GroupSource, param *FSParam) *refreshNode {
	return &refreshNode{
		ino:    0,
		source: source,
	}
}

func (n *refreshNode) Ino() uint64 {
	return n.ino
}

func (n *refreshNode) Mode() uint32 {
	return fuse.S_IFREG
}

func (n *refreshNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	return 0
}

func (n *refreshNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	n.source.InvalidateContentCache()
	return nil, 0, 0
}
