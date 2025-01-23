package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/matthewgaim/loudmouth_api/internal/errors"
)

var hubs = make(map[string]*Hub)
var hubsMutex sync.Mutex

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var comment_id_tester = 1

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	mediaID := r.URL.Query().Get("media_id")
	if mediaID == "" {
		errors.RespondWithError(w, http.StatusBadRequest, "media_id is required")
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %s\n", err.Error())
		return
	}
	log.Printf("Client connected to media_id: %s\n", mediaID)

	hub := getHub(mediaID)
	client := &Client{conn: conn, hub: hub}
	hub.register <- client

	// cleanup
	defer func() {
		hub.unregister <- client
		conn.Close()
		log.Printf("Client disconnected from media_id: %s\n", mediaID)
	}()

	// Incoming websocket messages
	for {
		var data InboundMessage
		err := conn.ReadJSON(&data)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for media_id %s: %v", mediaID, err)
			}
			break
		}

		if data.Poster == "PING_POSTER" {
			log.Println("pong")
		} else {
			hub.broadcast <- data
		}
	}
}

func newHub(media_id string) *Hub {
	log.Printf("New hub: %s\n", media_id)
	new_hub := &Hub{
		media_id:   media_id,
		broadcast:  make(chan InboundMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
	go new_hub.run()
	return new_hub
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.conn.Close()
				if len(h.clients) == 0 {
					hubsMutex.Lock()
					delete(hubs, h.media_id)
					hubsMutex.Unlock()
					return
				}
			}
		case message := <-h.broadcast:
			log.Printf("new comment: %s\n", message.Message)
			for client := range h.clients {
				comment_id_tester += 1
				msg := OutboundMessage{
					CommentId:   comment_id_tester, // TODO: replace with db comment id
					Message:     message.Message,
					Poster:      message.Poster,
					TimeOfMedia: message.TimeOfMedia,
				}
				err := client.conn.WriteJSON(msg)
				if err != nil {
					log.Printf("Error sending message: %s\n", err.Error())
					client.conn.Close()
					h.unregister <- client
				}
			}
		}
	}
}

func getHub(movieID string) *Hub {
	hubsMutex.Lock()
	defer hubsMutex.Unlock()

	if hub, exists := hubs[movieID]; exists {
		return hub
	}

	hub := newHub(movieID)
	hubs[movieID] = hub
	return hub
}
