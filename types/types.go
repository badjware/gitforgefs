package types

import (
	"context"
	"time"
)

type GitClient interface {
	FetchLocalRepositoryPath(ctx context.Context, source RepositorySource) (string, error)
}

type GitForge interface {
	FetchRootGroupContent(ctx context.Context) (map[string]RepositoryGroupSource, error)
	FetchGroupContent(ctx context.Context, source RepositoryGroupSource) (RepositoryGroupContent, error)
}

type GitForgeCacher interface {
	GitForge
	InvalidateCache(source RepositoryGroupSource)
}

type RepositoryGroupSource interface {
	GetGroupID() uint64
	GetGroupName() string
	GetGroupPath() string
	GetLastModified() time.Time
}

type RepositorySource interface {
	GetRepositoryID() uint64
	GetRepositoryName() string
	GetRepositoryPath() string
	GetLastModified() time.Time
	GetCloneURL() string
	GetDefaultBranch() string
}

type RepositoryGroupContent struct {
	// a map of the subgroups contained in this group, keyed by group name
	// must not be nil
	Groups map[string]RepositoryGroupSource
	// a map of the repositories contained in this group, keyed by repository name
	// must not be nil
	Repositories map[string]RepositorySource
}
