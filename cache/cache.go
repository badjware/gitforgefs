package cache

import (
	"context"
	"sync"
	"time"

	"github.com/badjware/gitforgefs/types"
)

type Cache struct {
	backend types.GitForge

	rootContentLock   sync.RWMutex
	cachedRootContent map[string]types.GroupSource

	contentLock   sync.RWMutex
	cachedContent map[string]CachedContent
}

func NewForgeCache(backend types.GitForge) types.GitForge {
	return &Cache{
		backend: backend,

		cachedContent: map[string]CachedContent{},
	}
}

type CachedContent struct {
	types.GroupContent
	creationTime time.Time
}

func (c *Cache) FetchRootGroupContent(ctx context.Context) (map[string]types.GroupSource, error) {
	c.rootContentLock.RLock()
	if c.cachedRootContent == nil {
		c.rootContentLock.RUnlock()

		// acquire write lock
		c.rootContentLock.Lock()
		defer c.rootContentLock.Unlock()

		// check to make sure the data is still not there,
		// since RWMutex is not upgradeable and another thread may have grabbed the lock in the meantime
		if c.cachedRootContent == nil {
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

func (c *Cache) FetchGroupContent(ctx context.Context, source types.GroupSource) (types.GroupContent, error) {
	c.contentLock.RLock()
	if cachedContent, found := c.cachedContent[source.GetGroupPath()]; !found {
		c.contentLock.RUnlock()

		// acquire write lock
		c.contentLock.Lock()
		defer c.contentLock.Unlock()

		// read the map again to make sure the data is still not there
		if cachedContent, found := c.cachedContent[source.GetGroupPath()]; found {
			return cachedContent.GroupContent, nil
		}

		// fetch content from backend and cache it
		content, err := c.backend.FetchGroupContent(ctx, source)
		if err != nil {
			return types.GroupContent{}, err
		}
		c.cachedContent[source.GetGroupPath()] = CachedContent{
			GroupContent: content,
			creationTime: time.Now(),
		}
		return content, nil
	} else {
		c.contentLock.RUnlock()
		return cachedContent.GroupContent, nil
	}
}

func (c *Cache) InvalidateCache(path string) {
	c.contentLock.Lock()
	defer c.contentLock.Unlock()

	delete(c.cachedContent, path)
}
