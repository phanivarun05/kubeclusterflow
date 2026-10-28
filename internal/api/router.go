package api

import (
	"fmt"
	"net/http"

	"github.com/phanivarun05/kubeclusterflow/internal/localcache"
)

type Server struct {
	router     http.ServeMux
	cacheStore *localcache.MemoryCacheStore
	sseBroker  *SSEBroker
}

// constructor: Initializes the server and attatch dependencies
func NewServer(cacheStore *localcache.MemoryCacheStore, sseBroker *SSEBroker) *Server {
	mux := http.NewServeMux()

	server := &Server{
		router:     *mux,
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

	// Static Frontend UI Serving
	// Using Go's embed package to serve compiled react/vite dashboard build
	frontendFiles := GET_EMBED_UI_FILES()
	fileServer := http.FileServer(http.FS(frontendFiles))

	s.router.Handle("GET /", fileServer)
}

// Start the server
func (s *Server) Start(port string) error {
	print("Starting Real-Time Dashboard Server on port", port)

	// Listen and serve using our configured router multiplexer
	err := http.ListenAndServe(port, &s.router)
	if err != nil {
		fmt.Errorf("server failed: %w", err)
	}

	return nil
}
