package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/phanivarun05/kubeclusterflow/internal/api"
	"github.com/phanivarun05/kubeclusterflow/internal/k8s"
	"github.com/phanivarun05/kubeclusterflow/internal/lineage"
	"github.com/phanivarun05/kubeclusterflow/internal/localcache"
	"github.com/spf13/cobra"
)

var portFlag string

// Initialize the constructor for the in-memory cache store
func NewMemoryCacheStore() *localcache.MemoryCacheStore {
	return &localcache.MemoryCacheStore{
		Store: make(map[string]map[string]map[string]lineage.PodLineageInfo),
	}
}

// Serve function
func runServe(cmd *cobra.Command, args []string) error {
	slog.Info("Initializing KubeClusterFlow Dashboard...")

	// Initialize the Hierarchical in-memory cache store
	cacheStore := NewMemoryCacheStore()

	// Discover all available Kubernetes contexts from ~/.kube/config
	allContexts, err := k8s.ListAllContexts()
	if err != nil {
		return fmt.Errorf("failed to retrieve Kubernetes contexts: %v", err)
	}

	slog.Info("Total contexts", "contexts", strconv.Itoa(len(allContexts)))

	// Set up root context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Loop through every discovered context and start a ClusterInformerManager for each
	for _, clusterContext := range allContexts {
		slog.Info("Starting informer for context", "context", clusterContext)

		// Get typed clientset for the current context
		clientSet, err := k8s.GetClientForContext(clusterContext)
		if err != nil {
			slog.Info("Skipping, Error creating clientset for context", "context", clusterContext, "err", err)
			continue
		}

		// Initialize informer manager for the current context
		informerManager := k8s.NewInformerManager(clientSet, cacheStore, clusterContext, "")

		// Start the informer manager in a separate goroutine
		go func(ctx context.Context, manager *k8s.ClusterInformerManager) {
			if err := manager.Start(ctx); err != nil {
				slog.Info("Error starting informer for context %s: %v\n", "context", clusterContext, "err", err)
			}
		}(ctx, informerManager)
	}

	// Initialize SSE Broker for real-time updates
	sseBroker := api.NewSSEBroker()

	// Initialize the HTTP server for the dashboard
	server := api.NewServer(cacheStore, sseBroker)

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Received shutdown signal. Stopping informers and server...")
		cancel()   // Cancel the root context to stop all informers
		os.Exit(0) // Exit the application
	}()

	// Start the HTTP server
	postString := fmt.Sprintf(":%s", portFlag)
	return server.Start(postString)

}

// cobra serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launch the real-time Kubernetes cluster monitoring web dashboard",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().StringVarP(&portFlag, "port", "p", "8080", "Port for the dashboard server")
	rootCmd.AddCommand(serveCmd)
}
