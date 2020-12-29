package fs

import (
	"fmt"

	"github.com/badjware/gitlabfs/git"
	"github.com/badjware/gitlabfs/gitlab"

	"github.com/hanwen/go-fuse/v2/fs"
)

func Start(gf gitlab.GroupFetcher, gp git.GitClonerPuller, mountpoint string, rootGrouptID int) error {
	fmt.Printf("Mounting in %v\n", mountpoint)

	opts := &fs.Options{}
	opts.Debug = true
	root, err := newRootGroupNode(gf, gp, rootGrouptID)
	if err != nil {
		return fmt.Errorf("root group fetch fail: %v", err)
	}
	server, err := fs.Mount(mountpoint, root, opts)
	if err != nil {
		return fmt.Errorf("mount failed: %v", err)
	}
	server.Wait()

	return nil
}
