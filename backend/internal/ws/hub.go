package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/mahjong/backend/internal/models"
)

// Client represents a connected WebSocket client.
type Client struct {
	Conn         *websocket.Conn
	Hub          *Hub
	UserID       string
	SessionToken string
	Nickname     string
	RoomCode     string
	Seat         int
	Send         chan []byte
	mu           sync.Mutex
}

// SendMessage marshals and sends a server message to this client.
func (c *Client) SendMessage(msg models.ServerMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}
	select {
	case c.Send <- data:
	default:
		log.Printf("client %s send buffer full, dropping message", c.UserID)
	}
}

// WritePump pumps messages from the Send channel to the WebSocket connection.
func (c *Client) WritePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		c.mu.Lock()
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		c.mu.Unlock()
		if err != nil {
			log.Printf("write error for %s: %v", c.UserID, err)
			return
		}
	}
}

// Hub manages all WebSocket connections grouped by room.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]bool             // all connected clients
	rooms   map[string]map[*Client]bool  // roomCode -> set of clients
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		rooms:   make(map[string]map[*Client]bool),
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = true
}

// Unregister removes a client from the hub and its room.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)
	}
	if client.RoomCode != "" {
		if room, ok := h.rooms[client.RoomCode]; ok {
			delete(room, client)
			if len(room) == 0 {
				delete(h.rooms, client.RoomCode)
			}
		}
	}
}

// JoinRoom adds a client to a room's connection group.
func (h *Hub) JoinRoom(client *Client, roomCode string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	client.RoomCode = roomCode
	if h.rooms[roomCode] == nil {
		h.rooms[roomCode] = make(map[*Client]bool)
	}
	h.rooms[roomCode][client] = true
}

// LeaveRoom removes a client from a room's connection group.
func (h *Hub) LeaveRoom(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if client.RoomCode != "" {
		if room, ok := h.rooms[client.RoomCode]; ok {
			delete(room, client)
			if len(room) == 0 {
				delete(h.rooms, client.RoomCode)
			}
		}
		client.RoomCode = ""
	}
}

// BroadcastToRoom sends a message to all clients in a room.
func (h *Hub) BroadcastToRoom(roomCode string, msg models.ServerMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("broadcast marshal error: %v", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room, ok := h.rooms[roomCode]; ok {
		for client := range room {
			select {
			case client.Send <- data:
			default:
				log.Printf("broadcast: client %s buffer full", client.UserID)
			}
		}
	}
}

// SendToSeat sends a message to the client at a specific seat in a room.
func (h *Hub) SendToSeat(roomCode string, seat int, msg models.ServerMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room, ok := h.rooms[roomCode]; ok {
		for client := range room {
			if client.Seat == seat {
				client.SendMessage(msg)
				return
			}
		}
	}
}

// GetClientByToken finds a client by session token in a room.
func (h *Hub) GetClientByToken(roomCode, token string) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room, ok := h.rooms[roomCode]; ok {
		for client := range room {
			if client.SessionToken == token {
				return client
			}
		}
	}
	return nil
}

// GetRoomClients returns all clients in a room.
func (h *Hub) GetRoomClients(roomCode string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var clients []*Client
	if room, ok := h.rooms[roomCode]; ok {
		for client := range room {
			clients = append(clients, client)
		}
	}
	return clients
}
