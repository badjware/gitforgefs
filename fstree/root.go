package fstree

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/badjware/gitforgefs/types"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type staticNode interface {
	fs.InodeEmbedder
	Mode() uint32
}

type FSParam struct {
	UseSymlinks bool

	GitClient types.GitClient
	Backend   types.GitForgeCacher

	logger *slog.Logger
}

type rootNode struct {
	fs.Inode
	param *FSParam
}

var _ = (fs.NodeOnAdder)((*rootNode)(nil))

func Start(logger *slog.Logger, mountpoint string, mountoptions []string, param *FSParam, debug bool) error {
	logger.Info("Mounting", "mountpoint", mountpoint)

	opts := &fs.Options{}
	opts.MountOptions.Options = mountoptions
	opts.Debug = debug

	param.logger = logger
	root := &rootNode{
		param: param,
	}

	server, err := fs.Mount(mountpoint, root, opts)
	if err != nil {
		return fmt.Errorf("mount failed: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	go signalHandler(logger, signalChan, server)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// server.Serve() is already called in fs.Mount() so we shouldn't call it ourself. We wait for the server to terminate.
	server.Wait()

	return nil
}

func (n *rootNode) OnAdd(ctx context.Context) {
	rootGroups, err := n.param.Backend.FetchRootGroupContent(ctx)
	if err != nil {
		panic(err)
	}

	for groupName, group := range rootGroups {
		groupNode, _ := newGroupNodeFromSource(ctx, group, n.param)
		persistentInode := n.NewPersistentInode(
			ctx,
			groupNode,
			fs.StableAttr{
				Mode: fuse.S_IFDIR,
			},
		)
		n.AddChild(groupName, persistentInode, false)
	}

	n.param.logger.Info("Mounted and ready to use")
}

func signalHandler(logger *slog.Logger, signalChan <-chan os.Signal, server *fuse.Server) {
	err := server.WaitMount()
	if err != nil {
		logger.Error("failed to start exit signal handler", "error", err)
		return
	}
	for {
		s := <-signalChan
		logger.Info("Caught signal", "signal", s)
		err := server.Unmount()
		if err != nil {
			logger.Error("Failed to unmount", "error", err)
		}
	}
}
