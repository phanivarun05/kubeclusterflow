package api

import (
	"sync"
	//"net/http"
	//"encoding/json"
	"github.com/phanivarun05/kubeclusterflow/internal/lineage"
)

type SSEClient struct {
	channel chan lineage.PodLineageInfo
}

type SSEBroker struct {
	mu         sync.Mutex
	clients    map[*SSEClient]bool
	register   chan *SSEClient
	unregister chan *SSEClient
	broadcast  chan lineage.PodLineageInfo
}

// constructor for SSE broker
func NewSSEBroker() *SSEBroker {
	broker := &SSEBroker{
		clients:    make(map[*SSEClient]bool),
		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),
		broadcast:  make(chan lineage.PodLineageInfo),
	}

	// start background event loop listener
	go broker.Listen()

	return broker
}

// Method runs indefinitely in background goroutine
func (b *SSEBroker) Listen() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = true
			b.mu.Unlock()

		case client := <-b.unregister:
			b.mu.Lock()
			if b.clients[client] {
				delete(b.clients, client)
				close(client.channel)
			}
			b.mu.Unlock()

		case event := <-b.broadcast:
			b.mu.Lock()
			for client := range b.clients {
				// non-blocking send to prevent hanging if a client is slow
				select {
				case client.channel <- event:
				default:
					continue
				}
			}
			b.mu.Unlock()
		}
	}
}
