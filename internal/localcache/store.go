package localcache

import (
	"sync"

	"github.com/phanivarun05/kubeclusterflow/internal/lineage"
)

type MemoryCacheStore struct {
	mu    sync.RWMutex
	Store map[string]lineage.PodLineageInfo
}

// ------------------------------
// Set Method to Write into store
// ------------------------------
func (c *MemoryCacheStore) Set(podName string, lineageInfo lineage.PodLineageInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Store[podName] = lineageInfo
}

// ---------------------------------
// Get Method to Read from the store
// ---------------------------------
func (c *MemoryCacheStore) Get(podName string) (lineage.PodLineageInfo, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	info, exists := c.Store[podName]
	return info, exists
}

// ----------------------------------
// Delete method to delete from store
// ----------------------------------
func (c *MemoryCacheStore) Delete(podName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.Store, podName)
}

// ----------------------------------
// ListAll snapshot slice of all pods
// ----------------------------------
func (c *MemoryCacheStore) ListAll() []lineage.PodLineageInfo {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]lineage.PodLineageInfo, 0, len(c.Store))
	for _, info := range c.Store {
		result = append(result, info)
	}
	return result
}
