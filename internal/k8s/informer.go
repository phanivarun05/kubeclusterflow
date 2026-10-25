package k8s

import (
	"sync"
	"time"

	"github.com/phanivarun05/kubeclusterflow/internal/lineage"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// ======================
// Structure Definitions
// ======================
type MemoryCacheStore struct {
	mutex sync.RWMutex
	store map[string]lineage.PodLineageInfo
}
type ClusterInformerManager struct {
	clientset  kubernetes.Interface
	factory    informers.SharedInformerFactory
	cacheStore *MemoryCacheStore
	namespace  string
}

// ======================================
// Constructor for the Informer Manager
// ======================================

func NewInformerManager(k8sClient kubernetes.Interface, cacheStore *MemoryCacheStore, targetNamespace string) *ClusterInformerManager {

	// Create an Informer Factory for the specified namespace or all namespaces
	var factory = informers.NewSharedInformerFactoryWithOptions(
		k8sClient,
		30*time.Minute,
		informers.WithNamespace(targetNamespace),
	)

	// Extarct the specific Pod Informer from the Core v1 API group
	podInformer := factory.Core().V1().Pods().Informer()

	// Add Event Handlers to the Pod Informer additions, updates, and deletions
	podInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod, ok := obj.(*v1.Pod)
				if !ok {
					return
				}
				// Map Pod Ownership
				lineage := lineage.ClassifyPodOwnership(pod)

				// save to local cache
				cacheStore.mutex.Lock()
				cacheStore.store[pod.Name] = lineage
				cacheStore.mutex.Unlock()
			},
		},
	)
	return &ClusterInformerManager{
		clientset:  k8sClient,
		factory:    factory,
		cacheStore: cacheStore,
		namespace:  targetNamespace,
	}
}
