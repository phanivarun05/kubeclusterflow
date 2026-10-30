package localcache

import (
	"sync"

	"github.com/phanivarun05/kubeclusterflow/internal/lineage"
)

type MemoryCacheStore struct {
	mu sync.RWMutex
	// Hierarchical nested map
	// MAP[ContextName]MAP[Namespace]MAP[PodName]PodLineageInfo
	Store map[string]map[string]map[string]lineage.PodLineageInfo
}

// Set Method to Write into store
func (c *MemoryCacheStore) Set(clusterContext, namespace, podName string, lineageInfo lineage.PodLineageInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize context map if context not exists yet means creating first time
	if c.Store[clusterContext] == nil {
		c.Store[clusterContext] = make(map[string]map[string]lineage.PodLineageInfo)
	}

	// Similarly initialize namespace map if it not exists yet
	if c.Store[clusterContext][namespace] == nil {
		c.Store[clusterContext][namespace] = make(map[string]lineage.PodLineageInfo)
	}

	c.Store[clusterContext][namespace][podName] = lineageInfo
}

// Get Method to Read from the store
func (c *MemoryCacheStore) Get(clusterContext, namespace, podName string) (lineage.PodLineageInfo, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	info, exists := c.Store[clusterContext][namespace][podName]
	return info, exists
}

// Delete method to delete from store
func (c *MemoryCacheStore) Delete(clusterContext, namespace, podName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Store[clusterContext] != nil && c.Store[clusterContext][namespace] != nil {
		delete(c.Store[clusterContext][namespace], podName)
	}
}

// ListByNamespace - Retrieve pods for a specific context and namespace, also support all namespaces
func (c *MemoryCacheStore) ListByNamespace(clusterContext, namespace string) []lineage.PodLineageInfo {
	c.mu.Lock()
	defer c.mu.Unlock()

	var results = make([]lineage.PodLineageInfo, 0)

	ctxMap := c.Store[clusterContext]
	if ctxMap == nil {
		return results
	}

	// if namespace is "all", aggregate pods across all namespaces in this cluster contex
	if namespace == "all" || namespace == "" {
		for _, nsMap := range ctxMap {
			for _, podInfo := range nsMap {
				results = append(results, podInfo)
			}
		}
	} else {
		nsMap := ctxMap[namespace]
		for _, podInfo := range nsMap {
			results = append(results, podInfo)
		}
	}

	return results
}
