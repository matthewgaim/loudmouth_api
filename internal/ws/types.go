package ws

import (
	"database/sql"
	"time"

	"github.com/gorilla/websocket"
)

type Hub struct {
	media_id               string
	clients                map[*Client]bool
	broadcastLiveMessage   chan InboundMessage
	broadcastStaleMessages chan *Client
	register               chan *Client
	unregister             chan *Client
	dbConn                 *sql.DB
}

type Client struct {
	hub            *Hub
	conn           *websocket.Conn
	send           chan []byte
	minTimeOfMedia int
	maxTimeOfMedia int
}

type OutboundMessage struct {
	CommentId   int       `json:"comment_id"`
	Message     string    `json:"message"`
	Poster      string    `json:"poster"`
	TimeOfMedia int       `json:"time_of_media"`
	CreatedAt   time.Time `json:"created_at"`
}

type InboundMessage struct {
	Message     string `json:"message"`
	Poster      string `json:"poster"`
	TimeOfMedia int    `json:"time_of_media"`
	MediaId     string `json:"media_id"`
}
