package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/phanivarun05/kubeclusterflow/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sClientWrapper struct {
	Clientset kubernetes.Interface
	Context   string
	Namespace string
}

// Client Constructor
func NewK8sClientWrapper(contextName, namespace string) (*K8sClientWrapper, error) {
	kubeconfigPath := config.GetKubeConfigPath()

	// setup loading rules for ~/.kube/config
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfigPath

	// apply context override if specified
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: contextName,
	}

	// build REST client configuration
	restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	).ClientConfig()

	if err != nil {
		return nil, fmt.Errorf("failed to build rest config: %w", err)
	}

	// create core typed kubernetes clientset
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	// default namespace if empty
	if namespace == "" {
		namespace = "default"
	}

	return &K8sClientWrapper{
		Clientset: clientSet,
		Context:   contextName,
		Namespace: namespace,
	}, nil
}

// Listing Clusters and Namespaces
func (w *K8sClientWrapper) ListNamespaces() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nsList, err := w.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var names []string
	for _, ns := range nsList.Items {
		names = append(names, ns.Name)
	}

	return names, nil
}
