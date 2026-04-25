package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	sharedtypes "gpsgo/shared/types"
)

// buildAllowedOrigins parses WS_ALLOWED_ORIGINS (comma-separated) into a
// fast-lookup map. If the env var is not set, the map is empty and all origins
// will be rejected, preventing cross-site WebSocket hijacking.
func buildAllowedOrigins() map[string]bool {
	raw := os.Getenv("WS_ALLOWED_ORIGINS")
	allowed := make(map[string]bool)
	if raw == "" {
		return allowed
	}
	for _, origin := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			allowed[trimmed] = true
		}
	}
	return allowed
}

var allowedOrigins = buildAllowedOrigins()

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Allow requests with no Origin header (e.g. same-origin, curl, server-to-server).
		if origin == "" {
			return true
		}
		return allowedOrigins[origin]
	},
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	tenantID string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan sharedtypes.LocationUpdatedEvent
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan sharedtypes.LocationUpdatedEvent),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.conn.Close()
			}
			h.mu.Unlock()
		case evt := <-h.broadcast:
			b, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			h.mu.Lock()
			for client := range h.clients {
				// Only broadcast to clients of the same tenant
				if client.tenantID == evt.TenantID {
					if err := client.conn.WriteMessage(websocket.TextMessage, b); err != nil {
						client.conn.Close()
						delete(h.clients, client)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Broadcast(evt sharedtypes.LocationUpdatedEvent) {
	h.broadcast <- evt
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, tenantID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: h, conn: conn, tenantID: tenantID}
	h.register <- client

	go func() {
		defer func() { h.unregister <- client }()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}
