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

	// contentLock   sync.RWMutex
	// cachedRepositoryGroupSource map[string]*CachedContent
	cachedRepositoryGroupSource sync.Map
}

func NewForgeCache(backend types.GitForge, logger *slog.Logger) types.GitForgeCacher {
	return &Cache{
		backend: backend,
		logger:  logger,
	}
}

type CachedContent struct {
	GetContent   func() (types.RepositoryGroupContent, error)
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

func (c *Cache) FetchGroupContent(ctx context.Context, source types.RepositoryGroupSource) (types.RepositoryGroupContent, error) {
	logger := c.logger.With("groupID", source.GetGroupID()).With("groupPath", source.GetGroupPath())

	cachedContent := CachedContent{
		GetContent: sync.OnceValues(func() (types.RepositoryGroupContent, error) {
			logger.Info("Fetching content from backend")
			return c.backend.FetchGroupContent(ctx, source)
		}),
		creationTime: time.Now(),
	}
	actual, loaded := c.cachedRepositoryGroupSource.LoadOrStore(source.GetGroupPath(), &cachedContent)
	if loaded {
		logger.Debug("Cache hit")
		// If already loaded, return the existing cached content or wait for it to be available
		return actual.(*CachedContent).GetContent()
	} else {
		logger.Info("Cache miss")
		// Do the actual fetch in the background
		content, err := cachedContent.GetContent()
		if err != nil {
			// If there was an error fetching the content, remove the cache entry to allow for retries
			c.cachedRepositoryGroupSource.Delete(source.GetGroupPath())
		}
		return content, err
	}
}

func (c *Cache) InvalidateCache(source types.RepositoryGroupSource) {
	c.cachedRepositoryGroupSource.Delete(source.GetGroupPath())
}
