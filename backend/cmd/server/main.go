package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/mahjong/backend/internal/db"
	"github.com/mahjong/backend/internal/engine"
	"github.com/mahjong/backend/internal/models"
	"github.com/mahjong/backend/internal/room"
	"github.com/mahjong/backend/internal/ws"
)

func main() {
	// Load .env file if present (ignored in production)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")

	// Initialize database (optional for dev — server works without it)
	var database *db.DB
	if databaseURL != "" {
		var err error
		database, err = db.New(databaseURL)
		if err != nil {
			log.Printf("Warning: database unavailable: %v", err)
		} else {
			defer database.Close()
		}
	}

	// Initialize WebSocket hub and room manager
	hub := ws.NewHub()
	roomMgr := room.NewManager(hub)

	// WebSocket handler with message routing
	wsHandler := &ws.Handler{
		Hub:       hub,
		OnMessage: createMessageRouter(roomMgr, database),
		OnDisconnect: func(client *ws.Client) {
			if client.RoomCode != "" && client.Seat >= 0 {
				roomMgr.HandleDisconnect(client.RoomCode, client.Seat)
			}
		},
	}

	// HTTP routes
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/mahjong/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// REST: Create guest user and get session token
	mux.HandleFunc("POST /api/mahjong/auth/guest", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Nickname string `json:"nickname"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Nickname == "" {
			http.Error(w, `{"error":"nickname required"}`, http.StatusBadRequest)
			return
		}

		if database == nil {
			// No DB: return a generated token
			token := fmt.Sprintf("guest_%s", req.Nickname)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"user_id":       token,
				"session_token": token,
				"nickname":      req.Nickname,
			})
			return
		}

		user, err := database.CreateGuestUser(r.Context(), req.Nickname)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"user_id":       user.ID,
			"session_token": user.SessionToken,
			"nickname":      user.Nickname,
		})
	})

	// REST: Create room
	mux.HandleFunc("POST /api/mahjong/rooms", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SessionToken string `json:"session_token"`
			Nickname     string `json:"nickname"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
			return
		}

		userID := req.SessionToken // In guest mode, token is the ID
		rm, err := roomMgr.CreateRoom(userID, req.SessionToken, req.Nickname)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"code": rm.Code,
		})
	})

	// REST: Get room info
	mux.HandleFunc("GET /api/mahjong/rooms/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		rm := roomMgr.GetRoom(code)
		if rm == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"room not found"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rm.GetInfo())
	})

	// WebSocket upgrade
	mux.HandleFunc("GET /api/mahjong/ws", wsHandler.ServeWS)

	// CORS middleware for development
	handler := corsMiddleware(mux)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting Mahjong server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// createMessageRouter returns a handler that routes incoming WebSocket messages.
func createMessageRouter(roomMgr *room.Manager, database *db.DB) ws.MessageHandler {
	return func(client *ws.Client, msg models.ClientMessage) {
		var err error

		switch msg.Type {
		case models.MsgJoinRoom:
			err = handleJoinRoom(client, roomMgr, msg)

		case models.MsgLeaveRoom:
			handleLeaveRoom(client, roomMgr)

		case models.MsgPlayerReady:
			err = handlePlayerReady(client, roomMgr)

		case models.MsgStartGame:
			err = handleStartGame(client, roomMgr)

		case models.MsgConfigureRoom:
			if msg.Config != nil {
				err = roomMgr.ConfigureRoom(client.RoomCode, client.UserID, *msg.Config)
				if err == nil {
					client.Hub.BroadcastToRoom(client.RoomCode, models.ServerMessage{
						Type:   models.MsgConfigUpdated,
						Config: msg.Config,
					})
				}
			}

		case models.MsgDiscard:
			err = roomMgr.HandleDiscard(client.RoomCode, client.Seat, msg.Tile)

		case models.MsgChi:
			if len(msg.Tiles) == 2 {
				err = roomMgr.HandleReaction(client.RoomCode, client.Seat, engine.PlayerReaction{
					Type:     engine.ReactionChi,
					ChiTiles: [2]models.TileCode{msg.Tiles[0], msg.Tiles[1]},
				})
			}

		case models.MsgPong:
			err = roomMgr.HandleReaction(client.RoomCode, client.Seat, engine.PlayerReaction{
				Type: engine.ReactionPong,
			})

		case models.MsgGang:
			switch msg.GangType {
			case "open":
				err = roomMgr.HandleReaction(client.RoomCode, client.Seat, engine.PlayerReaction{
					Type: engine.ReactionGang,
				})
			case "closed":
				err = roomMgr.HandleClosedKong(client.RoomCode, client.Seat, msg.Tile)
			case "add":
				err = roomMgr.HandleAddKong(client.RoomCode, client.Seat, msg.Tile)
			}

		case models.MsgHu:
			// Determine context: is this a reaction or self-draw?
			rm := roomMgr.GetRoom(client.RoomCode)
			if rm != nil && rm.Game != nil {
				switch rm.Game.Phase {
				case engine.PhasePlayerTurn:
					err = roomMgr.HandleSelfDrawWin(client.RoomCode, client.Seat)
				case engine.PhaseAwaitingReaction, engine.PhaseAwaitingRobKong:
					err = roomMgr.HandleReaction(client.RoomCode, client.Seat, engine.PlayerReaction{
						Type: engine.ReactionHu,
					})
				}
			}

		case models.MsgPass:
			err = roomMgr.HandleReaction(client.RoomCode, client.Seat, engine.PlayerReaction{
				Type: engine.ReactionPass,
			})
		}

		if err != nil {
			client.SendMessage(models.ServerMessage{
				Type:         models.MsgError,
				ErrorCode:    "action_failed",
				ErrorMessage: err.Error(),
			})
		}
	}
}

func handleJoinRoom(client *ws.Client, roomMgr *room.Manager, msg models.ClientMessage) error {
	client.SessionToken = msg.SessionToken
	client.Nickname = msg.Nickname
	client.UserID = msg.SessionToken // guest mode

	rm, seat, err := roomMgr.JoinRoom(msg.Code, client.UserID, msg.SessionToken, msg.Nickname)
	if err != nil {
		return err
	}

	client.Seat = seat
	client.Hub.JoinRoom(client, msg.Code)

	seatVal := seat
	client.SendMessage(models.ServerMessage{
		Type:     models.MsgRoomJoined,
		RoomCode: rm.Code,
		YourSeat: &seatVal,
		Players:  rm.GetPlayerInfos(),
		Config:   &rm.Config,
	})

	// If game is in progress, send full state (reconnection)
	if rm.Game != nil {
		if err := roomMgr.HandleReconnect(msg.Code, client); err != nil {
			log.Printf("reconnect error: %v", err)
		}
	} else {
		// Notify others
		client.Hub.BroadcastToRoom(msg.Code, models.ServerMessage{
			Type:     models.MsgPlayerJoined,
			Seat:     &seatVal,
			Nickname: msg.Nickname,
		})
	}

	return nil
}

func handleLeaveRoom(client *ws.Client, roomMgr *room.Manager) {
	if client.RoomCode == "" {
		return
	}
	seat := client.Seat
	code := client.RoomCode
	roomMgr.LeaveRoom(code, seat)
	client.Hub.LeaveRoom(client)

	client.Hub.BroadcastToRoom(code, models.ServerMessage{
		Type: models.MsgPlayerLeft,
		Seat: &seat,
	})
}

func handlePlayerReady(client *ws.Client, roomMgr *room.Manager) error {
	if err := roomMgr.SetReady(client.RoomCode, client.Seat); err != nil {
		return err
	}
	seat := client.Seat
	client.Hub.BroadcastToRoom(client.RoomCode, models.ServerMessage{
		Type: models.MsgPlayerReadyServer,
		Seat: &seat,
	})
	return nil
}

func handleStartGame(client *ws.Client, roomMgr *room.Manager) error {
	return roomMgr.StartGame(client.RoomCode, client.UserID)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

