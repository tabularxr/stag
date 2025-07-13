package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tabular/stag/pkg/api"
	"github.com/tabular/stag/pkg/logger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHub struct {
	clients    map[string]map[*websocket.Conn]bool
	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan WSBroadcast
	logger     logger.Logger
	mu         sync.RWMutex
}

type WSClient struct {
	conn      *websocket.Conn
	sessionID string
	send      chan api.WSMessage
}

type WSBroadcast struct {
	sessionID string
	message   api.WSMessage
}

func NewWSHub(logger logger.Logger) *WSHub {
	hub := &WSHub{
		clients:    make(map[string]map[*websocket.Conn]bool),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		broadcast:  make(chan WSBroadcast),
		logger:     logger,
	}

	go hub.run()
	return hub
}

func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.sessionID] == nil {
				h.clients[client.sessionID] = make(map[*websocket.Conn]bool)
			}
			h.clients[client.sessionID][client.conn] = true
			h.mu.Unlock()
			h.logger.Infof("Client connected to session %s", client.sessionID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.sessionID]; ok {
				if _, ok := clients[client.conn]; ok {
					delete(clients, client.conn)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.sessionID)
					}
				}
			}
			h.mu.Unlock()
			h.logger.Infof("Client disconnected from session %s", client.sessionID)

		case broadcast := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.clients[broadcast.sessionID]; ok {
				for conn := range clients {
					client := h.findClientByConn(conn, broadcast.sessionID)
					if client != nil {
						select {
						case client.send <- broadcast.message:
						default:
							close(client.send)
							delete(clients, conn)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *WSHub) findClientByConn(conn *websocket.Conn, sessionID string) *WSClient {
	return &WSClient{
		conn:      conn,
		sessionID: sessionID,
		send:      make(chan api.WSMessage, 256),
	}
}

func (h *WSHub) BroadcastToSession(sessionID string, message api.WSMessage) {
	h.broadcast <- WSBroadcast{
		sessionID: sessionID,
		message:   message,
	}
}

func WSHandler(hub *WSHub, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("session_id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Errorf("Failed to upgrade connection: %v", err)
			return
		}

		client := &WSClient{
			conn:      conn,
			sessionID: sessionID,
			send:      make(chan api.WSMessage, 256),
		}

		hub.register <- client

		go client.writePump(hub, logger)
		go client.readPump(hub, logger)
	}
}

func (c *WSClient) readPump(hub *WSHub, logger logger.Logger) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg api.WSMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		if msg.Type == api.WSMessageTypePing {
			c.send <- api.WSMessage{
				Type:      api.WSMessageTypePong,
				Timestamp: time.Now().UnixMilli(),
			}
		}
	}
}

func (c *WSClient) writePump(hub *WSHub, logger logger.Logger) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				logger.Errorf("Failed to write message: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}