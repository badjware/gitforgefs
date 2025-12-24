package cache

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/badjware/gitforgefs/types"
)

type Cache struct {
	backend types.GitForge
	logger  *slog.Logger

	rootContentLock   sync.RWMutex
	cachedRootContent map[string]types.RepositoryGroupSource

	contentLock   sync.RWMutex
	cachedContent map[string]CachedContent
}

func NewForgeCache(backend types.GitForge, logger *slog.Logger) types.GitForgeCacher {
	return &Cache{
		backend: backend,
		logger:  logger,

		cachedContent: map[string]CachedContent{},
	}
}

type CachedContent struct {
	types.RepositoryGroupContent
	creationTime time.Time
}

func (c *Cache) FetchRootGroupContent(ctx context.Context) (map[string]types.RepositoryGroupSource, error) {
	c.rootContentLock.RLock()
	if c.cachedRootContent == nil {
		c.rootContentLock.RUnlock()

		// acquire write lock
		c.rootContentLock.Lock()
		defer c.rootContentLock.Unlock()

		// check to make sure the data is still not there,
		// since RWMutex is not upgradeable and another thread may have grabbed the lock in the meantime
		if c.cachedRootContent == nil {
			c.logger.Info("Fetching root content from backend")
			content, err := c.backend.FetchRootGroupContent(ctx)
			if err != nil {
				return nil, err
			}
			c.cachedRootContent = content
		}
		return c.cachedRootContent, nil
	}
	c.rootContentLock.RUnlock()
	return c.cachedRootContent, nil
}

// TODO: improve locking strategy
func (c *Cache) FetchGroupContent(ctx context.Context, source types.RepositoryGroupSource) (types.RepositoryGroupContent, error) {
	logger := c.logger.With("groupID", source.GetGroupID()).With("groupPath", source.GetGroupPath())

	c.contentLock.RLock()
	if cachedContent, found := c.cachedContent[source.GetGroupPath()]; !found {
		c.contentLock.RUnlock()

		logger.Debug("Cache miss")

		// acquire write lock
		c.contentLock.Lock()
		defer c.contentLock.Unlock()

		// read the map again to make sure the data is still not there
		if cachedContent, found := c.cachedContent[source.GetGroupPath()]; found {
			return cachedContent.RepositoryGroupContent, nil
		}

		// fetch content from backend and cache it
		logger.Info("Fetching content from backend")
		content, err := c.backend.FetchGroupContent(ctx, source)
		if err != nil {
			return types.RepositoryGroupContent{}, err
		}
		c.cachedContent[source.GetGroupPath()] = CachedContent{
			RepositoryGroupContent: content,
			creationTime:           time.Now(),
		}
		return content, nil
	} else {
		c.contentLock.RUnlock()
		logger.Debug("Cache hit")
		return cachedContent.RepositoryGroupContent, nil
	}
}

func (c *Cache) InvalidateCache(source types.RepositoryGroupSource) {
	c.contentLock.Lock()
	defer c.contentLock.Unlock()

	delete(c.cachedContent, source.GetGroupPath())
}
