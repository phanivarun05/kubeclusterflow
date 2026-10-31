package api

import (
	"encoding/json"
	"net/http"

	"github.com/phanivarun05/kubeclusterflow/internal/k8s"
	"github.com/phanivarun05/kubeclusterflow/internal/lineage"
	"github.com/phanivarun05/kubeclusterflow/internal/localcache"
)

// Helper func to get query params
func GET_QUERY_PARAM(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// when browser first loads the dashboard, it fetches full initial snapshot of all classified pods
func HandleGetPodsSnapshot(w http.ResponseWriter, r *http.Request, cacheStore *localcache.MemoryCacheStore) {

	// set headers for json response
	w.Header().Set("Content-Type", "applicattion/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	targetContext := GET_QUERY_PARAM(r, "context")
	targetNamespace := GET_QUERY_PARAM(r, "namespace")

	if targetContext == "" {
		http.Error(w, "Missing 'context' query parameter", http.StatusBadRequest)
	}

	// retrieve all classified pods instantly from memory cache
	Pods := cacheStore.ListByNamespace(targetContext, targetNamespace)

	// encode struct slice directly to json response stream
	encoder := json.NewEncoder(w)
	err := encoder.Encode(Pods)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Handle SSE stream, opens a long-lived http connection, registers clients to SSEBroker and flush real time events down to browser.
func HandleSSEStream(w http.ResponseWriter, r *http.Request, broker *SSEBroker) {

	// verify response writer support HTTP flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// set strict headers for server-sent events (SSE).
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// create a new client registration
	client := &SSEClient{
		channel: make(chan lineage.PodLineageInfo, 10),
	}
	broker.register <- client

	// ensure client unregisters when http request context drops
	defer func() {
		broker.unregister <- client
	}()

	clientCtx := r.Context()

	// infinite streaming loop listening events or disconnects
	for {
		select {
		case <-clientCtx.Done():
			// browser closed or cancelled
			return

		case eventData := <-client.channel:
			jsonData, _ := json.Marshal(eventData)
			w.Write([]byte("data: " + string(jsonData) + "\n\n"))

			// flush tcp buffer immediately down wire
			flusher.Flush()
		}
	}
}

// Handle Get Contexts - Returns a list of all available Kubernetes contexts/clusters
func HandleGetContexts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Call K8s helper to list contexts
	contexts, err := k8s.ListAllContexts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(contexts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleGetNamespaces- Returns Namespaces available in  a specific cluster context
func HandleGetNamespaces(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// extract query params
	targetContext := GET_QUERY_PARAM(r, "context")
	if targetContext == "" {
		http.Error(w, "Missing 'context' query parameter", http.StatusBadRequest)
		return
	}

	// Fetch live namespaces
	namespaces, err := k8s.ListNamespacesForContext(targetContext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(namespaces)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//HandleGetHelmReleases - Lists deployed helm releases for the selected context and namespace
