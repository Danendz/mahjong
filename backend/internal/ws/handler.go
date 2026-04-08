package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mahjong/backend/internal/models"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev; restrict in production
	},
}

// MessageHandler is called for each incoming client message.
type MessageHandler func(client *Client, msg models.ClientMessage)

// Handler holds WebSocket upgrade and message routing dependencies.
type Handler struct {
	Hub            *Hub
	OnMessage      MessageHandler
	OnDisconnect   func(client *Client)
}

// ServeWS handles WebSocket upgrade requests.
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}

	client := &Client{
		Conn: conn,
		Hub:  h.Hub,
		Seat: -1,
		Send: make(chan []byte, 64),
	}

	h.Hub.Register(client)

	go client.WritePump()
	go h.readPump(client)
}

// readPump reads messages from the WebSocket connection.
func (h *Handler) readPump(client *Client) {
	defer func() {
		if h.OnDisconnect != nil {
			h.OnDisconnect(client)
		}
		h.Hub.Unregister(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start ping ticker
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for range ticker.C {
			client.mu.Lock()
			err := client.Conn.WriteMessage(websocket.PingMessage, nil)
			client.mu.Unlock()
			if err != nil {
				return
			}
		}
	}()

	for {
		_, data, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("read error from %s: %v", client.UserID, err)
			}
			return
		}

		var msg models.ClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			client.SendMessage(models.ServerMessage{
				Type:         models.MsgError,
				ErrorCode:    "invalid_message",
				ErrorMessage: "failed to parse message",
			})
			continue
		}

		if h.OnMessage != nil {
			h.OnMessage(client, msg)
		}
	}
}
