package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client - represents a connected merchant
type Client struct {
	MerchantID string
	Conn       *websocket.Conn
	Send       chan interface{}
	Close      chan bool
	Hub        *Hub
}

// Hub - central WebSocket hub
type Hub struct {
	clients    map[string]*Client // key: merchant_id
	register   chan *Client
	unregister chan *Client
	broadcast  chan interface{}
	mu         sync.RWMutex
}

// NewHub - create new hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan interface{}, 256),
	}
}

// Run - start hub (blocking, call in goroutine)
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.MerchantID] = client
			h.mu.Unlock()
			log.Printf("✓ Merchant %s connected [total: %d]", client.MerchantID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			delete(h.clients, client.MerchantID)
			h.mu.Unlock()
			close(client.Send)
			log.Printf("✗ Merchant %s disconnected [total: %d]", client.MerchantID, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// Channel full
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToMerchant - send notification ke specific merchant
func (h *Hub) SendToMerchant(merchantID string, notification interface{}) error {
	h.mu.RLock()
	client, exists := h.clients[merchantID]
	h.mu.RUnlock()

	if !exists {
		return nil // Merchant not connected (OK, not an error)
	}

	select {
	case client.Send <- notification:
		log.Printf("📨 Notification sent to merchant %s", merchantID)
		return nil
	default:
		return nil // Channel full, skip
	}
}

// GetConnectedCount - for monitoring
func (h *Hub) GetConnectedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}