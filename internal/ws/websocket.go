package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/matthewgaim/loudmouth_api/internal/comments"
	"github.com/matthewgaim/loudmouth_api/internal/db"
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

func HandleWebSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaID := r.URL.Query().Get("media_id")
		if mediaID == "" {
			errors.RespondWithError(w, http.StatusBadRequest, "media_id is required")
			return
		}
		if err := newMediaInDb(mediaID); err != nil {
			log.Printf("Failed to find media / upload new media: %s\n", err.Error())
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
		client := &Client{
			conn:           conn,
			hub:            hub,
			minTimeOfMedia: 0,
			maxTimeOfMedia: 0,
		}
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
				client.minTimeOfMedia = max(data.TimeOfMedia-10, 0)
				client.maxTimeOfMedia = max(data.TimeOfMedia+10, 0)
				hub.broadcastStaleMessages <- client
			} else {
				hub.broadcastLiveMessage <- data
			}
		}
	}
}

func newMediaInDb(media_host_id string) error {
	_, err := db.DBConn.Exec(`
		INSERT INTO media
			(media_host_id, title, media_type)
		VALUES
			($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, media_host_id, "CHANGE_THIS_TITLE", "CHANGE_THIS_MEDIA_TYPE")
	if err != nil {
		return err
	}
	return nil
}

func newHub(mediaID string) *Hub {
	log.Printf("New hub: %s\n", mediaID)
	new_hub := &Hub{
		media_id:               mediaID,
		broadcastLiveMessage:   make(chan InboundMessage),
		broadcastStaleMessages: make(chan *Client),
		register:               make(chan *Client),
		unregister:             make(chan *Client),
		clients:                make(map[*Client]bool),
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
		case client := <-h.broadcastStaleMessages:
			comments, err := comments.GetComments(client.hub.media_id, client.minTimeOfMedia, client.maxTimeOfMedia, db.DBConn)
			if err != nil {
				log.Println("Error getting stale comments")
			}
			comment_id_tester += 1
			msgs := []OutboundMessage{}
			for _, data := range comments {
				msg := OutboundMessage{
					CommentId:   data.Id,
					Message:     data.Message,
					Poster:      data.Poster,
					TimeOfMedia: data.TimeOfMedia,
					CreatedAt:   data.CreatedAt,
				}
				msgs = append(msgs, msg)
			}
			ws_write_err := client.conn.WriteJSON(msgs)
			if ws_write_err != nil {
				log.Printf("Error sending message: %s\n", ws_write_err.Error())
				client.conn.Close()
				h.unregister <- client
			}
		case message := <-h.broadcastLiveMessage:
			new_comment_id, created_at, err := comments.MakeComment(message.Message, message.Poster, message.TimeOfMedia, message.MediaId, db.DBConn)
			if err != nil {
				log.Printf("Error uploading comment to DB: %s\n", err.Error())
			}
			log.Printf("new comment: %d\n", new_comment_id)
			for client := range h.clients {
				comment_id_tester += 1
				msg := OutboundMessage{
					CommentId:   new_comment_id,
					Message:     message.Message,
					Poster:      message.Poster,
					TimeOfMedia: message.TimeOfMedia,
					CreatedAt:   created_at,
				}
				msgs := []OutboundMessage{msg}
				err := client.conn.WriteJSON(msgs)
				if err != nil {
					log.Printf("Error sending message: %s\n", err.Error())
					client.conn.Close()
					h.unregister <- client
				}
			}
		}

	}
}

func getHub(mediaID string) *Hub {
	hubsMutex.Lock()
	defer hubsMutex.Unlock()

	if hub, exists := hubs[mediaID]; exists {
		return hub
	}

	hub := newHub(mediaID)
	hubs[mediaID] = hub
	return hub
}
