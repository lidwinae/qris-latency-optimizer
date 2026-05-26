package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all (restrict in production)
	},
}

// HandleWebSocket - upgrade HTTP to WebSocket
func (h *Hub) HandleWebSocket(c *gin.Context) {
	merchantID := c.Query("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "merchant_id required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		MerchantID: merchantID,
		Conn:       conn,
		Send:       make(chan interface{}, 256),
		Close:      make(chan bool),
		Hub:        h,
	}

	h.register <- client

	go h.readPump(client)
	go h.writePump(client)
}

// readPump - handle incoming messages
func (h *Hub) readPump(client *Client) {
	defer func() {
		client.Hub.unregister <- client
		client.Conn.Close()
	}()

	for {
		_, msg, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error: %v", err)
			}
			return
		}

		// Handle ping/pong if needed
		_ = msg
	}
}

// writePump - send messages to client
func (h *Hub) writePump(client *Client) {
	defer client.Conn.Close()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteJSON(message); err != nil {
				log.Printf("❌ WebSocket write error: %v", err)
				return
			}

		case <-client.Close:
			return
		}
	}
}