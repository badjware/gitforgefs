package types

import (
	"context"
)

type GitForge interface {
	FetchRootGroupContent(ctx context.Context) (map[string]RepositoryGroupSource, error)
	FetchGroupContent(ctx context.Context, source RepositoryGroupSource) (RepositoryGroupContent, error)
}

type RepositoryGroupSource interface {
	GetGroupID() uint64
	GetGroupName() string
	GetGroupPath() string
}

type RepositorySource interface {
	GetRepositoryID() uint64
	GetRepositoryName() string
	GetRepositoryPath() string
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
