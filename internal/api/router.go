package api

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/phanivarun05/kubeclusterflow/internal/localcache"
)

//go:embed all:ui/dist
var frontendDist embed.FS

type Server struct {
	router     *http.ServeMux
	cacheStore *localcache.MemoryCacheStore
	sseBroker  *SSEBroker
}

// constructor: Initializes the server and attatch dependencies
func NewServer(cacheStore *localcache.MemoryCacheStore, sseBroker *SSEBroker) *Server {
	mux := http.NewServeMux()

	server := &Server{
		router:     mux,
		cacheStore: cacheStore,
		sseBroker:  sseBroker,
	}

	// register all application routes
	server.SetupRoutes()

	return server
}

// Setting up Routes
func (s *Server) SetupRoutes() {

	// REST Snapshot Endpoint: Returns all classified pods.
	s.router.HandleFunc("GET /api/pods", func(w http.ResponseWriter, r *http.Request) {
		HandleGetPodsSnapshot(w, r, s.cacheStore)
	})

	// Real time SSE stream Endpoint: Streams live pod changes to web browser
	s.router.HandleFunc("GET /api/stream", func(w http.ResponseWriter, r *http.Request) {
		HandleSSEStream(w, r, s.sseBroker)
	})

	// Streams Contexts
	s.router.HandleFunc("GET /api/contexts", HandleGetContexts)

	// Streams Namespaces for a given context
	s.router.HandleFunc("GET /api/namespaces", HandleGetNamespaces)

	// Streams Helm Releases for a given context and namespace
	s.router.HandleFunc("GET /api/releases", HandleGetHelmReleases)

	// Static Frontend UI Serving
	// Using Go's embed package to serve compiled react/vite dashboard build
	subFS, err := fs.Sub(frontendDist, "ui/dist")
	if err != nil || hasNoFiles(subFS) {
		s.router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "tect/html; charset=utf-8")
			w.Write([]byte(`<h2>KubeClusterFlow Backend Running 🚀</h2><p>Frontend UI not built yet. Access API endpoints at <a href="/api/pods">/api/pods</a> or <a href="/api/stream">/api/stream</a>.</p>`))
		})
	} else {
		fileServer := http.FileServer(http.FS(subFS))
		s.router.Handle("GET /", fileServer)
	}
}

// helper to check if embeded FS subfolder is empty
func hasNoFiles(fsys fs.FS) bool {
	entries, err := fs.ReadDir(fsys, ".")
	return err != nil || len(entries) == 0
}

// Start the server
func (s *Server) Start(port string) error {
	print("Starting Real-Time Dashboard Server on port", port)

	// Listen and serve using our configured router multiplexer
	err := http.ListenAndServe(port, s.router)
	if err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
