package ws

import "github.com/gorilla/websocket"

type Hub struct {
	media_id   string
	clients    map[*Client]bool
	broadcast  chan InboundMessage
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

type OutboundMessage struct {
	CommentId   int    `json:"comment_id"`
	Message     string `json:"message"`
	Poster      string `json:"poster"`
	TimeOfMedia int    `json:"time_of_media"`
}

type InboundMessage struct {
	Message     string `json:"message"`
	Poster      string `json:"poster"`
	TimeOfMedia int    `json:"time_of_media"`
	MediaId     string `json:"media_id"`
}
