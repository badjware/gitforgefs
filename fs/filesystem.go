package fs

import (
	"fmt"

	"github.com/badjware/gitlabfs/gitlab"

	"github.com/hanwen/go-fuse/v2/fs"
)

func Start(gf gitlab.GroupFetcher, mountpoint string, rootGrouptID int) error {
	fmt.Printf("Mounting in %v\n", mountpoint)

	opts := &fs.Options{}
	opts.Debug = true
	root, err := newRootGroupNode(gf, rootGrouptID)
	if err != nil {
		return fmt.Errorf("root group fetch fail: %w", err)
	}
	server, err := fs.Mount(mountpoint, root, opts)
	if err != nil {
		return fmt.Errorf("mount failed: %w", err)
	}
	server.Wait()

	return nil
}
