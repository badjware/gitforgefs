package fstree

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

const (
	staticInodeStart = uint64(int(^(uint(0))>>1)) + 1
)

type staticNode interface {
	fs.InodeEmbedder
	Ino() uint64
	Mode() uint32
}

type GitClient interface {
	FetchLocalRepositoryPath(source RepositorySource) (string, error)
}

type GitPlatform interface {
	FetchRootGroupContent() (map[string]GroupSource, error)
	FetchGroupContent(gid uint64) (map[string]GroupSource, map[string]RepositorySource, error)
}

type FSParam struct {
	GitClient   GitClient
	GitPlatform GitPlatform

	staticInoChan chan uint64
}

type rootNode struct {
	fs.Inode
	param *FSParam
}

var _ = (fs.NodeOnAdder)((*rootNode)(nil))

func Start(mountpoint string, mountoptions []string, param *FSParam, debug bool) error {
	fmt.Printf("Mounting in %v\n", mountpoint)

	opts := &fs.Options{}
	opts.MountOptions.Options = mountoptions
	opts.Debug = debug

	param.staticInoChan = make(chan uint64)
	root := &rootNode{
		param: param,
	}

	go staticInoGenerator(root.param.staticInoChan)

	server, err := fs.Mount(mountpoint, root, opts)
	if err != nil {
		return fmt.Errorf("mount failed: %v", err)
	}

	signalChan := make(chan os.Signal)
	go signalHandler(signalChan, server)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// server.Serve() is already called in fs.Mount() so we shouldn't call it ourself. We wait for the server to terminate.
	server.Wait()

	return nil
}

func (n *rootNode) OnAdd(ctx context.Context) {
	rootGroups, err := n.param.GitPlatform.FetchRootGroupContent()
	if err != nil {
		panic(err)
	}

	for groupName, group := range rootGroups {
		groupNode, _ := newGroupNodeFromSource(group, n.param)
		persistentInode := n.NewPersistentInode(
			ctx,
			groupNode,
			fs.StableAttr{
				Ino:  <-n.param.staticInoChan,
				Mode: fuse.S_IFDIR,
			},
		)
		n.AddChild(groupName, persistentInode, false)
	}

	fmt.Println("Mounted and ready to use")
}

func staticInoGenerator(staticInoChan chan<- uint64) {
	i := staticInodeStart
	for {
		staticInoChan <- i
		i++
	}
}

func signalHandler(signalChan <-chan os.Signal, server *fuse.Server) {
	err := server.WaitMount()
	if err != nil {
		fmt.Printf("failed to start exit signal handler: %v\n", err)
		return
	}
	for {
		s := <-signalChan
		fmt.Printf("Caught %v: stopping\n", s)
		err := server.Unmount()
		if err != nil {
			fmt.Printf("Failed to unmount: %v\n", err)
		}
	}
}
