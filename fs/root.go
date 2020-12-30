package fs

import (
	"context"
	"fmt"

	"github.com/badjware/gitlabfs/git"
	"github.com/badjware/gitlabfs/gitlab"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

const (
	staticInodeStart = uint64(int(^(uint(0))>>1)) + 1
)

type FSParam struct {
	Gitlab gitlab.GitlabFetcher
	Git    git.GitClonerPuller

	staticInoChan chan uint64
}

type rootNode struct {
	fs.Inode
	param        *FSParam
	rootGroupIds []int
	userIds      []int
}

var _ = (fs.NodeOnAdder)((*rootNode)(nil))

func (n *rootNode) OnAdd(ctx context.Context) {
	projectsInode := n.NewPersistentInode(
		ctx,
		newProjectsNode(
			n.rootGroupIds,
			n.param,
		),
		fs.StableAttr{
			Ino:  <-n.param.staticInoChan,
			Mode: fuse.S_IFDIR,
		},
	)
	n.AddChild("projects", projectsInode, false)

	usersInode := n.NewPersistentInode(
		ctx,
		newUsersNode(
			n.userIds,
			n.param,
		),
		fs.StableAttr{
			Ino:  <-n.param.staticInoChan,
			Mode: fuse.S_IFDIR,
		},
	)
	n.AddChild("users", usersInode, false)
}

func Start(mountpoint string, rootGroupIds []int, userIds []int, param *FSParam) error {
	fmt.Printf("Mounting in %v\n", mountpoint)

	opts := &fs.Options{}
	opts.Debug = true

	param.staticInoChan = make(chan uint64)
	root := &rootNode{
		param:        param,
		rootGroupIds: rootGroupIds,
		userIds:      userIds,
	}

	go staticInoGenerator(root.param.staticInoChan)

	server, err := fs.Mount(mountpoint, root, opts)
	if err != nil {
		return fmt.Errorf("mount failed: %v", err)
	}
	server.Wait()

	return nil
}

func staticInoGenerator(staticInoChan chan<- uint64) {
	i := staticInodeStart
	for {
		staticInoChan <- i
		i++
	}
}
