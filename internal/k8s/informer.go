package k8s

import (
	"context"
	"fmt"
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

// =====================
// Unwrap Deleted Object
// =====================

func UnwrapDeletedObject(obj interface{}) *v1.Pod {

	pod, ok := obj.(*v1.Pod)
	if ok {
		return pod
	}

	// check is it wrapped in  a tombstone.
	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		underlyingObj := tombstone.Obj

		pod, ok := underlyingObj.(*v1.Pod)
		if ok {
			return pod
		}
	}
	return nil
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

			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				newPod, ok := newObj.(*v1.Pod)
				if !ok {
					return
				}

				linege := lineage.ClassifyPodOwnership(newPod)

				cacheStore.mutex.Lock()
				cacheStore.store[newPod.Name] = linege
				cacheStore.mutex.Unlock()
			},

			DeleteFunc: func(obj interface{}) {
				pod, ok := obj.(*v1.Pod)
				if !ok {
					pod = UnwrapDeletedObject(obj)
				}

				if pod != nil {
					cacheStore.store[pod.Name] = lineage.PodLineageInfo{}
				}
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

// ===========================
// Background Lifecycle Runner
// ===========================

func (m *ClusterInformerManager) Start(stopCtx context.Context) error {

	// start background goroutines i.e informers to listen to API server
	m.factory.Start(stopCtx.Done())

	// Block until local cache finishes it's initial API list sync
	syncedResults := m.factory.WaitForCacheSync(stopCtx.Done())

	for informerType, isSynced := range syncedResults {
		if !isSynced {
			return fmt.Errorf("failed to sync informer cache")
		}
		_ = informerType
	}

	fmt.Println("Informer caches synced successfully. Running in Background...")

	<-stopCtx.Done()
	print("shutting down informers cleanly...")

	return nil
}
